# Journey Visual -- Player Interface (/play)

## Metadata

- Feature ID: trivia
- Interface: /play (Player / Team)
- Personas: Priya Nair (team captain), Jordan Kim (casual player)
- Date: 2026-03-29
- Artifact type: ASCII flow + TUI mockups + emotional annotations

---

## Emotional Arc

```
EMOTION
  ^
  |
  | UNCERTAIN       ENGAGED           FOCUSED            RELIEVED
  | (joining)       (answering)       (submitting)       (waiting)
  |
  |   [1]              [2][3]            [4][5]              [6]
  |    *                  *                  *                 *
  |     \                / \                / \               /
  |      *              *   *              *   *             *
  |       \            /     \            /     \           /
  |        *----------*       *----------*       *---------*
  |
  +-------------------------------------------------------------------> TIME
       Join         Answer Q1..N      Review & Submit   See scores
       (30 sec)     (per question)    (end of round)    (ceremony)
```

**Arc narrative:** Priya and Jordan start slightly uncertain -- new URL, new app, what do I do? Once team creation succeeds and they see questions appearing, engagement kicks in. The answer form becomes the team's shared work surface. Submitting feels like a commitment; confirmation dialog validates their confidence. Waiting for ceremony is anticipation.

---

## Journey Flow Diagram

```
  PLAYER opens /play on phone
         |
         v
  +------+----------+
  |  JOIN SCREEN    |   Team name entry
  |  Enter team     |   (first time only)
  |  name           |
  +------+----------+
         |
         | [Name entered]         [Return visit / refresh]
         |                               |
         v                               v
  +------+----------+          +---------+--------+
  |  LOBBY          |          | AUTO-REJOIN       |
  |  Waiting for    |          | "Welcome back,    |
  |  game to start  |          |  Team Awesome!"   |
  +------+----------+          +---------+--------+
         |                               |
         +---------------+---------------+
                         |
                         v
  +------+----------+                                          |
  |  ROUND ACTIVE   |  <---------------------------------------+
  |  Q1 appears     |
  |  Enter answer   |
  +------+----------+
         |
         | [Q2, Q3... revealed by quizmaster]
         v
  +------+----------+
  |  ANSWER FORM    |
  |  Q1  [____]     |
  |  Q2  [____]     |
  |  Q3  [____]     |
  |  (all editable) |
  +------+----------+
         |
         | [End of round -- all questions revealed]
         v
  +------+----------+
  |  REVIEW &       |
  |  SUBMIT         |
  |  Confirmation   |
  |  dialog         |
  +------+----------+
         |
         | [Confirmed]
         v
  +------+----------+
  |  WAITING        |
  |  Submitted!     |
  |  Waiting for    |
  |  other teams    |
  +------+----------+
         |
         | [Ceremony starts]
         v
  +------+----------+
  |  CEREMONY VIEW  |
  |  See current    |
  |  round scores   |
  |  (read-only)    |
  +------+----------+
         |
         | [Next round]         [Game over]
         |  YES -----------------> Round Active (new round)
         |  NO
         v
  +------+----------+
  |  FINAL SCORES   |
  |  Winner!        |
  +------------------+
```

---

## TUI Mockups (Mobile-First)

### Screen 1: Join -- First Visit

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
|  Join from the same browser to   |
|  stay connected.                 |
|                                  |
+----------------------------------+
```

**Emotional note:** One field, one button. No account creation, no email, no password. Minimal cognitive load at a moment when the social energy in the room is high. Players want to join and get in, not fill out a form.

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
|  [ That's not us -- change team ]|
|                                  |
+----------------------------------+
```

**Emotional note:** Jordan's anxiety ("everything was gone") is addressed directly. The app says "welcome back" by name and rejoins automatically. The escape hatch ("change team") is present but not prominent.

---

### Screen 3: Lobby -- Waiting for Game

```
+----------------------------------+
|  TRIVIA NIGHT          [ /play ] |
+----------------------------------+
|                                  |
|  Team Awesome                    |
|  3 players connected             |
|                                  |
|  ──────────────────────────────  |
|  Waiting for Marcus to           |
|  start the game...               |
|                                  |
|  Other teams:                    |
|  - The Brainiacs                 |
|  - Quiz Killers                  |
|                                  |
|  Get ready!                      |
|                                  |
+----------------------------------+
```

**Emotional note:** Players can see other teams are also waiting, which builds a sense of shared anticipation. The quizmaster's name ("Marcus") personalizes the wait. No clock, no pressure -- just readiness.

---

### Screen 4: Round Active -- Question Revealed

```
+----------------------------------+
|  Round 1: General Knowledge      |
+----------------------------------+
|  Q1 of 8 revealed                |
|                                  |
|  What is the capital of France?  |
|                                  |
|  Your answer:                    |
|  +----------------------------+  |
|  | Paris                      |  |
|  +----------------------------+  |
|                                  |
|  ──────────────────────────────  |
|  Q2, Q3... not yet revealed      |
|  (waiting for quizmaster)        |
|                                  |
+----------------------------------+
```

**Emotional note:** The current question is large and readable. Answer field is immediately below, no scrolling. The "Q2, Q3... not yet revealed" line sets expectation -- there are more questions coming, and the quizmaster controls the pace.

---

### Screen 5: Multiple Questions Revealed

```
+----------------------------------+
|  Round 1: General Knowledge      |
+----------------------------------+
|  3 of 8 questions revealed       |
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
|  Q3  [ IMAGE: landmark photo ]   |
|      Name this landmark.         |
|  +----------------------------+  |
|  |                            |  |
|  +----------------------------+  |
|                                  |
|  Q4-8 not yet revealed.          |
|                                  |
+----------------------------------+
```

**Emotional note:** As more questions are revealed, the answer sheet grows. All previously revealed questions stay editable. The team can discuss Q1 while Q3 is being revealed. Multi-part answers have expandable fields. Image appears inline above the question text.

---

### Screen 6: Multiple Choice Question

```
+----------------------------------+
|  Round 1: General Knowledge      |
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

**Emotional note:** Radio button style for multiple choice. No text input needed. Tapping an option selects it, and the selection is visually clear. Consistent with how every app the player has used before works.

---

### Screen 7: End of Round -- Submit Review

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
|  Q5  (no answer)  !              |
|  Q6  Michael Jackson             |
|  Q7  1969                        |
|  Q8  Brazil                      |
|                                  |
|  ! Q5 has no answer entered.     |
|    You can still go back.        |
|                                  |
|  [ Go Back & Edit ]              |
|  [ Submit Answers ]              |
|                                  |
+----------------------------------+
```

**Emotional note:** Pre-submission review is a crucial trust moment. The app surfaces the unanswered question with a warning but does not block submission -- a blank answer is still a valid choice. This mirrors the real trivia experience ("we just didn't know that one").

---

### Screen 8: Submission Confirmation Dialog

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

**Emotional note:** Final gate. Priya's anxiety about accidental submission is resolved here. The dialog is honest about the consequence. "Yes, Submit" is primary but styled assertively, not casually -- this is a commitment.

---

### Screen 9: Post-Submit Waiting Screen

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
|  The answer ceremony will        |
|  begin shortly.                  |
|                                  |
+----------------------------------+
```

**Emotional note:** "Locked in" is confident language, not scary. Showing other teams' status turns waiting into shared anticipation -- players watch as the last team ticks over to submitted.

---

## Shared Artifacts Produced/Consumed

| Artifact | Direction | Used By |
|----------|-----------|---------|
| Team name + ID | Produces | /host (team registry), /display |
| Answer entries per question | Produces | /host (scoring) |
| Submission state per round | Produces | /host (submission status) |
| Current game state | Consumes | Driven by /host |
| Revealed questions | Consumes | Driven by /host |
| Scores (ceremony) | Consumes | Driven by /host ceremony |

---

## Error Paths

| Error | Detection Point | User Experience |
|-------|----------------|-----------------|
| Team name already taken | Join screen | "That name is taken -- try another" |
| Game not started | Lobby | "Waiting for Marcus to start..." |
| Browser refresh mid-round | Auto | Rejoin screen; answers restored from localStorage |
| Connection lost | Any active screen | "Reconnecting..." banner; auto-reconnects |
| Attempting to edit after submit | Post-submit | Fields are read-only; reminder: "Answers are locked" |
| Joining a game in progress | Join screen | Auto-caught up to current round; prior round history shown as submitted |
| Audio/video fails to play | Question view | "Media unavailable -- ask the quizmaster" |
