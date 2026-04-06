package game

// Test-support helpers — not used in production logic.
// These functions live here (not in _test.go) so that tests in other packages
// (e.g., handler, hub) can obtain a pre-loaded GameSession without referencing
// QuizFull or QuestionFull directly, keeping those types confined to the game package.

// NewLoadedSession returns a GameSession pre-loaded with a minimal single-round,
// single-question quiz. Use this in handler and hub tests that need a loadedSession
// without importing QuizFull or QuestionFull.
func NewLoadedSession() *GameSession {
	s := NewGameSession()
	_ = s.Load(QuizFull{
		Title: "Test Quiz",
		Rounds: []Round{
			{Name: "Round 1", Questions: []QuestionFull{
				{Text: "Q1?", Answer: "A1"},
			}},
		},
	})
	return s
}

// NewLoadedSessionTwoQuestions returns a GameSession pre-loaded with a single round
// containing two questions. Use this in handler and hub tests that require more than
// one question per round (e.g., late-joiner reveal tests).
func NewLoadedSessionTwoQuestions() *GameSession {
	s := NewGameSession()
	_ = s.Load(QuizFull{
		Title: "Test Quiz",
		Rounds: []Round{
			{Name: "Round 1", Questions: []QuestionFull{
				{Text: "What is the capital of France?", Answer: "Paris"},
				{Text: "What color is the sky?", Answer: "Blue"},
			}},
		},
	})
	return s
}
