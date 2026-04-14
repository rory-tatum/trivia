package handler_test

// Tests for host-specific handler behaviors added in the host-ui feature.
//
// Behaviors:
//   1. host_start_round → round_started payload includes question_count
//   2. host_begin_scoring → scoring_data event sent to host room
//   3. host_reveal_question → question_revealed total_questions is per-round count

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"trivia/internal/game"
	"trivia/internal/handler"
	"trivia/internal/hub"
)

// stubQuizLoader satisfies handler.QuizLoader using an already-loaded session.
// It pre-loads the session during construction so LoadIntoSession is a no-op.
type stubQuizLoader struct{}

func (s *stubQuizLoader) LoadIntoSession(_ string, _ game.GamePort) (handler.QuizLoadedMeta, error) {
	return handler.QuizLoadedMeta{
		Title:         "Test Quiz",
		RoundCount:    1,
		QuestionCount: 2,
	}, nil
}

func startHostServer(t *testing.T) (*httptest.Server, *hub.Hub, *game.GameSession) {
	t.Helper()
	h := hub.NewHub()
	session := game.NewLoadedSessionTwoQuestions()
	_ = session.StartRound(0) // pre-start round 0 for tests that need it
	hh := handler.NewHostHandler(h, &stubQuizLoader{}, "http://localhost", session)
	srv := httptest.NewServer(hh)
	t.Cleanup(srv.Close)
	return srv, h, session
}

func dialHost(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("dial host: %v", err)
	}
	t.Cleanup(func() { conn.Close(websocket.StatusNormalClosure, "") })
	return conn
}

func readHostEvent(t *testing.T, conn *websocket.Conn, timeout time.Duration) map[string]interface{} {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var msg map[string]interface{}
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatalf("read host event: %v", err)
	}
	return msg
}

func sendHostMsg(t *testing.T, conn *websocket.Conn, msg map[string]interface{}) {
	t.Helper()
	if err := wsjson.Write(context.Background(), conn, msg); err != nil {
		t.Fatalf("send host msg: %v", err)
	}
}

// TestHostHandler_StartRound_PayloadIncludesQuestionCount verifies that
// host_start_round broadcasts a round_started event with question_count in the payload.
func TestHostHandler_StartRound_PayloadIncludesQuestionCount(t *testing.T) {
	h := hub.NewHub()
	session := game.NewLoadedSessionTwoQuestions()
	hh := handler.NewHostHandler(h, &stubQuizLoader{}, "http://localhost", session)
	srv := httptest.NewServer(hh)
	t.Cleanup(srv.Close)

	conn := dialHost(t, srv)
	sendHostMsg(t, conn, map[string]interface{}{
		"event":   "host_start_round",
		"payload": map[string]interface{}{"round_index": 0},
	})

	msg := readHostEvent(t, conn, 2*time.Second)
	if msg["event"] != "round_started" {
		t.Fatalf("expected round_started event, got %v", msg["event"])
	}
	payload, _ := msg["payload"].(map[string]interface{})
	qCount, _ := payload["question_count"].(float64)
	if qCount != 2 {
		t.Errorf("expected question_count=2 in round_started payload, got %v", payload["question_count"])
	}
}

// TestHostHandler_StartRound_PayloadIncludesRoundName verifies that
// host_start_round broadcasts a round_started event with round_name in the payload.
func TestHostHandler_StartRound_PayloadIncludesRoundName(t *testing.T) {
	h := hub.NewHub()
	session := game.NewLoadedSessionTwoQuestions()
	hh := handler.NewHostHandler(h, &stubQuizLoader{}, "http://localhost", session)
	srv := httptest.NewServer(hh)
	t.Cleanup(srv.Close)

	conn := dialHost(t, srv)
	sendHostMsg(t, conn, map[string]interface{}{
		"event":   "host_start_round",
		"payload": map[string]interface{}{"round_index": 0},
	})

	msg := readHostEvent(t, conn, 2*time.Second)
	if msg["event"] != "round_started" {
		t.Fatalf("expected round_started event, got %v", msg["event"])
	}
	payload, _ := msg["payload"].(map[string]interface{})
	roundName, _ := payload["round_name"].(string)
	if roundName == "" {
		t.Errorf("expected non-empty round_name in round_started payload, got %q", roundName)
	}
}

// TestHostHandler_BeginScoring_SendsScoringDataToHost verifies that
// host_begin_scoring causes a scoring_data event to be sent to the host.
func TestHostHandler_BeginScoring_SendsScoringDataToHost(t *testing.T) {
	_, _, session := startHostServer(t)

	// Manually transition session to scoring-ready state.
	_ = session.ForceEndRound(0)

	h2 := hub.NewHub()
	hh := handler.NewHostHandler(h2, &stubQuizLoader{}, "http://localhost", session)
	srv2 := httptest.NewServer(hh)
	t.Cleanup(srv2.Close)

	conn := dialHost(t, srv2)
	sendHostMsg(t, conn, map[string]interface{}{
		"event":   "host_begin_scoring",
		"payload": map[string]interface{}{"round_index": 0},
	})

	// Expect scoring_opened first, then scoring_data.
	sawScoringData := false
	for i := 0; i < 3; i++ {
		msg := readHostEvent(t, conn, 2*time.Second)
		if msg["event"] == "scoring_data" {
			sawScoringData = true
			payload, _ := msg["payload"].(map[string]interface{})
			questions, _ := payload["questions"].([]interface{})
			if len(questions) != 2 {
				t.Errorf("expected 2 questions in scoring_data, got %d", len(questions))
			}
			break
		}
	}
	if !sawScoringData {
		t.Error("expected scoring_data event after host_begin_scoring, but did not receive it")
	}
}

// TestHostHandler_RevealQuestion_TotalQuestionsIsPerRound verifies that
// question_revealed carries the per-round total_questions count, not the quiz total.
func TestHostHandler_RevealQuestion_TotalQuestionsIsPerRound(t *testing.T) {
	_, _, session := startHostServer(t)

	h2 := hub.NewHub()
	hh := handler.NewHostHandler(h2, &stubQuizLoader{}, "http://localhost", session)
	srv2 := httptest.NewServer(hh)
	t.Cleanup(srv2.Close)

	conn := dialHost(t, srv2)
	sendHostMsg(t, conn, map[string]interface{}{
		"event":   "host_reveal_question",
		"payload": map[string]interface{}{"round_index": 0, "question_index": 0},
	})

	msg := readHostEvent(t, conn, 2*time.Second)
	if msg["event"] != "question_revealed" {
		t.Fatalf("expected question_revealed event, got %v", msg["event"])
	}
	payload, _ := msg["payload"].(map[string]interface{})
	totalQ, _ := payload["total_questions"].(float64)
	if totalQ != 2 {
		t.Errorf("expected total_questions=2 (per-round), got %v", payload["total_questions"])
	}
}
