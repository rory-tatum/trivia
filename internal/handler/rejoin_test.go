package handler_test

// Test budget: 2 behaviors × 2 = 4 max unit tests. Using 2.
//
// Behaviors:
//   1. team_rejoin with valid device_token -> returns state_snapshot with draft_answers
//   2. team_rejoin with invalid device_token -> returns error event, no state sent

import (
	"testing"
	"time"
)

func TestPlayHandler_TeamRejoin_ValidToken_ReturnsSnapshotWithDrafts(t *testing.T) {
	srv, _, session := startPlayServer(t)

	// Register a team and save a draft to verify drafts are included in snapshot.
	team, err := session.RegisterTeam("Team Delta")
	if err != nil {
		t.Fatalf("setup: register team: %v", err)
	}
	_ = session.SaveDraft(team.ID, 0, 0, "draft answer")

	conn := dialPlay(t, srv)
	// Discard initial state_snapshot.
	skipStateSnapshot(t, conn)

	sendMsg(t, conn, map[string]interface{}{
		"event": "team_rejoin",
		"payload": map[string]interface{}{
			"team_id":      team.ID,
			"device_token": team.DeviceToken,
		},
	})

	msg := readEvent(t, conn, 2*time.Second)
	if msg["event"] != "state_snapshot" {
		t.Fatalf("expected state_snapshot on rejoin, got %v", msg["event"])
	}
	payload, _ := msg["payload"].(map[string]interface{})
	drafts, ok := payload["draft_answers"]
	if !ok {
		t.Fatal("state_snapshot on rejoin must include draft_answers field")
	}
	draftList, _ := drafts.([]interface{})
	if len(draftList) == 0 {
		t.Error("expected at least one draft answer in state_snapshot, got none")
	}
}

func TestPlayHandler_TeamRejoin_InvalidToken_ReturnsError(t *testing.T) {
	srv, _, session := startPlayServer(t)

	team, err := session.RegisterTeam("Team Epsilon")
	if err != nil {
		t.Fatalf("setup: register team: %v", err)
	}
	_ = session.StartRound(0)

	conn := dialPlay(t, srv)
	skipStateSnapshot(t, conn)

	sendMsg(t, conn, map[string]interface{}{
		"event": "team_rejoin",
		"payload": map[string]interface{}{
			"team_id":      team.ID,
			"device_token": "wrong-token",
		},
	})

	msg := readEvent(t, conn, 2*time.Second)
	if msg["event"] != "error" {
		t.Fatalf("expected error event for invalid device token, got %v", msg["event"])
	}
}
