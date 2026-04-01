package game

// StripAnswers converts a QuestionFull to a QuestionPublic, discarding all
// answer fields. This is the ONLY sanctioned conversion point (DEC-010).
// The QuestionPublic returned is safe for client transport.
func StripAnswers(q QuestionFull) QuestionPublic {
	return QuestionPublic{
		Text: q.Text,
	}
}
