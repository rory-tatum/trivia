# Acceptance Criteria — Host UI (Quizmaster Panel)

**Feature ID**: host-ui
**Date**: 2026-04-09

All acceptance criteria are expressed as testable Given/When/Then statements. Each item is traceable to a user story.

---

## US-01: Reliable Connection Status and Auth Error Handling

### AC-01-1: Connecting state before handshake
```gherkin
Given Marcus navigates to "/?token=pub-night-secret"
When the page loads but onOpen has not yet fired
Then the status indicator text is "Connecting..."
And no game controls are visible
```

### AC-01-2: Connected state after onOpen
```gherkin
Given Marcus navigates to "/?token=pub-night-secret"
When the WebSocket handshake completes and onOpen fires
Then the status indicator text is "Connected"
And the Load Quiz form is visible
```

### AC-01-3: Auth failure — permanent error, no retry
```gherkin
Given Marcus navigates to "/?token=bad-token"
When the WebSocket upgrade is rejected (HTTP 403)
And the connection closes with code 1006 before any onOpen has fired
Then the status indicator text is "Disconnected"
And the message "Connection refused — invalid token. Check HOST_TOKEN and reload." is displayed
And no further reconnect attempts occur
And no game controls are visible
```

### AC-01-4: Mid-game reconnect preserves state
```gherkin
Given Marcus is connected and the round active panel is showing Round 1
When the WebSocket connection drops (onClose fires after a prior onOpen)
Then the status indicator text is "Reconnecting..."
And the round active panel remains visible with current state
When onOpen fires again on successful reconnect
Then the status indicator text is "Connected"
And game controls are interactive
```

### AC-01-5: Reconnect exhausted — reload prompt
```gherkin
Given the WebSocket has failed to reconnect 10 consecutive times
When the reconnect_failed client event fires
Then the message "Could not reconnect. Please reload." is displayed
And a "Reload" button is visible
And clicking "Reload" triggers a page reload
```

---

## US-02: Load Quiz with Confirmation

### AC-02-1: Load Quiz form visible on connect
```gherkin
Given Marcus is connected (status "Connected")
Then a text input with label "Quiz file path" is visible
And a "Load Quiz" button is visible
And no round controls are visible
```

### AC-02-2: Successful quiz load
```gherkin
Given Marcus is connected
When Marcus enters "/quizzes/pub-night-vol3.yaml" in the file path input
And Marcus clicks "Load Quiz"
And the server responds with quiz_loaded event (title "Pub Night Vol. 3", 3 rounds, 15 questions)
Then the confirmation text "Pub Night Vol. 3 | 3 rounds | 15 questions" is displayed
And the player URL is displayed
And the display URL is displayed
And a "Start Round 1: General Knowledge" button is visible
And the Load Quiz form is no longer shown
```

### AC-02-3: Quiz load failure inline error
```gherkin
Given Marcus is connected
When Marcus enters "/nonexistent/file.yaml" in the file path input
And Marcus clicks "Load Quiz"
And the server responds with an error event
Then an error message is displayed below the file path input
And the file path input remains editable and focused
And no round controls appear
```

### AC-02-4: Empty path client-side validation
```gherkin
Given Marcus is on the Load Quiz form with an empty file path input
When Marcus clicks "Load Quiz"
Then no host_load_quiz message is sent over WebSocket
And the validation message "Please enter a quiz file path." is displayed
```

---

## US-03: Run a Round — Start, Reveal Questions, End

### AC-03-1: Start Round button after quiz load
```gherkin
Given the quiz "Pub Night Vol. 3" is loaded (3 rounds)
Then the button "Start Round 1: General Knowledge" is visible
And "Start Round 2" and "Start Round 3" buttons are not visible
```

### AC-03-2: Starting a round
```gherkin
Given the "Start Round 1: General Knowledge" button is visible
When Marcus clicks it
Then a host_start_round WebSocket message is sent with round_index 0
And the round panel heading reads "Round 1: General Knowledge"
And the counter reads "0 of 5 revealed"
And the "Reveal Next Question" button is visible
And the "End Round" button is not visible
```

### AC-03-3: Revealing a question
```gherkin
Given Marcus is in Round 1 and 0 of 5 questions are revealed
When Marcus clicks "Reveal Next Question"
Then a host_reveal_question message is sent with round_index 0 and question_index 0
And the question text from the question_revealed event appears in the revealed list
And the counter reads "1 of 5 revealed"
```

### AC-03-4: Sequential question reveal
```gherkin
Given Marcus has revealed 2 of 5 questions
When Marcus clicks "Reveal Next Question" again
Then a host_reveal_question message is sent with question_index 2
And the counter reads "3 of 5 revealed"
```

### AC-03-5: All questions revealed — End Round appears
```gherkin
Given Marcus has revealed 4 of 5 questions
When Marcus clicks "Reveal Next Question" for the 5th time
Then the counter reads "5 of 5 revealed"
And the "Reveal Next Question" button is no longer visible
And the "End Round" button is visible
```

### AC-03-6: End Round sends both commands
```gherkin
Given all 5 questions are revealed in Round 1
When Marcus clicks "End Round"
Then a host_end_round message is sent with round_index 0
And a host_begin_scoring message is sent with round_index 0
And the scoring panel becomes visible
And the round active panel is replaced
```

---

## US-04: Score a Round — Mark Answers and Publish Scores

### AC-04-1: Scoring panel shows questions with correct answers
```gherkin
Given Marcus has clicked "End Round" and begin_scoring has been sent
Then the scoring panel is visible
And Q1 shows question text "What is the capital of France?"
And Q1 shows correct answer "Paris"
```

### AC-04-2: Scoring panel shows team submissions
```gherkin
Given the scoring panel is visible for Round 1
Then team "The Brainiacs" is listed under Q1 with their answer "Paris"
And team "Quiz Killers" is listed under Q1 with their answer "Lyon"
And each team row has a "Correct" button and a "Wrong" button
```

### AC-04-3: Correct verdict
```gherkin
Given Marcus is on the scoring panel for Q1
When Marcus clicks "Correct" for "The Brainiacs"
Then a host_mark_answer message is sent with team_id "brainiacs-001", round_index 0, question_index 0, verdict "correct"
And the "Correct" button for The Brainiacs on Q1 is visually marked as applied
And the running total for "The Brainiacs" increases to reflect the new score
```

### AC-04-4: Wrong verdict
```gherkin
Given Marcus is on the scoring panel for Q1
When Marcus clicks "Wrong" for "Quiz Killers"
Then a host_mark_answer message is sent with verdict "wrong"
And the "Wrong" button for Quiz Killers on Q1 is visually marked as applied
And the running total for "Quiz Killers" is unchanged
```

### AC-04-5: Publish Scores
```gherkin
Given Marcus has applied verdicts for all answers in Round 1
When Marcus clicks "Publish Scores"
Then a host_publish_scores message is sent with round_index 0
And the scores_published event is received
And controls for the next phase (ceremony or next round) become available
```

---

## US-05: Run Answer Ceremony on Display

### AC-05-1: Ceremony panel after publish
```gherkin
Given Marcus has clicked "Publish Scores" for Round 1
Then the ceremony panel is visible
And the ceremony progress reads "Question 0 of 5 shown"
And the "Show Next Question" button is visible
And the "Reveal Answer" button is not visible
```

### AC-05-2: Show question on display and play
```gherkin
Given Marcus is on the ceremony panel
When Marcus clicks "Show Next Question"
Then a host_ceremony_show_question message is sent with question_index 0
And the display screen renders "Q1: What is the capital of France?"
And the play screen renders "Q1: What is the capital of France?"
And the ceremony progress reads "Question 1 of 5 shown"
And the "Reveal Answer" button becomes visible
```

### AC-05-3: Reveal answer to display only
```gherkin
Given Question 1 is shown on display and play
When Marcus clicks "Reveal Answer"
Then a host_ceremony_reveal_answer message is sent with question_index 0
And the display screen renders "Answer: Paris"
And the play screen does NOT render the answer text
```

### AC-05-4: Ceremony complete
```gherkin
Given Marcus has shown and revealed all 5 questions in Round 1
Then the ceremony progress reads "Question 5 of 5 shown"
And the "Show Next Question" button is not visible
And the message "Ceremony complete" is displayed
And "Start Round 2" or "End Game" buttons are visible
```

---

## US-06: End Game and View Final Leaderboard

### AC-06-1: End Game button available
```gherkin
Given at least one round has been scored and published
Then the "End Game" button is visible
```

### AC-06-2: End game triggers game_over
```gherkin
Given the "End Game" button is visible
When Marcus clicks "End Game"
Then a host_end_game message is sent with empty payload
And the game_over event is received with final scores
```

### AC-06-3: Leaderboard sorted descending
```gherkin
Given the game_over event has been received
And final scores are: The Brainiacs 12pts, Quiz Killers 9pts, Random Facts 7pts
When the leaderboard renders
Then "1. The Brainiacs — 12 pts" appears first
And "2. Quiz Killers — 9 pts" appears second
And "3. Random Facts — 7 pts" appears third
```

### AC-06-4: Game controls removed after game over
```gherkin
Given the game_over leaderboard is displayed
Then the "Start Round", "Reveal Next Question", "End Round", "Publish Scores", and "End Game" buttons are not visible
```
