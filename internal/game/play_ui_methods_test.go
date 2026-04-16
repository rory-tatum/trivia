package game_test

// Tests for GameSession methods added in the play-ui feature (step 01-01):
//   - VerdictsByQuestion  (DEP-02: verdicts array in ceremony_answer_revealed)
//   - RoundScoresWithNames (DEP-04: team names in round_scores_published)
//   - FinalScoresWithNames (DEP-04: team names in game_over)

import (
	"testing"

	"trivia/internal/game"
)

// makePlayUITestSession returns a session with one round of 2 questions,
// two registered teams, both submitted, and scoring applied.
// Team "Alpha" answered question 0 correctly; "Beta" answered incorrectly.
func makePlayUITestSession(t *testing.T) (*game.GameSession, string, string) {
	t.Helper()
	session := game.NewGameSession()
	_ = session.Load(game.QuizFull{
		Title: "Play UI Test Quiz",
		Rounds: []game.Round{
			{Name: "Round 1", Questions: []game.QuestionFull{
				{Text: "What is 2+2?", Answer: "4"},
				{Text: "Capital of Japan?", Answer: "Tokyo"},
			}},
		},
	})
	_ = session.StartRound(0)

	alpha, _ := session.RegisterTeam("Alpha")
	beta, _ := session.RegisterTeam("Beta")

	// Both teams submit answers for round 0.
	_ = session.SubmitAnswers(alpha.ID, 0, []game.Submission{
		{TeamID: alpha.ID, RoundIndex: 0, QuestionIndex: 0, Answer: "4"},
		{TeamID: alpha.ID, RoundIndex: 0, QuestionIndex: 1, Answer: "Tokyo"},
	})
	_ = session.SubmitAnswers(beta.ID, 0, []game.Submission{
		{TeamID: beta.ID, RoundIndex: 0, QuestionIndex: 0, Answer: "wrong"},
		{TeamID: beta.ID, RoundIndex: 0, QuestionIndex: 1, Answer: "wrong"},
	})

	// End round and begin scoring.
	_ = session.ForceEndRound(0)
	_ = session.BeginScoring()

	// Mark verdicts: Alpha correct on Q0, Beta incorrect on Q0.
	_ = session.MarkAnswerVerdict(alpha.ID, 0, 0, game.VerdictCorrect)
	_ = session.MarkAnswerVerdict(beta.ID, 0, 0, game.VerdictIncorrect)
	_ = session.MarkAnswerVerdict(alpha.ID, 0, 1, game.VerdictCorrect)
	_ = session.MarkAnswerVerdict(beta.ID, 0, 1, game.VerdictIncorrect)

	return session, alpha.ID, beta.ID
}

// --- VerdictsByQuestion ---

func TestVerdictsByQuestion_ReturnsVerdictForEachTeam(t *testing.T) {
	session, alphaID, betaID := makePlayUITestSession(t)

	verdicts := session.VerdictsByQuestion(0, 0)

	if len(verdicts) != 2 {
		t.Fatalf("expected 2 verdicts, got %d", len(verdicts))
	}
	found := map[string]string{}
	for _, v := range verdicts {
		found[v.TeamID] = v.Verdict
	}
	if found[alphaID] != "correct" {
		t.Errorf("expected Alpha verdict %q, got %q", "correct", found[alphaID])
	}
	if found[betaID] != "incorrect" {
		t.Errorf("expected Beta verdict %q, got %q", "incorrect", found[betaID])
	}
}

func TestVerdictsByQuestion_IncludesTeamName(t *testing.T) {
	session, alphaID, _ := makePlayUITestSession(t)

	verdicts := session.VerdictsByQuestion(0, 0)

	for _, v := range verdicts {
		if v.TeamID == alphaID && v.TeamName != "Alpha" {
			t.Errorf("expected TeamName %q for Alpha, got %q", "Alpha", v.TeamName)
		}
	}
}

// --- RoundScoresWithNames ---

func TestRoundScoresWithNames_ReturnsEntryWithTeamName(t *testing.T) {
	session, _, _ := makePlayUITestSession(t)
	_ = session.PublishRoundScores(0)

	entries := session.RoundScoresWithNames(0)

	if len(entries) == 0 {
		t.Fatal("expected at least one score entry, got none")
	}
	found := false
	for _, e := range entries {
		if e.TeamName == "Alpha" {
			found = true
			if e.RoundScore != 2 {
				t.Errorf("Alpha round score: expected 2, got %d", e.RoundScore)
			}
		}
	}
	if !found {
		t.Error("expected entry for team Alpha in round scores")
	}
}

func TestRoundScoresWithNames_IncludesRunningTotal(t *testing.T) {
	session, _, _ := makePlayUITestSession(t)
	_ = session.PublishRoundScores(0)

	entries := session.RoundScoresWithNames(0)

	for _, e := range entries {
		if e.TeamName == "Alpha" && e.RunningTotal != 2 {
			t.Errorf("Alpha running total: expected 2, got %d", e.RunningTotal)
		}
	}
}
