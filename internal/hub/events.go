package hub

import (
	"trivia/internal/game"
)

// ServerEvent is the envelope for all server→client JSON events.
// The Event field is the discriminator; Payload carries event-specific data.
type ServerEvent struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

// -- Outbound event payloads -------------------------------------------------

// StateSnapshotPayload carries the full current game state for a newly connecting client.
type StateSnapshotPayload struct {
	State             game.GameState        `json:"state"`
	Quiz              game.QuizPublic       `json:"quiz,omitempty"`
	Teams             []game.Team           `json:"teams"`
	CurrentRound      int                   `json:"current_round"`
	RevealedQuestions []game.QuestionPublic `json:"revealed_questions"`
	DraftAnswers      []game.DraftAnswer    `json:"draft_answers,omitempty"`
}

// TeamJoinedPayload is broadcast when a new team registers.
type TeamJoinedPayload struct {
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
}

// RoundStartedPayload is broadcast when the host starts a round.
type RoundStartedPayload struct {
	RoundIndex int `json:"round_index"`
}

// QuestionRevealedPayload is broadcast when a question is revealed.
// RevealedCount and TotalQuestions are included for the quizmaster panel display.
type QuestionRevealedPayload struct {
	Question       game.QuestionPublic `json:"question"`
	RevealedCount  int                 `json:"revealed_count"`
	TotalQuestions int                 `json:"total_questions"`
}

// SubmissionReceivedPayload is sent to the host when a team submits answers.
type SubmissionReceivedPayload struct {
	TeamID     string `json:"team_id"`
	TeamName   string `json:"team_name"`
	RoundIndex int    `json:"round_index"`
}

// ScoringOpenedPayload is broadcast when the host opens scoring.
type ScoringOpenedPayload struct {
	RoundIndex int `json:"round_index"`
}

// CeremonyQuestionShownPayload is broadcast when a ceremony question is shown.
type CeremonyQuestionShownPayload struct {
	QuestionIndex int                 `json:"question_index"`
	Question      game.QuestionPublic `json:"question"`
}

// CeremonyAnswerRevealedPayload is broadcast when an answer is revealed during ceremony.
type CeremonyAnswerRevealedPayload struct {
	QuestionIndex int    `json:"question_index"`
	Answer        string `json:"answer"`
}

// RoundScoresPayload is broadcast when round scores are published.
type RoundScoresPayload struct {
	RoundIndex int            `json:"round_index"`
	Scores     map[string]int `json:"scores"`
}

// ScoreUpdatedPayload is broadcast to the host when a verdict is marked.
type ScoreUpdatedPayload struct {
	TeamID       string `json:"team_id"`
	RoundIndex   int    `json:"round_index"`
	RunningTotal int    `json:"running_total"`
}

// GameOverPayload is broadcast when the host ends the game.
type GameOverPayload struct {
	FinalScores map[string]int `json:"final_scores"`
}

// ErrorPayload is sent to a single client when an error occurs.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SubmissionAckPayload is sent to the submitting client to confirm receipt.
type SubmissionAckPayload struct {
	TeamID     string `json:"team_id"`
	RoundIndex int    `json:"round_index"`
	Locked     bool   `json:"locked"`
}

// -- Constructor helpers -----------------------------------------------------

// NewStateSnapshotEvent builds a StateSnapshotEvent for a connecting client.
func NewStateSnapshotEvent(p StateSnapshotPayload) ServerEvent {
	return ServerEvent{Event: "state_snapshot", Payload: p}
}

// NewTeamJoinedEvent builds a TeamJoinedEvent.
func NewTeamJoinedEvent(teamID, teamName string) ServerEvent {
	return ServerEvent{Event: "team_joined", Payload: TeamJoinedPayload{TeamID: teamID, TeamName: teamName}}
}

// NewRoundStartedEvent builds a RoundStartedEvent.
func NewRoundStartedEvent(roundIndex int) ServerEvent {
	return ServerEvent{Event: "round_started", Payload: RoundStartedPayload{RoundIndex: roundIndex}}
}

// NewQuestionRevealedEvent builds a QuestionRevealedEvent.
func NewQuestionRevealedEvent(q game.QuestionPublic, revealedCount, totalQuestions int) ServerEvent {
	return ServerEvent{Event: "question_revealed", Payload: QuestionRevealedPayload{
		Question:       q,
		RevealedCount:  revealedCount,
		TotalQuestions: totalQuestions,
	}}
}

// NewSubmissionReceivedEvent builds a SubmissionReceivedEvent.
func NewSubmissionReceivedEvent(teamID, teamName string, roundIndex int) ServerEvent {
	return ServerEvent{Event: "submission_received", Payload: SubmissionReceivedPayload{
		TeamID: teamID, TeamName: teamName, RoundIndex: roundIndex,
	}}
}

// NewScoringOpenedEvent builds a ScoringOpenedEvent.
func NewScoringOpenedEvent(roundIndex int) ServerEvent {
	return ServerEvent{Event: "scoring_opened", Payload: ScoringOpenedPayload{RoundIndex: roundIndex}}
}

// NewScoreUpdatedEvent builds a ScoreUpdatedEvent for the host after a verdict is marked.
func NewScoreUpdatedEvent(teamID string, roundIndex, runningTotal int) ServerEvent {
	return ServerEvent{Event: "score_updated", Payload: ScoreUpdatedPayload{
		TeamID: teamID, RoundIndex: roundIndex, RunningTotal: runningTotal,
	}}
}

// NewCeremonyQuestionShownEvent builds a CeremonyQuestionShownEvent.
func NewCeremonyQuestionShownEvent(questionIndex int, q game.QuestionPublic) ServerEvent {
	return ServerEvent{Event: "ceremony_question_shown", Payload: CeremonyQuestionShownPayload{
		QuestionIndex: questionIndex, Question: q,
	}}
}

// NewCeremonyAnswerRevealedEvent builds a CeremonyAnswerRevealedEvent.
func NewCeremonyAnswerRevealedEvent(questionIndex int, answer string) ServerEvent {
	return ServerEvent{Event: "ceremony_answer_revealed", Payload: CeremonyAnswerRevealedPayload{
		QuestionIndex: questionIndex, Answer: answer,
	}}
}

// NewRoundScoresPublishedEvent builds a RoundScoresPublishedEvent.
func NewRoundScoresPublishedEvent(roundIndex int, scores map[string]int) ServerEvent {
	return ServerEvent{Event: "round_scores_published", Payload: RoundScoresPayload{
		RoundIndex: roundIndex, Scores: scores,
	}}
}

// NewGameOverEvent builds a GameOverEvent.
func NewGameOverEvent(finalScores map[string]int) ServerEvent {
	return ServerEvent{Event: "game_over", Payload: GameOverPayload{FinalScores: finalScores}}
}

// NewErrorEvent builds an ErrorEvent.
func NewErrorEvent(code, message string) ServerEvent {
	return ServerEvent{Event: "error", Payload: ErrorPayload{Code: code, Message: message}}
}

// NewSubmissionAckEvent builds a SubmissionAckEvent.
func NewSubmissionAckEvent(teamID string, roundIndex int, locked bool) ServerEvent {
	return ServerEvent{Event: "submission_ack", Payload: SubmissionAckPayload{
		TeamID: teamID, RoundIndex: roundIndex, Locked: locked,
	}}
}
