// steps.go registers all Gherkin step definitions with godog for the host-ui feature.
//
// Layer 2: Step methods translate business-language Gherkin phrases into
// calls on the HostUIDriver (Layer 3). No business logic lives here.
//
// Organisation:
//   - InitializeScenario: registers lifecycle hooks and all step patterns
//   - Given steps: arrange preconditions
//   - When steps: drive actions through the server's driving ports
//   - Then steps: assert observable outcomes
package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// InitializeScenario wires lifecycle hooks and all step definitions for host-ui scenarios.
func InitializeScenario(sc *godog.ScenarioContext) {
	w := newWorld()

	sc.Before(func(ctx context.Context, s *godog.Scenario) (context.Context, error) {
		return ctx, nil
	})

	sc.After(func(ctx context.Context, s *godog.Scenario, err error) (context.Context, error) {
		w.teardown()
		return ctx, nil
	})

	// -------------------------------------------------------------------------
	// Given steps — arrange preconditions
	// -------------------------------------------------------------------------

	sc.Step(`^the server is running with HOST_TOKEN "([^"]*)"$`,
		func(token string) error {
			return w.givenServerRunning(token)
		})

	sc.Step(`^a quiz file "([^"]*)" with (\d+) round of (\d+) text questions$`,
		func(filename string, rounds, questions int) error {
			return w.givenQuizFileExists(filename, rounds, questions)
		})

	sc.Step(`^a quiz file "([^"]*)" with (\d+) rounds of (\d+) text questions each$`,
		func(filename string, rounds, questions int) error {
			return w.givenQuizFileExistsMultiRound(filename, rounds, questions)
		})

	sc.Step(`^a quiz file "([^"]*)" with (\d+) rounds of (\d+) text questions each$`,
		func(filename string, rounds, questions int) error {
			return w.givenQuizFileExistsMultiRound(filename, rounds, questions)
		})

	sc.Step(`^Marcus opens the quizmaster panel with a valid host token$`,
		func() error {
			return w.givenMarcusConnectsToHostPanel()
		})

	sc.Step(`^Marcus has opened the quizmaster panel with a valid host token$`,
		func() error {
			return w.givenMarcusConnectsToHostPanel()
		})

	sc.Step(`^Marcus has opened the quizmaster panel with a valid token and is in Round (\d+)$`,
		func(roundNum int) error {
			return w.givenMarcusConnectedAndInRound(roundNum)
		})

	sc.Step(`^Marcus has opened the quizmaster panel with a valid token$`,
		func() error {
			return w.givenMarcusConnectsToHostPanel()
		})

	sc.Step(`^"([^"]*)" is connected in the lobby$`,
		func(teamName string) error {
			return w.givenTeamConnected(teamName)
		})

	sc.Step(`^"([^"]*)" is connected in the lobby$`,
		func(teamName string) error {
			return w.givenTeamConnected(teamName)
		})

	sc.Step(`^the display interface is connected$`,
		func() error {
			return w.givenDisplayConnected()
		})

	sc.Step(`^Marcus has loaded "([^"]*)" through the quizmaster panel$`,
		func(filename string) error {
			return w.givenMarcusLoadedQuiz(filename)
		})

	sc.Step(`^Marcus has started Round (\d+)$`,
		func(roundNum int) error {
			return w.givenMarcusStartedRound(roundNum - 1)
		})

	sc.Step(`^Marcus has started Round (\d+) and revealed (\d+) of (\d+) questions$`,
		func(roundNum, revealed, total int) error {
			if err := w.givenMarcusStartedRound(roundNum - 1); err != nil {
				return err
			}
			return w.givenQuestionsRevealed(roundNum-1, revealed)
		})

	sc.Step(`^Marcus has revealed all (\d+) questions in Round (\d+)$`,
		func(count, roundNum int) error {
			if err := w.givenMarcusStartedRound(roundNum - 1); err != nil {
				return err
			}
			return w.givenQuestionsRevealed(roundNum-1, count)
		})

	sc.Step(`^Round (\d+) has ended and all (\d+) questions have been revealed$`,
		func(roundNum, qCount int) error {
			return w.givenRoundEnded(roundNum-1, qCount)
		})

	sc.Step(`^"([^"]*)" has entered answers for all (\d+) questions$`,
		func(teamName string, count int) error {
			return w.givenTeamEnteredAnswers(teamName, count)
		})

	sc.Step(`^scoring is open for Round (\d+)$`,
		func(roundNum int) error {
			return w.givenScoringOpen(roundNum - 1)
		})

	sc.Step(`^"([^"]*)" submitted "([^"]*)" for question (\d+) \(correct answer: "([^"]*)"\)$`,
		func(teamName, answer string, qNum int, correct string) error {
			return w.givenTeamSubmittedAnswer(teamName, qNum-1, answer)
		})

	sc.Step(`^all answers for Round (\d+) have been marked correct or wrong$`,
		func(roundNum int) error {
			return w.givenAllAnswersMarked(roundNum - 1)
		})

	sc.Step(`^round scores have been published$`,
		func() error {
			return w.givenScoresPublished(0)
		})

	sc.Step(`^Round (\d+) has been fully played, scored, and ceremonialized$`,
		func(roundNum int) error {
			return w.givenRoundFullyComplete(roundNum - 1)
		})

	sc.Step(`^Round (\d+) has been fully played and both teams have equal scores$`,
		func(roundNum int) error {
			return w.givenRoundPlayedWithEqualScores(roundNum - 1)
		})

	sc.Step(`^Marcus is on the ceremony panel for Round (\d+)$`,
		func(roundNum int) error {
			return w.givenMarcusOnCeremonyPanel(roundNum - 1)
		})

	sc.Step(`^Marcus is showing question (\d+) on the ceremony panel$`,
		func(qNum int) error {
			return w.givenCeremonyShowingQuestion(qNum - 1)
		})

	sc.Step(`^Marcus has shown and revealed answers for all (\d+) questions in Round (\d+)$`,
		func(count, roundNum int) error {
			return w.givenCeremonyComplete(roundNum-1, count)
		})

	// -------------------------------------------------------------------------
	// When steps — drive actions
	// -------------------------------------------------------------------------

	sc.Step(`^Marcus loads "([^"]*)" through the quizmaster panel$`,
		func(filename string) error {
			return w.whenMarcusLoadsQuiz(filename)
		})

	sc.Step(`^Marcus attempts to load "([^"]*)" through the quizmaster panel$`,
		func(filePath string) error {
			return w.whenMarcusLoadsQuizByPath(filePath)
		})

	sc.Step(`^Marcus submits the load quiz form with an empty file path$`,
		func() error {
			return w.whenMarcusSubmitsEmptyFilePath()
		})

	sc.Step(`^the WebSocket handshake completes successfully$`,
		func() error {
			return w.whenWebSocketHandshakeCompletes()
		})

	sc.Step(`^Marcus opens the quizmaster panel with token "([^"]*)"$`,
		func(token string) error {
			return w.whenMarcusConnectsWithToken(token)
		})

	sc.Step(`^the WebSocket connection drops unexpectedly$`,
		func() error {
			return w.whenWebSocketDrops()
		})

	sc.Step(`^the WebSocket connection is restored$`,
		func() error {
			return w.whenWebSocketRestores()
		})

	sc.Step(`^the WebSocket fails to reconnect (\d+) consecutive times$`,
		func(count int) error {
			return w.whenWebSocketFailsToReconnect(count)
		})

	sc.Step(`^Marcus starts Round (\d+)$`,
		func(roundNum int) error {
			return w.whenMarcusStartsRound(roundNum - 1)
		})

	sc.Step(`^Marcus reveals question (\d+)$`,
		func(qNum int) error {
			return w.whenMarcusRevealsQuestion(0, qNum-1)
		})

	sc.Step(`^Marcus clicks "End Round"$`,
		func() error {
			return w.whenMarcusEndsRound()
		})

	sc.Step(`^Marcus ends Round (\d+)$`,
		func(roundNum int) error {
			return w.whenMarcusEndsRound()
		})

	sc.Step(`^Marcus marks "([^"]*)" answer for question (\d+) as correct$`,
		func(teamName string, qNum int) error {
			return w.whenMarcusMarksAnswer(teamName, 0, qNum-1, "correct")
		})

	sc.Step(`^Marcus marks "([^"]*)" answer for question (\d+) as wrong$`,
		func(teamName string, qNum int) error {
			return w.whenMarcusMarksAnswer(teamName, 0, qNum-1, "wrong")
		})

	sc.Step(`^Marcus publishes scores for Round (\d+)$`,
		func(roundNum int) error {
			return w.whenMarcusPublishesScores(roundNum - 1)
		})

	sc.Step(`^Marcus publishes scores for Round (\d+) without marking questions (\d+) and (\d+)$`,
		func(roundNum, q1, q2 int) error {
			return w.whenMarcusPublishesScores(roundNum - 1)
		})

	sc.Step(`^Marcus starts the answer ceremony$`,
		func() error {
			return w.whenMarcusStartsCeremony()
		})

	sc.Step(`^Marcus clicks "Show Next Question" for question (\d+)$`,
		func(qNum int) error {
			return w.whenMarcusShowsCeremonyQuestion(qNum - 1)
		})

	sc.Step(`^Marcus clicks "Reveal Answer" for question (\d+)$`,
		func(qNum int) error {
			return w.whenMarcusRevealsCeremonyAnswer(qNum - 1)
		})

	sc.Step(`^Marcus ends the game$`,
		func() error {
			return w.whenMarcusEndsGame()
		})

	sc.Step(`^Marcus ends the game before playing Round (\d+)$`,
		func(roundNum int) error {
			return w.whenMarcusEndsGame()
		})

	sc.Step(`^Marcus sends a mark-answer command before starting a round$`,
		func() error {
			return w.whenMarcusSendsMarkAnswerWithoutRound()
		})

	sc.Step(`^Marcus sends a start-round command with round index (\d+)$`,
		func(roundIndex int) error {
			return w.whenMarcusSendsStartRound(roundIndex)
		})

	sc.Step(`^Marcus sends a reveal-question command with question index (\d+) before revealing earlier questions$`,
		func(questionIndex int) error {
			return w.whenMarcusSendsRevealOutOfOrder(questionIndex)
		})

	sc.Step(`^Marcus dials the WebSocket endpoint with token "([^"]*)"$`,
		func(token string) error {
			return w.whenMarcusDialsWithWrongToken(token)
		})

	// -------------------------------------------------------------------------
	// Then steps — assert observable outcomes
	// -------------------------------------------------------------------------

	sc.Step(`^the quizmaster panel shows "Connected" status$`,
		func() error {
			return w.thenHostPanelShowsConnected()
		})

	sc.Step(`^the connection status shows "Connecting\.\.\." before the handshake is complete$`,
		func() error {
			return w.thenConnectionStatusConnecting()
		})

	sc.Step(`^the connection status shows "Connected"$`,
		func() error {
			return w.thenHostPanelShowsConnected()
		})

	sc.Step(`^the connection status shows "Disconnected"$`,
		func() error {
			return w.thenConnectionStatusDisconnected()
		})

	sc.Step(`^the connection status shows "Reconnecting\.\.\."$`,
		func() error {
			return w.thenConnectionStatusReconnecting()
		})

	sc.Step(`^the message "([^"]*)" is visible$`,
		func(msg string) error {
			return w.thenMessageVisible(msg)
		})

	sc.Step(`^no further connection attempts are made$`,
		func() error {
			return w.thenNoFurtherConnectionAttempts()
		})

	sc.Step(`^the load quiz form is visible with a file path input$`,
		func() error {
			return w.thenLoadQuizFormVisible()
		})

	sc.Step(`^the load quiz form is visible$`,
		func() error {
			return w.thenLoadQuizFormVisible()
		})

	sc.Step(`^a file path input labeled "Quiz file path" is visible$`,
		func() error {
			return w.thenFilePathInputVisible()
		})

	sc.Step(`^a "([^"]*)" button is visible$`,
		func(label string) error {
			return w.thenButtonVisible(label)
		})

	sc.Step(`^no "Start Round" button is visible$`,
		func() error {
			return w.thenStartRoundButtonNotVisible()
		})

	sc.Step(`^the quizmaster panel shows the quiz confirmation$`,
		func() error {
			return w.thenQuizConfirmationVisible()
		})

	sc.Step(`^the confirmation includes the quiz title and round count$`,
		func() error {
			return w.thenConfirmationIncludesTitleAndRoundCount()
		})

	sc.Step(`^the quizmaster panel shows "([^"]*)"$`,
		func(text string) error {
			return w.thenPanelShowsText(text)
		})

	sc.Step(`^the player join URL is displayed for Marcus to share$`,
		func() error {
			return w.thenPlayerURLDisplayed()
		})

	sc.Step(`^the player join URL is displayed$`,
		func() error {
			return w.thenPlayerURLDisplayed()
		})

	sc.Step(`^the display URL is displayed for Marcus to share$`,
		func() error {
			return w.thenDisplayURLDisplayed()
		})

	sc.Step(`^the display URL is displayed$`,
		func() error {
			return w.thenDisplayURLDisplayed()
		})

	sc.Step(`^the "([^"]*)" button is visible$`,
		func(label string) error {
			return w.thenButtonVisible(label)
		})

	sc.Step(`^the "([^"]*)" button is no longer visible$`,
		func(label string) error {
			return w.thenButtonNotVisible(label)
		})

	sc.Step(`^the round panel is visible showing "(\d+) of (\d+) revealed"$`,
		func(revealed, total int) error {
			return w.thenRoundPanelVisible(revealed, total)
		})

	sc.Step(`^the round panel shows "(\d+) of (\d+) revealed"$`,
		func(revealed, total int) error {
			return w.thenRoundPanelVisible(revealed, total)
		})

	sc.Step(`^the round panel shows "([^"]*)" and "(\d+) of (\d+) revealed"$`,
		func(roundName string, revealed, total int) error {
			return w.thenRoundPanelShowsNameAndCounter(roundName, revealed, total)
		})

	sc.Step(`^the "Reveal Next Question" button is visible$`,
		func() error {
			return w.thenRevealButtonVisible()
		})

	sc.Step(`^the "Reveal Next Question" button is no longer visible$`,
		func() error {
			return w.thenRevealButtonNotVisible()
		})

	sc.Step(`^the first revealed question appears in the question list$`,
		func() error {
			return w.thenFirstQuestionInList()
		})

	sc.Step(`^the revealed question list shows (\d+) questions in order$`,
		func(count int) error {
			return w.thenRevealedQuestionsCount(count)
		})

	sc.Step(`^the host panel receives confirmation that Round (\d+) has started$`,
		func(roundNum int) error {
			return w.thenRoundStartedConfirmed(roundNum - 1)
		})

	sc.Step(`^the host panel receives confirmation that Round (\d+) has ended$`,
		func(roundNum int) error {
			return w.thenRoundEndedConfirmed(roundNum - 1)
		})

	sc.Step(`^the scoring panel becomes visible$`,
		func() error {
			return w.thenScoringPanelVisible()
		})

	sc.Step(`^the scoring panel is visible$`,
		func() error {
			return w.thenScoringPanelVisible()
		})

	sc.Step(`^the scoring panel shows each question with its correct answer$`,
		func() error {
			return w.thenScoringPanelShowsCorrectAnswers()
		})

	sc.Step(`^the scoring panel shows submitted answers for "([^"]*)"$`,
		func(teamName string) error {
			return w.thenScoringPanelShowsTeamSubmissions(teamName)
		})

	sc.Step(`^"([^"]*)" submitted answers are listed under each question$`,
		func(teamName string) error {
			return w.thenScoringPanelShowsTeamSubmissions(teamName)
		})

	sc.Step(`^each team row has a "Correct" button and a "Wrong" button$`,
		func() error {
			return w.thenScoringRowsHaveVerdictButtons()
		})

	sc.Step(`^the running total for "([^"]*)" increases by (\d+) point$`,
		func(teamName string, points int) error {
			return w.thenRunningTotalIncreasedBy(teamName, points)
		})

	sc.Step(`^the running total for "([^"]*)" reflects (\d+) correct answers$`,
		func(teamName string, count int) error {
			return w.thenRunningTotalReflectsCorrectCount(teamName, count)
		})

	sc.Step(`^the running total for "([^"]*)" is unchanged$`,
		func(teamName string) error {
			return w.thenRunningTotalUnchanged(teamName)
		})

	sc.Step(`^the "Correct" button for "([^"]*)" on question (\d+) is visually marked as applied$`,
		func(teamName string, qNum int) error {
			return w.thenVerdictButtonMarked(teamName, qNum-1, "correct")
		})

	sc.Step(`^the "Wrong" button for "([^"]*)" on question (\d+) is visually marked as applied$`,
		func(teamName string, qNum int) error {
			return w.thenVerdictButtonMarked(teamName, qNum-1, "wrong")
		})

	sc.Step(`^the round score summary is shown to Marcus$`,
		func() error {
			return w.thenRoundScoreSummaryVisible()
		})

	sc.Step(`^the host panel accepts the publish without error$`,
		func() error {
			return w.thenPublishAcceptedWithoutError()
		})

	sc.Step(`^the ceremony panel is visible$`,
		func() error {
			return w.thenCeremonyPanelVisible()
		})

	sc.Step(`^the ceremony progress shows "Question (\d+) of (\d+) shown"$`,
		func(shown, total int) error {
			return w.thenCeremonyProgressShows(shown, total)
		})

	sc.Step(`^the display screen receives question (\d+)$`,
		func(qNum int) error {
			return w.thenDisplayReceivesQuestion(qNum - 1)
		})

	sc.Step(`^the display screen receives the answer for question (\d+)$`,
		func(qNum int) error {
			return w.thenDisplayReceivesAnswer(qNum - 1)
		})

	sc.Step(`^the play screen for "([^"]*)" does not receive the answer$`,
		func(teamName string) error {
			return w.thenPlayScreenDoesNotReceiveAnswer(teamName)
		})

	sc.Step(`^the final leaderboard is displayed$`,
		func() error {
			return w.thenFinalLeaderboardVisible()
		})

	sc.Step(`^"([^"]*)" appears on the leaderboard with their score$`,
		func(teamName string) error {
			return w.thenTeamOnLeaderboard(teamName)
		})

	sc.Step(`^the leaderboard shows all teams sorted by score from highest to lowest$`,
		func() error {
			return w.thenLeaderboardSortedDescending()
		})

	sc.Step(`^rank indicators \(1st, 2nd, etc\.\) are displayed next to each team$`,
		func() error {
			return w.thenRankIndicatorsDisplayed()
		})

	sc.Step(`^"([^"]*)" and "([^"]*)" appear at the same rank position$`,
		func(t1, t2 string) error {
			return w.thenTeamsAtSameRank(t1, t2)
		})

	sc.Step(`^game control buttons are no longer visible$`,
		func() error {
			return w.thenGameControlsRemoved()
		})

	sc.Step(`^the final leaderboard is displayed with scores from Round (\d+) only$`,
		func(roundNum int) error {
			return w.thenLeaderboardWithRoundScores(roundNum)
		})

	sc.Step(`^no error is shown$`,
		func() error {
			return w.thenNoErrorShown()
		})

	sc.Step(`^the round panel remains visible$`,
		func() error {
			return w.thenRoundPanelStillVisible()
		})

	sc.Step(`^game controls are available$`,
		func() error {
			return w.thenGameControlsAvailable()
		})

	sc.Step(`^a "Reload" button is visible$`,
		func() error {
			return w.thenReloadButtonVisible()
		})

	sc.Step(`^the game panel content is still visible beneath the overlay$`,
		func() error {
			return w.thenGamePanelVisibleBeneathOverlay()
		})

	sc.Step(`^the "Reveal Answer" button is now visible$`,
		func() error {
			return w.thenRevealAnswerButtonVisible()
		})

	sc.Step(`^the message "Ceremony complete" is visible$`,
		func() error {
			return w.thenMessageVisible("Ceremony complete")
		})

	sc.Step(`^no host_load_quiz command is sent to the server$`,
		func() error {
			return w.thenNoCommandSent("host_load_quiz")
		})

	sc.Step(`^the validation message "([^"]*)" is visible below the input$`,
		func(msg string) error {
			return w.thenValidationMessageVisible(msg)
		})

	sc.Step(`^an error message is displayed below the file path input$`,
		func() error {
			return w.thenLoadErrorDisplayed()
		})

	sc.Step(`^an error message is shown below the file path input$`,
		func() error {
			return w.thenLoadErrorDisplayed()
		})

	sc.Step(`^the file path input remains editable$`,
		func() error {
			return w.thenFilePathInputEditable()
		})

	sc.Step(`^no round controls appear$`,
		func() error {
			return w.thenNoRoundControlsVisible()
		})

	sc.Step(`^the server sends an error event in response$`,
		func() error {
			return w.thenServerSentError()
		})

	sc.Step(`^the quizmaster panel remains in the quiz-loaded state$`,
		func() error {
			return w.thenPanelInQuizLoadedState()
		})

	sc.Step(`^the host panel receives a quiz confirmation with (\d+) rounds and (\d+) questions$`,
		func(rounds, questions int) error {
			return w.thenHostReceivedQuizLoaded(rounds, questions)
		})

	sc.Step(`^the WebSocket dial is refused with an abnormal close$`,
		func() error {
			return w.thenWebSocketDialRefused()
		})

	sc.Step(`^no messages are received on the connection$`,
		func() error {
			return w.thenNoMessagesReceived()
		})
}
