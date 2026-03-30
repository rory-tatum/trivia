<!-- markdownlint-disable MD024 -->
# User Stories -- Releases 2-5: Display, Media, Advanced Types, Resilience

## Metadata

- Feature ID: trivia
- Releases: 2 (Display), 3 (Media), 4 (Advanced Question Types), 5 (Resilience)
- Phase: DISCUSS -- Phase 4 (Requirements)
- Date: 2026-03-29
- Note: These stories build on the Release 1 walking skeleton. Each release delivers a demonstrable user outcome.

---

# Release 2: Full Display Integration

**User outcome:** The /display screen shows the current question in real-time so the room can read questions from the TV -- Marcus no longer needs to read aloud.

---

## US-06: Display Holding Screen

### Problem

Marcus is a quizmaster who wants a welcoming, professional-looking screen on the room TV before the game starts. Guests arriving late should be able to see the join URL on the TV without Marcus having to repeat it.

### Who

- All room participants | arriving before game starts | motivated to see what's happening and how to join

### Solution

When /display is opened and the game is in LOBBY state, show a holding screen with the quiz title, a "game starting soon" message, and the /play URL.

### Domain Examples

#### 1: Happy Path

The TV shows "Friday Night Trivia -- March 2026 | Game starting soon... | Join: http://trivia.local/play"

#### 2: Edge Case

/display is opened before Marcus has loaded a quiz. The screen shows a generic "Trivia Night -- waiting for host..." message.

#### 3: Error Case

/display opens and the WebSocket connection fails. The screen shows "Connecting..." until it reconnects.

### UAT Scenarios (BDD)

#### Scenario: Holding screen shows quiz title and join URL

Given Marcus has loaded a quiz and is in the lobby
When /display is opened on the TV
Then the screen shows "Friday Night Trivia -- March 2026"
And the /play URL is visible: "http://trivia.local/play"
And no question content is visible

#### Scenario: Holding screen updates when game starts

Given /display is showing the holding screen
When Marcus clicks "Start Game"
Then /display transitions to the first question view within 1 second

### Acceptance Criteria

- [ ] Holding screen shows quiz title, "game starting soon" message, and /play URL
- [ ] Holding screen is shown when game_state = LOBBY
- [ ] Screen auto-transitions to question view when game starts
- [ ] If no quiz is loaded, a generic message is shown

### Outcome KPIs

- **Who:** Late-arriving room participants
- **Does what:** See join URL on TV without asking the host
- **By how much:** 0 verbal URL repeats needed (qualitative)
- **Measured by:** KPI-06 (quizmaster satisfaction)
- **Baseline:** Marcus repeats URL multiple times as guests arrive

### Technical Notes

- Consumes ART-09 (game_state) and ART-01 (game_session_id) to get quiz title and /play URL
- Shown when game_state = LOBBY

---

## US-11: Display Shows Current Question in Real-Time

### Problem

Marcus is a quizmaster who currently must read each question aloud because the TV cannot show it automatically. This splits his attention between his laptop and the room, and players who miss a question verbally have no way to catch up.

### Who

- All room participants | question is being revealed | motivated to read the question on the TV without relying on the quizmaster's voice

### Solution

When Marcus reveals a question, /display updates within 1 second to show the question text (and media if applicable) as a large, readable view designed for TV/projector distance.

### Domain Examples

#### 1: Happy Path

Marcus reveals Q3. Within 1 second, the TV shows "Round 1: General Knowledge | Question 3 of 8 | Name this landmark." in large text. No answer, no other questions.

#### 2: Edge Case

Q3 has a multiple choice component: choices A-D appear below the question text in large font, allowing the room to discuss options.

#### 3: Error Case

Media file fails to load for Q4. The TV shows the question text and "Media unavailable" placeholder. Game continues.

### UAT Scenarios (BDD)

#### Scenario: Display updates to current question within 1 second of reveal

Given Marcus clicks "Reveal Q3" on /host
Then /display shows "Q3: Name this landmark." within 1 second
And the round name and question counter are in the header
And the correct answer is NOT shown on /display

#### Scenario: /display shows only the current (most recent) question

Given Marcus has revealed Q1, Q2, Q3
When /display is viewed
Then only Q3 is shown (not Q1, Q2)

#### Scenario: Multiple choice options shown on /display

Given Q4 has choices Venus, Mercury, Mars, Earth
When Marcus reveals Q4
Then /display shows the question text and all four choices labeled A through D

### Acceptance Criteria

- [ ] /display updates to show the most recently revealed question within 1 second (IC-02)
- [ ] /display shows question text in large, room-readable font
- [ ] /display shows only the current question (not cumulative revealed questions)
- [ ] Correct answer is never shown on /display before ceremony
- [ ] Multiple choice options are shown when present
- [ ] Upcoming (unrevealed) question content is never visible on /display

### Outcome KPIs

- **Who:** Room participants
- **Does what:** Read question from TV
- **By how much:** Marcus no longer needs to read questions aloud (KPI-06)
- **Measured by:** KPI-06 -- quizmaster satisfaction
- **Baseline:** Marcus reads aloud; split attention; players miss questions

### Technical Notes

- Consumes ART-04 (revealed_question_set) -- /display renders last item in set
- Question content (text, choices) sent to /display; answer fields stripped (ART-02 risk)

---

## US-17: Ceremony Scores View on /play

### Problem

Priya is a team captain who wants to see the correct answers and round scores on her phone during the ceremony -- currently she can only look at the TV and might miss something.

### Who

- Priya Nair | during ceremony | motivated to see answers and scores on her own device

### Solution

During the ceremony, /play shows a read-only view of each answer as it is revealed by Marcus, mirroring the /display content. After ceremony, round scores are shown.

### Domain Examples

#### 1: Happy Path

As Marcus walks through the ceremony, Priya's phone shows the same question + answer as the TV. After the ceremony, her phone shows the round scores including her team's position.

#### 2: Edge Case

Priya's phone was face down during the ceremony. She picks it up mid-ceremony and sees the most recent answer revealed.

#### 3: Error Case

Priya lost connection during ceremony. When she reconnects, she sees the current ceremony position.

### UAT Scenarios (BDD)

#### Scenario: /play shows ceremony answers in sync with /display

Given Marcus advances the ceremony to Q3
When Priya views her /play screen
Then Priya sees Q3 question text and "Answer: Eiffel Tower" (after Marcus reveals it)
And she can also see her team's submitted answer for Q3 for comparison

#### Scenario: /play shows round scores after ceremony

Given Marcus clicks "End Ceremony & Show Scores"
Then Priya's /play screen shows the round scores matching /display

### Acceptance Criteria

- [ ] /play shows ceremony content (question + answer) in sync with /display
- [ ] /play shows the team's own submitted answer alongside the correct answer
- [ ] After ceremony, /play shows round scores
- [ ] /play ceremony view is read-only (no input possible)

### Outcome KPIs

- **Who:** Players
- **Does what:** Follow ceremony from their phone
- **By how much:** No dependency on being able to see the TV clearly (KPI-06)
- **Measured by:** KPI-06
- **Baseline:** Can only watch TV; misses answers if screen is obscured

### Technical Notes

- Ceremony state consumed from ART-09 game_state = CEREMONY
- Player's own submitted answers from ART-06 shown alongside ART-07 correct answer

---

## US-20: Game Over Screen

### Problem

Players on their phones see nothing after the game ends if there is no dedicated final state for /play. The game just stops and the screen goes stale.

### Who

- Priya, Jordan | end of game | motivated to see final results and feel the game completed properly

### Solution

After "End Game", /play shows the same final scores as /display: standings, winner, total points per team.

### Domain Examples

#### 1: Happy Path

Priya's phone shows "TRIVIA NIGHT COMPLETE | Team Awesome: WINNER! | ..." after Marcus ends the game.

#### 2: Edge Case

A player who disconnected before the game ended reconnects to a finished game and sees the final scores.

#### 3: Error Case

Player opens /play after server restart. The session is gone; they see the fresh join screen.

### UAT Scenarios (BDD)

#### Scenario: /play shows game over screen when game ends

Given Marcus clicks "End Game"
When Priya views her /play screen
Then Priya sees the final standings with her team's total score
And no answer entry or game controls are visible

### Acceptance Criteria

- [ ] /play shows final standings when game_state = GAME_OVER
- [ ] Final standings match the /display final scores screen
- [ ] No game controls or input fields are shown
- [ ] A disconnected player who reconnects to a finished game sees the final scores

### Outcome KPIs

- **Who:** All players
- **Does what:** See final results on their device
- **By how much:** Included in KPI-06
- **Measured by:** KPI-06
- **Baseline:** No phone view after game ends

### Technical Notes

- ART-09 game_state = GAME_OVER triggers final scores broadcast to all clients

---

# Release 3: Rich Media Questions

**User outcome:** Marcus can include image, audio, and video questions in the quiz and they render correctly on /play and /display.

---

## US-21: Image Question Support

### Problem

Marcus is a quizmaster who runs "picture rounds" where players identify landmarks, logos, or faces from photos. Currently he opens Google Images on a separate screen, which is clumsy and inconsistent with the game flow.

### Who

- Marcus Okafor | hosting a picture round | motivated to show images within the game flow, not from a separate app

### Solution

YAML questions with an `image` field render with the referenced image displayed above the question text on /play and on /display. Images are served as static files from the quiz directory.

### Domain Examples

#### 1: Happy Path

Q3 has `image: "eiffel.jpg"`. On /play, the image appears above "Name this landmark." Players see the photo and answer in the same field. On /display, the image is shown prominently for the room.

#### 2: Edge Case

The image file is a large PNG. It loads within 2 seconds on a local WiFi network. If it takes longer, a loading spinner is shown.

#### 3: Error Case

The image file is missing at runtime (file was deleted after YAML validation). The question shows "Image unavailable" placeholder and the text question is still shown.

### UAT Scenarios (BDD)

#### Scenario: Image question displays correctly on /play

Given Q3 has image "eiffel.jpg" and text "Name this landmark."
When Marcus reveals Q3
Then /play shows the image "eiffel.jpg" above the question text
And an answer field is shown below the question text

#### Scenario: Image question displays on /display

Given Marcus reveals Q3 with an image
Then /display shows the image prominently with the question text below

#### Scenario: Missing image shows placeholder

Given the image file for Q3 is not found at serve time
When the question is revealed
Then both /play and /display show "Image unavailable" placeholder text
And the question text is still shown and answerable

### Acceptance Criteria

- [ ] YAML `image` field renders the referenced image above question text on /play and /display
- [ ] Images are served from the quiz file's directory
- [ ] Missing image at serve time shows a placeholder without breaking the question
- [ ] Answer field is present and functional regardless of image status

### Technical Notes

- Image served as static file relative to YAML location (DEC-007)
- Image validation at YAML load time (US-01) catches missing files before game starts
- Runtime missing file handled with placeholder (in case file is deleted/moved during session)

---

## US-22: Audio Question Support

### Problem

Marcus hosts music rounds where teams must identify songs and artists. Currently he plays audio from Spotify or YouTube on a separate device, which is awkward and disconnected from the game state.

### Who

- Marcus Okafor | hosting a music round | motivated to play audio within the game flow with a single reveal action

### Solution

YAML questions with an `audio` field automatically play the audio file when the question is revealed. /display shows "Now playing..." and the question text. /play shows the question and answer fields.

### Domain Examples

#### 1: Happy Path

Q1 in Round 2 has `audio: "mystery-track.mp3"`. Marcus reveals the question. The audio begins playing (on Marcus's device which is casting to the TV). /display shows "Round 2: Music | Q1 | Now playing... | Name this song and artist." /play shows the question with a multi-field answer input (for song name and artist).

#### 2: Edge Case

The audio file is longer than 60 seconds. The quizmaster controls when to advance; audio plays until he reveals the next question or stops it explicitly. There is no auto-stop.

#### 3: Error Case

Audio file is missing at serve time. /display shows "Audio unavailable" and Marcus can play the track from an alternate source.

### UAT Scenarios (BDD)

#### Scenario: Audio plays automatically on question reveal

Given Q1 in Round 2 has audio "mystery-track.mp3"
When Marcus reveals Q1
Then the audio file begins playing
And /display shows "Now playing..." with the question text
And /play shows the question text and answer fields

#### Scenario: Missing audio file shows placeholder

Given the audio file for Q1 is not found
When Marcus reveals Q1
Then "Audio unavailable" is shown on /display and /play
And the question text is still shown and answerable

### Acceptance Criteria

- [ ] YAML `audio` field causes audio to auto-play on question reveal
- [ ] /display shows "Now playing..." indicator when audio is active
- [ ] /play shows question text and answer fields (no audio player controls on /play -- audio plays through /display/cast)
- [ ] Missing audio at serve time shows placeholder without breaking the question

### Technical Notes

- Audio plays on the device running /display (Marcus's device casting to TV)
- Audio is served as a static file (DEC-007)
- /play does not play audio (player devices are separate from the speaker output)

---

## US-23: Video Question Support

### Problem

Marcus occasionally uses short video clips as questions. Currently he opens YouTube or a local video file separately, disrupting the game flow.

### Who

- Marcus Okafor | hosting a video question round | motivated to show video within the game UI

### Solution

YAML questions with a `video` field render a video player on /display and /play (muted on /play; with audio on /display). Video plays on reveal.

### Domain Examples

#### 1: Happy Path

Q4 has `video: "movie-clip.mp4"`. On reveal, /display shows the video clip (with audio via cast). /play shows a muted preview of the video with the question text.

#### 2: Edge Case

A large video file (50MB) takes several seconds to buffer. A loading spinner is shown until buffering is complete.

#### 3: Error Case

Video file is missing. Placeholder shown; game continues.

### UAT Scenarios (BDD)

#### Scenario: Video question displays on /display and /play

Given Q4 has video "movie-clip.mp4"
When Marcus reveals Q4
Then /display shows the video player and begins playback (with audio)
And /play shows a muted video preview with question text and answer field

### Acceptance Criteria

- [ ] YAML `video` field renders a video player on reveal
- [ ] /display plays video with audio; /play shows muted preview
- [ ] Missing video shows placeholder without breaking the question
- [ ] Answer field is present regardless of video status

### Technical Notes

- Video served as static file (DEC-007)
- /play video is muted to avoid audio conflicts between player devices

---

# Release 4: Advanced Question Types

**User outcome:** Marcus can include multiple choice and multi-part questions, and players see appropriate UI.

---

## US-24: Multiple Choice Question UI

### Problem

Marcus frequently includes multiple choice questions in his quiz. Currently players must write out their chosen answer from memory, when the options are already listed. A multiple choice UI reduces ambiguity and speeds answer entry.

### Who

- Priya Nair | answering a multiple choice question | motivated to select an answer with one tap rather than typing

### Solution

YAML questions with a `choices` array render as selectable options (radio-button style) on /play. /display shows the choices labeled A through D.

### Domain Examples

#### 1: Happy Path

Q4 has choices: Venus, Mercury, Mars, Earth. Priya sees four tappable options labeled A-D. She taps Mercury. Mercury is selected.

#### 2: Edge Case

Priya taps Venus then changes her mind and taps Mercury. Only Mercury is selected; selection is changeable until submission.

#### 3: Error Case

A YAML question has `choices` with only 1 option. The system renders it as a single-option choice (unusual but valid). No crash.

### UAT Scenarios (BDD)

#### Scenario: Multiple choice question renders as selectable options

Given Q4 has choices [Venus, Mercury, Mars, Earth]
When Marcus reveals Q4
Then /play shows four radio-button style options: A) Venus, B) Mercury, C) Mars, D) Earth
And /display shows the same labeled options

#### Scenario: Player changes their multiple choice selection

Given Priya has selected "A) Venus" for Q4
When Priya taps "B) Mercury"
Then Mercury is selected and Venus is deselected

### Acceptance Criteria

- [ ] YAML questions with `choices` array render as selectable options on /play
- [ ] /display shows choices labeled A through D (or A-N for more choices)
- [ ] Player can change selection at any time before submission
- [ ] Selection is stored as the team's draft answer for that question
- [ ] On submission, the selected choice is recorded as the team's answer

### Technical Notes

- `choices` array in YAML maps to radio-button UI on /play
- Answer stored as the choice text (not the letter) in ART-05/ART-06

---

## US-25: Multi-Part Answer Entry

### Problem

Marcus creates questions that have multiple answers (e.g., "Name the three primary colors" or "Name the song and artist"). Players need to enter each part separately so the quizmaster can score them independently.

### Who

- Priya Nair | answering a multi-part question | motivated to enter each answer part clearly so the quizmaster can mark each one

### Solution

YAML questions with an `answers` array render multiple answer fields on /play. The number of fields matches the number of expected answers. Fields are individually editable.

### Domain Examples

#### 1: Happy Path (Unordered)

Q2: `answers: [Red, Blue, Yellow]`, `ordered: false`. Priya sees three fields. She enters "Red", "Blue", "Yellow" in any order. Correct.

#### 2: Happy Path (Ordered)

Q7: `answers: [1st place, 2nd place, 3rd place]`, `ordered: true`. Priya must enter them in the specified sequence. The /play UI labels the fields "Part 1", "Part 2", "Part 3".

#### 3: Edge Case

Priya only fills 2 of 3 fields for a 3-part answer. The blank field is flagged on the review screen.

### UAT Scenarios (BDD)

#### Scenario: Multi-part unordered question shows multiple fields

Given Q2 has `answers: [Red, Blue, Yellow]` and `ordered: false`
When Marcus reveals Q2
Then /play shows three answer input fields
And the fields have no enforced order (no "Part 1/2/3" labels)

#### Scenario: Multi-part ordered question labels fields

Given Q7 has `answers: [...]` and `ordered: true`
When Marcus reveals Q7
Then /play shows the fields labeled "Part 1", "Part 2", etc.

### Acceptance Criteria

- [ ] YAML questions with `answers` array render multiple input fields
- [ ] Field count matches the number of expected answers
- [ ] `ordered: true` labels fields as Part 1, Part 2, etc.
- [ ] `ordered: false` shows unlabeled fields in any order
- [ ] Each field is independently editable

### Technical Notes

- `answers` (plural) in YAML triggers multi-part rendering
- `ordered` flag from YAML passed to /play for UI labeling and to /host for scoring context

---

## US-26: Multi-Part Answer Scoring

### Problem

Marcus needs to score multi-part answers where each part can be correct or incorrect independently. For unordered answers, any permutation is acceptable. For ordered answers, position matters.

### Who

- Marcus Okafor | scoring a multi-part question | motivated to score each part fairly and quickly

### Solution

The scoring interface shows each part of a multi-part answer separately. For ordered answers, parts are shown in sequence. For unordered answers, submitted parts are compared to expected parts without regard to order. Marcus marks each part correct or incorrect.

### Domain Examples

#### 1: Happy Path (Unordered)

Q2 expected: [Red, Blue, Yellow]. Team submitted: [Yellow, Red, Blue]. The scoring UI shows all three matches as aligned (order-insensitive comparison). Marcus confirms as correct.

#### 2: Happy Path (Ordered)

Q7 expected: [1st, 2nd, 3rd]. Team submitted: [2nd, 1st, 3rd]. The scoring UI shows Part 1: expected "1st", got "2nd" -- mismatch. Part 2: expected "2nd", got "1st" -- mismatch. Part 3: match. Marcus marks 1 of 3 correct.

#### 3: Edge Case

Partial credit: Marcus marks 2 of 3 parts correct. The question awards 2 points.

### UAT Scenarios (BDD)

#### Scenario: Unordered multi-part answers scored without regard to order

Given Q2 expects [Red, Blue, Yellow] with `ordered: false`
And Team Awesome submitted [Yellow, Red, Blue]
When Marcus views the Q2 scoring row for Team Awesome
Then the submitted parts are shown alongside expected parts without position penalties
And Marcus can mark the answer as correct with one click (all-or-nothing) or review each part

#### Scenario: Ordered multi-part answer mismatch highlighted

Given Q7 expects [1st, 2nd, 3rd] with `ordered: true`
And Team Awesome submitted [2nd, 1st, 3rd]
When Marcus views Q7 scoring
Then each part is shown in position: Part 1 shows "expected 1st, got 2nd" (mismatch highlighted)
And Marcus marks each part individually

### Acceptance Criteria

- [ ] Scoring interface shows multi-part answers with each part visible
- [ ] `ordered: false` answers show parts alongside expected without position enforcement
- [ ] `ordered: true` answers show parts by position with mismatch highlighted
- [ ] Each part can be marked correct or incorrect independently
- [ ] Score for the question = count of correct parts

### Technical Notes

- Multi-part answer structure stored in ART-06 as array of strings
- Scoring per-part creates multiple verdict entries in ART-07

---

# Release 5: Resilience and Polish

**User outcome:** The game handles real-world disruptions gracefully -- late joiners, disconnections, and edge cases -- without Marcus having to intervene.

---

## US-27: Quizmaster Override -- Open Scoring Before All Teams Submit

*(Already specified as an edge case in US-10. This story formalizes the override as a first-class feature with UI and confirmation.)*

### Problem

Marcus occasionally needs to proceed to scoring even when one team hasn't submitted (they got distracted or left early). Without an override, the game is blocked.

### Who

- Marcus Okafor | waiting for a team that isn't submitting | motivated to keep the game moving for everyone else

### Solution

An "Open Scoring Anyway" button on the submission status panel that, after a confirmation, treats the non-submitting team as having submitted blank answers.

### UAT Scenarios (BDD)

#### Scenario: Override opens scoring with blank answers for non-submitting team

Given Quiz Killers has not submitted and Marcus clicks "Open Scoring Anyway"
When Marcus confirms the override dialog
Then the scoring interface opens
And Quiz Killers shows blank answers for all questions in the scoring grid
And all other teams' submitted answers are shown normally

### Acceptance Criteria

- [ ] Override button is available from the submission status panel at any time after round ends
- [ ] Override requires a confirmation dialog naming the non-submitting teams
- [ ] After override, scoring opens immediately
- [ ] Non-submitting teams have blank answers in the scoring grid
- [ ] Blank answers are scoreable (Marcus can mark each as wrong)

---

## US-28: Late Joiner Catch-Up

### Problem

A player arrives late or opens /play after the game has started. Without catch-up, they see the lobby when the game is already in Round 2.

### Who

- Jordan Kim | joins mid-game | motivated to start participating immediately without needing Marcus to explain the state

### Solution

When a new device connects to /play after the game has started, the server sends the full current game state: current round, all revealed questions, any draft answers for the team token (if the team existed already), and the game phase.

### UAT Scenarios (BDD)

#### Scenario: Late joiner receives current game state

Given the game is in Round 2 with Q1 and Q2 revealed
When a new player opens /play and registers "Late Squad"
Then the player immediately sees Round 2 with Q1 and Q2 in their answer form
And the player can begin entering answers for Q1 and Q2

### Acceptance Criteria

- [ ] New connections during an active game receive the full current state snapshot
- [ ] Late joiners see the current round and all revealed questions
- [ ] Late joiners can enter and submit answers like any other team
- [ ] Marcus sees the late-joining team appear on the /host submission status

---

## US-29: Reconnection Handling

### Problem

Any client (player, display) can lose its WebSocket connection due to network issues. Without reconnection handling, the client shows stale content and the game breaks for that user.

### Who

- Priya Nair, display TV, Marcus | any network interruption | motivated to reconnect automatically without losing game state

### Solution

All clients implement automatic WebSocket reconnection with exponential backoff. On reconnect, the client re-subscribes and receives the current full game state. A "Reconnecting..." banner is shown during the outage.

### UAT Scenarios (BDD)

#### Scenario: /play client reconnects and restores state

Given Priya's phone loses connection for 30 seconds during Round 1 with 4 answers entered
When the connection is restored
Then the WebSocket reconnects automatically
And Priya's 4 draft answers are restored
And the current revealed questions are re-synced

#### Scenario: /display reconnects and shows current state

Given /display loses its WebSocket connection during Round 1
When the connection is restored
Then /display shows the current question (whatever Marcus has since advanced to)
And no manual intervention is needed

### Acceptance Criteria

- [ ] All clients auto-reconnect on WebSocket disconnect
- [ ] "Reconnecting..." banner shown during outage; dismissed on reconnect
- [ ] On reconnect, full current game state is received from server
- [ ] Draft answers from localStorage are restored alongside server state
- [ ] No manual action required from Marcus or players on reconnect

---

## US-30: YAML Validation Error Messages

*(US-01 specifies validation; this story adds field-level precision and error message quality standards)*

### Problem

When Marcus's YAML has an error, a generic message like "invalid YAML" tells him nothing. He needs to know the exact line, round, and field that is wrong.

### Who

- Marcus Okafor | preparing a quiz | motivated to fix YAML errors quickly without trial and error

### Solution

YAML validation produces field-level error messages with round number, question number, and specific issue description. Multiple errors are shown together, not one at a time.

### UAT Scenarios (BDD)

#### Scenario: Multiple errors shown at once

Given "quiz-with-errors.yaml" has 3 errors across 2 rounds
When Marcus loads the file
Then all 3 errors are shown simultaneously
And each error identifies the round, question, and field

#### Scenario: Error message is specific and actionable

Given Round 2, Question 3 is missing the "answer" field
When the error is shown
Then the message reads: "Round 2, Question 3: missing required field 'answer'"
And not just "YAML parse error"

### Acceptance Criteria

- [ ] All validation errors are shown at once (not one at a time)
- [ ] Each error message includes: round name/number, question number, field name, and issue
- [ ] At least the following error types are covered: missing required field, unknown field type, media file not found, invalid data type
- [ ] Errors are shown in the order they appear in the file

### Technical Notes

- Extends US-01 validation logic with richer error reporting
- Error format: `{round_name}, Question {N}: {field}: {reason}`
