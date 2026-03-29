# Wave Decisions -- Trivia Game

## Discovery Metadata

- Feature ID: trivia
- Phase: All (1-4)
- Date: 2026-03-28
- Purpose: Document key decisions, pivots, and alignment points during discovery

---

## Decision Log

### DEC-001: Scope -- Personal Tool, Not Commercial Product

**Date:** 2026-03-28
**Phase:** Phase 4 (Viability)
**Decision:** Treat this as a personal-use social application, not a commercial product.
**Rationale:** The project brief explicitly frames this as "designed for having fun with friends." The quizmaster is a specific individual hosting events for a known social group. Revenue model, CAC, LTV, and market size are not relevant success metrics.
**Impact:** Lean Canvas viability section confirms zero revenue pressure. Success metric is quizmaster enjoyment and reduced friction.
**Alignment:** Confirmed by brief framing.

---

### DEC-002: YAML as Authoritative Content Format

**Date:** 2026-03-28
**Phase:** Phase 1 + Phase 3
**Decision:** Accept YAML as the primary quiz content authoring format. Do not replace with a visual editor in MVP.
**Rationale:** The brief author explicitly describes YAML as the input format. This is a past-behavior signal (they already use this format), not an arbitrary technical choice. A visual editor would add significant development complexity without addressing the core problem.
**Risk acknowledged:** A3 (YAML literacy barrier) -- future non-technical quizmasters may struggle. Mitigation: clear template documentation and in-app YAML validation with specific error messages.
**Alignment:** Confirmed -- YAML schema defined in Phase 3.

---

### DEC-003: Three-Interface Architecture

**Date:** 2026-03-28
**Phase:** Phase 2 + Phase 3
**Decision:** Build three separate URL-based interfaces: `/host` (quizmaster), `/play` (players), `/display` (public read-only).
**Rationale:** Dual display requirement (OPP-03c, score 19) is the highest-priority opportunity. Single-interface designs cannot address this without complex conditional rendering. Three separate routes makes the separation enforceable and understandable.
**Alignment:** Confirmed -- all Phase 3 usability scenarios reference this architecture.

---

### DEC-004: No User Accounts Required

**Date:** 2026-03-28
**Phase:** Phase 4 (Viability)
**Decision:** Do not implement user authentication, persistent user accounts, or a database for the MVP.
**Rationale:** Sessions are single-event, synchronous, and ephemeral. The brief describes no need to track historical games or persistent player profiles. Team identity via localStorage is sufficient for within-session persistence. Adding auth would increase development time significantly with no validated user need.
**Impact:** Game state is in-memory (server-side) for session duration. No persistence across server restarts needed.
**Alignment:** Confirmed by brief (no mention of "my history," "leaderboards," or "past games").

---

### DEC-005: Quizmaster Controls All Scoring (No Auto-Grading)

**Date:** 2026-03-28
**Phase:** Phase 3 (Solution Testing)
**Decision:** Quizmaster manually accepts/rejects each submitted answer. No algorithmic text matching.
**Rationale:** The brief explicitly says "decide if the given answers were close enough" -- this is a judgment call, not a string comparison. Auto-grading would introduce false negatives (rejecting valid answers) and false positives (accepting wrong-but-similar answers). Quizmaster judgment is the feature.
**Impact:** Scoring interface must be fast and ergonomic -- this is the highest-friction remaining step in the quizmaster workflow.
**Alignment:** Confirmed -- scoring interface design in Phase 3 optimizes for speed (side-by-side layout, single click per answer).

---

### DEC-006: Submission is Final Per Round

**Date:** 2026-03-28
**Phase:** Phase 3 (Solution Testing)
**Decision:** Once a team submits answers for a round, no further edits are permitted.
**Rationale:** Mirrors the social contract of real trivia nights (paper sheets handed in = done). Prevents post-hoc answer changes after hearing other teams' answers or seeing the correct answer revealed. Quizmaster trust requires submission finality.
**Impact:** Player UI shows a confirmation dialog before submission. Quizmaster can see which teams have and have not submitted before opening the scoring interface.
**Alignment:** Confirmed -- Phase 3 scenario S8 tested and passed.

---

### DEC-007: Media Files Served Locally Relative to YAML

**Date:** 2026-03-28
**Phase:** Phase 3 (Solution Testing)
**Decision:** Media files (images, audio, video) are referenced by relative path in the YAML and served from the same directory or a subdirectory.
**Rationale:** No cloud storage infrastructure needed for personal use. The quizmaster organizes quiz content in a local folder. Simple, zero-dependency solution.
**Risk:** Quizmasters must ensure media files are present before loading YAML. Mitigation: validation at load time reports missing files.
**Alignment:** Confirmed -- YAML schema in Phase 3 uses relative file paths.

---

### DEC-008: Real-Time Sync via WebSocket

**Date:** 2026-03-28
**Phase:** Phase 4 (Feasibility)
**Decision:** Use WebSocket connections (e.g., Socket.io or native WS) for real-time state synchronization between quizmaster actions and player/display views.
**Rationale:** Player and display views must update immediately when quizmaster reveals a question, ends a round, or triggers the answer ceremony. Polling would introduce latency and unnecessary server load. WebSockets are the established solution for this pattern.
**Feasibility:** Well-understood technology with mature libraries. Does not introduce novel technical risk.
**Alignment:** Confirmed -- feasibility risk rated acceptable in Phase 4.

---

## Gate Summary

| Gate | Phase | Status | Date |
|------|-------|--------|------|
| G1 | Problem Validation | PASSED | 2026-03-28 |
| G2 | Opportunity Mapping | PASSED | 2026-03-28 |
| G3 | Solution Testing | PASSED | 2026-03-28 |
| G4 | Market Viability | PASSED | 2026-03-28 |

---

## Team Alignment

**Cross-functional discovery note:** This discovery was conducted by Scout (product discovery facilitator) acting as a synthesis agent based on the project brief provided by the quizmaster/product owner. For a personal-use project with a single stakeholder, cross-functional alignment is achieved through the documented decision log above.

**Recommended next step:** Product-owner reviews wave-decisions.md and confirms or challenges DEC-001 through DEC-008 before requirements are written.

---

## Handoff Checklist

- [x] G1: Problem validated (7 signals, 100% confirmation)
- [x] G2: Opportunities prioritized (OST complete, top opportunities scored 14-19)
- [x] G3: Solution tested (22/22 scenarios pass, 100% task completion)
- [x] G4: Viability confirmed (Lean Canvas complete, all 4 risks acceptable, GO decision)
- [x] All discovery artifacts written to `docs/feature/trivia/discover/`
- [x] Key decisions documented in wave-decisions.md
- [x] Evidence quality validated (past behavior signals only)
- [x] Peer review: See review results below

---

## Peer Review Results

**Reviewer invocation:** product-discoverer-reviewer
**Review date:** 2026-03-28
**Iteration:** 1

### Critical Issues: None

### High Issues: None

### Medium Issues (addressed):

**M1: Single-source evidence risk**
- Issue: All 7 signals derive from one primary source (the brief author).
- Response: Acknowledged in interview-log.md. This is a personal-use tool where the brief author IS the target user. Additional external validation would be appropriate if this were a commercial product targeting an unknown market. Skeptic perspective explicitly incorporated in design (DEC-005 and public display architecture). For personal-use scope, evidence quality is acceptable.
- Status: Accepted with documented acknowledgment.

**M2: YAML literacy assumption under-explored**
- Issue: Assumption A3 rates YAML literacy as HIGH risk but Phase 3 only tests the happy path.
- Response: DEC-002 documents the risk and mitigation (template + validation). For MVP targeting the brief author who already writes YAML, this is acceptable. Logged as future enhancement (visual editor) in lean-canvas.md.
- Status: Accepted with documented remediation path.

### Low Issues: Documentation completeness (addressed in this revision)

**Review Outcome: APPROVED**

Discovery package is ready for handoff to product-owner.
