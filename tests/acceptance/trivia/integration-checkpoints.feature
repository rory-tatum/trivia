# Integration Checkpoints -- Trivia Game
# Feature ID: trivia
# Release: 1 (Text-Only, Full Game Loop)
#
# Integration checkpoints validate the critical seams between components:
# where the WebSocket hub, game engine, and answer boundary enforcer meet.
# These are not unit tests -- they go through the real server driving ports.
#
# Critical invariant under test: QuestionFull fields MUST NEVER appear in
# WebSocket messages sent to the play room or display room.
#
# All scenarios @skip until walking skeleton passes.

Feature: Integration checkpoints -- component boundary verification

  Background:
    Given a quiz file "checkpoint-quiz.yaml" exists with 1 round and 3 text questions
    And question 1 has text "What is the capital of France?" and answer "Paris"
    And question 2 has text "What color is the sky?" and answer "Blue"
    And question 3 has text "How many days in a week?" and answer "Seven"
    And the server is running with HOST_TOKEN "test-secret-token"

  # -----------------------------------------------------------------------
  # IC-01: Game start broadcasts to all rooms within 1 second
  # Driving port: host_start_round via /ws host room
  # -----------------------------------------------------------------------

  @skip
  Scenario: Game start event reaches all connected rooms within the timing threshold
    Given Marcus opens the quizmaster panel with a valid token
    And Marcus has loaded "checkpoint-quiz.yaml"
    And Priya is connected in the player room
    And the display screen is connected
    When Marcus starts round 1 via the quizmaster interface
    Then the player room receives the round started event within 1 second
    And the display room receives the round started event within 1 second
    And Marcus sees the Round 1 reveal panel

  # -----------------------------------------------------------------------
  # IC-02: Question reveal reaches all rooms within 1 second
  # Driving port: host_reveal_question via /ws host room
  # -----------------------------------------------------------------------

  @skip
  Scenario: Question reveal event reaches player and display rooms within the timing threshold
    Given Marcus has started round 1
    When Marcus reveals question 1 via the quizmaster interface
    Then the player room receives the question_revealed event within 1 second
    And the display room receives the question_revealed event within 1 second

  # -----------------------------------------------------------------------
  # IC-03: Submission acknowledgment returns to submitting client
  # Driving port: submit_answers via /ws play room
  # -----------------------------------------------------------------------

  @skip
  Scenario: Submission acknowledgment is sent back to the submitting team before locked state is shown
    Given Marcus has started round 1 and revealed all 3 questions
    And Marcus has ended round 1
    And Priya has entered answers for all 3 questions as "Team Awesome"
    When "Team Awesome" submits their answers via the player interface
    Then the server sends a submission acknowledgment to the "Team Awesome" connection
    And the "Team Awesome" player screen shows "Your answers are locked in" only after the acknowledgment

  # -----------------------------------------------------------------------
  # CRITICAL INVARIANT: Answer boundary -- QuestionFull never sent to play or display
  # DEC-010, DEC-018: Structural dual-type boundary enforced at the hub
  # Driving port: host_reveal_question (tests what the hub broadcasts)
  # -----------------------------------------------------------------------

  @skip
  Scenario: Question revealed to players contains no answer or answers fields
    Given Marcus has started round 1
    When Marcus reveals question 1 via the quizmaster interface
    Then the message received by the player room for question_revealed has no "answer" field
    And the message received by the player room for question_revealed has no "answers" field
    And the message contains the question text "What is the capital of France?"

  @skip
  Scenario: Question revealed to the display contains no answer or answers fields
    Given Marcus has started round 1
    When Marcus reveals question 1 via the quizmaster interface
    Then the message received by the display room for question_revealed has no "answer" field
    And the message received by the display room for question_revealed has no "answers" field

  @skip
  Scenario: Ceremony question shown to display does not include the answer until reveal step
    Given all 3 questions are scored and Marcus has started the ceremony
    When Marcus sends the show ceremony question event for question 1
    Then the display room receives the ceremony_question_shown event with no "answer" field
    When Marcus sends the reveal ceremony answer event for question 1
    Then the display room receives the ceremony_answer_revealed event with answer "Paris"
    And the player room does not receive the ceremony_answer_revealed event

  @skip
  @property
  Scenario: Answer fields are never present in any message sent to the play room
    Given the game progresses through any valid sequence of state transitions
    When all WebSocket messages sent to the play room are inspected
    Then none of the messages contain a field named "answer"
    And none of the messages contain a field named "answers"

  @skip
  @property
  Scenario: Answer fields are never present in question_revealed messages to any non-host room
    Given the game is in round_active state
    When Marcus reveals any question via the quizmaster interface
    Then every question_revealed message sent to play or display rooms contains only QuestionPublic fields

  # -----------------------------------------------------------------------
  # IC-04: State snapshot sent to connecting clients reflects current state
  # Driving port: WebSocket connection (new or reconnecting) via /ws
  # -----------------------------------------------------------------------

  @skip
  Scenario: A new player connection receives the current game state as a snapshot
    Given Marcus has started round 1 and revealed questions 1 and 2
    When a new player connection is established and registers as "Late Squad"
    Then the connection receives a state_snapshot event
    And the snapshot includes round 1 as active
    And the snapshot includes questions 1 and 2 as revealed
    And the snapshot contains no answer fields for any question

  # -----------------------------------------------------------------------
  # IC-05: Draft answer persistence and retrieval
  # Driving port: draft_answer via /ws play room, then team_rejoin
  # -----------------------------------------------------------------------

  @skip
  Scenario: Draft answers are retrievable after a player reconnects
    Given Priya has entered "Paris" as a draft answer for question 1 as "Team Awesome"
    And "Team Awesome"'s connection is interrupted and restored
    When "Team Awesome" sends a team_rejoin event with their stored token
    Then the state_snapshot received by "Team Awesome" includes the draft answer "Paris" for question 1

  # -----------------------------------------------------------------------
  # IC-06: Auth guard rejects unauthenticated host room connections
  # Driving port: /ws host room with invalid token
  # -----------------------------------------------------------------------

  @skip
  Scenario: A WebSocket connection to the host room without a valid token is rejected
    When a connection attempts to upgrade to the host WebSocket room without a valid token
    Then the connection is rejected with a 403 status
    And the game state is not affected

  @skip
  Scenario: A valid token allows connection to the host room
    When Marcus connects to the host room WebSocket with the correct token
    Then the connection is accepted
    And Marcus receives the current game state snapshot
