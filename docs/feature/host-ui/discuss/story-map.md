# User Story Map — Host UI (Quizmaster Panel)

**Feature ID**: host-ui
**Date**: 2026-04-09
**Persona**: Marcus, the quizmaster

---

## Backbone (User Activities — horizontal sequence)

```
[1. Connect]  →  [2. Load Quiz]  →  [3. Run Round]  →  [4. Score Round]  →  [5. Run Ceremony]  →  [6. End Game]
```

---

## Story Map

### Activity 1: Connect

**Goal**: Marcus opens the host page and reliably knows whether he is connected and authorized.

| Walking Skeleton | Release 1 | Release 2 |
|---|---|---|
| US-01a: Trusted connection status (onOpen hook) | US-01b: Auth failure message (wrong token) | US-01c: Reconnect handling (mid-game drop) |

---

### Activity 2: Load Quiz

**Goal**: Marcus loads a quiz file and sees confirmation of what was loaded.

| Walking Skeleton | Release 1 | Release 2 |
|---|---|---|
| US-02: Load quiz form with success confirmation | US-02b: Inline error on load failure | — |

---

### Activity 3: Run Round

**Goal**: Marcus starts a round and reveals questions one by one to participants.

| Walking Skeleton | Release 1 | Release 2 |
|---|---|---|
| US-03: Start round + reveal questions | US-03b: End round (lock answers) | — |

---

### Activity 4: Score Round

**Goal**: Marcus marks each team's answer correct or wrong and sees running totals.

| Walking Skeleton | Release 1 | Release 2 |
|---|---|---|
| US-04: Scoring panel with mark correct/wrong | US-04b: Publish scores | — |

---

### Activity 5: Run Ceremony

**Goal**: Marcus walks through the answers on the display screen for the room to see.

| Walking Skeleton | Release 1 | Release 2 |
|---|---|---|
| US-05: Show question on display + reveal answer | — | — |

---

### Activity 6: End Game

**Goal**: Marcus ends the game and sees the final leaderboard.

| Walking Skeleton | Release 1 | Release 2 |
|---|---|---|
| US-06: End game + final leaderboard | — | — |

---

## Walking Skeleton

The minimal end-to-end slice that verifies the system works:

```
Marcus opens /?token=correct-token
    → connected status shows "Connected" (onOpen hook wired)
    → enters file path, clicks Load Quiz
    → sees confirmation string (quiz_loaded handled)
    → clicks Start Round 1
    → clicks Reveal Next Question (first question appears)
    → clicks End Round → Begin Scoring
    → marks one answer Correct
    → clicks Publish Scores
    → clicks End Game
    → sees final leaderboard
```

**Walking skeleton stories**: US-01a, US-02, US-03, US-04, US-06
**Demonstrable in**: A single session with one round, one question, one team

---

## Story List (all)

| ID | Title | Activity | Release |
|----|-------|----------|---------|
| US-01 | Reliable connection status and auth error handling | Connect | Walking Skeleton + R1 |
| US-02 | Load quiz with confirmation | Load Quiz | Walking Skeleton |
| US-03 | Run a round: start, reveal questions, end | Run Round | Walking Skeleton |
| US-04 | Score a round: mark answers, publish scores | Score Round | Walking Skeleton |
| US-05 | Run answer ceremony on display | Run Ceremony | Release 1 |
| US-06 | End game and view final leaderboard | End Game | Walking Skeleton |

---

## Scope Assessment: PASS

- 6 user stories
- 2 bounded contexts: frontend Host.tsx (React), WsClient (WebSocket lifecycle)
- Backend is already implemented — only frontend wiring needed
- Estimated effort: 4–6 days total (split below by release)
- All stories deliver working, user-verifiable behavior
- No story requires >3 days

Note: The WsClient lifecycle fixes (IC-1, IC-2) are prerequisite technical work enabling US-01. They are captured as Technical Notes in US-01, not as separate stories, because the user-observable outcome (accurate status) is what matters.
