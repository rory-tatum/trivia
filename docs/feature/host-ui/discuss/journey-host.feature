Feature: Host UI — Quizmaster Panel
  As Marcus, the quizmaster
  I want a browser-based control panel connected via WebSocket
  So that I can run a trivia game session for my friends without confusion

  Background:
    Given the trivia server is running with HOST_TOKEN set to "pub-night-secret"
    And Marcus navigates to "/?token=pub-night-secret"

  # ─── Step 1 & 2: Connection lifecycle ───────────────────────────────────────

  Scenario: Quizmaster connects successfully and sees the load quiz form
    Given Marcus opens "/?token=pub-night-secret"
    When the WebSocket handshake completes (onOpen fires)
    Then the status indicator shows "Connected"
    And the "Load Quiz" form is visible with a file path input and a "Load Quiz" button
    And the "Start Round" button is not visible

  Scenario: Wrong token shows a permanent auth error — no retry loop
    Given Marcus opens "/?token=wrong-token"
    When the WebSocket upgrade is rejected with HTTP 403
    And the connection closes with CloseEvent code 1006 on the first attempt
    Then the status indicator shows "Disconnected"
    And the page displays "Connection refused — invalid token. Check HOST_TOKEN and reload."
    And no reconnect attempts are made

  Scenario: Status correctly reflects connecting state before handshake
    Given Marcus opens "/?token=pub-night-secret"
    When the page loads but before onOpen fires
    Then the status indicator shows "Connecting..." (not "Connected")

  # ─── Step 3: Load quiz ────────────────────────────────────────────────────

  Scenario: Quizmaster loads a quiz successfully
    Given Marcus is connected and on the idle screen
    When Marcus enters "/quizzes/pub-night-vol3.yaml" in the file path input
    And Marcus clicks "Load Quiz"
    Then the server receives a host_load_quiz message with file_path "/quizzes/pub-night-vol3.yaml"
    And the confirmation "Pub Night Vol. 3 | 3 rounds | 15 questions" is displayed
    And the "Start Round 1" button becomes visible
    And the player join URL "http://host/play" is displayed
    And the display URL "http://host/display" is displayed

  Scenario: Quiz load fails with invalid file path
    Given Marcus is connected and on the idle screen
    When Marcus enters "/nonexistent/path.yaml" in the file path input
    And Marcus clicks "Load Quiz"
    Then the server returns an error event with code "quiz_load_failed"
    And an error message is shown below the file path input
    And the file path input remains editable
    And the "Start Round" button remains hidden

  # ─── Step 4: Start round ─────────────────────────────────────────────────

  Scenario: Quizmaster starts the first round
    Given Marcus has loaded "Pub Night Vol. 3" (3 rounds, 5 questions each)
    When Marcus clicks "Start Round 1"
    Then the server receives a host_start_round message with round_index 0
    And the round panel shows "Round 1: General Knowledge"
    And the question counter shows "0 of 5 revealed"
    And the "Reveal Next Question" button is visible

  # ─── Step 5: Reveal questions ─────────────────────────────────────────────

  Scenario: Quizmaster reveals questions one by one
    Given Marcus has started Round 1 with 5 questions
    When Marcus clicks "Reveal Next Question" for the first time
    Then the server receives host_reveal_question with round_index 0 and question_index 0
    And "Q1: What is the capital of France?" appears in the revealed questions list
    And the question counter shows "1 of 5 revealed"

  Scenario: All questions revealed — End Round button appears
    Given Marcus has started Round 1 with 5 questions
    When Marcus reveals all 5 questions
    Then the question counter shows "5 of 5 revealed"
    And the "Reveal Next Question" button is replaced by the "End Round" button

  # ─── Step 6: End round and begin scoring ──────────────────────────────────

  Scenario: Quizmaster ends the round and enters scoring
    Given Marcus has revealed all 5 questions in Round 1
    When Marcus clicks "End Round"
    Then the server receives host_end_round with round_index 0
    And the server receives host_begin_scoring
    And the scoring panel is displayed
    And each question shows its correct answer
    And each question shows each team's submitted answer with "Correct" and "Wrong" buttons

  # ─── Step 7: Mark answers ────────────────────────────────────────────────

  Scenario: Quizmaster marks a team answer as correct
    Given Marcus is in the scoring panel for Round 1, Question 1
    And team "The Brainiacs" answered "Paris" for "What is the capital of France?"
    When Marcus clicks "Correct" for The Brainiacs on Question 1
    Then the server receives host_mark_answer with team_id "brainiacs-001", round_index 0, question_index 0, verdict "correct"
    And the running total for "The Brainiacs" increases by the question's point value
    And the "Correct" button is highlighted to indicate the verdict is applied

  Scenario: Quizmaster marks a team answer as wrong
    Given Marcus is in the scoring panel for Round 1, Question 1
    And team "Quiz Killers" answered "Lyon" for "What is the capital of France?"
    When Marcus clicks "Wrong" for Quiz Killers on Question 1
    Then the server receives host_mark_answer with team_id "killers-002", round_index 0, question_index 0, verdict "wrong"
    And the running total for "Quiz Killers" is unchanged

  # ─── Step 8: Publish scores ────────────────────────────────────────────────

  Scenario: Quizmaster publishes round scores
    Given Marcus has marked all answers for Round 1
    When Marcus clicks "Publish Scores"
    Then the server receives host_publish_scores with round_index 0
    And the round score summary is shown to Marcus
    And controls to start the next round (Round 2) or run ceremony become available

  # ─── Step 9: Ceremony ─────────────────────────────────────────────────────

  Scenario: Quizmaster shows a question on the display screen during ceremony
    Given Marcus has published scores for Round 1
    And Marcus has chosen to run the answer ceremony
    When Marcus clicks "Show Next Question" for the first time
    Then the server receives host_ceremony_show_question with question_index 0
    And the display screen shows "Q1: What is the capital of France?"
    And the ceremony progress shows "Question 1 of 5 shown"

  Scenario: Quizmaster reveals the answer for a ceremony question
    Given the ceremony is showing Question 1 ("What is the capital of France?")
    When Marcus clicks "Reveal Answer"
    Then the server receives host_ceremony_reveal_answer with question_index 0
    And the display screen shows the answer "Paris"
    And the play screen does NOT show the answer

  # ─── Step 10: End game ────────────────────────────────────────────────────

  Scenario: Quizmaster ends the game and sees the final leaderboard
    Given all 3 rounds have been played and scored
    When Marcus clicks "End Game"
    Then the server receives host_end_game
    And the final leaderboard is displayed showing all teams sorted by score descending
    And "The Brainiacs" with 12 points appears first
    And game control buttons are removed

  # ─── Error paths ──────────────────────────────────────────────────────────

  Scenario: Connection drops mid-game and reconnects automatically
    Given Marcus is in Round 2 with the round panel visible
    When the WebSocket connection drops unexpectedly (network hiccup)
    Then the status indicator shows "Reconnecting..."
    And the round panel remains visible (client-side state preserved)
    When the WebSocket reconnects successfully
    Then the status indicator returns to "Connected"
    And game controls are usable again

  Scenario: Reconnect attempts exhausted — reload prompt shown
    Given the WebSocket has dropped and reconnected unsuccessfully 10 times
    When the reconnect_failed event fires
    Then the status indicator shows "Disconnected"
    And the message "Could not reconnect. Please reload." is displayed
    And a "Reload" button is visible
