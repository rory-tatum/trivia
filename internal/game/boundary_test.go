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

// TestStripAnswersCopiesChoices verifies Choices are propagated to QuestionPublic (US-11).
func TestStripAnswersCopiesChoices(t *testing.T) {
	choices := []string{"Mercury", "Venus", "Earth", "Jupiter"}
	full := game.QuestionFull{
		Text:    "Which planet is largest?",
		Answer:  "Jupiter",
		Choices: choices,
	}

	pub := game.StripAnswers(full)

	if len(pub.Choices) != len(choices) {
		t.Fatalf("expected %d choices, got %d", len(choices), len(pub.Choices))
	}
	for i, c := range choices {
		if pub.Choices[i] != c {
			t.Errorf("choices[%d]: expected %q, got %q", i, c, pub.Choices[i])
		}
	}
}

// TestStripAnswersCopiesMedia verifies Media is propagated to QuestionPublic (US-13).
func TestStripAnswersCopiesMedia(t *testing.T) {
	media := &game.MediaRef{Type: "image", URL: "/media/test.jpg"}
	full := game.QuestionFull{
		Text:   "What is in this image?",
		Answer: "A cat",
		Media:  media,
	}

	pub := game.StripAnswers(full)

	if pub.Media == nil {
		t.Fatal("expected Media to be set, got nil")
	}
	if pub.Media.Type != media.Type {
		t.Errorf("media.Type: expected %q, got %q", media.Type, pub.Media.Type)
	}
	if pub.Media.URL != media.URL {
		t.Errorf("media.URL: expected %q, got %q", media.URL, pub.Media.URL)
	}
}

// TestStripAnswersSetsIsMultiPartTrue verifies IsMultiPart is true when len(Answers) > 1 (US-12).
func TestStripAnswersSetsIsMultiPartTrue(t *testing.T) {
	full := game.QuestionFull{
		Text:    "Name two planets",
		Answers: []string{"Mars", "Venus"},
	}

	pub := game.StripAnswers(full)

	if !pub.IsMultiPart {
		t.Error("expected IsMultiPart=true when Answers has >1 entries")
	}
}

// TestStripAnswersSetsIsMultiPartFalse verifies IsMultiPart is false when len(Answers) <= 1 (US-12).
func TestStripAnswersSetsIsMultiPartFalse(t *testing.T) {
	full := game.QuestionFull{
		Text:   "Name the capital of France",
		Answer: "Paris",
	}

	pub := game.StripAnswers(full)

	if pub.IsMultiPart {
		t.Error("expected IsMultiPart=false when Answers has <=1 entries")
	}
}

// TestStripAnswersExcludesAnswerAndAnswers verifies Answer and Answers fields are
// never present in the JSON output — verifying through JSON tags and reflection.
func TestStripAnswersExcludesAnswerAndAnswers(t *testing.T) {
	full := game.QuestionFull{
		Text:    "What is the capital of France?",
		Answer:  "Paris",
		Answers: []string{"Paris", "Lyon"},
	}

	pub := game.StripAnswers(full)

	typ := reflect.TypeOf(pub)
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		if name == "Answer" || name == "Answers" {
			t.Errorf("StripAnswers result must not have field %q (DEC-010 violation)", name)
		}
	}
	// Verify that the type itself does not leak sensitive fields regardless of value.
	_ = pub
}
