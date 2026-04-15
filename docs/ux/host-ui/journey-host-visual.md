# Journey Visual: Host UI (Quizmaster Panel)

**Feature ID**: host-ui
**Persona**: Marcus, the quizmaster — a friend running pub trivia for their group. Comfortable with tech but focused on the social experience; wants controls to be obvious so he can stay in conversation, not stare at a screen.
**Date**: 2026-04-09

---

## Emotional Arc

```
Anxiety          Confidence       Engagement        Authority         Satisfaction
(will it work?)  (it connected!)  (game is live)    (I control pace)  (game over, scores final)
     |                |                |                  |                  |
  [LAND]          [CONNECT]        [LOAD]            [PLAY LOOP]        [END GAME]
```

---

## Happy Path: ASCII Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  STEP 1: LAND                                                                   │
│  Marcus opens /?token=HOST_SECRET in browser                                    │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  Quizmaster Panel                           ● Connecting...             │   │
│  │                                                                         │   │
│  │  Waiting for WebSocket handshake...                                     │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
│  Emotional note: Anxiety — "Is my token right? Will it connect?"                │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │ onOpen fires
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  STEP 2: CONNECTED (IDLE)                                                       │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  Quizmaster Panel                           ● Connected                 │   │
│  │  ─────────────────────────────────────────────────────────────────────  │   │
│  │                                                                         │   │
│  │  Load Quiz                                                              │   │
│  │  Quiz file path: [______________________________]  [Load Quiz]          │   │
│  │                                                                         │   │
│  │  Players join at: http://host/play                                      │   │
│  │  Display at:      http://host/display                                   │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
│  Emotional note: Confidence — "I'm in. I can see what to do next."              │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │ Marcus enters path, clicks Load Quiz
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  STEP 3: QUIZ LOADED                                                            │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  Quizmaster Panel                           ● Connected                 │   │
│  │  ─────────────────────────────────────────────────────────────────────  │   │
│  │                                                                         │   │
│  │  ✓ Quiz loaded: "Pub Night Vol. 3 | 3 rounds | 15 questions"            │   │
│  │                                                                         │   │
│  │  Round 1: General Knowledge                                             │   │
│  │  [Start Round 1]                                                        │   │
│  │                                                                         │   │
│  │  Players join at: http://host/play                                      │   │
│  │  Display at:      http://host/display                                   │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
│  Emotional note: Engagement — "Great, quiz is ready. Friends can join now."     │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │ Marcus clicks Start Round 1
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  STEP 4: ROUND ACTIVE                                                           │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  Quizmaster Panel                           ● Connected                 │   │
│  │  ─────────────────────────────────────────────────────────────────────  │   │
│  │  Round 1: General Knowledge           Question 0 of 5 revealed          │   │
│  │                                                                         │   │
│  │  [ Reveal Next Question ]                                               │   │
│  │                                                                         │   │
│  │  Revealed so far:                                                       │   │
│  │  Q1: What is the capital of France?                                     │   │
│  │  Q2: How many bones in the human body?                                  │   │
│  │                                                                         │   │
│  │  [ End Round ]                                                          │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
│  Emotional note: Authority — "I control the pace. Friends are answering."       │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │ Marcus clicks End Round (all questions revealed)
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  STEP 5: SCORING                                                                │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  Quizmaster Panel                           ● Connected                 │   │
│  │  ─────────────────────────────────────────────────────────────────────  │   │
│  │  Scoring: Round 1                                                       │   │
│  │                                                                         │   │
│  │  Q1: What is the capital of France?   Answer: Paris                    │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐   │   │
│  │  │ Team: The Brainiacs    Answered: "Paris"     [✓ Correct] [✗ Wrong]│  │   │
│  │  │ Team: Quiz Killers     Answered: "paris"     [✓ Correct] [✗ Wrong]│  │   │
│  │  └─────────────────────────────────────────────────────────────────┘   │   │
│  │                                                                         │   │
│  │  Running totals:  The Brainiacs: 3pts   Quiz Killers: 2pts              │   │
│  │                                                                         │   │
│  │  [ Publish Scores ]                                                     │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
│  Emotional note: Authority + Focus — "I'm the judge. These calls are mine."     │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │ Marcus clicks Publish Scores → next round
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  STEP 6: CEREMONY (answer walkthrough on display)                               │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  Quizmaster Panel                           ● Connected                 │   │
│  │  ─────────────────────────────────────────────────────────────────────  │   │
│  │  Ceremony: Round 1 — showing answers on display screen                  │   │
│  │                                                                         │   │
│  │  Question 1 of 5 shown on display                                       │   │
│  │  [ Show Next Question ]   [ Reveal Answer ]                             │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
│  Emotional note: Engagement — "Everyone can see the answers being walked through"│
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │ After all rounds + ceremony
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  STEP 7: GAME OVER                                                              │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  Quizmaster Panel                           ● Connected                 │   │
│  │  ─────────────────────────────────────────────────────────────────────  │   │
│  │  Game Over!                                                             │   │
│  │                                                                         │   │
│  │  Final Scores:                                                          │   │
│  │  1. The Brainiacs      12 pts                                           │   │
│  │  2. Quiz Killers        9 pts                                           │   │
│  │  3. Random Facts        7 pts                                           │   │
│  │                                                                         │   │
│  │  [ End Game ]                                                           │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
│  Emotional note: Satisfaction — "We did it. Clear winner, great night."         │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Error Paths

### Auth Failure (wrong token)
```
Marcus navigates to /?token=WRONG
        │
        ▼ WS upgrade returns 403
        │ CloseEvent code 1006 on first attempt
        ▼
┌─────────────────────────────────────────────────────────┐
│  Quizmaster Panel                  ● Disconnected        │
│  ─────────────────────────────────────────────────────  │
│  Connection refused — invalid token.                    │
│  Check your HOST_TOKEN and reload.                      │
└─────────────────────────────────────────────────────────┘
        Emotional note: Clarity over confusion — Marcus knows exactly what went wrong
```

### Quiz Load Failure (bad file path)
```
Marcus enters wrong file path → clicks Load Quiz
        │ Server returns error event: quiz_load_failed
        ▼
┌─────────────────────────────────────────────────────────┐
│  Load Quiz                                              │
│  Quiz file path: [/wrong/path.yaml]  [Load Quiz]        │
│  ✗ Could not load quiz: file not found at /wrong/path   │
└─────────────────────────────────────────────────────────┘
        Emotional note: Immediate feedback, can correct and retry
```

### Connection Drop During Game
```
WS drops mid-round (network hiccup)
        │ onClose fires, reconnect loop starts
        ▼ status changes to "Reconnecting..."
        │ onOpen fires on successful reconnect
        ▼ status returns to "Connected"
        Game controls remain visible (client-side state preserved)
```

---

## Shared Artifacts

| Artifact | Source | Consumers |
|---|---|---|
| `token` | URL query param `?token=` | WsClient auth, Host.tsx |
| `gamePhase` | Derived from WS events | All render branches in Host.tsx |
| `quizMeta` | `quiz_loaded` payload | Confirmation string, round/question counts |
| `revealedQuestions` | Accumulated `question_revealed` payloads | Round active panel |
| `teamAnswers` | Per-team, per-question, from game session | Scoring panel |
| `scores` | `score_updated` events | Running totals, final scores |
| `roundIndex` | `round_started` payload | Round active + scoring + ceremony panels |
| `playerURL` | `quiz_loaded` payload (`meta.PlayerURL`) | Displayed in idle/loaded panels |
| `displayURL` | `quiz_loaded` payload (`meta.DisplayURL`) | Displayed in idle/loaded panels |
