package game_test

// Test budget: 4 behaviors × 2 = 8 max unit tests. Using 4.
//
// Behaviors:
//   1. SubmitAnswers via GamePort locks team — IsSubmitted returns true after submit
//   2. SubmitAnswers idempotency — resubmit does NOT overwrite stored answers
//   3. SubmitAnswers stores the answers provided in the first submission
//   4. IsSubmitted returns false before any submission

import (
	"testing"

	"trivia/internal/game"
)

func makeSubmittableSession(t *testing.T) *game.GameSession {
	t.Helper()
	session := game.NewGameSession()
	_ = session.Load(game.QuizFull{
		Title: "Test Quiz",
		Rounds: []game.Round{
			{Name: "Round 1", Questions: []game.QuestionFull{
				{Text: "What is the capital of France?", Answer: "Paris"},
				{Text: "Name the three primary colors.", Answer: "Red, Yellow, Blue"},
			}},
		},
	})
	_ = session.StartRound(0)
	_, _ = session.RegisterTeam("Team Awesome")
	return session
}

func TestSubmitAnswers_IsSubmittedTrueAfterSubmit(t *testing.T) {
	session := makeSubmittableSession(t)

	submissions := []game.Submission{
		{TeamID: "team-1", RoundIndex: 0, QuestionIndex: 0, Answer: "Paris"},
	}
	if err := session.SubmitAnswers("team-1", 0, submissions); err != nil {
		t.Fatalf("SubmitAnswers returned unexpected error: %v", err)
	}

	if !session.SubmissionStatus("team-1") {
		t.Error("expected SubmissionStatus to be true after submit, got false")
	}
}

func TestSubmitAnswers_IsSubmittedFalseBeforeSubmit(t *testing.T) {
	session := makeSubmittableSession(t)

	if session.SubmissionStatus("team-1") {
		t.Error("expected SubmissionStatus to be false before submit, got true")
	}
}

func TestSubmitAnswers_StoresAnswersFromFirstSubmission(t *testing.T) {
	session := makeSubmittableSession(t)

	submissions := []game.Submission{
		{TeamID: "team-1", RoundIndex: 0, QuestionIndex: 0, Answer: "Paris"},
		{TeamID: "team-1", RoundIndex: 0, QuestionIndex: 1, Answer: "Red, Yellow, Blue"},
	}
	if err := session.SubmitAnswers("team-1", 0, submissions); err != nil {
		t.Fatalf("SubmitAnswers returned unexpected error: %v", err)
	}

	stored := session.GetSubmissions("team-1")
	if len(stored) != 2 {
		t.Fatalf("expected 2 stored submissions, got %d", len(stored))
	}
	if stored[0].Answer != "Paris" {
		t.Errorf("expected first answer %q, got %q", "Paris", stored[0].Answer)
	}
}

func TestSubmitAnswers_ResubmitDoesNotOverwriteStoredAnswers(t *testing.T) {
	session := makeSubmittableSession(t)

	first := []game.Submission{
		{TeamID: "team-1", RoundIndex: 0, QuestionIndex: 0, Answer: "Paris"},
	}
	second := []game.Submission{
		{TeamID: "team-1", RoundIndex: 0, QuestionIndex: 0, Answer: "OVERWRITE"},
	}

	_ = session.SubmitAnswers("team-1", 0, first)
	_ = session.SubmitAnswers("team-1", 0, second) // resubmit — must be idempotent

	stored := session.GetSubmissions("team-1")
	if len(stored) == 0 {
		t.Fatal("expected stored submissions, got none")
	}
	if stored[0].Answer != "Paris" {
		t.Errorf("idempotency violated: expected original answer %q, got %q", "Paris", stored[0].Answer)
	}
}
