// step_impls.go contains the implementation bodies for all Given/When/Then step methods.
//
// All Given methods set up preconditions by driving the server through its ports.
// All When methods drive a single action through the server's driving port.
// All Then methods assert an observable outcome returned from the server.
//
// RED SCAFFOLD: All step implementations call t.Fatal("not yet implemented — RED scaffold").
// The software-crafter enables one scenario at a time, implementing steps until green.
package steps

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"nhooyr.io/websocket"
)

// gameState mirrors game.GameState string values for black-box assertions.
// Values must match internal/game/state.go constants exactly.
type gameState string

const (
	stateRoundActive gameState = "ROUND_ACTIVE"
	stateCeremony    gameState = "CEREMONY"
	stateRoundScores gameState = "ROUND_SCORES"
)

// Event type names sent by the server over WebSocket.
const (
	eventTeamRegistered        = "team_registered"
	eventStateSnapshot         = "state_snapshot"
	eventRoundStarted          = "round_started"
	eventQuestionRevealed      = "question_revealed"
	eventRoundEnded            = "round_ended"
	eventSubmissionAck         = "submission_ack"
	eventSubmissionReceived    = "submission_received"
	eventCeremonyQuestionShown = "ceremony_question_shown"
	eventCeremonyAnswerReveal  = "ceremony_answer_revealed"
	eventRoundScoresPublished  = "round_scores_published"
	eventGameOver              = "game_over"
	eventError                 = "error"
)

// eventWaitTimeout is the default deadline used by waitForEvent calls in Then steps.
const eventWaitTimeout = 2 * time.Second

// negativeEventWindow is the brief wait used when asserting that an event must NOT arrive.
const negativeEventWindow = 150 * time.Millisecond

// defaultQuestionCount is the number of questions used in focused test fixtures
// that do not specify an explicit question count.
const defaultQuestionCount = 2

// =============================================================================
// Internal helpers
// =============================================================================

// dialPlay opens a raw WebSocket connection to the play room without registering it in the world.
func dialPlay(ctx context.Context, server *httptest.Server) (*websocket.Conn, error) {
	url := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws?room=play"
	conn, _, err := websocket.Dial(ctx, url, nil)
	return conn, err
}

// =============================================================================
// Given implementations — arrange preconditions
// =============================================================================

func (w *World) givenServerRunning(token string) error {
	if w.server != nil {
		return nil // already running
	}
	w.hostToken = token
	w.server = NewPlayUITestServer(token)
	return nil
}

func (w *World) givenQuizFileExists(filename string, rounds, questions int) error {
	qs := make([]QuizQuestion, questions)
	for i := 0; i < questions; i++ {
		qs[i] = QuizQuestion{
			Text:   fmt.Sprintf("Question %d?", i+1),
			Answer: fmt.Sprintf("Answer %d", i+1),
		}
	}
	w.quizFixtures[filename] = SimpleQuizYAML(TitleFromFilename(filename), qs)
	return nil
}

func (w *World) givenQuizFileExistsMultiRound(filename string, rounds, questionsPerRound int) error {
	w.quizFixtures[filename] = MultiRoundQuizYAML(TitleFromFilename(filename), rounds, questionsPerRound)
	return nil
}

func (w *World) givenQuizFileWithMultipleChoice(filename string) error {
	w.quizFixtures[filename] = MultipleChoiceQuizYAML(TitleFromFilename(filename))
	return nil
}

func (w *World) givenQuizFileWithMultiPart(filename string) error {
	w.quizFixtures[filename] = MultiPartQuizYAML(TitleFromFilename(filename))
	return nil
}

func (w *World) givenQuizFileWithMedia(filename string) error {
	w.quizFixtures[filename] = MediaQuizYAML(TitleFromFilename(filename))
	return nil
}

func (w *World) givenQuizmasterLoadedQuiz(filename string) error {
	// Ensure server is running.
	if w.server == nil {
		return fmt.Errorf("server not started — call givenServerRunning first")
	}
	// Get or create host driver.
	d := w.ensureHostDriver()
	// Write the quiz fixture to disk.
	content, ok := w.quizFixtures[filename]
	if !ok {
		return fmt.Errorf("quiz fixture %q not registered — call givenQuizFileExists first", filename)
	}
	path, err := d.WriteQuizFixture(filename, content)
	if err != nil {
		return fmt.Errorf("write quiz fixture: %w", err)
	}
	w.quizFilePaths[filename] = path
	// Connect host if not already connected.
	if err := w.ensureHostConnected(d); err != nil {
		return fmt.Errorf("host connect: %w", err)
	}
	// Send host_load_quiz.
	if err := d.HostLoadQuiz(w.ctx, path); err != nil {
		return fmt.Errorf("host_load_quiz: %w", err)
	}
	// Wait for quiz_loaded confirmation.
	_, ok = w.waitForEvent(roleHost, "quiz_loaded", eventWaitTimeout)
	if !ok {
		return fmt.Errorf("quiz_loaded event not received — quiz load may have failed")
	}
	return nil
}

func (w *World) givenRoundStartedWithTeam(roundIndex int, teamName string) error {
	// Load the first registered quiz fixture into the server.
	if err := w.loadFirstQuizFixture(); err != nil {
		return err
	}
	// Connect and register the team in the play room.
	d := w.ensurePlayDriver(teamName)
	if err := w.ensurePlayConnected(d, teamName); err != nil {
		return err
	}
	if err := d.PlayRegisterTeam(w.ctx, teamName); err != nil {
		return err
	}
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventTeamRegistered, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not register in givenRoundStartedWithTeam", teamName)
	}
	// Start the round via the host port.
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostStartRound(w.ctx, roundIndex); err != nil {
		return err
	}
	if _, ok := w.waitForEvent(key, eventRoundStarted, eventWaitTimeout); !ok {
		return fmt.Errorf("round_started not received by team %q in givenRoundStartedWithTeam", teamName)
	}
	return nil
}

// loadFirstQuizFixture loads the first registered quiz fixture into the server.
// It is a no-op if no fixtures are registered or the quiz is already loaded.
func (w *World) loadFirstQuizFixture() error {
	if len(w.quizFilePaths) > 0 {
		return nil // already loaded
	}
	if len(w.quizFixtures) == 0 {
		return fmt.Errorf("no quiz fixture registered — call givenQuizFileExists first")
	}
	// Pick the first (and typically only) fixture.
	var filename string
	for f := range w.quizFixtures {
		filename = f
		break
	}
	return w.givenQuizmasterLoadedQuiz(filename)
}

func (w *World) givenRoundStartedAndQuestionsRevealed(roundIndex, questionCount int) error {
	// Load the first registered quiz fixture into the server.
	if err := w.loadFirstQuizFixture(); err != nil {
		return err
	}
	// Start the round via the host port (no team in play room yet).
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostStartRound(w.ctx, roundIndex); err != nil {
		return err
	}
	// Reveal the required number of questions.
	for i := 0; i < questionCount; i++ {
		if err := hd.HostRevealQuestion(w.ctx, roundIndex, i); err != nil {
			return fmt.Errorf("reveal question %d: %w", i, err)
		}
		// Brief pause to allow server to process each reveal in sequence.
		time.Sleep(10 * time.Millisecond)
	}
	w.revealedCount = questionCount
	return nil
}

func (w *World) givenRoundEndedWithTeam(roundIndex int, teamName string) error {
	// Start the round with the team registered, then end it.
	if err := w.givenRoundStartedWithTeam(roundIndex, teamName); err != nil {
		return err
	}
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostEndRound(w.ctx, roundIndex); err != nil {
		return err
	}
	// Wait for the team to receive round_ended confirmation.
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventRoundEnded, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive round_ended in givenRoundEndedWithTeam", teamName)
	}
	return nil
}

func (w *World) givenRoundEnded(roundIndex int) error {
	// Load the first registered quiz fixture if not already loaded.
	if err := w.loadFirstQuizFixture(); err != nil {
		return err
	}
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostStartRound(w.ctx, roundIndex); err != nil {
		return err
	}
	time.Sleep(20 * time.Millisecond)
	return hd.HostEndRound(w.ctx, roundIndex)
}

func (w *World) givenQuizmasterRevealedQuestion(roundIndex, questionIndex int) error {
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	return hd.HostRevealQuestion(w.ctx, roundIndex, questionIndex)
}

func (w *World) givenAllQuestionsRevealed(count int) error {
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	roundIndex := w.currentRoundIndex
	if roundIndex < 0 {
		roundIndex = 0
	}
	for i := 0; i < count; i++ {
		if err := hd.HostRevealQuestion(w.ctx, roundIndex, i); err != nil {
			return fmt.Errorf("reveal question %d: %w", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	w.revealedCount = count
	return nil
}

func (w *World) givenTeamAlreadyRegistered(teamName string) error {
	// Precondition: connect and register the team so the name is taken.
	if w.server == nil {
		return fmt.Errorf("server not started")
	}
	d := w.ensurePlayDriver(teamName)
	if err := w.ensurePlayConnected(d, teamName); err != nil {
		return err
	}
	if err := d.PlayRegisterTeam(w.ctx, teamName); err != nil {
		return err
	}
	// Wait for team_registered to confirm the name is now taken.
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventTeamRegistered, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not register successfully in givenTeamAlreadyRegistered", teamName)
	}
	return nil
}

func (w *World) givenTeamRegistered(teamName string) error {
	if w.server == nil {
		return fmt.Errorf("server not started — call givenServerRunning first")
	}
	d := w.ensurePlayDriver(teamName)
	if err := w.ensurePlayConnected(d, teamName); err != nil {
		return err
	}
	if err := d.PlayRegisterTeam(w.ctx, teamName); err != nil {
		return err
	}
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventTeamRegistered, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive team_registered in givenTeamRegistered", teamName)
	}
	return nil
}

func (w *World) givenTwoTeamsRegistered(team1, team2 string) error {
	for _, teamName := range []string{team1, team2} {
		d := w.ensurePlayDriver(teamName)
		if err := w.ensurePlayConnected(d, teamName); err != nil {
			return err
		}
		if err := d.PlayRegisterTeam(w.ctx, teamName); err != nil {
			return err
		}
		key := connectionKey(rolePlay, teamName)
		if _, ok := w.waitForEvent(key, eventTeamRegistered, eventWaitTimeout); !ok {
			return fmt.Errorf("team %q did not register in givenTwoTeamsRegistered", teamName)
		}
	}
	return nil
}

func (w *World) givenTeamRegisteredAndRoundActiveWithQuestions(teamName string, roundIndex, questionCount int) error {
	// Ensure quiz is loaded.
	if err := w.loadFirstQuizFixture(); err != nil {
		return err
	}
	// Connect and register the team in the play room.
	d := w.ensurePlayDriver(teamName)
	if err := w.ensurePlayConnected(d, teamName); err != nil {
		return err
	}
	if err := d.PlayRegisterTeam(w.ctx, teamName); err != nil {
		return err
	}
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventTeamRegistered, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not register in givenTeamRegisteredAndRoundActiveWithQuestions", teamName)
	}
	// Start the round via the host port.
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostStartRound(w.ctx, roundIndex); err != nil {
		return err
	}
	if _, ok := w.waitForEvent(key, eventRoundStarted, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive round_started", teamName)
	}
	// Reveal the requested number of questions.
	for i := 0; i < questionCount; i++ {
		if err := hd.HostRevealQuestion(w.ctx, roundIndex, i); err != nil {
			return fmt.Errorf("reveal question %d: %w", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !w.waitForEventCount(key, eventQuestionRevealed, questionCount, eventWaitTimeout) {
		return fmt.Errorf("team %q did not receive all %d question_revealed events", teamName, questionCount)
	}
	return nil
}

func (w *World) givenTeamSavedDraft(teamName string, roundIndex, questionIndex int, answer string) error {
	d := w.ensurePlayDriver(teamName)
	if err := d.PlayDraftAnswer(w.ctx, teamName, roundIndex, questionIndex, answer); err != nil {
		return fmt.Errorf("save draft for %q round %d question %d: %w", teamName, roundIndex, questionIndex, err)
	}
	// Brief pause to allow the server to persist the draft before the next step.
	time.Sleep(20 * time.Millisecond)
	return nil
}

func (w *World) givenTeamSubmitted(teamName string, roundIndex int) error {
	// The team must be registered and the round must be ended before they can submit.
	// Precondition: givenRoundEndedWithTeam has already been called.
	d := w.ensurePlayDriver(teamName)
	answers := make([]map[string]interface{}, w.totalQuestions)
	for i := 0; i < w.totalQuestions; i++ {
		answers[i] = map[string]interface{}{
			"question_index": i,
			"answer":         "",
		}
	}
	if err := d.PlaySubmitAnswers(w.ctx, teamName, roundIndex, answers); err != nil {
		return err
	}
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventSubmissionAck, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive submission_ack in givenTeamSubmitted", teamName)
	}
	return nil
}

func (w *World) givenTeamSubmittedAndCeremonyStarted(teamName string, roundIndex int) error {
	// Set up: register team, end round, team submits, begin scoring to transition to SCORING state.
	if err := w.givenRoundEndedWithTeam(roundIndex, teamName); err != nil {
		return err
	}
	if err := w.givenTeamSubmitted(teamName, roundIndex); err != nil {
		return err
	}
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostBeginScoring(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("begin scoring: %w", err)
	}
	time.Sleep(20 * time.Millisecond)
	return nil
}

func (w *World) givenTeamSubmittedAndCeremonyAtQuestion(teamName string, roundIndex, questionIndex int) error {
	// Set up: register team, end round, team submits, then advance ceremony to the given question.
	if err := w.givenRoundEndedWithTeam(roundIndex, teamName); err != nil {
		return err
	}
	if err := w.givenTeamSubmitted(teamName, roundIndex); err != nil {
		return err
	}
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostBeginScoring(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("begin scoring: %w", err)
	}
	time.Sleep(20 * time.Millisecond)
	// Show each question up to and including questionIndex.
	for i := 0; i <= questionIndex; i++ {
		if err := hd.HostCeremonyShowQuestion(w.ctx, i); err != nil {
			return fmt.Errorf("show ceremony question %d: %w", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (w *World) givenTeamHasValidToken(teamName string) error {
	// Ensure the team is registered so its device_token is stored in the World.
	// If already registered (teamID present), this is a no-op.
	if w.teamID(teamName) != "" {
		return nil
	}
	return w.givenTeamRegistered(teamName)
}

func (w *World) givenTwoTeamsCompletedRoundWithScoring(team1, team2 string, roundIndex int) error {
	// Load quiz fixture if not already loaded.
	if err := w.loadFirstQuizFixture(); err != nil {
		return err
	}
	// Register both teams in the play room.
	for _, teamName := range []string{team1, team2} {
		d := w.ensurePlayDriver(teamName)
		if err := w.ensurePlayConnected(d, teamName); err != nil {
			return fmt.Errorf("connect team %q: %w", teamName, err)
		}
		if err := d.PlayRegisterTeam(w.ctx, teamName); err != nil {
			return fmt.Errorf("register team %q: %w", teamName, err)
		}
		key := connectionKey(rolePlay, teamName)
		if _, ok := w.waitForEvent(key, eventTeamRegistered, eventWaitTimeout); !ok {
			return fmt.Errorf("team %q did not register", teamName)
		}
	}
	// Start round via host.
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostStartRound(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("start round: %w", err)
	}
	// Wait for both teams to receive round_started.
	for _, teamName := range []string{team1, team2} {
		key := connectionKey(rolePlay, teamName)
		if _, ok := w.waitForEvent(key, eventRoundStarted, eventWaitTimeout); !ok {
			return fmt.Errorf("team %q did not receive round_started", teamName)
		}
	}
	// End round.
	if err := hd.HostEndRound(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("end round: %w", err)
	}
	for _, teamName := range []string{team1, team2} {
		key := connectionKey(rolePlay, teamName)
		if _, ok := w.waitForEvent(key, eventRoundEnded, eventWaitTimeout); !ok {
			return fmt.Errorf("team %q did not receive round_ended", teamName)
		}
	}
	// Both teams submit.
	for _, teamName := range []string{team1, team2} {
		d := w.ensurePlayDriver(teamName)
		answers := make([]map[string]interface{}, w.totalQuestions)
		for i := 0; i < w.totalQuestions; i++ {
			answers[i] = map[string]interface{}{"question_index": i, "answer": ""}
		}
		if err := d.PlaySubmitAnswers(w.ctx, teamName, roundIndex, answers); err != nil {
			return fmt.Errorf("submit team %q: %w", teamName, err)
		}
		key := connectionKey(rolePlay, teamName)
		if _, ok := w.waitForEvent(key, eventSubmissionAck, eventWaitTimeout); !ok {
			return fmt.Errorf("team %q did not receive submission_ack", teamName)
		}
	}
	// Begin scoring to enter SCORING/CEREMONY state.
	if err := hd.HostBeginScoring(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("begin scoring: %w", err)
	}
	time.Sleep(20 * time.Millisecond)
	// Run ceremony: show and reveal each question so server reaches SCORES_PUBLISHED state.
	for i := 0; i < w.totalQuestions; i++ {
		if err := hd.HostCeremonyShowQuestion(w.ctx, i); err != nil {
			return fmt.Errorf("ceremony show question %d: %w", i, err)
		}
		time.Sleep(10 * time.Millisecond)
		if err := hd.HostCeremonyRevealAnswer(w.ctx, i); err != nil {
			return fmt.Errorf("ceremony reveal answer %d: %w", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (w *World) givenTeamCompletedRoundWithScoring(teamName string, roundIndex int) error {
	// Single-team variant used by ADP-02.
	if err := w.loadFirstQuizFixture(); err != nil {
		return err
	}
	d := w.ensurePlayDriver(teamName)
	if err := w.ensurePlayConnected(d, teamName); err != nil {
		return fmt.Errorf("connect team %q: %w", teamName, err)
	}
	if err := d.PlayRegisterTeam(w.ctx, teamName); err != nil {
		return fmt.Errorf("register team %q: %w", teamName, err)
	}
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventTeamRegistered, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not register", teamName)
	}
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostStartRound(w.ctx, roundIndex); err != nil {
		return err
	}
	if _, ok := w.waitForEvent(key, eventRoundStarted, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive round_started", teamName)
	}
	if err := hd.HostEndRound(w.ctx, roundIndex); err != nil {
		return err
	}
	if _, ok := w.waitForEvent(key, eventRoundEnded, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive round_ended", teamName)
	}
	answers := make([]map[string]interface{}, w.totalQuestions)
	for i := 0; i < w.totalQuestions; i++ {
		answers[i] = map[string]interface{}{"question_index": i, "answer": ""}
	}
	if err := d.PlaySubmitAnswers(w.ctx, teamName, roundIndex, answers); err != nil {
		return err
	}
	if _, ok := w.waitForEvent(key, eventSubmissionAck, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive submission_ack", teamName)
	}
	if err := hd.HostBeginScoring(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("begin scoring: %w", err)
	}
	time.Sleep(20 * time.Millisecond)
	for i := 0; i < w.totalQuestions; i++ {
		if err := hd.HostCeremonyShowQuestion(w.ctx, i); err != nil {
			return fmt.Errorf("ceremony show question %d: %w", i, err)
		}
		time.Sleep(10 * time.Millisecond)
		if err := hd.HostCeremonyRevealAnswer(w.ctx, i); err != nil {
			return fmt.Errorf("ceremony reveal answer %d: %w", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (w *World) givenTwoTeamsRoundCompleteAndScoresPublished(team1, team2 string, roundIndex int) error {
	// Complete ceremony for two teams and then publish scores.
	if err := w.givenTwoTeamsCompletedRoundWithScoring(team1, team2, roundIndex); err != nil {
		return err
	}
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostPublishScores(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("publish scores: %w", err)
	}
	// Wait for team1 to receive round_scores_published.
	key := connectionKey(rolePlay, team1)
	if _, ok := w.waitForEvent(key, eventRoundScoresPublished, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive round_scores_published in setup", team1)
	}
	return nil
}

func (w *World) givenTeamOnScoresScreen(teamName string, roundIndex int) error {
	// Single-team: complete round, ceremony, publish scores — team is now on scores screen.
	if err := w.givenTeamCompletedRoundWithScoring(teamName, roundIndex); err != nil {
		return err
	}
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostPublishScores(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("publish scores: %w", err)
	}
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventRoundScoresPublished, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive round_scores_published in givenTeamOnScoresScreen", teamName)
	}
	return nil
}

// ensureDefaultQuizFixture registers and loads a simple default quiz fixture
// if no quiz fixtures are already registered. Used by Given steps that don't
// have an explicit quiz file step.
func (w *World) ensureDefaultQuizFixture() error {
	if len(w.quizFixtures) > 0 {
		return nil // already registered
	}
	const defaultFilename = "default-setup.yaml"
	if err := w.givenQuizFileExists(defaultFilename, 1, defaultQuestionCount); err != nil {
		return err
	}
	return w.givenQuizmasterLoadedQuiz(defaultFilename)
}

func (w *World) givenGameInCeremonyPhase(roundIndex int) error {
	// We need at least one registered team to drive through the full flow.
	// Use a temporary team name that doesn't interfere with the scenario's main team.
	const setupTeam = "Setup Team"
	if err := w.ensureDefaultQuizFixture(); err != nil {
		return err
	}
	if err := w.givenRoundEndedWithTeam(roundIndex, setupTeam); err != nil {
		return err
	}
	if err := w.givenTeamSubmitted(setupTeam, roundIndex); err != nil {
		return err
	}
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostBeginScoring(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("begin scoring: %w", err)
	}
	// Wait briefly for the state transition to CEREMONY to be applied server-side.
	time.Sleep(30 * time.Millisecond)
	return nil
}

func (w *World) givenGameAtScoresScreen(roundIndex int) error {
	// Drive the game through: round → end → submit → ceremony → publish scores.
	const setupTeam = "Setup Team"
	if err := w.ensureDefaultQuizFixture(); err != nil {
		return err
	}
	if err := w.givenTeamCompletedRoundWithScoring(setupTeam, roundIndex); err != nil {
		return err
	}
	hd := w.ensureHostDriver()
	if err := w.ensureHostConnected(hd); err != nil {
		return err
	}
	if err := hd.HostPublishScores(w.ctx, roundIndex); err != nil {
		return fmt.Errorf("publish scores: %w", err)
	}
	key := connectionKey(rolePlay, setupTeam)
	if _, ok := w.waitForEvent(key, eventRoundScoresPublished, eventWaitTimeout); !ok {
		return fmt.Errorf("setup team did not receive round_scores_published")
	}
	return nil
}

func (w *World) givenTeamRegisteredRoundActiveQuestionsRevealed(teamName string, roundIndex, questionCount int) error {
	return w.givenTeamRegisteredAndRoundActiveWithQuestions(teamName, roundIndex, questionCount)
}

// =============================================================================
// When implementations — drive actions
// =============================================================================

func (w *World) whenTeamConnects(teamName string) error {
	if w.server == nil {
		return fmt.Errorf("server not started")
	}
	d := w.ensurePlayDriver(teamName)
	return w.ensurePlayConnected(d, teamName)
}

func (w *World) whenTeamRegisters(teamName string) error {
	d := w.ensurePlayDriver(teamName)
	return d.PlayRegisterTeam(w.ctx, teamName)
}

func (w *World) whenTeamConnectsAndRegisters(teamName string) error {
	d := w.ensurePlayDriver(teamName)
	if err := w.ensurePlayConnected(d, teamName); err != nil {
		return err
	}
	if err := d.PlayRegisterTeam(w.ctx, teamName); err != nil {
		return err
	}
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventTeamRegistered, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive team_registered in whenTeamConnectsAndRegisters", teamName)
	}
	return nil
}

func (w *World) whenTeamAttemptsRegistrationFromSecondDevice(teamName string) error {
	// Simulate a second physical device: open a fresh play connection under a distinct key.
	const secondDeviceSuffix = ":device2"
	secondKey := connectionKey(rolePlay, teamName) + secondDeviceSuffix
	d2 := NewPlayUIDriver(w.server, w.hostToken, w)
	w.connections[secondKey] = &WSConnection{
		Role:      rolePlay,
		Name:      teamName + secondDeviceSuffix,
		Connected: false,
		driver:    d2,
	}
	// Connect using the play endpoint.
	conn, err := dialPlay(w.ctx, w.server)
	if err != nil {
		return fmt.Errorf("second device connect: %w", err)
	}
	d2.wsConns[secondKey] = conn
	go d2.readLoop(w.ctx, secondKey, conn)
	w.connections[secondKey].Connected = true
	// Attempt to register with the already-taken name.
	return d2.sendMessage(w.ctx, secondKey, map[string]interface{}{
		"event": "team_register",
		"payload": map[string]interface{}{
			"team_name": teamName,
		},
	})
}

func (w *World) whenPlayerAttemptsEmptyRegistration() error {
	// An anonymous player connects and sends team_register with an empty team_name.
	if w.server == nil {
		return fmt.Errorf("server not started")
	}
	const anonKey = "play:anonymous"
	conn, err := dialPlay(w.ctx, w.server)
	if err != nil {
		return fmt.Errorf("anonymous player connect: %w", err)
	}
	anonDriver := NewPlayUIDriver(w.server, w.hostToken, w)
	anonDriver.wsConns[anonKey] = conn
	go anonDriver.readLoop(w.ctx, anonKey, conn)
	w.connections[anonKey] = &WSConnection{
		Role:      rolePlay,
		Name:      "anonymous",
		Connected: true,
		driver:    anonDriver,
	}
	// Send team_register with an empty team_name.
	return anonDriver.sendMessage(w.ctx, anonKey, map[string]interface{}{
		"event": "team_register",
		"payload": map[string]interface{}{
			"team_name": "",
		},
	})
}

func (w *World) whenTeamReconnectsWithStoredToken(teamName string) error {
	// Retrieve stored credentials from the previous connection.
	teamID := w.teamID(teamName)
	token := w.deviceToken(teamName)
	if teamID == "" || token == "" {
		return fmt.Errorf("no stored credentials for team %q — ensure team registered before reconnect", teamName)
	}

	key := connectionKey(rolePlay, teamName)

	// Close the existing connection if present.
	if conn, ok := w.connections[key]; ok && conn.driver != nil {
		conn.driver.CloseConnection(rolePlay, teamName)
		conn.Connected = false
	}

	// Clear old messages so Then steps only see messages from the new connection.
	w.mu.Lock()
	delete(w.receivedMessages, key)
	w.mu.Unlock()

	// Open a fresh connection under the same key.
	d := w.ensurePlayDriver(teamName)
	if err := d.ConnectPlay(w.ctx, teamName); err != nil {
		return fmt.Errorf("reconnect for %q: %w", teamName, err)
	}
	if conn, ok := w.connections[key]; ok {
		conn.Connected = true
	}

	// Wait briefly for the initial state_snapshot on the new connection (sent on connect).
	time.Sleep(20 * time.Millisecond)

	// Send team_rejoin with stored credentials.
	return d.PlayRejoinTeam(w.ctx, teamName, teamID, token)
}

func (w *World) whenPlayerAttemptsBadRejoin() error {
	// Open an anonymous connection and send team_rejoin with a bad token.
	const anonKey = "play:anonymous"
	if w.server == nil {
		return fmt.Errorf("server not started")
	}
	conn, err := dialPlay(w.ctx, w.server)
	if err != nil {
		return fmt.Errorf("anonymous rejoin connect: %w", err)
	}
	anonDriver := NewPlayUIDriver(w.server, w.hostToken, w)
	anonDriver.wsConns[anonKey] = conn
	go anonDriver.readLoop(w.ctx, anonKey, conn)
	w.connections[anonKey] = &WSConnection{
		Role:      rolePlay,
		Name:      "anonymous",
		Connected: true,
		driver:    anonDriver,
	}
	return anonDriver.PlayRejoinTeamWithBadToken(w.ctx, "anonymous")
}

func (w *World) whenTeamRequestsSnapshot(teamName string) error {
	// state_snapshot is sent automatically on connection; this step is a no-op
	// that makes the scenario read naturally.
	return nil
}

func (w *World) whenTeamSavesDraft(teamName string, roundIndex, questionIndex int, answer string) error {
	d := w.ensurePlayDriver(teamName)
	return d.PlayDraftAnswer(w.ctx, teamName, roundIndex, questionIndex, answer)
}

func (w *World) whenTeamSubmitsAnswers(teamName string, roundIndex int) error {
	d := w.ensurePlayDriver(teamName)
	// Build answers from drafts saved by this team.
	answers := make([]map[string]interface{}, w.totalQuestions)
	for i := 0; i < w.totalQuestions; i++ {
		answers[i] = map[string]interface{}{
			"question_index": i,
			"answer":         "",
		}
	}
	return d.PlaySubmitAnswers(w.ctx, teamName, roundIndex, answers)
}

func (w *World) whenTeamSubmitsBlankAnswers(teamName string, roundIndex int) error {
	d := w.ensurePlayDriver(teamName)
	// Build answers with empty string values.
	count := w.totalQuestions
	if count == 0 {
		count = defaultQuestionCount
	}
	answers := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		answers[i] = map[string]interface{}{
			"question_index": i,
			"answer":         "",
		}
	}
	return d.PlaySubmitAnswers(w.ctx, teamName, roundIndex, answers)
}

func (w *World) whenTeamAttemptsSubmitBeforeRound(teamName string, roundIndex int) error {
	d := w.ensurePlayDriver(teamName)
	answers := []map[string]interface{}{
		{"question_index": 0, "answer": "some answer"},
	}
	return d.PlaySubmitAnswers(w.ctx, teamName, roundIndex, answers)
}

func (w *World) whenPlayerAttemptsSubmitWithUnknownID() error {
	// Use a fresh anonymous connection with a fake team_id.
	// Reuse the "play:anonymous" key so thenAnonymousPlayerReceivesError can find the event.
	if w.server == nil {
		return fmt.Errorf("server not started")
	}
	const anonKey = "play:anonymous"
	conn, err := dialPlay(w.ctx, w.server)
	if err != nil {
		return fmt.Errorf("unknown team connect: %w", err)
	}
	anonDriver := NewPlayUIDriver(w.server, w.hostToken, w)
	anonDriver.wsConns[anonKey] = conn
	go anonDriver.readLoop(w.ctx, anonKey, conn)
	w.connections[anonKey] = &WSConnection{
		Role:      rolePlay,
		Name:      "anonymous",
		Connected: true,
		driver:    anonDriver,
	}
	return anonDriver.PlaySubmitAnswersWithID(w.ctx, "anonymous",
		"00000000-0000-0000-0000-000000000000", 0,
		[]map[string]interface{}{{"question_index": 0, "answer": "x"}},
	)
}

func (w *World) whenQuizmasterStartsRound(roundIndex int) error {
	d := w.ensureHostDriver()
	if err := w.ensureHostConnected(d); err != nil {
		return err
	}
	return d.HostStartRound(w.ctx, roundIndex)
}

func (w *World) whenQuizmasterRevealsQuestion(roundIndex, questionIndex int) error {
	d := w.ensureHostDriver()
	if err := w.ensureHostConnected(d); err != nil {
		return err
	}
	return d.HostRevealQuestion(w.ctx, roundIndex, questionIndex)
}

func (w *World) whenQuizmasterEndsRound(roundIndex int) error {
	d := w.ensureHostDriver()
	if err := w.ensureHostConnected(d); err != nil {
		return err
	}
	return d.HostEndRound(w.ctx, roundIndex)
}

func (w *World) whenQuizmasterShowsCeremonyQuestion(questionIndex int) error {
	d := w.ensureHostDriver()
	if err := w.ensureHostConnected(d); err != nil {
		return err
	}
	// Ceremony requires SCORING state. For the first ceremony question (index 0),
	// begin scoring first to transition through SCORING → CEREMONY.
	if questionIndex == 0 {
		if err := d.HostBeginScoring(w.ctx, w.currentRoundIndex); err != nil {
			return fmt.Errorf("begin scoring before ceremony: %w", err)
		}
		// Wait briefly for the scoring state to be applied server-side.
		time.Sleep(20 * time.Millisecond)
	}
	return d.HostCeremonyShowQuestion(w.ctx, questionIndex)
}

func (w *World) whenQuizmasterRevealsCeremonyAnswer(questionIndex int) error {
	d := w.ensureHostDriver()
	if err := w.ensureHostConnected(d); err != nil {
		return err
	}
	return d.HostCeremonyRevealAnswer(w.ctx, questionIndex)
}

func (w *World) whenQuizmasterShowsAndRevealsCeremonyQuestion(questionIndex int) error {
	d := w.ensureHostDriver()
	if err := w.ensureHostConnected(d); err != nil {
		return err
	}
	if err := d.HostCeremonyShowQuestion(w.ctx, questionIndex); err != nil {
		return err
	}
	return d.HostCeremonyRevealAnswer(w.ctx, questionIndex)
}

func (w *World) whenQuizmasterPublishesScores(roundIndex int) error {
	d := w.ensureHostDriver()
	if err := w.ensureHostConnected(d); err != nil {
		return err
	}
	return d.HostPublishScores(w.ctx, roundIndex)
}

func (w *World) whenQuizmasterEndsGame() error {
	d := w.ensureHostDriver()
	if err := w.ensureHostConnected(d); err != nil {
		return err
	}
	return d.HostEndGame(w.ctx)
}

// =============================================================================
// Then implementations — assert observable outcomes
// =============================================================================

func (w *World) thenTeamReceivesIdentity(teamName string) error {
	// Observable: play connection received team_registered event with team_id and device_token.
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventTeamRegistered, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive team_registered event", teamName)
	}
	if w.teamID(teamName) == "" {
		return fmt.Errorf("team_registered for %q missing team_id", teamName)
	}
	return nil
}

func (w *World) thenTeamIdentityHasBothTokens() error {
	// Observable: the most recently registered team has both team_id and device_token
	// already captured by captureTeamRegistered when the team_registered event arrived.
	// We check all known teams — at least one must have both fields set.
	w.mu.Lock()
	defer w.mu.Unlock()
	for teamName, id := range w.teamIDs {
		token := w.deviceTokens[teamName]
		if id != "" && token != "" {
			return nil
		}
	}
	return fmt.Errorf("no registered team has both team_id and device_token set")
}

func (w *World) thenTeamInLobby(teamName string) error {
	// Observable: play connection received state_snapshot with state == "LOBBY".
	key := connectionKey(rolePlay, teamName)
	msg, ok := w.waitForEvent(key, eventStateSnapshot, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive state_snapshot", teamName)
	}
	state, _ := msg.Payload["state"].(string)
	if state != "LOBBY" {
		return fmt.Errorf("expected state %q, got %q", "LOBBY", state)
	}
	return nil
}

func (w *World) thenTeamReceivesRoundStarted(teamName string) error {
	// Observable: play connection received round_started event.
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventRoundStarted, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive round_started", teamName)
	}
	return nil
}

func (w *World) thenTeamReceivesRoundStartedForRound(teamName string, roundIndex int) error {
	return w.thenTeamReceivesRoundStarted(teamName)
}

func (w *World) thenTeamSeesRoundBegun(teamName string, roundIndex, questionCount int) error {
	// Observable: round_started payload includes correct round_index and question_count.
	key := connectionKey(rolePlay, teamName)
	msg, ok := w.waitForEvent(key, eventRoundStarted, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive round_started", teamName)
	}
	ri, _ := msg.Payload["round_index"].(float64)
	qc, _ := msg.Payload["question_count"].(float64)
	if int(ri) != roundIndex {
		return fmt.Errorf("round_started round_index: expected %d, got %d", roundIndex, int(ri))
	}
	if int(qc) != questionCount {
		return fmt.Errorf("round_started question_count: expected %d, got %d", questionCount, int(qc))
	}
	return nil
}

func (w *World) thenTeamReceivesFirstQuestion(teamName string) error {
	// Observable: play connection received at least 1 question_revealed event.
	key := connectionKey(rolePlay, teamName)
	if !w.waitForEventCount(key, eventQuestionRevealed, 1, eventWaitTimeout) {
		return fmt.Errorf("team %q did not receive first question_revealed", teamName)
	}
	return nil
}

func (w *World) thenTeamHasReceivedAllQuestions(teamName string, count int) error {
	// Observable: play connection received exactly count question_revealed events.
	key := connectionKey(rolePlay, teamName)
	if !w.waitForEventCount(key, eventQuestionRevealed, count, eventWaitTimeout) {
		got := w.countEvents(key, eventQuestionRevealed)
		return fmt.Errorf("team %q expected %d question_revealed events, got %d", teamName, count, got)
	}
	return nil
}

func (w *World) thenTeamReceivesQuestion(teamName string) error {
	return w.thenTeamReceivesFirstQuestion(teamName)
}

func (w *World) thenQuestionHasText() error {
	// Observable: the most recently received question_revealed event has a non-empty text field
	// in the nested question object. Per AC: the question must not include an answer field.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event != eventQuestionRevealed {
				continue
			}
			question, _ := msg.Payload["question"].(map[string]interface{})
			if question == nil {
				// Fallback: text may be at top level.
				text, _ := msg.Payload["text"].(string)
				if text == "" {
					return fmt.Errorf("question_revealed payload missing text in question object: %v", msg.Payload)
				}
				if _, hasAnswer := msg.Payload["answer"]; hasAnswer {
					return fmt.Errorf("question_revealed payload must not include answer field: %v", msg.Payload)
				}
				return nil
			}
			text, _ := question["text"].(string)
			if text == "" {
				return fmt.Errorf("question_revealed question.text is empty: %v", msg.Payload)
			}
			if _, hasAnswer := question["answer"]; hasAnswer {
				return fmt.Errorf("question_revealed question must not include answer field: %v", msg.Payload)
			}
			return nil
		}
	}
	return fmt.Errorf("no question_revealed event received on any connection")
}

func (w *World) thenTeamReceivesRoundEnded(teamName string) error {
	// Observable: play connection received round_ended event.
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventRoundEnded, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive round_ended", teamName)
	}
	return nil
}

func (w *World) thenRoundEndedHasRoundNumber() error {
	// Observable: any play connection has received a round_ended event with a round_index field.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event == eventRoundEnded {
				if _, ok := msg.Payload["round_index"]; ok {
					return nil
				}
				return fmt.Errorf("round_ended payload missing round_index field: %v", msg.Payload)
			}
		}
	}
	return fmt.Errorf("no round_ended event found on any connection")
}

func (w *World) thenTeamReceivesSubmissionAck(teamName string) error {
	// Observable: play connection received submission_ack event.
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventSubmissionAck, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive submission_ack", teamName)
	}
	return nil
}

func (w *World) thenSubmissionAckShowsLocked(roundIndex int) error {
	// Observable: submission_ack payload has locked:true and matching round_index.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event == eventSubmissionAck {
				locked, _ := msg.Payload["locked"].(bool)
				if !locked {
					return fmt.Errorf("submission_ack locked field is not true: %v", msg.Payload)
				}
				ri, _ := msg.Payload["round_index"].(float64)
				if int(ri) != roundIndex {
					return fmt.Errorf("submission_ack round_index: expected %d, got %d", roundIndex, int(ri))
				}
				return nil
			}
		}
	}
	return fmt.Errorf("no submission_ack event found")
}

func (w *World) thenPlayRoomReceivesSubmissionNotification(teamName string) error {
	// Observable: play connection received submission_received broadcast (DEP-03).
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventSubmissionReceived, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("play room did not receive submission_received for team %q", teamName)
	}
	return nil
}

func (w *World) thenTeamReceivesOtherTeamSubmissionNotification(observerTeam, submittingTeam string) error {
	// Observable: observer's play connection received submission_received with submittingTeam's name.
	key := connectionKey(rolePlay, observerTeam)
	return pollUntil(eventWaitTimeout, 10*time.Millisecond, func() (bool, error) {
		for _, msg := range w.messagesFor(key) {
			if msg.Event != eventSubmissionReceived {
				continue
			}
			name, _ := msg.Payload["team_name"].(string)
			if name == submittingTeam {
				return true, nil
			}
		}
		return false, fmt.Errorf("team %q has not received submission_received for %q", observerTeam, submittingTeam)
	})
}

func (w *World) thenNotificationIncludesTeamName(teamName string) error {
	// Observable: any play connection has a submission_received with the given team_name.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event == eventSubmissionReceived {
				name, _ := msg.Payload["team_name"].(string)
				if name == teamName {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("no submission_received with team_name %q found on any connection", teamName)
}

func (w *World) thenTeamReceivesOwnSubmissionNotification(teamName string) error {
	// Observable: the submitting team's play connection received submission_received for itself.
	key := connectionKey(rolePlay, teamName)
	return pollUntil(eventWaitTimeout, 10*time.Millisecond, func() (bool, error) {
		for _, msg := range w.messagesFor(key) {
			if msg.Event == eventSubmissionReceived {
				name, _ := msg.Payload["team_name"].(string)
				if name == teamName {
					return true, nil
				}
			}
		}
		return false, fmt.Errorf("team %q has not received submission_received for itself", teamName)
	})
}

func (w *World) thenNotificationIncludesTeamNameAndRound(teamName string) error {
	// Observable: submission_received for teamName includes team_name and round_index fields.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event != eventSubmissionReceived {
				continue
			}
			name, _ := msg.Payload["team_name"].(string)
			if name != teamName {
				continue
			}
			if _, ok := msg.Payload["round_index"]; !ok {
				return fmt.Errorf("submission_received for %q missing round_index field: %v", teamName, msg.Payload)
			}
			return nil
		}
	}
	return fmt.Errorf("no submission_received with team_name %q found", teamName)
}

func (w *World) thenTeamReceivesCeremonyQuestion(teamName string) error {
	// Observable: play connection received ceremony_question_shown event.
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventCeremonyQuestionShown, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive ceremony_question_shown", teamName)
	}
	return nil
}

func (w *World) thenCeremonyQuestionHasText() error {
	// Observable: any ceremony_question_shown event has a non-empty question text.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event != eventCeremonyQuestionShown {
				continue
			}
			// Check question.text or top-level text field.
			if q, ok := msg.Payload["question"].(map[string]interface{}); ok {
				if text, _ := q["text"].(string); text != "" {
					return nil
				}
			}
			if text, _ := msg.Payload["text"].(string); text != "" {
				return nil
			}
			return fmt.Errorf("ceremony_question_shown payload has no question text: %v", msg.Payload)
		}
	}
	return fmt.Errorf("no ceremony_question_shown event received on any connection")
}

func (w *World) thenTeamReceivesCeremonyAnswer(teamName string) error {
	// Observable: play connection received ceremony_answer_revealed event (DEP-02).
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventCeremonyAnswerReveal, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive ceremony_answer_revealed", teamName)
	}
	return nil
}

func (w *World) thenVerdictsShowTeamResults() error {
	// Observable: latest ceremony_answer_revealed payload has a non-empty verdicts array.
	// We check any play connection that has received this event.
	for key, msgs := range w.receivedMessages {
		_ = key
		for _, msg := range msgs {
			if msg.Event == eventCeremonyAnswerReveal {
				verdicts, ok := msg.Payload["verdicts"].([]interface{})
				if ok && len(verdicts) > 0 {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("no ceremony_answer_revealed with non-empty verdicts found on any connection")
}

func (w *World) thenVerdictsIncludeTeam(teamName string) error {
	// Observable: ceremony_answer_revealed verdicts array contains an entry with team_name == teamName.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event != eventCeremonyAnswerReveal {
				continue
			}
			verdicts, ok := msg.Payload["verdicts"].([]interface{})
			if !ok {
				return fmt.Errorf("ceremony_answer_revealed missing verdicts array: %v", msg.Payload)
			}
			for _, raw := range verdicts {
				entry, _ := raw.(map[string]interface{})
				if entry["team_name"] == teamName {
					return nil
				}
			}
			return fmt.Errorf("verdicts array does not include team %q: %v", teamName, verdicts)
		}
	}
	return fmt.Errorf("no ceremony_answer_revealed event found on any connection")
}

func (w *World) thenEachVerdictHasResult() error {
	// Observable: every verdict entry in ceremony_answer_revealed has a verdict field.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event != eventCeremonyAnswerReveal {
				continue
			}
			verdicts, ok := msg.Payload["verdicts"].([]interface{})
			if !ok {
				return fmt.Errorf("ceremony_answer_revealed missing verdicts array: %v", msg.Payload)
			}
			for _, raw := range verdicts {
				entry, _ := raw.(map[string]interface{})
				if _, ok := entry["verdict"]; !ok {
					return fmt.Errorf("verdict entry missing verdict field: %v", entry)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("no ceremony_answer_revealed event found on any connection")
}

func (w *World) thenTeamHasReceivedCeremonyQuestionCount(teamName string, count int) error {
	// Observable: play connection for teamName has received exactly count ceremony_question_shown events.
	key := connectionKey(rolePlay, teamName)
	if !w.waitForEventCount(key, eventCeremonyQuestionShown, count, eventWaitTimeout) {
		got := w.countEvents(key, eventCeremonyQuestionShown)
		return fmt.Errorf("team %q expected %d ceremony_question_shown events, got %d", teamName, count, got)
	}
	return nil
}

func (w *World) thenTeamHasReceivedAnswerRevealCount(teamName string, count int) error {
	// Observable: play connection for teamName has received exactly count ceremony_answer_revealed events.
	key := connectionKey(rolePlay, teamName)
	if !w.waitForEventCount(key, eventCeremonyAnswerReveal, count, eventWaitTimeout) {
		got := w.countEvents(key, eventCeremonyAnswerReveal)
		return fmt.Errorf("team %q expected %d ceremony_answer_revealed events, got %d", teamName, count, got)
	}
	return nil
}

func (w *World) thenTeamReceivesRoundScores(teamName string) error {
	// Observable: play connection received round_scores_published event (DEP-04).
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventRoundScoresPublished, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive round_scores_published", teamName)
	}
	return nil
}

func (w *World) thenScoresListHasTeamNames() error {
	// Observable: round_scores_published payload has a scores array with team_name in each entry.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event == eventRoundScoresPublished {
				scores, ok := msg.Payload["scores"].([]interface{})
				if !ok || len(scores) == 0 {
					return fmt.Errorf("round_scores_published has no scores array")
				}
				for _, raw := range scores {
					entry, _ := raw.(map[string]interface{})
					if _, ok := entry["team_name"]; !ok {
						return fmt.Errorf("score entry missing team_name field: %v", entry)
					}
				}
				return nil
			}
		}
	}
	return fmt.Errorf("no round_scores_published event found")
}

func (w *World) thenScoresListIncludesTeam(teamName string) error {
	// Observable: round_scores_published payload scores array contains an entry for teamName.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event == eventRoundScoresPublished {
				scores, _ := msg.Payload["scores"].([]interface{})
				for _, raw := range scores {
					entry, _ := raw.(map[string]interface{})
					if entry["team_name"] == teamName {
						return nil
					}
				}
				return fmt.Errorf("team %q not found in round scores: %v", teamName, scores)
			}
		}
	}
	return fmt.Errorf("no round_scores_published event found")
}

func (w *World) thenScoresListIncludesTeamRoundScore(teamName string) error {
	return w.thenScoresListIncludesTeam(teamName)
}

func (w *World) thenTeamReceivesFinalScores(teamName string) error {
	// Observable: play connection received game_over event.
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventGameOver, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive game_over event", teamName)
	}
	return nil
}

func (w *World) thenFinalScoresHaveTeamNames() error {
	// Observable: game_over payload has a final_scores array with team_name in each entry.
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event == eventGameOver {
				finalScores, ok := msg.Payload["final_scores"].([]interface{})
				if !ok || len(finalScores) == 0 {
					return fmt.Errorf("game_over payload has no final_scores array: %v", msg.Payload)
				}
				for _, raw := range finalScores {
					entry, _ := raw.(map[string]interface{})
					if _, ok := entry["team_name"]; !ok {
						return fmt.Errorf("final_scores entry missing team_name field: %v", entry)
					}
					if _, ok := entry["total"]; !ok {
						return fmt.Errorf("final_scores entry missing total field: %v", entry)
					}
				}
				return nil
			}
		}
	}
	return fmt.Errorf("no game_over event found on any connection")
}

func (w *World) thenTeamReceivesGameState(teamName string) error {
	// Observable: play connection received a state_snapshot event after connecting.
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventStateSnapshot, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive state_snapshot", teamName)
	}
	return nil
}

func (w *World) thenGameStateIsLobby() error {
	// Observable: the most recently received state_snapshot for any play connection has state == "LOBBY".
	for key, msgs := range w.receivedMessages {
		if !strings.HasPrefix(key, rolePlay+":") && key != rolePlay {
			continue
		}
		for _, msg := range msgs {
			if msg.Event == eventStateSnapshot {
				state, _ := msg.Payload["state"].(string)
				if state == "LOBBY" {
					return nil
				}
				return fmt.Errorf("state_snapshot state expected %q, got %q", "LOBBY", state)
			}
		}
	}
	return fmt.Errorf("no state_snapshot received on any play connection")
}

func (w *World) thenTeamReceivesStateSnapshot(teamName string) error {
	// Observable: the team's play connection received a state_snapshot event (after rejoin).
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventStateSnapshot, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive state_snapshot after rejoin", teamName)
	}
	return nil
}

func (w *World) thenSnapshotHasDraftAnswers(teamName string) error {
	// Observable: at least one state_snapshot sent to teamName includes a non-empty draft_answers field.
	// After reconnect the server sends two snapshots: one on connect and one on rejoin.
	// We poll until we find a snapshot that has a non-empty draft_answers field.
	key := connectionKey(rolePlay, teamName)
	return pollUntil(eventWaitTimeout, 20*time.Millisecond, func() (bool, error) {
		var lastErr error
		for _, msg := range w.messagesFor(key) {
			if msg.Event != eventStateSnapshot {
				continue
			}
			drafts, ok := msg.Payload["draft_answers"]
			if !ok {
				lastErr = fmt.Errorf("state_snapshot for %q missing draft_answers field: %v", teamName, msg.Payload)
				continue // check next snapshot
			}
			// draft_answers may be a map or a list — either way must be non-nil/non-empty.
			switch v := drafts.(type) {
			case map[string]interface{}:
				if len(v) > 0 {
					return true, nil
				}
				lastErr = fmt.Errorf("state_snapshot draft_answers for %q is empty map", teamName)
			case []interface{}:
				if len(v) > 0 {
					return true, nil
				}
				lastErr = fmt.Errorf("state_snapshot draft_answers for %q is empty slice", teamName)
			default:
				if drafts != nil {
					return true, nil
				}
				lastErr = fmt.Errorf("state_snapshot draft_answers for %q is nil", teamName)
			}
		}
		if lastErr != nil {
			return false, lastErr
		}
		return false, fmt.Errorf("no state_snapshot received for team %q", teamName)
	})
}

func (w *World) thenSnapshotShowsRoundActive() error {
	// Observable: at least one play connection has received a state_snapshot with state == "ROUND_ACTIVE".
	// Multiple snapshots may arrive (initial connect + rejoin); we accept any one with the right state.
	return pollUntil(eventWaitTimeout, 20*time.Millisecond, func() (bool, error) {
		for key, msgs := range w.receivedMessages {
			if !strings.HasPrefix(key, rolePlay+":") && key != rolePlay {
				continue
			}
			for _, msg := range msgs {
				if msg.Event == eventStateSnapshot {
					state, _ := msg.Payload["state"].(string)
					if state == string(stateRoundActive) {
						return true, nil
					}
				}
			}
		}
		return false, fmt.Errorf("no state_snapshot with state %q found on any play connection", stateRoundActive)
	})
}

func (w *World) thenSnapshotShowsCeremony() error {
	// Observable: at least one play connection has received a state_snapshot with state == "CEREMONY".
	// Multiple snapshots may arrive (initial connect + rejoin); we accept any one with the right state.
	return pollUntil(eventWaitTimeout, 20*time.Millisecond, func() (bool, error) {
		for key, msgs := range w.receivedMessages {
			if !strings.HasPrefix(key, rolePlay+":") && key != rolePlay {
				continue
			}
			for _, msg := range msgs {
				if msg.Event == eventStateSnapshot {
					state, _ := msg.Payload["state"].(string)
					if state == string(stateCeremony) {
						return true, nil
					}
				}
			}
		}
		return false, fmt.Errorf("no state_snapshot with state %q found on any play connection", stateCeremony)
	})
}

func (w *World) thenSnapshotShowsRoundScores() error {
	// Observable: at least one play connection has received a state_snapshot with state == "ROUND_SCORES".
	// Multiple snapshots may arrive (initial connect + rejoin); we accept any one with the right state.
	return pollUntil(eventWaitTimeout, 20*time.Millisecond, func() (bool, error) {
		for key, msgs := range w.receivedMessages {
			if !strings.HasPrefix(key, rolePlay+":") && key != rolePlay {
				continue
			}
			for _, msg := range msgs {
				if msg.Event == eventStateSnapshot {
					state, _ := msg.Payload["state"].(string)
					if state == string(stateRoundScores) {
						return true, nil
					}
				}
			}
		}
		return false, fmt.Errorf("no state_snapshot with state %q found on any play connection", stateRoundScores)
	})
}

func (w *World) thenSnapshotHasRevealedQuestions(teamName string, count int) error {
	// Observable: the state_snapshot sent to the team (after mid-round join) includes
	// a revealed_questions array with the expected number of entries.
	key := connectionKey(rolePlay, teamName)
	return pollUntil(eventWaitTimeout, 20*time.Millisecond, func() (bool, error) {
		for _, msg := range w.messagesFor(key) {
			if msg.Event != eventStateSnapshot {
				continue
			}
			revealed, _ := msg.Payload["revealed_questions"].([]interface{})
			if len(revealed) >= count {
				return true, nil
			}
			return false, fmt.Errorf("state_snapshot for %q has %d revealed questions, expected %d (payload: %v)",
				teamName, len(revealed), count, msg.Payload)
		}
		return false, fmt.Errorf("no state_snapshot received for team %q", teamName)
	})
}

func (w *World) thenSnapshotShowsRevealedQuestions(count int) error {
	// Observable: at least one play connection has a state_snapshot with revealed_questions
	// array containing at least count entries.
	return pollUntil(eventWaitTimeout, 20*time.Millisecond, func() (bool, error) {
		for key, msgs := range w.receivedMessages {
			if !strings.HasPrefix(key, rolePlay+":") && key != rolePlay {
				continue
			}
			for _, msg := range msgs {
				if msg.Event != eventStateSnapshot {
					continue
				}
				revealed, _ := msg.Payload["revealed_questions"].([]interface{})
				if len(revealed) >= count {
					return true, nil
				}
				return false, fmt.Errorf("state_snapshot has %d revealed_questions, expected %d (payload: %v)",
					len(revealed), count, msg.Payload)
			}
		}
		return false, fmt.Errorf("no state_snapshot with revealed_questions found on any play connection")
	})
}

func (w *World) thenSnapshotContainsDraftForQuestion(teamName string, roundIndex, questionIndex int) error {
	// Observable: the state_snapshot for teamName contains a draft_answers entry for the given question.
	key := connectionKey(rolePlay, teamName)
	return pollUntil(eventWaitTimeout, 20*time.Millisecond, func() (bool, error) {
		for _, msg := range w.messagesFor(key) {
			if msg.Event != eventStateSnapshot {
				continue
			}
			drafts, ok := msg.Payload["draft_answers"]
			if !ok {
				return false, fmt.Errorf("state_snapshot missing draft_answers field: %v", msg.Payload)
			}
			// draft_answers may be a map[string]interface{} or []interface{}.
			switch v := drafts.(type) {
			case map[string]interface{}:
				if len(v) > 0 {
					return true, nil
				}
			case []interface{}:
				if len(v) > 0 {
					return true, nil
				}
			}
			return false, fmt.Errorf("draft_answers present but empty for team %q: %v", teamName, drafts)
		}
		return false, fmt.Errorf("no state_snapshot received for team %q", teamName)
	})
}

func (w *World) thenSnapshotHasNoDraftAnswers(teamName string) error {
	// Observable: the state_snapshot for teamName either has no draft_answers field,
	// or the field is present but empty.
	key := connectionKey(rolePlay, teamName)
	return pollUntil(eventWaitTimeout, 20*time.Millisecond, func() (bool, error) {
		for _, msg := range w.messagesFor(key) {
			if msg.Event != eventStateSnapshot {
				continue
			}
			drafts, exists := msg.Payload["draft_answers"]
			if !exists {
				return true, nil // field absent — no drafts
			}
			// Field present — must be nil or empty.
			switch v := drafts.(type) {
			case map[string]interface{}:
				if len(v) == 0 {
					return true, nil
				}
				return false, fmt.Errorf("state_snapshot draft_answers for fresh connection is non-empty: %v", drafts)
			case []interface{}:
				if len(v) == 0 {
					return true, nil
				}
				return false, fmt.Errorf("state_snapshot draft_answers for fresh connection is non-empty: %v", drafts)
			case nil:
				return true, nil
			default:
				return false, fmt.Errorf("state_snapshot draft_answers unexpected type %T: %v", drafts, drafts)
			}
		}
		return false, fmt.Errorf("no state_snapshot received for team %q", teamName)
	})
}

func (w *World) thenSecondDeviceReceivesDuplicateError() error {
	// Observable: the second device connection received an error event with code team_name_taken.
	// Look through all connection keys for the second device suffix.
	const secondDeviceSuffix = ":device2"
	timeout := time.After(eventWaitTimeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			// Report what we found for diagnostics.
			for key, msgs := range w.receivedMessages {
				if strings.HasSuffix(key, secondDeviceSuffix) {
					for _, msg := range msgs {
						if msg.Event == eventError {
							code, _ := msg.Payload["code"].(string)
							if code == "team_name_taken" {
								return nil
							}
						}
					}
					return fmt.Errorf("second device received no team_name_taken error; messages: %v", msgs)
				}
			}
			return fmt.Errorf("second device connection not found in received messages")
		case <-ticker.C:
			for key, msgs := range w.receivedMessages {
				if strings.HasSuffix(key, secondDeviceSuffix) {
					for _, msg := range msgs {
						if msg.Event == eventError {
							code, _ := msg.Payload["code"].(string)
							if code == "team_name_taken" {
								return nil
							}
						}
					}
				}
			}
		}
	}
}

func (w *World) thenReceivesTeamNotFoundError() error {
	// Observable: the anonymous play connection received an error event with code "invalid_token"
	// or "team_not_found" (server sends one of these on bad rejoin token).
	const anonKey = "play:anonymous"
	timeout := time.After(eventWaitTimeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			for _, msg := range w.messagesFor(anonKey) {
				if msg.Event == eventError {
					code, _ := msg.Payload["code"].(string)
					return fmt.Errorf("received error with code %q (expected invalid_token or team_not_found): %v", code, msg.Payload)
				}
			}
			return fmt.Errorf("no error event received after bad rejoin attempt")
		case <-ticker.C:
			for _, msg := range w.messagesFor(anonKey) {
				if msg.Event != eventError {
					continue
				}
				code, _ := msg.Payload["code"].(string)
				if code == "invalid_token" || code == "team_not_found" {
					return nil
				}
			}
		}
	}
}

func (w *World) thenNoStateSnapshotSent() error {
	// Observable: no additional state_snapshot is sent after the bad rejoin.
	// The initial state_snapshot on connect is expected; no second one should follow.
	// Wait briefly then count state_snapshot events on the anonymous connection.
	time.Sleep(negativeEventWindow)
	const anonKey = "play:anonymous"
	count := w.countEvents(anonKey, eventStateSnapshot)
	// The server sends one state_snapshot on initial connect (before rejoin).
	// After a failed rejoin, no additional snapshot should be sent.
	// We allow 0 or 1 (the initial connect snapshot); reject > 1.
	if count > 1 {
		return fmt.Errorf("expected at most 1 state_snapshot (initial connect), got %d after bad rejoin", count)
	}
	return nil
}

func (w *World) thenTeamReceivesAlreadySubmittedError(teamName string) error {
	// Observable: play connection received an error event with code "already_submitted".
	key := connectionKey(rolePlay, teamName)
	timeout := time.After(eventWaitTimeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			for _, msg := range w.messagesFor(key) {
				if msg.Event == eventError {
					code, _ := msg.Payload["code"].(string)
					return fmt.Errorf("team %q received error but code was %q (expected already_submitted): %v", teamName, code, msg.Payload)
				}
			}
			return fmt.Errorf("team %q did not receive an error event", teamName)
		case <-ticker.C:
			for _, msg := range w.messagesFor(key) {
				if msg.Event == eventError {
					code, _ := msg.Payload["code"].(string)
					if code == "already_submitted" {
						return nil
					}
				}
			}
		}
	}
}

func (w *World) thenTeamReceivesErrorResponse(teamName string) error {
	// Observable: play connection received any error event.
	key := connectionKey(rolePlay, teamName)
	if _, ok := w.waitForEvent(key, eventError, eventWaitTimeout); !ok {
		return fmt.Errorf("team %q did not receive an error event", teamName)
	}
	return nil
}

func (w *World) thenAnonymousPlayerReceivesError() error {
	// Observable: the anonymous player connection received an error event.
	const anonKey = "play:anonymous"
	_, ok := w.waitForEvent(anonKey, eventError, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("anonymous player did not receive an error event")
	}
	return nil
}

func (w *World) thenNoTeamIdentityIssued() error {
	// Observable: no team_registered event was received on the anonymous connection.
	const anonKey = "play:anonymous"
	time.Sleep(negativeEventWindow)
	for _, msg := range w.messagesFor(anonKey) {
		if msg.Event == eventTeamRegistered {
			return fmt.Errorf("unexpected team_registered event received by anonymous player")
		}
	}
	return nil
}

func (w *World) thenNoErrorReturnedToTeam(teamName string) error {
	// Observable: no error event received on the team's play connection within the window.
	time.Sleep(negativeEventWindow)
	key := connectionKey(rolePlay, teamName)
	for _, msg := range w.messagesFor(key) {
		if msg.Event == eventError {
			return fmt.Errorf("unexpected error event received for team %q: %v", teamName, msg.Payload)
		}
	}
	return nil
}

func (w *World) thenDraftSavedWithoutError() error {
	// draft_answer is fire-and-forget; assert no error event received on any play connection.
	time.Sleep(negativeEventWindow)
	for key, msgs := range w.receivedMessages {
		if !strings.HasPrefix(key, rolePlay+":") {
			continue
		}
		for _, msg := range msgs {
			if msg.Event == eventError {
				return fmt.Errorf("unexpected error event received after draft_answer on %q: %v", key, msg.Payload)
			}
		}
	}
	return nil
}

func (w *World) thenQuestionHasChoices(teamName string) error {
	return godog.ErrPending
}

func (w *World) thenChoicesListHasCount(count int) error {
	return godog.ErrPending
}

func (w *World) thenQuestionHasNoChoices(teamName string) error {
	return godog.ErrPending
}

func (w *World) thenQuestionHasMultiPartIndicator(teamName string) error {
	return godog.ErrPending
}

func (w *World) thenQuestionHasNoMultiPartIndicator(teamName string) error {
	return godog.ErrPending
}

func (w *World) thenQuestionHasMediaReference(teamName string) error {
	return godog.ErrPending
}

func (w *World) thenMediaReferenceHasTypeAndURL() error {
	return godog.ErrPending
}

func (w *World) thenQuestionHasNoMedia(teamName string) error {
	return godog.ErrPending
}

func (w *World) thenConnectionAcceptedWithSnapshot(teamName string) error {
	// Observable: connection was upgraded to WebSocket and a state_snapshot
	// was sent immediately — before any team_register.
	key := connectionKey(rolePlay, teamName)
	_, ok := w.waitForEvent(key, eventStateSnapshot, eventWaitTimeout)
	if !ok {
		return fmt.Errorf("team %q did not receive state_snapshot after connecting", teamName)
	}
	return nil
}

func (w *World) thenRoundScoresPayloadHasStructuredList(teamName string) error {
	// Observable: round_scores_published payload has a non-empty scores array.
	key := connectionKey(rolePlay, teamName)
	return pollUntil(eventWaitTimeout, 10*time.Millisecond, func() (bool, error) {
		for _, msg := range w.messagesFor(key) {
			if msg.Event != eventRoundScoresPublished {
				continue
			}
			scores, ok := msg.Payload["scores"].([]interface{})
			if !ok || len(scores) == 0 {
				return false, fmt.Errorf("round_scores_published has no scores array: %v", msg.Payload)
			}
			return true, nil
		}
		return false, fmt.Errorf("team %q has not received round_scores_published", teamName)
	})
}

func (w *World) thenEachScoreEntryHasTeamName() error {
	return w.thenScoresListHasTeamNames()
}

func (w *World) thenCeremonyAnswerPayloadHasVerdicts(teamName string) error {
	return godog.ErrPending
}

func (w *World) thenVerdictsListPresent() error {
	// Observable: ceremony_answer_revealed payload has a verdicts array (may be empty if no teams submitted).
	for _, msgs := range w.receivedMessages {
		for _, msg := range msgs {
			if msg.Event != eventCeremonyAnswerReveal {
				continue
			}
			if _, ok := msg.Payload["verdicts"]; ok {
				return nil
			}
			return fmt.Errorf("ceremony_answer_revealed payload missing verdicts field: %v", msg.Payload)
		}
	}
	return fmt.Errorf("no ceremony_answer_revealed event found on any connection")
}
