# Journey Visual -- Host Interface (/host)

## Metadata

- Feature ID: trivia
- Interface: /host (Quizmaster)
- Persona: Marcus Okafor
- Date: 2026-03-29
- Artifact type: ASCII flow + TUI mockups + emotional annotations

---

## Emotional Arc

```
EMOTION
  ^
  |
  | ANXIOUS          FOCUSED              IN CONTROL         ACCOMPLISHED
  | (setup)          (reveal)             (scoring)          (ceremony)
  |
  |   [1]              [2][3]               [4][5]              [6]
  |    *                 *  *                 *  *                *
  |     \               /    \               /    \              /
  |      *             *      *             *      *            *
  |       \           /        \           /        \          /
  |        *---------*          *---------*          *--------*
  |
  +-------------------------------------------------------------------> TIME
       Setup        Reveal         Submission        Scoring   Ceremony
       (2-3 min)    (per Q)        (wait)            (3 min)   (5 min)
```

**Arc narrative:** Marcus starts slightly anxious -- will the setup work? Will everyone connect? Once the first question is revealed and players respond, anxiety dissolves into focus. Scoring is ergonomic and efficient, building a sense of control. The ceremony is the payoff -- showman mode.

---

## Journey Flow Diagram

```
  MARCUS arrives at venue
         |
         v
  +------+----------+
  |  LOAD QUIZ      |   /host page
  |  Upload YAML    |
  |  Validate file  |
  +------+----------+
         |
         | [YAML valid]        [YAML invalid]
         |                           |
         v                           v
  +------+----------+     +----------+--------+
  |  GAME LOBBY     |     |  VALIDATION ERROR  |
  |  Share /play URL|     |  Field-level msgs  |
  |  Wait for teams |     |  Fix & reload      |
  +------+----------+     +--------------------+
         |
         | [Start Game]
         v
  +------+----------+
  |  ROUND ACTIVE   |  <--------------------------------------------+
  |  Reveal Q1      |                                               |
  |  [Reveal Next Q]|                                               |
  |  See Q count    |                                               |
  +------+----------+                                               |
         |                                                          |
         | [All questions revealed]                                 |
         v                                                          |
  +------+----------+                                               |
  |  END ROUND      |                                               |
  |  See submitted  |                                               |
  |  teams (live)   |                                               |
  +------+----------+                                               |
         |                                                          |
         | [All submitted OR timeout]                               |
         v                                                          |
  +------+----------+                                               |
  |  SCORING        |                                               |
  |  Q-by-Q grid    |                                               |
  |  Mark right/    |                                               |
  |  wrong          |                                               |
  +------+----------+                                               |
         |                                                          |
         | [All marked]                                             |
         v                                                          |
  +------+----------+                                               |
  |  CEREMONY       |                                               |
  |  Drive display  |                                               |
  |  Step through   |                                               |
  |  answers        |                                               |
  +------+----------+                                               |
         |                                                          |
         | [Next Round?]     [Last Round?]                          |
         |  YES ---------------------------------------->-----------+
         |  NO
         v
  +------+----------+
  |  GAME OVER      |
  |  Final scores   |
  |  Winner!        |
  +------------------+
```

---

## TUI Mockups

### Screen 1: Load Quiz

```
+----------------------------------------------------------------+
|  TRIVIA NIGHT                                    [ /host ]     |
+----------------------------------------------------------------+
|                                                                |
|  Load Your Quiz                                                |
|  ─────────────────────────────────────────────────────────    |
|                                                                |
|  YAML file path:  [ /home/marcus/quizzes/march-2026.yaml  ]   |
|                                                                |
|  [ Load Quiz ]                                                 |
|                                                                |
|  ─────────────────────────────────────────────────────────    |
|  Or drag and drop a .yaml file here                           |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Plain, task-focused. No decorative chrome. Marcus just wants to load the file and get started.

---

### Screen 2: YAML Validation Error

```
+----------------------------------------------------------------+
|  TRIVIA NIGHT                                    [ /host ]     |
+----------------------------------------------------------------+
|                                                                |
|  Quiz Load Failed                                              |
|  ─────────────────────────────────────────────────────────    |
|                                                                |
|  ! Round 2, Question 3: missing required field "answer"       |
|  ! Round 3, Question 1: image file "logo.png" not found       |
|                                                                |
|  Fix these errors in your YAML file and reload.               |
|                                                                |
|  [ Load Different File ]                                       |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Errors are specific and actionable. Marcus can fix immediately. No vague "something went wrong."

---

### Screen 3: Game Lobby

```
+----------------------------------------------------------------+
|  TRIVIA NIGHT                                    [ /host ]     |
+----------------------------------------------------------------+
|                                                                |
|  Friday Night Trivia -- March 2026                            |
|  4 rounds  |  32 questions                                     |
|                                                                |
|  ─── Player URL ───────────────────────────────────────────   |
|  http://trivia.local/play                                      |
|  [ Copy Link ]                                                 |
|                                                                |
|  ─── Display URL ──────────────────────────────────────────   |
|  http://trivia.local/display                                   |
|  [ Copy Link ]                                                 |
|                                                                |
|  ─── Teams Connected ──────────────────────────────────────   |
|  Team Awesome    (3 players)    connected 0:42 ago             |
|  The Brainiacs   (2 players)    connected 0:18 ago             |
|  Quiz Killers    (4 players)    connected 0:05 ago             |
|                                                                |
|  Waiting for more teams...                                     |
|                                                                |
|  [ Start Game ]                                               |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Marcus can see all teams are present. The "Start Game" button only appears when Marcus is ready -- no auto-start pressure. He feels in control of the moment.

---

### Screen 4: Round Active -- Reveal Panel

```
+----------------------------------------------------------------+
|  TRIVIA NIGHT -- Round 1: General Knowledge      [ /host ]    |
+----------------------------------------------------------------+
|                                                                |
|  Questions revealed: 3 / 8                                    |
|                                                                |
|  Q1  What is the capital of France?              [revealed]   |
|  Q2  Name the three primary colors.              [revealed]   |
|  Q3  Name this landmark.  [image: eiffel.jpg]    [revealed]   |
|  Q4  ● ● ● ● ● ● ● ● ● ● ● ●  (hidden)                       |
|  Q5  ● ● ● ● ● ● ● ● ● ● ● ●  (hidden)                       |
|  Q6  ● ● ● ● ● ● ● ● ● ● ● ●  (hidden)                       |
|  Q7  ● ● ● ● ● ● ● ● ● ● ● ●  (hidden)                       |
|  Q8  ● ● ● ● ● ● ● ● ● ● ● ●  (hidden)                       |
|                                                                |
|  [ Reveal Q4 ]                          [ End Round Early ]   |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Revealed questions are clearly visible; upcoming questions are hidden but their count is known. Marcus can see progress and knows what's coming without players seeing it.

---

### Screen 5: End Round -- Submission Status

```
+----------------------------------------------------------------+
|  TRIVIA NIGHT -- Round 1 Complete                [ /host ]    |
+----------------------------------------------------------------+
|                                                                |
|  Submission Status                                             |
|  ─────────────────────────────────────────────────────────    |
|                                                                |
|  Team Awesome       [submitted 2:14 ago]                      |
|  The Brainiacs      [submitted 0:45 ago]                      |
|  Quiz Killers       [waiting...]                               |
|                                                                |
|  1 team has not yet submitted.                                |
|                                                                |
|  [ Open Scoring ]   (available when all submitted)            |
|  [ Open Scoring Anyway ]  (override -- use with caution)      |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Marcus can see exactly who is holding things up. The social pressure is visible. "Open Scoring Anyway" exists for real-world events where one team is distracted.

---

### Screen 6: Scoring Interface

```
+----------------------------------------------------------------+
|  TRIVIA NIGHT -- Round 1: Scoring                [ /host ]    |
+----------------------------------------------------------------+
|                                                                |
|  Q1  What is the capital of France?                           |
|      Answer: Paris                                            |
|  ─────────────────────────────────────────────────────────    |
|  Team Awesome     "paris"              [ correct ] [ wrong ]  |
|  The Brainiacs    "PARIS"              [ correct ] [ wrong ]  |
|  Quiz Killers     "Lyon"               [ correct ] [ wrong ]  |
|                                                                |
|  ─────────────────────────────────────────────────────────    |
|  Q2  Name the three primary colors.                           |
|      Answers: Red, Blue, Yellow  (any order)                  |
|  ─────────────────────────────────────────────────────────    |
|  Team Awesome     Red / Blue / Green   [ correct ] [ wrong ]  |
|  The Brainiacs    Yellow / Red / Blue  [ correct ] [ wrong ]  |
|  Quiz Killers     Red / Yellow / Blue  [ correct ] [ wrong ]  |
|                                                                |
|  Round 1 Scores: Awesome: 1  |  Brainiacs: 2  |  Killers: 0  |
|  Running Total:  Awesome: 1  |  Brainiacs: 2  |  Killers: 0  |
|                                                                |
|  [ Save Scores & Start Ceremony ]                             |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Side-by-side layout. Marcus's eyes move left-to-right: question, expected, team answer, click. No scrolling per team; all teams visible per question. Running scores auto-update. The job is to click through quickly, and the layout supports that.

---

### Screen 7: Ceremony Control

```
+----------------------------------------------------------------+
|  TRIVIA NIGHT -- Round 1: Answer Ceremony        [ /host ]    |
+----------------------------------------------------------------+
|                                                                |
|  Driving the display at:  http://trivia.local/display         |
|                                                                |
|  Currently showing:   Q3 -- Name this landmark.              |
|  Correct answer:      Eiffel Tower                            |
|                                                                |
|  Who got it:  Team Awesome [YES]  Brainiacs [YES]  Killers [NO]|
|                                                                |
|  ─────────────────────────────────────────────────────────    |
|  [ Previous Answer ]              [ Next Answer ]             |
|                                                                |
|  Progress:  3 of 8 answers revealed                           |
|                                                                |
|  [ End Ceremony & Show Scores ]                               |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Marcus controls the room from this screen. He reveals answers at his own pace, building tension. The "who got it" panel lets him call out teams by name, adding showmanship.

---

## Shared Artifacts Produced/Consumed

| Artifact | Direction | Used By |
|----------|-----------|---------|
| Game session ID | Produces | /play, /display |
| Quiz content (from YAML) | Produces | /play, /display |
| Current question state | Produces | /play, /display |
| Team registry | Consumes | Built by /play join flow |
| Submitted answers | Consumes | Built by /play submission |
| Scores | Produces | /display (ceremony) |

---

## Error Paths

| Error | Detection Point | User Experience |
|-------|----------------|-----------------|
| YAML file not found | Load screen | Specific error: "File not found: quiz.yaml" |
| YAML missing required field | Load screen | "Round 2, Q3: missing answer field" |
| Media file missing | Load screen | "Round 1, Q4: image.jpg not found" |
| No teams connected | Lobby | Warning: "No teams yet -- share the player URL" |
| Team disconnects mid-game | Reveal panel | Status dot goes grey; reconnects auto |
| All teams disconnect | Any screen | Warning banner; game pauses |
| Team submits after scoring opens | Scoring | Late submission notification; quizmaster can include |
