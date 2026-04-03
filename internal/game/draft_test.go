package game_test

// Test budget: 2 behaviors × 2 = 4 max unit tests. Using 3.
//
// Behaviors:
//   1. SaveDraft persists text for a team/round/question triple via GamePort
//   2. SaveDraft updates overwrite the previous draft (fire-and-update semantics)

import (
	"testing"

	"trivia/internal/game"
)

func makeRevealableSession(t *testing.T) *game.GameSession {
	t.Helper()
	session := game.NewGameSession()
	_ = session.Load(game.QuizFull{
		Title: "Test Quiz",
		Rounds: []game.Round{
			{Name: "Round 1", Questions: []game.QuestionFull{
				{Text: "What is the capital of France?", Answer: "Paris"},
				{Text: "Name the three primary colors.", Answer: "Red, Green, Blue"},
			}},
		},
	})
	_ = session.StartRound(0)
	_ = session.RevealQuestion(0, 0)
	_, _ = session.RegisterTeam("Team Awesome")
	return session
}

func TestSaveDraft_PersistsDraftAnswer(t *testing.T) {
	session := makeRevealableSession(t)

	if err := session.SaveDraft("team-1", 0, 0, "Paris"); err != nil {
		t.Fatalf("SaveDraft returned unexpected error: %v", err)
	}

	draft := session.GetDraft("team-1", 0, 0)
	if draft != "Paris" {
		t.Errorf("expected draft %q, got %q", "Paris", draft)
	}
}

func TestSaveDraft_UpdateOverwritesPreviousDraft(t *testing.T) {
	session := makeRevealableSession(t)

	_ = session.SaveDraft("team-1", 0, 0, "Paris, France")
	_ = session.SaveDraft("team-1", 0, 0, "Paris")

	draft := session.GetDraft("team-1", 0, 0)
	if draft != "Paris" {
		t.Errorf("expected updated draft %q, got %q", "Paris", draft)
	}
}

func TestSaveDraft_IndependentAcrossQuestions(t *testing.T) {
	session := makeRevealableSession(t)
	_ = session.RevealQuestion(0, 1)

	_ = session.SaveDraft("team-1", 0, 0, "Paris")
	_ = session.SaveDraft("team-1", 0, 1, "Red, Yellow, Blue")

	if got := session.GetDraft("team-1", 0, 0); got != "Paris" {
		t.Errorf("q0 draft: expected %q, got %q", "Paris", got)
	}
	if got := session.GetDraft("team-1", 0, 1); got != "Red, Yellow, Blue" {
		t.Errorf("q1 draft: expected %q, got %q", "Red, Yellow, Blue", got)
	}
}
