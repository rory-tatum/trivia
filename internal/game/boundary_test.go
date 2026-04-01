package game_test

import (
	"reflect"
	"testing"

	"trivia/internal/game"
)

// TestQuestionPublicHasNoAnswerFields uses reflection to assert DEC-010:
// QuestionPublic must have NO field named "Answer" or "Answers".
// This is the mandatory safety net for the answer boundary.
func TestQuestionPublicHasNoAnswerFields(t *testing.T) {
	t.Helper()
	typ := reflect.TypeOf(game.QuestionPublic{})
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		if name == "Answer" || name == "Answers" {
			t.Errorf("QuestionPublic must not have field %q (DEC-010 violation)", name)
		}
	}
}

// TestStripAnswersProducesQuestionPublic verifies StripAnswers converts
// a QuestionFull to a QuestionPublic preserving the text.
func TestStripAnswersProducesQuestionPublic(t *testing.T) {
	full := game.QuestionFull{
		Text:    "What is the capital of France?",
		Answer:  "Paris",
		Answers: []string{"Paris"},
	}

	pub := game.StripAnswers(full)

	if pub.Text != full.Text {
		t.Errorf("expected text %q, got %q", full.Text, pub.Text)
	}
}
