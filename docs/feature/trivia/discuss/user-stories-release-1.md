<!-- markdownlint-disable MD024 -->
# User Stories -- Release 1: Walking Skeleton

## Metadata

- Feature ID: trivia
- Release: 1 (Walking Skeleton -- text-only, full game loop)
- Phase: DISCUSS -- Phase 4 (Requirements)
- Date: 2026-03-29
- Job story coverage: JS-01 through JS-08

---

## US-01: Load YAML Quiz File

### Problem

Marcus is a quizmaster who authors quiz content in YAML files. He finds it frustrating to have no tool that reads his existing format -- he currently pastes questions into a Google Doc and reads aloud from there, with no validation that his file is complete and correct before the game starts.

### Who

- Marcus Okafor | quizmaster, host of the evening | motivated to start the game quickly and confidently

### Solution

A /host page that accepts a YAML file path or upload, validates the structure and referenced files, and either initializes a game session or returns specific, actionable errors.

### Domain Examples

#### 1: Happy Path -- Valid YAML Loads Successfully

Marcus has prepared "friday-march-2026.yaml" in `/home/marcus/quizzes/`. He opens /host, provides the file path, and clicks Load. The system validates all 4 rounds, 32 questions, and confirms all fields are present. A game session is created. Marcus sees "Friday Night Trivia -- March 2026 | 4 rounds | 32 questions" and two shareable URLs.

#### 2: Edge Case -- YAML Missing Required Field

Marcus forgot to add an `answer` field to Round 2, Question 3. When he loads the file, the system shows: "Round 2, Question 3: missing required field 'answer'". No game session is created. Marcus opens the file, adds the field, re-saves, and re-loads without losing his /host session.

#### 3: Error Case -- Referenced Media File Not Found

Marcus's YAML references "eiffel.jpg" in Round 1, Question 4 but the file is in a different folder. The system reports: "Round 1, Question 4: image file 'eiffel.jpg' not found relative to quiz file location." Marcus moves the file and reloads.

### UAT Scenarios (BDD)

#### Scenario: Valid YAML loads and creates game session

Given Marcus opens /host in his browser
When Marcus provides the path to "friday-march-2026.yaml" containing 4 rounds and 32 text questions
Then the system validates the file structure successfully
And Marcus sees the quiz title "Friday Night Trivia -- March 2026", round count "4", and question count "32"
And a game session is initialized with a unique game session ID
And the lobby screen shows a /play URL and a /display URL Marcus can share

#### Scenario: YAML with missing required field shows specific error

Given Marcus provides "quiz-incomplete.yaml" where Round 2, Question 3 has no "answer" field
When the system validates the file
Then Marcus sees the error message "Round 2, Question 3: missing required field 'answer'"
And no game session is created
And Marcus can correct the YAML and reload without losing the /host page

#### Scenario: YAML referencing a missing media file shows specific error

Given "quiz-with-media.yaml" references "eiffel.jpg" which is not in the expected directory
When Marcus loads the file
Then Marcus sees "Round 1, Question 4: image file 'eiffel.jpg' not found"
And the load is rejected until the file is present or the reference is corrected

#### Scenario: YAML file not found at provided path

Given Marcus enters a path "/home/marcus/quizzes/wrong-name.yaml" that does not exist
When Marcus clicks Load
Then Marcus sees "File not found: /home/marcus/quizzes/wrong-name.yaml"
And Marcus can correct the path and retry

### Acceptance Criteria

- [ ] Valid YAML (all required fields present, all media found) creates a game session and shows lobby
- [ ] Missing required field produces error identifying the exact round, question, and field name
- [ ] Missing media file produces error identifying the exact round, question, and file path
- [ ] File not found produces an error with the provided path
- [ ] Errors do not create a partial game session
- [ ] After an error, Marcus can re-load without refreshing the /host page
- [ ] Game session ID is unique per load

### Outcome KPIs

- **Who:** Marcus (quizmaster)
- **Does what:** Loads a prepared YAML quiz file
- **By how much:** Under 2 minutes from opening /host to game session created
- **Measured by:** KPI-04 -- end-to-end setup timing
- **Baseline:** Currently no dedicated tool; manual setup is 5-10 minutes

### Technical Notes

- YAML parsing must validate: `title` (string), `rounds` (array), each round has `name` and `questions` array, each question has `text` and either `answer` (string) or `answers` (array)
- Media files referenced by relative path from YAML file location
- Answer/answers fields in the parsed quiz content tree must NEVER be sent to /play or /display clients (see ART-02 in shared-artifacts-registry.md)
- Server creates game_session_id (ART-01) on successful validation

---

## US-02: Game Lobby -- Share URLs and Monitor Teams

### Problem

Marcus is a quizmaster who needs all teams to be connected before starting the game. Currently he shouts out a URL and has no way to confirm everyone is ready without asking each person individually.

### Who

- Marcus Okafor | quizmaster in a room with guests | motivated to confirm all teams are ready before starting

### Solution

A lobby screen on /host that shows the shareable /play and /display URLs, lists connected teams in real-time, and offers a "Start Game" button once Marcus is satisfied.

### Domain Examples

#### 1: Happy Path -- All Teams Present, Game Starts

Marcus loads the quiz. He copies the /play URL from the lobby and pastes it in the group chat. Over 3 minutes, "Team Awesome", "The Brainiacs", and "Quiz Killers" appear on the lobby list. Marcus clicks "Start Game".

#### 2: Edge Case -- One Team Connected with Multiple Devices

The Brainiacs join from two different phones (Jordan and another player). The lobby shows "The Brainiacs (2 devices)". Marcus starts the game; both devices participate as one team.

#### 3: Error Case -- Attempting to Start with No Teams

Marcus loads the quiz but forgets to share the URL. He tries to click "Start Game" before anyone has joined. The button is greyed out with the label "Waiting for teams..." and a tooltip: "At least one team must join before starting."

### UAT Scenarios (BDD)

#### Scenario: Teams appear in lobby as they join

Given Marcus has loaded "friday-march-2026.yaml" and is on the lobby screen
When Team Awesome connects to /play and registers their team name
Then Marcus sees "Team Awesome" appear in the connected teams list within 2 seconds
And the team's connection timestamp is shown

#### Scenario: Quizmaster copies player URL

Given Marcus is on the lobby screen
When Marcus clicks "Copy Link" next to the /play URL
Then the URL "http://trivia.local/play" is copied to the clipboard
And Marcus sees a brief confirmation "Copied!"

#### Scenario: Quizmaster starts game with teams present

Given Team Awesome, The Brainiacs, and Quiz Killers are all listed in the lobby
When Marcus clicks "Start Game"
Then the game state transitions to ROUND_ACTIVE for Round 1
And all connected /play clients receive the game start event
And Marcus sees the Round 1 reveal panel

#### Scenario: Start Game blocked when no teams are present

Given Marcus has loaded a quiz but no teams have joined yet
When Marcus tries to click "Start Game"
Then the button is inactive
And Marcus sees the message "Waiting for teams to join..."

### Acceptance Criteria

- [ ] /play URL and /display URL are shown on the lobby screen and can be copied with one click
- [ ] Connected teams appear within 2 seconds of joining
- [ ] Each team entry shows team name and connection time
- [ ] Multiple devices for the same team show a device count, not duplicate team entries
- [ ] "Start Game" requires at least one team to be connected
- [ ] Clicking "Start Game" broadcasts game start to all /play and /display clients

### Outcome KPIs

- **Who:** Marcus (quizmaster)
- **Does what:** Confirms all teams are connected and starts the game
- **By how much:** Part of the under-2-minute setup target (KPI-04)
- **Measured by:** KPI-04
- **Baseline:** Verbal confirmation -- requires asking each person; 3-5 minutes

### Technical Notes

- Game lobby consumes ART-03 (team registry) updated in real-time via WebSocket
- Start Game action broadcasts game_state change (ART-09) to all clients (IC-01)

---

## US-03: Start Game -- Broadcast to All Clients

### Problem

Marcus is a quizmaster who needs all player devices and the display TV to simultaneously transition to the first round when he starts the game. Without this synchronization, some players see the lobby while others see questions, creating confusion.

### Who

- Marcus Okafor | quizmaster ready to begin | motivated to start the game cleanly, with the whole room transitioning together

### Solution

A "Start Game" action on /host that broadcasts a game_state change to all connected /play and /display clients simultaneously, causing them to transition from their lobby/waiting state to the round-active state.

### Domain Examples

#### 1: Happy Path -- All Clients Sync on Start

Marcus clicks "Start Game". Within 1 second, all six player devices (3 teams, 2 devices each) transition from "waiting for game" to showing the Round 1 answer form. The TV (/display) transitions from the holding screen to the Round 1 question view. Marcus's /host shows the Round 1 reveal panel.

#### 2: Edge Case -- A Player Joins Immediately After Start

Jordan opens /play 10 seconds after Marcus clicked "Start Game". The server recognizes the game is in ROUND_ACTIVE state. Jordan's device shows the current round and any questions already revealed, not the lobby.

#### 3: Error Case -- Player Device Was Offline During Start

Priya's phone was on airplane mode when Marcus clicked "Start Game". When Priya reconnects, her device receives the current game state (ROUND_ACTIVE) and catches up automatically, never showing the lobby.

### UAT Scenarios (BDD)

#### Scenario: All connected clients transition on game start

Given Team Awesome, The Brainiacs, and Quiz Killers are connected in the lobby
And /display is showing the holding screen
When Marcus clicks "Start Game"
Then all three /play clients show "Round 1: General Knowledge" within 1 second
And /display transitions from the holding screen to the question view within 1 second
And Marcus sees the Round 1 reveal panel

#### Scenario: Late-connecting player receives current state

Given Marcus has already started the game and revealed Q1 in Round 1
When a new device opens /play and registers as a new team
Then the device immediately sees Round 1 as active and Q1 revealed in their answer form
And the device does not see a lobby screen

#### Scenario: Reconnecting player after missed start

Given Priya was in the lobby on /play and lost connection before game start
When Priya's device reconnects after Marcus has started Round 1 and revealed Q2
Then Priya's device shows Round 1 active with Q1 and Q2 revealed
And any draft answers Priya had entered before disconnecting are restored

### Acceptance Criteria

- [ ] Game start broadcasts game_state=ROUND_ACTIVE to all connected clients within 1 second
- [ ] /display transitions from holding screen to question view on game start
- [ ] Clients that connect after game start receive the current game state, not the lobby
- [ ] Clients that were offline during game start catch up on reconnection
- [ ] Marcus sees the Round 1 reveal panel on /host after clicking Start Game

### Outcome KPIs

- **Who:** Marcus and all players
- **Does what:** All clients synchronize to game start simultaneously
- **By how much:** All clients transition within 1 second (IC-01)
- **Measured by:** WebSocket event timing in integration test
- **Baseline:** Verbal "ready" -- no technical synchronization today

### Technical Notes

- Broadcasts game_state (ART-09) change via WebSocket (DEC-008)
- Server must maintain current game_state for catch-up on new connections (IC-06)
- Integration checkpoint IC-01 must be verified

---

## US-04: Player Joins Game (First Visit)

### Problem

Priya is a team captain who needs to join the trivia game from her phone. She finds it disruptive to be interrupted during the social event to create an account, verify an email, or navigate a complex registration flow. She just wants to enter a team name and start playing.

### Who

- Priya Nair | team captain, first-time app user at a social event | motivated to join quickly and not look confused in front of her team

### Solution

A /play join screen that requests only a team name, registers the team, writes a persistence token to localStorage, and shows the lobby (or current game state if the game has started).

### Domain Examples

#### 1: Happy Path -- Team Registers and Sees Lobby

Priya opens `http://trivia.local/play` on her iPhone. She sees a single field: "Enter your team name" and a "Join Game" button. She types "Team Awesome" and clicks Join. She immediately sees the lobby showing "Team Awesome -- 1 player connected" and the other teams.

#### 2: Edge Case -- Multiple Devices, Same Team

Jordan also opens /play and types "Team Awesome". The server recognizes this as an additional device for the existing "Team Awesome" team (same team name, different device token). Jordan sees "Team Awesome -- 2 players connected" in the lobby.

#### 3: Error Case -- Team Name Taken by a Different Team

Another group has already registered "Quiz Killers". When Priya accidentally types "Quiz Killers" for her team, she sees: "That name is taken -- try a different team name." The field stays populated so she can edit it rather than retyping.

### UAT Scenarios (BDD)

#### Scenario: Player registers a new team on first visit

Given a game session is active in the lobby
When Priya opens /play and enters "Team Awesome" and clicks "Join Game"
Then a team identity token is written to Priya's browser localStorage
And Priya sees the lobby screen showing "Team Awesome" and a list of other connected teams
And Marcus sees "Team Awesome" appear on his /host lobby list within 2 seconds

#### Scenario: Second device joins the same team

Given Priya has already registered "Team Awesome"
When Jordan opens /play on a different device and enters "Team Awesome"
Then Jordan's device is added as a second device for the Team Awesome team
And the lobby shows "Team Awesome (2 devices)" on /host

#### Scenario: Duplicate team name is rejected

Given "Quiz Killers" is already registered
When Priya accidentally types "Quiz Killers" in the team name field
Then Priya sees the error "That name is taken -- try a different team name"
And the field remains editable with "Quiz Killers" pre-filled
And no new team registration is created

#### Scenario: Player joins after game has started

Given Marcus has started the game and Q1 has been revealed
When a new player opens /play, enters "Late Squad", and clicks Join
Then the player sees Round 1 active with Q1 already revealed
And the player can begin entering answers for Q1 immediately

### Acceptance Criteria

- [ ] Join form requires only team name (no email, no password, no account creation)
- [ ] Successful join writes a team identity token to localStorage
- [ ] Successful join shows the lobby (or current game state if game has started)
- [ ] Multiple devices with the same team name are treated as the same team
- [ ] Duplicate team name (different intent) shows an inline error without page reload
- [ ] Marcus sees the new team within 2 seconds on /host lobby list
- [ ] Joining after game start shows current round state, not lobby

### Outcome KPIs

- **Who:** Jordan (casual player, first time)
- **Does what:** Opens app URL and enters first answer
- **By how much:** Under 60 seconds (KPI-03)
- **Measured by:** KPI-03 -- timed test with new user
- **Baseline:** Paper: 2-3 minutes

### Technical Notes

- Team identity token is a UUID written to localStorage keyed by game_session_id (ART-01)
- Team name uniqueness enforced server-side; case-insensitive comparison recommended
- ART-03 (team registry) updated on join

---

## US-05: Auto-Rejoin on Browser Refresh

### Problem

Jordan is a casual player who accidentally refreshes the browser during Round 1. He finds it frustrating that he loses his team identity and has to re-enter his team name and explain to Marcus that he needs to be re-added to the game.

### Who

- Jordan Kim | casual player, mid-game | motivated to get back into the game immediately without disrupting Marcus or his team

### Solution

When a player's browser refreshes or the tab is re-opened, the app reads the localStorage team identity token, automatically rejoins the team to the current game session, and restores the current game state and any draft answers.

### Domain Examples

#### 1: Happy Path -- Refresh Mid-Round, All Answers Restored

Priya has entered 3 answers in Round 1 (Q1: "Paris", Q2: "Red, Blue, Yellow", Q3: "Eiffel Tower"). She accidentally refreshes. The page reloads, reads the localStorage token, rejoins "Team Awesome", and shows all three answers restored. The "Welcome back, Team Awesome!" message appears briefly then dismisses.

#### 2: Edge Case -- Tab Closed and Reopened 10 Minutes Later

Jordan closes his browser tab and returns 10 minutes later during Round 2 scoring. The app reads his token, rejoins "The Brainiacs", and shows the current game state (scoring in progress; answers locked). Jordan cannot edit answers -- they were already submitted -- but he can see the waiting screen.

#### 3: Error Case -- Token Present but Game Session No Longer Active

Priya opens /play the day after the game. The server has been restarted; the game session is gone. The app finds a localStorage token but the session_id is not recognized. Priya sees the join screen again with a note: "Your previous session has ended. Enter your team name to start a new game."

### UAT Scenarios (BDD)

#### Scenario: Player restores session on refresh with draft answers

Given Priya has entered 3 answers in Round 1 and refreshes the browser
When the page reloads
Then the app reads Priya's localStorage token and rejoins "Team Awesome"
And Priya sees "Welcome back, Team Awesome!"
And all 3 draft answers (Q1: Paris, Q2: Red/Blue/Yellow, Q3: Eiffel Tower) are restored
And the current game state (Round 1 active, 3 questions revealed) is shown

#### Scenario: Reconnection during scoring shows locked state

Given Team Awesome submitted Round 1 answers and Jordan refreshes during scoring
When the page reloads and Jordan rejoins
Then Jordan sees the waiting screen showing "Answers are locked for Round 1"
And Jordan cannot edit any answer fields

#### Scenario: Token present but session expired shows join screen

Given Priya has a valid localStorage token from a previous game
And the server has been restarted (session no longer exists)
When Priya opens /play
Then the app detects the token is invalid
And Priya sees the join screen with the message "Your previous session has ended"
And Priya can enter a new team name to join a new game

#### Scenario: Player with no localStorage token sees fresh join screen

Given Jordan opens /play on a new device with no stored token
When the page loads
Then Jordan sees the standard join screen with no "welcome back" behavior

### Acceptance Criteria

- [ ] Valid localStorage token on refresh automatically rejoins the team within 2 seconds
- [ ] Auto-rejoin shows a "Welcome back, {team name}!" message
- [ ] Draft answers from localStorage are restored after rejoin
- [ ] If submitted answers exist, the post-submission state is restored (read-only)
- [ ] An invalid or expired session token shows the join screen with an explanatory message
- [ ] No token shows the standard fresh join screen
- [ ] Auto-rejoin does not require Marcus to take any action

### Outcome KPIs

- **Who:** Priya, Jordan (players)
- **Does what:** Refresh browser without losing answers
- **By how much:** 0% answer loss on refresh (KPI-02)
- **Measured by:** KPI-02 -- automated test
- **Baseline:** Any refresh loses all progress in an app without persistence

### Technical Notes

- localStorage token: `{game_session_id}_{team_id}` keyed pair
- Server must maintain per-team state (ART-05 draft answers, ART-06 submitted answers) keyed by team token
- Reconnection triggers full state snapshot from server (IC-05)
- Draft answers stored in localStorage as backup; server draft store is authoritative

---

## US-07: Quizmaster Reveals Questions One at a Time

### Problem

Marcus is a quizmaster who needs to control the pace of question delivery. He finds it frustrating that without a tool, all questions are visible to players at once (if using a document) or he must manually track which question he's on while reading aloud. He wants to press a button and have the next question appear for everyone simultaneously.

### Who

- Marcus Okafor | quizmaster running an active round | motivated to maintain suspense and keep the room synchronized

### Solution

A reveal panel on /host that shows the current round's questions in order, with "Reveal Q[N]" buttons. Each click sends the question to all /play clients and updates /display. Already-revealed questions are shown as revealed; upcoming questions are hidden (content not visible).

### Domain Examples

#### 1: Happy Path -- Sequential Reveal

Marcus is in Round 1. He clicks "Reveal Q1". All players see "What is the capital of France?" and can begin answering. Two minutes later, Marcus clicks "Reveal Q2". Players now see Q1 and Q2. The /display shows Q2 as the current question.

#### 2: Edge Case -- Revealing an Image Question

Marcus clicks "Reveal Q3" which references "eiffel.jpg". The image appears on all /play clients above the question text. /display shows the image prominently. The image is served from the same directory as the YAML file.

#### 3: Error Case -- Revealing the Last Question

Marcus reveals Q8 (the last question in Round 1). The "Reveal Next Question" button is replaced by an "End Round" button. The reveal panel shows "8 of 8 revealed". Players see all 8 questions in their answer form.

### UAT Scenarios (BDD)

#### Scenario: Quizmaster reveals first question and players see it

Given Marcus is on the Round 1 reveal panel and no questions have been revealed
When Marcus clicks "Reveal Q1"
Then Q1 "What is the capital of France?" appears in all connected /play answer forms within 1 second
And /display updates to show Q1 as the current question within 1 second
And the /host reveal panel shows "1 of 8 revealed" with Q1 marked as [revealed]

#### Scenario: Revealing Q2 preserves Q1 on player view

Given Marcus has revealed Q1 and players have entered answers for Q1
When Marcus clicks "Reveal Q2"
Then Q2 appears below Q1 in all /play answer forms
And Q1 answer entries are preserved and remain editable
And /display now shows Q2 as the current question

#### Scenario: Upcoming questions are hidden from players

Given Marcus has revealed Q1 and Q2
When Priya views her /play answer form
Then only Q1 and Q2 are visible
And the text "Q3-Q8 not yet revealed" is shown
And the content of Q3-Q8 is not accessible to Priya's browser

#### Scenario: All questions revealed activates End Round

Given Marcus has revealed all 8 questions in Round 1
Then the "Reveal Next Question" button is replaced by "End Round"
And a note shows "All questions revealed"

### Acceptance Criteria

- [ ] Each reveal button sends only the corresponding question to /play and /display
- [ ] /play shows all revealed questions (cumulative); /display shows only the most recently revealed
- [ ] Answer entries for previously revealed questions are preserved when a new question is revealed
- [ ] Question content for unrevealed questions is NOT sent to /play or /display clients
- [ ] After all questions are revealed, "End Round" replaces the reveal button
- [ ] Reveal propagates to all clients within 1 second (IC-02)

### Outcome KPIs

- **Who:** Marcus (quizmaster)
- **Does what:** Controls question reveal pace
- **By how much:** Questions appear on all clients within 1 second
- **Measured by:** IC-02 latency measurement
- **Baseline:** Reading aloud; no technical control

### Technical Notes

- Reveal action broadcasts ART-04 (revealed_question_set) update via WebSocket (IC-02)
- ART-02 (quiz_content_tree) answer fields must be stripped before sending question data to /play and /display -- only question text, media ref, and choices (for MC) are sent
- Media files served as static assets; relative path from YAML file location (DEC-007)

---

## US-08: Player Enters and Edits Answers

### Problem

Priya is a team captain who needs to enter her team's answers during a round. She finds paper answer sheets frustrating because crossing out answers looks messy and suspicious to the quizmaster, and editing under time pressure creates illegible responses that cost points.

### Who

- Priya Nair | team captain actively playing a round | motivated to enter and edit answers freely without penalty until her team submits

### Solution

An answer form on /play that shows all currently revealed questions with editable text fields. Players can type, edit, delete, and re-enter answers at any time during the round. Answers persist across refreshes via localStorage.

### Domain Examples

#### 1: Happy Path -- Entering and Changing an Answer

Priya first types "Paris, France" in Q1. After discussion with her team, she changes it to "Paris". The field updates immediately. When Q2 is revealed, Q1 remains editable and retains "Paris".

#### 2: Edge Case -- Answering Before Discussion

Jordan quickly types "Venus" for the planet question. After the team discusses, they agree it's "Mercury". Jordan changes his answer to "Mercury". The old answer "Venus" is gone with no trace.

#### 3: Error Case -- No Answer Entered

The team simply doesn't know the answer to Q5. Priya leaves Q5 blank. The app does not warn about this during the round (only on the submit review screen). A blank answer is a legitimate choice.

### UAT Scenarios (BDD)

#### Scenario: Player enters an answer and it persists on next reveal

Given Q1 has been revealed and Priya has typed "Paris" in the Q1 answer field
When Marcus reveals Q2
Then Q1's answer field still shows "Paris"
And Q2's answer field is blank and ready for input
And Priya can still edit Q1's answer after Q2 appears

#### Scenario: Player changes an answer

Given Priya has typed "Venus" for Q4 (multiple-choice, but treating as text here)
When Priya clears the field and types "Mercury"
Then the field shows "Mercury"
And the draft answer stored is "Mercury", not "Venus"

#### Scenario: Answer persists after browser refresh

Given Priya has entered "Paris" for Q1, "Red/Blue/Yellow" for Q2, and "Eiffel Tower" for Q3
When Priya refreshes the browser
Then after auto-rejoin, all three answers are restored exactly as entered

#### Scenario: Blank answer is allowed

Given Q5 has been revealed and Priya's team does not know the answer
When Priya leaves the Q5 field empty
Then no error or warning is shown during the round
And the blank answer is recorded as the team's draft response for Q5

### Acceptance Criteria

- [ ] Answer field is present for each revealed question
- [ ] Answer field accepts text input and displays the current value
- [ ] Editing an answer overwrites the previous value with no confirmation needed
- [ ] All answer fields remain editable until the team submits
- [ ] Draft answers are saved to localStorage on every change
- [ ] Answers persist after browser refresh (via auto-rejoin, US-05)
- [ ] Blank answers are allowed and recorded as empty

### Outcome KPIs

- **Who:** Priya (team captain)
- **Does what:** Enters and edits answers freely during the round
- **By how much:** 0% forced errors from accidental answer overwrites vs. paper
- **Measured by:** Functional test -- enter, edit, verify; refresh and verify restore
- **Baseline:** Paper: editing means crossing out -- messy and irreversible

### Technical Notes

- Draft answers stored in ART-05 (draft_answers) in localStorage and synced to server
- Answer form renders one field per entry in ART-04 (revealed_question_set)
- For Release 1: text input only. Multi-part and multiple choice handled in Release 4 (US-25, US-26)

---

## US-09: Player Submits Answers with Confirmation

### Problem

Priya is a team captain who needs to submit her team's final answers at the end of a round. She is anxious about accidentally submitting before her team is ready, and she is also anxious that she might miss submitting altogether (which happened with paper sheets at a previous trivia night).

### Who

- Priya Nair | team captain at end of round | motivated to submit confidently, knowing the submission is final and correct

### Solution

A submit review screen on /play showing all answers, flagging blanks, with a two-step confirmation (review → confirm dialog → locked). After submission, answers are locked and a waiting screen shows other teams' submission status.

### Domain Examples

#### 1: Happy Path -- Full Submission with All Answers

Priya has answered all 8 questions. She opens the review screen, sees all 8 answers clearly, clicks "Submit Answers", sees the confirmation dialog ("Once submitted, you cannot change your answers"), clicks "Yes, Submit", and sees "Your answers are locked in." The other teams' submission status appears.

#### 2: Edge Case -- Submission with One Blank Answer

Q5 is blank. The review screen shows "!" beside Q5 with the note "no answer entered". Priya decides her team genuinely doesn't know Q5 and clicks "Submit Answers" anyway. The confirmation dialog mentions "You have 1 unanswered question (Q5)." Priya clicks "Yes, Submit".

#### 3: Error Case -- Player Cancels After Seeing Dialog

Priya opens the review screen, sees Q3 has a typo ("Efiel Tower"). She clicks "Submit Answers", sees the dialog, then clicks "Go Back". She returns to the answer form, corrects Q3 to "Eiffel Tower", and submits again.

### UAT Scenarios (BDD)

#### Scenario: Player reviews all answers before submitting

Given all 8 questions in Round 1 have been revealed and Priya has answered 7 questions
When Marcus ends the round and the submit button becomes available
And Priya clicks "Submit Answers"
Then Priya sees a review screen listing all 8 answers
And Q5 (blank) is flagged with "!" and the note "no answer entered"
And both "Go Back & Edit" and "Submit Answers" buttons are shown

#### Scenario: Confirmation dialog warns about blank and requires explicit confirm

Given Priya is on the review screen with Q5 blank
When Priya clicks "Submit Answers"
Then a modal dialog appears: "You have 1 unanswered question (Q5). Once submitted, you cannot change your answers."
And Priya must click "Yes, Submit" to proceed or "Go Back" to return to the form

#### Scenario: Successful submission locks answers and shows team status

Given Priya clicks "Yes, Submit" in the confirmation dialog
Then the submission is persisted on the server and acknowledged
And all answer fields on Priya's /play view become read-only
And Priya sees "Your answers are locked in." and a team submission status panel
And Marcus sees Team Awesome as "submitted" on his /host submission status screen

#### Scenario: Player cancels from confirmation dialog and corrects an answer

Given the confirmation dialog is open
When Priya clicks "Go Back"
Then the dialog closes and Priya returns to the answer form
And all previously entered answers are intact and editable
And Priya corrects Q3 and can re-submit

#### Scenario: Submitted answers cannot be edited

Given Team Awesome has submitted Round 1 answers
When Priya tries to edit Q1's answer field
Then the field is read-only and does not accept input
And a banner shows "Answers are locked for Round 1"

### Acceptance Criteria

- [ ] Submit button appears only after the round has ended (all questions revealed or quizmaster ends round)
- [ ] Review screen shows all questions and entered answers before final submission
- [ ] Blank answers are flagged with a visual indicator; submission is still allowed
- [ ] Confirmation dialog clearly states submission is final and shows blank count
- [ ] "Go Back" in dialog returns to editable answer form without losing any data
- [ ] "Yes, Submit" persists submission to server and returns acknowledgment before UI locks
- [ ] All answer fields become read-only immediately after confirmed submission
- [ ] Marcus sees team submission status update within 2 seconds of player submission
- [ ] Submission state (ART-06) is immutable once written (DEC-006)

### Outcome KPIs

- **Who:** Priya (team captain)
- **Does what:** Submits answers with confidence, no accidental submissions
- **By how much:** 100% of teams submit on their own (KPI-05)
- **Measured by:** KPI-05 -- session log of override use
- **Baseline:** Paper: manual hand-in; quizmaster must collect and chase teams

### Technical Notes

- Submission persists ART-06 (submitted_answers) on server
- Server must acknowledge before UI shows "locked in" -- client retry on network failure (IC-03)
- Once written, ART-06 is immutable (DEC-006)
- Submission state change broadcasts to /host via WebSocket (submission_status update)

---

## US-10: Quizmaster Monitors Submission Status

### Problem

Marcus is a quizmaster who needs to know when all teams have submitted before opening the scoring interface. Currently he has to ask each team "did you hand in your sheet?" verbally, which interrupts the social flow and sometimes leads to him starting scoring before everyone is done.

### Who

- Marcus Okafor | quizmaster waiting for round submissions | motivated to know exactly which teams have and have not submitted so he can proceed to scoring at the right time

### Solution

A submission status panel on /host that lists all teams with their submission state (submitted / waiting) updating in real-time. Scoring opens automatically when all teams have submitted, with a manual override available.

### Domain Examples

#### 1: Happy Path -- All Submit, Scoring Opens

Two minutes after the round ends, Marcus sees: "Team Awesome [submitted 2:14 ago], The Brainiacs [submitted 0:45 ago], Quiz Killers [submitted 0:05 ago]". The "Open Scoring" button activates.

#### 2: Edge Case -- One Team Takes a Long Time

Quiz Killers is still discussing. Marcus can see "Quiz Killers -- waiting..." and the other two teams submitted. Marcus decides to wait; a few moments later Quiz Killers submits and "Open Scoring" activates.

#### 3: Error Case -- Using Override

Quiz Killers never submits after 10 minutes. Marcus uses "Open Scoring Anyway" to proceed. Quiz Killers gets blank submissions for all questions in the scoring grid.

### UAT Scenarios (BDD)

#### Scenario: Submission status updates in real-time

Given the round has ended and Marcus is on the submission status panel
When Team Awesome submits their answers
Then Team Awesome's status changes from "waiting..." to "submitted X seconds ago" within 2 seconds

#### Scenario: Open Scoring activates when all teams submit

Given Team Awesome, The Brainiacs, and Quiz Killers have all submitted
Then the "Open Scoring" button becomes active
And Marcus can click it to proceed to the scoring interface

#### Scenario: Open Scoring remains inactive while any team is waiting

Given Team Awesome and The Brainiacs have submitted but Quiz Killers has not
Then the "Open Scoring" button is inactive
And Marcus sees "1 team has not yet submitted"

#### Scenario: Quizmaster uses override to proceed

Given Quiz Killers has not submitted after a long wait
When Marcus clicks "Open Scoring Anyway"
Then a confirmation dialog warns "Quiz Killers has not submitted -- their answers will be blank"
And on confirm, the scoring interface opens with Quiz Killers showing blank answers for all questions

### Acceptance Criteria

- [ ] Submission status panel shows all teams with submitted/waiting status
- [ ] Status updates within 2 seconds of a team's submission
- [ ] "Open Scoring" is inactive until all teams have submitted
- [ ] "Open Scoring" activates when the last team submits
- [ ] Override button is available at all times after round ends
- [ ] Override requires a confirmation dialog before proceeding
- [ ] Teams that did not submit show blank answers in the scoring interface

### Outcome KPIs

- **Who:** Marcus (quizmaster)
- **Does what:** Knows exactly when to proceed to scoring
- **By how much:** Zero instances of scoring opening before all teams ready (KPI-05)
- **Measured by:** KPI-05 -- override never needed in normal play
- **Baseline:** Verbal check -- unreliable, disruptive

### Technical Notes

- Submission status consumes ART-06 updates via WebSocket
- Override writes blank submitted_answers for the non-submitting team before opening scoring

---

## US-12: Quizmaster Scoring Interface

### Problem

Marcus is a quizmaster who needs to mark each team's submitted answers as correct or incorrect. He finds manual scoring -- reading paper sheets, comparing to an answer key, tallying points, and writing them down -- to be the most time-consuming and error-prone part of running a trivia night, often taking 10-15 minutes per round.

### Who

- Marcus Okafor | quizmaster in the scoring phase | motivated to mark answers as quickly and accurately as possible so the game maintains momentum

### Solution

A scoring interface on /host that presents questions one at a time with the expected answer visible, and for each question shows all teams' submitted answers as rows with "correct" and "wrong" buttons. Scores are tallied automatically as Marcus clicks.

### Domain Examples

#### 1: Happy Path -- Scoring Q1 Across Three Teams

Marcus sees Q1 ("What is the capital of France?" | Answer: "Paris"). Below it: "Team Awesome: paris [correct] [wrong]", "The Brainiacs: PARIS [correct] [wrong]", "Quiz Killers: Lyon [correct] [wrong]". Marcus clicks "correct" for Awesome and Brainiacs (case-insensitive), "wrong" for Killers. The running scores update: Awesome 1, Brainiacs 1, Killers 0.

#### 2: Edge Case -- Partial Correct on a 2-Point Question

Marcus uses this story only for 1-point questions. Multi-part scoring is handled in US-26 (Release 4). For Release 1, each question is worth 1 point.

#### 3: Error Case -- Accidentally Clicked Wrong on a Correct Answer

Marcus clicks "wrong" for Team Awesome's answer but then sees it's actually correct. He clicks "correct" for Team Awesome on the same row. The verdict switches from wrong to correct and the score adjusts. Toggle behavior.

### UAT Scenarios (BDD)

#### Scenario: Scoring interface shows expected answer and all team submissions

Given all teams have submitted Round 1 answers and Marcus opens scoring
When Marcus views Q1 in the scoring interface
Then Marcus sees "Q1: What is the capital of France? | Answer: Paris"
And below it, Team Awesome's submission "paris", The Brainiacs' "PARIS", and Quiz Killers' "Lyon"
And each row has [correct] and [wrong] buttons

#### Scenario: Marking an answer correct increments score

Given Marcus is scoring Q1
When Marcus clicks "correct" for Team Awesome's answer "paris"
Then Team Awesome's round score increments by 1
And the score display shows "Team Awesome: 1 | The Brainiacs: 0 | Quiz Killers: 0"
And the answer row for Team Awesome is highlighted green

#### Scenario: Marking an answer wrong does not increment score

Given Marcus clicks "wrong" for Quiz Killers' answer "Lyon"
Then Quiz Killers' score does not change
And the answer row for Quiz Killers is highlighted red

#### Scenario: Quizmaster can change a verdict

Given Marcus has clicked "wrong" for Team Awesome's answer
When Marcus clicks "correct" for Team Awesome on the same question
Then the verdict changes to correct
And Team Awesome's score is recalculated to include the point

#### Scenario: Running totals update after each verdict

Given Marcus has scored Q1 through Q3
Then the running total display shows the cumulative score for each team across all marked questions
And the display updates in real-time without requiring a page action

#### Scenario: All answers marked enables Start Ceremony

Given Marcus has marked correct or wrong for every question and every team
Then the "Save Scores & Start Ceremony" button becomes active

### Acceptance Criteria

- [ ] Scoring interface shows each question with expected answer and all team submissions
- [ ] Correct/wrong buttons are present for each team per question
- [ ] Marking correct increments the team's score by 1 point
- [ ] Marking wrong leaves the score unchanged
- [ ] Verdict can be toggled before the ceremony starts
- [ ] Running round and running total scores update automatically on each verdict
- [ ] "Save Scores & Start Ceremony" becomes active when all question/team combinations have a verdict
- [ ] Expected answers are not visible to players or /display during scoring phase

### Outcome KPIs

- **Who:** Marcus (quizmaster)
- **Does what:** Scores a complete round
- **By how much:** Under 3 minutes for 5 teams x 10 questions (KPI-01)
- **Measured by:** KPI-01 -- timed test
- **Baseline:** 10-15 minutes manual paper scoring

### Technical Notes

- Scoring interface consumes ART-06 (submitted_answers)
- Scoring results written to ART-07 (scored_answers)
- Running totals computed from ART-07 and stored in ART-08 (round_scores)
- Expected answers (from ART-02) are shown only on /host during scoring -- never sent to /play or /display

---

## US-14: Auto-Tally Scores

### Problem

Marcus is a quizmaster who currently adds up scores manually in a spreadsheet or on paper after marking answers. Even with a scoring interface, if he has to manually total the points, he introduces calculation errors and adds 2-3 minutes to scoring time.

### Who

- Marcus Okafor | quizmaster completing scoring | motivated to have scores ready instantly without manual calculation

### Solution

As Marcus marks each answer correct or incorrect, running round scores and cumulative game totals update automatically in the scoring interface. When all answers are marked, final scores are ready with no additional action required.

### Domain Examples

#### 1: Happy Path -- 3 Teams, 8 Questions, Auto-Tallied

Marcus marks all 24 answer slots (3 teams x 8 questions). As he marks each, he can see in the corner: "Team Awesome: 6 | The Brainiacs: 5 | Quiz Killers: 3" updating after each click. He never picks up a calculator.

#### 2: Edge Case -- Changing a Verdict After Seeing the Total

Marcus marks Q7 wrong for The Brainiacs but then reconsiders. He changes it to correct. The Brainiacs' score updates from 4 to 5. The running total adjusts.

#### 3: Error Case -- Partial Scoring

Marcus scores Q1 through Q4 and needs a break. The running totals show partial scores (only Q1-Q4 counted). When Marcus returns and finishes Q5-Q8, the totals incorporate all questions.

### UAT Scenarios (BDD)

#### Scenario: Score auto-updates after each verdict

Given Marcus is on the scoring interface for Round 1
When Marcus marks Team Awesome's Q1 answer as correct
Then Team Awesome's round score shows "1" immediately
And the running total for Team Awesome updates to include this round's score

#### Scenario: Changing a verdict recalculates the score

Given Marcus has marked Q3 wrong for The Brainiacs (score: 2)
When Marcus changes Q3 to correct for The Brainiacs
Then The Brainiacs' round score updates from 2 to 3 immediately

#### Scenario: Final scores are ready when all answers marked

Given Marcus has marked correct or wrong for all 24 question-team combinations
Then round scores show the final totals: Team Awesome: 6, The Brainiacs: 5, Quiz Killers: 3
And running totals are computed across all previously completed rounds
And no further calculation is required before the ceremony

### Acceptance Criteria

- [ ] Round score for each team updates after every verdict without any additional action
- [ ] Running total updates automatically to reflect the current round's score plus all previous rounds
- [ ] Changing a verdict recalculates the score correctly (no stale values)
- [ ] Final scores are accurate and match the sum of all correct verdict marks
- [ ] Scores are available to the ceremony UI without any export or manual entry step

### Outcome KPIs

- **Who:** Marcus (quizmaster)
- **Does what:** Obtains final scores without manual calculation
- **By how much:** 0 arithmetic errors; included in KPI-01 timing target
- **Measured by:** Automated test -- mark N correct answers; verify score = N
- **Baseline:** Spreadsheet/mental math -- error-prone, slow

### Technical Notes

- ART-08 (round_scores) computed server-side from ART-07 (scored_answers) after each verdict write
- Score computation must be deterministic and recalculate fully on verdict change (not incremental to avoid drift)

---

## US-15: Answer Ceremony -- Quizmaster Drives Display

### Problem

Marcus is a quizmaster who wants to make the answer reveal feel like a performance moment, not just an announcement. With paper scoring, he reads out "Question 1 -- Paris. Question 2 -- Red, Blue, Yellow." in a flat list. He wants each answer to appear on the TV with the question, so the room can read and react together.

### Who

- Marcus Okafor | quizmaster in ceremony mode | motivated to create a moment of suspense and delight for each revealed answer

### Solution

A ceremony control panel on /host that allows Marcus to step through each question one at a time. For each step, the /display shows the question text and correct answer. Marcus can see which teams got it right and can announce them verbally.

### Domain Examples

#### 1: Happy Path -- Walking Through All 8 Answers

Marcus clicks "Start Ceremony". The /display shows Q1 ("What is the capital of France?") without the answer. Marcus clicks "Reveal Answer". "Paris" appears on /display. Marcus sees on his /host panel that Team Awesome and The Brainiacs got it right; Quiz Killers did not. He announces: "Team Awesome and the Brainiacs got that one!" He clicks "Next Question" to advance to Q2.

#### 2: Edge Case -- Pausing on a Question

Marcus pauses on Q5 to tell a story about that question. The /display continues showing Q5. Nothing advances until Marcus clicks "Next Question". He controls the pace entirely.

#### 3: Error Case -- Accidentally Skipping a Question

Marcus inadvertently clicks "Next Question" twice and skips from Q4 to Q6. He sees a "Previous Question" button and clicks back to Q5. The /display returns to Q5.

### UAT Scenarios (BDD)

#### Scenario: Ceremony starts and display shows first question

Given Marcus has completed scoring and clicks "Save Scores & Start Ceremony"
Then /display transitions to ceremony mode showing Q1 question text
And the correct answer for Q1 is NOT yet shown on /display
And Marcus sees the ceremony control panel on /host with Q1 selected and the correct answer "Paris" visible to him

#### Scenario: Marcus reveals an answer and it appears on display

Given the ceremony is showing Q1 question text
When Marcus clicks "Reveal Answer"
Then "Answer: Paris" appears on /display below the Q1 question text
And Marcus sees on /host which teams answered Q1 correctly

#### Scenario: Marcus advances to the next question

Given Q1 answer has been revealed
When Marcus clicks "Next Question"
Then /display transitions to Q2 question text (without answer)
And Marcus sees Q2 in his ceremony panel with the correct answer visible to him

#### Scenario: Marcus can go back to a previous question

Given Marcus is on Q3 in the ceremony
When Marcus clicks "Previous Question"
Then /display returns to Q2 showing question + answer (already revealed)

#### Scenario: Ceremony complete transitions to round scores

Given Marcus has stepped through all 8 questions and revealed all 8 answers
When Marcus clicks "End Ceremony & Show Scores"
Then /display shows the Round 1 scores in rank order
And /play clients show the round scores view
And Marcus sees an option to start Round 2 or end the game

### Acceptance Criteria

- [ ] Ceremony mode on /host shows a step control panel for each question in the round
- [ ] /display shows question text only (no answer) when Marcus advances to a question
- [ ] Marcus can reveal the answer with a separate button; answer appears on /display on click
- [ ] Marcus can navigate forward and backward through questions
- [ ] Marcus can see on /host which teams got each question right (not shown on /display)
- [ ] "End Ceremony & Show Scores" transitions /display to round scores
- [ ] Ceremony is driven entirely by Marcus -- no auto-advance, no timer

### Outcome KPIs

- **Who:** Marcus (quizmaster), room participants
- **Does what:** Experience a structured, paced answer reveal
- **By how much:** Ceremony creates observable moments of engagement (groans, cheers) -- JS-05 fulfilled
- **Measured by:** KPI-06 -- quizmaster satisfaction
- **Baseline:** Flat verbal read-out; no shared visual; no suspense

### Technical Notes

- Ceremony control on /host drives /display state changes via WebSocket broadcast per step
- ART-07 (scored_answers) used to compute per-team correctness for /host ceremony panel display
- Answer content (ART-02) is safe to send to /display ONLY during ceremony, keyed by current ceremony question index

---

## US-16: Round Scores on /display

### Problem

Marcus is a quizmaster who wants everyone in the room to see the scores after each round without requiring them to check their phones. Currently he reads them aloud from his paper notes, which is easy to mishear or not see.

### Who

- All room participants | post-ceremony, waiting for next round | motivated to see their standing in the game at a glance from across the room

### Solution

After the ceremony, /display shows a round scores screen with team names ranked by round score, showing round points and running total. Stays visible until Marcus starts the next round.

### Domain Examples

#### 1: Happy Path -- Scores Show After Ceremony

After the Round 1 ceremony, /display transitions to: "Round 1 Scores. 1. Team Awesome: 8 pts (Running: 8). 2. The Brainiacs: 6 pts (Running: 6). 3. Quiz Killers: 4 pts (Running: 4)."

#### 2: Edge Case -- Tied Teams

Team Awesome and The Brainiacs both scored 6 in Round 1. The display shows them tied at rank 1 with the same score. No arbitrary tie-breaking is applied.

#### 3: Error Case -- Scores Screen Visible Too Long

Marcus forgets to start Round 2 for 5 minutes. The scores screen stays up -- this is correct behavior. The display waits for Marcus to start the next round.

### UAT Scenarios (BDD)

#### Scenario: Round scores displayed in rank order after ceremony

Given Marcus clicks "End Ceremony & Show Scores" after Round 1 ceremony
Then /display shows "Round 1: Complete" header
And teams are listed in descending score order: Team Awesome (8), The Brainiacs (6), Quiz Killers (4)
And each row shows round points and running total
And the footer shows "Round 2 starts shortly..."

#### Scenario: Tied teams shown at same rank

Given Team Awesome and The Brainiacs both scored 6 in Round 1
When /display shows the round scores
Then both teams appear at rank 1 with the same score
And no arbitrary ordering separates them

#### Scenario: /play clients also show round scores

Given the ceremony is complete
Then all /play clients show the current round scores matching /display
And /play clients cannot edit answers or interact during the scores screen

### Acceptance Criteria

- [ ] /display shows round scores in descending score order after ceremony completes
- [ ] Each team row shows: rank, team name, round points, running total
- [ ] Tied teams are shown at the same rank without arbitrary ordering
- [ ] Scores screen persists until Marcus starts the next round
- [ ] /play clients see the same score data (read-only)
- [ ] Scores are derived from ART-08 (auto-calculated, not manually entered)

### Outcome KPIs

- **Who:** All room participants
- **Does what:** See scores on TV without checking phones
- **By how much:** 100% of rounds end with scores on TV (KPI-07 -- display information correctly shown)
- **Measured by:** Manual session verification
- **Baseline:** Verbal announcement; whiteboard; not visible to all

### Technical Notes

- /display scores screen state driven by ART-09 game_state = ROUND_SCORES
- ART-08 (round_scores) sent to /display and /play on ceremony_complete event

---

## US-18: Advance to Next Round

### Problem

Marcus is a quizmaster who needs to move from one round to the next. After scores are shown, he wants to start Round 2 cleanly -- resetting the question reveal state, clearing submitted answers from the previous round, and transitioning all clients to the new round's context.

### Who

- Marcus Okafor | quizmaster between rounds | motivated to keep game momentum with a clean, fast round transition

### Solution

A "Start Round 2" button on /host that resets the game state for the next round (new question set, empty answer fields, fresh submission status) and broadcasts the transition to all clients.

### Domain Examples

#### 1: Happy Path -- Clean Round Transition

After Round 1 scores are shown, Marcus clicks "Start Round 2: Music". All /play clients clear the Round 1 answer form and show "Round 2: Music -- waiting for first question." /display transitions to a new question view ready for Round 2 reveals. Marcus sees the Round 2 reveal panel.

#### 2: Edge Case -- Final Round

After Round 4 (the last round), the option shown is "End Game" instead of "Start Next Round". Marcus clicks "End Game" and the final scores screen appears.

#### 3: Error Case -- Accidentally Starting Round 2 Early

Marcus accidentally clicks "Start Round 2" before reviewing the scores. Round 2 begins. Marcus cannot go back to Round 1 questions (they are locked and submitted). The Round 1 scores remain visible on the scores panel sidebar.

### UAT Scenarios (BDD)

#### Scenario: Quizmaster starts the next round and all clients transition

Given Round 1 is complete and scores are displayed
When Marcus clicks "Start Round 2"
Then all /play clients transition to "Round 2: Music" with an empty answer form
And /display shows the Round 2 question view ready for reveals
And Marcus sees the Round 2 reveal panel with "0 of 6 questions revealed"

#### Scenario: Final round shows End Game instead of Start Next Round

Given Round 4 (the last round) ceremony is complete and scores are shown
When Marcus views the /host scores panel
Then Marcus sees "End Game" instead of "Start Round 5"

#### Scenario: End Game triggers final scores screen

Given Marcus clicks "End Game" after the last round ceremony
Then /display transitions to the final scores screen with full standings
And /play clients show the final scores view
And no further game actions are available

### Acceptance Criteria

- [ ] "Start Round [N]" button is available on /host after round scores are shown
- [ ] Clicking the button resets revealed_question_set (ART-04) for the new round
- [ ] All /play clients transition to the new round's empty answer form
- [ ] /display transitions to the new round question view
- [ ] Submitted answers from the previous round remain stored (not deleted) for reference
- [ ] After the last round, "End Game" replaces "Start Next Round"
- [ ] "End Game" triggers the final scores screen on /display and /play

### Acceptance Criteria (continued)

- [ ] Round number and round name are shown on all interfaces during the new round

### Outcome KPIs

- **Who:** Marcus (quizmaster), all players
- **Does what:** Transition between rounds seamlessly
- **By how much:** Round transition takes under 10 seconds from button click to all clients updated
- **Measured by:** WebSocket broadcast timing
- **Baseline:** Verbal "okay let's move on"; reshuffling paper

### Technical Notes

- Round transition broadcasts ART-09 game_state = ROUND_ACTIVE with new round_number
- ART-04 (revealed_question_set) resets to empty for new round
- ART-05 and ART-06 from previous round remain in server memory (for scoring review if needed)

---

## US-19: Final Scores and Winner Announcement

### Problem

Marcus is a quizmaster who wants the end of the game to feel like a proper finish, not just silence after the last round. With paper scoring, he tallies the final numbers manually and announces the winner verbally -- anticlimactic and error-prone.

### Who

- All room participants | end of game | motivated to see the final standings and celebrate the winner

### Solution

A final scores screen on /display (and mirrored on /play) that shows the complete final standings in rank order with the winner highlighted. Shown after the last round ceremony completes.

### Domain Examples

#### 1: Happy Path -- Clear Winner

After the Round 4 ceremony and Marcus clicks "End Game", /display shows: "TRIVIA NIGHT COMPLETE. Final Standings. 1. Team Awesome: 32 pts WINNER! 2. The Brainiacs: 27 pts. 3. Quiz Killers: 19 pts. Thanks for playing!"

#### 2: Edge Case -- Tied Winner

Team Awesome and The Brainiacs both score 30 points. The display shows them both at rank 1 with the notation "TIE!" No automatic tie-breaker -- Marcus handles tie resolution verbally or with a tiebreaker question (out of scope for this story).

#### 3: Error Case -- Final Screen After Server Restart

The server is accidentally restarted after the game ends but before Marcus screenshots the scores. The in-memory scores are gone. This is an accepted limitation (DEC-004 -- no persistence). Marcus should screenshot before closing.

### UAT Scenarios (BDD)

#### Scenario: Final scores displayed with winner highlighted

Given Marcus has clicked "End Game" after the Round 4 ceremony
Then /display shows "TRIVIA NIGHT COMPLETE"
And all teams are listed in descending order by total score
And the top-ranked team has a "WINNER!" visual indicator
And each row shows team name and total points across all rounds

#### Scenario: Tied top scores both show as winners

Given Team Awesome (30 pts) and The Brainiacs (30 pts) are tied for first
When /display shows the final scores
Then both Team Awesome and The Brainiacs appear at rank 1
And both show a "TIE!" indicator

#### Scenario: /play clients show the same final scores

Given the game is over
When Priya views her /play screen
Then Priya sees the same final standings as /display
And all answer and submission UI is permanently locked

### Acceptance Criteria

- [ ] Final scores screen shows all teams ranked by total score (descending)
- [ ] Winner (top score) has a visual distinction (e.g., "WINNER!" label)
- [ ] Tied scores at any rank are shown at the same rank without arbitrary ordering
- [ ] /play clients mirror the final scores view
- [ ] Final scores screen persists indefinitely (until server restart)
- [ ] No game actions are available after the game over state

### Outcome KPIs

- **Who:** All room participants
- **Does what:** See and celebrate the winner on the TV
- **By how much:** 100% of games end with a visible winner on /display (KPI-06)
- **Measured by:** Session verification
- **Baseline:** Verbal announcement; written on whiteboard; forgettable

### Technical Notes

- ART-08 (round_scores) aggregated across all rounds for final totals
- ART-09 game_state = GAME_OVER triggers final scores broadcast
- In-memory only -- no persistence (DEC-004)
