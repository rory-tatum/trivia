# Wave Decisions — Host UI (Quizmaster Panel)

**Feature ID**: host-ui
**Date**: 2026-04-09
**Wave**: PRODUCT OWNER (requirements)
**Next wave**: DESIGN (solution-architect)

---

## Decisions Made in This Wave

### WD-01: JTBD skipped by explicit instruction
**Decision**: Jobs-to-be-done analysis was skipped per feature brief instruction.
**Rationale**: The job is unambiguous — "run a trivia game session" — and the RCA provides sufficient problem definition. JTBD would not surface new requirements.

### WD-02: UX research depth is lightweight
**Decision**: No user interviews conducted. Journey mapped from project brief + RCA + existing code.
**Rationale**: Brownfield project; the persona (Marcus, quizmaster) is the product author. RCA provides concrete evidence of all failure paths.

### WD-03: Walking skeleton excludes ceremony (US-05)
**Decision**: US-05 (answer ceremony) is Release 1, not walking skeleton.
**Rationale**: A complete game is playable without the visual ceremony — scores can be verbally announced. The ceremony also depends on the display route functioning, which is a separate feature concern.

### WD-04: IC-4 is a DESIGN wave open item
**Decision**: How the scoring panel receives team submitted answers is deferred to DESIGN wave.
**Rationale**: The user-facing behavior is clear (Marcus sees each team's answer text). The mechanism — state snapshot, new event, or existing event extension — is a solution design concern.

### WD-05: Route alias (D root cause) is a DESIGN wave concern
**Decision**: The PO wave recommends a catch-all `<Route path="*">` with a "not found" message as the minimum fix. The decision of whether to also add `/host` as a named route is deferred to DESIGN wave.
**Rationale**: Not user-journey-blocking (correct URL is `/?token=...`); it is a navigational safety net.

### WD-06: "End Game" available any time, not gated on all rounds
**Decision**: Marcus can click "End Game" after any round is complete, not only after the final round.
**Rationale**: Real-world trivia nights sometimes cut short. The behavior requirement is: Marcus has agency over when the game ends. The server enforces any hard game state constraints.

### WD-07: Scoring panel "Publish Scores" is not gated on all verdicts applied
**Decision**: Marcus can publish scores even if not all team-answer pairs have been marked.
**Rationale**: Marcus is the judge; he may choose to leave some answers unmarked (e.g., team did not submit). The tool must not block him.

---

## Constraints Passed to DESIGN Wave

1. `Host.tsx` communicates only via WsClient (WebSocket) — no REST calls
2. No new IncomingMessage event types may be added without a backend change
3. The `QuizFull`/`QuestionFull` types must not appear in handler or hub packages — scoring panel data must come from existing event payloads or a new server-side event that does not expose internal types
4. The verdict values accepted by `host_mark_answer` are `"correct"` and `"wrong"` (from `game.Verdict` type)
5. The display route (`/display`) is a separate feature; US-05 depends on it being functional

---

## Handoff Package Contents

| Artifact | Path | Purpose |
|---|---|---|
| Journey visual | docs/feature/host-ui/discuss/journey-host-visual.md | ASCII flow, emotional arc, TUI mockups |
| Journey schema | docs/feature/host-ui/discuss/journey-host.yaml | Structured step definitions, shared artifacts |
| Gherkin feature | docs/feature/host-ui/discuss/journey-host.feature | Acceptance scenarios |
| Shared artifacts | docs/feature/host-ui/discuss/shared-artifacts-registry.md | State variables and integration checkpoints |
| Story map | docs/feature/host-ui/discuss/story-map.md | Backbone, walking skeleton, release slices |
| Prioritization | docs/feature/host-ui/discuss/prioritization.md | Delivery sequence, risks |
| Requirements | docs/feature/host-ui/discuss/requirements.md | Functional + non-functional requirements |
| User stories | docs/feature/host-ui/discuss/user-stories.md | 6 LeanUX stories with domain examples |
| Acceptance criteria | docs/feature/host-ui/discuss/acceptance-criteria.md | Testable Given/When/Then per story |
| DoR checklist | docs/feature/host-ui/discuss/dor-checklist.md | All 6 stories pass; 3 open items for DESIGN |
| Outcome KPIs | docs/feature/host-ui/discuss/outcome-kpis.md | 7 KPIs with measurable targets |
| Wave decisions | docs/feature/host-ui/discuss/wave-decisions.md | This file |
