# Walking Skeleton -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DISTILL (Acceptance Test Design)
- Date: 2026-03-29

---

## Definition

The walking skeleton answers: **"Can Marcus run a complete trivia night from YAML load through final scores?"**

This is a user-centric question, not a technical one. The skeleton is demo-able to Marcus today. It does not assert architectural wiring -- it asserts that a quizmaster can accomplish their goal.

---

## Walking Skeleton Scenario

File: `tests/acceptance/trivia/walking-skeleton.feature`

```
Scenario: Marcus runs one complete round of trivia from YAML load to final scores
```

This scenario covers the full game loop end-to-end in a single scenario:

1. YAML loads and game session created (US-01)
2. Player joins lobby (US-04)
3. Quizmaster starts game (US-03)
4. Quizmaster reveals questions (US-07)
5. Player enters answers (US-08)
6. Player submits answers (US-09)
7. Quizmaster marks answers (US-12)
8. Answer ceremony (US-15)
9. Scores published on display (US-16)
10. Game ends with winner on display (US-19)

Each step uses a When/Then pair to verify the observable user outcome at that step.

---

## Litmus Test

The walking skeleton passes the user-centric litmus test:

**Question: "Can Marcus run a complete trivia night on this system?"**

The scenario is demonstrable to Marcus as a stakeholder. He can watch the test run and see each step of his workflow succeed: load quiz, players join, questions reveal, answers submit, scores appear.

It is NOT technically framed. It does not say "all layers connect" or "state machine transitions" -- it says what Marcus and Priya observe.

---

## Driving Ports Exercised

The walking skeleton exercises all primary driving ports in sequence:

| Step | Driving Port (WebSocket event) | Package |
|------|-------------------------------|---------|
| Load quiz | `host_load_quiz` | handler/host.go → quiz/loader.go |
| Player joins | `team_register` | handler/play.go → hub.go |
| Start game | `host_start_round` | handler/host.go → game.go |
| Reveal question | `host_reveal_question` | handler/host.go → game.go → hub.go |
| Draft answer | `draft_answer` | handler/play.go |
| Submit answers | `submit_answers` + `submission_ack` | handler/play.go → game.go |
| Mark answer | `host_mark_answer` | handler/host.go → game/scoring.go |
| Ceremony | `host_ceremony_show_question` + `host_ceremony_reveal_answer` | handler/host.go → hub.go |
| Publish scores | `host_publish_scores` | handler/host.go → hub.go |
| End game | `host_end_game` | handler/host.go → hub.go |

No internal packages are invoked directly. All interaction is through driving ports only.

---

## What the Walking Skeleton Does NOT Cover

The following are tested in focused scenarios (milestone files), not the skeleton:

- YAML error cases (missing fields, missing files, file not found)
- Duplicate team name handling
- Auto-rejoin on refresh
- Blank answer flagging on submit review
- Submission confirmation dialog cancel
- Submission idempotency
- Quizmaster override (out of scope for Release 1)
- Multi-round progression
- Tied scores
- Reconnection resilience

These are all @skip until the walking skeleton passes.

---

## Infrastructure Required

The skeleton uses `net/http/httptest.NewServer` with the real production server wired in. The software-crafter must:

1. Wire the real `cmd/server` handler to `httptest.NewServer` in `givenServerRunning()`
2. Ensure `godog` is in `go.mod` dependencies
3. Ensure `nhooyr.io/websocket` is available for the test driver's WebSocket client

No mocks. The test hits the real server over real HTTP and WebSocket connections.

---

## First Scenario Executable Requirement

Before handoff to DELIVER wave is complete, the walking skeleton must:

1. Compile without errors (`go test ./tests/acceptance/trivia/steps/... -run=^$ -v`)
2. Run and fail for a business logic reason (e.g., `godog.ErrPending` from `givenServerRunning`)
3. NOT fail due to compilation errors or missing step definitions

Steps 1 and 2 are the software-crafter's first action in the DELIVER wave.
