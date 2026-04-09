package game_test

// Tests for host-specific GameSession methods added in the host-ui feature:
//   - RoundQuestionCount
//   - ScoringData
//   - CeremonyQuestion
//   - CeremonyAnswer

import (
	"testing"

	"trivia/internal/game"
)

// makeHostTestSession returns a loaded session with two rounds:
//   round 0: 2 questions
//   round 1: 1 question
func makeHostTestSession(t *testing.T) *game.GameSession {
	t.Helper()
	session := game.NewGameSession()
	_ = session.Load(game.QuizFull{
		Title: "Host Test Quiz",
		Rounds: []game.Round{
			{Name: "Round 1", Questions: []game.QuestionFull{
				{Text: "Capital of France?", Answer: "Paris"},
				{Text: "Capital of Germany?", Answer: "Berlin"},
			}},
			{Name: "Round 2", Questions: []game.QuestionFull{
				{Text: "Capital of Spain?", Answer: "Madrid"},
			}},
		},
	})
	return session
}

// --- RoundQuestionCount ---

func TestRoundQuestionCount_ReturnsCorrectCount(t *testing.T) {
	session := makeHostTestSession(t)

	if got := session.RoundQuestionCount(0); got != 2 {
		t.Errorf("expected 2 questions in round 0, got %d", got)
	}
	if got := session.RoundQuestionCount(1); got != 1 {
		t.Errorf("expected 1 question in round 1, got %d", got)
	}
}

func TestRoundQuestionCount_OutOfRangeReturnsZero(t *testing.T) {
	session := makeHostTestSession(t)

	if got := session.RoundQuestionCount(99); got != 0 {
		t.Errorf("expected 0 for out-of-range round index, got %d", got)
	}
	if got := session.RoundQuestionCount(-1); got != 0 {
		t.Errorf("expected 0 for negative round index, got %d", got)
	}
	// Exact boundary: 2 rounds → index 2 is one past the last valid index.
	if got := session.RoundQuestionCount(2); got != 0 {
		t.Errorf("expected 0 for exact-boundary round index 2, got %d", got)
	}
}

func TestRoundQuestionCount_QuizNotLoadedReturnsZero(t *testing.T) {
	session := game.NewGameSession()
	if got := session.RoundQuestionCount(0); got != 0 {
		t.Errorf("expected 0 when quiz not loaded, got %d", got)
	}
}

// --- ScoringData ---

func TestScoringData_ReturnsQuestionsWithCorrectAnswers(t *testing.T) {
	session := makeHostTestSession(t)
	_ = session.StartRound(0)

	data := session.ScoringData(0)
	if len(data) != 2 {
		t.Fatalf("expected 2 scoring questions, got %d", len(data))
	}
	if data[0].CorrectAnswer != "Paris" {
		t.Errorf("expected correct answer %q, got %q", "Paris", data[0].CorrectAnswer)
	}
	if data[1].CorrectAnswer != "Berlin" {
		t.Errorf("expected correct answer %q, got %q", "Berlin", data[1].CorrectAnswer)
	}
}

func TestScoringData_IncludesTeamSubmissions(t *testing.T) {
	session := makeHostTestSession(t)
	_ = session.StartRound(0)
	team, _ := session.RegisterTeam("Team A")
	subs := []game.Submission{
		{TeamID: team.ID, RoundIndex: 0, QuestionIndex: 0, Answer: "Paris"},
		{TeamID: team.ID, RoundIndex: 0, QuestionIndex: 1, Answer: "Berlin"},
	}
	_ = session.SubmitAnswers(team.ID, 0, subs)

	data := session.ScoringData(0)
	if len(data) == 0 {
		t.Fatal("expected scoring data, got none")
	}
	found := false
	for _, s := range data[0].Submissions {
		if s.TeamID == team.ID && s.Answer == "Paris" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected team submission with answer %q in scoring data", "Paris")
	}
}

func TestScoringData_MissingSubmissionHasEmptyAnswer(t *testing.T) {
	session := makeHostTestSession(t)
	_ = session.StartRound(0)
	team, _ := session.RegisterTeam("Team B")
	// Team B does not submit anything.

	data := session.ScoringData(0)
	if len(data) == 0 || len(data[0].Submissions) == 0 {
		t.Fatal("expected scoring data with at least one submission entry")
	}
	for _, s := range data[0].Submissions {
		if s.TeamID == team.ID && s.Answer != "" {
			t.Errorf("expected empty answer for non-submitting team, got %q", s.Answer)
		}
	}
}

func TestScoringData_OutOfRangeReturnsNil(t *testing.T) {
	session := makeHostTestSession(t)
	if got := session.ScoringData(99); got != nil {
		t.Errorf("expected nil for out-of-range round, got %v", got)
	}
	// Exact boundary: 2 rounds → index 2 is one past the last valid index.
	if got := session.ScoringData(2); got != nil {
		t.Errorf("expected nil for exact-boundary round index 2, got %v", got)
	}
}

func TestScoringData_QuizNotLoadedReturnsNil(t *testing.T) {
	session := game.NewGameSession()
	if got := session.ScoringData(0); got != nil {
		t.Errorf("expected nil when quiz not loaded, got %v", got)
	}
}

// --- CeremonyQuestion ---

func TestCeremonyQuestion_ReturnsQuestionWithoutAnswer(t *testing.T) {
	session := makeHostTestSession(t)

	q := session.CeremonyQuestion(0, 0)
	if q.Text != "Capital of France?" {
		t.Errorf("expected text %q, got %q", "Capital of France?", q.Text)
	}
	// QuestionPublic intentionally has no Answer field — the type itself enforces the strip.
}

func TestCeremonyQuestion_ReturnsCorrectIndex(t *testing.T) {
	session := makeHostTestSession(t)

	q := session.CeremonyQuestion(0, 1)
	if q.Index != 1 {
		t.Errorf("expected index 1, got %d", q.Index)
	}
}

func TestCeremonyQuestion_OutOfRangeReturnsEmpty(t *testing.T) {
	session := makeHostTestSession(t)

	q := session.CeremonyQuestion(0, 99)
	if q.Text != "" {
		t.Errorf("expected empty QuestionPublic for out-of-range index, got %q", q.Text)
	}
	q2 := session.CeremonyQuestion(99, 0)
	if q2.Text != "" {
		t.Errorf("expected empty QuestionPublic for out-of-range round, got %q", q2.Text)
	}
}

func TestCeremonyQuestion_QuizNotLoadedReturnsEmpty(t *testing.T) {
	session := game.NewGameSession()
	q := session.CeremonyQuestion(0, 0)
	if q.Text != "" {
		t.Errorf("expected empty QuestionPublic when quiz not loaded, got %q", q.Text)
	}
}

// --- CeremonyAnswer ---

func TestCeremonyAnswer_ReturnsCorrectAnswer(t *testing.T) {
	session := makeHostTestSession(t)

	answer := session.CeremonyAnswer(0, 0)
	if answer != "Paris" {
		t.Errorf("expected answer %q, got %q", "Paris", answer)
	}
	answer2 := session.CeremonyAnswer(0, 1)
	if answer2 != "Berlin" {
		t.Errorf("expected answer %q, got %q", "Berlin", answer2)
	}
}

func TestCeremonyAnswer_OutOfRangeReturnsEmpty(t *testing.T) {
	session := makeHostTestSession(t)

	if got := session.CeremonyAnswer(0, 99); got != "" {
		t.Errorf("expected empty string for out-of-range question, got %q", got)
	}
	if got := session.CeremonyAnswer(99, 0); got != "" {
		t.Errorf("expected empty string for out-of-range round, got %q", got)
	}
	// Exact boundary: round 0 has 2 questions → index 2 is one past the last valid.
	if got := session.CeremonyAnswer(0, 2); got != "" {
		t.Errorf("expected empty string for exact-boundary question index 2, got %q", got)
	}
}

func TestCeremonyAnswer_QuizNotLoadedReturnsEmpty(t *testing.T) {
	session := game.NewGameSession()
	if got := session.CeremonyAnswer(0, 0); got != "" {
		t.Errorf("expected empty string when quiz not loaded, got %q", got)
	}
}
