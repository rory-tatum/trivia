package game

// QuestionPublic is the transport-safe question representation sent to clients.
// It intentionally has NO Answer or Answers fields (DEC-010).
type QuestionPublic struct {
	Text  string
	Index int
}

// RoundPublic is the transport-safe round representation.
type RoundPublic struct {
	Name      string
	Questions []QuestionPublic
}

// QuizPublic is the transport-safe quiz summary sent to clients.
type QuizPublic struct {
	Title       string
	RoundCount  int
	QuestionCount int
}
