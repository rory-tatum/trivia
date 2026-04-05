Feature: Milestone 3 -- Scoring, Ceremony, and Round Scores
  As Marcus the quizmaster
  I want to mark answers correct or wrong, step through the ceremony, and publish scores
  So that players receive the theatrical payoff of the round and see where they stand

  # All scenarios @skip until walking skeleton and milestones 1-2 pass.
  #
  # Driving ports:
  #   host_mark_answer             -> /ws (host room) -> updates ART-07, recomputes ART-08
  #   host_ceremony_show_question  -> /ws (host room) -> broadcasts to display room
  #   host_ceremony_reveal_answer  -> /ws (host room) -> broadcasts to display room only
  #   host_publish_scores          -> /ws (host room) -> broadcasts to all rooms
  #   host_end_game                -> /ws (host room) -> broadcasts to all rooms

  Background:
    Given Marcus has loaded "friday-march-2026.yaml" with 1 round of 8 text questions
    And "Team Awesome", "The Brainiacs", and "Quiz Killers" played the round
    And all three teams submitted their answers
    And Marcus has opened scoring

  # -------------------------------------------------------------------------
  # US-12: Quizmaster Scoring Interface
  # -------------------------------------------------------------------------

  @skip
  Scenario: Scoring interface shows all submitted answers for a question
    When Marcus opens the scoring panel for question 1
    Then Marcus sees the correct answer "Paris" for question 1
    And Marcus sees "Team Awesome" answered "paris"
    And Marcus sees "The Brainiacs" answered "Paris"
    And Marcus sees "Quiz Killers" answered "Lyon"

  Scenario: Quizmaster marks an answer correct and the score increments
    Given the scoring panel shows "Team Awesome" answered "paris" for question 1
    When Marcus marks "paris" as correct for "Team Awesome" question 1
    Then "Team Awesome"'s score increments by 1 point
    And the scoring panel shows "Team Awesome"'s updated running total

  @skip
  Scenario: Quizmaster marks an answer wrong and the score does not change
    Given the scoring panel shows "Quiz Killers" answered "Lyon" for question 1
    When Marcus marks "Lyon" as wrong for "Quiz Killers" question 1
    Then "Quiz Killers"' score does not change
    And the scoring panel shows "Lyon" marked as wrong

  @skip
  Scenario: Quizmaster toggles a verdict from wrong to correct
    Given Marcus marked "paris" as wrong for "Team Awesome" question 1 by mistake
    When Marcus changes the verdict to correct for "Team Awesome" question 1
    Then "Team Awesome"'s score increments by 1 point
    And the running total is recalculated immediately

  @skip
  Scenario: Scoring answers does not send answer content to player connections
    Given Marcus is on the scoring panel showing correct answers
    When Marcus marks answers for all teams
    Then no message containing answer field data is sent to the player connections
    And no message containing answer field data is sent to the display connection

  # -------------------------------------------------------------------------
  # US-14: Auto-Tally Scores
  # -------------------------------------------------------------------------

  @skip
  Scenario: Running totals are calculated automatically as verdicts are entered
    Given Marcus has marked 3 questions correct for "Team Awesome"
    And 2 questions correct for "The Brainiacs"
    And 1 question correct for "Quiz Killers"
    When Marcus views the score summary
    Then "Team Awesome"'s total shows 3
    And "The Brainiacs"' total shows 2
    And "Quiz Killers"' total shows 1

  @skip
  Scenario: Changing a verdict recalculates the total instantly
    Given "Team Awesome" has a total of 4 points after all 8 questions scored
    When Marcus changes question 3 verdict from correct to wrong for "Team Awesome"
    Then "Team Awesome"'s total immediately shows 3

  @skip
  Scenario: Final totals match the exact count of correct verdicts
    Given Marcus has marked 6 questions correct and 2 wrong for "The Brainiacs"
    When Marcus reviews the total for "The Brainiacs"
    Then "The Brainiacs"' total is exactly 6

  # -------------------------------------------------------------------------
  # US-15: Answer Ceremony
  # -------------------------------------------------------------------------

  @skip
  Scenario: Starting the ceremony transitions the display to ceremony mode
    Given all 8 questions are fully scored
    When Marcus starts the answer ceremony via the quizmaster interface
    Then the display screen shows question 1 text of the ceremony
    And no answer is shown yet on the display screen
    And the quizmaster panel shows the ceremony control for question 1

  @skip
  Scenario: Advancing the ceremony first shows the question then reveals the answer
    Given the ceremony is at question 2 and only the question text is shown
    When Marcus advances to reveal the answer for question 2
    Then the display screen shows the answer "Paris" below question 2 text
    And the player screens update to the ceremony step for question 2

  @skip
  Scenario: Ceremony answer reveal is shown only on the display, not the player screens
    Given the ceremony is in progress
    When Marcus reveals the answer for question 3 during ceremony
    Then the display screen shows the answer for question 3
    And no answer text is sent to the player connections at this ceremony step

  @skip
  Scenario: Quizmaster can navigate back to a previous ceremony question
    Given the ceremony has advanced to question 4
    When Marcus navigates back to question 3
    Then the display screen shows question 3 again with its answer
    And the ceremony progress indicator reflects question 3

  @skip
  Scenario: Ceremony completes only after all questions have been stepped through
    Given the ceremony has gone through questions 1 through 7
    When Marcus views the ceremony panel
    Then the "End Ceremony and Publish Scores" action is not yet available
    When Marcus steps through question 8 and reveals its answer
    Then the "End Ceremony and Publish Scores" action becomes available

  # -------------------------------------------------------------------------
  # US-16: Round Scores on Display
  # -------------------------------------------------------------------------

  @skip
  Scenario: Publishing scores shows ranked results on the display screen
    Given the ceremony for round 1 is complete
    And "Team Awesome" scored 8, "The Brainiacs" scored 6, "Quiz Killers" scored 4
    When Marcus publishes the round scores
    Then the display screen shows the round 1 scores in rank order:
      | Rank | Team         | Round Points | Running Total |
      | 1    | Team Awesome | 8            | 8             |
      | 2    | The Brainiacs| 6            | 6             |
      | 3    | Quiz Killers | 4            | 4             |

  @skip
  Scenario: Tied teams are shown at the same rank
    Given "Team Awesome" and "The Brainiacs" both scored 7 in round 1
    When Marcus publishes the round scores
    Then both "Team Awesome" and "The Brainiacs" appear at rank 1 on the display screen

  @skip
  Scenario: Player screens show the same scores as the display after publishing
    Given Marcus has published the round 1 scores
    When Priya views her player screen
    Then Priya sees the round 1 scores matching the display screen

  # -------------------------------------------------------------------------
  # US-18: Advance to Next Round
  # -------------------------------------------------------------------------

  @skip
  Scenario: Starting the next round resets the question reveal state
    Given round 1 scores are published and "friday-march-2026.yaml" has a round 2
    When Marcus starts round 2 via the quizmaster interface
    Then the quizmaster panel shows the round 2 reveal panel with 0 questions revealed
    And all player screens show the round 2 answer form with no questions yet
    And the display screen shows the round 2 waiting state

  @skip
  Scenario: Round 1 scores are preserved when round 2 begins
    Given "Team Awesome" scored 8 in round 1
    When Marcus starts round 2
    Then "Team Awesome"'s round 1 score of 8 is still visible in the running total

  @skip
  Scenario: The final round offers "End Game" instead of "Next Round"
    Given "friday-march-2026.yaml" has 4 rounds and all 4 rounds have been played and scored
    When Marcus publishes the round 4 scores
    Then the quizmaster panel offers "End Game" and not "Start Round 5"

  # -------------------------------------------------------------------------
  # US-19: Final Scores and Winner
  # -------------------------------------------------------------------------

  @skip
  Scenario: Ending the game shows final standings on the display screen
    Given all rounds have been played and scored
    And "Team Awesome" has the highest total across all rounds
    When Marcus ends the game via the quizmaster interface
    Then the display screen shows "Trivia Night Complete"
    And "Team Awesome" is highlighted as the winner
    And all teams' final totals are shown in rank order

  @skip
  Scenario: Tied winners are shown together on the final display
    Given "Team Awesome" and "The Brainiacs" have the same total after all rounds
    When Marcus ends the game
    Then both "Team Awesome" and "The Brainiacs" are shown as joint winners on the display screen

  @skip
  Scenario: Player screens show final scores after game end
    Given Marcus has ended the game
    When Priya views her player screen
    Then Priya sees the final standings with "Team Awesome"'s total score
    And all answer fields on the player screen are permanently locked
