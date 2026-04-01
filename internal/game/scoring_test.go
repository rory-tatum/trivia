package game_test

import (
	"testing"

	"trivia/internal/game"
)

// TestCorrectVerdictIncrementsTeamScore verifies that applying a correct verdict
// adds one point to the team's score for that round.
func TestCorrectVerdictIncrementsTeamScore(t *testing.T) {
	scores := game.NewRoundScores()
	scores.ApplyVerdict("team-1", game.VerdictCorrect)

	if scores.TeamScore("team-1") != 1 {
		t.Errorf("expected score 1 after correct verdict, got %d", scores.TeamScore("team-1"))
	}
}

// TestIncorrectVerdictDoesNotIncrementScore verifies that an incorrect verdict
// leaves the team's score unchanged.
func TestIncorrectVerdictDoesNotIncrementScore(t *testing.T) {
	scores := game.NewRoundScores()
	scores.ApplyVerdict("team-1", game.VerdictIncorrect)

	if scores.TeamScore("team-1") != 0 {
		t.Errorf("expected score 0 after incorrect verdict, got %d", scores.TeamScore("team-1"))
	}
}

// TestMultipleCorrectVerdictsAccumulate verifies score accumulates across
// multiple correct answers from the same team.
func TestMultipleCorrectVerdictsAccumulate(t *testing.T) {
	scores := game.NewRoundScores()
	scores.ApplyVerdict("team-1", game.VerdictCorrect)
	scores.ApplyVerdict("team-1", game.VerdictCorrect)

	if scores.TeamScore("team-1") != 2 {
		t.Errorf("expected score 2 after two correct verdicts, got %d", scores.TeamScore("team-1"))
	}
}
