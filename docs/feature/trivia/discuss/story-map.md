# Story Map -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DISCUSS -- Phase 2.5 (User Story Mapping)
- Date: 2026-03-29
- Method: User Story Mapping (Patton) -- backbone → walking skeleton → release slices

---

## Story Map Backbone (User Activities -- Horizontal)

The backbone is the complete game event sequence, organized left to right as it happens in a real trivia night:

```
+------------+  +------------+  +------------+  +------------+  +------------+  +------------+
|  LOAD QUIZ |  |  ONBOARD   |  |  PLAY ROUND|  |  SCORE     |  |  CEREMONY  |  |  NEXT      |
|  & SETUP   |  |  PLAYERS   |  |  (repeat)  |  |  ANSWERS   |  |  & SCORES  |  |  ROUND /   |
|            |  |            |  |            |  |            |  |            |  |  END GAME  |
+------------+  +------------+  +------------+  +------------+  +------------+  +------------+
     A                B               C               D               E               F
```

---

## Story Map Full Detail

```
ACTIVITY   A: Load & Setup       B: Onboard Players    C: Play Round (x4)      D: Score Answers        E: Ceremony & Scores    F: Progress / End
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
USER       [Marcus]              [Marcus]              [Marcus]                [Marcus]                [Marcus]                [Marcus]
TASKS      [Players]             [Priya/Jordan]        [Priya/Jordan]          [Priya/Jordan]          [Priya/Jordan]          [Priya/Jordan]
           [Display]             [Display]             [Display]               [Display]               [Display]               [Display]
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

STORIES    US-01: Load YAML      US-04: Player joins   US-07: QM reveals Qs   US-12: QM scoring       US-15: Answer ceremony  US-18: Next round
           (validation errors)   (first visit)         (one at a time)         (side-by-side grid)     (driven from /host)     (return to C)

           US-02: Game lobby     US-05: Auto-rejoin    US-08: Player enters    US-13: Multi-part       US-16: Round scores     US-19: Final scores
           (share URLs)          (device persistence)  & edits answers         answer scoring          (on /display)           (on /display)
                                                       (all revealed)
           US-03: Start game     US-06: Display        US-09: Player submits   US-14: Auto-tally       US-17: Ceremony         US-20: Game over
           (broadcast to all)    holding screen        (confirm + lock)        scores                  scores view             screen
                                                                                                       (/play read-only)
                                                       US-10: QM monitors
                                                       submissions

                                                       US-11: Display shows
                                                       current question
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
```

---

## Walking Skeleton (Release 1)

The walking skeleton is the thinnest slice that produces an end-to-end working game. A real trivia night can be run with this slice, even if it is basic.

**Definition of walking skeleton for this project:**

> A quizmaster loads a quiz, players join and answer text-only questions, players submit at end of round, quizmaster scores answers, scores are displayed -- one complete round from start to finish.

**Walking Skeleton Stories:**

| Story | Title | Why it's in the skeleton |
|-------|-------|--------------------------|
| US-01 | Load YAML (text questions only) | Game cannot start without this |
| US-02 | Game lobby -- share URLs | Teams cannot join without this |
| US-03 | Start game -- broadcast to all | Game state synchronization foundation |
| US-04 | Player joins (first visit) | Players cannot participate without this |
| US-05 | Auto-rejoin on refresh | Without this, one accidental refresh breaks the game |
| US-07 | Quizmaster reveals questions | Core game loop start |
| US-08 | Player enters and edits answers | Core player interaction |
| US-09 | Player submits answers (with confirm) | Round cannot complete without this |
| US-10 | Quizmaster monitors submission status | QM cannot proceed to scoring without this |
| US-12 | Quizmaster scoring interface | Core value proposition -- replaces paper scoring |
| US-14 | Auto-tally scores | Scores must be computed for ceremony |
| US-15 | Answer ceremony (step through) | Core showmanship value |
| US-16 | Round scores on /display | Room must see scores |
| US-18 | Advance to next round | Multi-round game requires this |
| US-19 | Final scores and winner | Game must end |

**Explicitly excluded from walking skeleton (deferred to Release 2+):**

- Media questions (US: image/audio/video support)
- Multiple choice questions
- Multi-part answers
- Display holding screen (nice to have, not blocking)
- /display show current question (Release 1 /display just shows scores; QM reads aloud for Round 1 MVP)
- Running scoreboard on /host
- Late joiner catch-up
- QM override for submissions

---

## Release Plan

### Release 1: Walking Skeleton -- Text-Only, One Complete Game

**User outcome:** Marcus can run a complete trivia night with text-only questions from YAML, teams join on phones, submit answers, and scores are displayed.

**Stories:** US-01 through US-05, US-07 through US-10, US-12, US-14 through US-16, US-18, US-19

**Estimated effort:** 10-12 days

**Note:** /display in R1 shows only scores and ceremony. Questions are read aloud by Marcus.

---

### Release 2: Full Display Integration

**User outcome:** The /display screen becomes a proper shared focal point -- it shows the current question in real-time, so the room can see questions on the TV without Marcus reading aloud.

**New stories:**
- US-06: Display holding screen
- US-11: Display shows current question (text only, real-time sync)
- US-17: Ceremony scores view on /play
- US-20: Game over screen

**Estimated effort:** 3-4 days

---

### Release 3: Rich Media Questions

**User outcome:** Marcus can include image, audio, and video questions in the quiz, and players and the display show the media inline.

**New stories:**
- US-21: Image question support (/play + /display)
- US-22: Audio question support (/display plays audio; /play shows indicator)
- US-23: Video question support (/play + /display)

**Estimated effort:** 3-4 days

---

### Release 4: Advanced Question Types

**User outcome:** Marcus can include multiple choice and multi-part questions in the quiz, and players see appropriate UI for each type.

**New stories:**
- US-24: Multiple choice question UI (/play radio buttons, /display labeled choices)
- US-25: Multi-part answer entry (/play expandable fields, ordered/unordered)
- US-26: Multi-part answer scoring (per-part correct/incorrect on /host)

**Estimated effort:** 3-4 days

---

### Release 5: Resilience & Polish

**User outcome:** The game handles real-world disruptions gracefully -- late joiners, disconnections, quizmaster override -- without Marcus having to intervene.

**New stories:**
- US-27: QM override (open scoring before all teams submit)
- US-28: Late joiner catch-up (full state sync on new join)
- US-29: Reconnection handling (/play, /display, /host)
- US-30: YAML validation error messages (field-level, specific)

**Estimated effort:** 3-4 days

---

## Scope Assessment

| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Total user stories | 30 | 10 per release | PASS (split into 5 releases) |
| Bounded contexts | 3 (/host, /play, /display) | -- | Normal for this domain |
| Walking skeleton stories | 15 | 5-10 | ACCEPTABLE -- game loop is inherently multi-step |
| Release 1 estimated effort | 10-12 days | -- | OVERSIZED -- see note below |
| Stories per release (R2-R5) | 3-4 each | 1-3 days each | PASS |

### Scope Assessment: Release 1 Oversized -- Approved Split Rationale

Release 1 at 10-12 days is larger than the 1-3 day story target. However:

1. The walking skeleton for a real-time multi-party game has an inherent minimum -- you cannot have a working game without player join, question reveal, answer submission, and scoring all functioning together
2. Each individual user story within Release 1 is scoped to 1-3 days
3. The release is a thin end-to-end slice (text-only), not a complete feature
4. Releases 2-5 each deliver a clear, demonstrable user outcome in 3-4 days

**Recommendation:** Proceed with Release 1 as defined. Track progress at the individual story level (each US-01 through US-19 is independently estimable and demonstrable as a unit).

**Scope Assessment: Release 1 -- OVERSIZED but APPROVED for walking skeleton reasons. Releases 2-5 PASS.**
