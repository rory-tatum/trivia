# Evolution: host-ui-round-control

**Date**: 2026-04-15
**Feature ID**: host-ui-round-control
**Feature**: Enable 5 acceptance scenarios for the host UI round control (US03-01 through US03-05)
**Delivery type**: Acceptance-test-only (zero production code modified)

---

## Business Context

This delivery activated five previously-skipped acceptance scenarios covering the host UI round control flow. The scenarios validate that a host can start a round, reveal questions one at a time, observe live counters, and end the round — transitioning to the scoring panel. All assertions are driven by real WebSocket events; no mock state was substituted for observable protocol behaviour.

User stories covered: US03-01, US03-02, US03-03, US03-04, US03-05.

---

## Steps Completed

| Step  | Name                                                         | Completed          |
|-------|--------------------------------------------------------------|--------------------|
| 01-01 | Enable US03-01: start round confirmation and round panel counter | 2026-04-15T05:18:35Z |
| 01-02 | Enable US03-02: reveal question 1 increments counter and populates list | 2026-04-15T05:23:35Z |
| 01-03 | Enable US03-03: sequential reveal order for 3 questions      | 2026-04-15T05:26:16Z |
| 02-01 | Enable US03-04: last question triggers End Round button, hides Reveal Next | 2026-04-15T05:28:38Z |
| 02-02 | Enable US03-05: end round sends end+scoring commands, scoring panel visible | 2026-04-15T05:31:25Z |

All steps followed the RED_ACCEPTANCE → GREEN → COMMIT cycle. RED_UNIT was skipped on all steps (NOT_APPLICABLE: acceptance-test-only delivery, no production code to unit test).

A post-merge refactoring pass (L1–L4) was committed separately after step 02-02 to eliminate duplication across the step definitions.

---

## Key Decisions

### Decision 1: World fields only updated from real server events

`currentRoundName`, `revealedCount`, and `totalQuestions` are set exclusively from `round_started` payload fields. `revealedQuestions` is populated only from `question_revealed` events. No world field is pre-populated by test setup or Given steps. This eliminates Testing Theater — every Then assertion reflects observable protocol state, not fixture data.

### Decision 2: `thenRevealButtonNotVisible` required polling

A synchronous snapshot check raced the final `question_revealed` event. The check could run before the last event was processed, producing a false positive (button still visible). A polling loop with a 2-second deadline and a ticker was added to tolerate WebSocket propagation latency. This bug was discovered during the post-merge integration gate and fixed before the refactoring pass (commit `6a67d94`).

### Decision 3: L4 refactoring introduced `pollUntil` helper

The deadline-ticker-select pattern appeared more than five times across step implementations after the polling fix was applied. The pattern was extracted to a `World.pollUntil()` helper in the refactoring pass. Named constants for event types (`eventRoundStarted`, `eventQuestionRevealed`, `eventRoundEnded`, `eventScoringData`) and connection roles (`roleHost`, `rolePlay`) were also extracted, eliminating more than 20 magic string literals.

### Decision 4: `question_revealed` events filtered to host connection

The game server broadcasts `question_revealed` to all room connections (host, play, display). The `addMessage` handler processed events from all connections, causing `revealedCount` to increment once per receiving connection rather than once per event. The fix filters `addMessage` to only process events when the message key matches `roleHost`, preventing double-counting.

### Decision 5: Mutation testing skipped

No production code was modified in this delivery. All changes are confined to acceptance test step definitions and test infrastructure (`tests/acceptance/host-ui/steps/`). Per-feature mutation testing applies only to production code paths; skipping is correct per the project mutation testing strategy.

---

## Issues Encountered

| Issue | Resolution |
|-------|-----------|
| `thenRevealButtonNotVisible` raced final `question_revealed` event on synchronous check | Replaced snapshot check with 2-second polling loop; fixed in commit `6a67d94` before refactoring |
| `revealedCount` double-incremented when server broadcast to all room connections | Filtered `addMessage` to host connection key only (Decision 4) |

---

## Lessons Learned

- Assertions on "button not visible" are inherently more fragile than "button visible" assertions because they require the system to have fully settled before the check runs. Polling with a deadline is the correct default for negative-visibility checks on async state.
- Magic string literals in step definitions accumulate rapidly when scenarios multiply. Introducing named constants at the first refactoring pass (not deferred) reduces the cognitive load of adding future scenarios.
- Filtering broadcast events to the correct role connection must be a deliberate design step, not an afterthought. Broadcasting to all connections is the server's correct behaviour; the test infrastructure must match the client's consumption model.

---

## Artifacts

No lasting artifacts to migrate: this was an acceptance-test-only delivery with no design/, distill/, or discuss/ directories. The implementation lives in the test package at `tests/acceptance/host-ui/steps/`.

**Git commits (this feature)**:
- `b7cb4ed` feat(host-ui): enable US03-01 start round confirmation and panel counter - step 01-01
- `6e4be55` feat(host-ui): enable US03-02 reveal question counter and list - step 01-02
- `0748d5e` feat(host-ui): enable US03-03 sequential reveal order - step 01-03
- `88c6e5b` feat(host-ui): enable US03-04 end round button replaces reveal next - step 02-01
- `6a67d94` fix(host-ui): add polling to thenRevealButtonNotVisible - step 02-01
- `f3289d1` feat(host-ui): enable US03-05 end round and scoring panel - step 02-02
- `57410e9` refactor(host-ui-round-control): L1-L4 refactoring pass on acceptance test steps
