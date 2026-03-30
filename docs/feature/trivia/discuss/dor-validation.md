# Definition of Ready Validation -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DISCUSS -- Phase 5 (Validate and Handoff)
- Date: 2026-03-29
- Validator: Luna (product-owner agent)

---

## DoR Checklist (9-Item Gate)

Applied to all Release 1 stories (US-01 through US-19). Release 2+ stories are written to the same template but are not gate-blocked for the current handoff.

---

### DoR Item 1: Problem statement clear, in domain language

| Story | Status | Evidence |
|-------|--------|----------|
| US-01 Load YAML | PASS | "Marcus is a quizmaster who authors quiz content in YAML files..." -- domain language, specific pain |
| US-02 Lobby | PASS | "He has no way to confirm everyone is ready without asking each person individually" |
| US-03 Start Game | PASS | "Without this synchronization, some players see the lobby while others see questions" |
| US-04 Player Joins | PASS | "She finds it disruptive to create an account, verify an email, or navigate a complex registration flow" |
| US-05 Auto-Rejoin | PASS | "He loses his team identity and has to re-enter his team name and explain to Marcus" |
| US-07 Reveal Qs | PASS | "He wants to press a button and have the next question appear for everyone simultaneously" |
| US-08 Enter Answers | PASS | "Crossing out answers looks messy and suspicious to the quizmaster" |
| US-09 Submit | PASS | "Anxious about accidentally submitting before her team is ready" |
| US-10 Monitor Status | PASS | "He has to ask each team 'did you hand in your sheet?' verbally" |
| US-12 Scoring | PASS | "The most time-consuming and error-prone part of running a trivia night" |
| US-14 Auto-Tally | PASS | "If he has to manually total the points, he introduces calculation errors" |
| US-15 Ceremony | PASS | "With paper scoring, he reads out answers in a flat list" |
| US-16 Round Scores | PASS | "He reads them aloud from his paper notes, which is easy to mishear" |
| US-18 Next Round | PASS | "He wants to start Round 2 cleanly -- resetting question reveal state" |
| US-19 Final Scores | PASS | "Anticlimactic and error-prone" |

**Item 1 Result: ALL PASS**

---

### DoR Item 2: User/persona with specific characteristics

| Story | Persona | Specific Characteristics |
|-------|---------|--------------------------|
| US-01 | Marcus Okafor | Quizmaster, YAML-fluent, hosts monthly trivia nights, home server user |
| US-02 | Marcus Okafor | Same -- in-room with guests settling in |
| US-03 | Marcus + all players | Both sides of the synchronization |
| US-04 | Priya Nair | Team captain, iPhone SE, first app use at social event |
| US-05 | Jordan Kim | Casual player, mid-game, accidentally refreshed |
| US-07 | Marcus Okafor | Running an active round, wants pacing control |
| US-08 | Priya Nair | Team captain actively playing, editing answers collaboratively |
| US-09 | Priya Nair | Team captain at end of round, anxious about finality |
| US-10 | Marcus Okafor | Waiting for round submissions before scoring |
| US-12 | Marcus Okafor | Scoring phase, motivated for speed |
| US-14 | Marcus Okafor | Completing scoring, wants instant totals |
| US-15 | Marcus Okafor | Ceremony mode, showman motivation |
| US-16 | All room participants | Post-ceremony, waiting for next round |
| US-18 | Marcus Okafor | Between rounds, maintains momentum |
| US-19 | All room participants | End of game, celebrating |

**Item 2 Result: ALL PASS**

---

### DoR Item 3: 3+ domain examples with real data

| Story | Example 1 | Example 2 | Example 3 |
|-------|-----------|-----------|-----------|
| US-01 | Valid YAML loads (Marcus, "friday-march-2026.yaml") | Missing "answer" field (Round 2, Q3) | Missing media file ("eiffel.jpg") |
| US-02 | All teams present, game starts | Multiple devices same team | Start blocked, no teams |
| US-03 | All clients sync (6 devices, 3 teams) | Late joiner after start | Offline during start, catch up |
| US-04 | Priya joins "Team Awesome" | Jordan joins existing "Team Awesome" | "Quiz Killers" already taken |
| US-05 | Refresh mid-round, 3 answers restored | Tab closed 10 min, Round 2 scoring | Expired session, fresh join |
| US-07 | Sequential reveal Q1-Q8 | Image question (eiffel.jpg) | Last question enables End Round |
| US-08 | Priya changes "Paris, France" to "Paris" | Jordan changes "Venus" to "Mercury" | Blank Q5 allowed |
| US-09 | Full submission, all answers | Blank Q5, proceeds anyway | Cancels dialog, corrects typo |
| US-10 | All submit, scoring opens | One team waits | Override after timeout |
| US-12 | Scores Q1 "paris" correct | Q1 "Lyon" wrong | Accidentally wrong, toggles |
| US-14 | 3 teams, 8 Qs, auto-tallied | Changing verdict recalculates | Partial scoring saved |
| US-15 | Walk through all 8 answers | Pause on Q5 | Back-navigate to Q4 |
| US-16 | 3 teams ranked after Round 1 | Tied teams same rank | Scores stay until Round 2 |
| US-18 | Clean Round 1 to Round 2 transition | Final round shows End Game | Accidental early advance |
| US-19 | Clear winner (Team Awesome 32 pts) | Tied winner (both 30 pts) | Final screen after server restart |

**Item 3 Result: ALL PASS -- all examples use real persona names and specific data**

---

### DoR Item 4: UAT in Given/When/Then (3-7 scenarios per story)

| Story | Scenario Count | Status |
|-------|---------------|--------|
| US-01 | 4 scenarios | PASS |
| US-02 | 4 scenarios | PASS |
| US-03 | 3 scenarios | PASS |
| US-04 | 4 scenarios | PASS |
| US-05 | 4 scenarios | PASS |
| US-07 | 4 scenarios | PASS |
| US-08 | 4 scenarios | PASS |
| US-09 | 5 scenarios | PASS |
| US-10 | 4 scenarios | PASS |
| US-12 | 6 scenarios | PASS |
| US-14 | 3 scenarios | PASS |
| US-15 | 5 scenarios | PASS |
| US-16 | 3 scenarios | PASS |
| US-18 | 3 scenarios | PASS |
| US-19 | 3 scenarios | PASS |

**Item 4 Result: ALL PASS -- all between 3-7 scenarios, all in Given/When/Then format**

---

### DoR Item 5: Acceptance Criteria derived from UAT

| Story | AC Count | Derived from Scenarios | Status |
|-------|---------|----------------------|--------|
| US-01 | 7 AC | Each scenario has corresponding AC | PASS |
| US-02 | 6 AC | Lobby list, copy link, start game, guard | PASS |
| US-03 | 5 AC | Sync on start, late join, catch-up | PASS |
| US-04 | 7 AC | Join, multi-device, duplicate, after-start | PASS |
| US-05 | 6 AC | Restore answers, locked state, expired, no-token | PASS |
| US-07 | 6 AC | Send per-question, cumulative/current, hidden, end-round | PASS |
| US-08 | 7 AC | Field per question, edit, persist, blank | PASS |
| US-09 | 9 AC | Review screen, blank flag, dialog, go-back, lock, marcus update | PASS |
| US-10 | 7 AC | Real-time update, active/inactive, override, blank answers | PASS |
| US-12 | 8 AC | Show answers, correct/wrong, toggle, auto-update, ceremony gate | PASS |
| US-14 | 5 AC | Auto-update, recalculate, final totals accurate | PASS |
| US-15 | 7 AC | Ceremony panel, question-only then answer, navigate, end ceremony | PASS |
| US-16 | 6 AC | Rank order, row data, tied, persist, /play mirror | PASS |
| US-18 | 7 AC | Reset, all clients, /display, previous round preserved, end game | PASS |
| US-19 | 6 AC | Final standings, winner, tied, /play mirror, persist, locked | PASS |

**Item 5 Result: ALL PASS**

---

### DoR Item 6: Right-sized (1-3 days, 3-7 scenarios)

| Story | Scenarios | Estimated Days | Status |
|-------|-----------|---------------|--------|
| US-01 | 4 | 1-2 | PASS |
| US-02 | 4 | 1 | PASS |
| US-03 | 3 | 1 | PASS |
| US-04 | 4 | 1-2 | PASS |
| US-05 | 4 | 1-2 | PASS |
| US-07 | 4 | 1-2 | PASS |
| US-08 | 4 | 1 | PASS |
| US-09 | 5 | 1-2 | PASS |
| US-10 | 4 | 1 | PASS |
| US-12 | 6 | 2-3 | PASS |
| US-14 | 3 | 1 | PASS |
| US-15 | 5 | 2 | PASS |
| US-16 | 3 | 1 | PASS |
| US-18 | 3 | 1 | PASS |
| US-19 | 3 | 1 | PASS |

**Item 6 Result: ALL PASS -- all within 1-3 day estimates, all have 3-7 scenarios**

---

### DoR Item 7: Technical notes -- constraints and dependencies

| Story | Notes Present | Key Constraints |
|-------|--------------|-----------------|
| US-01 | PASS | YAML schema fields, media relative paths, answer field filtering (ART-02 risk) |
| US-02 | PASS | ART-03 consumed, WebSocket broadcast for start |
| US-03 | PASS | IC-01 integration checkpoint, state snapshot for catch-up |
| US-04 | PASS | localStorage token format, ART-03, case-insensitive uniqueness |
| US-05 | PASS | localStorage key format, IC-05, server draft store |
| US-07 | PASS | ART-04 broadcast, answer field stripping, media static serving, IC-02 |
| US-08 | PASS | ART-05 draft storage, Release 1 text-only scope |
| US-09 | PASS | ART-06 immutability (DEC-006), server ACK required (IC-03) |
| US-10 | PASS | ART-06 updates via WebSocket, blank answers on override |
| US-12 | PASS | ART-06 consumed, ART-07 written, ART-08 computed, answer privacy |
| US-14 | PASS | ART-08 server-side computation, deterministic recalculation |
| US-15 | PASS | Ceremony per-step broadcast, ART-02 answer release keyed by index |
| US-16 | PASS | ART-09 = ROUND_SCORES, ART-08 sent to /display and /play |
| US-18 | PASS | ART-04 reset, ART-09 broadcast, previous round data preserved |
| US-19 | PASS | ART-08 aggregate, ART-09 = GAME_OVER, in-memory only (DEC-004) |

**Item 7 Result: ALL PASS**

---

### DoR Item 8: Dependencies resolved or tracked

| Dependency | Status | Notes |
|------------|--------|-------|
| WebSocket infrastructure | TRACKED | Required by US-02, 03, 07, 09, 10, 12, 15, 18, 19. Must be implemented before any real-time stories. Solution architect to design. |
| localStorage browser API | RESOLVED | Standard browser capability; no external dependency |
| YAML parsing library | TRACKED | Standard library (Node: js-yaml, Python: PyYAML). Solution architect to select. |
| Static file serving | RESOLVED | Built into any web server; no special dependency |
| DEC-004 (no database) | RESOLVED | Confirmed -- in-memory state only |
| DEC-006 (submission finality) | RESOLVED | Documented; enforced in US-09 |
| DEC-008 (WebSocket) | RESOLVED | Technology decision made; library selection deferred to solution architect |
| Answer field privacy (ART-02 risk) | TRACKED | Critical -- server must filter answer fields from /play and /display payloads. Must be in solution-architect design brief. |

**Item 8 Result: ALL PASS -- dependencies identified and tracked**

---

### DoR Item 9: Outcome KPIs defined with measurable targets

| Story | KPI Reference | Measurable Target |
|-------|--------------|-------------------|
| US-01 | KPI-04 | Under 2 minutes setup |
| US-02 | KPI-04 | Under 2 minutes setup |
| US-03 | IC-01 timing | Within 1 second broadcast |
| US-04 | KPI-03 | Under 60 seconds join |
| US-05 | KPI-02 | 0% answer loss on refresh |
| US-07 | IC-02 timing | Within 1 second reveal |
| US-08 | Functional test | 0 forced errors from answer entry |
| US-09 | KPI-05 | 100% teams submit before override |
| US-10 | KPI-05 | Override never needed in normal play |
| US-12 | KPI-01 | Under 3 minutes for full round |
| US-14 | Functional test | Score = count of correct verdicts |
| US-15 | KPI-06 | Quizmaster satisfaction: "I'd use this again" |
| US-16 | KPI-07 | 0 display information leaks |
| US-18 | Timing test | Round transition under 10 seconds |
| US-19 | KPI-06 | 100% games end with winner on /display |

**Item 9 Result: ALL PASS -- every story references a KPI with a measurable target**

---

## DoR Summary

| Story | D1 | D2 | D3 | D4 | D5 | D6 | D7 | D8 | D9 | Result |
|-------|----|----|----|----|----|----|----|----|----|----|
| US-01 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-02 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-03 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-04 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-05 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-07 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-08 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-09 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-10 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-12 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-14 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-15 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-16 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-18 | P | P | P | P | P | P | P | P | P | **PASS** |
| US-19 | P | P | P | P | P | P | P | P | P | **PASS** |

### DoR Gate: ALL 15 RELEASE 1 STORIES PASS

**Blocker count: 0**
**No partial handoffs. All stories are ready for DESIGN wave.**

---

## Anti-Pattern Check

| Anti-Pattern | Checked | Result |
|---|---|---|
| "Implement-X" title pattern | Checked all US titles | PASS -- all start from user pain, not technical implementation |
| Generic data (user123, test@test.com) | Checked all examples | PASS -- Marcus Okafor, Priya Nair, Jordan Kim, "Team Awesome", "The Brainiacs", "Quiz Killers", "friday-march-2026.yaml" |
| Technical AC (e.g., "Use JWT tokens") | Checked all AC lists | PASS -- all AC describe observable user outcomes, not implementation |
| Oversized story (>7 scenarios) | Checked all story scenario counts | PASS -- max 6 scenarios (US-12), all others ≤5 |
| Abstract requirements | Checked all domain examples | PASS -- all examples have real names, specific data values, and concrete outcomes |
