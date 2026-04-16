package game

// QuestionFull is the server-internal question representation.
// It contains the answer and must NEVER be sent to clients.
// Use StripAnswers to produce a QuestionPublic for client transport.
type QuestionFull struct {
	Text    string
	Answer  string
	Answers []string
	Choices []string  // multiple choice options (presentation metadata, not answers)
	Media   *MediaRef // optional media attachment
}

// Round is a server-internal round containing its full questions.
type Round struct {
	Name      string
	Questions []QuestionFull
}

// QuizFull is the server-internal quiz representation loaded from YAML.
// It must never appear in handler or hub packages.
type QuizFull struct {
	Title  string
	Rounds []Round
}

// Team represents a participating team in a game session.
type Team struct {
	ID          string
	Name        string
	DeviceToken string
}

// Verdict represents the host's judgment on a submitted answer.
type Verdict string

const (
	VerdictCorrect   Verdict = "correct"
	VerdictIncorrect Verdict = "incorrect"
)

// Submission holds a team's submitted answer for a specific question.
type Submission struct {
	TeamID        string
	RoundIndex    int
	QuestionIndex int
	Answer        string
	Verdict       Verdict
}
