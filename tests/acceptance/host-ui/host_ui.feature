Feature: Host UI — Quizmaster Panel
  As Marcus, the quizmaster
  I want a browser-based control panel that accurately reflects the game state
  So that I can run a trivia game session for my friends with confidence

  # ============================================================================
  # WALKING SKELETON (Strategy C — Real Local WebSocket)
  # Answers: "Can Marcus connect, load a quiz, run a round, score it, and end the game?"
  # Traces:  US-01a, US-02, US-03, US-04, US-06
  # ============================================================================

  # @walking_skeleton @driving_port @real-io @US-01 @US-02 @US-03 @US-04 @US-06
  @walking_skeleton @driving_port @real-io @US-01 @US-02 @US-03 @US-04 @US-06
  Scenario: Marcus opens the host panel with a valid token and runs a complete game session
    Given a quiz file "friday-quiz.yaml" with 1 round of 3 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "The Brainiacs" is connected in the lobby
    When Marcus opens the quizmaster panel with a valid host token
    Then the quizmaster panel shows "Connected" status
    And the load quiz form is visible with a file path input

    When Marcus loads "friday-quiz.yaml" through the quizmaster panel
    Then the quizmaster panel shows the quiz confirmation
    And the confirmation includes the quiz title and round count
    And the player join URL is displayed for Marcus to share
    And the display URL is displayed for Marcus to share
    And the "Start Round 1" button is visible

    When Marcus starts Round 1
    Then the round panel is visible showing "0 of 3 revealed"

    When Marcus reveals question 1
    Then the round panel shows "1 of 3 revealed"

    When Marcus reveals question 2
    And Marcus reveals question 3
    Then the round panel shows "3 of 3 revealed"
    And the "End Round" button is visible

    When Marcus ends Round 1
    Then the scoring panel is visible
    And the scoring panel shows each question with its correct answer
    And the scoring panel shows submitted answers for "The Brainiacs"

    When Marcus marks "The Brainiacs" answer for question 1 as correct
    And Marcus marks "The Brainiacs" answer for question 2 as correct
    And Marcus marks "The Brainiacs" answer for question 3 as wrong
    Then the running total for "The Brainiacs" reflects 2 correct answers

    When Marcus publishes scores for Round 1
    Then the "End Game" button is visible

    When Marcus ends the game
    Then the final leaderboard is displayed
    And "The Brainiacs" appears on the leaderboard with their score
    And game control buttons are no longer visible


  # ============================================================================
  # US-01: Reliable Connection Status and Auth Error Handling
  # ============================================================================

  # @driving_port @real-io @US-01
  @driving_port @real-io @US-01
  Scenario: Connection status shows "Connecting" before the handshake completes
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When Marcus opens the quizmaster panel with a valid host token
    Then the connection status shows "Connecting..." before the handshake is complete

  # @driving_port @real-io @US-01
  @driving_port @real-io @US-01
  Scenario: Connection status shows "Connected" only after the handshake succeeds
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When Marcus opens the quizmaster panel with a valid host token
    And the WebSocket handshake completes successfully
    Then the connection status shows "Connected"
    And the load quiz form is visible

  # @driving_port @real-io @US-01
  @driving_port @real-io @US-01
  Scenario: Wrong token shows a permanent auth error and stops retrying
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When Marcus opens the quizmaster panel with token "wrong-token"
    Then the connection status shows "Disconnected"
    And the message "Connection refused — invalid token. Check HOST_TOKEN and reload." is visible
    And no further connection attempts are made

  # @driving_port @real-io @US-01
  @driving_port @real-io @US-01
  Scenario: Mid-game network drop shows "Reconnecting" and preserves the round panel
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And Marcus has opened the quizmaster panel with a valid token and is in Round 2
    When the WebSocket connection drops unexpectedly
    Then the connection status shows "Reconnecting..."
    And the round panel remains visible
    When the WebSocket connection is restored
    Then the connection status shows "Connected"
    And game controls are available

  # @driving_port @real-io @US-01
  @driving_port @real-io @US-01
  Scenario: Reconnect attempts exhausted after 10 failures shows a reload prompt
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And Marcus has opened the quizmaster panel with a valid token
    When the WebSocket fails to reconnect 10 consecutive times
    Then the message "Could not reconnect. Please reload." is visible
    And a "Reload" button is visible
    And the game panel content is still visible beneath the overlay


  # ============================================================================
  # US-02: Load Quiz with Confirmation
  # ============================================================================

  # @driving_port @real-io @US-02
  @driving_port @real-io @US-02
  Scenario: Load quiz form is visible immediately after connecting
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When Marcus opens the quizmaster panel with a valid host token
    And the WebSocket handshake completes successfully
    Then a file path input labeled "Quiz file path" is visible
    And a "Load Quiz" button is visible
    And no "Start Round" button is visible

  # @driving_port @real-io @US-02
  @driving_port @real-io @US-02
  Scenario: Successful quiz load shows confirmation and all session URLs
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "pub-night-vol3.yaml" with 3 rounds of 5 text questions each
    And Marcus has opened the quizmaster panel with a valid host token
    When Marcus loads "pub-night-vol3.yaml" through the quizmaster panel
    Then the quizmaster panel shows "Pub Night Vol. 3 | 3 rounds | 15 questions"
    And the player join URL is displayed
    And the display URL is displayed
    And the "Start Round 1: Round 1" button is visible

  # @skip @driving_port @real-io @US-02
  @skip @driving_port @real-io @US-02
  Scenario: Quiz load with a file that does not exist shows an inline error
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And Marcus has opened the quizmaster panel with a valid host token
    When Marcus attempts to load "/quizzes/does-not-exist.yaml" through the quizmaster panel
    Then an error message is displayed below the file path input
    And the file path input remains editable
    And no "Start Round" button is visible

  # @skip @US-02
  @skip @US-02
  Scenario: Submitting an empty file path is blocked before sending to the server
    Given Marcus has opened the quizmaster panel with a valid host token
    When Marcus submits the load quiz form with an empty file path
    Then no host_load_quiz command is sent to the server
    And the validation message "Please enter a quiz file path." is visible below the input


  # ============================================================================
  # US-03: Run a Round — Start, Reveal Questions, End
  # ============================================================================

  # @skip @driving_port @real-io @US-03
  @skip @driving_port @real-io @US-03
  Scenario: Starting a round sends the correct command and shows the round panel
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "round-test.yaml" with 1 round of 5 text questions
    And Marcus has loaded "round-test.yaml" through the quizmaster panel
    When Marcus starts Round 1
    Then the host panel receives confirmation that Round 1 has started
    And the round panel shows "Round 1" and "0 of 5 revealed"
    And the "Reveal Next Question" button is visible

  # @skip @driving_port @real-io @US-03
  @skip @driving_port @real-io @US-03
  Scenario: Revealing a question appends it to the revealed list and increments the counter
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "round-test.yaml" with 1 round of 5 text questions
    And Marcus has started Round 1
    When Marcus reveals question 1
    Then the first revealed question appears in the question list
    And the round panel shows "1 of 5 revealed"

  # @skip @driving_port @real-io @US-03
  @skip @driving_port @real-io @US-03
  Scenario: Questions are revealed in sequential order matching the quiz file
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "seq-test.yaml" with 1 round of 3 text questions
    And Marcus has started Round 1
    When Marcus reveals question 1
    And Marcus reveals question 2
    And Marcus reveals question 3
    Then the revealed question list shows 3 questions in order
    And the round panel shows "3 of 3 revealed"

  # @skip @driving_port @real-io @US-03
  @skip @driving_port @real-io @US-03
  Scenario: Revealing the last question replaces "Reveal Next Question" with "End Round"
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "round-test.yaml" with 1 round of 5 text questions
    And Marcus has started Round 1 and revealed 4 of 5 questions
    When Marcus reveals question 5
    Then the "Reveal Next Question" button is no longer visible
    And the "End Round" button is visible
    And the round panel shows "5 of 5 revealed"

  # @skip @driving_port @real-io @US-03
  @skip @driving_port @real-io @US-03
  Scenario: Ending the round sends the end command followed by the scoring command
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "round-test.yaml" with 1 round of 5 text questions
    And Marcus has revealed all 5 questions in Round 1
    When Marcus clicks "End Round"
    Then the host panel receives confirmation that Round 1 has ended
    And the scoring panel becomes visible


  # ============================================================================
  # US-04: Score a Round — Mark Answers and Publish Scores
  # ============================================================================

  # @skip @driving_port @real-io @US-04
  @skip @driving_port @real-io @US-04
  Scenario: Scoring panel shows each question with its correct answer and team submissions
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "score-test.yaml" with 1 round of 2 text questions
    And "The Brainiacs" is connected in the lobby
    And Round 1 has ended and all 2 questions have been revealed
    And "The Brainiacs" has entered answers for all 2 questions
    When scoring is open for Round 1
    Then the scoring panel is visible
    And the scoring panel shows the correct answer for each question
    And "The Brainiacs" submitted answers are listed under each question
    And each team row has a "Correct" button and a "Wrong" button

  # @skip @driving_port @real-io @US-04
  @skip @driving_port @real-io @US-04
  Scenario: Marking a team answer as correct increases the running total and highlights the button
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "score-test.yaml" with 1 round of 2 text questions
    And "The Brainiacs" is connected in the lobby
    And scoring is open for Round 1
    And "The Brainiacs" submitted "Paris" for question 1 (correct answer: "Paris")
    When Marcus marks "The Brainiacs" answer for question 1 as correct
    Then the running total for "The Brainiacs" increases by 1 point
    And the "Correct" button for "The Brainiacs" on question 1 is visually marked as applied

  # @skip @driving_port @real-io @US-04
  @skip @driving_port @real-io @US-04
  Scenario: Marking a team answer as wrong leaves the running total unchanged
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "score-test.yaml" with 1 round of 2 text questions
    And "Quiz Killers" is connected in the lobby
    And scoring is open for Round 1
    And "Quiz Killers" submitted "Lyon" for question 1 (correct answer: "Paris")
    When Marcus marks "Quiz Killers" answer for question 1 as wrong
    Then the running total for "Quiz Killers" is unchanged
    And the "Wrong" button for "Quiz Killers" on question 1 is visually marked as applied

  # @skip @driving_port @real-io @US-04
  @skip @driving_port @real-io @US-04
  Scenario: Publishing scores makes the next-round and end-game controls available
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "multi-round.yaml" with 2 rounds of 2 text questions each
    And "The Brainiacs" is connected in the lobby
    And all answers for Round 1 have been marked correct or wrong
    When Marcus publishes scores for Round 1
    Then the round score summary is shown to Marcus
    And the "Start Round 2" button is visible
    And the "End Game" button is visible

  # @skip @driving_port @real-io @US-04
  @skip @driving_port @real-io @US-04
  Scenario: Marcus can publish scores without marking all answers — partial scoring allowed
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "partial-score.yaml" with 1 round of 3 text questions
    And "The Brainiacs" is connected in the lobby
    And scoring is open for Round 1
    And "The Brainiacs" submitted answers for all 3 questions
    When Marcus marks "The Brainiacs" answer for question 1 as correct
    And Marcus publishes scores for Round 1 without marking questions 2 and 3
    Then the host panel accepts the publish without error
    And the "End Game" button is visible


  # ============================================================================
  # US-05: Run Answer Ceremony on Display
  # ============================================================================

  # @skip @driving_port @real-io @US-05
  @skip @driving_port @real-io @US-05
  Scenario: Ceremony panel appears after publishing round scores
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "ceremony-test.yaml" with 1 round of 2 text questions
    And "The Brainiacs" is connected in the lobby
    And the display interface is connected
    And round scores have been published
    When Marcus starts the answer ceremony
    Then the ceremony panel is visible
    And the ceremony progress shows "Question 0 of 2 shown"
    And the "Show Next Question" button is visible

  # @skip @driving_port @real-io @US-05
  @skip @driving_port @real-io @US-05
  Scenario: Showing the next ceremony question sends it to both display and play rooms
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "ceremony-test.yaml" with 1 round of 2 text questions
    And "The Brainiacs" is connected in the lobby
    And the display interface is connected
    And Marcus is on the ceremony panel for Round 1
    When Marcus clicks "Show Next Question" for question 1
    Then the display screen receives question 1
    And the ceremony progress shows "Question 1 of 2 shown"
    And the "Reveal Answer" button is now visible

  # @skip @driving_port @real-io @US-05
  @skip @driving_port @real-io @US-05
  Scenario: Revealing the answer sends it to display only — play room does not receive it
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "ceremony-test.yaml" with 1 round of 2 text questions
    And "The Brainiacs" is connected in the lobby
    And the display interface is connected
    And Marcus is showing question 1 on the ceremony panel
    When Marcus clicks "Reveal Answer" for question 1
    Then the display screen receives the answer for question 1
    And the play screen for "The Brainiacs" does not receive the answer

  # @skip @driving_port @real-io @US-05
  @skip @driving_port @real-io @US-05
  Scenario: All questions walked through — ceremony complete shown and next controls available
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "ceremony-test.yaml" with 1 round of 2 text questions
    And the display interface is connected
    And Marcus has shown and revealed answers for all 2 questions in Round 1
    Then the ceremony progress shows "Question 2 of 2 shown"
    And the message "Ceremony complete" is visible
    And the "Show Next Question" button is no longer visible
    And the "End Game" button is visible


  # ============================================================================
  # US-06: End Game and View Final Leaderboard
  # ============================================================================

  # @skip @driving_port @real-io @US-06
  @skip @driving_port @real-io @US-06
  Scenario: End Game sends the game-over command and displays the final leaderboard
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "final-game.yaml" with 1 round of 2 text questions
    And "The Brainiacs" is connected in the lobby
    And "Quiz Killers" is connected in the lobby
    And Round 1 has been fully played, scored, and ceremonialized
    When Marcus ends the game
    Then the final leaderboard is displayed
    And the leaderboard shows all teams sorted by score from highest to lowest
    And rank indicators (1st, 2nd, etc.) are displayed next to each team
    And game control buttons are no longer visible

  # @skip @driving_port @real-io @US-06
  @skip @driving_port @real-io @US-06
  Scenario: Final leaderboard with tied teams shows both at the same rank position
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "tie-game.yaml" with 1 round of 2 text questions
    And "The Brainiacs" is connected in the lobby
    And "Quiz Killers" is connected in the lobby
    And Round 1 has been fully played and both teams have equal scores
    When Marcus ends the game
    Then the final leaderboard is displayed
    And "The Brainiacs" and "Quiz Killers" appear at the same rank position

  # @skip @driving_port @real-io @US-06
  @skip @driving_port @real-io @US-06
  Scenario: Marcus can end the game early after only one of three rounds
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "three-round-game.yaml" with 3 rounds of 2 text questions each
    And "The Brainiacs" is connected in the lobby
    And Round 1 has been fully played, scored, and ceremonialized
    When Marcus ends the game before playing Round 2
    Then the final leaderboard is displayed with scores from Round 1 only
    And no error is shown


  # ============================================================================
  # INFRASTRUCTURE FAILURE SCENARIOS
  # (Focused tests with in-memory server or direct port invocation)
  # ============================================================================

  # @skip @infrastructure-failure @in-memory @US-02
  @skip @infrastructure-failure @in-memory @US-02
  Scenario: Quiz load fails when the specified file cannot be read from disk
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And Marcus has opened the quizmaster panel with a valid host token
    When Marcus attempts to load "/tmp/nonexistent-quiz.yaml" through the quizmaster panel
    Then the server sends an error event in response
    And the error is displayed below the file path input
    And the file path input remains editable
    And no round controls appear

  # @skip @infrastructure-failure @in-memory @US-02
  @skip @infrastructure-failure @in-memory @US-02
  Scenario: Quiz load fails when the file path contains no extension or is malformed
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And Marcus has opened the quizmaster panel with a valid host token
    When Marcus attempts to load "/tmp/notaquiz" through the quizmaster panel
    Then the server sends an error event in response
    And an error message is displayed below the file path input

  # @skip @infrastructure-failure @in-memory @US-04
  @skip @infrastructure-failure @in-memory @US-04
  Scenario: Scoring command rejected when round has not been started
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And Marcus has opened the quizmaster panel with a valid host token
    And a quiz file "infra-test.yaml" with 1 round of 2 text questions
    And Marcus has loaded "infra-test.yaml" through the quizmaster panel
    When Marcus sends a mark-answer command before starting a round
    Then the server sends an error event in response

  # @skip @infrastructure-failure @in-memory @US-03
  @skip @infrastructure-failure @in-memory @US-03
  Scenario: Starting a round with an invalid round index is rejected by the server
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "infra-test.yaml" with 1 round of 2 text questions
    And Marcus has loaded "infra-test.yaml" through the quizmaster panel
    When Marcus sends a start-round command with round index 99
    Then the server sends an error event in response
    And the quizmaster panel remains in the quiz-loaded state


  # ============================================================================
  # ADAPTER INTEGRATION — REAL I/O
  # At least one scenario per driven adapter exercising real I/O
  # ============================================================================

  # @skip @real-io @adapter-integration @US-02
  @skip @real-io @adapter-integration @US-02
  Scenario: Quiz loader reads a real YAML file from disk and confirms the content
    Given a quiz file "real-io-quiz.yaml" with 2 rounds of 3 text questions each
    And the server is running with HOST_TOKEN "pub-night-secret"
    And Marcus has opened the quizmaster panel with a valid host token
    When Marcus loads "real-io-quiz.yaml" through the quizmaster panel
    Then the host panel receives a quiz confirmation with 2 rounds and 6 questions
    And no error is shown

  # @skip @real-io @adapter-integration @US-01
  @skip @real-io @adapter-integration @US-01
  Scenario: WebSocket upgrade is rejected with a real HTTP 403 for a wrong token
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When Marcus dials the WebSocket endpoint with token "definitely-wrong"
    Then the WebSocket dial is refused with an abnormal close
    And no messages are received on the connection

  # @skip @infrastructure-failure @in-memory @US-03
  @skip @infrastructure-failure @in-memory @US-03
  Scenario: Revealing a question out of order is rejected by the server
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And a quiz file "order-test.yaml" with 1 round of 3 text questions
    And Marcus has started Round 1
    When Marcus sends a reveal-question command with question index 5 before revealing earlier questions
    Then the server sends an error event in response
