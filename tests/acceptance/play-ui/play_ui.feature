Feature: Play UI — Player Interface
  As a player (Priya or Jordan)
  I want a mobile-friendly interface at /play that connects me to the game via WebSocket
  So that I can join, answer questions, submit, and follow the ceremony on my phone

  # ============================================================================
  # WALKING SKELETON (Strategy C — Real Local WebSocket)
  # Answers: "Can a player join, answer questions, submit, and see scores?"
  # Covers the complete game loop: register → lobby → question → review → submit
  #         → ceremony → scores
  # Traces: US-01, US-03, US-05, US-06, US-08, US-09
  # ============================================================================

  @walking_skeleton @driving_port @real-io @US-01 @US-03 @US-05 @US-06 @US-08 @US-09
  Scenario: Player joins the game, answers questions, submits, and sees round scores
    Given a quiz file "friday-quiz.yaml" with 1 round of 3 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has loaded "friday-quiz.yaml"
    When "Team Awesome" connects to the play room
    And "Team Awesome" registers as a new team
    Then "Team Awesome" receives their team identity
    And "Team Awesome" is in the lobby waiting for the round to start

    When the quizmaster starts Round 1
    Then "Team Awesome" receives the round started notification
    And "Team Awesome" sees that Round 1 has begun with 3 questions to answer

    When the quizmaster reveals question 1
    Then "Team Awesome" receives the first question on their device

    When the quizmaster reveals question 2
    And the quizmaster reveals question 3
    Then "Team Awesome" has received all 3 questions for Round 1

    When "Team Awesome" saves a draft answer "Paris" for Round 1 question 1
    And "Team Awesome" saves a draft answer "Shakespeare" for Round 1 question 2
    And "Team Awesome" saves a draft answer "Mercury" for Round 1 question 3

    When the quizmaster ends Round 1
    Then "Team Awesome" receives the round ended notification

    When "Team Awesome" submits their Round 1 answers
    Then "Team Awesome" receives confirmation that their answers are locked in
    And the play room receives a notification that "Team Awesome" has submitted

    When the quizmaster shows the ceremony question 1
    Then "Team Awesome" receives the ceremony question on their device

    When the quizmaster reveals the answer for ceremony question 1
    Then "Team Awesome" receives the revealed answer with team verdicts
    And the verdicts show whether each team answered correctly

    When the quizmaster publishes Round 1 scores
    Then "Team Awesome" receives the round scores with each team's name and total
    And the scores list includes "Team Awesome" with their round score


  # ============================================================================
  # US-01: Team Registration
  # ============================================================================

  @skip @driving_port @real-io @US-01
  Scenario: Player registers a unique team name and receives their team identity
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When "Team Awesome" connects to the play room
    And "Team Awesome" registers as a new team
    Then "Team Awesome" receives their team identity
    And the team identity includes a team identifier and a device token

  @skip @driving_port @real-io @US-01
  Scenario: Player receives a duplicate name rejection when their team name is already taken
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And "The Brainiacs" is already registered in the game
    When "The Brainiacs" attempts to register from a second device
    Then the second device receives a name-already-taken error

  @skip @driving_port @real-io @US-01
  Scenario: Player connects to the play room and receives an initial game state snapshot
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When "Team Awesome" connects to the play room
    Then "Team Awesome" receives the current game state
    And the game state indicates the game is in the lobby


  # ============================================================================
  # US-02: Auto-Rejoin and Device Recognition
  # ============================================================================

  @skip @driving_port @real-io @US-02
  Scenario: Player rejoins the game and receives their saved draft answers
    Given a quiz file "rejoin-quiz.yaml" with 1 round of 4 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has loaded "rejoin-quiz.yaml"
    And "Team Awesome" has registered and Round 1 is active with 4 questions revealed
    And "Team Awesome" has saved a draft answer "Paris" for Round 1 question 1
    When "Team Awesome" reconnects with their stored device token
    Then "Team Awesome" receives a game state snapshot
    And the snapshot includes "Team Awesome"'s previously saved draft answers
    And the snapshot shows the game is in round active state

  @skip @driving_port @real-io @US-02
  Scenario: Player rejoins during ceremony and is routed to the ceremony screen
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And the game is in the ceremony phase for Round 1
    And "Team Awesome" has a valid device token from earlier in the game
    When "Team Awesome" reconnects with their stored device token
    Then "Team Awesome" receives a game state snapshot showing the ceremony is in progress

  @skip @driving_port @real-io @US-02
  Scenario: Player rejoin is rejected when their device token is not recognised
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When a player attempts to rejoin with an unrecognised device token
    Then they receive a team-not-found error
    And no game state snapshot is sent


  # ============================================================================
  # US-03: Question Reveal and Answer Capture
  # ============================================================================

  @skip @driving_port @real-io @US-03
  Scenario: Play room receives the question when the quizmaster reveals it
    Given a quiz file "reveal-quiz.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    When the quizmaster reveals question 1
    Then "Team Awesome" receives the question on their device
    And the question includes the question text

  @skip @driving_port @real-io @US-03
  Scenario: Questions accumulate as the quizmaster reveals them one by one
    Given a quiz file "multi-reveal.yaml" with 1 round of 3 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    When the quizmaster reveals question 1
    And the quizmaster reveals question 2
    And the quizmaster reveals question 3
    Then "Team Awesome" has received all 3 questions for Round 1

  @skip @driving_port @real-io @US-03
  Scenario: Player joining mid-round receives all previously revealed questions in the state snapshot
    Given a quiz file "mid-round.yaml" with 1 round of 5 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 and revealed 4 of 5 questions
    When "Quiz Killers" connects to the play room and registers
    And "Quiz Killers" requests a game state snapshot
    Then the snapshot for "Quiz Killers" includes all 4 revealed questions

  @skip @driving_port @real-io @US-03
  Scenario: Player saves a draft answer to the server for a revealed question
    Given a quiz file "draft-quiz.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    And the quizmaster has revealed question 1
    When "Team Awesome" saves a draft answer "Paris" for Round 1 question 1
    Then the draft is saved on the server without error


  # ============================================================================
  # US-04: Draft Answer Persistence (Rejoin)
  # ============================================================================

  @skip @driving_port @real-io @US-04
  Scenario: Draft answers are included in the state snapshot when a player rejoins
    Given a quiz file "persist-quiz.yaml" with 1 round of 3 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    And the quizmaster has revealed question 1
    And "Team Awesome" has saved a draft answer "Mercury" for Round 1 question 1
    When "Team Awesome" reconnects with their stored device token
    Then the state snapshot contains "Team Awesome"'s draft answer for Round 1 question 1

  @skip @driving_port @real-io @US-04
  Scenario: Draft answers are not included in a fresh connection state snapshot
    Given a quiz file "fresh-quiz.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1
    When "Quiz Killers" connects to the play room and registers
    Then the state snapshot for "Quiz Killers" contains no draft answers


  # ============================================================================
  # US-05: End-of-Round Review
  # ============================================================================

  @skip @driving_port @real-io @US-05
  Scenario: Play room is notified when the quizmaster ends the round
    Given a quiz file "review-quiz.yaml" with 1 round of 3 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    And the quizmaster has revealed all 3 questions
    When the quizmaster ends Round 1
    Then "Team Awesome" receives the round ended notification
    And the round ended notification includes the round number


  # ============================================================================
  # US-06: Submission Confirmation and Locking
  # ============================================================================

  @skip @driving_port @real-io @US-06
  Scenario: Team submits answers and receives a locked confirmation
    Given a quiz file "submit-quiz.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 and ended Round 1 with "Team Awesome" in the play room
    When "Team Awesome" submits their Round 1 answers
    Then "Team Awesome" receives confirmation that their answers are locked in
    And the confirmation shows the answers are locked for Round 1

  @skip @driving_port @real-io @US-06
  Scenario: Team submitting a second time receives an already-submitted response
    Given a quiz file "double-submit.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 and ended Round 1 with "Team Awesome" in the play room
    And "Team Awesome" has already submitted their Round 1 answers
    When "Team Awesome" submits their Round 1 answers again
    Then "Team Awesome" receives an already-submitted error

  @skip @driving_port @real-io @US-06
  Scenario: Submission is rejected when the round has not yet started
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" is registered in the play room
    When "Team Awesome" attempts to submit answers for Round 1 before the round starts
    Then "Team Awesome" receives an error response

  @skip @driving_port @real-io @US-06
  Scenario: Team can submit with all answers blank
    Given a quiz file "blank-submit.yaml" with 1 round of 3 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 and ended Round 1 with "Team Awesome" in the play room
    When "Team Awesome" submits Round 1 answers with all fields empty
    Then "Team Awesome" receives confirmation that their answers are locked in


  # ============================================================================
  # US-07: Post-Submit Waiting Screen — submission_received to play room (DEP-03)
  # ============================================================================

  @skip @driving_port @real-io @US-07
  Scenario: Play room is notified in real time when another team submits their answers
    Given a quiz file "multi-team.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" and "The Brainiacs" are registered in the play room
    And the quizmaster has started Round 1 and ended Round 1
    And "Team Awesome" has already submitted their Round 1 answers
    When "The Brainiacs" submits their Round 1 answers
    Then "Team Awesome" receives a notification that "The Brainiacs" has submitted
    And the notification includes "The Brainiacs" team name

  @skip @driving_port @real-io @US-07
  Scenario: Submitting team receives a notification about their own submission in the play room
    Given a quiz file "self-notify.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" is registered in the play room
    And the quizmaster has started Round 1 and ended Round 1
    When "Team Awesome" submits their Round 1 answers
    Then "Team Awesome" receives a submission notification in the play room
    And the notification includes "Team Awesome" team name and round number


  # ============================================================================
  # US-08: Ceremony View — ceremony events to play room (DEP-02)
  # ============================================================================

  @skip @driving_port @real-io @US-08
  Scenario: Play room receives the ceremony question when the quizmaster shows it
    Given a quiz file "ceremony-quiz.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" has submitted for Round 1 and the quizmaster has started the ceremony
    When the quizmaster shows the ceremony question 1
    Then "Team Awesome" receives the ceremony question on their device
    And the ceremony question includes the question text

  @skip @driving_port @real-io @US-08
  Scenario: Play room receives the revealed answer and team verdicts during ceremony
    Given a quiz file "verdict-quiz.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" has submitted for Round 1 and the ceremony is at question 1
    When the quizmaster reveals the answer for ceremony question 1
    Then "Team Awesome" receives the revealed answer with team verdicts
    And the verdicts include a result for "Team Awesome"
    And each verdict shows whether the team answered correctly or not

  @skip @driving_port @real-io @US-08
  Scenario: Play room receives ceremony events for each question in sequence
    Given a quiz file "seq-ceremony.yaml" with 1 round of 3 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" has submitted for Round 1 and the quizmaster has started the ceremony
    When the quizmaster shows and reveals the answer for ceremony question 1
    And the quizmaster shows and reveals the answer for ceremony question 2
    And the quizmaster shows and reveals the answer for ceremony question 3
    Then "Team Awesome" has received 3 ceremony question events
    And "Team Awesome" has received 3 answer reveal events

  @skip @driving_port @real-io @US-08
  Scenario: Player rejoining during ceremony receives a ceremony-state snapshot
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And the game is in the ceremony phase for Round 1
    And "Team Awesome" has a valid device token from earlier in the game
    When "Team Awesome" reconnects with their stored device token
    Then the state snapshot shows the game is in ceremony state


  # ============================================================================
  # US-09: Round Scores Display — team names in payload (DEP-04)
  # ============================================================================

  @skip @driving_port @real-io @US-09
  Scenario: Play room receives round scores with team names when the quizmaster publishes them
    Given a quiz file "scores-quiz.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" and "The Brainiacs" have completed Round 1 with the quizmaster scoring
    When the quizmaster publishes Round 1 scores
    Then "Team Awesome" receives the round scores
    And the scores list includes team names alongside each score
    And the scores list includes "Team Awesome" with their round score and running total
    And the scores list includes "The Brainiacs" with their round score and running total

  @skip @driving_port @real-io @US-09
  Scenario: Play room receives the final leaderboard with team names at game over
    Given a quiz file "final-quiz.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" and "The Brainiacs" have completed Round 1 and scores are published
    When the quizmaster ends the game
    Then "Team Awesome" receives the final scores notification
    And the final scores include team names and totals for all teams

  @skip @driving_port @real-io @US-09
  Scenario: Next round starts from the scores screen
    Given a quiz file "two-round.yaml" with 2 rounds of 2 text questions each
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" is on the Round 1 scores screen
    When the quizmaster starts Round 2
    Then "Team Awesome" receives the round started notification for Round 2


  # ============================================================================
  # US-10: Connection Status and State Restore on Reconnect
  # ============================================================================

  @skip @driving_port @real-io @US-10
  Scenario: Player reconnecting receives a state snapshot restoring their game position
    Given a quiz file "reconnect-quiz.yaml" with 1 round of 3 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" is registered with Round 1 active and 3 questions revealed
    When "Team Awesome" reconnects with their stored device token
    Then "Team Awesome" receives a game state snapshot
    And the snapshot shows 3 revealed questions for Round 1

  @skip @driving_port @real-io @US-10
  Scenario: Player reconnecting during scores phase receives a scores-state snapshot
    Given a quiz file "scores-reconnect.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the game is at the Round 1 scores screen with all teams having submitted
    And "Team Awesome" has a valid device token
    When "Team Awesome" reconnects with their stored device token
    Then the state snapshot shows the game is in the round scores phase


  # ============================================================================
  # US-11: Multiple Choice Questions (DEP-01 — choices field)
  # ============================================================================

  @skip @driving_port @real-io @US-11
  Scenario: Play room receives a multiple choice question with the choices list
    Given a quiz file "mc-quiz.yaml" with 1 round including a multiple choice question
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    When the quizmaster reveals the multiple choice question
    Then "Team Awesome" receives the question with a non-empty list of answer choices
    And the choices list contains 4 options

  @skip @driving_port @real-io @US-11
  Scenario: Text questions are revealed without a choices list
    Given a quiz file "text-only.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    When the quizmaster reveals question 1
    Then "Team Awesome" receives the question with no choices list


  # ============================================================================
  # US-12: Multi-Part Answers (DEP-01 — is_multi_part field)
  # ============================================================================

  @skip @driving_port @real-io @US-12
  Scenario: Play room receives a multi-part question with the multi-part indicator
    Given a quiz file "multipart-quiz.yaml" with 1 round including a multi-part question
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    When the quizmaster reveals the multi-part question
    Then "Team Awesome" receives the question with the multi-part indicator set

  @skip @driving_port @real-io @US-12
  Scenario: Single-answer questions are revealed without the multi-part indicator
    Given a quiz file "single-answer.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    When the quizmaster reveals question 1
    Then "Team Awesome" receives the question without the multi-part indicator


  # ============================================================================
  # US-13: Media Questions (DEP-01 — media field)
  # ============================================================================

  @skip @driving_port @real-io @US-13
  Scenario: Play room receives an image question with the media attachment
    Given a quiz file "media-quiz.yaml" with 1 round including an image question
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    When the quizmaster reveals the image question
    Then "Team Awesome" receives the question with a media reference
    And the media reference includes the media type and a URL

  @skip @driving_port @real-io @US-13
  Scenario: Text questions are revealed without a media attachment
    Given a quiz file "no-media.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 with "Team Awesome" in the play room
    When the quizmaster reveals question 1
    Then "Team Awesome" receives the question with no media attachment


  # ============================================================================
  # INFRASTRUCTURE FAILURE SCENARIOS
  # (Behaviour at invalid state or protocol boundary)
  # ============================================================================

  @skip @infrastructure-failure @in-memory @US-01
  Scenario: Team registration rejected when the team name is empty
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When a player attempts to register with an empty team name
    Then the player receives an error response
    And no team identity is issued

  @skip @infrastructure-failure @in-memory @US-06
  Scenario: Submit rejected when the team identifier is unknown
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When a player attempts to submit answers using an unknown team identifier
    Then the player receives an error response

  @skip @infrastructure-failure @in-memory @US-06
  Scenario: Submit rejected before a round has started
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" is registered in the play room
    When "Team Awesome" attempts to submit answers for Round 1 before the round starts
    Then "Team Awesome" receives an error response

  @skip @infrastructure-failure @in-memory @US-02
  Scenario: Rejoin with a malformed device token receives a team-not-found error
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When a player sends a rejoin request with a token that does not match any registered team
    Then the player receives a team-not-found error

  @skip @infrastructure-failure @in-memory @US-06
  Scenario: Second submission from the same team is treated as already submitted
    Given a quiz file "idempotent.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And the quizmaster has started Round 1 and ended Round 1 with "Team Awesome" in the play room
    And "Team Awesome" has already submitted their Round 1 answers
    When "Team Awesome" submits their Round 1 answers again
    Then "Team Awesome" receives an already-submitted error

  @skip @infrastructure-failure @in-memory @US-03
  Scenario: Draft answer sent before a round has started is silently accepted
    Given the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" is registered in the play room
    When "Team Awesome" saves a draft answer "Paris" for Round 1 question 1
    Then no error is returned to "Team Awesome"


  # ============================================================================
  # ADAPTER INTEGRATION — REAL I/O
  # At least one scenario per driven adapter exercising real I/O
  # ============================================================================

  @skip @real-io @adapter-integration @US-01
  Scenario: Player connecting to the play room receives an immediate game state snapshot
    Given the server is running with HOST_TOKEN "pub-night-secret"
    When "Team Awesome" connects to the play room
    Then the connection is accepted and "Team Awesome" receives a state snapshot

  @skip @real-io @adapter-integration @US-09
  Scenario: Round scores received by the player include team names alongside each score
    Given a quiz file "adapter-scores.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" has completed Round 1 with the quizmaster scoring
    When the quizmaster publishes Round 1 scores
    Then "Team Awesome" receives the round scores notification
    And the scores notification includes a structured list with a team name in each entry

  @skip @real-io @adapter-integration @US-08
  Scenario: Ceremony answer received by the player includes a result for each team
    Given a quiz file "adapter-ceremony.yaml" with 1 round of 2 text questions
    And the server is running with HOST_TOKEN "pub-night-secret"
    And "Team Awesome" has submitted for Round 1 and the ceremony is at question 1
    When the quizmaster reveals the answer for ceremony question 1
    Then "Team Awesome" receives the revealed answer with team verdicts
    And the verdicts list is present in the answer notification
