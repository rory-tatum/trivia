package quiz_test

import (
	"path/filepath"
	"strings"
	"testing"

	"trivia/internal/quiz"
)

// Test Budget: 5 behaviors x 2 = 10 max unit tests
// Behaviors:
//  1. Load valid YAML -> QuizFull with correct data
//  2. Missing title -> descriptive error
//  3. Missing round name -> error identifying round number
//  4. Missing question text -> error identifying round+question
//  5. File not found -> file-not-found error

func TestLoaderLoadFromPath_ValidFile_ReturnsPopulatedQuizFull(t *testing.T) {
	loader := quiz.NewLoader()
	path := filepath.Join("testdata", "valid.yaml")

	result, err := loader.LoadFromPath(path)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Title != "Science Quiz Night" {
		t.Errorf("expected title %q, got %q", "Science Quiz Night", result.Title)
	}
	if len(result.Rounds) != 2 {
		t.Fatalf("expected 2 rounds, got %d", len(result.Rounds))
	}
	if result.Rounds[0].Name != "Round 1: Chemistry" {
		t.Errorf("expected round 1 name %q, got %q", "Round 1: Chemistry", result.Rounds[0].Name)
	}
	if len(result.Rounds[0].Questions) != 2 {
		t.Fatalf("expected 2 questions in round 1, got %d", len(result.Rounds[0].Questions))
	}
	if result.Rounds[0].Questions[0].Text != "What is the chemical symbol for water?" {
		t.Errorf("unexpected question text: %q", result.Rounds[0].Questions[0].Text)
	}
	if result.Rounds[0].Questions[0].Answer != "H2O" {
		t.Errorf("unexpected answer: %q", result.Rounds[0].Questions[0].Answer)
	}
}

func TestLoaderLoadFromPath_MissingTitle_ReturnsError(t *testing.T) {
	loader := quiz.NewLoader()
	path := filepath.Join("testdata", "missing_title.yaml")

	_, err := loader.LoadFromPath(path)

	if err == nil {
		t.Fatal("expected error for missing title, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "title") {
		t.Errorf("expected error to mention 'title', got: %v", err)
	}
}

func TestLoaderLoadFromPath_MissingRoundName_ReturnsErrorWithRoundIndex(t *testing.T) {
	loader := quiz.NewLoader()
	path := filepath.Join("testdata", "missing_round_name.yaml")

	_, err := loader.LoadFromPath(path)

	if err == nil {
		t.Fatal("expected error for missing round name, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "round") {
		t.Errorf("expected error to mention 'round', got: %v", err)
	}
	if !strings.Contains(errMsg, "1") {
		t.Errorf("expected error to identify round number, got: %v", err)
	}
}

func TestLoaderLoadFromPath_MissingQuestionText_ReturnsErrorWithLocation(t *testing.T) {
	loader := quiz.NewLoader()
	path := filepath.Join("testdata", "missing_question_text.yaml")

	_, err := loader.LoadFromPath(path)

	if err == nil {
		t.Fatal("expected error for missing question text, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "question") {
		t.Errorf("expected error to mention 'question', got: %v", err)
	}
	// Error should identify both round 1 and question 1
	if !strings.Contains(errMsg, "1") {
		t.Errorf("expected error to identify location, got: %v", err)
	}
}

func TestLoaderLoadFromPath_NonExistentFile_ReturnsFileNotFoundError(t *testing.T) {
	loader := quiz.NewLoader()

	_, err := loader.LoadFromPath("testdata/does-not-exist.yaml")

	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") &&
		!strings.Contains(strings.ToLower(err.Error()), "no such file") &&
		!strings.Contains(strings.ToLower(err.Error()), "does not exist") {
		t.Errorf("expected file-not-found error, got: %v", err)
	}
}
