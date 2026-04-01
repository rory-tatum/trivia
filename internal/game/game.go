// Package game contains the domain core of the trivia game.
// It has zero imports from infrastructure packages.
package game

import (
	"fmt"
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
}

// GameSession is the in-memory implementation of the game domain.
// It implements both GamePort and StateReader.
type GameSession struct {
	mu sync.RWMutex

	state      GameState
	quiz       QuizFull
	quizLoaded bool

	currentRound     int
	revealedUpTo     int // index of last revealed question (-1 = none)
	teams            map[string]Team
	teamOrder        []string
	submissions      map[string][]Submission // teamID -> submissions
	submittedTeams   map[string]bool
	roundScoresMap   map[int]*RoundScores
	totals           *TotalScores
	ceremonyQuestion int
	nextTeamSeq      int
}

// NewGameSession creates a new GameSession in the LOBBY state.
func NewGameSession() *GameSession {
	return &GameSession{
		state:          StateLobby,
		currentRound:   -1,
		revealedUpTo:   -1,
		teams:          make(map[string]Team),
		submissions:    make(map[string][]Submission),
		submittedTeams: make(map[string]bool),
		roundScoresMap: make(map[int]*RoundScores),
		totals:         NewTotalScores(),
	}
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
	g.submissions[teamID] = answers
	g.submittedTeams[teamID] = true
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

func (g *GameSession) Quiz() QuizPublic {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.quizLoaded {
		return QuizPublic{}
	}
	total := 0
	for _, r := range g.quiz.Rounds {
		total += len(r.Questions)
	}
	return QuizPublic{
		Title:         g.quiz.Title,
		RoundCount:    len(g.quiz.Rounds),
		QuestionCount: total,
	}
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
