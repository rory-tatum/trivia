# Journey Visual — Player Interface (/play)

## Metadata

- Feature ID: play-ui
- Interface: /play (Player / Team)
- Personas: Priya Nair (team captain, anxious about losing answers), Jordan Kim (casual player, rejoins mid-game)
- Date: 2026-04-15
- Extends: docs/ux/trivia/journey-play-visual.md (original trivia DESIGN wave)
- JTBD grounding: J1 (join), J2 (rejoin), J3 (answer), J4 (submit), J5 (ceremony), J6 (orientation), J7 (question types)

---

## Personas

### Priya Nair — Team Captain
- Role: The one responsible for entering answers on behalf of Team Awesome.
- Device: Android phone, Chrome browser.
- Fear: Accidentally submitting too early or losing typed answers.
- Goal: Get everyone's input captured and submitted correctly every round.
- Emotional starting point: Slightly anxious (responsible for the team's score).

### Jordan Kim — Casual Team Member
- Role: Participates in discussion, occasionally contributes an answer.
- Device: iPhone, Safari.
- Context: Often steps away for a drink; rejoins mid-round.
- Fear: Coming back and finding the game does not recognise them.
- Goal: Be part of the experience without causing problems.
- Emotional starting point: Relaxed but needs quick onboarding.

---

## Emotional Arc

```
EMOTION
  ^
  |
  | UNCERTAIN   RELIEVED   ENGAGED      FOCUSED        CONFIDENT    ELATED
  | (joining)   (lobby)    (answering)  (submitting)   (waiting)    (ceremony)
  |
  |   [1,2]       [3]       [4,5,6,7]     [8,9]          [10]        [11,12]
  |    * *          *          * * * *       * *              *          * *
  |     \ \        / \        /  |  |  \     |  \            |         /   \
  |      \ *------*   *------*   |  |   *----*   *----------*         *     *
  |       \                      |  |                                        \
  |        *--------------------*    *------------------------------------------
  |
  +----------------------------------------------------------------------> TIME
       Join      Lobby    Answer Q1..N    Review+Submit  Waiting   Ceremony
       (30 sec)  (wait)   (per question)  (end of round) (other    (scores)
                                                          teams)
```

**Arc narrative:**
Priya and Jordan start uncertain — opening a URL at a pub with social pressure to be ready quickly. Once they create a team and see the lobby confirmation, relief kicks in ("we're in"). As questions appear, engagement deepens — this is the core experience, and the answer form becomes the team's shared focus. Submission creates a moment of focused commitment, not panic. The waiting phase is collective anticipation as teams watch each other submit. The ceremony is the emotional payoff — excitement, reactions, shared experience.

---

## Journey Flow Diagram

```
  PLAYER opens /play on phone
         |
         |--- [localStorage has team_id + device_token] ---> AUTO-REJOIN
         |
         v (no stored identity)
  +------+-----------+
  |  SCREEN 1: JOIN  |   Team name entry
  |  Enter team name |
  +------+-----------+
         |
         | [Name submitted]
         |
         |--- [DUPLICATE_TEAM_NAME error] ---> inline error, retry
         |
         v
  +------+-----------+
  |  SCREEN 3: LOBBY |   Waiting for game to start
  |  See other teams |
  +------+-----------+
         |
         ^-- AUTO-REJOIN also lands here if game is in LOBBY state
         |
         | [host sends round_started]
         v
  +------+-----------+
  |  SCREEN 4:       |   <------ question_revealed events
  |  ROUND ACTIVE    |           (accumulating list)
  |  Q1 appears      |
  |  Enter answer    |
  +------+-----------+
         |
         | [More questions revealed by quizmaster]
         v
  +------+-----------+
  |  SCREEN 5:       |   All revealed questions visible + editable
  |  ANSWER FORM     |   draft_answer sent on each keystroke/blur
  |  Q1..Qn          |   localStorage draft updated continuously
  +------+-----------+
         |
         | [host sends round_ended — or all questions revealed]
         v
  +------+-----------+
  |  SCREEN 8:       |   Pre-submission review
  |  END-OF-ROUND    |   Blank answers flagged with warning
  |  REVIEW          |   "Go Back & Edit" or "Submit Answers"
  +------+-----------+
         |
         | [player taps Submit Answers]
         v
  +------+-----------+
  |  SCREEN 9:       |   Confirmation dialog
  |  CONFIRM SUBMIT  |   "This cannot be undone"
  |  DIALOG          |   [ Go Back ] [ Yes, Submit ]
  +------+-----------+
         |
         | [player confirms]
         v
  +------+-----------+
  |  SCREEN 10:      |   Fields become read-only
  |  POST-SUBMIT     |   Team submission status list (live)
  |  WAITING         |   "Waiting for other teams..."
  +------+-----------+
         |
         | [host sends ceremony_question_shown / ceremony_answer_revealed]
         v
  +------+-----------+
  |  SCREEN 11:      |   Current ceremony question + answer visible
  |  CEREMONY VIEW   |   Which teams got it right
  |  (read-only)     |
  +------+-----------+
         |
         | [host sends round_scores_published]
         v
  +------+-----------+
  |  SCREEN 12:      |   Round score + running total
  |  SCORES          |   All teams ranked
  +------+-----------+
         |
         | [Next round — round_started] --> back to ROUND ACTIVE
         | [Game over — game_over]
         v
  +------+-----------+
  |  SCREEN 12:      |   Final scores, winner highlighted
  |  FINAL SCORES    |
  +------------------+

  --- ERROR PATHS ---

  AUTO-REJOIN path:
  +------+-----------+
  |  SCREEN 2:       |   "Welcome back, Team Awesome!"
  |  AUTO-REJOIN     |   Auto-reconnects to current game state
  +------+-----------+
         |
         | [If state = ROUND_ACTIVE] -> restored to ANSWER FORM (Q1..Qn)
         | [If state = ROUND_ENDED / SCORING] -> post-submit waiting
         | [If state = CEREMONY] -> ceremony view
         | [TEAM_NOT_FOUND error] -> falls back to JOIN screen

  CONNECTION LOST path:
  Any screen ---> "Reconnecting..." banner overlaid
              ---> auto-reconnects via WsClient backoff
              ---> state_snapshot received ---> UI restores current screen

  JOINING GAME IN PROGRESS:
  New player joins when state = ROUND_ACTIVE
  ---> state_snapshot includes revealed questions
  ---> Lands on ANSWER FORM with all currently-revealed questions visible
  ---> No answers yet (new team, starting blank)
```

---

## Screen Mockups (Mobile-First)

### Screen 1: Join — First Visit

```
+----------------------------------+
|  TRIVIA NIGHT          [ /play ] |
+----------------------------------+
|                                  |
|  Welcome!                        |
|                                  |
|  Enter your team name:           |
|  +----------------------------+  |
|  |                            |  |
|  +----------------------------+  |
|                                  |
|  [ Join Game ]                   |
|                                  |
|  ──────────────────────────────  |
|  Use the same browser each       |
|  round to stay connected.        |
|                                  |
+----------------------------------+
```

**Emotional target:** Uncertain → Welcomed
**Emotional design levers:** One field, one button. No account, no password. Minimal cognitive load at a high-energy social moment.
**JTBD:** J1 (join quickly)
**Error state:** Name collision → inline red text "That name is taken — try another" appears below the input, input retains focus.

---

### Screen 2: Auto-Rejoin (Return Visit)

```
+----------------------------------+
|  TRIVIA NIGHT          [ /play ] |
+----------------------------------+
|                                  |
|  Welcome back,                   |
|  Team Awesome!                   |
|                                  |
|  Rejoining the game...           |
|  ████████████████████ 100%       |
|                                  |
|  [ That's not us — change team ] |
|                                  |
+----------------------------------+
```

**Emotional target:** Anxious → Relieved
**Emotional design levers:** Personalised greeting ("Team Awesome") confirms recognition before any delay. Progress bar gives feedback. Escape hatch is present but subordinate.
**JTBD:** J2 (rejoin seamlessly)
**Draft restore:** After rejoin completes, the answer form re-renders with localStorage draft answers restored.

---

### Screen 3: Lobby — Waiting for Game

```
+----------------------------------+
|  TRIVIA NIGHT          [ /play ] |
+----------------------------------+
|                                  |
|  Team Awesome                    |
|  You're in!                      |
|                                  |
|  ──────────────────────────────  |
|  Waiting for Marcus to           |
|  start the game...               |
|                                  |
|  Teams here:                     |
|  • Team Awesome (you)            |
|  • The Brainiacs                 |
|  • Quiz Killers                  |
|                                  |
|  Get ready!                      |
|                                  |
+----------------------------------+
```

**Emotional target:** Relieved → Engaged (anticipation building)
**Emotional design levers:** "You're in!" confirms success. Other teams visible creates shared anticipation. Quizmaster's name personalises the wait.
**JTBD:** J1 (join outcome), J6 (orientation)
**Shared artifact:** team_name from team_register response; team list from state_snapshot.

---

### Screen 4: Round Active — First Question Revealed

```
+----------------------------------+
|  Round 1: General Knowledge      |
|  ■ 1 of 8 questions revealed     |
+----------------------------------+
|                                  |
|  Q1  What is the capital         |
|      of France?                  |
|                                  |
|  Your answer:                    |
|  +----------------------------+  |
|  |                            |  |
|  +----------------------------+  |
|                                  |
|  ──────────────────────────────  |
|  Q2–Q8 not yet revealed.         |
|  Answers saved as you type.      |
|                                  |
+----------------------------------+
```

**Emotional target:** Engaged → Focused
**Emotional design levers:** Round name and progress counter set context instantly. Single question dominant. Draft-save hint removes anxiety about losing work.
**JTBD:** J3 (capture answers), J6 (orientation)
**Shared artifact:** question.text from question_revealed event. revealed_count / total_questions from event payload.

---

### Screen 5: Multiple Questions Revealed

```
+----------------------------------+
|  Round 1: General Knowledge      |
|  ■■■ 3 of 8 questions revealed   |
+----------------------------------+
|                                  |
|  Q1  What is the capital         |
|      of France?                  |
|  +----------------------------+  |
|  | Paris                      |  |
|  +----------------------------+  |
|                                  |
|  Q2  Name the three primary      |
|      colors.                     |
|  +----------------------------+  |
|  | Red, Blue               ▼  |  |
|  +----------------------------+  |
|  [ + add part ]                  |
|                                  |
|  Q3  [ IMAGE: eiffel.jpg ]       |
|      Name this landmark.         |
|  +----------------------------+  |
|  |                            |  |
|  +----------------------------+  |
|                                  |
|  Q4–Q8 not yet revealed.         |
|                                  |
+----------------------------------+
```

**Emotional target:** Focused (team discussion, collaborative entry)
**Emotional design levers:** All revealed questions remain visible and editable — no forced sequential progression. Growing progress indicator shows advancement. Multi-part answer "add part" control matches the numbered blanks on paper.
**JTBD:** J3 (answer capture), J7 (question types)
**Shared artifact:** revealed_questions array (accumulated from question_revealed events). draft_answers from localStorage per question_index.

---

### Screen 6: Multiple Choice Question

```
+----------------------------------+
|  Round 1: General Knowledge      |
|  ■■■■ 4 of 8 questions revealed  |
+----------------------------------+
|                                  |
|  Q4  Which planet is closest     |
|      to the sun?                 |
|                                  |
|  (  ) Venus                      |
|  ( * ) Mercury                   |
|  (  ) Mars                       |
|  (  ) Earth                      |
|                                  |
+----------------------------------+
```

**Emotional target:** Focused (low friction for a known-answer type)
**Emotional design levers:** Radio buttons match a universal mental model. No typing required. Selection is visually immediate.
**JTBD:** J7 (question types)
**Shared artifact:** question.choices from question_revealed event. Selection stored as draft_answer.

---

### Screen 7: Media Question (Image)

```
+----------------------------------+
|  Round 1: General Knowledge      |
|  ■■■■■ 5 of 8 questions revealed |
+----------------------------------+
|                                  |
|  Q5                              |
|  +----------------------------+  |
|  |  [  IMAGE: landmark.jpg  ] |  |
|  |  (tap to expand)           |  |
|  +----------------------------+  |
|      Name this landmark.         |
|                                  |
|  +----------------------------+  |
|  |                            |  |
|  +----------------------------+  |
|                                  |
+----------------------------------+
```

**Emotional target:** Focused (image provides visual anchor for team discussion)
**Error state:** Image fails → "Media unavailable — ask the quizmaster" replaces image block.
**JTBD:** J7 (question types)
**Shared artifact:** question.media.url from question_revealed event.

---

### Screen 8: End of Round — Review Screen

```
+----------------------------------+
|  Round 1: Submit Answers         |
+----------------------------------+
|                                  |
|  Review your answers:            |
|                                  |
|  Q1  Paris                       |
|  Q2  Red, Blue, Yellow           |
|  Q3  Eiffel Tower                |
|  Q4  Mercury                     |
|  Q5  (no answer)  ⚠              |
|  Q6  Michael Jackson             |
|  Q7  1969                        |
|  Q8  Brazil                      |
|                                  |
|  ⚠ Q5 has no answer.             |
|    You can still go back.        |
|                                  |
|  [ Go Back & Edit ]              |
|  [ Submit Answers ]              |
|                                  |
+----------------------------------+
```

**Emotional target:** Focused → Confident (review builds assurance)
**Emotional design levers:** Blank answer warning is advisory, not blocking. "Go Back & Edit" is available but does not hide "Submit Answers." The blank is honest and mirrors real trivia ("we just didn't know that one").
**JTBD:** J4 (submit with confidence)

---

### Screen 9: Submission Confirmation Dialog

```
+----------------------------------+
|  Round 1: Submit Answers         |
+----------------------------------+
|                                  |
|  +------------------------------+|
|  | Are you sure?               ||
|  |                             ||
|  | You have 1 unanswered       ||
|  | question (Q5).              ||
|  |                             ||
|  | Once submitted, you cannot  ||
|  | change your answers.        ||
|  |                             ||
|  | [ Go Back ]  [ Yes, Submit ]||
|  +------------------------------+|
|                                  |
+----------------------------------+
```

**Emotional target:** Confident (irreversibility made explicit, removes regret)
**JTBD:** J4 (submit with confidence)
**Domain note:** Submit sends submit_answers message with all answers. Server sends submission_ack with locked: true. After ack, all answer fields become read-only.

---

### Screen 10: Post-Submit Waiting

```
+----------------------------------+
|  Round 1: Submitted!             |
+----------------------------------+
|                                  |
|  Your answers are locked in.     |
|                                  |
|  Waiting for other teams...      |
|                                  |
|  Team Awesome      [submitted]   |
|  The Brainiacs     [submitted]   |
|  Quiz Killers      [waiting...]  |
|                                  |
|  The ceremony will begin soon.   |
|                                  |
+----------------------------------+
```

**Emotional target:** Confident → Anticipation
**Emotional design levers:** "Locked in" is assertive, not scary. Other teams' status turns a potentially anxious wait into shared anticipation — watching the last team tick over.
**JTBD:** J5 (ceremony), J6 (orientation)
**Shared artifact:** submission_received events update other teams' status in real time.

---

### Screen 11: Ceremony View

```
+----------------------------------+
|  Round 1: Ceremony               |
+----------------------------------+
|                                  |
|  Q3  Name this landmark.         |
|                                  |
|  Answer: Eiffel Tower            |
|                                  |
|  Team Awesome     ✓ got it       |
|  The Brainiacs    ✓ got it       |
|  Quiz Killers     ✗ missed it    |
|                                  |
|  ──────────────────────────────  |
|  Waiting for quizmaster to       |
|  reveal the next answer...       |
|                                  |
+----------------------------------+
```

**Emotional target:** Anticipation → Elated (or "we knew it!" vs "oh no")
**Emotional design levers:** Revealed answer is bold and prominent. Team results are clear checkmarks/crosses. Waiting message prevents confusion when quizmaster pauses.
**JTBD:** J5 (ceremony)
**Shared artifact:** ceremony_answer_revealed event (question_index, answer). Scoring verdicts from scored_answers.

---

### Screen 12: Round Scores

```
+----------------------------------+
|  Round 1: Scores                 |
+----------------------------------+
|                                  |
|  Round 1 Results                 |
|                                  |
|  1. Team Awesome    6 pts  (+6)  |
|  2. The Brainiacs   5 pts  (+5)  |
|  3. Quiz Killers    3 pts  (+3)  |
|                                  |
|  Your team: Team Awesome  6 pts  |
|                                  |
|  ──────────────────────────────  |
|  Round 2 starts soon...          |
|                                  |
+----------------------------------+
```

**Emotional target:** Elated / Proud (or motivated to do better next round)
**JTBD:** J5 (ceremony payoff)
**Shared artifact:** round_scores_published event payload. team_name from stored identity.

---

### Connection Status Banner (Overlay — any screen)

```
+----------------------------------+
|  ⚠ Reconnecting...              |
+----------------------------------+
```

```
+----------------------------------+
|  ⚠ Connection lost. Tap to retry.|
+----------------------------------+
```

**JTBD:** J6 (orientation)
**Trigger:** WsClient fires reconnect_failed; banner replaces status indicator.

---

## Error Paths

| Error | Detection Point | Player Experience |
|-------|----------------|------------------|
| Team name taken | Join → server sends DUPLICATE_TEAM_NAME | Inline: "That name is taken — try another." Input retains focus. |
| TEAM_NOT_FOUND on rejoin | Auto-rejoin → server sends TEAM_NOT_FOUND | Falls back to join screen. No localStorage data cleared yet. |
| Game not started | Lobby | "Waiting for Marcus to start the game..." (expected idle state) |
| Browser refresh mid-round | Auto-rejoin flow | state_snapshot received; draft answers restored from localStorage |
| Connection lost | Any screen | "Reconnecting..." overlay banner. Auto-reconnects via WsClient exponential backoff. |
| Attempt to edit after submit | Post-submit screen | Fields are read-only; "Answers are locked" reminder if user taps a field. |
| Joining game in progress | Join → ROUND_ACTIVE state_snapshot | Lands on Answer Form with all currently-revealed questions pre-populated (no answers yet). |
| Image/audio/video fails | Media question display | "Media unavailable — ask the quizmaster" inline in the media block. |
| INVALID_STATE_TRANSITION | Any action in wrong state | Transient error banner: "Action not available right now." Clears after 3 seconds. |

---

## Shared Artifacts Produced/Consumed

| Artifact | Direction | Source | Consumers |
|----------|-----------|--------|-----------|
| team_id | Produces | team_register → team_registered response | localStorage, submit_answers, rejoin |
| device_token | Produces | team_registered response | localStorage, team_rejoin message |
| team_name | Produces | player input on join | Lobby display, post-submit, scores |
| draft_answers[round][q_index] | Produces | player typing → localStorage | Answer form restore on rejoin |
| revealed_questions | Consumes | question_revealed events | Answer form question list |
| submission_status | Consumes | submission_received events | Post-submit waiting screen |
| ceremony question + answer | Consumes | ceremony_question_shown + ceremony_answer_revealed | Ceremony view |
| round_scores | Consumes | round_scores_published event | Scores screen |
| game_state | Consumes | state_snapshot on connect/reconnect | Screen routing logic |
