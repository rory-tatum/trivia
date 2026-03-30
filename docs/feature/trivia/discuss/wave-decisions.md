# Wave Decisions -- DISCUSS Phase

## Metadata

- Feature ID: trivia
- Phase: DISCUSS (Requirements)
- Date: 2026-03-29
- Carries forward: All DISCOVER decisions (DEC-001 through DEC-008)

---

## Inherited Decisions from DISCOVER Wave

All eight DISCOVER decisions are confirmed and carried forward. No reversals.

| Decision | Summary | Status |
|----------|---------|--------|
| DEC-001 | Personal tool, not commercial product | CONFIRMED |
| DEC-002 | YAML as primary content format | CONFIRMED |
| DEC-003 | Three-interface architecture (/host, /play, /display) | CONFIRMED |
| DEC-004 | No user accounts; in-memory state only | CONFIRMED |
| DEC-005 | Quizmaster manual scoring (no auto-grading) | CONFIRMED |
| DEC-006 | Submission is final per round | CONFIRMED |
| DEC-007 | Media files served locally relative to YAML | CONFIRMED |
| DEC-008 | Real-time sync via WebSocket | CONFIRMED |

---

## New Decisions from DISCUSS Wave

### DEC-009: Walking Skeleton is Release 1 (Text-Only)

**Date:** 2026-03-29
**Phase:** Story Mapping (Phase 2.5)
**Decision:** Release 1 implements the complete game loop with text-only questions. Media, multiple choice, and multi-part questions are deferred to Releases 3 and 4.
**Rationale:** The walking skeleton proves the real-time architecture and the core value proposition (load → join → reveal → answer → submit → score → ceremony → scores) without the additional complexity of media rendering or complex answer types. This is the most valuable deliverable slice that can be shipped first.
**Impact:** US-01 is scoped to text questions only in Release 1. YAML schema validation for media/MC/multi-part fields deferred.
**Alignment:** Confirmed by user (D2: walking skeleton as Release 1).

---

### DEC-010: Answer Fields Must Never Leave the Server to /play or /display

**Date:** 2026-03-29
**Phase:** Shared Artifact Registry (Phase 3), ART-02
**Decision:** The server must filter all `answer`/`answers` fields from quiz content before sending any question data to /play or /display clients. This applies at all game states until the ceremony reveal for that specific question.
**Rationale:** A single oversight in server-side serialization would expose correct answers to all players. This is a security invariant, not a feature -- it must be enforced at the API/serialization layer, not assumed by client behavior.
**Impact:** Solution architect must design a question payload type that structurally excludes answer fields (e.g., separate QuestionPublic and QuestionFull types). This requirement must appear in the technical handoff brief.
**Risk if violated:** All teams see answers; game integrity is destroyed.
**Alignment:** See ART-02 in shared-artifacts-registry.md.

---

### DEC-011: /display Shows Only the Most Recently Revealed Question (Not All Revealed)

**Date:** 2026-03-29
**Phase:** Journey Design (Phase 2), journey-host-visual.md
**Decision:** /display renders only the current question (last revealed). /play renders all revealed questions cumulatively. These are two different views of the same ART-04 (revealed_question_set) artifact.
**Rationale:** The display is a shared focal point for the room -- showing all 8 revealed questions at once would require tiny text and provide no value. The current question is what the room should be focused on. Players use their phones to see all revealed questions.
**Impact:** /display WebSocket handler subscribes to the same reveal events as /play but only renders the latest item. /play renders all items in sequence.
**Alignment:** Confirmed in journey-display-visual.md and shared-artifacts-registry.md.

---

### DEC-012: Submission Acknowledgment Required Before UI Locks

**Date:** 2026-03-29
**Phase:** Requirements (Phase 4), US-09
**Decision:** The /play submission UI must not show "Your answers are locked in" until the server has acknowledged the submission (written ART-06). Client must retry on network failure.
**Rationale:** If the UI shows "locked" before the server has persisted the submission, a network failure between click and server write could silently lose the team's answers. The quizmaster would open scoring with no submission from that team, and the team would believe they submitted. This is a high-trust moment for players.
**Impact:** Requires server-side acknowledgment pattern (not fire-and-forget). Client shows spinner or "Submitting..." between click and acknowledgment.
**Alignment:** See IC-03 in journey-trivia.yaml.

---

### DEC-013: Ceremony Answer Reveal Is Per-Question, Not Per-Round Batch

**Date:** 2026-03-29
**Phase:** Requirements (Phase 4), US-15
**Decision:** During the ceremony, Marcus steps through one question at a time. Each step first shows the question (without answer), then a separate action reveals the answer. This is a two-step reveal per question, not a full-round batch reveal.
**Rationale:** The two-step ceremony (question shown → pause → answer revealed) creates the theatrical suspense moment that is the emotional payoff of the round. A batch reveal would flatten this to a data table, not a performance.
**Impact:** Ceremony control panel has: "Show Q[N]" → question appears on /display → "Reveal Answer" → answer appears on /display → "Next Question" → repeat. Three distinct actions per question.
**Alignment:** Confirmed by JS-05 (showman job story) and journey-host-visual.md Screen 7.

---

### DEC-014: Release 2+ Stories Are Out of Scope for This DISCUSS Handoff

**Date:** 2026-03-29
**Phase:** Scope Assessment (Phase 2.7)
**Decision:** The DESIGN wave handoff package covers Release 1 only (US-01 through US-19). Release 2+ stories (US-06, US-11, US-17, US-20 through US-30) are written and documented but not included in the DoR gate for this handoff.
**Rationale:** Release 2+ stories depend on Release 1 being built and validated. Handing off all 30 stories at once would overload the DESIGN wave and prevent iterative delivery.
**Impact:** DESIGN wave should implement Release 1, validate the walking skeleton in a real game session, then receive the Release 2+ handoff as a second iteration.
**Alignment:** Consistent with DEC-009 (walking skeleton is Release 1).

---

## Vocabulary Decisions (Canonical Terms)

Confirmed in shared-artifacts-registry.md. All requirements, UI copy, and technical artifacts must use these terms:

| Canonical Term | Context |
|---------------|---------|
| quizmaster | Person running the game |
| player | Person playing |
| team | Group of players |
| round | Set of questions |
| question | Single question |
| reveal | Quizmaster action of showing a question |
| submission | Team's final locked answers |
| mark correct / mark wrong | Quizmaster scoring action |
| score | Points total |
| ceremony | Post-round answer reveal |
| display | The /display interface |

---

## Open Questions for DESIGN Wave

The following questions are intentionally deferred to the solution-architect (DESIGN wave). They are not requirements questions -- they are design and technology questions.

| # | Question | Why Deferred |
|---|----------|-------------|
| OQ-01 | How should the server maintain WebSocket connections at scale? (e.g., Socket.io rooms, raw WS, SSE) | Technology choice -- belongs in DESIGN |
| OQ-02 | What is the server-side state machine implementation? (class, FSM library, event sourcing) | Architecture -- belongs in DESIGN |
| OQ-03 | How is the quizmaster session protected? (URL token, simple password, no auth for localhost) | Security design -- belongs in DESIGN |
| OQ-04 | What is the reconnection backoff strategy? | Technical implementation |
| OQ-05 | What web framework and language? (Node/Express, Python/FastAPI, etc.) | Technology choice |
| OQ-06 | How are static media files served? (Express static, nginx, etc.) | Infrastructure |

---

## Handoff Package Contents

| Artifact | Location | Status |
|----------|----------|--------|
| JTBD Analysis | `discuss/jtbd-analysis.md` | COMPLETE |
| Journey Visual -- /host | `discuss/journey-host-visual.md` | COMPLETE |
| Journey Visual -- /play | `discuss/journey-play-visual.md` | COMPLETE |
| Journey Visual -- /display | `discuss/journey-display-visual.md` | COMPLETE |
| Journey YAML Schema | `discuss/journey-trivia.yaml` | COMPLETE |
| Journey Gherkin | `discuss/journey-trivia.feature` | COMPLETE |
| Story Map | `discuss/story-map.md` | COMPLETE |
| Prioritization | `discuss/prioritization.md` | COMPLETE |
| Shared Artifacts Registry | `discuss/shared-artifacts-registry.md` | COMPLETE |
| Outcome KPIs | `discuss/outcome-kpis.md` | COMPLETE |
| User Stories R1 | `discuss/user-stories-release-1.md` | COMPLETE (15 stories) |
| User Stories R2+ | `discuss/user-stories-release-2-plus.md` | COMPLETE (15 stories) |
| DoR Validation | `discuss/dor-validation.md` | COMPLETE -- ALL PASS |
| Wave Decisions | `discuss/wave-decisions.md` | THIS FILE |

**DISCUSS wave is ready for handoff to solution-architect (DESIGN wave).**
