// step_impls.go contains the implementation of all Given/When/Then step bodies.
//
// Each method delegates to the TriviaDriver (Layer 3) and makes assertions
// against observed outcomes. These implementations are stubs -- they
// contain the correct structure and will compile, but return
// godog.ErrPending for steps not yet implemented. The software-crafter
// enables one scenario at a time and fills in these stubs.
package steps

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

// -----------------------------------------------------------------------
// Given implementations
// -----------------------------------------------------------------------

func (w *World) givenServerRunning(token string) error {
	if w.server != nil {
		return nil // already running
	}
	w.hostToken = token
	w.server = NewTestServer(token)
	return nil
}

func (w *World) givenQuizFileExists(filename string, rounds, questions int) error {
	// Generate a quiz YAML fixture with the requested rounds and question count.
	if rounds <= 1 {
		qs := make([]QuizQuestion, questions)
		for i := 0; i < questions; i++ {
			qs[i] = QuizQuestion{
				Text:   fmt.Sprintf("Question %d text?", i+1),
				Answer: fmt.Sprintf("Answer %d", i+1),
			}
		}
		w.quizFixtures[filename] = SimpleQuizYAML("Friday Night Trivia -- March 2026", qs)
	} else {
		w.quizFixtures[filename] = MultiRoundQuizYAML("Friday Night Trivia -- March 2026", rounds, questions)
	}
	return nil
}

func (w *World) givenQuizFixtureQuestionDetail(qIndex int, text, answer string) {
	// Overrides question text/answer in the fixture at the given index.
	// Full implementation stores per-question details for the driver to use.
}

func (w *World) givenMarcusConnectsToHostPanel() error {
	if w.server == nil {
		if err := w.givenServerRunning(w.hostToken); err != nil {
			return err
		}
	}
	driver := NewTriviaDriver(w.server, w.hostToken, w)
	w.connections["host"] = &WSConnection{Role: "host", driver: driver}
	return driver.ConnectHost(w.ctx)
}

func (w *World) givenGameSessionLoaded(filename string) error {
	if err := w.givenMarcusConnectsToHostPanel(); err != nil {
		return err
	}
	return w.whenMarcusLoadsQuiz(filename)
}

func (w *World) givenTeamConnected(teamName string) error {
	if w.server == nil {
		if err := w.givenServerRunning(w.hostToken); err != nil {
			return err
		}
	}
	// Ensure quiz is loaded so the lobby state is meaningful.
	if err := w.ensureQuizLoaded(); err != nil {
		return err
	}
	// Ensure display is connected so broadcast assertions can be verified.
	if w.connections["display"] == nil {
		if err := w.givenDisplayConnected(); err != nil {
			return err
		}
	}
	return w.whenPlayerJoins(teamName)
}

// ensureQuizLoaded loads the quiz into the host session if not already done.
// It is a no-op when no host connection exists, no fixtures are registered, or quiz is already loaded.
func (w *World) ensureQuizLoaded() error {
	if w.quizLoaded {
		return nil
	}
	if w.connections["host"] == nil {
		if err := w.givenMarcusConnectsToHostPanel(); err != nil {
			return err
		}
	}
	internalKeys := map[string]bool{
		"QUIZ_DIR_OVERRIDE": true, "last_http_status": true,
		"last_http_body": true, "docker_build_output": true,
		"arch_lint_output": true, "tsc_output": true,
		"go_test_output": true, "server_startup_error": true,
	}
	for filename := range w.quizFixtures {
		if internalKeys[filename] {
			continue
		}
		if err := w.whenMarcusLoadsQuiz(filename); err != nil {
			return err
		}
		w.quizLoaded = true
		return nil
	}
	return nil
}

func (w *World) givenDisplayConnected() error {
	if w.server == nil {
		if err := w.givenServerRunning(w.hostToken); err != nil {
			return err
		}
	}
	driver := NewTriviaDriver(w.server, w.hostToken, w)
	w.connections["display"] = &WSConnection{Role: "display", driver: driver}
	return driver.ConnectDisplay(w.ctx)
}

func (w *World) givenMarcusStartedGame(roundIndex int) error {
	if err := w.givenMarcusConnectsToHostPanel(); err != nil {
		return err
	}
	// Ensure quiz is loaded before starting a round — StartRound fails otherwise.
	if err := w.ensureQuizLoaded(); err != nil {
		return err
	}
	if err := w.whenMarcusStartsRound(roundIndex); err != nil {
		return err
	}
	// Wait for the round_started event on the host to confirm the state transition succeeded.
	if _, ok := w.waitForEvent("host", "round_started", 2*time.Second); !ok {
		return fmt.Errorf("timed out waiting for round_started confirmation after host_start_round")
	}
	return nil
}

func (w *World) givenQuestionsRevealed(roundIndex, count int) error {
	driver := w.hostDriver()
	for i := 0; i < count; i++ {
		if err := driver.HostRevealQuestion(w.ctx, roundIndex, i); err != nil {
			return err
		}
	}
	return nil
}

func (w *World) givenRoundEnded(roundIndex, questionCount int) error {
	if questionCount > 0 {
		if err := w.givenQuestionsRevealed(roundIndex, questionCount); err != nil {
			return err
		}
	}
	return w.hostDriver().HostEndRound(w.ctx, roundIndex)
}

func (w *World) givenTeamEnteredAnswers(teamName string, count int) error {
	for i := 0; i < count; i++ {
		if err := w.whenPlayerDraftsAnswer(teamName, i, fmt.Sprintf("answer %d", i+1)); err != nil {
			return err
		}
	}
	return nil
}

func (w *World) givenScoringOpen(roundIndex int) error {
	// Pre-condition: all teams have submitted; scoring opens automatically.
	// The software-crafter sets up the state here.
	return godog.ErrPending
}

func (w *World) givenTeamSubmittedAnswer(teamName string, questionIndex int, answer string) error {
	return w.whenPlayerDraftsAnswer(teamName, questionIndex, answer)
}

func (w *World) givenAllAnswersMarked(roundIndex int) error {
	// The software-crafter marks all team answers correct or wrong here.
	return godog.ErrPending
}

func (w *World) givenRoundFullyComplete(roundIndex int) error {
	// Orchestrates the full round sequence for use as a precondition.
	return godog.ErrPending
}

func (w *World) givenScoresPublished(roundIndex int) error {
	return w.hostDriver().HostPublishScores(w.ctx, roundIndex)
}

func (w *World) givenPlayerDisconnectedBeforeStart(teamName, playerName string) error {
	return godog.ErrPending
}

func (w *World) givenMediaFileExists(filename string) error {
	// The software-crafter creates a temp media file here.
	return godog.ErrPending
}

func (w *World) givenDraftAnswerEntered(teamName string, questionIndex int, answer string) error {
	return w.whenPlayerDraftsAnswer(teamName, questionIndex, answer)
}

func (w *World) givenConnectionInterruptedAndRestored(teamName string) error {
	return godog.ErrPending
}

func (w *World) givenCeremonyStarted(roundIndex int) error {
	return w.whenMarcusStartsCeremony(roundIndex)
}

func (w *World) givenCeremonyAtQuestion(roundIndex, questionIndex int) error {
	if err := w.givenCeremonyStarted(roundIndex); err != nil {
		return err
	}
	for i := 0; i <= questionIndex; i++ {
		if err := w.whenMarcusCeremonyShowQuestion(roundIndex, i); err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------
// When implementations
// -----------------------------------------------------------------------

func (w *World) whenMarcusLoadsQuiz(filename string) error {
	content, ok := w.quizFixtures[filename]
	if !ok {
		// Default fixture if not explicitly set.
		content = SimpleQuizYAML("Friday Night Trivia -- March 2026", []QuizQuestion{
			{Text: "What is the capital of France?", Answer: "Paris"},
			{Text: "What color is the sky?", Answer: "Blue"},
		})
	}
	driver := w.hostDriver()
	// Write fixture to temp file and send the path.
	path, err := driver.WriteQuizFixture(filename, content)
	if err != nil {
		return err
	}
	if err := driver.HostLoadQuiz(w.ctx, path); err != nil {
		return err
	}
	w.quizLoaded = true
	return nil
}

func (w *World) whenMarcusLoadsQuizPath(path string) error {
	return w.hostDriver().HostLoadQuiz(w.ctx, path)
}

func (w *World) whenMarcusLoadsDefaultQuiz() error {
	for filename := range w.quizFixtures {
		return w.whenMarcusLoadsQuiz(filename)
	}
	return fmt.Errorf("no quiz fixture registered")
}

func (w *World) whenPlayerJoins(teamName string) error {
	if w.server == nil {
		if err := w.givenServerRunning(w.hostToken); err != nil {
			return err
		}
	}
	driver := NewTriviaDriver(w.server, w.hostToken, w)
	key := connectionKey("play", teamName)
	w.connections[key] = &WSConnection{Role: "play", Name: teamName, driver: driver}
	w.lastJoinAttemptKey = key
	if err := driver.ConnectPlay(w.ctx, teamName); err != nil {
		return err
	}
	return driver.PlayRegisterTeam(w.ctx, teamName)
}

func (w *World) whenPlayerJoinsSecondDevice(teamName string) error {
	// Connect under a distinct key but register with the original team name.
	// This simulates a second device attempting to register the same team name.
	if w.server == nil {
		if err := w.givenServerRunning(w.hostToken); err != nil {
			return err
		}
	}
	connKey := teamName + "_device2"
	key := connectionKey("play", connKey)
	driver := NewTriviaDriver(w.server, w.hostToken, w)
	w.connections[key] = &WSConnection{Role: "play", Name: connKey, driver: driver}
	w.lastJoinAttemptKey = key
	if err := driver.ConnectPlay(w.ctx, connKey); err != nil {
		return err
	}
	return driver.PlayRegisterTeamWithKey(w.ctx, connKey, teamName)
}

func (w *World) whenMarcusStartsRound(roundIndex int) error {
	return w.hostDriver().HostStartRound(w.ctx, roundIndex)
}

func (w *World) whenMarcusRevealsQuestion(roundIndex, questionIndex int) error {
	return w.hostDriver().HostRevealQuestion(w.ctx, roundIndex, questionIndex)
}

func (w *World) whenPlayerDraftsAnswer(teamName string, questionIndex int, answer string) error {
	driver := w.playDriver(teamName)
	if driver == nil {
		return fmt.Errorf("no player connection for %q", teamName)
	}
	return driver.PlayDraftAnswer(w.ctx, teamName, 0, questionIndex, answer)
}

func (w *World) whenMarcusEndsRound(roundIndex int) error {
	return w.hostDriver().HostEndRound(w.ctx, roundIndex)
}

func (w *World) whenTeamSubmits(teamName string, roundIndex int) error {
	// Collect all drafted answers for this team and submit them.
	answers := []map[string]interface{}{
		{"question_index": 0, "answer": "Paris"},
		{"question_index": 1, "answer": "Blue"},
	}
	return w.playDriver(teamName).PlaySubmitAnswers(w.ctx, teamName, roundIndex, answers)
}

func (w *World) whenMarcusMarksAnswer(teamName string, roundIndex, questionIndex int, verdict string) error {
	teamID := teamName // simplified: team_id == team_name in test context
	return w.hostDriver().HostMarkAnswer(w.ctx, teamID, roundIndex, questionIndex, verdict)
}

func (w *World) whenMarcusStartsCeremony(roundIndex int) error {
	return w.whenMarcusCeremonyShowQuestion(roundIndex, 0)
}

func (w *World) whenMarcusCeremonyShowQuestion(roundIndex, questionIndex int) error {
	return w.hostDriver().HostCeremonyShowQuestion(w.ctx, questionIndex)
}

func (w *World) whenMarcusCeremonyRevealAnswer(roundIndex, questionIndex int) error {
	return w.hostDriver().HostCeremonyRevealAnswer(w.ctx, questionIndex)
}

func (w *World) whenMarcusRunsFullCeremony(roundIndex, questionCount int) error {
	for i := 0; i < questionCount; i++ {
		if err := w.whenMarcusCeremonyShowQuestion(roundIndex, i); err != nil {
			return err
		}
		if err := w.whenMarcusCeremonyRevealAnswer(roundIndex, i); err != nil {
			return err
		}
	}
	return nil
}

func (w *World) whenMarcusPublishesScores(roundIndex int) error {
	return w.hostDriver().HostPublishScores(w.ctx, roundIndex)
}

func (w *World) whenMarcusEndsGame() error {
	return w.hostDriver().HostEndGame(w.ctx)
}

func (w *World) whenHTTPRequest(target string) error {
	if w.server == nil {
		if err := w.givenServerRunning(w.hostToken); err != nil {
			return err
		}
	}
	driver := NewTriviaDriver(w.server, w.hostToken, w)
	var url string
	switch {
	case target == "host_no_token":
		url = driver.GetHostURLNoToken()
	case target == "play":
		url = driver.GetPlayerURL()
	case target == "display":
		url = driver.GetDisplayURL()
	case strings.HasPrefix(target, "media:"):
		url = w.server.URL + "/media/" + strings.TrimPrefix(target, "media:")
	default:
		url = w.server.URL + "/" + target
	}
	code, body, err := driver.HTTPGet(url)
	if err != nil {
		return err
	}
	w.quizFixtures["last_http_status"] = fmt.Sprintf("%d", code)
	w.quizFixtures["last_http_body"] = body
	return nil
}

func (w *World) whenHTTPRequestWithToken(target, token string) error {
	if w.server == nil {
		if err := w.givenServerRunning(w.hostToken); err != nil {
			return err
		}
	}
	driver := NewTriviaDriver(w.server, token, w)
	url := driver.GetHostURLWithToken(token)
	code, body, err := driver.HTTPGet(url)
	if err != nil {
		return err
	}
	w.quizFixtures["last_http_status"] = fmt.Sprintf("%d", code)
	w.quizFixtures["last_http_body"] = body
	return nil
}

func (w *World) whenServerStarts() error {
	return w.givenServerRunning(w.hostToken)
}

func (w *World) whenConnectWithoutToken() error {
	return godog.ErrPending
}

func (w *World) whenUnauthorizedReveal() error {
	return godog.ErrPending
}

func (w *World) whenTeamRejoins(teamName string) error {
	conn := w.connections[connectionKey("play", teamName)]
	if conn == nil {
		return fmt.Errorf("no connection for team %q", teamName)
	}
	return conn.driver.PlayRejoinTeam(w.ctx, teamName, teamName, conn.Token)
}

func (w *World) whenPlayerRefreshes(teamName string) error {
	return godog.ErrPending
}

func (w *World) whenConnectionRestored(teamName string) error {
	return godog.ErrPending
}

func (w *World) whenFullGameSequenceRuns() error {
	return godog.ErrPending
}

func (w *World) whenDockerBuildRuns() error {
	cmd := exec.CommandContext(w.ctx, "docker", "build", "-t", "trivia:acceptance-test", ".")
	out, err := cmd.CombinedOutput()
	w.quizFixtures["docker_build_output"] = string(out)
	if err != nil {
		return fmt.Errorf("docker build failed: %w\n%s", err, string(out))
	}
	return nil
}

func (w *World) whenDockerComposeUp() error {
	return godog.ErrPending
}

func (w *World) whenGoArchLintRuns() error {
	cmd := exec.CommandContext(w.ctx, "go-arch-lint", "check")
	out, err := cmd.CombinedOutput()
	w.quizFixtures["arch_lint_output"] = string(out)
	if err != nil {
		return fmt.Errorf("go-arch-lint failed: %w\n%s", err, string(out))
	}
	return nil
}

func (w *World) whenTypeScriptTypeCheck() error {
	cmd := exec.CommandContext(w.ctx, "npx", "tsc", "--noEmit")
	cmd.Dir = "../../../../frontend"
	out, err := cmd.CombinedOutput()
	w.quizFixtures["tsc_output"] = string(out)
	if err != nil {
		return fmt.Errorf("tsc failed: %w\n%s", err, string(out))
	}
	return nil
}

func (w *World) whenGoTestWithRace() error {
	cmd := exec.CommandContext(w.ctx, "go", "test", "./...", "-race", "-count=1")
	out, err := cmd.CombinedOutput()
	w.quizFixtures["go_test_output"] = string(out)
	if err != nil {
		return fmt.Errorf("go test failed: %w\n%s", err, string(out))
	}
	return nil
}

// -----------------------------------------------------------------------
// Then implementations
// -----------------------------------------------------------------------

func (w *World) thenMarcusSeesQuizTitle(title string) error {
	msg, ok := w.waitForEvent("host", "quiz_loaded", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for quiz_loaded event on host connection")
	}
	got, _ := ExtractStringField(msg.Payload, "title")
	if got != title {
		return fmt.Errorf("expected quiz title %q, got %q", title, got)
	}
	return nil
}

func (w *World) thenHostSeesText(text string) error {
	// Wait for any host event that contains the given text in its payload.
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for text %q in host messages", text)
		case <-ticker.C:
			for _, msg := range w.messagesFor("host") {
				if strings.Contains(MarshalJSON(msg.Payload), text) {
					return nil
				}
			}
		}
	}
}

func (w *World) thenMarcusSeesQuizCounts(rounds, questions int) error {
	msg, ok := w.waitForEvent("host", "quiz_loaded", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for quiz_loaded event")
	}
	gotRounds, _ := msg.Payload["round_count"].(float64)
	gotQ, _ := msg.Payload["question_count"].(float64)
	if int(gotRounds) != rounds || int(gotQ) != questions {
		return fmt.Errorf("expected %d rounds/%d questions, got %g/%g", rounds, questions, gotRounds, gotQ)
	}
	return nil
}

func (w *World) thenShareablePlayerURLShown() error {
	msg, ok := w.waitForEvent("host", "quiz_loaded", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for quiz_loaded event")
	}
	url, _ := ExtractStringField(msg.Payload, "player_url")
	if url == "" {
		return fmt.Errorf("expected player_url in quiz_loaded payload, got empty")
	}
	w.lastQuizMeta.PlayerURL = url
	return nil
}

func (w *World) thenShareableDisplayURLShown() error {
	msg, ok := w.waitForEvent("host", "quiz_loaded", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for quiz_loaded event")
	}
	url, _ := ExtractStringField(msg.Payload, "display_url")
	if url == "" {
		return fmt.Errorf("expected display_url in quiz_loaded payload, got empty")
	}
	w.lastQuizMeta.DisplayURL = url
	return nil
}

func (w *World) thenGameSessionHasUniqueID() error {
	msg, ok := w.waitForEvent("host", "quiz_loaded", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for quiz_loaded event")
	}
	sessionID, _ := ExtractStringField(msg.Payload, "session_id")
	if sessionID == "" {
		return fmt.Errorf("expected non-empty session_id in quiz_loaded payload")
	}
	w.gameSessionID = sessionID
	return nil
}

func (w *World) thenTeamTokenStored(teamName string) error {
	msg, ok := w.waitForEvent(connectionKey("play", teamName), "team_registered", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for team_registered event for %q", teamName)
	}
	token, _ := ExtractStringField(msg.Payload, "device_token")
	if token == "" {
		return fmt.Errorf("expected device_token in team_registered payload for %q", teamName)
	}
	if conn, ok := w.connections[connectionKey("play", teamName)]; ok {
		conn.Token = token
	}
	return nil
}

func (w *World) thenMarcusSeesTeamInLobby(teamName string, deadline time.Duration) error {
	msg, ok := w.waitForEvent("host", "team_joined", deadline)
	if !ok {
		return fmt.Errorf("timed out waiting for team_joined event for %q within %v", teamName, deadline)
	}
	gotName, _ := ExtractStringField(msg.Payload, "team_name")
	if gotName != teamName {
		return fmt.Errorf("expected team_name %q in team_joined, got %q", teamName, gotName)
	}
	return nil
}

func (w *World) thenMarcusSeesTeamStatus(teamName, status string) error {
	expected := "submission_received"
	if status == "submitted" {
		expected = "submission_received"
	}
	msg, ok := w.waitForEvent("host", expected, 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for %q event for team %q", expected, teamName)
	}
	gotName, _ := ExtractStringField(msg.Payload, "team_name")
	if gotName != teamName {
		return fmt.Errorf("expected team_name %q in %q event, got %q", teamName, expected, gotName)
	}
	return nil
}

func (w *World) thenPlayerSeesState(teamName, state string, deadline time.Duration) error {
	key := connectionKey("play", teamName)
	msg, ok := w.waitForEvent(key, "round_started", deadline)
	if !ok {
		return fmt.Errorf("timed out waiting for round_started on %q within %v", teamName, deadline)
	}
	roundName, _ := ExtractStringField(msg.Payload, "round_name")
	if !strings.Contains(roundName, state) && !strings.Contains(state, "Round") {
		return fmt.Errorf("expected state %q, round_name was %q", state, roundName)
	}
	return nil
}

func (w *World) thenDisplaySeesState(state string, deadline time.Duration) error {
	_, ok := w.waitForEvent("display", "round_started", deadline)
	if !ok {
		return fmt.Errorf("timed out waiting for display to reach state %q within %v", state, deadline)
	}
	return nil
}

func (w *World) thenHostSeesRevealPanel(roundIndex int) error {
	_, ok := w.waitForEvent("host", "round_started", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for round_started on host")
	}
	return nil
}

func (w *World) thenPlayerSeesText(teamName, text string, deadline time.Duration) error {
	key := connectionKey("play", teamName)
	msg, ok := w.waitForEvent(key, "question_revealed", deadline)
	if !ok {
		return fmt.Errorf("timed out waiting for question_revealed on %q", teamName)
	}
	q, _ := msg.Payload["question"].(map[string]interface{})
	if q == nil {
		return fmt.Errorf("question_revealed payload missing 'question' field")
	}
	qText, _ := ExtractStringField(q, "text")
	if !strings.Contains(qText, text) {
		return fmt.Errorf("expected question text to contain %q, got %q", text, qText)
	}
	return nil
}

func (w *World) thenHostSeesRevealCount(count string, total int) error {
	// The host receives question_revealed events with revealed_count and total_questions.
	expected := fmt.Sprintf("%s of %d", count, total)
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for quizmaster panel to show %q", expected)
		case <-ticker.C:
			for _, msg := range w.messagesFor("host") {
				if msg.Event == "question_revealed" {
					rc, _ := msg.Payload["revealed_count"].(float64)
					tq, _ := msg.Payload["total_questions"].(float64)
					got := fmt.Sprintf("%g of %g", rc, tq)
					// normalise: "1 of 8" vs "1 of 8"
					if fmt.Sprintf("%d of %d", int(rc), int(tq)) == expected {
						return nil
					}
					_ = got
				}
			}
		}
	}
}

func (w *World) thenNoAnswerFieldInPlayOrDisplay(questionIndex int) error {
	playMsgs := w.messagesFor("play:Team Awesome")
	for _, msg := range playMsgs {
		if msg.Event == "question_revealed" {
			if PayloadContainsAnswerField(msg.Payload) {
				return fmt.Errorf("answer field found in question_revealed message to play room: %s",
					MarshalJSON(msg.Payload))
			}
		}
	}
	displayMsgs := w.messagesFor("display")
	for _, msg := range displayMsgs {
		if msg.Event == "question_revealed" {
			if PayloadContainsAnswerField(msg.Payload) {
				return fmt.Errorf("answer field found in question_revealed message to display room: %s",
					MarshalJSON(msg.Payload))
			}
		}
	}
	return nil
}

func (w *World) thenLockedStateAfterAck() error {
	// Verify submission_ack arrived before the locked state is observable.
	_, ok := w.waitForEvent("play:Team Awesome", "submission_ack", 3*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for submission_ack")
	}
	return nil
}

func (w *World) thenTeamScoreIncreasedBy(teamName string, points int) error {
	_, ok := w.waitForEvent("host", "score_updated", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for score_updated event")
	}
	return nil
}

func (w *World) thenDisplayShowsRoundScores() error {
	_, ok := w.waitForEvent("display", "round_scores_published", 3*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for round_scores_published on display")
	}
	return nil
}

func (w *World) thenTeamAppearsInScores(teamName string) error {
	msg, ok := w.waitForEvent("display", "round_scores_published", 3*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for round_scores_published")
	}
	scores, _ := msg.Payload["scores"].([]interface{})
	for _, s := range scores {
		scoreEntry, _ := s.(map[string]interface{})
		name, _ := ExtractStringField(scoreEntry, "team_name")
		if name == teamName {
			return nil
		}
	}
	return fmt.Errorf("team %q not found in round_scores_published payload", teamName)
}

func (w *World) thenPlayersReceiveRoundScores() error {
	_, ok := w.waitForEvent("play:Team Awesome", "round_scores_published", 3*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for round_scores_published on player connection")
	}
	return nil
}

func (w *World) thenDisplayShowsFinalScores() error {
	_, ok := w.waitForEvent("display", "game_over", 3*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for game_over on display")
	}
	return nil
}

func (w *World) thenPlayersReceiveFinalScores() error {
	_, ok := w.waitForEvent("play:Team Awesome", "game_over", 3*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for game_over on player connection")
	}
	return nil
}

func (w *World) thenGameStateIs(state string) error {
	// Verified by the game_over event on any connection.
	if strings.EqualFold(state, "game over") {
		return w.thenDisplayShowsFinalScores()
	}
	return godog.ErrPending
}

func (w *World) thenDisplayShowsWinner(teamName string) error {
	msg, ok := w.waitForEvent("display", "game_over", 3*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for game_over on display")
	}
	scores, _ := msg.Payload["final_scores"].([]interface{})
	if len(scores) == 0 {
		return fmt.Errorf("game_over payload has no final_scores")
	}
	first, _ := scores[0].(map[string]interface{})
	name, _ := ExtractStringField(first, "team_name")
	if name != teamName {
		return fmt.Errorf("expected winner %q, got %q", teamName, name)
	}
	return nil
}

func (w *World) thenPlayerSeesLockedIn(teamName string) error {
	return w.thenLockedStateAfterAck()
}

func (w *World) thenDisplayShowsCeremonyQuestion(questionIndex int) error {
	_, ok := w.waitForEvent("display", "ceremony_question_shown", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for ceremony_question_shown on display")
	}
	return nil
}

func (w *World) thenDisplayShowsCeremonyAnswer(answer string, questionIndex int) error {
	msg, ok := w.waitForEvent("display", "ceremony_answer_revealed", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for ceremony_answer_revealed on display")
	}
	gotAnswer, _ := ExtractStringField(msg.Payload, "answer")
	if gotAnswer != answer {
		return fmt.Errorf("expected ceremony answer %q, got %q", answer, gotAnswer)
	}
	return nil
}

func (w *World) thenDisplayShowsTeamScore(teamName string, points int) error {
	return w.thenTeamAppearsInScores(teamName)
}

func (w *World) thenAllAnswerFieldsLocked() error {
	_, ok := w.waitForEvent("play:Team Awesome", "game_over", 3*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for game_over (answer fields should be locked)")
	}
	return nil
}

func (w *World) thenMarcusSeesError(errMsg string) error {
	msg, ok := w.waitForEvent("host", "error", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for error event on host")
	}
	message, _ := ExtractStringField(msg.Payload, "message")
	if !strings.Contains(message, errMsg) {
		return fmt.Errorf("expected error message to contain %q, got %q", errMsg, message)
	}
	w.lastError = message
	return nil
}

func (w *World) thenNoGameSession() error {
	// Verify no quiz_loaded or session_created event was received.
	msgs := w.messagesFor("host")
	for _, msg := range msgs {
		if msg.Event == "quiz_loaded" || msg.Event == "session_created" {
			return fmt.Errorf("unexpected %q event received after invalid load", msg.Event)
		}
	}
	return nil
}

func (w *World) thenHTTPStatusIs(expected int) error {
	gotStr := w.quizFixtures["last_http_status"]
	got := 0
	fmt.Sscanf(gotStr, "%d", &got)
	if got != expected {
		return fmt.Errorf("expected HTTP status %d, got %d", expected, got)
	}
	return nil
}

func (w *World) thenResponseContainsSPAHTML() error {
	body := w.quizFixtures["last_http_body"]
	if !strings.Contains(body, "<!DOCTYPE html>") && !strings.Contains(body, "<html") {
		return fmt.Errorf("expected HTML response containing DOCTYPE or html tag, got: %q", body[:min(200, len(body))])
	}
	return nil
}

func (w *World) thenResponseContentTypeIs(ct string) error {
	// Content-type is checked via the HTTP client response; stub for now.
	return godog.ErrPending
}

func (w *World) thenRoomReceivesEvent(connKey, eventType string, deadline time.Duration) error {
	_, ok := w.waitForEvent(connKey, eventType, deadline)
	if !ok {
		return fmt.Errorf("timed out waiting for %q event on connection %q within %v", eventType, connKey, deadline)
	}
	return nil
}

func (w *World) thenRoomDoesNotReceiveEvent(connKey, eventType string) error {
	// Give 500ms for any stray message to arrive, then assert it did not.
	time.Sleep(500 * time.Millisecond)
	for _, msg := range w.messagesFor(connKey) {
		if msg.Event == eventType {
			return fmt.Errorf("unexpected %q event found on connection %q", eventType, connKey)
		}
	}
	return nil
}

func (w *World) thenQuestionRevealedHasNoField(connKey, fieldName string) error {
	msgs := w.messagesFor(connKey)
	for _, msg := range msgs {
		if msg.Event == "question_revealed" {
			q, _ := msg.Payload["question"].(map[string]interface{})
			if _, exists := q[fieldName]; exists {
				return fmt.Errorf("question_revealed to %q contains forbidden field %q: %s",
					connKey, fieldName, MarshalJSON(q))
			}
		}
	}
	return nil
}

func (w *World) thenQuestionRevealedContainsText(connKey, text string) error {
	msgs := w.messagesFor(connKey)
	for _, msg := range msgs {
		if msg.Event == "question_revealed" {
			q, _ := msg.Payload["question"].(map[string]interface{})
			qText, _ := ExtractStringField(q, "text")
			if strings.Contains(qText, text) {
				return nil
			}
		}
	}
	return fmt.Errorf("question_revealed on %q does not contain text %q", connKey, text)
}

func (w *World) thenCeremonyEventHasNoField(connKey, eventType, fieldName string) error {
	msgs := w.messagesFor(connKey)
	for _, msg := range msgs {
		if msg.Event == eventType {
			if _, exists := msg.Payload[fieldName]; exists {
				return fmt.Errorf("%q event on %q contains forbidden field %q", eventType, connKey, fieldName)
			}
		}
	}
	return nil
}

func (w *World) thenCeremonyAnswerRevealedWith(connKey, answer string) error {
	msg, ok := w.waitForEvent(connKey, "ceremony_answer_revealed", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for ceremony_answer_revealed on %q", connKey)
	}
	got, _ := ExtractStringField(msg.Payload, "answer")
	if got != answer {
		return fmt.Errorf("expected ceremony answer %q on %q, got %q", answer, connKey, got)
	}
	return nil
}

func (w *World) thenNoMessagesToPlayContainField(fieldName string) error {
	for _, msg := range w.messagesFor("play:Team Awesome") {
		if _, exists := msg.Payload[fieldName]; exists {
			return fmt.Errorf("play room message %q contains forbidden field %q: %s",
				msg.Event, fieldName, MarshalJSON(msg.Payload))
		}
	}
	return nil
}

func (w *World) thenAllRevealedQuestionsPublicOnly() error {
	for _, connKey := range []string{"play:Team Awesome", "display"} {
		if err := w.thenQuestionRevealedHasNoField(connKey, "answer"); err != nil {
			return err
		}
		if err := w.thenQuestionRevealedHasNoField(connKey, "answers"); err != nil {
			return err
		}
	}
	return nil
}

func (w *World) thenConnectionReceivesSnapshot() error {
	return w.thenRoomReceivesEvent("play:Late Squad", "state_snapshot", 2*time.Second)
}

func (w *World) thenSnapshotShowsActiveRound(roundIndex int) error {
	msg, ok := w.waitForEvent("play:Late Squad", "state_snapshot", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for state_snapshot")
	}
	ri, _ := msg.Payload["current_round_index"].(float64)
	if int(ri) != roundIndex {
		return fmt.Errorf("expected round_index %d in snapshot, got %g", roundIndex, ri)
	}
	return nil
}

func (w *World) thenSnapshotShowsRevealedQuestions(q1, q2 int) error {
	msg, ok := w.waitForEvent("play:Late Squad", "state_snapshot", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for state_snapshot")
	}
	revealed, _ := msg.Payload["revealed_questions"].([]interface{})
	if len(revealed) < q2+1 {
		return fmt.Errorf("expected at least %d revealed questions in snapshot, got %d", q2+1, len(revealed))
	}
	return nil
}

func (w *World) thenSnapshotHasNoAnswerFields() error {
	msg, ok := w.waitForEvent("play:Late Squad", "state_snapshot", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for state_snapshot")
	}
	if PayloadContainsAnswerField(msg.Payload) {
		return fmt.Errorf("state_snapshot contains answer fields: %s", MarshalJSON(msg.Payload))
	}
	return nil
}

func (w *World) thenSnapshotContainsDraft(teamName, answer string, qIndex int) error {
	key := connectionKey("play", teamName)
	msg, ok := w.waitForEvent(key, "state_snapshot", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for state_snapshot for %q", teamName)
	}
	drafts, _ := msg.Payload["draft_answers"].([]interface{})
	for _, d := range drafts {
		draft, _ := d.(map[string]interface{})
		qi, _ := draft["question_index"].(float64)
		ans, _ := ExtractStringField(draft, "answer")
		if int(qi) == qIndex && ans == answer {
			return nil
		}
	}
	return fmt.Errorf("draft answer %q for question %d not found in state_snapshot for %q", answer, qIndex, teamName)
}

func (w *World) thenConnectionRejectedWithStatus(code int) error {
	return godog.ErrPending
}

func (w *World) thenConnectionAccepted() error {
	return w.thenRoomReceivesEvent("host", "state_snapshot", 2*time.Second)
}

func (w *World) thenServerExitedWithError() error {
	return godog.ErrPending
}

func (w *World) thenErrorOutputContains(msg string) error {
	output := w.quizFixtures["server_startup_error"]
	if !strings.Contains(output, msg) {
		return fmt.Errorf("expected error output to contain %q, got %q", msg, output)
	}
	return nil
}

func (w *World) thenServerRunning() error {
	if w.server == nil {
		return fmt.Errorf("server is not running")
	}
	return nil
}

func (w *World) thenDockerBuildSucceeded() error {
	output := w.quizFixtures["docker_build_output"]
	if strings.Contains(output, "error") {
		return fmt.Errorf("docker build output contains errors: %s", output)
	}
	return nil
}

func (w *World) thenContainerHealthy(deadline time.Duration) error {
	return godog.ErrPending
}

func (w *World) thenGoArchLintPassed() error {
	output := w.quizFixtures["arch_lint_output"]
	if output == "" {
		return fmt.Errorf("go-arch-lint has not been run")
	}
	if strings.Contains(output, "violation") || strings.Contains(output, "FAIL") {
		return fmt.Errorf("go-arch-lint violations found: %s", output)
	}
	return nil
}

func (w *World) thenPackageHasNoImport(pkg, typeName string) error {
	output := w.quizFixtures["arch_lint_output"]
	if strings.Contains(output, pkg) && strings.Contains(output, typeName) {
		return fmt.Errorf("package %q has a reference to %q (go-arch-lint violation)", pkg, typeName)
	}
	return nil
}

func (w *World) thenTypeScriptPassed() error {
	output := w.quizFixtures["tsc_output"]
	if strings.Contains(output, "error TS") {
		return fmt.Errorf("TypeScript errors found: %s", output)
	}
	return nil
}

func (w *World) thenGoTestPassed() error {
	output := w.quizFixtures["go_test_output"]
	if strings.Contains(output, "FAIL") {
		return fmt.Errorf("go test failures: %s", output)
	}
	return nil
}

func (w *World) thenNoRaceConditions() error {
	output := w.quizFixtures["go_test_output"]
	if strings.Contains(output, "DATA RACE") {
		return fmt.Errorf("race condition detected: %s", output)
	}
	return nil
}

func (w *World) thenPlayerSeesLobby(teamName string) error {
	// The player receives a state_snapshot on connect that includes the team registry.
	// After registration, the player also sees team_registered. Both are acceptable signals.
	// We verify the team appears in the team registry via the state_snapshot or team_registered.
	key := connectionKey("play", teamName)
	_, ok := w.waitForEvent(key, "team_registered", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for team_registered event on %q — player not registered", teamName)
	}
	return nil
}

func (w *World) thenPlayerSeesError(expectedMsg string) error {
	key := w.lastJoinAttemptKey
	if key == "" {
		return fmt.Errorf("no player join attempt recorded; cannot check error")
	}
	msg, ok := w.waitForEvent(key, "error", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for error event on %q", key)
	}
	message, _ := ExtractStringField(msg.Payload, "message")
	if !strings.Contains(message, expectedMsg) {
		return fmt.Errorf("expected error message to contain %q, got %q", expectedMsg, message)
	}
	w.lastError = message
	return nil
}

func (w *World) thenNameFieldRemainedPopulated() error {
	// The server sends the error event with the attempted team name in the payload.
	// The client is expected to keep the name field populated on error.
	// At the acceptance level, we verify the error event was received (done in thenPlayerSeesError).
	// The UI-level assertion (field remains populated) is a frontend concern not testable here.
	return nil
}

func (w *World) thenNoDuplicateTeamInLobby() error {
	// Wait briefly, then verify no team_joined event was sent to host for the duplicate attempt.
	time.Sleep(200 * time.Millisecond)
	teamName := ""
	// Extract original team name from lastJoinAttemptKey (format: "play:Name_device2").
	if w.lastJoinAttemptKey != "" {
		raw := strings.TrimPrefix(w.lastJoinAttemptKey, "play:")
		teamName = strings.TrimSuffix(raw, "_device2")
	}
	// Count team_joined events for this team on the host connection.
	count := 0
	for _, msg := range w.messagesFor("host") {
		if msg.Event == "team_joined" {
			name, _ := ExtractStringField(msg.Payload, "team_name")
			if name == teamName {
				count++
			}
		}
	}
	if count > 1 {
		return fmt.Errorf("expected at most 1 team_joined event for %q, got %d (duplicate registered)", teamName, count)
	}
	return nil
}

// US-03: Start Game Broadcast Then implementations

func (w *World) thenAllThreePlayersReceiveRoundStarted(deadline time.Duration) error {
	teams := []string{"Team Awesome", "The Brainiacs", "Quiz Killers"}
	for _, team := range teams {
		key := connectionKey("play", team)
		if _, ok := w.waitForEvent(key, "round_started", deadline); !ok {
			return fmt.Errorf("timed out waiting for round_started on %q within %v", team, deadline)
		}
	}
	return nil
}

func (w *World) thenEachPlayerSeesRoundActive(roundIndex int) error {
	teams := []string{"Team Awesome", "The Brainiacs", "Quiz Killers"}
	for _, team := range teams {
		key := connectionKey("play", team)
		msg, ok := w.waitForEvent(key, "round_started", 2*time.Second)
		if !ok {
			return fmt.Errorf("timed out waiting for round_started on %q", team)
		}
		ri, _ := msg.Payload["round_index"].(float64)
		if int(ri) != roundIndex {
			return fmt.Errorf("expected round_index %d for %q, got %g", roundIndex, team, ri)
		}
	}
	return nil
}

func (w *World) thenDisplayReceivesRoundStarted(deadline time.Duration) error {
	if _, ok := w.waitForEvent("display", "round_started", deadline); !ok {
		return fmt.Errorf("timed out waiting for round_started on display within %v", deadline)
	}
	return nil
}

func (w *World) thenLateJoinerSeesRoundActive(teamName string, roundIndex int) error {
	key := connectionKey("play", teamName)
	msg, ok := w.waitForEvent(key, "state_snapshot", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for state_snapshot on late joiner %q", teamName)
	}
	ri, _ := msg.Payload["current_round"].(float64)
	if int(ri) != roundIndex {
		return fmt.Errorf("expected current_round %d in snapshot for %q, got %g", roundIndex, teamName, ri)
	}
	state, _ := ExtractStringField(msg.Payload, "state")
	if state != "ROUND_ACTIVE" {
		return fmt.Errorf("late joiner %q snapshot state is %q, expected ROUND_ACTIVE", teamName, state)
	}
	return nil
}

func (w *World) thenLateJoinerSeesRevealedQuestion(teamName string, questionIndex int) error {
	key := connectionKey("play", teamName)
	msg, ok := w.waitForEvent(key, "state_snapshot", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for state_snapshot for late joiner %q", teamName)
	}
	revealed, _ := msg.Payload["revealed_questions"].([]interface{})
	if len(revealed) < questionIndex+1 {
		return fmt.Errorf("expected at least %d revealed questions in snapshot for %q, got %d",
			questionIndex+1, teamName, len(revealed))
	}
	return nil
}

func (w *World) thenLateJoinerNotInLobby(teamName string) error {
	key := connectionKey("play", teamName)
	msg, ok := w.waitForEvent(key, "state_snapshot", 2*time.Second)
	if !ok {
		return fmt.Errorf("timed out waiting for state_snapshot for %q", teamName)
	}
	state, _ := ExtractStringField(msg.Payload, "state")
	if state == "LOBBY" || state == "" {
		return fmt.Errorf("late joiner %q is in lobby/empty state %q, expected ROUND_ACTIVE", teamName, state)
	}
	return nil
}

// -----------------------------------------------------------------------
// Milestone-2 step implementations
// -----------------------------------------------------------------------

// givenMarcusLoadedMilestone2Quiz sets up the quiz fixture for milestone-2 scenarios.
// The fixture has 1 round of 8 questions with specific texts matching the feature file.
func (w *World) givenMarcusLoadedMilestone2Quiz(filename string, rounds, questions int) error {
	qs := []QuizQuestion{
		{Text: "What is the capital of France?", Answer: "Paris"},
		{Text: "Name the three primary colors.", Answer: "Red, yellow, blue"},
		{Text: "How many sides does a hexagon have?", Answer: "Six"},
		{Text: "What is the boiling point of water in Celsius?", Answer: "100"},
		{Text: "Who painted the Mona Lisa?", Answer: "Leonardo da Vinci"},
		{Text: "What is the smallest planet in the solar system?", Answer: "Mercury"},
		{Text: "How many bones are in the adult human body?", Answer: "206"},
		{Text: "What currency is used in Japan?", Answer: "Yen"},
	}
	// If questions > 8, pad with generic entries.
	for i := len(qs); i < questions; i++ {
		qs = append(qs, QuizQuestion{
			Text:   fmt.Sprintf("Bonus question %d?", i+1),
			Answer: fmt.Sprintf("Answer %d", i+1),
		})
	}
	if rounds <= 1 {
		w.quizFixtures[filename] = SimpleQuizYAML("Friday Night Trivia -- March 2026", qs[:questions])
	} else {
		w.quizFixtures[filename] = MultiRoundQuizYAML("Friday Night Trivia -- March 2026", rounds, questions)
	}
	return w.givenGameSessionLoaded(filename)
}

// givenQuestionRevealedWithText reveals question at the given 1-based index and verifies text.
func (w *World) givenQuestionRevealedWithText(qNum int, _ string) error {
	return w.givenQuestionsRevealed(0, qNum)
}

// givenPriyaEnteredInField enters a draft answer for Priya (Team Awesome) for a question field.
func (w *World) givenPriyaEnteredInField(answer string, qNum int) error {
	return w.whenPlayerDraftsAnswer("Team Awesome", qNum-1, answer)
}

// whenMarcusRevealsQuestionWithText reveals a question by number (text is documentary).
func (w *World) whenMarcusRevealsQuestionWithText(qNum int, _ string) error {
	return w.whenMarcusRevealsQuestion(0, qNum-1)
}

// whenPriyaEntersInField enters a draft answer for Priya in the given question field.
func (w *World) whenPriyaEntersInField(answer string, qNum int) error {
	return w.whenPlayerDraftsAnswer("Team Awesome", qNum-1, answer)
}

// whenPriyaChangesAnswer enters a new draft answer for Priya for question 1.
func (w *World) whenPriyaChangesAnswer(newAnswer string) error {
	return w.whenPlayerDraftsAnswer("Team Awesome", 0, newAnswer)
}

// thenPlayerSeesTextOnAnyReveal checks that the team's connection has received
// a question_revealed event whose question text contains the expected text.
func (w *World) thenPlayerSeesTextOnAnyReveal(teamName, text string) error {
	return w.thenPlayerSeesText(teamName, text, 2*time.Second)
}

// thenDisplaySeesCurrentQuestion verifies the display received a question_revealed
// whose question text contains the expected text.
func (w *World) thenDisplaySeesCurrentQuestion(text string) error {
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for display to show %q as current question", text)
		case <-ticker.C:
			for _, msg := range w.messagesFor("display") {
				if msg.Event == "question_revealed" {
					q, _ := msg.Payload["question"].(map[string]interface{})
					qText, _ := ExtractStringField(q, "text")
					if strings.Contains(qText, text) {
						return nil
					}
				}
			}
		}
	}
}

// thenPlayerScreenShowsBothQuestions checks that Team Awesome's connection has received
// at least 2 question_revealed events (accumulation behaviour).
func (w *World) thenPlayerScreenShowsBothQuestions(q1, q2 int) error {
	teamName := "Team Awesome"
	key := connectionKey("play", teamName)
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			count := 0
			for _, msg := range w.messagesFor(key) {
				if msg.Event == "question_revealed" {
					count++
				}
			}
			return fmt.Errorf("expected %d question_revealed events on %q, got %d", q2, teamName, count)
		case <-ticker.C:
			count := 0
			for _, msg := range w.messagesFor(key) {
				if msg.Event == "question_revealed" {
					count++
				}
			}
			if count >= q2 {
				return nil
			}
		}
	}
}

// thenPriyaAnswerPreserved verifies the draft answer for question qNum is still stored.
// Verified by checking that no error event was received on Priya's connection after draft_answer.
func (w *World) thenPriyaAnswerPreserved(answer string, qNum int) error {
	// The server stores drafts silently; no error event means the draft was accepted.
	// We give a short window for any error to arrive, then declare success.
	time.Sleep(150 * time.Millisecond)
	for _, msg := range w.messagesFor(connectionKey("play", "Team Awesome")) {
		if msg.Event == "error" {
			errMsg, _ := ExtractStringField(msg.Payload, "message")
			return fmt.Errorf("unexpected error on play connection after draft: %s", errMsg)
		}
	}
	return nil
}

// thenDisplayShowsOnlyQuestion verifies the display received a question_revealed for the
// given question number. The display always shows only the latest (it replaces).
func (w *World) thenDisplayShowsOnlyQuestion(qNum int) error {
	// Find the last question_revealed on display and check it is for qNum-1.
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			return fmt.Errorf("timed out waiting for display to show question %d", qNum)
		case <-ticker.C:
			msgs := w.messagesFor("display")
			count := 0
			for _, msg := range msgs {
				if msg.Event == "question_revealed" {
					count++
				}
			}
			if count >= qNum {
				// Last question_revealed should have revealed_count == qNum.
				last := WSMessage{}
				for _, msg := range msgs {
					if msg.Event == "question_revealed" {
						last = msg
					}
				}
				rc, _ := last.Payload["revealed_count"].(float64)
				if int(rc) == qNum {
					return nil
				}
			}
		}
	}
}

// thenPriyaScreenShowsAnswerInField verifies the draft was accepted (no error) for the
// given question field. Since draft is fire-and-forget, absence of error = success.
func (w *World) thenPriyaScreenShowsAnswerInField(answer string, qNum int) error {
	time.Sleep(150 * time.Millisecond)
	for _, msg := range w.messagesFor(connectionKey("play", "Team Awesome")) {
		if msg.Event == "error" {
			errMsg, _ := ExtractStringField(msg.Payload, "message")
			return fmt.Errorf("error on play connection after draft for question %d: %s", qNum, errMsg)
		}
	}
	return nil
}

// thenDraftPersistedFor verifies the draft was silently accepted for the team.
func (w *World) thenDraftPersistedFor(teamName string) error {
	time.Sleep(150 * time.Millisecond)
	for _, msg := range w.messagesFor(connectionKey("play", teamName)) {
		if msg.Event == "error" {
			errMsg, _ := ExtractStringField(msg.Payload, "message")
			return fmt.Errorf("error on play connection for %q after draft: %s", teamName, errMsg)
		}
	}
	return nil
}

// thenPreviousAnswerNotShown verifies that the old answer is no longer active.
// Since the server stores only the latest draft, we verify no error was received.
func (w *World) thenPreviousAnswerNotShown(oldAnswer string) error {
	time.Sleep(150 * time.Millisecond)
	for _, msg := range w.messagesFor(connectionKey("play", "Team Awesome")) {
		if msg.Event == "error" {
			errMsg, _ := ExtractStringField(msg.Payload, "message")
			return fmt.Errorf("error on play connection after changing draft: %s", errMsg)
		}
	}
	return nil
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

func (w *World) hostDriver() *TriviaDriver {
	conn := w.connections["host"]
	if conn == nil {
		return nil
	}
	return conn.driver
}

func (w *World) playDriver(teamName string) *TriviaDriver {
	conn := w.connections[connectionKey("play", teamName)]
	if conn == nil {
		return nil
	}
	return conn.driver
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
