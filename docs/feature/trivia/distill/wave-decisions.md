# Wave Decisions -- DISTILL Phase

## Metadata

- Feature ID: trivia
- Phase: DISTILL (Acceptance Test Design)
- Date: 2026-03-29
- Designer: Quinn (acceptance-test-designer)
- Carries forward: DEC-001 through DEC-033

---

## Inherited Decisions (All Confirmed)

All decisions from DISCOVER, DISCUSS, DESIGN, and DEVOPS waves are confirmed and carried forward unchanged. Key decisions that directly shaped the acceptance test design:

| Decision | Impact on Test Design |
|----------|-----------------------|
| DEC-003 | Three feature files by interface milestone + integration checkpoints |
| DEC-008 | All driving ports are WebSocket events; TriviaDriver wraps nhooyr/websocket |
| DEC-010 | Dedicated integration checkpoint scenarios + @property tags for invariant |
| DEC-012 | Explicit scenario: locked state only after submission_ack received |
| DEC-013 | Two-step ceremony: ceremony_question_shown then ceremony_answer_revealed |
| DEC-018 | `PayloadContainsAnswerField()` helper used in Then steps |
| DEC-020 | Auth guard tested in both infrastructure.feature and milestone-4-resilience.feature |
| DEC-022 | go:embed verified via "React application served from root path" scenario |
| DEC-023 | Reconnection limit scenario (10 attempts) in milestone-4-resilience.feature |
| DEC-031 | go-arch-lint tested as an @infrastructure scenario |

---

## New Decisions from DISTILL Wave

### DEC-034: Godog as BDD Test Runner

**Date:** 2026-03-29
**Decision:** Use `github.com/cucumber/godog` as the Gherkin/BDD test runner for Go acceptance tests.
**Rationale:** D2 (user decision: Go-native BDD). godog is the official Cucumber for Go implementation. Feature files are standard Gherkin, readable by non-technical stakeholders. Step definitions are idiomatic Go functions.
**Impact:** `acceptance_test.go` uses `godog.TestSuite` wired to `go test`. Feature files in `tests/acceptance/trivia/` directory. Step definitions in `tests/acceptance/trivia/steps/` package.

---

### DEC-035: Real In-Process Server for Acceptance Tests

**Date:** 2026-03-29
**Decision:** Acceptance tests start a real Go HTTP server via `net/http/httptest.NewServer`. No mocks at the acceptance level.
**Rationale:** D3 (user decision: real services). Mocking at the acceptance level creates Testing Theater -- the tests pass but don't validate the real behavior. `httptest.NewServer` provides a real HTTP+WebSocket server with no Docker overhead, making tests fast enough for CI.
**Impact:** `givenServerRunning()` in `step_impls.go` must wire the real `cmd/server` handler. This is the software-crafter's first DELIVER action.

---

### DEC-036: Three-Layer Step Abstraction

**Date:** 2026-03-29
**Decision:** Step definitions use a strict three-layer pattern:
- Layer 1 (Gherkin): business language in .feature files
- Layer 2 (Step methods in steps.go): delegates, no assertions
- Layer 3 (TriviaDriver in driver.go): speaks driving ports only
**Rationale:** Separating the Gherkin translation from the production port invocation makes steps reusable and maintainable. The TriviaDriver can be updated independently of the Gherkin phrasing.
**Impact:** All step implementations in `step_impls.go` must call `TriviaDriver` methods only. Assertions belong only in Then steps.

---

### DEC-037: @skip Tag on All Non-Walking-Skeleton Scenarios

**Date:** 2026-03-29
**Decision:** All scenarios in milestone files 1-4, integration-checkpoints, and infrastructure are tagged `@skip` except the walking skeleton scenario.
**Rationale:** One test at a time (Principle 4). Multiple failing tests break the TDD feedback loop. The software-crafter removes `@skip` from exactly one scenario, implements the production code to make it pass, commits, then moves to the next.
**Impact:** The `godog.Options.Tags` field in `acceptance_test.go` uses the default (no tag filter), so @skip scenarios are skipped automatically by godog's built-in skip support.

---

### DEC-038: @property Tag for Property-Shaped Invariants

**Date:** 2026-03-29
**Decision:** Two scenarios in `integration-checkpoints.feature` are tagged `@property`. These signal that the DELIVER wave crafter should implement them as property-based tests (using `testing/quick` or a Go property testing library), not as single-example assertions.
**Rationale:** The answer-boundary invariant ("QuestionFull fields never sent to /play or /display") is a universal statement ("for any valid game sequence"). A single example cannot prove a universal negative. Property-based testing generates hundreds of random valid game sequences and verifies the invariant holds for all of them.
**Impact:** The software-crafter implements these scenarios using a generator that produces valid game state sequences and asserts the invariant holds for each generated sequence.

---

### DEC-039: @infrastructure Tag Separates Docker-Dependent Scenarios

**Date:** 2026-03-29
**Decision:** Scenarios requiring Docker are tagged `@infrastructure`. They run via `TestAcceptanceInfrastructure` (separate test function), not `TestAcceptance`.
**Rationale:** Docker-dependent tests are slow (3-5 minutes for an image build) and require Docker to be installed. Separating them allows the fast acceptance suite to run on every commit while infrastructure tests run less frequently.
**Impact:** `TestAcceptance` excludes `@infrastructure` tags. `TestAcceptanceInfrastructure` is gated with `testing.Short()` and must be explicitly invoked.

---

### DEC-040: World Pattern for Per-Scenario State

**Date:** 2026-03-29
**Decision:** All per-scenario state is held in a `World` struct created fresh for each scenario via `Before` hook and torn down via `After` hook.
**Rationale:** Godog does not support test-scoped DI natively. The World pattern is the idiomatic godog approach for isolating state between scenarios. Without it, state bleeds between scenarios.
**Impact:** All step methods are bound to a `*World` receiver. The `After` hook calls `world.teardown()` to close WebSocket connections and shut down the test server.

---

## Handoff Package Contents

| Artifact | Location | Status |
|----------|----------|--------|
| Walking skeleton | `tests/acceptance/trivia/walking-skeleton.feature` | COMPLETE |
| Milestone 1 scenarios | `tests/acceptance/trivia/milestone-1-game-session.feature` | COMPLETE |
| Milestone 2 scenarios | `tests/acceptance/trivia/milestone-2-question-flow.feature` | COMPLETE |
| Milestone 3 scenarios | `tests/acceptance/trivia/milestone-3-scoring.feature` | COMPLETE |
| Milestone 4 scenarios | `tests/acceptance/trivia/milestone-4-resilience.feature` | COMPLETE |
| Integration checkpoints | `tests/acceptance/trivia/integration-checkpoints.feature` | COMPLETE |
| Infrastructure scenarios | `tests/acceptance/trivia/infrastructure.feature` | COMPLETE |
| godog entry point | `tests/acceptance/trivia/steps/acceptance_test.go` | COMPLETE |
| World struct | `tests/acceptance/trivia/steps/world.go` | COMPLETE |
| TriviaDriver | `tests/acceptance/trivia/steps/driver.go` | COMPLETE |
| Step registrations | `tests/acceptance/trivia/steps/steps.go` | COMPLETE |
| Step implementations | `tests/acceptance/trivia/steps/step_impls.go` | COMPLETE (stubs with ErrPending) |
| Test scenarios doc | `docs/feature/trivia/distill/test-scenarios.md` | COMPLETE |
| Walking skeleton doc | `docs/feature/trivia/distill/walking-skeleton.md` | COMPLETE |
| Acceptance review | `docs/feature/trivia/distill/acceptance-review.md` | COMPLETE |
| Wave decisions | `docs/feature/trivia/distill/wave-decisions.md` | THIS FILE |

**DISTILL wave is ready for handoff to software-crafter (DELIVER wave).**
