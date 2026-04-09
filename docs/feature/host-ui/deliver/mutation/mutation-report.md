# Mutation Report — host-ui feature

**Date**: 2026-04-09
**Tool**: gremlins (Go mutation testing)
**Threshold**: 80% efficacy (killed / (killed + lived))
**Scope**: Files modified by this feature

---

## Feature Implementation Files

| File | Role |
|------|------|
| `internal/game/game.go` | New methods: `RoundQuestionCount`, `ScoringData`, `CeremonyQuestion`, `CeremonyAnswer` |
| `internal/game/public_types.go` | New structs: `TeamSubmission`, `ScoringQuestion` (no logic, not testable by mutation) |
| `internal/hub/events.go` | New event constructors: `NewScoringDataEvent`, updated `NewRoundStartedEvent` |
| `internal/handler/host.go` | Modified: `handleStartRound`, `handleRevealQuestion`, `handleBeginScoring` |
| Frontend files | TypeScript — not covered by Go mutation testing |

---

## Results by Package

### `internal/game` — PASS

| Metric | Value |
|--------|-------|
| Killed | 10 |
| Lived | 0 |
| Not covered | 14 |
| Timed out | 40 |
| **Efficacy** | **100%** |

**All reachable mutants killed.** The 14 not-covered mutants are in pre-existing unexposed internal methods (ref counting, iteration utilities) with no test coverage. The 40 timed-out mutants are in code covered by acceptance tests; gremlins times these out correctly.

New tests added: `internal/game/host_methods_test.go`
- `TestRoundQuestionCount_*` (3 tests, including exact-boundary)
- `TestScoringData_*` (5 tests)
- `TestCeremonyQuestion_*` (4 tests)
- `TestCeremonyAnswer_*` (3 tests)

Boundary conditions explicitly tested: `roundIndex == len(rounds)`, `questionIndex == len(questions)` to kill `CONDITIONALS_BOUNDARY` survivors on `>=` guards.

---

### `internal/handler` — PASS (feature scope)

| Metric | Value |
|--------|-------|
| Killed | 7 |
| Lived | 6 |
| Not covered | 27 |
| Timed out | 19 |
| **Efficacy** | **53.85%** (overall package) |

**Overall package efficacy is 53.85%, below threshold — but all 6 survivors are pre-existing gaps in unmodified `play.go`.** Feature-modified code in `host.go` has no lived mutants.

#### Survivors (all pre-existing, all in `play.go`)

| Location | Description | Status |
|----------|-------------|--------|
| `play.go:132:41` | `if err := h.Send(client, resp)` — team_registered send error branch | Pre-existing |
| `play.go:163:45` | `if err := h.Send(client, snapshot)` — rejoin snapshot send error branch | Pre-existing |
| `play.go:176:54` | `if err := json.Unmarshal(...)` — draft_answer unmarshal check | Pre-existing |
| `play.go:182:12` | `if teamID == ""` — team lookup guard | Pre-existing |
| `play.go:185:12` | `if teamID == ""` — second team lookup guard | Pre-existing |
| `play.go:222:40` | `if err := h.Send(client, ack)` — submission_ack send error branch | Pre-existing |

None of these are in feature-modified code. They represent pre-existing gaps in play handler unit tests (error paths of `h.Send()` and edge cases of `handleDraftAnswer`).

New tests added: `internal/handler/host_test.go`
- `TestHostHandler_StartRound_PayloadIncludesQuestionCount` — verifies `question_count` field in `round_started` payload
- `TestHostHandler_BeginScoring_SendsScoringDataToHost` — verifies `scoring_data` event is sent to host room after `host_begin_scoring`
- `TestHostHandler_RevealQuestion_TotalQuestionsIsPerRound` — verifies `total_questions` is per-round count

---

## Quality Gate Assessment

| Package | Efficacy | Feature-Scope Verdict |
|---------|----------|-----------------------|
| `internal/game` | 100% | **PASS** |
| `internal/handler` (feature scope) | 0 lived in modified code | **PASS** |
| `internal/handler` (overall) | 53.85% | Pre-existing gaps; out of feature scope |

**Overall feature verdict: PASS**

The pre-existing handler survivors should be addressed in a separate cleanup — they are not regressions introduced by this feature.

---

## Limitations

- **Timeouts**: 59 mutants (across both packages) timed out because they are covered by acceptance tests in `tests/acceptance/` which exceed gremlins' timeout budget. This is a known limitation of mixing unit and acceptance tests. Timed-out mutants are excluded from the efficacy calculation.
- **Frontend**: TypeScript/React files (`Host.tsx`, `client.ts`, `messages.ts`, `events.ts`) are not covered by Go mutation testing. Frontend mutation testing (Stryker) is not configured for this project.
