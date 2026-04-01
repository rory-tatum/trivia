// Package quiz handles loading and validation of YAML quiz files.
package quiz

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"trivia/internal/game"
)

// Loader reads and validates quiz YAML files, returning domain types.
type Loader struct{}

// NewLoader creates a new Loader.
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFromPath reads the YAML file at path, validates it, and returns a QuizFull.
// Returns a descriptive error if the file is missing, malformed, or fails validation.
func (l *Loader) LoadFromPath(path string) (game.QuizFull, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return game.QuizFull{}, fmt.Errorf("quiz file not found: %s", path)
		}
		return game.QuizFull{}, fmt.Errorf("reading quiz file %s: %w", path, err)
	}

	var raw yamlQuiz
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return game.QuizFull{}, fmt.Errorf("parsing quiz file %s: %w", path, err)
	}

	if err := validate(raw); err != nil {
		return game.QuizFull{}, err
	}

	return mapToQuizFull(raw), nil
}

// mapToQuizFull converts the raw YAML schema to domain types.
func mapToQuizFull(raw yamlQuiz) game.QuizFull {
	rounds := make([]game.Round, len(raw.Rounds))
	for i, r := range raw.Rounds {
		questions := make([]game.QuestionFull, len(r.Questions))
		for j, q := range r.Questions {
			questions[j] = game.QuestionFull{
				Text:    q.Text,
				Answer:  q.Answer,
				Answers: q.Answers,
			}
		}
		rounds[i] = game.Round{
			Name:      r.Name,
			Questions: questions,
		}
	}
	return game.QuizFull{
		Title:  raw.Title,
		Rounds: rounds,
	}
}
