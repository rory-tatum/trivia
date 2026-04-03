Feature: Milestone 1 -- Game Session Management
  As Marcus the quizmaster and players joining the game
  I want reliable game session creation, lobby management, and game start
  So that all participants enter the game together from a shared starting point

  # All scenarios in this file are @skip until the walking skeleton passes.
  # Enable one at a time, implement, commit, repeat.
  #
  # Driving ports:
  #   host_load_quiz   -> /ws (host room)
  #   team_register    -> /ws (play room)
  #   team_rejoin      -> /ws (play room)
  #   host_start_round -> /ws (host room)

  Background:
    Given a quiz file "friday-march-2026.yaml" containing 4 rounds and 32 text questions exists on the server
    And the quizmaster panel is accessible with a valid host token

  # -------------------------------------------------------------------------
  # US-01: Load YAML Quiz File
  # -------------------------------------------------------------------------

  Scenario: Valid YAML file creates a game session
    When Marcus loads "friday-march-2026.yaml" via the quizmaster interface
    Then Marcus sees "Friday Night Trivia -- March 2026 | 4 rounds | 32 questions"
    And the panel shows a shareable player join URL
    And the panel shows a shareable display screen URL
    And the game session has a unique session identifier

  @skip
  Scenario: YAML file with a missing required field shows a specific error
    Given a quiz file "quiz-missing-answer.yaml" where round 2 question 3 has no answer field
    When Marcus loads "quiz-missing-answer.yaml" via the quizmaster interface
    Then Marcus sees the error "Round 2, Question 3: missing required field 'answer'"
    And no game session is created
    And Marcus can correct the file and reload without losing the quizmaster page

  @skip
  Scenario: YAML file referencing a missing media file shows a specific error
    Given a quiz file "quiz-missing-image.yaml" that references "eiffel.jpg" in round 1 question 4
    And "eiffel.jpg" is not present in the quiz file directory
    When Marcus loads "quiz-missing-image.yaml" via the quizmaster interface
    Then Marcus sees the error "Round 1, Question 4: image file 'eiffel.jpg' not found"
    And the load is rejected until the file is present

  @skip
  Scenario: Providing a path to a file that does not exist shows a clear error
    When Marcus provides the path "/quizzes/no-such-file.yaml" via the quizmaster interface
    Then Marcus sees "File not found: /quizzes/no-such-file.yaml"
    And Marcus can correct the path and retry without reloading the page

  # -------------------------------------------------------------------------
  # US-02: Game Lobby
  # -------------------------------------------------------------------------

  @skip
  Scenario: Connected teams appear in the lobby within 2 seconds
    Given Marcus has loaded "friday-march-2026.yaml" and the lobby is open
    When Priya connects to the player interface and joins as "Team Awesome"
    Then Marcus sees "Team Awesome" in the connected teams list within 2 seconds
    And the entry shows "Team Awesome" and a connection time

  @skip
  Scenario: Multiple devices for one team show a device count, not duplicate entries
    Given Marcus has loaded "friday-march-2026.yaml" and the lobby is open
    When Priya joins as "Team Awesome" from her phone
    And Jordan also connects from a second device and joins as "Team Awesome"
    Then the lobby shows "Team Awesome (2 devices)" as a single entry
    And Marcus sees exactly one entry for "Team Awesome"

  @skip
  Scenario: Start game is blocked when no teams have joined
    Given Marcus has loaded "friday-march-2026.yaml" and the lobby is open
    And no teams have connected yet
    Then the "Start Game" action is unavailable
    And Marcus sees "Waiting for teams to join..."

  # -------------------------------------------------------------------------
  # US-03: Start Game Broadcast
  # -------------------------------------------------------------------------

  Scenario: Starting the game transitions all connected players simultaneously
    Given "Team Awesome", "The Brainiacs", and "Quiz Killers" are connected in the lobby
    When Marcus starts the game via the quizmaster interface
    Then all three player connections receive the round started event within 1 second
    And each player screen shows round 1 is active
    And the display screen transitions from the waiting state to the question view

  Scenario: A player who joins after game start receives the current game state
    Given Marcus has started the game and revealed question 1 in round 1
    When a new player connects and joins as "Late Arrivals"
    Then "Late Arrivals" immediately sees round 1 as active
    And "Late Arrivals" sees question 1 already revealed in their answer form
    And "Late Arrivals" does not see a lobby screen

  @skip
  Scenario: A player who was offline during game start catches up on reconnection
    Given Priya was connected in the lobby but lost her connection before game start
    And Marcus started the game and revealed question 1 and question 2
    When Priya's player connection reconnects
    Then Priya's player screen shows round 1 active with both questions revealed

  # -------------------------------------------------------------------------
  # US-04: Player Joins (First Visit)
  # -------------------------------------------------------------------------

  Scenario: Player registers a new team name successfully
    Given Marcus has loaded "friday-march-2026.yaml" and the lobby is open
    When Priya connects to the player interface and joins as "Team Awesome"
    Then a persistence token is stored in Priya's browser for "Team Awesome"
    And Priya sees the lobby showing "Team Awesome" and other connected teams
    And Marcus sees "Team Awesome" in his connected teams list

  Scenario: Registering a team name that is already taken shows a clear error
    Given "Quiz Killers" is already registered in the lobby
    When another player tries to join as "Quiz Killers" from a different device
    Then that player sees "That name is taken -- try a different team name"
    And the name field remains populated so the player can edit it
    And no duplicate team entry appears in the lobby

  Scenario: Team name matching is case-insensitive for the uniqueness check
    Given "Team Awesome" is already registered in the lobby
    When a new player tries to join as "team awesome"
    Then that player sees "That name is taken -- try a different team name"

  # -------------------------------------------------------------------------
  # US-05: Auto-Rejoin on Refresh
  # -------------------------------------------------------------------------

  @skip
  Scenario: Player restores their session after accidental page refresh
    Given Priya has joined as "Team Awesome" and the game is in round 1
    And Priya has entered "Paris" for question 1 and "Mercury" for question 2
    When Priya's player page is refreshed
    Then Priya sees "Welcome back, Team Awesome!" on reload
    And Priya's answer for question 1 shows "Paris"
    And Priya's answer for question 2 shows "Mercury"

  @skip
  Scenario: Player with an expired or unknown session token starts fresh
    Given a player has a browser token that does not match any active session
    When that player connects to the player interface
    Then the player sees the join form to register a new team name
    And no error about the previous session is shown

  @skip
  Scenario: Rejoining player during scoring sees the locked state
    Given "Team Awesome" submitted their answers and the game entered scoring
    When Priya refreshes the player page and rejoins as "Team Awesome"
    Then Priya's player screen shows the submitted answers in read-only form
    And a banner reads "Answers are locked for round 1"
