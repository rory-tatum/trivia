<!-- markdownlint-disable MD024 -->
# User Stories — Host UI (Quizmaster Panel)

**Feature ID**: host-ui
**Date**: 2026-04-09

---

## US-01: Reliable Connection Status and Auth Error Handling

### Problem
Marcus is a quizmaster who navigates to the host page with his token. He finds it alarming to see "Connected" immediately on page load when the WebSocket handshake has not completed — and infuriating to see "Connected" after entering a wrong token while the page silently retries in a loop with no explanation.

### Who
- Quizmaster | Opening the host page for the first time | Needs to trust the status indicator before doing anything else

### Solution
Show accurate connection lifecycle states: "Connecting..." before handshake, "Connected" only after onOpen fires, and a clear permanent error on auth failure (wrong token) with no retry loop.

### Domain Examples

#### 1: Happy path — correct token
Marcus opens `/?token=pub-night-secret`. The status shows "Connecting..." for ~200ms while the handshake completes. Once onOpen fires, it changes to "Connected" and the Load Quiz form appears.

#### 2: Wrong token
Marcus accidentally opens `/?token=old-token`. The WS upgrade returns 403. The page detects CloseEvent code 1006 on the first attempt (no prior onOpen). It shows "Connection refused — invalid token. Check HOST_TOKEN and reload." and stops retrying.

#### 3: Mid-game network drop
Marcus is in Round 2 when his Wi-Fi hiccups. The status shows "Reconnecting..." while the client retries. The round panel stays visible. When the connection restores, status returns to "Connected" and controls are usable again.

### UAT Scenarios (BDD)

#### Scenario: Status shows Connecting before handshake
```gherkin
Given Marcus navigates to "/?token=pub-night-secret"
When the page loads but before onOpen fires
Then the status indicator shows "Connecting..."
```

#### Scenario: Status shows Connected only after onOpen
```gherkin
Given Marcus navigates to "/?token=pub-night-secret"
When the WebSocket handshake completes and onOpen fires
Then the status indicator shows "Connected"
And the Load Quiz form is visible
```

#### Scenario: Wrong token shows permanent error, no retry
```gherkin
Given Marcus navigates to "/?token=wrong-token"
When the WebSocket upgrade is rejected with HTTP 403
And CloseEvent fires with code 1006 on the first attempt without a prior onOpen
Then the status indicator shows "Disconnected"
And the message "Connection refused — invalid token. Check HOST_TOKEN and reload." is visible
And no further reconnect attempts are made
```

#### Scenario: Mid-game drop shows Reconnecting with state preserved
```gherkin
Given Marcus is connected and the round panel is showing Round 2
When the WebSocket connection drops unexpectedly
Then the status indicator shows "Reconnecting..."
And the round panel remains visible
When the connection is restored and onOpen fires again
Then the status indicator shows "Connected"
And game controls are interactive
```

#### Scenario: Reconnect exhausted shows reload prompt
```gherkin
Given the WebSocket has failed to reconnect 10 times
When the reconnect_failed event fires
Then the message "Could not reconnect. Please reload." is displayed
And a "Reload" button is visible
```

### Acceptance Criteria
- [ ] Status shows "Connecting..." between `connect()` call and `onOpen` firing
- [ ] Status shows "Connected" only inside the `onOpen` callback
- [ ] Wrong token (first-attempt close code 1006 with no prior onOpen) shows permanent error message and stops retrying
- [ ] Mid-game disconnect (close after prior onOpen) shows "Reconnecting..." and preserves client-side state
- [ ] After 10 failed reconnects, shows reload prompt with button

### Outcome KPIs
- **Who**: Quizmasters opening the host page
- **Does what**: Correctly interpret the connection status without confusion
- **By how much**: 0 instances of "Connected" showing while actually disconnected
- **Measured by**: Manual QA + automated test of onOpen timing
- **Baseline**: Currently shows "Connected" immediately on every page load regardless of WS state

### Technical Notes
- Requires `onOpen(handler: () => void): void` added to WsClient (IC-1)
- Requires CloseEvent.code inspection in WsClient.onclose (IC-2): code 1006 on first attempt (attempt === 0, closed === false, onOpen never fired) → emit `connection_refused` event, stop retrying
- Auth failure detection: `connection_refused` event in Host.tsx sets `authError = true`, renders error branch
- Do not use HTTP status codes or protocol terms in user-facing messages

---

## US-02: Load Quiz with Confirmation

### Problem
Marcus is a quizmaster who connects successfully to the host page. He finds it bewildering that there is nothing to do after connecting — no form, no indication of what to do next, no way to start the game.

### Who
- Quizmaster | Just connected, ready to prepare the game | Needs to get the quiz into the system before friends arrive

### Solution
Show a "Load Quiz" form with a file path input and button. On success, show a confirmation string and the URLs for players and the display screen. On failure, show an inline error.

### Domain Examples

#### 1: Happy path — valid quiz file
Marcus enters `/home/marcus/quizzes/pub-night-vol3.yaml` and clicks "Load Quiz". The server responds with `quiz_loaded`. The page shows: "Pub Night Vol. 3 | 3 rounds | 15 questions" and the URLs `http://host/play` and `http://host/display`. The "Start Round 1" button appears.

#### 2: File not found
Marcus misremembers the path and enters `/quizzes/pub-night.yaml`. The server returns an error event. An inline message appears: "Could not load quiz: file not found." The input remains editable and the button is available for retry.

#### 3: Empty path submitted
Marcus accidentally clicks "Load Quiz" with an empty input. The command is not sent. An inline validation message appears: "Please enter a quiz file path."

### UAT Scenarios (BDD)

#### Scenario: Load quiz form visible when connected
```gherkin
Given Marcus is connected (status shows "Connected")
Then a text input labeled "Quiz file path" is visible
And a "Load Quiz" button is visible
And no "Start Round" button is visible
```

#### Scenario: Successful quiz load shows confirmation and controls
```gherkin
Given Marcus is connected
When Marcus enters "/quizzes/pub-night-vol3.yaml" and clicks "Load Quiz"
Then the host receives a quiz_loaded event with title "Pub Night Vol. 3"
And the confirmation "Pub Night Vol. 3 | 3 rounds | 15 questions" is displayed
And the player URL "http://host/play" is displayed
And the display URL "http://host/display" is displayed
And the "Start Round 1: General Knowledge" button is visible
```

#### Scenario: Quiz load failure shows inline error
```gherkin
Given Marcus is connected
When Marcus enters "/nonexistent/path.yaml" and clicks "Load Quiz"
Then the server sends an error event with code "quiz_load_failed"
And an error message is shown below the file path input
And the input remains editable
And no "Start Round" button appears
```

#### Scenario: Empty path prevented client-side
```gherkin
Given Marcus is on the load quiz form
When Marcus clicks "Load Quiz" with an empty file path input
Then no host_load_quiz message is sent
And the validation message "Please enter a quiz file path." appears below the input
```

### Acceptance Criteria
- [ ] Load Quiz form (input + button) visible immediately after connecting
- [ ] Successful load shows confirmation string in format `{title} | {N} rounds | {M} questions`
- [ ] Successful load shows `player_url` and `display_url`
- [ ] Successful load shows "Start Round N" button for the first round
- [ ] Failed load shows inline error; form remains editable
- [ ] Empty path submission prevented client-side with validation message

### Outcome KPIs
- **Who**: Quizmasters after connecting
- **Does what**: Successfully load a quiz on first attempt
- **By how much**: Load success rate >= 95% of attempts (failures = wrong path, not UX confusion)
- **Measured by**: Browser DevTools network / WS frame inspection in QA session
- **Baseline**: Currently impossible — no load form exists

### Technical Notes
- Send: `{ event: "host_load_quiz", payload: { file_path: string } }`
- Receive success: `{ event: "quiz_loaded", payload: QuizLoadedMeta }`
- Receive failure: `{ event: "error", payload: { message: string } }`
- `QuizLoadedMeta.confirmation` is pre-formatted by server; use it verbatim
- Display `playerURL` and `displayURL` as copyable text (or links) — these are what Marcus shares with friends

---

## US-03: Run a Round — Start, Reveal Questions, End

### Problem
Marcus is a quizmaster who has loaded a quiz. He finds it impossible to start the game — there are no round controls, no way to reveal questions to players, no progression mechanism.

### Who
- Quizmaster | Quiz is loaded, friends are in the `/play` room | Needs to start the round and pace question reveals

### Solution
Show a "Start Round N" button. Once clicked, show a "Reveal Next Question" button with a counter. Each click reveals the next question. When all questions are revealed, show "End Round". At all times, show the list of revealed question texts.

### Domain Examples

#### 1: Happy path — single round, 5 questions
Marcus clicks "Start Round 1". The panel shows "Round 1: General Knowledge — 0 of 5 revealed". Marcus clicks "Reveal Next Question" five times. After each click, a new question appears in the list. After the fifth, the button changes to "End Round".

#### 2: Partial reveal — Marcus pauses after Q3
Marcus reveals Q1, Q2, Q3. He pauses to let teams think. The counter shows "3 of 5 revealed". The "Reveal Next Question" button is still available. End Round is not yet shown.

#### 3: Multi-round game — advancing to round 2
After scoring Round 1 and publishing scores, the "Start Round 2: History" button appears. Marcus clicks it and the flow repeats with a fresh question counter.

### UAT Scenarios (BDD)

#### Scenario: Start Round button visible after quiz load
```gherkin
Given Marcus has successfully loaded "Pub Night Vol. 3" (3 rounds)
Then the "Start Round 1: General Knowledge" button is visible
```

#### Scenario: Starting a round sends correct message and updates UI
```gherkin
Given Marcus has loaded the quiz and sees "Start Round 1: General Knowledge"
When Marcus clicks "Start Round 1: General Knowledge"
Then a host_start_round message is sent with round_index 0
And the round panel shows "Round 1: General Knowledge"
And the question counter shows "0 of 5 revealed"
And the "Reveal Next Question" button is visible
```

#### Scenario: Revealing a question appends it to the list
```gherkin
Given Marcus has started Round 1 and 0 of 5 questions are revealed
When Marcus clicks "Reveal Next Question"
Then a host_reveal_question message is sent with round_index 0 and question_index 0
And "Q1: What is the capital of France?" appears in the revealed list
And the counter shows "1 of 5 revealed"
```

#### Scenario: All questions revealed — End Round replaces Reveal button
```gherkin
Given Marcus has revealed 4 of 5 questions in Round 1
When Marcus clicks "Reveal Next Question" for the 5th time
Then the question counter shows "5 of 5 revealed"
And the "Reveal Next Question" button is no longer visible
And the "End Round" button is visible
```

#### Scenario: Ending the round sends end_round and begin_scoring commands
```gherkin
Given all 5 questions in Round 1 are revealed
When Marcus clicks "End Round"
Then a host_end_round message is sent with round_index 0
And a host_begin_scoring message is sent
And the scoring panel becomes visible
```

### Acceptance Criteria
- [ ] "Start Round N: {name}" button visible after quiz load for the next unstarted round
- [ ] Clicking Start Round sends `host_start_round` with correct `round_index`
- [ ] Round panel shows round name and "N of M revealed" counter
- [ ] Each "Reveal Next Question" sends `host_reveal_question` with incrementing `question_index`
- [ ] Revealed question texts are displayed in an ordered list
- [ ] "Reveal Next Question" replaced by "End Round" only when all questions are revealed
- [ ] "End Round" sends `host_end_round` then `host_begin_scoring` in sequence

### Outcome KPIs
- **Who**: Quizmasters during an active round
- **Does what**: Reveal all questions in a round without confusion or wrong clicks
- **By how much**: Zero accidental double-reveals or out-of-order reveals
- **Measured by**: QA session walkthrough; all questions revealed in sequential order
- **Baseline**: Currently impossible — no round controls exist

### Technical Notes
- `question_index` is 0-based and increments with each reveal; track in local state
- `host_end_round` payload: `{ round_index: number }`
- `host_begin_scoring` payload: `{ round_index: number }`
- Round name comes from `question_revealed` payload or from `round_started` broadcast (check `RoundStartedMsg` payload in messages.ts — includes `round_name`)

---

## US-04: Score a Round — Mark Answers and Publish Scores

### Problem
Marcus is a quizmaster who has ended a round. He finds it frustrating that there is no scoring interface — he has no way to see what teams answered, mark answers correct or wrong, or publish scores for everyone to see.

### Who
- Quizmaster | Round ended, teams have submitted answers | Needs to judge each answer and record scores

### Solution
Show a scoring panel with each question's correct answer and each team's submitted answer. Each row has "Correct" and "Wrong" buttons. Running totals update as verdicts are applied. A "Publish Scores" button sends the final scores to all rooms.

### Domain Examples

#### 1: Happy path — two teams, one question
For "What is the capital of France?", The Brainiacs answered "Paris" and Quiz Killers answered "Lyon". Marcus marks The Brainiacs correct (score increases) and Quiz Killers wrong (score unchanged). He clicks "Publish Scores".

#### 2: Generous marking — close enough
Quiz Killers answered "paris" (lowercase). Marcus decides this is close enough and marks it correct. The running total for Quiz Killers increases.

#### 3: Multi-question scoring flow
Marcus works through Q1–Q5 for Round 1, marking each team's answer. Running totals update after each verdict. When all are marked, he clicks "Publish Scores".

### UAT Scenarios (BDD)

#### Scenario: Scoring panel shows correct answers and team submissions
```gherkin
Given Marcus has clicked "End Round" for Round 1
And host_begin_scoring has been sent
Then the scoring panel is visible
And Q1 "What is the capital of France?" shows correct answer "Paris"
And team "The Brainiacs" answer "Paris" is listed under Q1
And team "Quiz Killers" answer "Lyon" is listed under Q1
And each team row has a "Correct" button and a "Wrong" button
```

#### Scenario: Marking an answer correct updates running total
```gherkin
Given Marcus is on the scoring panel, Round 1
And "The Brainiacs" has a running total of 0
When Marcus clicks "Correct" for The Brainiacs on Q1
Then a host_mark_answer message is sent with team_id "brainiacs-001", round_index 0, question_index 0, verdict "correct"
And the running total for "The Brainiacs" increases
And the "Correct" button is visually marked as applied
```

#### Scenario: Marking an answer wrong does not change running total
```gherkin
Given Marcus is on the scoring panel, Round 1
And "Quiz Killers" has a running total of 0
When Marcus clicks "Wrong" for Quiz Killers on Q1
Then a host_mark_answer message is sent with verdict "wrong"
And the running total for "Quiz Killers" is unchanged
And the "Wrong" button is visually marked as applied
```

#### Scenario: Publish scores sends scores to all rooms
```gherkin
Given Marcus has marked all answers for Round 1
When Marcus clicks "Publish Scores"
Then a host_publish_scores message is sent with round_index 0
And controls for "Start Round 2" and "Run Ceremony" become available
```

### Acceptance Criteria
- [ ] Scoring panel shows each question with correct answer
- [ ] Each question lists each team's submitted answer
- [ ] Each team-answer row has "Correct" and "Wrong" buttons
- [ ] Clicking a verdict button sends `host_mark_answer` with team_id, round_index, question_index, verdict
- [ ] Running total per team updates visually after each verdict
- [ ] Applied verdict button is visually distinguished (highlighted/disabled)
- [ ] "Publish Scores" button visible after scoring begins; sends `host_publish_scores`
- [ ] After publish, controls offer starting next round or running ceremony

### Outcome KPIs
- **Who**: Quizmasters during scoring phase
- **Does what**: Mark all team answers without re-marking or missing any
- **By how much**: Zero missed verdicts per round (all answers marked before publish)
- **Measured by**: Scoring panel completion — publish button only enabled when all answers have a verdict (or Marcus explicitly overrides)
- **Baseline**: Currently impossible — no scoring interface exists

### Technical Notes
- Scoring panel requires team submission data: submitted answers per team per question. The `score_updated` event provides running totals but not submitted answer text. Investigate whether `host_begin_scoring` triggers a state snapshot that includes submissions, or whether a new mechanism is needed. Flag as IC-4 open item for DESIGN wave.
- `host_mark_answer` payload: `{ team_id, round_index, question_index, verdict }` where verdict is `"correct"` or `"wrong"`
- `score_updated` payload: `{ team_id, team_name, score }` — server-sent after each verdict
- Do not block "Publish Scores" on all verdicts being applied — Marcus may choose not to mark some (partial scoring edge case)

---

## US-05: Run Answer Ceremony on Display

### Problem
Marcus is a quizmaster who has published scores for a round. He finds it awkward to verbally describe answers while everyone stares at their phones — he wants to walk the room through the answers one by one on the shared display screen, building anticipation before revealing the answer.

### Who
- Quizmaster | Round scores published, room wants to see the answers | Needs to control the display screen to show each question and its answer

### Solution
Show a ceremony panel with "Show Next Question" and "Reveal Answer" buttons. Each question is shown first (building suspense), then the answer is revealed. The display screen reflects each step.

### Domain Examples

#### 1: Happy path — ceremony for 5 questions
Marcus clicks "Show Next Question" — Q1 appears on the display. Marcus gives people a moment to remember what they answered. He clicks "Reveal Answer" — "Paris" appears on the display. He repeats for Q2–Q5.

#### 2: Rapid walkthrough
For a quick ceremony, Marcus clicks through all questions and answers in sequence without pausing. The display screen updates correctly at each step.

#### 3: Ceremony complete — next round available
After walking through all 5 questions, the ceremony panel shows "Ceremony complete". Controls to "Start Round 2" become available.

### UAT Scenarios (BDD)

#### Scenario: Ceremony panel appears after publishing scores
```gherkin
Given Marcus has clicked "Publish Scores" for Round 1
Then the ceremony panel is visible with a "Show Next Question" button
And the ceremony progress shows "Question 0 of 5 shown"
```

#### Scenario: Show next question sends to display and play rooms
```gherkin
Given Marcus is on the ceremony panel
When Marcus clicks "Show Next Question" for the first time
Then a host_ceremony_show_question message is sent with question_index 0
And the display screen shows "Q1: What is the capital of France?"
And the play screen shows "Q1: What is the capital of France?"
And the ceremony progress shows "Question 1 of 5 shown"
And the "Reveal Answer" button is now visible
```

#### Scenario: Reveal answer sends to display room only
```gherkin
Given the ceremony is showing Question 1 on display and play
When Marcus clicks "Reveal Answer"
Then a host_ceremony_reveal_answer message is sent with question_index 0
And the display screen shows "Answer: Paris"
And the play screen does NOT show the answer
```

#### Scenario: All questions walked through — ceremony complete
```gherkin
Given Marcus has shown and revealed answers for all 5 questions in Round 1
Then the ceremony progress shows "Question 5 of 5 shown"
And the message "Ceremony complete" is shown
And the "Show Next Question" button is not visible
And controls to "Start Round 2" or "End Game" are available
```

### Acceptance Criteria
- [ ] Ceremony panel visible after publishing scores
- [ ] "Show Next Question" sends `host_ceremony_show_question` with incrementing `question_index`
- [ ] Ceremony progress counter updates after each question shown
- [ ] "Reveal Answer" sends `host_ceremony_reveal_answer` with matching `question_index`
- [ ] Answer shown on display room only (not play room)
- [ ] "Ceremony complete" shown when all questions walked through
- [ ] Post-ceremony controls: "Start Next Round" or "End Game"

### Outcome KPIs
- **Who**: Quizmasters running answer ceremony
- **Does what**: Walk through all answers on the display screen without skipping or double-sending
- **By how much**: Correct question/answer sequence on display for 100% of questions
- **Measured by**: QA session: display screen content matches expected question/answer at each step
- **Baseline**: Currently impossible — no ceremony controls exist

### Technical Notes
- `host_ceremony_show_question` at index 0 triggers `session.StartCeremony()` on server; subsequent indexes use `AdvanceCeremony()`
- `host_ceremony_reveal_answer` broadcasts only to `RoomDisplay` (confirmed in handler.go line 296)
- Track `ceremonyCursor` (0-based) in local state; increment after each "Show Next Question" click
- "Reveal Answer" is only valid after "Show Next Question" for the same index

---

## US-06: End Game and View Final Leaderboard

### Problem
Marcus is a quizmaster who has completed all rounds and scored everything. He finds it unsatisfying that there is no ceremonial end to the game — no final scores, no clear winner announcement, no way to formally close the session.

### Who
- Quizmaster | All rounds played, scored, and published | Needs to officially end the game and see the final standings

### Solution
Show an "End Game" button. Clicking it sends `host_end_game`, which triggers a `game_over` broadcast with final scores. The host panel displays the final leaderboard in rank order.

### Domain Examples

#### 1: Happy path — three teams, clear winner
After 3 rounds, Marcus clicks "End Game". The leaderboard shows: 1st The Brainiacs (12pts), 2nd Quiz Killers (9pts), 3rd Random Facts (7pts). Game controls disappear.

#### 2: Tie for first
Two teams are tied at 10pts. The leaderboard shows them in the tied positions. Marcus announces a tiebreaker round verbally (future feature).

#### 3: Early end game
Marcus decides to end after Round 2 without playing Round 3 (group needs to leave). He clicks "End Game" — the leaderboard shows scores from Rounds 1 and 2 only.

### UAT Scenarios (BDD)

#### Scenario: End Game button available after all rounds scored
```gherkin
Given Marcus has published scores for all rounds
Then the "End Game" button is visible
```

#### Scenario: End Game sends host_end_game and shows leaderboard
```gherkin
Given the "End Game" button is visible
When Marcus clicks "End Game"
Then a host_end_game message is sent
And the game_over event is received with final scores
And the leaderboard shows all teams sorted by score descending
And "The Brainiacs" with 12 points is listed first
And "Quiz Killers" with 9 points is listed second
And "Random Facts" with 7 points is listed third
And game control buttons (Start Round, End Round, etc.) are removed
```

#### Scenario: Final leaderboard sorted descending by score
```gherkin
Given the game_over event has been received
When the leaderboard renders
Then teams are sorted from highest to lowest score
And rank numbers (1st, 2nd, 3rd) are displayed
```

### Acceptance Criteria
- [ ] "End Game" button visible when appropriate (after all rounds scored, or any time Marcus chooses)
- [ ] Clicking "End Game" sends `host_end_game` with empty payload
- [ ] `game_over` event payload drives the leaderboard display
- [ ] Leaderboard shows all teams sorted descending by score
- [ ] Rank numbers displayed (1st, 2nd, etc.)
- [ ] Game control buttons removed from the panel after game_over

### Outcome KPIs
- **Who**: Quizmasters ending a game session
- **Does what**: See a clear final leaderboard and announce the winner
- **By how much**: Final leaderboard displayed correctly for 100% of game sessions
- **Measured by**: QA session: leaderboard matches expected scores after scoring all rounds
- **Baseline**: Currently impossible — no end game control or leaderboard exists

### Technical Notes
- `host_end_game` payload: `Record<string, never>` (empty object `{}`)
- `game_over` payload: `{ scores: Array<{ team_id, team_name, score }> }` — already sorted or sort client-side descending by score
- Remove all active game controls (Start Round, Reveal, Score, Publish, Ceremony) on game_over state
- Do not require all rounds to be played — "End Game" should be available whenever Marcus decides the game is over
