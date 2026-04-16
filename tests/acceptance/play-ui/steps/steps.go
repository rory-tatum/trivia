// steps.go registers all Gherkin step definitions with godog for the play-ui feature.
//
// Layer 2: Step methods translate business-language Gherkin phrases into
// calls on the PlayUIDriver (Layer 3). No business logic lives here.
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

// InitializeScenario wires lifecycle hooks and all step definitions for play-ui scenarios.
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
		func(filename string, rounds, questionsPerRound int) error {
			return w.givenQuizFileExistsMultiRound(filename, rounds, questionsPerRound)
		})

	sc.Step(`^a quiz file "([^"]*)" with (\d+) round including a multiple choice question$`,
		func(filename string, rounds int) error {
			return w.givenQuizFileWithMultipleChoice(filename)
		})

	sc.Step(`^a quiz file "([^"]*)" with (\d+) round including a multi-part question$`,
		func(filename string, rounds int) error {
			return w.givenQuizFileWithMultiPart(filename)
		})

	sc.Step(`^a quiz file "([^"]*)" with (\d+) round including an image question$`,
		func(filename string, rounds int) error {
			return w.givenQuizFileWithMedia(filename)
		})

	sc.Step(`^the quizmaster has loaded "([^"]*)"$`,
		func(filename string) error {
			return w.givenQuizmasterLoadedQuiz(filename)
		})

	sc.Step(`^the quizmaster has started Round (\d+) with "([^"]*)" in the play room$`,
		func(roundNum int, teamName string) error {
			return w.givenRoundStartedWithTeam(roundNum-1, teamName)
		})

	sc.Step(`^the quizmaster has started Round (\d+) and revealed (\d+) of (\d+) questions$`,
		func(roundNum, revealed, total int) error {
			return w.givenRoundStartedAndQuestionsRevealed(roundNum-1, revealed)
		})

	sc.Step(`^the quizmaster has started Round (\d+) and ended Round (\d+) with "([^"]*)" in the play room$`,
		func(startRound, endRound int, teamName string) error {
			return w.givenRoundEndedWithTeam(startRound-1, teamName)
		})

	sc.Step(`^the quizmaster has started Round (\d+) and ended Round (\d+)$`,
		func(startRound, endRound int) error {
			return w.givenRoundEnded(startRound - 1)
		})

	sc.Step(`^the quizmaster has revealed all (\d+) questions$`,
		func(count int) error {
			return w.givenAllQuestionsRevealed(count)
		})

	sc.Step(`^"([^"]*)" is already registered in the game$`,
		func(teamName string) error {
			return w.givenTeamAlreadyRegistered(teamName)
		})

	sc.Step(`^"([^"]*)" is registered in the play room$`,
		func(teamName string) error {
			return w.givenTeamRegistered(teamName)
		})

	sc.Step(`^"([^"]*)" and "([^"]*)" are registered in the play room$`,
		func(team1, team2 string) error {
			return w.givenTwoTeamsRegistered(team1, team2)
		})

	sc.Step(`^"([^"]*)" has registered and Round (\d+) is active with (\d+) questions revealed$`,
		func(teamName string, roundNum, questionCount int) error {
			return w.givenTeamRegisteredAndRoundActiveWithQuestions(teamName, roundNum-1, questionCount)
		})

	sc.Step(`^"([^"]*)" has saved a draft answer "([^"]*)" for Round (\d+) question (\d+)$`,
		func(teamName, answer string, roundNum, qNum int) error {
			return w.givenTeamSavedDraft(teamName, roundNum-1, qNum-1, answer)
		})

	sc.Step(`^"([^"]*)" has already submitted their Round (\d+) answers$`,
		func(teamName string, roundNum int) error {
			return w.givenTeamSubmitted(teamName, roundNum-1)
		})

	sc.Step(`^"([^"]*)" has submitted for Round (\d+) and the quizmaster has started the ceremony$`,
		func(teamName string, roundNum int) error {
			return w.givenTeamSubmittedAndCeremonyStarted(teamName, roundNum-1)
		})

	sc.Step(`^"([^"]*)" has submitted for Round (\d+) and the ceremony is at question (\d+)$`,
		func(teamName string, roundNum, qNum int) error {
			return w.givenTeamSubmittedAndCeremonyAtQuestion(teamName, roundNum-1, qNum-1)
		})

	sc.Step(`^"([^"]*)" has a valid device token from earlier in the game$`,
		func(teamName string) error {
			return w.givenTeamHasValidToken(teamName)
		})

	sc.Step(`^"([^"]*)" and "([^"]*)" have completed Round (\d+) with the quizmaster scoring$`,
		func(team1, team2 string, roundNum int) error {
			return w.givenTwoTeamsCompletedRoundWithScoring(team1, team2, roundNum-1)
		})

	sc.Step(`^"([^"]*)" has completed Round (\d+) with the quizmaster scoring$`,
		func(teamName string, roundNum int) error {
			return w.givenTeamCompletedRoundWithScoring(teamName, roundNum-1)
		})

	sc.Step(`^"([^"]*)" and "([^"]*)" have completed Round (\d+) and scores are published$`,
		func(team1, team2 string, roundNum int) error {
			return w.givenTwoTeamsRoundCompleteAndScoresPublished(team1, team2, roundNum-1)
		})

	sc.Step(`^"([^"]*)" is on the Round (\d+) scores screen$`,
		func(teamName string, roundNum int) error {
			return w.givenTeamOnScoresScreen(teamName, roundNum-1)
		})

	sc.Step(`^the game is in the ceremony phase for Round (\d+)$`,
		func(roundNum int) error {
			return w.givenGameInCeremonyPhase(roundNum - 1)
		})

	sc.Step(`^the game is at the Round (\d+) scores screen with all teams having submitted$`,
		func(roundNum int) error {
			return w.givenGameAtScoresScreen(roundNum - 1)
		})

	sc.Step(`^"([^"]*)" is registered with Round (\d+) active and (\d+) questions revealed$`,
		func(teamName string, roundNum, questionCount int) error {
			return w.givenTeamRegisteredRoundActiveQuestionsRevealed(teamName, roundNum-1, questionCount)
		})

	// -------------------------------------------------------------------------
	// When steps — drive actions
	// -------------------------------------------------------------------------

	sc.Step(`^"([^"]*)" connects to the play room$`,
		func(teamName string) error {
			return w.whenTeamConnects(teamName)
		})

	sc.Step(`^"([^"]*)" registers as a new team$`,
		func(teamName string) error {
			return w.whenTeamRegisters(teamName)
		})

	sc.Step(`^"([^"]*)" connects to the play room and registers$`,
		func(teamName string) error {
			return w.whenTeamConnectsAndRegisters(teamName)
		})

	sc.Step(`^"([^"]*)" attempts to register from a second device$`,
		func(teamName string) error {
			return w.whenTeamAttemptsRegistrationFromSecondDevice(teamName)
		})

	sc.Step(`^a player attempts to register with an empty team name$`,
		func() error {
			return w.whenPlayerAttemptsEmptyRegistration()
		})

	sc.Step(`^"([^"]*)" reconnects with their stored device token$`,
		func(teamName string) error {
			return w.whenTeamReconnectsWithStoredToken(teamName)
		})

	sc.Step(`^a player attempts to rejoin with an unrecognised device token$`,
		func() error {
			return w.whenPlayerAttemptsBadRejoin()
		})

	sc.Step(`^a player sends a rejoin request with a token that does not match any registered team$`,
		func() error {
			return w.whenPlayerAttemptsBadRejoin()
		})

	sc.Step(`^"([^"]*)" requests a game state snapshot$`,
		func(teamName string) error {
			return w.whenTeamRequestsSnapshot(teamName)
		})

	sc.Step(`^"([^"]*)" saves a draft answer "([^"]*)" for Round (\d+) question (\d+)$`,
		func(teamName, answer string, roundNum, qNum int) error {
			return w.whenTeamSavesDraft(teamName, roundNum-1, qNum-1, answer)
		})

	sc.Step(`^"([^"]*)" submits their Round (\d+) answers$`,
		func(teamName string, roundNum int) error {
			return w.whenTeamSubmitsAnswers(teamName, roundNum-1)
		})

	sc.Step(`^"([^"]*)" submits Round (\d+) answers with all fields empty$`,
		func(teamName string, roundNum int) error {
			return w.whenTeamSubmitsBlankAnswers(teamName, roundNum-1)
		})

	sc.Step(`^"([^"]*)" attempts to submit answers for Round (\d+) before the round starts$`,
		func(teamName string, roundNum int) error {
			return w.whenTeamAttemptsSubmitBeforeRound(teamName, roundNum-1)
		})

	sc.Step(`^a player attempts to submit answers using an unknown team identifier$`,
		func() error {
			return w.whenPlayerAttemptsSubmitWithUnknownID()
		})

	sc.Step(`^the quizmaster starts Round (\d+)$`,
		func(roundNum int) error {
			return w.whenQuizmasterStartsRound(roundNum - 1)
		})

	sc.Step(`^the quizmaster reveals question (\d+)$`,
		func(qNum int) error {
			return w.whenQuizmasterRevealsQuestion(0, qNum-1)
		})

	sc.Step(`^the quizmaster reveals the multiple choice question$`,
		func() error {
			return w.whenQuizmasterRevealsQuestion(0, 0)
		})

	sc.Step(`^the quizmaster reveals the multi-part question$`,
		func() error {
			return w.whenQuizmasterRevealsQuestion(0, 0)
		})

	sc.Step(`^the quizmaster reveals the image question$`,
		func() error {
			return w.whenQuizmasterRevealsQuestion(0, 0)
		})

	sc.Step(`^the quizmaster ends Round (\d+)$`,
		func(roundNum int) error {
			return w.whenQuizmasterEndsRound(roundNum - 1)
		})

	sc.Step(`^the quizmaster shows the ceremony question (\d+)$`,
		func(qNum int) error {
			return w.whenQuizmasterShowsCeremonyQuestion(qNum - 1)
		})

	sc.Step(`^the quizmaster reveals the answer for ceremony question (\d+)$`,
		func(qNum int) error {
			return w.whenQuizmasterRevealsCeremonyAnswer(qNum - 1)
		})

	sc.Step(`^the quizmaster shows and reveals the answer for ceremony question (\d+)$`,
		func(qNum int) error {
			return w.whenQuizmasterShowsAndRevealsCeremonyQuestion(qNum - 1)
		})

	sc.Step(`^the quizmaster publishes Round (\d+) scores$`,
		func(roundNum int) error {
			return w.whenQuizmasterPublishesScores(roundNum - 1)
		})

	sc.Step(`^the quizmaster ends the game$`,
		func() error {
			return w.whenQuizmasterEndsGame()
		})

	// -------------------------------------------------------------------------
	// Then steps — assert observable outcomes
	// -------------------------------------------------------------------------

	sc.Step(`^"([^"]*)" receives their team identity$`,
		func(teamName string) error {
			return w.thenTeamReceivesIdentity(teamName)
		})

	sc.Step(`^the team identity includes a team identifier and a device token$`,
		func() error {
			return w.thenTeamIdentityHasBothTokens()
		})

	sc.Step(`^"([^"]*)" is in the lobby waiting for the round to start$`,
		func(teamName string) error {
			return w.thenTeamInLobby(teamName)
		})

	sc.Step(`^"([^"]*)" receives the round started notification$`,
		func(teamName string) error {
			return w.thenTeamReceivesRoundStarted(teamName)
		})

	sc.Step(`^"([^"]*)" receives the round started notification for Round (\d+)$`,
		func(teamName string, roundNum int) error {
			return w.thenTeamReceivesRoundStartedForRound(teamName, roundNum-1)
		})

	sc.Step(`^"([^"]*)" sees that Round (\d+) has begun with (\d+) questions to answer$`,
		func(teamName string, roundNum, questionCount int) error {
			return w.thenTeamSeesRoundBegun(teamName, roundNum-1, questionCount)
		})

	sc.Step(`^"([^"]*)" receives the first question on their device$`,
		func(teamName string) error {
			return w.thenTeamReceivesFirstQuestion(teamName)
		})

	sc.Step(`^"([^"]*)" has received all (\d+) questions for Round (\d+)$`,
		func(teamName string, count, roundNum int) error {
			return w.thenTeamHasReceivedAllQuestions(teamName, count)
		})

	sc.Step(`^"([^"]*)" receives the question on their device$`,
		func(teamName string) error {
			return w.thenTeamReceivesQuestion(teamName)
		})

	sc.Step(`^the question includes the question text$`,
		func() error {
			return w.thenQuestionHasText()
		})

	sc.Step(`^"([^"]*)" receives the round ended notification$`,
		func(teamName string) error {
			return w.thenTeamReceivesRoundEnded(teamName)
		})

	sc.Step(`^the round ended notification includes the round number$`,
		func() error {
			return w.thenRoundEndedHasRoundNumber()
		})

	sc.Step(`^"([^"]*)" receives confirmation that their answers are locked in$`,
		func(teamName string) error {
			return w.thenTeamReceivesSubmissionAck(teamName)
		})

	sc.Step(`^the confirmation shows the answers are locked for Round (\d+)$`,
		func(roundNum int) error {
			return w.thenSubmissionAckShowsLocked(roundNum - 1)
		})

	sc.Step(`^the play room receives a notification that "([^"]*)" has submitted$`,
		func(teamName string) error {
			return w.thenPlayRoomReceivesSubmissionNotification(teamName)
		})

	sc.Step(`^"([^"]*)" receives a notification that "([^"]*)" has submitted$`,
		func(observerTeam, submittingTeam string) error {
			return w.thenTeamReceivesOtherTeamSubmissionNotification(observerTeam, submittingTeam)
		})

	sc.Step(`^the notification includes "([^"]*)" team name$`,
		func(teamName string) error {
			return w.thenNotificationIncludesTeamName(teamName)
		})

	sc.Step(`^"([^"]*)" receives a submission notification in the play room$`,
		func(teamName string) error {
			return w.thenTeamReceivesOwnSubmissionNotification(teamName)
		})

	sc.Step(`^the notification includes "([^"]*)" team name and round number$`,
		func(teamName string) error {
			return w.thenNotificationIncludesTeamNameAndRound(teamName)
		})

	sc.Step(`^"([^"]*)" receives the ceremony question on their device$`,
		func(teamName string) error {
			return w.thenTeamReceivesCeremonyQuestion(teamName)
		})

	sc.Step(`^the ceremony question includes the question text$`,
		func() error {
			return w.thenCeremonyQuestionHasText()
		})

	sc.Step(`^"([^"]*)" receives the revealed answer with team verdicts$`,
		func(teamName string) error {
			return w.thenTeamReceivesCeremonyAnswer(teamName)
		})

	sc.Step(`^the verdicts show whether each team answered correctly$`,
		func() error {
			return w.thenVerdictsShowTeamResults()
		})

	sc.Step(`^the verdicts include a result for "([^"]*)"$`,
		func(teamName string) error {
			return w.thenVerdictsIncludeTeam(teamName)
		})

	sc.Step(`^each verdict shows whether the team answered correctly or not$`,
		func() error {
			return w.thenEachVerdictHasResult()
		})

	sc.Step(`^"([^"]*)" has received (\d+) ceremony question events$`,
		func(teamName string, count int) error {
			return w.thenTeamHasReceivedCeremonyQuestionCount(teamName, count)
		})

	sc.Step(`^"([^"]*)" has received (\d+) answer reveal events$`,
		func(teamName string, count int) error {
			return w.thenTeamHasReceivedAnswerRevealCount(teamName, count)
		})

	sc.Step(`^"([^"]*)" receives the round scores with each team's name and total$`,
		func(teamName string) error {
			return w.thenTeamReceivesRoundScores(teamName)
		})

	sc.Step(`^the scores list includes team names alongside each score$`,
		func() error {
			return w.thenScoresListHasTeamNames()
		})

	sc.Step(`^the scores list includes "([^"]*)" with their round score and running total$`,
		func(teamName string) error {
			return w.thenScoresListIncludesTeam(teamName)
		})

	sc.Step(`^the scores list includes "([^"]*)" with their round score$`,
		func(teamName string) error {
			return w.thenScoresListIncludesTeamRoundScore(teamName)
		})

	sc.Step(`^"([^"]*)" receives the final scores notification$`,
		func(teamName string) error {
			return w.thenTeamReceivesFinalScores(teamName)
		})

	sc.Step(`^the final scores include team names and totals for all teams$`,
		func() error {
			return w.thenFinalScoresHaveTeamNames()
		})

	sc.Step(`^"([^"]*)" receives the current game state$`,
		func(teamName string) error {
			return w.thenTeamReceivesGameState(teamName)
		})

	sc.Step(`^the game state indicates the game is in the lobby$`,
		func() error {
			return w.thenGameStateIsLobby()
		})

	sc.Step(`^"([^"]*)" receives a game state snapshot$`,
		func(teamName string) error {
			return w.thenTeamReceivesStateSnapshot(teamName)
		})

	sc.Step(`^the snapshot includes "([^"]*)"'s previously saved draft answers$`,
		func(teamName string) error {
			return w.thenSnapshotHasDraftAnswers(teamName)
		})

	sc.Step(`^the snapshot shows the game is in round active state$`,
		func() error {
			return w.thenSnapshotShowsRoundActive()
		})

	sc.Step(`^the snapshot shows the game is in ceremony state$`,
		func() error {
			return w.thenSnapshotShowsCeremony()
		})

	sc.Step(`^the snapshot shows the game is in the round scores phase$`,
		func() error {
			return w.thenSnapshotShowsRoundScores()
		})

	sc.Step(`^the snapshot for "([^"]*)" includes all (\d+) revealed questions$`,
		func(teamName string, count int) error {
			return w.thenSnapshotHasRevealedQuestions(teamName, count)
		})

	sc.Step(`^the snapshot shows (\d+) revealed questions for Round (\d+)$`,
		func(count, roundNum int) error {
			return w.thenSnapshotShowsRevealedQuestions(count)
		})

	sc.Step(`^the state snapshot contains "([^"]*)"'s draft answer for Round (\d+) question (\d+)$`,
		func(teamName string, roundNum, qNum int) error {
			return w.thenSnapshotContainsDraftForQuestion(teamName, roundNum-1, qNum-1)
		})

	sc.Step(`^the state snapshot for "([^"]*)" contains no draft answers$`,
		func(teamName string) error {
			return w.thenSnapshotHasNoDraftAnswers(teamName)
		})

	sc.Step(`^the second device receives a name-already-taken error$`,
		func() error {
			return w.thenSecondDeviceReceivesDuplicateError()
		})

	sc.Step(`^they receive a team-not-found error$`,
		func() error {
			return w.thenReceivesTeamNotFoundError()
		})

	sc.Step(`^the player receives a team-not-found error$`,
		func() error {
			return w.thenReceivesTeamNotFoundError()
		})

	sc.Step(`^no game state snapshot is sent$`,
		func() error {
			return w.thenNoStateSnapshotSent()
		})

	sc.Step(`^"([^"]*)" receives an already-submitted error$`,
		func(teamName string) error {
			return w.thenTeamReceivesAlreadySubmittedError(teamName)
		})

	sc.Step(`^"([^"]*)" receives an error response$`,
		func(teamName string) error {
			return w.thenTeamReceivesErrorResponse(teamName)
		})

	sc.Step(`^the player receives an error response$`,
		func() error {
			return w.thenAnonymousPlayerReceivesError()
		})

	sc.Step(`^no team identity is issued$`,
		func() error {
			return w.thenNoTeamIdentityIssued()
		})

	sc.Step(`^no error is returned to "([^"]*)"$`,
		func(teamName string) error {
			return w.thenNoErrorReturnedToTeam(teamName)
		})

	sc.Step(`^the draft is saved on the server without error$`,
		func() error {
			return w.thenDraftSavedWithoutError()
		})

	sc.Step(`^"([^"]*)" receives the question with a non-empty list of answer choices$`,
		func(teamName string) error {
			return w.thenQuestionHasChoices(teamName)
		})

	sc.Step(`^the choices list contains (\d+) options$`,
		func(count int) error {
			return w.thenChoicesListHasCount(count)
		})

	sc.Step(`^"([^"]*)" receives the question with no choices list$`,
		func(teamName string) error {
			return w.thenQuestionHasNoChoices(teamName)
		})

	sc.Step(`^"([^"]*)" receives the question with the multi-part indicator set$`,
		func(teamName string) error {
			return w.thenQuestionHasMultiPartIndicator(teamName)
		})

	sc.Step(`^"([^"]*)" receives the question without the multi-part indicator$`,
		func(teamName string) error {
			return w.thenQuestionHasNoMultiPartIndicator(teamName)
		})

	sc.Step(`^"([^"]*)" receives the question with a media reference$`,
		func(teamName string) error {
			return w.thenQuestionHasMediaReference(teamName)
		})

	sc.Step(`^the media reference includes the media type and a URL$`,
		func() error {
			return w.thenMediaReferenceHasTypeAndURL()
		})

	sc.Step(`^"([^"]*)" receives the question with no media attachment$`,
		func(teamName string) error {
			return w.thenQuestionHasNoMedia(teamName)
		})

	sc.Step(`^the connection is accepted and "([^"]*)" receives a state snapshot$`,
		func(teamName string) error {
			return w.thenConnectionAcceptedWithSnapshot(teamName)
		})

	sc.Step(`^"([^"]*)" receives the round scores notification$`,
		func(teamName string) error {
			return w.thenTeamReceivesRoundScores(teamName)
		})

	sc.Step(`^the scores notification includes a structured list with a team name in each entry$`,
		func() error {
			return w.thenScoresListHasTeamNames()
		})

	sc.Step(`^the verdicts list is present in the answer notification$`,
		func() error {
			return w.thenVerdictsListPresent()
		})
}
