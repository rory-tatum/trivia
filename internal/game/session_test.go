package game_test

import (
	"testing"

	"trivia/internal/game"
)

// Test Budget: 2 behaviors × 2 = 4 max unit tests.
// Behavior 1: NewGameSession produces a non-empty session ID.
// Behavior 2: GetSessionID is consistent across calls.

func TestNewGameSession_HasNonEmptySessionID(t *testing.T) {
	session := game.NewGameSession()
	if session.GetSessionID() == "" {
		t.Error("expected NewGameSession to produce a non-empty session ID")
	}
}

func TestNewGameSession_SessionIDIsConsistent(t *testing.T) {
	session := game.NewGameSession()
	first := session.GetSessionID()
	second := session.GetSessionID()
	if first != second {
		t.Errorf("GetSessionID must return the same value on repeated calls: %q != %q", first, second)
	}
}

func TestNewGameSession_SessionIDsAreDifferentAcrossSessions(t *testing.T) {
	s1 := game.NewGameSession()
	s2 := game.NewGameSession()
	if s1.GetSessionID() == s2.GetSessionID() {
		t.Error("expected different session IDs for different GameSession instances")
	}
}
