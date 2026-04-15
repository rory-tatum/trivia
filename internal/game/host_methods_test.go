package game_test

// Tests for host-specific GameSession methods added in the host-ui feature:
//   - RoundName         (added in host-ui feature)
//   - RoundQuestionCount (added in host-ui feature)
//   - ScoringData       (added in host-ui feature)
//   - CeremonyQuestion  (pre-existing method, first unit tests added here)
//   - CeremonyAnswer    (pre-existing method, first unit tests added here)
//
// CeremonyQuestion and CeremonyAnswer were introduced before this feature (commit fb833a9)
// and are called by handleCeremonyShowQuestion / handleCeremonyRevealAnswer in
// internal/handler/host.go. These are the first unit-level tests for those methods.

import (
	"testing"

	"trivia/internal/game"
)

// makeHostTestSession returns a loaded session with two rounds:
//
//	round 0: 2 questions
//	round 1: 1 question
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

// --- RoundName ---

func TestRoundName_ReturnsHumanReadableName(t *testing.T) {
	session := makeHostTestSession(t)
	cases := []struct {
		idx      int
		expected string
	}{
		{0, "Round 1"},
		{1, "Round 2"},
	}
	for _, c := range cases {
		if got := session.RoundName(c.idx); got != c.expected {
			t.Errorf("RoundName(%d): expected %q, got %q", c.idx, c.expected, got)
		}
	}
}

func TestRoundName_InvalidIndexReturnsEmpty(t *testing.T) {
	session := makeHostTestSession(t)
	for _, idx := range []int{-1, 2, 99} {
		if got := session.RoundName(idx); got != "" {
			t.Errorf("RoundName(%d): expected empty string, got %q", idx, got)
		}
	}
	unloaded := game.NewGameSession()
	if got := unloaded.RoundName(0); got != "" {
		t.Errorf("RoundName(0) on unloaded session: expected empty string, got %q", got)
	}
}

// --- RoundQuestionCount ---

func TestRoundQuestionCount_ReturnsCorrectCount(t *testing.T) {
	session := makeHostTestSession(t)
	cases := []struct {
		round    int
		expected int
	}{
		{0, 2},
		{1, 1},
	}
	for _, c := range cases {
		if got := session.RoundQuestionCount(c.round); got != c.expected {
			t.Errorf("RoundQuestionCount(%d): expected %d, got %d", c.round, c.expected, got)
		}
	}
}

func TestRoundQuestionCount_InvalidIndexReturnsZero(t *testing.T) {
	session := makeHostTestSession(t)
	// Exact boundary: 2 rounds → index 2 is one past the last valid index.
	for _, idx := range []int{-1, 2, 99} {
		if got := session.RoundQuestionCount(idx); got != 0 {
			t.Errorf("RoundQuestionCount(%d): expected 0, got %d", idx, got)
		}
	}
	unloaded := game.NewGameSession()
	if got := unloaded.RoundQuestionCount(0); got != 0 {
		t.Errorf("RoundQuestionCount(0) on unloaded session: expected 0, got %d", got)
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

func TestScoringData_InvalidIndexReturnsNil(t *testing.T) {
	session := makeHostTestSession(t)
	// Exact boundary: 2 rounds → index 2 is one past the last valid index.
	for _, idx := range []int{2, 99} {
		if got := session.ScoringData(idx); got != nil {
			t.Errorf("ScoringData(%d): expected nil, got %v", idx, got)
		}
	}
	unloaded := game.NewGameSession()
	if got := unloaded.ScoringData(0); got != nil {
		t.Errorf("ScoringData(0) on unloaded session: expected nil, got %v", got)
	}
}

// --- CeremonyQuestion ---
// CeremonyQuestion is a pre-existing method (commit fb833a9) called by
// handleCeremonyShowQuestion in internal/handler/host.go. These are its first unit tests.

func TestCeremonyQuestion_HappyPath(t *testing.T) {
	session := makeHostTestSession(t)
	cases := []struct {
		round    int
		question int
		wantText string
		wantIdx  int
	}{
		{0, 0, "Capital of France?", 0},
		{0, 1, "Capital of Germany?", 1},
	}
	for _, c := range cases {
		q := session.CeremonyQuestion(c.round, c.question)
		if q.Text != c.wantText {
			t.Errorf("CeremonyQuestion(%d,%d).Text: expected %q, got %q", c.round, c.question, c.wantText, q.Text)
		}
		if q.Index != c.wantIdx {
			t.Errorf("CeremonyQuestion(%d,%d).Index: expected %d, got %d", c.round, c.question, c.wantIdx, q.Index)
		}
		// QuestionPublic intentionally has no Answer field — the type itself enforces the strip.
	}
}

func TestCeremonyQuestion_InvalidIndexReturnsEmpty(t *testing.T) {
	session := makeHostTestSession(t)
	cases := []struct{ round, question int }{
		{0, 99},  // question out of range
		{99, 0},  // round out of range
	}
	for _, c := range cases {
		if q := session.CeremonyQuestion(c.round, c.question); q.Text != "" {
			t.Errorf("CeremonyQuestion(%d,%d): expected empty QuestionPublic, got text %q", c.round, c.question, q.Text)
		}
	}
	unloaded := game.NewGameSession()
	if q := unloaded.CeremonyQuestion(0, 0); q.Text != "" {
		t.Errorf("CeremonyQuestion(0,0) on unloaded session: expected empty, got %q", q.Text)
	}
}

// --- CeremonyAnswer ---
// CeremonyAnswer is a pre-existing method (commit fb833a9) called by
// handleCeremonyRevealAnswer in internal/handler/host.go. These are its first unit tests.

func TestCeremonyAnswer_ReturnsCorrectAnswer(t *testing.T) {
	session := makeHostTestSession(t)
	cases := []struct {
		round    int
		question int
		expected string
	}{
		{0, 0, "Paris"},
		{0, 1, "Berlin"},
	}
	for _, c := range cases {
		if got := session.CeremonyAnswer(c.round, c.question); got != c.expected {
			t.Errorf("CeremonyAnswer(%d,%d): expected %q, got %q", c.round, c.question, c.expected, got)
		}
	}
}

func TestCeremonyAnswer_InvalidIndexReturnsEmpty(t *testing.T) {
	session := makeHostTestSession(t)
	// Exact boundary: round 0 has 2 questions → index 2 is one past the last valid.
	cases := []struct{ round, question int }{
		{0, 99}, // question far out of range
		{99, 0}, // round out of range
		{0, 2},  // exact boundary
	}
	for _, c := range cases {
		if got := session.CeremonyAnswer(c.round, c.question); got != "" {
			t.Errorf("CeremonyAnswer(%d,%d): expected empty string, got %q", c.round, c.question, got)
		}
	}
	unloaded := game.NewGameSession()
	if got := unloaded.CeremonyAnswer(0, 0); got != "" {
		t.Errorf("CeremonyAnswer(0,0) on unloaded session: expected empty string, got %q", got)
	}
}
