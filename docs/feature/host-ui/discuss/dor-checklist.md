# Definition of Ready Checklist — Host UI (Quizmaster Panel)

**Feature ID**: host-ui
**Date**: 2026-04-09
**Validator**: Luna (Product Owner wave)

---

## DoR Gate: Per-Story Validation

### US-01: Reliable Connection Status and Auth Error Handling

| # | DoR Item | Status | Evidence |
|---|----------|--------|----------|
| 1 | Problem statement clear, domain language | PASS | "Marcus sees 'Connected' before the handshake completes and even after wrong token, with no explanation" — domain language, no tech jargon |
| 2 | User/persona with specific characteristics | PASS | Marcus, quizmaster, runs pub trivia, comfort with tech, focused on social experience |
| 3 | >= 3 domain examples with real data | PASS | (1) correct token / happy path, (2) wrong token / auth failure, (3) mid-game Wi-Fi drop |
| 4 | UAT in Given/When/Then (3-7 scenarios) | PASS | 5 scenarios in AC-01-1 through AC-01-5 |
| 5 | AC derived from UAT | PASS | Each AC maps to a scenario in user-stories.md |
| 6 | Right-sized (1-3 days, 3-7 scenarios) | PASS | 5 scenarios; estimated 1 day (WsClient hook + Host.tsx wiring) |
| 7 | Technical notes: constraints/dependencies | PASS | IC-1 (onOpen hook), IC-2 (CloseEvent.code), event names documented |
| 8 | Dependencies resolved or tracked | PASS | Dependencies on WsClient are in-project; no external blockers |
| 9 | Outcome KPIs defined with measurable targets | PASS | KPI-01 (zero false "Connected"), KPI-02 (100% auth failure clarity) |

**US-01 DoR**: PASS

---

### US-02: Load Quiz with Confirmation

| # | DoR Item | Status | Evidence |
|---|----------|--------|----------|
| 1 | Problem statement clear, domain language | PASS | "There is nothing to do after connecting — no form, no indication of what to do next" — domain language |
| 2 | User/persona with specific characteristics | PASS | Marcus, just connected, quiz file ready, friends arriving |
| 3 | >= 3 domain examples with real data | PASS | (1) valid path /home/marcus/quizzes/pub-night-vol3.yaml, (2) wrong path, (3) empty path |
| 4 | UAT in Given/When/Then (3-7 scenarios) | PASS | 4 scenarios (AC-02-1 through AC-02-4) |
| 5 | AC derived from UAT | PASS | Acceptance criteria in acceptance-criteria.md directly derived |
| 6 | Right-sized (1-3 days, 3-7 scenarios) | PASS | 4 scenarios; estimated 0.5 days |
| 7 | Technical notes: constraints/dependencies | PASS | Message schema documented (host_load_quiz, quiz_loaded, error); confirmation format noted |
| 8 | Dependencies resolved or tracked | PASS | US-01 must be complete (connected state must be accurate) |
| 9 | Outcome KPIs defined with measurable targets | PASS | KPI-03 (>= 95% load success rate when path is valid) |

**US-02 DoR**: PASS

---

### US-03: Run a Round — Start, Reveal Questions, End

| # | DoR Item | Status | Evidence |
|---|----------|--------|----------|
| 1 | Problem statement clear, domain language | PASS | "No round controls, no way to reveal questions to players, no progression mechanism" |
| 2 | User/persona with specific characteristics | PASS | Marcus, quiz loaded, friends in /play room, needs to pace reveals |
| 3 | >= 3 domain examples with real data | PASS | (1) single round 5 questions, (2) partial reveal pause at Q3, (3) multi-round advance to round 2 |
| 4 | UAT in Given/When/Then (3-7 scenarios) | PASS | 6 scenarios (AC-03-1 through AC-03-6) |
| 5 | AC derived from UAT | PASS | All ACs in acceptance-criteria.md |
| 6 | Right-sized (1-3 days, 3-7 scenarios) | PASS | 6 scenarios; estimated 1 day |
| 7 | Technical notes: constraints/dependencies | PASS | host_start_round, host_reveal_question, host_end_round, host_begin_scoring schemas documented; round_name source noted |
| 8 | Dependencies resolved or tracked | PASS | US-02 must be complete; round_name comes from round_started event (confirmed in messages.ts RoundStartedMsg) |
| 9 | Outcome KPIs defined with measurable targets | PASS | KPI-04 (zero out-of-order reveals) |

**US-03 DoR**: PASS

---

### US-04: Score a Round — Mark Answers and Publish Scores

| # | DoR Item | Status | Evidence |
|---|----------|--------|----------|
| 1 | Problem statement clear, domain language | PASS | "No scoring interface — no way to see what teams answered, mark answers, or publish scores" |
| 2 | User/persona with specific characteristics | PASS | Marcus, round ended, teams submitted, acting as judge |
| 3 | >= 3 domain examples with real data | PASS | (1) two teams Q1 one correct one wrong, (2) generous marking lowercase "paris", (3) multi-question walkthrough |
| 4 | UAT in Given/When/Then (3-7 scenarios) | PASS | 5 scenarios (AC-04-1 through AC-04-5) |
| 5 | AC derived from UAT | PASS | All ACs in acceptance-criteria.md |
| 6 | Right-sized (1-3 days, 3-7 scenarios) | PASS | 5 scenarios; estimated 1.5 days (most complex UI state) |
| 7 | Technical notes: constraints/dependencies | PASS | host_mark_answer schema documented; IC-4 open item (team submission data source) flagged for DESIGN wave |
| 8 | Dependencies resolved or tracked | CONDITIONAL PASS | IC-4 (scoring panel team answer data) is an open item for DESIGN wave; risk tracked in prioritization.md |
| 9 | Outcome KPIs defined with measurable targets | PASS | KPI-05 (zero unpublished verdicts) |

**US-04 DoR**: PASS (IC-4 open item is a DESIGN wave concern, not a PO wave blocker — the behavior requirement is clear)

---

### US-05: Run Answer Ceremony on Display

| # | DoR Item | Status | Evidence |
|---|----------|--------|----------|
| 1 | Problem statement clear, domain language | PASS | "No way to walk the room through answers on the display screen, building anticipation" |
| 2 | User/persona with specific characteristics | PASS | Marcus, scores published, room watching display |
| 3 | >= 3 domain examples with real data | PASS | (1) 5-question walkthrough, (2) rapid walkthrough, (3) ceremony complete → next round |
| 4 | UAT in Given/When/Then (3-7 scenarios) | PASS | 4 scenarios (AC-05-1 through AC-05-4) |
| 5 | AC derived from UAT | PASS | All ACs in acceptance-criteria.md |
| 6 | Right-sized (1-3 days, 3-7 scenarios) | PASS | 4 scenarios; estimated 0.5 days |
| 7 | Technical notes: constraints/dependencies | PASS | host_ceremony_show_question / reveal schemas; answer-to-display-only boundary verified in handler.go line 296; ceremonyCursor state documented |
| 8 | Dependencies resolved or tracked | PASS | US-04 must be complete; display route must be functional (separate feature — noted as dependency) |
| 9 | Outcome KPIs defined with measurable targets | PASS | KPI-06 (100% correct content on display per step) |

**US-05 DoR**: PASS

---

### US-06: End Game and View Final Leaderboard

| # | DoR Item | Status | Evidence |
|---|----------|--------|----------|
| 1 | Problem statement clear, domain language | PASS | "No ceremonial end — no final scores, no winner announcement, no way to close the session" |
| 2 | User/persona with specific characteristics | PASS | Marcus, all rounds played, ready to announce winner |
| 3 | >= 3 domain examples with real data | PASS | (1) three teams clear winner The Brainiacs 12pts, (2) tie for first 10pts each, (3) early end game after round 2 |
| 4 | UAT in Given/When/Then (3-7 scenarios) | PASS | 4 scenarios (AC-06-1 through AC-06-4) |
| 5 | AC derived from UAT | PASS | All ACs in acceptance-criteria.md |
| 6 | Right-sized (1-3 days, 3-7 scenarios) | PASS | 4 scenarios; estimated 0.5 days |
| 7 | Technical notes: constraints/dependencies | PASS | host_end_game payload (empty {}); game_over payload shape documented; sort order specified |
| 8 | Dependencies resolved or tracked | PASS | No external dependencies; game_over event already typed in messages.ts (GameOverMsg) |
| 9 | Outcome KPIs defined with measurable targets | PASS | KPI-07 (>= 90% sessions reach game_over) |

**US-06 DoR**: PASS

---

## Feature-Level DoR Summary

| Story | DoR Result |
|-------|------------|
| US-01: Connection Status + Auth | PASS |
| US-02: Load Quiz | PASS |
| US-03: Run a Round | PASS |
| US-04: Score a Round | PASS (IC-4 flagged for DESIGN wave) |
| US-05: Run Ceremony | PASS |
| US-06: End Game | PASS |

**All 6 stories pass DoR. Feature is READY for DESIGN wave handoff.**

---

## Outstanding Items for DESIGN Wave

1. **IC-4 — Scoring panel team answer data**: The `score_updated` event provides running totals but not submitted answer text. The DESIGN wave must determine whether `host_begin_scoring` triggers a state snapshot with submission data, or whether a new server event is needed. Behavior requirement is clear: Marcus must see each team's submitted answer text in the scoring panel.

2. **IC-4 — score_updated includes team_name**: Verify the `score_updated` event payload includes `team_name` (not just `team_id`) for display in the running totals. Current `ScoreUpdatedMsg` in messages.ts has `{ team_id, team_name, score }` — this appears to already be correct but should be confirmed against the actual server event sent in `hub.NewScoreUpdatedEvent`.

3. **Route alias**: Whether to add a `/host` route alias is a DESIGN wave concern. The requirement is: navigating to a wrong URL must not show a blank page. The PO recommendation is a catch-all `<Route path="*">` with a "not found" message.
