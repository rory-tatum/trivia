# Shared Artifacts Registry -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DISCUSS -- Phase 3 (Coherence Validation)
- Date: 2026-03-29

---

## Purpose

This registry tracks every data artifact that crosses interface boundaries (/host, /play, /display, server). Each artifact has a single authoritative source. Integration failures most often occur when two components independently maintain different versions of the same artifact.

---

## Artifact Registry

### ART-01: Game Session

| Property | Value |
|----------|-------|
| ID | `game_session_id` |
| Description | Unique identifier for the active game instance |
| Type | UUID string |
| Authoritative source | Server (created on YAML load) |
| Produced by | /host on successful YAML validation |
| Consumed by | /play (join), /display (WebSocket subscription) |
| Persistence | Server in-memory; player localStorage (for rejoin) |
| Lifecycle | Created at YAML load; destroyed when quizmaster ends game or server restarts |
| Risk | If server restarts mid-game, all state is lost -- acceptable for personal use (DEC-004) |

---

### ART-02: Quiz Content Tree

| Property | Value |
|----------|-------|
| ID | `quiz_content_tree` |
| Description | Fully parsed and validated quiz structure: title, rounds array, questions array per round, answer/answers per question, media refs, question type flags |
| Type | Nested object |
| Authoritative source | Server (parsed from YAML on load) |
| Produced by | /host YAML load (US-01) |
| Consumed by | /host (reveal panel, scoring), /play (question rendering), /display (question rendering) |
| Persistence | Server in-memory |
| Sensitive fields | `answer`/`answers` fields -- MUST NOT be sent to /play or /display until ceremony reveal |
| Risk | If answer fields leak to /play or /display, game is compromised -- requires server-side filtering |

---

### ART-03: Team Registry

| Property | Value |
|----------|-------|
| ID | `team_registry` |
| Description | Map of team ID → team name → list of connected device tokens |
| Type | Map/object |
| Authoritative source | Server |
| Produced by | /play on team creation (US-04) |
| Consumed by | /host (lobby list, scoring interface, submission status), /display (submission count) |
| Persistence | Server in-memory; team token in player localStorage |
| Sensitive fields | None -- team names are public |
| Risk | If two devices register the same team name, duplicates appear -- requires uniqueness check on join |

---

### ART-04: Revealed Question Set

| Property | Value |
|----------|-------|
| ID | `revealed_questions` |
| Description | Ordered list of question indices that have been revealed in the current round |
| Type | Array of integers (question indices within the round) |
| Authoritative source | Server (updated by quizmaster reveal action) |
| Produced by | /host reveal action (US-07) |
| Consumed by | /play (answer form -- shows revealed questions only), /display (current question = last revealed) |
| Persistence | Server in-memory; synced to clients via WebSocket |
| Lifecycle | Reset to empty at start of each round |
| Risk | /play must show all revealed questions; /display must show only the most recently revealed -- these are different views of the same artifact |

---

### ART-05: Draft Answers

| Property | Value |
|----------|-------|
| ID | `draft_answers` |
| Description | Team's in-progress (unsaved) answer entries per question per round |
| Type | Map of question_index → answer_value (string or array for multi-part) |
| Authoritative source | Player localStorage (primary); server draft store (secondary sync) |
| Produced by | /play answer entry (US-08) |
| Consumed by | /play (review screen before submit, restore after refresh) |
| Persistence | localStorage (client-side); server draft store for cross-device sync consideration |
| Lifecycle | Overwritten on each keystroke; cleared on successful submission |
| Risk | If localStorage is cleared between refresh and rejoin, draft answers are lost -- acceptable edge case; warn user |

---

### ART-06: Submitted Answers

| Property | Value |
|----------|-------|
| ID | `submitted_answers` |
| Description | Final, locked answer set per team per round |
| Type | Map of team_id → round_number → question_index → answer_value |
| Authoritative source | Server (written on submission confirmation) |
| Produced by | /play submission (US-09) |
| Consumed by | /host scoring interface (US-12) |
| Persistence | Server in-memory |
| Lifecycle | Written on submission; immutable once written (per DEC-006) |
| Risk | If submission is acknowledged but not persisted (network failure), team loses answers -- requires server-side acknowledgment before /play shows "locked in" |

---

### ART-07: Scored Answers

| Property | Value |
|----------|-------|
| ID | `scored_answers` |
| Description | Each submitted answer with quizmaster's correct/incorrect judgment and point value |
| Type | Map of team_id → round_number → question_index → {answer, verdict, points} |
| Authoritative source | Server (written by quizmaster marking) |
| Produced by | /host scoring interface (US-12) |
| Consumed by | /host (auto-tally US-14, ceremony US-15), /display (round scores US-16), /play (ceremony view) |
| Persistence | Server in-memory |
| Lifecycle | Written during scoring; finalized when quizmaster starts ceremony |

---

### ART-08: Round Scores

| Property | Value |
|----------|-------|
| ID | `round_scores` |
| Description | Per-team point totals for each completed round + running total |
| Type | Map of team_id → array of round_score + running_total |
| Authoritative source | Server (auto-calculated from scored_answers) |
| Produced by | Auto-tally function (US-14) |
| Consumed by | /host (scores panel), /display (round scores screen US-16, final scores US-19), /play (ceremony view) |
| Persistence | Server in-memory |
| Lifecycle | Updated after each round scoring; final value = game result |

---

### ART-09: Game State

| Property | Value |
|----------|-------|
| ID | `game_state` |
| Description | Current phase of the game: LOBBY, ROUND_ACTIVE, ROUND_ENDED, SCORING, CEREMONY, ROUND_SCORES, GAME_OVER |
| Type | Enum string + current_round + current_question_index |
| Authoritative source | Server (state machine) |
| Produced by | /host actions (US-03, US-07, US-10, US-12, US-15, US-18, US-19) |
| Consumed by | /play (determines which screen to show), /display (determines which display mode to show) |
| Persistence | Server in-memory; sent to new clients on connection |
| Risk | IC-01 through IC-05 -- all integration checkpoints depend on this artifact being consistent across clients |

---

## Vocabulary Consistency Check

All three interfaces and all requirements must use the same terminology:

| Concept | Approved Term | DO NOT USE |
|---------|--------------|------------|
| Person running the game | quizmaster | host, admin, moderator, game master |
| Person playing | player | user, participant, contestant |
| Group of players | team | group, party, side |
| Single set of questions | round | level, stage, section |
| Single question | question | challenge, puzzle, item |
| Showing a question | reveal | publish, release, unlock, show |
| Team's final answers | submission | response, entry, sheet |
| Quizmaster's judgment | mark correct / mark wrong | grade, score (as verb), accept/reject |
| Points total | score | points tally, total |
| Post-round reveal | ceremony | review, walkthrough |
| TV/projector display | display | screen, TV, projector, audience view |

---

## Coherence Validation

| Check | Result |
|-------|--------|
| CLI/Web vocabulary consistent across all three journey visual files | PASS |
| Emotional arc smooth (no jarring transitions) in all three journeys | PASS |
| All shared artifacts have single authoritative source | PASS -- 9 artifacts, each with defined source |
| Answer fields never sent to /play or /display until ceremony | FLAGGED -- ART-02 requires server-side filtering (see Risk) |
| Integration checkpoints defined | PASS -- IC-01 through IC-06 in journey-trivia.yaml |
| No template variables (${variables}) without documented source | PASS -- no unresolved variables in mockups |

**Coherence Gate: PASSED with one tracked risk (ART-02 answer field filtering -- must be addressed in technical notes on relevant user stories)**
