# Test Scenarios -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DISTILL (Acceptance Test Design)
- Date: 2026-03-29
- Release: 1 (Text-Only Walking Skeleton)

---

## Scenario Inventory

### Walking Skeleton (`walking-skeleton.feature`)

| # | Scenario | Tag | Driving Port |
|---|----------|-----|-------------|
| 1 | Marcus runs one complete round from YAML load to final scores | `@walking_skeleton` | WebSocket host room + play room |

Total: 1 scenario (no @skip -- first to implement)

---

### Milestone 1: Game Session Management (`milestone-1-game-session.feature`)

| # | Scenario | Story | Error/Edge? |
|---|----------|-------|-------------|
| 1 | Valid YAML file creates a game session | US-01 | Happy |
| 2 | YAML file with a missing required field shows a specific error | US-01 | **Error** |
| 3 | YAML file referencing a missing media file shows a specific error | US-01 | **Error** |
| 4 | Providing a path to a file that does not exist shows a clear error | US-01 | **Error** |
| 5 | Connected teams appear in the lobby within 2 seconds | US-02 | Happy |
| 6 | Multiple devices for one team show a device count, not duplicate entries | US-02 | Edge |
| 7 | Start game is blocked when no teams have joined | US-02 | **Error** |
| 8 | Starting the game transitions all connected players simultaneously | US-03 | Happy |
| 9 | A player who joins after game start receives the current game state | US-03 | Edge |
| 10 | A player who was offline during game start catches up on reconnection | US-03 | Edge |
| 11 | Player registers a new team name successfully | US-04 | Happy |
| 12 | Registering a team name that is already taken shows a clear error | US-04 | **Error** |
| 13 | Team name matching is case-insensitive for the uniqueness check | US-04 | Edge |
| 14 | Player restores their session after accidental page refresh | US-05 | Happy |
| 15 | Player with an expired or unknown session token starts fresh | US-05 | **Error** |
| 16 | Rejoining player during scoring sees the locked state | US-05 | Edge |

Total: 16 scenarios | Error/Edge count: 9 (56%) -- exceeds 40% target

---

### Milestone 2: Question Reveal and Answer Entry (`milestone-2-question-flow.feature`)

| # | Scenario | Story | Error/Edge? |
|---|----------|-------|-------------|
| 1 | Revealing question 1 sends it to all player screens | US-07 | Happy |
| 2 | Revealing a second question cumulates on player screens but replaces the display | US-07 | Happy |
| 3 | Questions are revealed in the correct sequence without skipping | US-07 | Edge |
| 4 | After all questions are revealed, end round becomes available | US-07 | Edge |
| 5 | Player enters a text answer into a revealed question field | US-08 | Happy |
| 6 | Player changes an answer after initial entry | US-08 | Happy |
| 7 | Player leaves an answer blank and moves on | US-08 | Edge |
| 8 | Draft answers survive a player page refresh mid-round | US-08 | **Error** |
| 9 | Player reviews all answers before submitting | US-09 | Happy |
| 10 | Player sees a blank question flagged on the review screen | US-09 | Edge |
| 11 | Player confirms submission through the confirmation step | US-09 | Happy |
| 12 | Player cancels the submission confirmation and returns to editing | US-09 | Edge |
| 13 | Player cannot edit answers after submission is confirmed | US-09 | **Error** |
| 14 | Player screen does not show locked state until server confirms the submission | US-09 | Edge |
| 15 | Resubmitting after a network interruption does not duplicate or overwrite | US-09 | **Error** |
| 16 | Quizmaster sees real-time submission status as teams submit | US-10 | Happy |
| 17 | Quizmaster sees all teams submitted and scoring becomes available | US-10 | Happy |
| 18 | Scoring remains blocked while at least one team has not submitted | US-10 | **Error** |
| 19 | A team that submitted with a blank answer shows the blank in scoring | US-10 | Edge |

Total: 19 scenarios | Error/Edge count: 11 (58%) -- exceeds 40% target

---

### Milestone 3: Scoring, Ceremony, and Round Scores (`milestone-3-scoring.feature`)

| # | Scenario | Story | Error/Edge? |
|---|----------|-------|-------------|
| 1 | Scoring interface shows all submitted answers for a question | US-12 | Happy |
| 2 | Quizmaster marks an answer correct and the score increments | US-12 | Happy |
| 3 | Quizmaster marks an answer wrong and the score does not change | US-12 | Happy |
| 4 | Quizmaster toggles a verdict from wrong to correct | US-12 | Edge |
| 5 | Scoring answers does not send answer content to player connections | US-12 | **Invariant** |
| 6 | Running totals are calculated automatically as verdicts are entered | US-14 | Happy |
| 7 | Changing a verdict recalculates the total instantly | US-14 | Edge |
| 8 | Final totals match the exact count of correct verdicts | US-14 | Happy |
| 9 | Starting the ceremony transitions the display to ceremony mode | US-15 | Happy |
| 10 | Advancing the ceremony first shows the question then reveals the answer | US-15 | Happy |
| 11 | Ceremony answer reveal is shown only on the display, not the player screens | US-15 | **Invariant** |
| 12 | Quizmaster can navigate back to a previous ceremony question | US-15 | Edge |
| 13 | Ceremony completes only after all questions have been stepped through | US-15 | **Error** |
| 14 | Publishing scores shows ranked results on the display screen | US-16 | Happy |
| 15 | Tied teams are shown at the same rank | US-16 | Edge |
| 16 | Player screens show the same scores as the display after publishing | US-16 | Happy |
| 17 | Starting the next round resets the question reveal state | US-18 | Happy |
| 18 | Round 1 scores are preserved when round 2 begins | US-18 | Edge |
| 19 | The final round offers "End Game" instead of "Next Round" | US-18 | Edge |
| 20 | Ending the game shows final standings on the display screen | US-19 | Happy |
| 21 | Tied winners are shown together on the final display | US-19 | Edge |
| 22 | Player screens show final scores after game end | US-19 | Happy |

Total: 22 scenarios | Error/Invariant/Edge count: 13 (59%) -- exceeds 40% target

---

### Milestone 4: Reconnection and Edge Cases (`milestone-4-resilience.feature`)

| # | Scenario | Coverage | Error/Edge? |
|---|----------|----------|-------------|
| 1 | Player reconnects after a brief connection loss and sees current state | US-05 extended | **Error** |
| 2 | Player connection drops and game state advances before reconnect | US-05 extended | **Error** |
| 3 | Player reconnects after submitting and sees the locked state | US-05 extended | **Error** |
| 4 | Player reaches reconnection limit and sees a manual reconnect prompt | DEC-023 | **Error** |
| 5 | Display screen reconnects and shows the current game state | Display resilience | **Error** |
| 6 | Quizmaster reopens the host panel and game state is fully restored | Host resilience | **Error** |
| 7 | Attempting to reveal question 2 before question 1 is not allowed | State machine | **Error** |
| 8 | Attempting to start scoring before all questions are revealed is not allowed | State machine | **Error** |
| 9 | Attempting to start the ceremony before all answers are scored is not allowed | State machine | **Error** |
| 10 | An unauthenticated request to the host interface is rejected | DEC-020 | **Security** |
| 11 | Server fails to start when the host token is not configured | Platform | **Error** |
| 12 | Server fails to start when the quiz directory is not accessible | Platform | **Error** |

Total: 12 scenarios | All Error/Edge/Security (100%)

---

### Integration Checkpoints (`integration-checkpoints.feature`)

| # | Scenario | Invariant | Tag |
|---|----------|-----------|-----|
| 1 | Game start event reaches all connected rooms within 1 second | IC-01 | `@skip` |
| 2 | Question reveal event reaches play and display rooms within 1 second | IC-02 | `@skip` |
| 3 | Submission acknowledgment sent before locked state is shown | IC-03/DEC-012 | `@skip` |
| 4 | Question revealed to players contains no answer or answers fields | **DEC-010** | `@skip` |
| 5 | Question revealed to display contains no answer or answers fields | **DEC-010** | `@skip` |
| 6 | Ceremony question shown before answer reveal contains no answer field | DEC-013/DEC-010 | `@skip` |
| 7 | Answer fields are never present in any message to play room | DEC-010 | `@skip @property` |
| 8 | Answer fields never in question_revealed to any non-host room | DEC-010 | `@skip @property` |
| 9 | New player connection receives current game state snapshot | IC-04 | `@skip` |
| 10 | Draft answers retrievable after player reconnects | IC-05 | `@skip` |
| 11 | WebSocket connection to host room without token is rejected | DEC-020 | `@skip` |
| 12 | Valid token allows connection to host room | DEC-020 | `@skip` |

Total: 12 scenarios | 2 @property-tagged (property-based test signals)

---

### Infrastructure (`infrastructure.feature`)

| # | Scenario | Coverage | Tag |
|---|----------|----------|-----|
| 1 | Server starts successfully with required env vars | Startup | `@skip` |
| 2 | Server refuses to start without HOST_TOKEN | Startup | `@skip` |
| 3 | Server refuses to start when QUIZ_DIR not accessible | Startup | `@skip` |
| 4 | Requests to host panel without token receive 403 | Auth guard | `@skip` |
| 5 | Requests with incorrect token receive 403 | Auth guard | `@skip` |
| 6 | Requests with correct token are accepted | Auth guard | `@skip` |
| 7 | Player interface accessible without token | Access control | `@skip` |
| 8 | Display interface accessible without token | Access control | `@skip` |
| 9 | Media files in quiz directory served under media path | Media serving | `@skip` |
| 10 | Requesting missing media file returns 404 | Media serving | `@skip` |
| 11 | React application served from root path | go:embed | `@skip` |
| 12 | Docker image builds without errors | CI/CD | `@skip @infrastructure` |
| 13 | Built container starts and serves application | CI/CD | `@skip @infrastructure` |
| 14 | Package dependency rules pass go-arch-lint | DEC-031/DEC-018 | `@skip @infrastructure` |
| 15 | TypeScript type checking passes with zero errors | CI gate | `@skip @infrastructure` |
| 16 | Go tests pass with race detector | CI gate | `@skip @infrastructure` |

Total: 16 scenarios | @infrastructure-tagged run separately (require Docker)

---

## Summary

| File | Scenarios | Error/Edge % | Walking Skeleton? |
|------|-----------|-------------|-------------------|
| walking-skeleton.feature | 1 | N/A | Yes (no @skip) |
| milestone-1-game-session.feature | 16 | 56% | No |
| milestone-2-question-flow.feature | 19 | 58% | No |
| milestone-3-scoring.feature | 22 | 59% | No |
| milestone-4-resilience.feature | 12 | 100% | No |
| integration-checkpoints.feature | 12 | N/A (all invariants) | No |
| infrastructure.feature | 16 | N/A (all infra) | No |
| **Total** | **98** | **>40% across all** | **1 active** |

Error path ratio across milestone files: (9+11+13+12) / (16+19+22+12) = 45/69 = **65%** -- well above 40% gate.

---

## Answer Boundary Coverage

The critical invariant (DEC-010: QuestionFull fields never sent to /play or /display) is covered at three levels:

1. **Walking skeleton** (implicit): `thenNoAnswerFieldInPlayOrDisplay` asserted on every question reveal
2. **Integration checkpoints** (explicit): 4 dedicated scenarios + 2 @property scenarios for this invariant
3. **Milestone 3** (explicit): "Scoring answers does not send answer content to player connections" and "Ceremony answer reveal is shown only on the display, not the player screens"
4. **CI gate** (structural): `go-arch-lint` enforces at compile time via `@infrastructure` scenario

---

## One-at-a-Time Implementation Sequence

1. **Enable**: walking-skeleton.feature (already no @skip)
2. After walking skeleton passes: remove @skip from `milestone-1-game-session.feature` scenarios one at a time, starting with "Valid YAML file creates a game session"
3. After milestone 1 fully passes: proceed to milestone-2-question-flow.feature
4. After milestone 2: milestone-3-scoring.feature
5. After milestone 3: milestone-4-resilience.feature
6. After milestone 4: integration-checkpoints.feature
7. Infrastructure scenarios run separately via `TestAcceptanceInfrastructure`
