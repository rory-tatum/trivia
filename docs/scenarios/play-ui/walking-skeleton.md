# Walking Skeleton — play-ui

**Feature ID**: play-ui
**Date**: 2026-04-16
**Strategy**: C — Real Local (in-process httptest.Server + real WebSocket connections)

---

## What the Walking Skeleton Covers

The walking skeleton answers: **"Can a player join, answer questions, submit, and see scores — end to end?"**

It covers the complete game loop across 6 user stories (WS Release 0):

| Activity | Story | Protocol event |
|----------|-------|----------------|
| A: Register team | US-01 | `team_register` → `team_registered` |
| B: Wait in lobby, receive round start | US-03 (precondition) | `round_started` |
| C: Receive 3 text questions | US-03 | `question_revealed` × 3 |
| C: Save draft answers | US-03 | `draft_answer` (fire-and-forget) |
| D: Receive round ended | US-05 | `round_ended` |
| D: Submit answers, receive locked confirmation | US-06 | `submit_answers` → `submission_ack` |
| D: Play room notified of own submission | US-06/US-07 boundary | `submission_received` (DEP-03) |
| E: Receive ceremony question | US-08 | `ceremony_question_shown` |
| E: Receive revealed answer with verdicts | US-08 | `ceremony_answer_revealed` with verdicts (DEP-02) |
| F: Receive round scores with team names | US-09 | `round_scores_published` with team names (DEP-04) |

The skeleton also implicitly exercises DEP-02, DEP-03, and DEP-04 — the three most significant backend wiring changes in the DESIGN wave.

---

## What the Walking Skeleton Does NOT Cover

The following are intentionally deferred to @skip focused scenarios:

| Excluded behaviour | Reason for exclusion | Covered by |
|--------------------|----------------------|------------|
| Auto-rejoin (`team_rejoin`) | Second release priority | US02-01 through US02-03 |
| Draft answer restoration on rejoin | Second release priority | US04-01, US04-02 |
| State snapshot game-phase routing (all states) | Second release — tests rejoin | US02-01, US08-04, US10-01 |
| Connection status banner / reconnect | Second release priority | US10-01, US10-02 |
| Multiple choice questions (`choices` field) | Release 2 question types | US11-01, US11-02 |
| Multi-part questions (`is_multi_part` field) | Release 2 question types | US12-01, US12-02 |
| Media questions (`media` field) | Release 2 question types | US13-01, US13-02 |
| Post-submit team status list (US-07) | Release 3 social polish — submission_received to play room verified structurally in WS-01 | US07-01, US07-02 |
| Final leaderboard at game_over | Not in walking skeleton scope | US09-02 |
| All error paths | WS = happy path only | INF-01 through INF-06 |
| Infrastructure failures (empty names, bad tokens) | Focused infrastructure scenarios | INF-01 through INF-06 |

---

## Reusable Infrastructure from host-ui

The play-ui test infrastructure reuses the following patterns directly from `tests/acceptance/host-ui/`:

| Component | Location in host-ui | Reused in play-ui |
|-----------|---------------------|-------------------|
| `NewHostUITestServer` pattern | `steps/server_setup.go` | `NewPlayUITestServer` — identical wiring |
| `quizLoaderAdapter` struct | `steps/server_setup.go` | Copied verbatim — same architectural invariant |
| `WSMessage` struct | `steps/world.go` | Identical struct in `play-ui/steps/world.go` |
| `pollUntil` / `waitForEvent` helpers | `steps/world.go` | Identical helpers in `play-ui/steps/world.go` |
| `SimpleQuizYAML` / `MultiRoundQuizYAML` helpers | `steps/driver.go` | Copied to `play-ui/steps/driver.go` |
| `TestAcceptance` / `TestAcceptanceWalkingSkeleton` pattern | `steps/acceptance_test.go` | Identical pattern |
| `connectionKey(role, name)` | `steps/world.go` | Identical in `play-ui/steps/world.go` |
| Host command methods (`HostLoadQuiz`, `HostStartRound`, etc.) | `steps/driver.go` | Copied to `PlayUIDriver` for Given steps |

The key difference from host-ui: play-ui tests use `PlayUIDriver` rather than `HostUIDriver` as the primary driver. `PlayUIDriver` owns play room connections (`ConnectPlay`, `PlayRegisterTeam`, `PlaySubmitAnswers`, etc.) and delegates host commands only for Given-step precondition setup.

---

## PlayUIDriver — New Struct

`PlayUIDriver` is the Layer 3 test driver for play-ui. It is not a subtype of `HostUIDriver` — it is a parallel implementation that happens to share some host-driving methods.

Key methods unique to `PlayUIDriver`:

```
ConnectPlay(ctx, teamName)                         — opens /ws?room=play
PlayRegisterTeam(ctx, teamName)                    — sends team_register
PlayRejoinTeam(ctx, teamName, teamID, deviceToken) — sends team_rejoin
PlayRejoinTeamWithBadToken(ctx, teamName)          — sends team_rejoin with invalid token
PlayDraftAnswer(ctx, teamName, round, q, answer)   — sends draft_answer
PlaySubmitAnswers(ctx, teamName, round, answers)   — sends submit_answers
PlaySubmitAnswersWithID(ctx, teamName, teamID, ...) — submit with explicit team_id (error scenarios)
```

Internal helpers carried from host-ui pattern:
```
readLoop(ctx, key, conn)            — background WebSocket reader
sendMessage(ctx, connKey, msg)      — typed JSON sender
WriteQuizFixture(filename, content) — temp file writer
```

New quiz fixture generators:
```
MultipleChoiceQuizYAML(title) — quiz with one MC question (choices field)
MultiPartQuizYAML(title)      — quiz with one multi-part question (answers array)
MediaQuizYAML(title)          — quiz with one image question (media field)
```

---

## Scaffold Files Required to Compile

The following files exist under `tests/acceptance/play-ui/steps/` as RED scaffolds:

| File | Purpose | State |
|------|---------|-------|
| `acceptance_test.go` | Test entry point; `TestAcceptance`, `TestAcceptanceWalkingSkeleton`, `TestAcceptanceAdapterIntegration` | Compiles |
| `server_setup.go` | `NewPlayUITestServer` — wires real production components | Compiles |
| `driver.go` | `PlayUIDriver` — Layer 3 WebSocket black-box driver | Compiles |
| `world.go` | `World` — per-scenario state; `waitForEvent`, `pollUntil`, `addMessage` | Compiles |
| `steps.go` | `InitializeScenario` — registers all 90+ Gherkin step patterns | Compiles |
| `step_impls.go` | Step method bodies — all return `godog.ErrPending` | Compiles (RED) |

Compile verification: `go build ./tests/acceptance/play-ui/steps/...` exits 0.

Walking skeleton test verification: `go test ./tests/acceptance/play-ui/steps/... -run TestAcceptanceWalkingSkeleton` exits PASS with 1 pending scenario (first pending step is `givenQuizmasterLoadedQuiz`).

---

## Implementation Entry Point for Software-Crafter

The software-crafter should begin by implementing `givenQuizmasterLoadedQuiz` in `step_impls.go`. This requires:

1. Ensuring `w.server` is started (handled by `givenServerRunning`)
2. Writing the quiz YAML to a temp file using `driver.WriteQuizFixture`
3. Connecting a host WebSocket using `driver.ConnectHost`
4. Sending `HostLoadQuiz` with the file path
5. Waiting for `quiz_loaded` event on the host connection
6. Storing the driver and connection in `w.connections`

Once `givenQuizmasterLoadedQuiz` is green, the next failing step is `whenTeamConnects`, which requires `ConnectPlay` to succeed and `waitForEvent(eventStateSnapshot)` to confirm connection.

Follow the implementation sequence in `test-scenarios.md` one scenario at a time.
