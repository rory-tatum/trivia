package handler_test

// Test budget: 4 distinct behaviors x 2 = 8 max unit tests. Using 4.
//
// Behaviors:
//   1. team_register with new name -> returns team_registered event with ID and DeviceToken
//   2. team_register with duplicate name -> returns error event
//   3. submit_answers -> returns submission_ack event with locked=true
//   4. draft_answer -> no error response (fire-and-forget)

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

// helpers

func startPlayServer(t *testing.T) (*httptest.Server, *hub.Hub, *game.GameSession) {
	t.Helper()
	h := hub.NewHub()
	session := game.NewLoadedSession()
	ph := handler.NewPlayHandler(h, session, session)
	srv := httptest.NewServer(ph)
	t.Cleanup(srv.Close)
	return srv, h, session
}

func dialPlay(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "/ws?room=play"
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("dial play: %v", err)
	}
	t.Cleanup(func() { conn.Close(websocket.StatusNormalClosure, "") })
	return conn
}

func sendMsg(t *testing.T, conn *websocket.Conn, msg map[string]interface{}) {
	t.Helper()
	if err := wsjson.Write(context.Background(), conn, msg); err != nil {
		t.Fatalf("send: %v", err)
	}
}

func readEvent(t *testing.T, conn *websocket.Conn, timeout time.Duration) map[string]interface{} {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var msg map[string]interface{}
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatalf("read event: %v", err)
	}
	return msg
}

// skipStateSnapshot discards the initial state_snapshot event sent on connection.
func skipStateSnapshot(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	msg := readEvent(t, conn, 2*time.Second)
	if msg["event"] != "state_snapshot" {
		t.Fatalf("expected initial state_snapshot, got %v", msg["event"])
	}
}

func TestPlayHandler_TeamRegister_ReturnsTeamRegisteredWithIDAndToken(t *testing.T) {
	srv, _, _ := startPlayServer(t)
	conn := dialPlay(t, srv)
	skipStateSnapshot(t, conn)

	sendMsg(t, conn, map[string]interface{}{
		"event":   "team_register",
		"payload": map[string]interface{}{"team_name": "Team Alpha"},
	})

	msg := readEvent(t, conn, 2*time.Second)
	if msg["event"] != "team_registered" {
		t.Fatalf("expected event team_registered, got %v", msg["event"])
	}
	payload, _ := msg["payload"].(map[string]interface{})
	if payload["team_id"] == "" || payload["team_id"] == nil {
		t.Error("expected non-empty team_id in team_registered payload")
	}
	if payload["device_token"] == "" || payload["device_token"] == nil {
		t.Error("expected non-empty device_token in team_registered payload")
	}
}

func TestPlayHandler_TeamRegister_DuplicateName_ReturnsError(t *testing.T) {
	srv, _, session := startPlayServer(t)

	// Pre-register the name
	_, err := session.RegisterTeam("Team Beta")
	if err != nil {
		t.Fatalf("setup: register team: %v", err)
	}

	conn := dialPlay(t, srv)
	skipStateSnapshot(t, conn)

	sendMsg(t, conn, map[string]interface{}{
		"event":   "team_register",
		"payload": map[string]interface{}{"team_name": "Team Beta"},
	})

	msg := readEvent(t, conn, 2*time.Second)
	if msg["event"] != "error" {
		t.Fatalf("expected error event for duplicate team name, got %v", msg["event"])
	}
}

func TestPlayHandler_SubmitAnswers_ReturnsSubmissionAck(t *testing.T) {
	srv, _, session := startPlayServer(t)
	conn := dialPlay(t, srv)
	skipStateSnapshot(t, conn)

	// Register a team first to get a team_id.
	sendMsg(t, conn, map[string]interface{}{
		"event":   "team_register",
		"payload": map[string]interface{}{"team_name": "Team Gamma"},
	})
	regMsg := readEvent(t, conn, 2*time.Second)
	payload, _ := regMsg["payload"].(map[string]interface{})
	teamID, _ := payload["team_id"].(string)

	// Start and end the round so submission is accepted.
	if err := session.StartRound(0); err != nil {
		t.Fatalf("start round: %v", err)
	}
	if err := session.ForceEndRound(0); err != nil {
		t.Fatalf("end round: %v", err)
	}

	sendMsg(t, conn, map[string]interface{}{
		"event": "submit_answers",
		"payload": map[string]interface{}{
			"team_id":     teamID,
			"round_index": 0,
			"answers": []interface{}{
				map[string]interface{}{"question_index": 0, "answer": "A1"},
			},
		},
	})

	msg := readEvent(t, conn, 2*time.Second)
	if msg["event"] != "submission_ack" {
		t.Fatalf("expected submission_ack event, got %v", msg["event"])
	}
	ackPayload, _ := msg["payload"].(map[string]interface{})
	locked, _ := ackPayload["locked"].(bool)
	if !locked {
		t.Error("expected locked=true in submission_ack")
	}
}

func TestPlayHandler_DraftAnswer_NoErrorResponse(t *testing.T) {
	srv, _, _ := startPlayServer(t)
	conn := dialPlay(t, srv)
	skipStateSnapshot(t, conn)

	sendMsg(t, conn, map[string]interface{}{
		"event": "draft_answer",
		"payload": map[string]interface{}{
			"round_index":    0,
			"question_index": 0,
			"answer":         "draft answer",
		},
	})

	// draft_answer is fire-and-forget: no error event should arrive within timeout
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	var msg map[string]interface{}
	err := wsjson.Read(ctx, conn, &msg)
	if err == nil {
		if msg["event"] == "error" {
			t.Errorf("draft_answer should not return an error event, got: %v", msg)
		}
	}
	// context deadline or EOF is acceptable (no error event received)
}
