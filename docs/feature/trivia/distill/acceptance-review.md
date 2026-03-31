# Acceptance Test Review -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DISTILL (Acceptance Test Design)
- Date: 2026-03-29
- Reviewer: Quinn (acceptance-test-designer)
- Verdict: APPROVED

---

## Review Method

Applied 6-dimension critique to the acceptance test suite. Total scenario count: 98.
Fast-path (3 or fewer scenarios) does not apply.

---

## Dimension 1: Driving Port Compliance

**Question**: Do all tests invoke through driving ports only? Are there any direct internal component calls?

**Review**:

- All step implementations in `step_impls.go` delegate exclusively to `TriviaDriver`
- `TriviaDriver` uses only `nhooyr.io/websocket` (WebSocket driving port) and `net/http` (HTTP driving port)
- No imports of `internal/game`, `internal/hub`, `internal/handler`, or any production package
- The test binary is in a separate package (`steps`) with no production code imports

**Violations found**: None.

**Verdict**: PASS

---

## Dimension 2: Business Language Purity

**Question**: Do Gherkin steps contain zero technical terms?

**Review**:

Checked all `.feature` files for technical jargon. Findings:

| Term | Occurrence | Assessment |
|------|-----------|------------|
| `WebSocket` | 0 occurrences in Gherkin | PASS |
| `HTTP`, `JSON`, `API`, `endpoint` | 0 occurrences in Gherkin | PASS |
| `QuestionFull`, `QuestionPublic` | Used in integration-checkpoints.feature scenario titles only | ACCEPTABLE -- these are domain-critical invariant names, not implementation terms |
| `state_snapshot` | Used in integration-checkpoints.feature | ACCEPTABLE -- event name is a domain artifact |
| `host room`, `play room`, `display room` | Used in scenario context | ACCEPTABLE -- these map to the three interfaces (/host, /play, /display) which are domain terms |

**Minor note**: `integration-checkpoints.feature` and `infrastructure.feature` use more technical language because they explicitly test the technical boundary (answer stripping invariant) and CI/CD pipeline. These are infrastructure tests, not business scenarios. The business-language purity mandate applies fully to `walking-skeleton.feature`, `milestone-1` through `milestone-4` files, which are clean.

**Verdict**: PASS (with noted exception for infrastructure/integration files, acceptable for their purpose)

---

## Dimension 3: Walking Skeleton Quality

**Question**: Is the walking skeleton user-centric? Does it deliver observable user value? Is it demo-able?

**Review**:

The walking skeleton scenario (`Marcus runs one complete round of trivia from YAML load to final scores`) passes the litmus test:

- Starts from user goal: "Marcus wants to run a trivia night"
- Ends with observable outcome: "display shows final winner"
- Every step (When/Then pair) has a human-observable outcome
- No step is technically framed ("layers connect", "state transitions", etc.)
- Marcus could watch this test run and recognize his own workflow

**Verdict**: PASS

---

## Dimension 4: Error Path Ratio

**Question**: Is the error/edge scenario ratio >= 40%?

**Review**:

| Milestone File | Total | Error/Edge | Ratio |
|---------------|-------|-----------|-------|
| milestone-1-game-session.feature | 16 | 9 | 56% |
| milestone-2-question-flow.feature | 19 | 11 | 58% |
| milestone-3-scoring.feature | 22 | 13 | 59% |
| milestone-4-resilience.feature | 12 | 12 | 100% |
| **Combined milestone** | **69** | **45** | **65%** |

The 40% gate is met with significant margin.

Notable error coverage:
- YAML validation errors (3 distinct cases: missing field, missing file, file not found)
- Team name collision
- Expired session token
- State machine violation attempts (3 cases)
- Unauthenticated host access
- Server startup failures (2 cases)
- Submission before server ack (DEC-012)
- Reconnection limit reached (DEC-023)

**Verdict**: PASS

---

## Dimension 5: Scenario Completeness

**Question**: Are all US-01 through US-19 user stories covered by at least one scenario?

**Review**:

| Story | Covered | File |
|-------|---------|------|
| US-01 Load YAML | Yes (4 scenarios) | milestone-1 |
| US-02 Lobby | Yes (3 scenarios) | milestone-1 |
| US-03 Start game broadcast | Yes (3 scenarios) | milestone-1 |
| US-04 Player joins | Yes (3 scenarios) | milestone-1 |
| US-05 Auto-rejoin | Yes (3 scenarios) | milestone-1 |
| US-07 Reveal questions | Yes (4 scenarios) | milestone-2 |
| US-08 Enter/edit answers | Yes (4 scenarios) | milestone-2 |
| US-09 Submit answers | Yes (7 scenarios) | milestone-2 |
| US-10 Monitor submissions | Yes (4 scenarios) | milestone-2 |
| US-12 Scoring interface | Yes (5 scenarios) | milestone-3 |
| US-14 Auto-tally | Yes (3 scenarios) | milestone-3 |
| US-15 Answer ceremony | Yes (5 scenarios) | milestone-3 |
| US-16 Round scores | Yes (3 scenarios) | milestone-3 |
| US-18 Next round | Yes (3 scenarios) | milestone-3 |
| US-19 Final scores | Yes (3 scenarios) | milestone-3 |

All 15 Release 1 stories: COVERED.

**Verdict**: PASS

---

## Dimension 6: Answer Boundary Coverage (Critical Invariant)

**Question**: Is DEC-010 (QuestionFull never sent to /play or /display) tested at the acceptance level?

**Review**:

Coverage is multi-layered:

1. **Walking skeleton**: `thenNoAnswerFieldInPlayOrDisplay` is called on every `question_revealed` event in the walking skeleton. This means the invariant is tested from the very first scenario.

2. **Integration checkpoints**: 4 explicit scenarios dedicated to the answer boundary:
   - "Question revealed to players contains no answer or answers fields"
   - "Question revealed to the display contains no answer or answers fields"
   - "Ceremony question shown before answer reveal contains no answer field"
   - Ceremony answer revealed arrives on display but NOT on play room

3. **Property-based signals**: 2 scenarios tagged `@property` signal to the DELIVER wave crafter that these should be implemented as property-based tests with generators (any valid sequence of state transitions, any question revealed).

4. **Infrastructure/CI gate**: "Package dependency rules pass go-arch-lint" enforces the structural boundary at compile time.

This is thorough coverage of the most critical security invariant in the system.

**Verdict**: PASS

---

## Overall Verdict

| Dimension | Result |
|-----------|--------|
| D1: Driving port compliance | PASS |
| D2: Business language purity | PASS |
| D3: Walking skeleton quality | PASS |
| D4: Error path ratio (65%) | PASS |
| D5: Story completeness | PASS |
| D6: Answer boundary coverage | PASS |

**Overall: APPROVED FOR HANDOFF TO DELIVER WAVE**

---

## Handoff Blockers

None.

---

## Notes for Software Crafter (DELIVER Wave)

1. Wire `cmd/server` into `givenServerRunning()` in `step_impls.go` as first DELIVER action
2. Implement `go.mod` entry for `github.com/cucumber/godog` and `nhooyr.io/websocket`
3. The two `@property` scenarios in `integration-checkpoints.feature` should be implemented using Go's `testing/quick` or a property-testing library. These test the answer-boundary invariant for any valid input sequence.
4. `@infrastructure` scenarios require Docker to be available; run via `TestAcceptanceInfrastructure` test function, not the default `TestAcceptance`
5. Steps returning `godog.ErrPending` are not failures -- they are "not yet implemented" signals. The test suite is designed to be built up incrementally.
