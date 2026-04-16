// Package game contains the domain core of the trivia game.
// It has zero imports from infrastructure packages.
package game

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

// GamePort defines the commands the host can issue to a game session.
// Implemented by GameSession; consumed by the handler package.
type GamePort interface {
	Load(quiz QuizFull) error
	StartRound(roundIndex int) error
	RevealQuestion(roundIndex, questionIndex int) error
	ForceEndRound(roundIndex int) error
	MarkAnswerVerdict(teamID string, roundIndex, questionIndex int, verdict Verdict) error
	StartCeremony() error
	AdvanceCeremony(questionIndex int) error
	PublishRoundScores(roundIndex int) error
	EndGame() error
	RegisterTeam(name string) (Team, error)
	SubmitAnswers(teamID string, roundIndex int, answers []Submission) error
	SaveDraft(teamID string, roundIndex, questionIndex int, answer string) error
	ValidateTeamToken(teamID, deviceToken string) bool
}

// StateReader defines the observable state queries consumed by the hub package.
type StateReader interface {
	CurrentState() GameState
	CurrentRoundIndex() int
	TeamRegistry() []Team
	RevealedQuestions() []QuestionPublic
	SubmissionStatus(teamID string) bool
	RoundScores(roundIndex int) map[string]int
	Quiz() QuizPublic
	GetDraft(teamID string, roundIndex, questionIndex int) string
	GetAllDrafts(teamID string) []DraftAnswer
}

// draftKey is the composite key for a draft answer (teamID + round + question).
type draftKey struct {
	teamID        string
	roundIndex    int
	questionIndex int
}

// verdictKey is the composite key for a stored verdict (teamID + round + question).
type verdictKey struct {
	teamID        string
	roundIndex    int
	questionIndex int
}

// GameSession is the in-memory implementation of the game domain.
// It implements both GamePort and StateReader.
type GameSession struct {
	mu sync.RWMutex

	sessionID  string
	state      GameState
	quiz       QuizFull
	quizLoaded bool

	currentRound     int
	revealedUpTo     int // index of last revealed question (-1 = none)
	teams            map[string]Team
	teamOrder        []string
	submissions      map[string][]Submission // teamID -> submissions
	submittedTeams   map[string]bool
	drafts           map[draftKey]string    // teamID+round+question -> draft answer text
	verdicts         map[verdictKey]Verdict // teamID+round+question -> verdict
	roundScoresMap   map[int]*RoundScores
	totals           *TotalScores
	ceremonyQuestion int
	nextTeamSeq      int
}

// newSessionID generates a random 16-character hex session identifier.
func newSessionID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// NewGameSession creates a new GameSession in the LOBBY state.
func NewGameSession() *GameSession {
	return &GameSession{
		sessionID:      newSessionID(),
		state:          StateLobby,
		currentRound:   -1,
		revealedUpTo:   -1,
		teams:          make(map[string]Team),
		submissions:    make(map[string][]Submission),
		submittedTeams: make(map[string]bool),
		drafts:         make(map[draftKey]string),
		verdicts:       make(map[verdictKey]Verdict),
		roundScoresMap: make(map[int]*RoundScores),
		totals:         NewTotalScores(),
	}
}

// GetSessionID returns the unique identifier for this game session.
func (g *GameSession) GetSessionID() string {
	return g.sessionID
}

// -- GamePort implementation ------------------------------------------------

func (g *GameSession) Load(quiz QuizFull) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.quiz = quiz
	g.quizLoaded = true
	return nil
}

func (g *GameSession) StartRound(roundIndex int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.quizLoaded {
		return fmt.Errorf("no quiz loaded")
	}
	if roundIndex < 0 || roundIndex >= len(g.quiz.Rounds) {
		return fmt.Errorf("round index %d out of range", roundIndex)
	}
	if err := g.transition(StateRoundActive); err != nil {
		return err
	}
	g.currentRound = roundIndex
	g.revealedUpTo = -1
	g.roundScoresMap[roundIndex] = NewRoundScores()
	return nil
}

func (g *GameSession) RevealQuestion(roundIndex, questionIndex int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.state != StateRoundActive {
		return fmt.Errorf("cannot reveal question in state %q", g.state)
	}
	if roundIndex != g.currentRound {
		return fmt.Errorf("round index mismatch")
	}
	if questionIndex != g.revealedUpTo+1 {
		return fmt.Errorf("questions must be revealed in order: expected %d, got %d", g.revealedUpTo+1, questionIndex)
	}
	round := g.quiz.Rounds[roundIndex]
	if questionIndex >= len(round.Questions) {
		return fmt.Errorf("question index %d out of range", questionIndex)
	}
	g.revealedUpTo = questionIndex
	return nil
}

func (g *GameSession) ForceEndRound(roundIndex int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if roundIndex != g.currentRound {
		return fmt.Errorf("round index mismatch")
	}
	return g.transition(StateRoundEnded)
}

func (g *GameSession) MarkAnswerVerdict(teamID string, roundIndex, questionIndex int, verdict Verdict) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.state != StateScoring {
		return fmt.Errorf("cannot mark verdict in state %q", g.state)
	}
	rs, ok := g.roundScoresMap[roundIndex]
	if !ok {
		rs = NewRoundScores()
		g.roundScoresMap[roundIndex] = rs
	}
	rs.ApplyVerdict(teamID, verdict)
	g.verdicts[verdictKey{teamID: teamID, roundIndex: roundIndex, questionIndex: questionIndex}] = verdict
	return nil
}

func (g *GameSession) StartCeremony() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.transition(StateCeremony); err != nil {
		return err
	}
	g.ceremonyQuestion = 0
	return nil
}

func (g *GameSession) AdvanceCeremony(questionIndex int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.state != StateCeremony {
		return fmt.Errorf("not in ceremony state")
	}
	g.ceremonyQuestion = questionIndex
	return nil
}

func (g *GameSession) PublishRoundScores(roundIndex int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if rs, ok := g.roundScoresMap[roundIndex]; ok {
		g.totals.AddRound(rs)
	}
	return g.transition(StateRoundScores)
}

func (g *GameSession) EndGame() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.transition(StateGameOver)
}

func (g *GameSession) RegisterTeam(name string) (Team, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	// Reject duplicate names (case-insensitive).
	nameLower := strings.ToLower(name)
	for _, t := range g.teams {
		if strings.ToLower(t.Name) == nameLower {
			return Team{}, fmt.Errorf("That name is taken -- try a different team name")
		}
	}
	g.nextTeamSeq++
	id := fmt.Sprintf("team-%d", g.nextTeamSeq)
	token := fmt.Sprintf("tok-%s-%d", id, g.nextTeamSeq)
	t := Team{ID: id, Name: name, DeviceToken: token}
	g.teams[id] = t
	g.teamOrder = append(g.teamOrder, id)
	return t, nil
}

func (g *GameSession) SubmitAnswers(teamID string, roundIndex int, answers []Submission) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	// Reject unknown team identifiers.
	if _, ok := g.teams[teamID]; !ok {
		return fmt.Errorf("unknown team: %s", teamID)
	}
	// Reject submissions when no round is active or the round has not ended.
	if g.state != StateRoundActive && g.state != StateRoundEnded {
		return fmt.Errorf("no round is currently active")
	}
	// Reject submissions for a round index that does not match the current round.
	if roundIndex != g.currentRound {
		return fmt.Errorf("round %d is not the current round", roundIndex)
	}
	// Already submitted: return a distinct error so the handler can send the right code.
	if g.submittedTeams[teamID] {
		return fmt.Errorf("already_submitted")
	}
	g.submissions[teamID] = answers
	g.submittedTeams[teamID] = true
	return nil
}

// GetSubmissions returns the stored submissions for a team.
// Returns nil if the team has not submitted.
func (g *GameSession) GetSubmissions(teamID string) []Submission {
	g.mu.RLock()
	defer g.mu.RUnlock()
	subs := g.submissions[teamID]
	if subs == nil {
		return nil
	}
	result := make([]Submission, len(subs))
	copy(result, subs)
	return result
}

// SaveDraft persists the player's current draft answer for a question.
// Overwrites any previous draft for the same team/round/question.
func (g *GameSession) SaveDraft(teamID string, roundIndex, questionIndex int, answer string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.drafts[draftKey{teamID: teamID, roundIndex: roundIndex, questionIndex: questionIndex}] = answer
	return nil
}

// -- StateReader implementation ---------------------------------------------

func (g *GameSession) CurrentState() GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.state
}

func (g *GameSession) CurrentRoundIndex() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentRound
}

func (g *GameSession) TeamRegistry() []Team {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]Team, 0, len(g.teamOrder))
	for _, id := range g.teamOrder {
		result = append(result, g.teams[id])
	}
	return result
}

func (g *GameSession) RevealedQuestions() []QuestionPublic {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.currentRound < 0 || !g.quizLoaded {
		return nil
	}
	round := g.quiz.Rounds[g.currentRound]
	var result []QuestionPublic
	for i := 0; i <= g.revealedUpTo && i < len(round.Questions); i++ {
		pub := StripAnswers(round.Questions[i])
		pub.Index = i
		result = append(result, pub)
	}
	return result
}

func (g *GameSession) SubmissionStatus(teamID string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.submittedTeams[teamID]
}

func (g *GameSession) RoundScores(roundIndex int) map[string]int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if rs, ok := g.roundScoresMap[roundIndex]; ok {
		return rs.AllScores()
	}
	return map[string]int{}
}

// TeamRunningTotal returns the sum of a team's scores across all scored rounds so far.
func (g *GameSession) TeamRunningTotal(teamID string) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	total := 0
	for _, rs := range g.roundScoresMap {
		total += rs.TeamScore(teamID)
	}
	return total
}

// GetDraft returns the current draft answer for a team/round/question.
// Returns empty string if no draft has been saved.
func (g *GameSession) GetDraft(teamID string, roundIndex, questionIndex int) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.drafts[draftKey{teamID: teamID, roundIndex: roundIndex, questionIndex: questionIndex}]
}

// GetAllDrafts returns all draft answers saved by the given team.
func (g *GameSession) GetAllDrafts(teamID string) []DraftAnswer {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var result []DraftAnswer
	for k, answer := range g.drafts {
		if k.teamID == teamID {
			result = append(result, DraftAnswer{
				RoundIndex:    k.roundIndex,
				QuestionIndex: k.questionIndex,
				Answer:        answer,
			})
		}
	}
	return result
}

// ValidateTeamToken checks whether the provided deviceToken matches the stored token for teamID.
func (g *GameSession) ValidateTeamToken(teamID, deviceToken string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	t, ok := g.teams[teamID]
	if !ok {
		return false
	}
	return t.DeviceToken == deviceToken
}

func (g *GameSession) Quiz() QuizPublic {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.quizLoaded {
		return QuizPublic{}
	}
	return QuizPublic{
		Title:         g.quiz.Title,
		RoundCount:    len(g.quiz.Rounds),
		QuestionCount: g.totalQuestionCount(),
	}
}

// totalQuestionCount returns the total number of questions across all rounds.
// Must be called with at least a read lock held.
func (g *GameSession) totalQuestionCount() int {
	total := 0
	for _, r := range g.quiz.Rounds {
		total += len(r.Questions)
	}
	return total
}

// transition performs a validated state transition.
// Must be called with the lock held.
func (g *GameSession) transition(to GameState) error {
	if err := ValidateTransition(g.state, to); err != nil {
		return err
	}
	g.state = to
	return nil
}

// BeginScoring transitions from ROUND_ENDED to SCORING.
func (g *GameSession) BeginScoring() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.transition(StateScoring)
}

// CeremonyQuestion returns the QuestionPublic for a ceremony question.
// Safe for the handler package: never exposes QuestionFull or QuizFull.
func (g *GameSession) CeremonyQuestion(roundIndex, questionIndex int) QuestionPublic {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.roundInBounds(roundIndex) {
		return QuestionPublic{}
	}
	round := g.quiz.Rounds[roundIndex]
	if questionIndex < 0 || questionIndex >= len(round.Questions) {
		return QuestionPublic{}
	}
	pub := StripAnswers(round.Questions[questionIndex])
	pub.Index = questionIndex
	return pub
}

// CeremonyAnswer returns the answer string for a question in the current round.
// This is safe to expose from the game package: it returns a plain string,
// never QuestionFull or QuizFull.
func (g *GameSession) CeremonyAnswer(roundIndex, questionIndex int) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.roundInBounds(roundIndex) {
		return ""
	}
	round := g.quiz.Rounds[roundIndex]
	if questionIndex < 0 || questionIndex >= len(round.Questions) {
		return ""
	}
	return round.Questions[questionIndex].Answer
}

// roundInBounds reports whether the quiz is loaded and roundIndex is a valid round index.
// Must be called with at least a read lock held.
func (g *GameSession) roundInBounds(roundIndex int) bool {
	return g.quizLoaded && roundIndex >= 0 && roundIndex < len(g.quiz.Rounds)
}

// RoundName returns the human-readable name of the given round.
// Returns an empty string if the quiz is not loaded or the round index is out of range.
func (g *GameSession) RoundName(roundIndex int) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.roundInBounds(roundIndex) {
		return ""
	}
	return g.quiz.Rounds[roundIndex].Name
}

// RoundQuestionCount returns the number of questions in the given round.
// Returns 0 if the quiz is not loaded or the round index is out of range.
func (g *GameSession) RoundQuestionCount(roundIndex int) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.roundInBounds(roundIndex) {
		return 0
	}
	return len(g.quiz.Rounds[roundIndex].Questions)
}

// ScoringData returns the host-only scoring panel data for a round:
// each question with its correct answer and all team submissions.
func (g *GameSession) ScoringData(roundIndex int) []ScoringQuestion {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.roundInBounds(roundIndex) {
		return nil
	}
	round := g.quiz.Rounds[roundIndex]
	questions := make([]ScoringQuestion, len(round.Questions))
	for questionIndex, q := range round.Questions {
		questions[questionIndex] = ScoringQuestion{
			QuestionIndex: questionIndex,
			Text:          q.Text,
			CorrectAnswer: q.Answer,
			Submissions:   g.collectTeamSubmissions(roundIndex, questionIndex),
		}
	}
	return questions
}

// collectTeamSubmissions builds the ordered list of TeamSubmission entries for one question.
// Must be called with at least a read lock held.
func (g *GameSession) collectTeamSubmissions(roundIndex, questionIndex int) []TeamSubmission {
	result := make([]TeamSubmission, 0, len(g.teamOrder))
	for _, teamID := range g.teamOrder {
		result = append(result, TeamSubmission{
			TeamID:   teamID,
			TeamName: g.teams[teamID].Name,
			Answer:   g.teamAnswerFor(teamID, roundIndex, questionIndex),
		})
	}
	return result
}

// teamAnswerFor returns a team's submitted answer for a specific question, or empty string if not submitted.
// Must be called with at least a read lock held.
func (g *GameSession) teamAnswerFor(teamID string, roundIndex, questionIndex int) string {
	for _, sub := range g.submissions[teamID] {
		if sub.RoundIndex == roundIndex && sub.QuestionIndex == questionIndex {
			return sub.Answer
		}
	}
	return ""
}

// FinalScores returns the cumulative scores across all rounds.
func (g *GameSession) FinalScores() map[string]int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.totals.AllTotals()
}

// TeamVerdict carries one team's verdict for a specific question.
type TeamVerdict struct {
	TeamID   string
	TeamName string
	Verdict  string
}

// VerdictsByQuestion returns the verdict for each registered team for the given
// round and question. Teams that did not submit receive verdict "".
func (g *GameSession) VerdictsByQuestion(roundIndex, questionIndex int) []TeamVerdict {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]TeamVerdict, 0, len(g.teamOrder))
	for _, teamID := range g.teamOrder {
		verdict := g.verdictFor(teamID, roundIndex, questionIndex)
		result = append(result, TeamVerdict{
			TeamID:   teamID,
			TeamName: g.teams[teamID].Name,
			Verdict:  verdict,
		})
	}
	return result
}

// verdictFor returns the verdict string for a team's answer to a specific question.
// Returns empty string if no verdict has been recorded.
// Must be called with at least a read lock held.
func (g *GameSession) verdictFor(teamID string, roundIndex, questionIndex int) string {
	v, ok := g.verdicts[verdictKey{teamID: teamID, roundIndex: roundIndex, questionIndex: questionIndex}]
	if !ok {
		return ""
	}
	return string(v)
}

// ScoreEntry carries one team's round score and cumulative running total.
type ScoreEntry struct {
	TeamID       string
	TeamName     string
	RoundScore   int
	RunningTotal int
}

// RoundScoresWithNames returns round scores for each registered team, enriched
// with team names and running totals.
func (g *GameSession) RoundScoresWithNames(roundIndex int) []ScoreEntry {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]ScoreEntry, 0, len(g.teamOrder))
	rs, hasRound := g.roundScoresMap[roundIndex]
	for _, teamID := range g.teamOrder {
		roundScore := 0
		if hasRound {
			roundScore = rs.TeamScore(teamID)
		}
		result = append(result, ScoreEntry{
			TeamID:       teamID,
			TeamName:     g.teams[teamID].Name,
			RoundScore:   roundScore,
			RunningTotal: g.totals.TeamTotal(teamID),
		})
	}
	return result
}

// FinalScoreEntry carries one team's final cumulative total at game over.
type FinalScoreEntry struct {
	TeamID   string
	TeamName string
	Total    int
}

// FinalScoresWithNames returns the final cumulative scores for each registered team,
// enriched with team names.
func (g *GameSession) FinalScoresWithNames() []FinalScoreEntry {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]FinalScoreEntry, 0, len(g.teamOrder))
	for _, teamID := range g.teamOrder {
		result = append(result, FinalScoreEntry{
			TeamID:   teamID,
			TeamName: g.teams[teamID].Name,
			Total:    g.totals.TeamTotal(teamID),
		})
	}
	return result
}
