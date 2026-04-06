package game_test

import (
	"testing"

	"trivia/internal/game"
)

// makeMinimalLoadedSession returns a session with a quiz loaded, ready for StartRound.
func makeMinimalLoadedSession(t *testing.T) *game.GameSession {
	t.Helper()
	s := game.NewGameSession()
	_ = s.Load(game.QuizFull{
		Title: "Minimal Quiz",
		Rounds: []game.Round{
			{Name: "Round 1", Questions: []game.QuestionFull{
				{Text: "Q1", Answer: "A1"},
			}},
		},
	})
	return s
}

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

// TestStartRound_UpdatesGameState verifies that a valid transition via StartRound
// actually persists the new state. This kills the CONDITIONALS_NEGATION mutant on
// transition() which would skip the g.state = to assignment for valid transitions.
func TestStartRound_UpdatesGameState(t *testing.T) {
	s := makeMinimalLoadedSession(t)

	if err := s.StartRound(0); err != nil {
		t.Fatalf("StartRound failed: %v", err)
	}

	if got := s.CurrentState(); got != game.StateRoundActive {
		t.Errorf("expected state %q after StartRound, got %q", game.StateRoundActive, got)
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
