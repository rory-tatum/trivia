package handler_test

// Test Budget: 2 distinct behaviors × 2 = 4 max unit tests. Using 2.
//
// Behaviors driven through PlayHandler (driving port):
//   1. When game has started, connecting player receives state_snapshot with ROUND_ACTIVE state
//      and correct current_round index (not LOBBY, not -1).
//   2. When question has been revealed, connecting player receives state_snapshot with
//      revealed_questions containing QuestionPublic fields (no answer field).

import (
	"context"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"trivia/internal/game"
	"trivia/internal/handler"
	"trivia/internal/hub"

	"net/http/httptest"
)

func startPlayServerWithStartedGame(t *testing.T) (*httptest.Server, *game.GameSession) {
	t.Helper()
	h := hub.NewHub()
	session := game.NewGameSession()
	_ = session.Load(game.QuizFull{
		Title: "Test Quiz",
		Rounds: []game.Round{
			{Name: "Round 1", Questions: []game.QuestionFull{
				{Text: "What is the capital of France?", Answer: "Paris"},
				{Text: "What color is the sky?", Answer: "Blue"},
			}},
		},
	})
	if err := session.StartRound(0); err != nil {
		t.Fatalf("setup: StartRound: %v", err)
	}
	ph := handler.NewPlayHandler(h, session, session)
	srv := httptest.NewServer(ph)
	t.Cleanup(srv.Close)
	return srv, session
}

func dialPlayConn(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "/ws?room=play"
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("dial play: %v", err)
	}
	t.Cleanup(func() { conn.Close(websocket.StatusNormalClosure, "") })
	return conn
}

func readEventWithTimeout(t *testing.T, conn *websocket.Conn, timeout time.Duration) map[string]interface{} {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var msg map[string]interface{}
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatalf("read event: %v", err)
	}
	return msg
}

// TestPlayHandler_LateJoiner_SnapshotShowsActiveRound verifies that a player connecting
// after game start receives a state_snapshot with ROUND_ACTIVE state and correct round index.
func TestPlayHandler_LateJoiner_SnapshotShowsActiveRound(t *testing.T) {
	srv, _ := startPlayServerWithStartedGame(t)
	conn := dialPlayConn(t, srv)

	msg := readEventWithTimeout(t, conn, 2*time.Second)

	if msg["event"] != "state_snapshot" {
		t.Fatalf("expected state_snapshot as first event, got %v", msg["event"])
	}
	payload, _ := msg["payload"].(map[string]interface{})
	state, _ := payload["state"].(string)
	if state != "ROUND_ACTIVE" {
		t.Errorf("expected state ROUND_ACTIVE in snapshot, got %q", state)
	}
	currentRound, _ := payload["current_round"].(float64)
	if int(currentRound) != 0 {
		t.Errorf("expected current_round 0 in snapshot, got %g", currentRound)
	}
}

// TestPlayHandler_LateJoiner_SnapshotRevealedQuestionsHaveNoAnswerField verifies that
// the state_snapshot sent to a late-joining player contains revealed questions as
// QuestionPublic only — the answer field must not be present.
func TestPlayHandler_LateJoiner_SnapshotRevealedQuestionsHaveNoAnswerField(t *testing.T) {
	srv, session := startPlayServerWithStartedGame(t)
	if err := session.RevealQuestion(0, 0); err != nil {
		t.Fatalf("setup: RevealQuestion: %v", err)
	}

	conn := dialPlayConn(t, srv)
	msg := readEventWithTimeout(t, conn, 2*time.Second)

	if msg["event"] != "state_snapshot" {
		t.Fatalf("expected state_snapshot as first event, got %v", msg["event"])
	}
	payload, _ := msg["payload"].(map[string]interface{})
	revealed, _ := payload["revealed_questions"].([]interface{})
	if len(revealed) == 0 {
		t.Fatal("expected at least 1 revealed question in state_snapshot, got none")
	}
	for i, q := range revealed {
		qMap, _ := q.(map[string]interface{})
		if _, hasAnswer := qMap["answer"]; hasAnswer {
			t.Errorf("revealed_questions[%d] in state_snapshot contains forbidden 'answer' field", i)
		}
	}
}
