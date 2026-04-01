package quiz

import "fmt"

// validate checks the parsed YAML quiz for structural completeness.
// It returns a descriptive error identifying the exact location of any violation.
func validate(q yamlQuiz) error {
	if q.Title == "" {
		return fmt.Errorf("quiz title is required and must not be empty")
	}
	if len(q.Rounds) == 0 {
		return fmt.Errorf("quiz must contain at least one round")
	}
	for i, r := range q.Rounds {
		roundNum := i + 1
		if r.Name == "" {
			return fmt.Errorf("round %d is missing a name", roundNum)
		}
		if len(r.Questions) == 0 {
			return fmt.Errorf("round %d %q must contain at least one question", roundNum, r.Name)
		}
		for j, q := range r.Questions {
			questionNum := j + 1
			if q.Text == "" {
				return fmt.Errorf("round %d, question %d is missing text", roundNum, questionNum)
			}
			if q.Answer == "" && len(q.Answers) == 0 {
				return fmt.Errorf("round %d, question %d must have at least one answer", roundNum, questionNum)
			}
		}
	}
	return nil
}
