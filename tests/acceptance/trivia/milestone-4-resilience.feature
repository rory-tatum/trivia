Feature: Milestone 4 -- Reconnection and Edge Cases
  As all trivia night participants
  I want the game to recover from real-world disruptions gracefully
  So that a dropped connection or accidental refresh does not ruin game night

  # All scenarios @skip until walking skeleton and milestones 1-3 pass.
  #
  # Driving ports:
  #   team_rejoin      -> /ws (play room) -- reconnects existing team with stored token
  #   state_snapshot   -> /ws (all rooms) -- server sends on any new connection
  #   exponential backoff reconnection: 1s base, 2x multiplier, 30s max, 10 attempts

  Background:
    Given Marcus has loaded "friday-march-2026.yaml" with 2 rounds of 8 text questions each
    And "Team Awesome", "The Brainiacs", and "Quiz Killers" are all connected
    And Marcus has started the game and is in round 1 with questions 1 through 4 revealed

  # -------------------------------------------------------------------------
  # Player reconnection (US-05 extended: mid-game reconnect)
  # -------------------------------------------------------------------------

  @skip
  Scenario: Player reconnects after a brief connection loss and sees current state
    Given Priya has entered "Paris" for question 1 and "Mercury" for question 2
    And Priya's player connection is interrupted for 10 seconds
    When Priya's connection is restored
    Then the player interface reconnects automatically without requiring a page reload
    And Priya sees a brief "Reconnecting..." indicator that clears on success
    And Priya sees questions 1 through 4 still revealed
    And Priya's answers "Paris" and "Mercury" are still present

  @skip
  Scenario: Player connection drops and game state advances before reconnect
    Given Priya's connection drops while questions 1 and 2 are revealed
    And Marcus reveals questions 3 and 4 while Priya is disconnected
    When Priya's connection is restored
    Then Priya's player screen shows all 4 revealed questions
    And Priya sees no gap in the question list

  @skip
  Scenario: Player reconnects after submitting and sees the locked state
    Given "Team Awesome" submitted their round 1 answers
    And Priya's connection dropped after submission
    When Priya's connection is restored
    Then Priya's player screen shows the submitted answers in read-only form
    And a banner reads "Answers are locked for round 1"

  @skip
  Scenario: Player reaches reconnection limit and sees a manual reconnect prompt
    Given Priya's connection has failed 10 consecutive reconnection attempts
    When the reconnection limit is reached
    Then Priya sees a "Disconnected" message with a manual reconnect option
    And the page does not continue retrying automatically

  # -------------------------------------------------------------------------
  # Display reconnection
  # -------------------------------------------------------------------------

  @skip
  Scenario: Display screen reconnects and shows the current game state
    Given the display screen is showing question 3 as the current question
    And the display screen's connection is interrupted
    When the display connection is restored
    Then the display screen shows a "Reconnecting..." overlay briefly
    And the overlay clears when the connection is re-established
    And the display screen shows the current question as it stands at reconnect time

  # -------------------------------------------------------------------------
  # Host reconnection
  # -------------------------------------------------------------------------

  @skip
  Scenario: Quizmaster reopens the host panel and game state is fully restored
    Given the game is in round 2 in the scoring phase
    When Marcus closes and reopens the quizmaster panel tab
    Then the quizmaster panel restores the scoring interface for round 2
    And all previously entered verdicts are preserved

  # -------------------------------------------------------------------------
  # Invalid state transition attempts
  # -------------------------------------------------------------------------

  @skip
  Scenario: Attempting to reveal question 2 before question 1 is not allowed
    Given no questions have been revealed in round 1
    When Marcus attempts to reveal question 2 without first revealing question 1
    Then the quizmaster interface does not allow this action
    And an error is shown indicating that questions must be revealed in sequence

  @skip
  Scenario: Attempting to start scoring before all questions are revealed is not allowed
    Given only 4 of 8 questions have been revealed in round 1
    When Marcus attempts to end the round and open scoring
    Then the quizmaster interface does not allow this action
    And Marcus sees a message indicating all questions must be revealed first

  @skip
  Scenario: Attempting to start the ceremony before all answers are scored is not allowed
    Given Marcus is in the scoring phase and has only scored 5 of 8 questions for all teams
    When Marcus attempts to start the ceremony
    Then the quizmaster interface does not allow this action

  @skip
  Scenario: An unauthenticated request to the host interface is rejected
    Given a connection is made to the quizmaster panel without a valid host token
    When that connection attempts to send a host_reveal_question event
    Then the connection receives an error response
    And the event has no effect on the game state

  # -------------------------------------------------------------------------
  # Server startup edge cases
  # -------------------------------------------------------------------------

  @skip
  Scenario: Server fails to start when the host token is not configured
    Given the server is started without the HOST_TOKEN environment variable set
    Then the server process exits with a startup error
    And the error message states "HOST_TOKEN environment variable is required"

  @skip
  Scenario: Server fails to start when the quiz directory is not accessible
    Given the server is started with QUIZ_DIR set to a path that does not exist
    Then the server process exits with a startup error
    And the error message identifies the inaccessible path
