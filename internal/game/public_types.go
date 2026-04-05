package game

// QuestionPublic is the transport-safe question representation sent to clients.
// It intentionally has NO Answer or Answers fields (DEC-010).
type QuestionPublic struct {
	Text  string `json:"text"`
	Index int    `json:"index"`
}

// RoundPublic is the transport-safe round representation.
type RoundPublic struct {
	Name      string           `json:"name"`
	Questions []QuestionPublic `json:"questions"`
}

// QuizPublic is the transport-safe quiz summary sent to clients.
type QuizPublic struct {
	Title         string `json:"title"`
	RoundCount    int    `json:"round_count"`
	QuestionCount int    `json:"question_count"`
}

// DraftAnswer is the transport-safe representation of a team's draft answer.
type DraftAnswer struct {
	RoundIndex    int    `json:"round_index"`
	QuestionIndex int    `json:"question_index"`
	Answer        string `json:"answer"`
}
