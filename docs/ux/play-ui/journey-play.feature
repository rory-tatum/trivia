Feature: Player Interface — /play

  Players join a pub trivia game on their phones, answer questions as they are revealed
  by the quizmaster, submit answers at the end of each round, and watch the score ceremony.
  The interface must feel immediate, reliable, and social — like a digital paper answer sheet.

  # Personas
  # Priya Nair — team captain, enters answers on behalf of the team, anxious about losing work
  # Jordan Kim — casual member, rejoins mid-game, expects seamless recovery

  # JTBD grounding: J1 (join), J2 (rejoin), J3 (answer), J4 (submit), J5 (ceremony), J6 (orientation), J7 (types)

  Background:
    Given the trivia game server is running
    And the quizmaster has loaded a quiz with 2 rounds of 8 questions each

  # ───────────────────────────────────────────────────────────────────
  # S01 — JOIN (J1)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Player joins the game for the first time
    Given Priya opens /play on her Android phone
    And she has no stored team identity in localStorage
    When she enters "Team Awesome" and taps Join Game
    Then she sees the lobby screen with "Team Awesome — You're in!"
    And "Team Awesome" appears in the connected teams list

  Scenario: Team name is already taken
    Given "The Brainiacs" has already registered
    When Jordan enters "The Brainiacs" and taps Join Game
    Then an inline error reads "That name is taken — try another"
    And the input field retains focus for immediate retry
    And no navigation occurs

  Scenario: Empty team name is rejected before sending
    Given Priya is on the join screen
    When she taps Join Game without entering a team name
    Then no team_register message is sent
    And the input field displays a validation hint "Please enter a team name"

  # ───────────────────────────────────────────────────────────────────
  # S02 — AUTO-REJOIN (J2)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Player rejoins after phone goes to sleep
    Given Jordan closed his Safari tab during Round 1
    And his team_id "abc-123" and device_token "tok-789" are stored in localStorage
    When he reopens /play
    Then the screen shows "Welcome back, Team Awesome!"
    And a rejoining progress indicator appears
    And his draft answers for Round 1 are restored in the answer form

  Scenario: Rejoin token not recognised
    Given Priya's device_token is no longer valid
    When she opens /play and the server returns TEAM_NOT_FOUND
    Then the message reads "We couldn't find your team — please join again"
    And the join form appears ready for a fresh team name

  Scenario: Player rejoins while game is in ceremony state
    Given Jordan's device_token is valid
    And the game is currently showing the Round 1 ceremony
    When he reopens /play
    Then his screen shows the ceremony view for the current question
    And he does not see the lobby or answer form

  # ───────────────────────────────────────────────────────────────────
  # S03 — LOBBY (J1, J6)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Player sees other teams while waiting in lobby
    Given Team Awesome, The Brainiacs, and Quiz Killers have all joined
    When Priya's lobby screen renders
    Then she sees all three teams listed
    And the message reads "Waiting for Marcus to start the game..."

  Scenario: New team joins lobby in real time
    Given Priya is on the lobby screen with two other teams visible
    When "Quiz Killers" connects and registers
    Then "Quiz Killers" appears in the teams list without a page reload

  Scenario: Game starts while player is in lobby
    Given Priya is on the lobby screen
    When the quizmaster starts Round 1
    Then her screen transitions to the Round Active view automatically
    And Round 1, Q1 appears immediately

  # ───────────────────────────────────────────────────────────────────
  # S04 — FIRST QUESTION REVEALED (J3, J6)
  # ───────────────────────────────────────────────────────────────────

  Scenario: First question appears when the round starts
    Given Team Awesome is in the lobby
    When Marcus reveals Question 1 "What is the capital of France?"
    Then Priya's screen shows the question text prominently
    And an empty answer field appears below it
    And the header reads "Round 1 — 1 of 8 questions revealed"
    And the hint "Answers saved as you type" is visible

  Scenario: Player types an answer and it persists in localStorage
    Given Q1 "What is the capital of France?" is visible on Priya's screen
    When she types "Paris" into the Q1 answer field
    Then "Paris" is saved to localStorage draft for Round 1, Question 0 immediately
    And the draft_answer message is sent to the server on blur

  # ───────────────────────────────────────────────────────────────────
  # S05 — MULTIPLE QUESTIONS REVEALED (J3, J7)
  # ───────────────────────────────────────────────────────────────────

  Scenario: New question appears without clearing existing answers
    Given Priya has typed "Paris" for Q1
    When Marcus reveals Q2 "Name the three primary colors"
    Then "Paris" remains in the Q1 answer field
    And a new Q2 answer section appears below Q1
    And the header updates to "Round 1 — 2 of 8 questions revealed"

  Scenario: Player can edit a previous answer after a new question appears
    Given Q1 shows "Paris" and Q2 has just been revealed
    When Priya taps the Q1 answer field and changes it to "Rome"
    Then "Rome" replaces "Paris" in the Q1 field
    And the localStorage draft for Q1 updates to "Rome"

  Scenario: Multi-part answer question appears
    Given Q2 "Name the three primary colors" has a multi-part answer type
    When it is revealed on Priya's screen
    Then an expandable list appears with one text field and a "+ add part" button
    And each part is stored as a separate draft entry

  Scenario: Image question renders inline
    Given Q3 has an image attachment "eiffel.jpg"
    When it is revealed on Priya's screen
    Then the image renders above the question text "Name this landmark"
    And an answer text field appears below the question text

  Scenario: Image fails to load
    Given Q3 has an image attachment that the browser cannot load
    When the image block renders
    Then the image block shows "Media unavailable — ask the quizmaster"
    And the answer field remains available for text input

  # ───────────────────────────────────────────────────────────────────
  # S06 — MULTIPLE CHOICE QUESTION (J7)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Player selects a multiple choice answer
    Given Q4 "Which planet is closest to the sun?" is revealed with choices Venus, Mercury, Mars, Earth
    When Priya taps "Mercury"
    Then Mercury is visually selected with a filled radio indicator
    And the selection is saved as the draft answer for Q4

  Scenario: Player changes a multiple choice selection
    Given Priya has selected "Venus" for Q4
    When she taps "Mercury"
    Then "Mercury" is selected and "Venus" is deselected
    And the localStorage draft for Q4 updates to "Mercury"

  # ───────────────────────────────────────────────────────────────────
  # S07 — AUDIO/VIDEO QUESTION (J7)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Audio question renders with playback control
    Given Q6 has an audio attachment "clip.mp3"
    When it is revealed on Priya's screen
    Then an audio player with a Play button appears above the question text
    And the answer field is visible below
    And audio does not autoplay (explicit user action required)

  # ───────────────────────────────────────────────────────────────────
  # S08 — END-OF-ROUND REVIEW (J4)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Review screen shows all answers at end of round
    Given Priya has answered Q1–Q8 (leaving Q5 blank)
    When Marcus ends the round
    Then Priya's screen shows all 8 questions with their draft answers in order
    And Q5 shows "(no answer) ⚠"
    And a warning reads "Q5 has no answer. You can still go back."

  Scenario: Player goes back to edit before submitting
    Given Priya is on the review screen
    When she taps "Go Back & Edit"
    Then she returns to the answer form with all fields editable
    And her existing answers are preserved in each field

  Scenario: All answers blank warning appears but does not block submission
    Given Priya has answered none of the questions
    When Marcus ends the round and she reaches the review screen
    Then the warning reads "You haven't answered any questions yet."
    And the "Submit Answers" button is still available

  # ───────────────────────────────────────────────────────────────────
  # S09 — SUBMISSION CONFIRMATION (J4)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Player submits answers through confirmation dialog
    Given Priya is on the review screen with 1 blank answer (Q5)
    When she taps "Submit Answers"
    Then a confirmation dialog appears reading "You have 1 unanswered question (Q5)"
    And the dialog states "Once submitted, you cannot change your answers"
    When she taps "Yes, Submit"
    Then the submit_answers message is sent to the server
    And her screen transitions to "Round 1: Submitted!"
    And all answer fields become read-only

  Scenario: Player cancels submission from the dialog
    Given Priya is on the confirmation dialog
    When she taps "Go Back"
    Then the dialog dismisses
    And she returns to the review screen with all options available

  Scenario: Confirmation dialog has no warning when all questions are answered
    Given Priya has answered all 8 questions
    When she taps "Submit Answers"
    Then the confirmation dialog appears without any unanswered question warning
    And the primary action reads "Yes, Submit"

  # ───────────────────────────────────────────────────────────────────
  # S10 — POST-SUBMIT WAITING (J5, J6)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Player sees other teams' submission status in real time
    Given Team Awesome has submitted
    When The Brainiacs submit their answers
    Then "The Brainiacs" status changes from "[waiting...]" to "[submitted]"
    Without a page reload

  Scenario: Player cannot edit answers after submission
    Given Team Awesome has submitted and Priya taps an answer field
    Then the field does not become editable
    And a reminder reads "Answers are locked"

  # ───────────────────────────────────────────────────────────────────
  # S11 — CEREMONY VIEW (J5)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Ceremony answer is revealed on player's phone
    Given Team Awesome has submitted and the ceremony has started
    When Marcus reveals the answer to Q3 as "Eiffel Tower"
    Then Priya's screen shows "Answer: Eiffel Tower"
    And "Team Awesome" shows "✓ got it"
    And "Quiz Killers" shows "✗ missed it"

  Scenario: Ceremony question shown before answer is revealed
    Given Marcus has shown Q4 on the display screen
    When the ceremony_question_shown event arrives at Priya's phone
    Then her screen shows the Q4 question text
    And the answer area reads "Waiting for quizmaster to reveal..."

  # ───────────────────────────────────────────────────────────────────
  # S12 — ROUND SCORES (J5)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Round scores appear after ceremony
    Given the Round 1 ceremony is complete
    When Marcus publishes scores
    Then Priya's screen shows all teams ranked with round scores and running totals
    And Team Awesome is visually highlighted as "Your team"

  Scenario: Final scores appear at game over
    Given Round 2 is the last round and its ceremony is complete
    When Marcus ends the game
    Then Priya's screen shows the final leaderboard
    And the winning team is visually highlighted

  # ───────────────────────────────────────────────────────────────────
  # S_CONNECTION — CONNECTION RESILIENCE (J2, J6)
  # ───────────────────────────────────────────────────────────────────

  Scenario: Reconnecting banner appears on connection drop
    Given Priya is entering answers for Round 1
    When her WebSocket connection drops
    Then a "Reconnecting..." banner overlays the current screen
    And the answer form remains visible behind the banner

  Scenario: State is restored after successful reconnect
    Given Priya's connection dropped during Round 1
    When the WebSocket reconnects
    Then the reconnecting banner disappears
    And the server sends a state_snapshot
    And Priya's screen shows the correct Round 1 answer form with her draft answers intact

  @property
  Scenario: Draft answers survive any connection interruption
    Given Priya's draft answers are stored in localStorage
    When the WebSocket connection is lost and subsequently restored
    Then all draft answers are present in the answer form
    And no draft answer is lost

  @property
  Scenario: Answer form does not lose data when new questions are revealed
    Given the answer form contains draft answers for Q1 and Q2
    When Q3 is revealed via a question_revealed event
    Then drafts for Q1 and Q2 remain unchanged
    And Q3 appears with an empty answer field
