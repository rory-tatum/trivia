package game_test

import (
	"testing"

	"trivia/internal/game"
)

// Test Budget: 2 behaviors × 2 = 4 max unit tests.
// Behavior 1: RegisterTeam with duplicate name returns exact error message.
// Behavior 2: Duplicate detection is case-insensitive.

func TestRegisterTeam_DuplicateName_ReturnsExactErrorMessage(t *testing.T) {
	session := game.NewGameSession()
	_, err := session.RegisterTeam("Quiz Killers")
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	_, err = session.RegisterTeam("Quiz Killers")
	if err == nil {
		t.Fatal("expected error for duplicate team name, got nil")
	}
	want := "That name is taken -- try a different team name"
	if err.Error() != want {
		t.Errorf("wrong error message\n got:  %q\n want: %q", err.Error(), want)
	}
}

func TestRegisterTeam_DuplicateNameCaseInsensitive_ReturnsExactErrorMessage(t *testing.T) {
	session := game.NewGameSession()
	_, err := session.RegisterTeam("Team Awesome")
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	_, err = session.RegisterTeam("team awesome")
	if err == nil {
		t.Fatal("expected error for case-insensitive duplicate team name, got nil")
	}
	want := "That name is taken -- try a different team name"
	if err.Error() != want {
		t.Errorf("wrong error message\n got:  %q\n want: %q", err.Error(), want)
	}
}
