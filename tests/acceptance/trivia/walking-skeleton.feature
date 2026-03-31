Feature: Walking Skeleton -- One Complete Trivia Round
  As Marcus the quizmaster
  I want to run one complete round of trivia from YAML load to scores on screen
  So that I can validate the end-to-end game loop works before game night

  # THE WALKING SKELETON -- no @skip tag, this is the FIRST scenario to implement.
  #
  # This scenario validates the minimum observable user outcome:
  #   Can Marcus run one complete round of trivia on this system?
  #
  # Driving ports exercised (in order):
  #   1. host_load_quiz WebSocket event  -> game session created
  #   2. team_register WebSocket event   -> team joins
  #   3. host_start_round WebSocket event -> game starts
  #   4. host_reveal_question             -> question appears for players
  #   5. draft_answer                     -> player drafts an answer
  #   6. submit_answers                   -> player locks in answers
  #   7. host_mark_answer                 -> quizmaster scores
  #   8. host_ceremony_show_question      -> ceremony begins
  #   9. host_ceremony_reveal_answer      -> answer revealed on display
  #  10. host_publish_scores              -> scores visible to all
  #  11. host_end_game                    -> game concludes

  @walking_skeleton
  Scenario: Marcus runs one complete round of trivia from YAML load to final scores
    Given Marcus has a valid quiz file "friday-march-2026.yaml" with 1 round and 2 text questions
    And Marcus opens the quizmaster panel with a valid host token
    When Marcus loads the quiz file via the quizmaster interface
    Then Marcus sees "Friday Night Trivia -- March 2026 | 1 round | 2 questions"
    And the panel shows a join URL that players can use to connect

    When Priya connects to the player interface and joins as "Team Awesome"
    Then Marcus sees "Team Awesome" appear in the connected teams list

    When Marcus starts the game
    Then the game enters the first round

    When Marcus reveals question 1
    Then "Team Awesome" sees "What is the capital of France?" on their player screen

    When Priya enters "Paris" as her answer to question 1
    And Marcus reveals question 2
    Then "Team Awesome" sees both revealed questions on their player screen

    When Priya enters "Blue" as her answer to question 2
    And Marcus ends the round
    And Priya submits "Team Awesome"'s answers
    Then Priya sees confirmation that "Team Awesome"'s answers are locked in

    When Marcus opens scoring
    Then Marcus sees "Team Awesome"'s submitted answers

    When Marcus marks "Paris" as correct for question 1
    And Marcus marks "Blue" as correct for question 2
    And Marcus starts the answer ceremony
    Then the display screen shows question 1 of the ceremony

    When Marcus reveals the answer to question 1 during ceremony
    Then the display screen shows the answer "Paris" for question 1

    When Marcus steps through all ceremony questions and publishes the round scores
    Then the display screen shows "Team Awesome" with 2 points
    And the player screen for "Team Awesome" shows their final round score

    When Marcus ends the game
    Then the display screen shows the final winner as "Team Awesome"
    And all answer fields on the player screen are permanently locked
