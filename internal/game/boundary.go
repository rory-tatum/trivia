package game

// StripAnswers converts a QuestionFull to a QuestionPublic, discarding all
// answer fields. This is the ONLY sanctioned conversion point (DEC-010).
// The QuestionPublic returned is safe for client transport.
// Answer and Answers fields are NEVER copied.
func StripAnswers(q QuestionFull) QuestionPublic {
	return QuestionPublic{
		Text:        q.Text,
		Choices:     q.Choices,
		IsMultiPart: len(q.Answers) > 1,
		Media:       q.Media,
	}
}
