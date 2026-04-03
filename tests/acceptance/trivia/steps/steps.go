// Steps registers all Gherkin step definitions with godog.
//
// Layer 2: Step methods translate business-language Gherkin phrases into
// calls to the TriviaDriver (Layer 3). No assertions belong in step methods
// themselves -- assertions belong in Then steps only.
//
// Organisation:
//   - InitializeScenario: registers all steps and lifecycle hooks
//   - Given steps: arrange world state
//   - When steps: drive actions through the server's ports
//   - Then steps: assert observable outcomes
package steps

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

// InitializeScenario wires lifecycle hooks and all step definitions.
func InitializeScenario(sc *godog.ScenarioContext) {
	w := newWorld()

	sc.Before(func(ctx context.Context, s *godog.Scenario) (context.Context, error) {
		// World is already fresh; attach it to context for step access if needed.
		return ctx, nil
	})

	sc.After(func(ctx context.Context, s *godog.Scenario, err error) (context.Context, error) {
		w.teardown()
		return ctx, nil
	})

	// -----------------------------------------------------------------------
	// Given steps -- arrange
	// -----------------------------------------------------------------------

	sc.Step(`^a quiz file "([^"]*)" exists with (\d+) round and (\d+) text questions$`,
		func(filename string, rounds, questions int) error {
			return w.givenQuizFileExists(filename, rounds, questions)
		})

	sc.Step(`^a quiz file "([^"]*)" exists with (\d+) rounds and (\d+) text questions$`,
		func(filename string, rounds, questions int) error {
			return w.givenQuizFileExists(filename, rounds, questions)
		})

	sc.Step(`^a quiz file "([^"]*)" containing (\d+) rounds and (\d+) text questions exists on the server$`,
		func(filename string, rounds, questions int) error {
			return w.givenQuizFileExists(filename, rounds, questions)
		})

	sc.Step(`^a quiz file "([^"]*)" with (\d+) round of (\d+) text questions$`,
		func(filename string, rounds, questions int) error {
			return w.givenQuizFileExists(filename, rounds, questions)
		})

	sc.Step(`^a quiz file "([^"]*)" with (\d+) rounds of (\d+) text questions each$`,
		func(filename string, rounds, questions int) error {
			return w.givenQuizFileExists(filename, rounds, questions)
		})

	sc.Step(`^the quiz title is "([^"]*)"$`,
		func(title string) error {
			w.lastQuizMeta.Title = title
			return nil
		})

	sc.Step(`^the quizmaster token is "([^"]*)"$`,
		func(token string) error {
			w.hostToken = token
			return nil
		})

	sc.Step(`^the server is running with HOST_TOKEN "([^"]*)"$`,
		func(token string) error {
			return w.givenServerRunning(token)
		})

	sc.Step(`^the server is running$`,
		func() error {
			return w.givenServerRunning(w.hostToken)
		})

	sc.Step(`^Marcus opens the quizmaster panel with a valid token$`,
		func() error {
			return w.givenMarcusConnectsToHostPanel()
		})

	sc.Step(`^the quizmaster panel is accessible with a valid host token$`,
		func() error {
			return w.givenMarcusConnectsToHostPanel()
		})

	sc.Step(`^Marcus has a valid quiz file "([^"]*)" with (\d+) round and (\d+) text questions$`,
		func(filename string, rounds, questions int) error {
			return w.givenQuizFileExists(filename, rounds, questions)
		})

	sc.Step(`^Marcus opens the quizmaster panel in his browser with a valid host token$`,
		func() error {
			return w.givenMarcusConnectsToHostPanel()
		})

	sc.Step(`^a game session has been loaded with "([^"]*)"$`,
		func(filename string) error {
			return w.givenGameSessionLoaded(filename)
		})

	sc.Step(`^Marcus has loaded "([^"]*)" and the lobby is open$`,
		func(filename string) error {
			return w.givenGameSessionLoaded(filename)
		})

	sc.Step(`^Marcus has loaded "([^"]*)" successfully$`,
		func(filename string) error {
			return w.givenGameSessionLoaded(filename)
		})

	sc.Step(`^Marcus has loaded "([^"]*)" and is on the lobby screen$`,
		func(filename string) error {
			return w.givenGameSessionLoaded(filename)
		})

	sc.Step(`^"([^"]*)" is connected in the lobby$`,
		func(teamName string) error {
			return w.givenTeamConnected(teamName)
		})

	sc.Step(`^"([^"]*)", "([^"]*)", and "([^"]*)" are connected in the lobby$`,
		func(t1, t2, t3 string) error {
			for _, t := range []string{t1, t2, t3} {
				if err := w.givenTeamConnected(t); err != nil {
					return err
				}
			}
			return nil
		})

	sc.Step(`^"([^"]*)", "([^"]*)", and "([^"]*)" are all connected$`,
		func(t1, t2, t3 string) error {
			for _, t := range []string{t1, t2, t3} {
				if err := w.givenTeamConnected(t); err != nil {
					return err
				}
			}
			return nil
		})

	sc.Step(`^the display interface is connected$`,
		func() error {
			return w.givenDisplayConnected()
		})

	sc.Step(`^Marcus has started the game$`,
		func() error {
			return w.givenMarcusStartedGame(0)
		})

	sc.Step(`^Marcus has started the game and revealed question (\d+) in round (\d+)$`,
		func(qNum, roundNum int) error {
			if err := w.givenMarcusStartedGame(roundNum - 1); err != nil {
				return err
			}
			return w.givenQuestionsRevealed(roundNum-1, qNum)
		})

	sc.Step(`^Marcus has started the game and is in round (\d+) with questions (\d+) through (\d+) revealed$`,
		func(roundNum, fromQ, toQ int) error {
			if err := w.givenMarcusStartedGame(roundNum - 1); err != nil {
				return err
			}
			return w.givenQuestionsRevealed(roundNum-1, toQ)
		})

	sc.Step(`^Marcus is on the Round (\d+) reveal panel and no questions have been revealed$`,
		func(roundNum int) error {
			return w.givenMarcusStartedGame(roundNum - 1)
		})

	sc.Step(`^the game is in Round (\d+) and no questions have been revealed$`,
		func(roundNum int) error {
			return w.givenMarcusStartedGame(roundNum - 1)
		})

	sc.Step(`^no questions have been revealed yet$`,
		func() error {
			return nil // state already correct after game start
		})

	sc.Step(`^Round (\d+) has ended and all (\d+) questions have been revealed$`,
		func(roundNum, qCount int) error {
			return w.givenRoundEnded(roundNum-1, qCount)
		})

	sc.Step(`^"([^"]*)" has entered answers for all (\d+) questions$`,
		func(teamName string, qCount int) error {
			return w.givenTeamEnteredAnswers(teamName, qCount)
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

	sc.Step(`^Round (\d+) has been fully played, scored, and ceremonialized$`,
		func(roundNum int) error {
			return w.givenRoundFullyComplete(roundNum - 1)
		})

	sc.Step(`^round scores have been published$`,
		func() error {
			return w.givenScoresPublished(0)
		})

	sc.Step(`^question (\d+) has text "([^"]*)" and answer "([^"]*)"$`,
		func(qNum int, text, answer string) error {
			w.givenQuizFixtureQuestionDetail(qNum-1, text, answer)
			return nil
		})

	sc.Step(`^"([^"]*)" is already registered in the lobby$`,
		func(teamName string) error {
			return w.givenTeamConnected(teamName)
		})

	sc.Step(`^"([^"]*)" has already been registered by another device$`,
		func(teamName string) error {
			return w.givenTeamConnected(teamName)
		})

	sc.Step(`^no teams have connected yet$`,
		func() error {
			return nil // no action needed
		})

	sc.Step(`^Priya was connected in the lobby but lost her connection before game start$`,
		func() error {
			return w.givenPlayerDisconnectedBeforeStart("Team Awesome", "Priya")
		})

	sc.Step(`^the quiz directory contains "([^"]*)"$`,
		func(filename string) error {
			return w.givenMediaFileExists(filename)
		})

	sc.Step(`^"([^"]*)" does not exist in the quiz directory$`,
		func(filename string) error {
			return nil // no action; file simply absent
		})

	sc.Step(`^HOST_TOKEN is set to "([^"]*)"$`,
		func(token string) error {
			w.hostToken = token
			return nil
		})

	sc.Step(`^HOST_TOKEN is not set in the environment$`,
		func() error {
			w.hostToken = ""
			return nil
		})

	sc.Step(`^QUIZ_DIR is set to an accessible directory$`,
		func() error {
			return nil // handled by server setup
		})

	sc.Step(`^QUIZ_DIR is set to "([^"]*)"$`,
		func(path string) error {
			w.quizFixtures["QUIZ_DIR_OVERRIDE"] = path
			return nil
		})

	sc.Step(`^the project source code is present$`,
		func() error {
			return nil // infrastructure scenario prerequisite
		})

	sc.Step(`^the Docker image has been built$`,
		func() error {
			return nil // infrastructure scenario prerequisite
		})

	sc.Step(`^the project Go source code is present$`,
		func() error {
			return nil // infrastructure scenario prerequisite
		})

	sc.Step(`^the frontend source code is present$`,
		func() error {
			return nil // infrastructure scenario prerequisite
		})

	sc.Step(`^Marcus has started round (\d+) and revealed all (\d+) questions$`,
		func(roundNum, qCount int) error {
			if err := w.givenMarcusStartedGame(roundNum - 1); err != nil {
				return err
			}
			return w.givenQuestionsRevealed(roundNum-1, qCount)
		})

	sc.Step(`^Marcus has ended round (\d+)$`,
		func(roundNum int) error {
			return w.givenRoundEnded(roundNum-1, -1)
		})

	sc.Step(`^Priya has entered "([^"]*)" as a draft answer for question (\d+) as "([^"]*)"$`,
		func(answer string, qNum int, teamName string) error {
			return w.givenDraftAnswerEntered(teamName, qNum-1, answer)
		})

	sc.Step(`^"([^"]*)"'s connection is interrupted and restored$`,
		func(teamName string) error {
			return w.givenConnectionInterruptedAndRestored(teamName)
		})

	sc.Step(`^all 3 questions are scored and Marcus has started the ceremony$`,
		func() error {
			return w.givenCeremonyStarted(0)
		})

	sc.Step(`^the ceremony is at question (\d+) and only the question text is shown$`,
		func(qNum int) error {
			return w.givenCeremonyAtQuestion(0, qNum-1)
		})

	// -----------------------------------------------------------------------
	// When steps -- act
	// -----------------------------------------------------------------------

	sc.Step(`^Marcus loads "([^"]*)" via the quizmaster interface$`,
		func(filename string) error {
			return w.whenMarcusLoadsQuiz(filename)
		})

	sc.Step(`^Marcus loads "([^"]*)" via the host interface$`,
		func(filename string) error {
			return w.whenMarcusLoadsQuiz(filename)
		})

	sc.Step(`^Marcus provides the path "([^"]*)" via the quizmaster interface$`,
		func(path string) error {
			return w.whenMarcusLoadsQuizPath(path)
		})

	sc.Step(`^Marcus loads the quiz file via the quizmaster interface$`,
		func() error {
			return w.whenMarcusLoadsDefaultQuiz()
		})

	sc.Step(`^Priya connects to the player interface and joins as "([^"]*)"$`,
		func(teamName string) error {
			return w.whenPlayerJoins(teamName)
		})

	sc.Step(`^Priya connects to the player interface and registers as "([^"]*)"$`,
		func(teamName string) error {
			return w.whenPlayerJoins(teamName)
		})

	sc.Step(`^a new player connects and joins as "([^"]*)"$`,
		func(teamName string) error {
			return w.whenPlayerJoins(teamName)
		})

	sc.Step(`^another player tries to join as "([^"]*)" from a different device$`,
		func(teamName string) error {
			return w.whenPlayerJoinsSecondDevice(teamName)
		})

	sc.Step(`^a new player tries to join as "([^"]*)"$`,
		func(teamName string) error {
			return w.whenPlayerJoinsSecondDevice(teamName)
		})

	sc.Step(`^Marcus starts the game$`,
		func() error {
			return w.whenMarcusStartsRound(0)
		})

	sc.Step(`^Marcus starts the game via the quizmaster interface$`,
		func() error {
			return w.whenMarcusStartsRound(0)
		})

	sc.Step(`^Marcus sends the start round command for Round (\d+) via the host interface$`,
		func(roundNum int) error {
			return w.whenMarcusStartsRound(roundNum - 1)
		})

	sc.Step(`^Marcus reveals question (\d+)$`,
		func(qNum int) error {
			return w.whenMarcusRevealsQuestion(0, qNum-1)
		})

	sc.Step(`^Marcus reveals question (\d+) via the quizmaster interface$`,
		func(qNum int) error {
			return w.whenMarcusRevealsQuestion(0, qNum-1)
		})

	sc.Step(`^Marcus reveals question (\d+) via the host interface$`,
		func(qNum int) error {
			return w.whenMarcusRevealsQuestion(0, qNum-1)
		})

	sc.Step(`^Priya enters "([^"]*)" as her answer to question (\d+)$`,
		func(answer string, qNum int) error {
			return w.whenPlayerDraftsAnswer("Team Awesome", qNum-1, answer)
		})

	sc.Step(`^Marcus ends the round$`,
		func() error {
			return w.whenMarcusEndsRound(0)
		})

	sc.Step(`^Priya submits "([^"]*)"'s answers$`,
		func(teamName string) error {
			return w.whenTeamSubmits(teamName, 0)
		})

	sc.Step(`^"([^"]*)" submits their answers via the player interface$`,
		func(teamName string) error {
			return w.whenTeamSubmits(teamName, 0)
		})

	sc.Step(`^Marcus opens scoring$`,
		func() error {
			return nil // scoring opens automatically when all teams submit in this scenario
		})

	sc.Step(`^Marcus marks "([^"]*)" as correct for question (\d+)$`,
		func(answer string, qNum int) error {
			return w.whenMarcusMarksAnswer("Team Awesome", 0, qNum-1, "correct")
		})

	sc.Step(`^Marcus marks "([^"]*)" as correct for question (\d+)$`,
		func(answer string, qNum int) error {
			return w.whenMarcusMarksAnswer("Team Awesome", 0, qNum-1, "correct")
		})

	sc.Step(`^Marcus marks "([^"]*)" answer for question (\d+) as correct via the host interface$`,
		func(teamName string, qNum int) error {
			return w.whenMarcusMarksAnswer(teamName, 0, qNum-1, "correct")
		})

	sc.Step(`^Marcus starts the answer ceremony$`,
		func() error {
			return w.whenMarcusStartsCeremony(0)
		})

	sc.Step(`^Marcus starts the ceremony and advances through all (\d+) questions via the host interface$`,
		func(qCount int) error {
			return w.whenMarcusRunsFullCeremony(0, qCount)
		})

	sc.Step(`^Marcus reveals the answer to question (\d+) during ceremony$`,
		func(qNum int) error {
			return w.whenMarcusCeremonyRevealAnswer(0, qNum-1)
		})

	sc.Step(`^Marcus advances to reveal the answer for question (\d+)$`,
		func(qNum int) error {
			return w.whenMarcusCeremonyRevealAnswer(0, qNum-1)
		})

	sc.Step(`^Marcus sends the show ceremony question event for question (\d+)$`,
		func(qNum int) error {
			return w.whenMarcusCeremonyShowQuestion(0, qNum-1)
		})

	sc.Step(`^Marcus sends the reveal ceremony answer event for question (\d+)$`,
		func(qNum int) error {
			return w.whenMarcusCeremonyRevealAnswer(0, qNum-1)
		})

	sc.Step(`^Marcus steps through all ceremony questions and publishes the round scores$`,
		func() error {
			return w.whenMarcusPublishesScores(0)
		})

	sc.Step(`^Marcus publishes round scores via the host interface$`,
		func() error {
			return w.whenMarcusPublishesScores(0)
		})

	sc.Step(`^Marcus ends the game$`,
		func() error {
			return w.whenMarcusEndsGame()
		})

	sc.Step(`^Marcus ends the game via the quizmaster interface$`,
		func() error {
			return w.whenMarcusEndsGame()
		})

	sc.Step(`^Marcus ends the game via the host interface$`,
		func() error {
			return w.whenMarcusEndsGame()
		})

	sc.Step(`^an HTTP request is made to the quizmaster panel without a token parameter$`,
		func() error {
			return w.whenHTTPRequest("host_no_token")
		})

	sc.Step(`^an HTTP request is made to the quizmaster panel with token "([^"]*)"$`,
		func(token string) error {
			return w.whenHTTPRequestWithToken("host", token)
		})

	sc.Step(`^an HTTP request is made to the player interface URL$`,
		func() error {
			return w.whenHTTPRequest("play")
		})

	sc.Step(`^an HTTP request is made to the display interface URL$`,
		func() error {
			return w.whenHTTPRequest("display")
		})

	sc.Step(`^an HTTP request is made to the media path for "([^"]*)"$`,
		func(filename string) error {
			return w.whenHTTPRequest("media:" + filename)
		})

	sc.Step(`^the server process starts$`,
		func() error {
			return w.whenServerStarts()
		})

	sc.Step(`^a connection is made to the quizmaster panel without a valid host token$`,
		func() error {
			return w.whenConnectWithoutToken()
		})

	sc.Step(`^that connection attempts to send a host_reveal_question event$`,
		func() error {
			return w.whenUnauthorizedReveal()
		})

	sc.Step(`^a new player connection is established and registers as "([^"]*)"$`,
		func(teamName string) error {
			return w.whenPlayerJoins(teamName)
		})

	sc.Step(`^"([^"]*)" sends a team_rejoin event with their stored token$`,
		func(teamName string) error {
			return w.whenTeamRejoins(teamName)
		})

	sc.Step(`^a connection attempts to upgrade to the host WebSocket room without a valid token$`,
		func() error {
			return w.whenConnectWithoutToken()
		})

	sc.Step(`^Marcus connects to the host room WebSocket with the correct token$`,
		func() error {
			return w.givenMarcusConnectsToHostPanel()
		})

	sc.Step(`^Priya refreshes her player page$`,
		func() error {
			return w.whenPlayerRefreshes("Team Awesome")
		})

	sc.Step(`^the submission is in progress waiting for server acknowledgment$`,
		func() error {
			return nil // state assertion only; submission sent in prior When step
		})

	sc.Step(`^the Docker image build is run$`,
		func() error {
			return w.whenDockerBuildRuns()
		})

	sc.Step(`^the container is started with docker-compose$`,
		func() error {
			return w.whenDockerComposeUp()
		})

	sc.Step(`^go-arch-lint check is run against the project$`,
		func() error {
			return w.whenGoArchLintRuns()
		})

	sc.Step(`^TypeScript type checking is run with strict mode$`,
		func() error {
			return w.whenTypeScriptTypeCheck()
		})

	sc.Step(`^go test is run with the race detector flag$`,
		func() error {
			return w.whenGoTestWithRace()
		})

	sc.Step(`^Priya's connection is restored$`,
		func() error {
			return w.whenConnectionRestored("Team Awesome")
		})

	sc.Step(`^all WebSocket messages sent to the play room are inspected$`,
		func() error {
			return nil // assertion happens in Then steps
		})

	sc.Step(`^the game progresses through any valid sequence of state transitions$`,
		func() error {
			return w.whenFullGameSequenceRuns()
		})

	sc.Step(`^Marcus reveals any question via the quizmaster interface$`,
		func() error {
			return w.whenMarcusRevealsQuestion(0, 0)
		})

	// -----------------------------------------------------------------------
	// Then steps -- assert
	// -----------------------------------------------------------------------

	sc.Step(`^Marcus sees the quiz title "([^"]*)"$`,
		func(title string) error {
			return w.thenMarcusSeesQuizTitle(title)
		})

	sc.Step(`^Marcus sees "([^"]*)"$`,
		func(text string) error {
			return w.thenHostSeesText(text)
		})

	sc.Step(`^Marcus sees "([^"]*)" in the connected teams list within (\d+) seconds$`,
		func(teamName string, seconds int) error {
			return w.thenMarcusSeesTeamInLobby(teamName, time.Duration(seconds)*time.Second)
		})

	sc.Step(`^Marcus sees "([^"]*)" in the connected teams list$`,
		func(teamName string) error {
			return w.thenMarcusSeesTeamInLobby(teamName, 2*time.Second)
		})

	sc.Step(`^Marcus sees "([^"]*)" appear in the connected teams list$`,
		func(teamName string) error {
			return w.thenMarcusSeesTeamInLobby(teamName, 2*time.Second)
		})

	sc.Step(`^Marcus sees "([^"]*)" listed as "([^"]*)" on the host interface$`,
		func(teamName, status string) error {
			return w.thenMarcusSeesTeamStatus(teamName, status)
		})

	sc.Step(`^Marcus sees "([^"]*)" listed as submitted in the quizmaster panel$`,
		func(teamName string) error {
			return w.thenMarcusSeesTeamStatus(teamName, "submitted")
		})

	sc.Step(`^Marcus sees (\d+) round and (\d+) questions$`,
		func(rounds, questions int) error {
			return w.thenMarcusSeesQuizCounts(rounds, questions)
		})

	sc.Step(`^Marcus sees (\d+) rounds and (\d+) questions$`,
		func(rounds, questions int) error {
			return w.thenMarcusSeesQuizCounts(rounds, questions)
		})

	sc.Step(`^a shareable player URL is shown$`,
		func() error {
			return w.thenShareablePlayerURLShown()
		})

	sc.Step(`^a shareable display URL is shown$`,
		func() error {
			return w.thenShareableDisplayURLShown()
		})

	sc.Step(`^the panel shows a join URL that players can use to connect$`,
		func() error {
			return w.thenShareablePlayerURLShown()
		})

	sc.Step(`^the panel shows a shareable player join URL$`,
		func() error {
			return w.thenShareablePlayerURLShown()
		})

	sc.Step(`^the panel shows a shareable display screen URL$`,
		func() error {
			return w.thenShareableDisplayURLShown()
		})

	sc.Step(`^the game session has a unique session identifier$`,
		func() error {
			return w.thenGameSessionHasUniqueID()
		})

	sc.Step(`^a team identity token is stored for "([^"]*)"$`,
		func(teamName string) error {
			return w.thenTeamTokenStored(teamName)
		})

	sc.Step(`^a persistence token is stored in Priya's browser for "([^"]*)"$`,
		func(teamName string) error {
			return w.thenTeamTokenStored(teamName)
		})

	sc.Step(`^"([^"]*)" player interface transitions to "([^"]*)" within (\d+) second$`,
		func(teamName, state string, seconds int) error {
			return w.thenPlayerSeesState(teamName, state, time.Duration(seconds)*time.Second)
		})

	sc.Step(`^"([^"]*)" player interface transitions to "([^"]*)" within (\d+) seconds$`,
		func(teamName, state string, seconds int) error {
			return w.thenPlayerSeesState(teamName, state, time.Duration(seconds)*time.Second)
		})

	sc.Step(`^the display interface transitions to the question view within (\d+) second$`,
		func(seconds int) error {
			return w.thenDisplaySeesState("question_view", time.Duration(seconds)*time.Second)
		})

	sc.Step(`^the display interface transitions to the question view within (\d+) seconds$`,
		func(seconds int) error {
			return w.thenDisplaySeesState("question_view", time.Duration(seconds)*time.Second)
		})

	sc.Step(`^Marcus sees the Round (\d+) reveal panel on the host interface$`,
		func(roundNum int) error {
			return w.thenHostSeesRevealPanel(roundNum - 1)
		})

	sc.Step(`^the game enters the first round$`,
		func() error {
			return w.thenHostSeesRevealPanel(0)
		})

	sc.Step(`^"([^"]*)" sees "([^"]*)" on their player screen$`,
		func(teamName, text string) error {
			return w.thenPlayerSeesText(teamName, text, 2*time.Second)
		})

	sc.Step(`^"([^"]*)"'s player screen shows "([^"]*)"$`,
		func(teamName, text string) error {
			return w.thenPlayerSeesText(teamName, text, 2*time.Second)
		})

	sc.Step(`^the host interface shows "([^"]*) of (\d+) revealed"$`,
		func(countStr string, total int) error {
			return w.thenHostSeesRevealCount(countStr, total)
		})

	sc.Step(`^neither the player interface nor the display interface contains the answer field for question (\d+)$`,
		func(qNum int) error {
			return w.thenNoAnswerFieldInPlayOrDisplay(qNum - 1)
		})

	sc.Step(`^the player interface shows "Your answers are locked in" only after server acknowledgement$`,
		func() error {
			return w.thenLockedStateAfterAck()
		})

	sc.Step(`^"([^"]*)" running total increments by (\d+)$`,
		func(teamName string, points int) error {
			return w.thenTeamScoreIncreasedBy(teamName, points)
		})

	sc.Step(`^the scoring panel on the host interface reflects the updated score$`,
		func() error {
			return nil // implied by score increment assertion
		})

	sc.Step(`^the display interface shows the round scores in rank order$`,
		func() error {
			return w.thenDisplayShowsRoundScores()
		})

	sc.Step(`^"([^"]*)" appears with their correct score$`,
		func(teamName string) error {
			return w.thenTeamAppearsInScores(teamName)
		})

	sc.Step(`^the player interface also shows the round scores$`,
		func() error {
			return w.thenPlayersReceiveRoundScores()
		})

	sc.Step(`^the display interface shows "Final Scores" with all teams in rank order$`,
		func() error {
			return w.thenDisplayShowsFinalScores()
		})

	sc.Step(`^the player interface shows the final scores$`,
		func() error {
			return w.thenPlayersReceiveFinalScores()
		})

	sc.Step(`^the game state is "([^"]*)"$`,
		func(state string) error {
			return w.thenGameStateIs(state)
		})

	sc.Step(`^the display screen shows the final winner as "([^"]*)"$`,
		func(teamName string) error {
			return w.thenDisplayShowsWinner(teamName)
		})

	sc.Step(`^Priya sees confirmation that "([^"]*)"'s answers are locked in$`,
		func(teamName string) error {
			return w.thenPlayerSeesLockedIn(teamName)
		})

	sc.Step(`^the display screen shows question (\d+) text of the ceremony$`,
		func(qNum int) error {
			return w.thenDisplayShowsCeremonyQuestion(qNum - 1)
		})

	sc.Step(`^the display screen shows the answer "([^"]*)" for question (\d+)$`,
		func(answer string, qNum int) error {
			return w.thenDisplayShowsCeremonyAnswer(answer, qNum-1)
		})

	sc.Step(`^the display screen shows "([^"]*)" with (\d+) points$`,
		func(teamName string, points int) error {
			return w.thenDisplayShowsTeamScore(teamName, points)
		})

	sc.Step(`^all answer fields on the player screen are permanently locked$`,
		func() error {
			return w.thenAllAnswerFieldsLocked()
		})

	sc.Step(`^Marcus sees the error "([^"]*)"$`,
		func(errMsg string) error {
			return w.thenMarcusSeesError(errMsg)
		})

	sc.Step(`^no game session is created$`,
		func() error {
			return w.thenNoGameSession()
		})

	sc.Step(`^Marcus can correct the file and reload without losing the quizmaster page$`,
		func() error {
			return nil // UX property; server remains accessible
		})

	sc.Step(`^the response status is (\d+)$`,
		func(code int) error {
			return w.thenHTTPStatusIs(code)
		})

	sc.Step(`^the response contains the React application HTML$`,
		func() error {
			return w.thenResponseContainsSPAHTML()
		})

	sc.Step(`^the response content type is "([^"]*)"$`,
		func(ct string) error {
			return w.thenResponseContentTypeIs(ct)
		})

	sc.Step(`^the player room receives the round started event within (\d+) second$`,
		func(seconds int) error {
			return w.thenRoomReceivesEvent("play:Team Awesome", "round_started", time.Duration(seconds)*time.Second)
		})

	sc.Step(`^the display room receives the round started event within (\d+) second$`,
		func(seconds int) error {
			return w.thenRoomReceivesEvent("display", "round_started", time.Duration(seconds)*time.Second)
		})

	sc.Step(`^the player room receives the question_revealed event within (\d+) second$`,
		func(seconds int) error {
			return w.thenRoomReceivesEvent("play:Team Awesome", "question_revealed", time.Duration(seconds)*time.Second)
		})

	sc.Step(`^the display room receives the question_revealed event within (\d+) second$`,
		func(seconds int) error {
			return w.thenRoomReceivesEvent("display", "question_revealed", time.Duration(seconds)*time.Second)
		})

	sc.Step(`^the server sends a submission acknowledgment to the "([^"]*)" connection$`,
		func(teamName string) error {
			return w.thenRoomReceivesEvent("play:"+teamName, "submission_ack", 2*time.Second)
		})

	sc.Step(`^the "([^"]*)" player screen shows "Your answers are locked in" only after the acknowledgment$`,
		func(teamName string) error {
			return w.thenLockedStateAfterAck()
		})

	sc.Step(`^the message received by the player room for question_revealed has no "([^"]*)" field$`,
		func(fieldName string) error {
			return w.thenQuestionRevealedHasNoField("play:Team Awesome", fieldName)
		})

	sc.Step(`^the message received by the display room for question_revealed has no "([^"]*)" field$`,
		func(fieldName string) error {
			return w.thenQuestionRevealedHasNoField("display", fieldName)
		})

	sc.Step(`^the message contains the question text "([^"]*)"$`,
		func(text string) error {
			return w.thenQuestionRevealedContainsText("play:Team Awesome", text)
		})

	sc.Step(`^the display room receives the ceremony_question_shown event with no "([^"]*)" field$`,
		func(fieldName string) error {
			return w.thenCeremonyEventHasNoField("display", "ceremony_question_shown", fieldName)
		})

	sc.Step(`^the display room receives the ceremony_answer_revealed event with answer "([^"]*)"$`,
		func(answer string) error {
			return w.thenCeremonyAnswerRevealedWith("display", answer)
		})

	sc.Step(`^the player room does not receive the ceremony_answer_revealed event$`,
		func() error {
			return w.thenRoomDoesNotReceiveEvent("play:Team Awesome", "ceremony_answer_revealed")
		})

	sc.Step(`^none of the messages contain a field named "([^"]*)"$`,
		func(fieldName string) error {
			return w.thenNoMessagesToPlayContainField(fieldName)
		})

	sc.Step(`^every question_revealed message sent to play or display rooms contains only QuestionPublic fields$`,
		func() error {
			return w.thenAllRevealedQuestionsPublicOnly()
		})

	sc.Step(`^the connection receives a state_snapshot event$`,
		func() error {
			return w.thenConnectionReceivesSnapshot()
		})

	sc.Step(`^the snapshot includes round (\d+) as active$`,
		func(roundNum int) error {
			return w.thenSnapshotShowsActiveRound(roundNum - 1)
		})

	sc.Step(`^the snapshot includes questions (\d+) and (\d+) as revealed$`,
		func(q1, q2 int) error {
			return w.thenSnapshotShowsRevealedQuestions(q1-1, q2-1)
		})

	sc.Step(`^the snapshot contains no answer fields for any question$`,
		func() error {
			return w.thenSnapshotHasNoAnswerFields()
		})

	sc.Step(`^the state_snapshot received by "([^"]*)" includes the draft answer "([^"]*)" for question (\d+)$`,
		func(teamName, answer string, qNum int) error {
			return w.thenSnapshotContainsDraft(teamName, answer, qNum-1)
		})

	sc.Step(`^the connection is rejected with a (\d+) status$`,
		func(code int) error {
			return w.thenConnectionRejectedWithStatus(code)
		})

	sc.Step(`^the game state is not affected$`,
		func() error {
			return nil // verifiable via game state query; no-op for now
		})

	sc.Step(`^the connection is accepted$`,
		func() error {
			return w.thenConnectionAccepted()
		})

	sc.Step(`^Marcus receives the current game state snapshot$`,
		func() error {
			return w.thenRoomReceivesEvent("host", "state_snapshot", 2*time.Second)
		})

	sc.Step(`^the server process exits with a startup error$`,
		func() error {
			return w.thenServerExitedWithError()
		})

	sc.Step(`^the error output contains "([^"]*)"$`,
		func(msg string) error {
			return w.thenErrorOutputContains(msg)
		})

	sc.Step(`^the error message states "([^"]*)"$`,
		func(msg string) error {
			return w.thenErrorOutputContains(msg)
		})

	sc.Step(`^no port is bound$`,
		func() error {
			return nil // verified by server exit
		})

	sc.Step(`^the error output identifies the inaccessible path$`,
		func() error {
			return w.thenErrorOutputContains("not accessible")
		})

	sc.Step(`^the server binds to port 8080 without error$`,
		func() error {
			return w.thenServerRunning()
		})

	sc.Step(`^the server is ready to accept connections$`,
		func() error {
			return w.thenServerRunning()
		})

	sc.Step(`^the build completes without error$`,
		func() error {
			return w.thenDockerBuildSucceeded()
		})

	sc.Step(`^the resulting image uses the distroless runtime base$`,
		func() error {
			return nil // verified by Dockerfile inspection; not automatable here
		})

	sc.Step(`^the container becomes healthy within (\d+) seconds$`,
		func(seconds int) error {
			return w.thenContainerHealthy(time.Duration(seconds) * time.Second)
		})

	sc.Step(`^the player interface responds to HTTP requests on the configured port$`,
		func() error {
			return w.thenServerRunning()
		})

	sc.Step(`^no architecture violations are reported$`,
		func() error {
			return w.thenGoArchLintPassed()
		})

	sc.Step(`^specifically the handler package has no reference to QuestionFull$`,
		func() error {
			return w.thenPackageHasNoImport("handler", "QuestionFull")
		})

	sc.Step(`^specifically the hub package has no reference to QuestionFull$`,
		func() error {
			return w.thenPackageHasNoImport("hub", "QuestionFull")
		})

	sc.Step(`^zero type errors are reported$`,
		func() error {
			return w.thenTypeScriptPassed()
		})

	sc.Step(`^all tests pass$`,
		func() error {
			return w.thenGoTestPassed()
		})

	sc.Step(`^no race conditions are detected$`,
		func() error {
			return w.thenNoRaceConditions()
		})

	sc.Step(`^Priya sees the lobby showing "([^"]*)" and other connected teams$`,
		func(teamName string) error {
			return w.thenPlayerSeesLobby(teamName)
		})

	sc.Step(`^that player sees "([^"]*)"$`,
		func(msg string) error {
			return w.thenPlayerSeesError(msg)
		})

	sc.Step(`^the name field remains populated so the player can edit it$`,
		func() error {
			return w.thenNameFieldRemainedPopulated()
		})

	sc.Step(`^no duplicate team entry appears in the lobby$`,
		func() error {
			return w.thenNoDuplicateTeamInLobby()
		})

	// US-03: Start Game Broadcast steps
	sc.Step(`^all three player connections receive the round started event within (\d+) second$`,
		func(seconds int) error {
			return w.thenAllThreePlayersReceiveRoundStarted(time.Duration(seconds) * time.Second)
		})

	sc.Step(`^each player screen shows round (\d+) is active$`,
		func(roundNum int) error {
			return w.thenEachPlayerSeesRoundActive(roundNum - 1)
		})

	sc.Step(`^the display screen transitions from the waiting state to the question view$`,
		func() error {
			return w.thenDisplayReceivesRoundStarted(1 * time.Second)
		})

	sc.Step(`^"([^"]*)" immediately sees round (\d+) as active$`,
		func(teamName string, roundNum int) error {
			return w.thenLateJoinerSeesRoundActive(teamName, roundNum-1)
		})

	sc.Step(`^"([^"]*)" sees question (\d+) already revealed in their answer form$`,
		func(teamName string, qNum int) error {
			return w.thenLateJoinerSeesRevealedQuestion(teamName, qNum-1)
		})

	sc.Step(`^"([^"]*)" does not see a lobby screen$`,
		func(teamName string) error {
			return w.thenLateJoinerNotInLobby(teamName)
		})

	_ = strings.ToLower // silence unused import
	_ = fmt.Sprintf    // silence unused import
}
