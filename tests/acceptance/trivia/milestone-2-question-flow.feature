Feature: Milestone 2 -- Question Reveal and Answer Entry
  As Marcus the quizmaster and Priya the team captain
  I want questions to appear on player devices when revealed
  And players to enter, edit, and submit their answers
  So that the game round proceeds smoothly with all teams participating

  # All scenarios @skip until walking skeleton and milestone 1 pass.
  #
  # Driving ports:
  #   host_reveal_question -> /ws (broadcasts to play and display rooms with QuestionPublic)
  #   draft_answer         -> /ws (play room, server stores draft)
  #   submit_answers       -> /ws (play room, server writes ART-06, returns submission_ack)

  Background:
    Given Marcus has loaded "friday-march-2026.yaml" with 1 round of 8 text questions
    And "Team Awesome", "The Brainiacs", and "Quiz Killers" are connected in the lobby
    And Marcus has started the game

  # -------------------------------------------------------------------------
  # US-07: Quizmaster Reveals Questions One at a Time
  # -------------------------------------------------------------------------

  Scenario: Revealing question 1 sends it to all player screens
    Given no questions have been revealed yet
    When Marcus reveals question 1 via the quizmaster interface
    Then "Team Awesome"'s player screen shows "What is the capital of France?"
    And "The Brainiacs"' player screen shows "What is the capital of France?"
    And "Quiz Killers"' player screen shows "What is the capital of France?"
    And the display screen shows "What is the capital of France?" as the current question
    And the quizmaster panel shows "1 of 8 revealed"

  Scenario: Revealing a second question cumulates on player screens but replaces the display
    Given Marcus has revealed question 1 "What is the capital of France?"
    And Priya has entered "Paris" in the question 1 answer field
    When Marcus reveals question 2 "Name the three primary colors."
    Then Priya's player screen shows both question 1 and question 2
    And Priya's answer "Paris" for question 1 is preserved
    And the display screen shows only question 2 as the current question

  @skip
  Scenario: Questions are revealed in the correct sequence without skipping
    Given Marcus has revealed questions 1 through 6
    When Marcus reveals question 7
    Then all players see questions 1 through 7 on their screens
    And the display screen shows question 7 as the current question
    And the quizmaster panel shows "7 of 8 revealed"

  @skip
  Scenario: After all questions are revealed, end round becomes available
    Given Marcus has revealed all 8 questions
    When Marcus reviews the quizmaster panel
    Then the "End Round" action is available
    And the quizmaster panel shows "8 of 8 revealed"

  # -------------------------------------------------------------------------
  # US-08: Player Enters and Edits Answers
  # -------------------------------------------------------------------------

  Scenario: Player enters a text answer into a revealed question field
    Given question 1 "What is the capital of France?" has been revealed
    When Priya enters "Paris" in the question 1 answer field
    Then Priya's player screen shows "Paris" in the question 1 field
    And the draft is persisted for "Team Awesome"

  Scenario: Player changes an answer after initial entry
    Given Priya has entered "Paris, France" in the question 1 answer field
    When Priya changes the answer to "Paris"
    Then Priya's player screen shows "Paris" in the question 1 field
    And the previous value "Paris, France" is no longer shown

  @skip
  Scenario: Player leaves an answer blank and moves on
    Given questions 1 through 8 have all been revealed
    When Priya enters answers for questions 1 through 4 and question 6 through 8
    And Priya leaves question 5 blank
    Then question 5 shows an empty answer field with no error
    And all other entered answers are preserved

  @skip
  Scenario: Draft answers survive a player page refresh mid-round
    Given Priya has entered "Paris" for question 1 and "Mercury" for question 2
    When Priya refreshes her player page
    Then after rejoining as "Team Awesome" Priya sees "Paris" for question 1
    And "Mercury" for question 2

  # -------------------------------------------------------------------------
  # US-09: Player Submits Answers
  # -------------------------------------------------------------------------

  @skip
  Scenario: Player reviews all answers before submitting
    Given all 8 questions have been revealed
    And Priya has answered questions 1 through 8
    And Marcus has ended the round
    When Priya opens the submit review screen
    Then Priya sees a list showing all 8 of her answers
    And a "Submit Answers" action is available

  @skip
  Scenario: Player sees a blank question flagged on the review screen
    Given all 8 questions have been revealed
    And Priya has answered questions 1 through 4 and 6 through 8 but left question 5 blank
    And Marcus has ended the round
    When Priya opens the submit review screen
    Then question 5 is flagged with a blank warning
    And a "Go Back and Edit" action is available alongside "Submit Answers"

  @skip
  Scenario: Player confirms submission through the confirmation step
    Given Priya is on the submit review screen with all answers filled
    When Priya initiates submission
    Then Priya sees a confirmation asking her to confirm because answers cannot be changed after submission
    When Priya confirms the submission
    Then the server acknowledges the submission for "Team Awesome"
    And Priya sees "Your answers are locked in"
    And Marcus sees "Team Awesome" listed as submitted in the quizmaster panel

  @skip
  Scenario: Player cancels the submission confirmation and returns to editing
    Given Priya is on the submission confirmation step
    When Priya cancels the confirmation
    Then Priya is returned to the answer form
    And all previously entered answers are intact and editable

  @skip
  Scenario: Player cannot edit answers after submission is confirmed
    Given "Team Awesome" has confirmed their submission for round 1
    When Priya tries to edit the question 1 answer field
    Then the field is read-only
    And a banner reads "Answers are locked for round 1"

  @skip
  Scenario: Player screen does not show locked state until server confirms the submission
    Given Priya has initiated submission for "Team Awesome"
    When the submission is in progress waiting for server acknowledgment
    Then Priya sees a "Submitting..." indicator
    And the locked state message does not appear until acknowledgment arrives

  @skip
  Scenario: Resubmitting after a network interruption does not duplicate or overwrite the submission
    Given "Team Awesome" successfully submitted in round 1
    When the submission event is sent again due to a client retry
    Then the server re-sends the acknowledgment to the client
    And "Team Awesome"'s round 1 answers remain unchanged

  # -------------------------------------------------------------------------
  # US-10: Quizmaster Monitors Submission Status
  # -------------------------------------------------------------------------

  @skip
  Scenario: Quizmaster sees real-time submission status as teams submit
    Given Marcus has ended round 1 and the submission panel is open
    When "Team Awesome" submits their answers
    Then the quizmaster panel shows "Team Awesome" as submitted
    And "The Brainiacs" and "Quiz Killers" are still shown as waiting

  @skip
  Scenario: Quizmaster sees all teams submitted and scoring becomes available
    Given Marcus has ended round 1
    When "Team Awesome", "The Brainiacs", and "Quiz Killers" all submit
    Then the quizmaster panel shows all 3 teams as submitted
    And the "Open Scoring" action becomes available

  @skip
  Scenario: Scoring remains blocked while at least one team has not submitted
    Given "Team Awesome" and "The Brainiacs" have submitted
    And "Quiz Killers" has not submitted
    When Marcus reviews the quizmaster panel
    Then the "Open Scoring" action is not available
    And Marcus sees "1 team has not yet submitted"

  @skip
  Scenario: A team that submitted with a blank answer shows the blank in scoring
    Given "Quiz Killers" submitted with no answer for question 3
    When Marcus opens scoring
    Then question 3 for "Quiz Killers" shows a blank answer in the scoring grid
