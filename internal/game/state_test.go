package game_test

import (
	"testing"

	"trivia/internal/game"
)

// TestValidStateTransitions verifies that the state machine allows all
// documented valid transitions.
func TestValidStateTransitions(t *testing.T) {
	transitions := []struct {
		from game.GameState
		to   game.GameState
	}{
		{game.StateLobby, game.StateRoundActive},
		{game.StateRoundActive, game.StateRoundEnded},
		{game.StateRoundEnded, game.StateScoring},
		{game.StateScoring, game.StateCeremony},
		{game.StateCeremony, game.StateRoundScores},
		{game.StateRoundScores, game.StateRoundActive},
		{game.StateRoundScores, game.StateGameOver},
	}

	for _, tc := range transitions {
		err := game.ValidateTransition(tc.from, tc.to)
		if err != nil {
			t.Errorf("expected valid transition %v -> %v, got error: %v", tc.from, tc.to, err)
		}
	}
}

// TestInvalidStateTransitions verifies that the state machine rejects invalid transitions.
func TestInvalidStateTransitions(t *testing.T) {
	invalids := []struct {
		from game.GameState
		to   game.GameState
	}{
		{game.StateLobby, game.StateGameOver},
		{game.StateRoundActive, game.StateLobby},
		{game.StateRoundEnded, game.StateLobby},
		{game.StateScoring, game.StateLobby},
		{game.StateCeremony, game.StateLobby},
		{game.StateGameOver, game.StateLobby},
	}

	for _, tc := range invalids {
		err := game.ValidateTransition(tc.from, tc.to)
		if err == nil {
			t.Errorf("expected error for invalid transition %v -> %v, got nil", tc.from, tc.to)
		}
	}
}
