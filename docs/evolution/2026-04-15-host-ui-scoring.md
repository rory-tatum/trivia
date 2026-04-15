# Evolution: host-ui-scoring

**Date**: 2026-04-15
**Feature**: Enable 6 acceptance scenarios for the host UI scoring panel (US04-01 through US04-05 + INF-03)
**Type**: Acceptance test delivery (test-only, no production code changes)
**Wave**: DELIVER

## Feature Summary

Activated 6 previously-skipped acceptance scenarios covering the host UI scoring panel. The walking skeleton had pre-built step definitions for 5 of the 6 steps; only `thenRunningTotalUnchanged` required new implementation. Two pre-existing bugs were discovered and fixed during red-phase execution.

## Business Context

The host UI scoring panel allows the host (Marcus) to review team answers, mark them correct or wrong, publish scores, and proceed to the next round or end the game. These scenarios validate the full scoring workflow end-to-end, providing confidence that the UI interacts correctly with the WebSocket server.

## Steps Completed

| Step | Name | Outcome |
|------|------|---------|
| 01-01 | Unskip US04-01 scoring panel content scenario | PASS |
| 01-02 | Unskip US04-02 mark correct increases total | PASS |
| 02-01 | Implement thenRunningTotalUnchanged and unskip US04-03 | PASS |
| 03-01 | Unskip US04-04 publish scores shows next-round controls | PASS |
| 03-02 | Unskip US04-05 partial scoring allowed | PASS |
| 04-01 | Unskip INF-03 mark-answer rejected without active round | PASS |

All phases across all steps executed as PASS. RED_UNIT skipped for all steps (acceptance-test-only delivery).

## Key Decisions

**Decision 1: Most step definitions pre-built — 5 of 6 steps were @skip-removal only**

Only `thenRunningTotalUnchanged` required new implementation (step 02-01). The other 5 steps only required removing `@skip` tags from the feature file. This confirms the walking skeleton pre-build strategy: substantial step infrastructure is laid down during skeleton construction, reducing active delivery to targeted, focused work.

**Decision 2: givenTeamEnteredAnswers used wrong driver — pre-existing bug fixed in 01-01**

`givenTeamEnteredAnswers` called `w.hostDriver()` instead of `w.playDriver(teamName)` for play-room commands. This meant team answer submission was being sent through the host WebSocket connection rather than the player connection. Caught by RED_ACCEPTANCE in step 01-01. Fixed to use the correct driver routing.

**Decision 3: Event name mismatch — scores_published vs round_scores_published**

`thenRoundScoreSummaryVisible` waited for event `scores_published` but the server emits `round_scores_published`. The step definition was waiting on the wrong event name and would time out. Caught in step 03-01 (RED_ACCEPTANCE). Fixed by using the correct `EventRoundScoresPublished` constant.

**Decision 4: thenRunningTotalUnchanged — delta assertion not absolute**

Adversarial review caught that asserting `running_total == 0` would be brittle (only correct for fresh teams with no prior correct answers). The implementation was revised to track `teamRunningTotals map[string]float64` in `world.go` — a pre-verdict snapshot — and assert the post-verdict total equals the snapshot. This makes the assertion robust regardless of accumulated score.

**Decision 5: givenAllAnswersMarked used defaultQuestionCount=2 — hardcoded constant replaced**

Adversarial review caught that the marking loop used a hardcoded `defaultQuestionCount=2` constant rather than `w.totalQuestions` (populated from the `round_started` event). This would silently under-mark in rounds with more than 2 questions. Fixed to use the real observable from the event stream.

**Decision 6: Mutation testing skipped**

No production code was modified in this delivery — all changes are acceptance test step definitions. Mutation testing targets production source; it is not applicable here.

## Issues Encountered

| Issue | Discovery Point | Root Cause | Resolution |
|-------|----------------|------------|-----------|
| `givenTeamEnteredAnswers` wrong driver | RED_ACCEPTANCE 01-01 | Pre-existing bug: `hostDriver()` used instead of `playDriver(teamName)` | Fixed driver routing |
| Event name mismatch `scores_published` vs `round_scores_published` | RED_ACCEPTANCE 03-01 | Pre-existing bug: wrong event constant in step definition | Replaced with `EventRoundScoresPublished` |
| `thenRunningTotalUnchanged` absolute assertion | Adversarial review (02-01) | Initial impl asserted `== 0` instead of delta | Added pre-verdict snapshot in world.go |
| `givenAllAnswersMarked` hardcoded question count | Adversarial review (02-01) | Used `defaultQuestionCount=2` instead of `w.totalQuestions` | Replaced with event-sourced value |

## Lessons Learned

1. **Walking skeleton pre-builds pay dividends**: 5 of 6 steps reduced to tag removal. Investment in skeleton fidelity directly shrinks per-step delivery cost.

2. **RED_ACCEPTANCE is a real bug net**: Both pre-existing bugs (driver routing, event name) were discovered at the first red phase, not during integration. The mandatory red phase justifies its cost by catching integration assumptions that compile cleanly but fail at runtime.

3. **Adversarial review catches brittle assertions before they become flaky tests**: Both `thenRunningTotalUnchanged` issues (absolute assertion, hardcoded count) were caught in review before commit. Without adversarial review, these would have created tests that pass today and fail when game state changes.

4. **Test-only deliveries still carry bug risk**: Even when no production code changes, step definition bugs can mask product defects. The same discipline (RED, fix, GREEN) applies.

## Migrated Artifacts

No lasting artifacts to migrate — this feature had no design, distill, or discuss directories. All value is captured in the acceptance tests themselves (committed to `tests/acceptance/host-ui/`) and this evolution document.
