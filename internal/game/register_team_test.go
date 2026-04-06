package game_test

import (
	"fmt"
	"testing"

	"trivia/internal/game"
)

// Test Budget: 3 behaviors × 2 = 6 max unit tests.
// Behavior 1: RegisterTeam with duplicate name returns exact error message.
// Behavior 2: Duplicate detection is case-insensitive.
// Behavior 3: Sequential registrations produce distinct, monotonically increasing IDs.

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

// TestRegisterTeam_SequentialIDsAreDistinctAndAscending verifies that successive
// RegisterTeam calls assign distinct IDs with strictly ascending sequence numbers.
// This kills the INCREMENT_DECREMENT mutant that changes nextTeamSeq++ to --.
func TestRegisterTeam_SequentialIDsAreDistinctAndAscending(t *testing.T) {
	session := game.NewGameSession()
	t1, err := session.RegisterTeam("Alpha")
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}
	t2, err := session.RegisterTeam("Beta")
	if err != nil {
		t.Fatalf("second registration failed: %v", err)
	}

	if t1.ID == t2.ID {
		t.Errorf("expected distinct team IDs, got %q twice", t1.ID)
	}

	var seq1, seq2 int
	fmt.Sscanf(t1.ID, "team-%d", &seq1)
	fmt.Sscanf(t2.ID, "team-%d", &seq2)
	if seq2 <= seq1 {
		t.Errorf("expected second team sequence (%d) > first (%d)", seq2, seq1)
	}
}
