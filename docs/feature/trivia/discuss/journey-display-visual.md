# Journey Visual -- Display Interface (/display)

## Metadata

- Feature ID: trivia
- Interface: /display (Public Read-Only)
- Audience: All room participants (via TV/projector), controlled by Marcus
- Date: 2026-03-29
- Artifact type: ASCII flow + TUI mockups + emotional annotations

---

## Role of /display

The display interface is not interacted with directly by any user. It is:
1. Opened by Marcus on the room TV (cast, HDMI, or second monitor)
2. Driven by Marcus's actions on /host (current question state, ceremony steps)
3. Seen by everyone in the room -- players, spectators, Marcus himself

It has no controls, no answer entry, no private information. Its job is to be a **shared focal point for the room**.

---

## Emotional Arc (Room Perspective)

```
EMOTION
  ^
  |
  | ANTICIPATION     CURIOSITY         TENSION            DELIGHT
  | (waiting)        (question shown)  (answer pending)   (correct revealed)
  |
  |   [1]               [2][3]             [4]               [5][6]
  |    *                   *                 *                  *
  |     \                 / \               / \                /
  |      *               *   *             *   *              *
  |       \             /     \           /     \            /
  |        *-----------*       *---------*       *----------*
  |
  +-------------------------------------------------------------------> TIME
       Pre-game      Question shown    Teams answering   Ceremony
       (lobby)       (reveal pacing)   (submit wait)     (answer reveal)
```

**Arc narrative:** The display creates the shared emotional pulse for the room. When a question appears, everyone looks up. When the correct answer is revealed during the ceremony, there is an audible reaction. The display must be legible at distance (TV resolution), uncluttered (no UI noise), and always in sync with the quizmaster's pacing.

---

## States of the Display

The display has distinct visual modes mapped to game states:

| Game State | Display Mode | Content |
|------------|-------------|---------|
| Pre-game (lobby) | Holding screen | Quiz title + "Game starting soon" |
| Round active | Question view | Current question text (+ media if applicable) |
| Round ended, awaiting submission | Waiting screen | "Round N complete -- submitting answers..." |
| Ceremony | Answer reveal | Question + correct answer (one at a time) |
| Between rounds | Scores screen | Round scores, running totals |
| Game over | Final screen | Winner announcement + full scoreboard |

---

## TUI Mockups (TV/Large Screen)

### State 1: Holding Screen (Pre-Game)

```
+----------------------------------------------------------------+
|                                                                |
|                                                                |
|            FRIDAY NIGHT TRIVIA                                 |
|            March 2026                                          |
|                                                                |
|                                                                |
|               Game starting soon...                           |
|                                                                |
|                                                                |
|         Join: http://trivia.local/play                        |
|                                                                |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Clean and welcoming. Players who just arrived can see the join URL on the big screen. No quizmaster controls, no scores, no hint of what's coming. Building anticipation.

---

### State 2: Question View -- Text Only

```
+----------------------------------------------------------------+
|  Round 1: General Knowledge          Question 3 of 8          |
+----------------------------------------------------------------+
|                                                                |
|                                                                |
|                                                                |
|    Name this song and artist.                                 |
|                                                                |
|                                                                |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Question text is large and centered. Round name and question counter in the header give context without clutter. For text-only questions, maximum whitespace maximizes legibility from across a room.

---

### State 3: Question View -- With Image

```
+----------------------------------------------------------------+
|  Round 1: General Knowledge          Question 3 of 8          |
+----------------------------------------------------------------+
|                                                                |
|      +-------------------------------+                        |
|      |                               |                        |
|      |      [LANDMARK IMAGE]         |                        |
|      |                               |                        |
|      +-------------------------------+                        |
|                                                                |
|    Name this landmark.                                        |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Image is given maximum space -- the point is for everyone to see it clearly. Question text below is still large. Nothing else on screen.

---

### State 4: Question View -- Audio

```
+----------------------------------------------------------------+
|  Round 2: Music                       Question 1 of 6         |
+----------------------------------------------------------------+
|                                                                |
|                                                                |
|         ♪  Now playing...                                     |
|                                                                |
|    Name this song and artist.                                 |
|                                                                |
|         [audio plays automatically when revealed]             |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Audio questions don't need a waveform or player chrome -- the audio plays automatically when Marcus reveals the question. A simple "Now playing..." indicator confirms the audio is active without distracting from the listening experience.

---

### State 5: Question View -- Multiple Choice

```
+----------------------------------------------------------------+
|  Round 1: General Knowledge          Question 4 of 8          |
+----------------------------------------------------------------+
|                                                                |
|    Which planet is closest to the sun?                        |
|                                                                |
|       A)  Venus                                               |
|       B)  Mercury                                             |
|       C)  Mars                                                |
|       D)  Earth                                               |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Choices are displayed on the TV so players can discuss them. The choices are lettered (A/B/C/D) matching what players see on their phones. Players can call out "I think B" in conversation.

---

### State 6: Waiting for Submissions

```
+----------------------------------------------------------------+
|  Round 1: General Knowledge          All questions revealed   |
+----------------------------------------------------------------+
|                                                                |
|                                                                |
|             Submitting answers...                             |
|                                                                |
|             3 teams connected                                 |
|             2 of 3 have submitted                             |
|                                                                |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** The room can see how many teams have submitted. This creates gentle social pressure on the last team. No team names are shown (to avoid embarrassment), just a count.

---

### State 7: Answer Ceremony -- Question + Answer Reveal

```
+----------------------------------------------------------------+
|  Round 1 -- Answers                   Question 3 of 8         |
+----------------------------------------------------------------+
|                                                                |
|    Name this landmark.                                        |
|                                                                |
|      +-------------------------------+                        |
|      |                               |                        |
|      |      [LANDMARK IMAGE]         |                        |
|      |                               |                        |
|      +-------------------------------+                        |
|                                                                |
|    Answer:  Eiffel Tower                                      |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** The correct answer appears below the question (and image if applicable). The reveal is a visual moment -- teams cheer or groan. Marcus controls when this appears. The display does not show which teams got it right/wrong (that information is on Marcus's screen, which he can announce verbally for theatrical effect).

---

### State 8: Round Scores

```
+----------------------------------------------------------------+
|  Round 1: Complete                                             |
+----------------------------------------------------------------+
|                                                                |
|  Round 1 Scores                                               |
|  ─────────────────────────────────────────────────────────    |
|                                                                |
|  1.  Team Awesome      8 points      (Running: 8)             |
|  2.  The Brainiacs     6 points      (Running: 6)             |
|  3.  Quiz Killers      4 points      (Running: 4)             |
|                                                                |
|  ─────────────────────────────────────────────────────────    |
|  Round 2 starts shortly...                                    |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Scores are announced in rank order. Running total shows the overall standing. This screen stays up between rounds -- it gives the room something to look at while teams prepare mentally for the next round.

---

### State 9: Final Scores -- Game Over

```
+----------------------------------------------------------------+
|  TRIVIA NIGHT COMPLETE                                         |
+----------------------------------------------------------------+
|                                                                |
|  Final Standings                                               |
|  ─────────────────────────────────────────────────────────    |
|                                                                |
|  1.  Team Awesome      32 points      WINNER!                 |
|  2.  The Brainiacs     27 points                              |
|  3.  Quiz Killers      19 points                              |
|                                                                |
|  ─────────────────────────────────────────────────────────    |
|  Thanks for playing!                                          |
|                                                                |
+----------------------------------------------------------------+
```

**Emotional note:** Final screen is celebratory but simple. "WINNER!" next to the top team is the only decoration needed. The room does the rest with noise.

---

## Design Constraints for /display

1. **No controls on screen** -- no buttons, no forms, no navigation. Any accidental touch does nothing.
2. **Auto-syncs to game state** -- WebSocket-driven; display always reflects the current authoritative state from /host.
3. **Large, high-contrast text** -- legible from 5+ meters at TV resolution (1080p minimum)
4. **No private information** -- upcoming questions are never visible; quizmaster scores and answer comparisons are never visible
5. **Graceful on disconnect** -- if WebSocket drops, display shows "Reconnecting..." and auto-reconnects without intervention
6. **Audio plays on the host device** -- for audio questions, the display is cast from Marcus's device, so audio plays through the casting device's speakers naturally

---

## Shared Artifacts Consumed

| Artifact | Source | When Used |
|----------|--------|-----------|
| Quiz title | /host (YAML load) | Holding screen |
| Current question | /host (reveal action) | Question view |
| Question media (image/audio/video) | YAML file (served locally) | Media question view |
| Multiple choice options | YAML file | MC question view |
| Submission count | /play (submission events) | Waiting screen |
| Correct answers | /host (ceremony trigger) | Ceremony state |
| Round scores | /host (scoring complete) | Scores screen |
| Final scores | /host (game over) | Final screen |

---

## Error Paths

| Error | Detection Point | User Experience |
|-------|----------------|-----------------|
| WebSocket connection lost | Any state | "Reconnecting..." overlay; auto-reconnects |
| Media file fails to load | Question state | Placeholder shown; quizmaster can read question aloud |
| Display opened before game starts | Pre-game | Holding screen shown; waits for game start event |
| Display opened mid-game | Mid-round | Catches up to current state immediately |
