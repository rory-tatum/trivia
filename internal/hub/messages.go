package hub

// InboundMessage is the envelope for all client→server JSON messages.
type InboundMessage struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

// -- Play room: client→server messages ---------------------------------------

// TeamRegisterMsg is sent by a new team client to register with a name.
type TeamRegisterMsg struct {
	TeamName string `json:"team_name"`
}

// TeamRejoinMsg is sent by a returning client to rejoin with their existing identity.
type TeamRejoinMsg struct {
	TeamID      string `json:"team_id"`
	DeviceToken string `json:"device_token"`
}

// DraftAnswerMsg is sent when a player drafts (but does not submit) an answer.
type DraftAnswerMsg struct {
	TeamName      string `json:"team_name"`
	RoundIndex    int    `json:"round_index"`
	QuestionIndex int    `json:"question_index"`
	Answer        string `json:"answer"`
}

// SubmitAnswersMsg is sent when a player locks in all answers for a round.
type SubmitAnswersMsg struct {
	TeamName   string        `json:"team_name"`
	RoundIndex int           `json:"round_index"`
	Answers    []AnswerEntry `json:"answers"`
}

// AnswerEntry is one answer within a SubmitAnswersMsg.
type AnswerEntry struct {
	QuestionIndex int    `json:"question_index"`
	Answer        string `json:"answer"`
}

// -- Host room: client→server messages ---------------------------------------

// HostLoadQuizMsg instructs the server to load a quiz from a file path.
type HostLoadQuizMsg struct {
	FilePath string `json:"file_path"`
}

// HostStartRoundMsg instructs the server to start the given round.
type HostStartRoundMsg struct {
	RoundIndex int `json:"round_index"`
}

// HostRevealQuestionMsg instructs the server to reveal a question.
type HostRevealQuestionMsg struct {
	RoundIndex    int `json:"round_index"`
	QuestionIndex int `json:"question_index"`
}

// HostMarkAnswerMsg instructs the server to mark a team's answer with a verdict.
type HostMarkAnswerMsg struct {
	TeamID        string `json:"team_id"`
	RoundIndex    int    `json:"round_index"`
	QuestionIndex int    `json:"question_index"`
	Verdict       string `json:"verdict"`
}

// HostCeremonyShowQuestionMsg instructs the server to show a ceremony question.
type HostCeremonyShowQuestionMsg struct {
	QuestionIndex int `json:"question_index"`
}

// HostCeremonyRevealAnswerMsg instructs the server to reveal the answer for a ceremony question.
type HostCeremonyRevealAnswerMsg struct {
	QuestionIndex int `json:"question_index"`
}

// HostPublishScoresMsg instructs the server to publish round scores.
type HostPublishScoresMsg struct {
	RoundIndex int `json:"round_index"`
}

// HostEndGameMsg instructs the server to end the game.
type HostEndGameMsg struct{}
