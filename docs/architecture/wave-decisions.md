# Wave Decisions -- DESIGN Phase

## Metadata

- Feature ID: trivia
- Phase: DESIGN
- Date: 2026-03-29
- Architect: Morgan (solution-architect)
- Carries forward: DEC-001 through DEC-014 (DISCOVER + DISCUSS waves)

---

## Inherited Decisions (All Confirmed)

| Decision | Summary | Status |
|----------|---------|--------|
| DEC-001 | Personal tool, not commercial product | CONFIRMED |
| DEC-002 | YAML as primary content format | CONFIRMED |
| DEC-003 | Three-interface architecture (/host, /play, /display) | CONFIRMED |
| DEC-004 | No user accounts; in-memory state only | CONFIRMED |
| DEC-005 | Quizmaster manual scoring | CONFIRMED |
| DEC-006 | Submission is final per round | CONFIRMED |
| DEC-007 | Media files served locally relative to YAML | CONFIRMED |
| DEC-008 | Real-time sync via WebSocket | CONFIRMED |
| DEC-009 | Walking skeleton is Release 1 (text-only) | CONFIRMED |
| DEC-010 | Answer fields must never leave the server to /play or /display | CONFIRMED -- resolved via dual-type structural boundary |
| DEC-011 | /display shows only most recently revealed question | CONFIRMED |
| DEC-012 | Submission acknowledgment required before UI locks | CONFIRMED -- resolved via server ack + client retry |
| DEC-013 | Ceremony answer reveal is per-question, two-step | CONFIRMED |
| DEC-014 | Release 2+ out of scope for this handoff | CONFIRMED |

---

## New Decisions from DESIGN Wave

### DEC-015: Go + TypeScript/React Stack

**Date:** 2026-03-29
**Decision:** Go 1.23 backend + TypeScript 5 / React 18 frontend.
**Rationale:** User-specified (Q2). Go's interface system maps cleanly to ports-and-adapters. Single compiled binary + go:embed simplifies Docker packaging. TypeScript provides type safety on WebSocket message handling.
**ADR:** ADR-001

---

### DEC-016: nhooyr.io/websocket as WebSocket Library (resolves OQ-01)

**Date:** 2026-03-29
**Decision:** Use nhooyr.io/websocket v1.x for server-side WebSocket handling.
**Rationale:** gorilla/websocket organization is archived (last release 2022). nhooyr/websocket is the actively maintained, context-aware successor with idiomatic Go API.
**ADR:** ADR-002

---

### DEC-017: Single Container with Docker Compose (resolves OQ-06 deployment)

**Date:** 2026-03-29
**Decision:** Single Docker container serving both Go backend and embedded SPA. Docker Compose manages volume mount for quiz files and environment variable injection.
**Rationale:** Solo developer, personal tool, 2-10 devices on local network. Multi-container would add orchestration overhead with no benefit.
**ADR:** ADR-003

---

### DEC-018: Structural Dual-Type Answer Boundary (enforces DEC-010)

**Date:** 2026-03-29
**Decision:** Two distinct Go struct types -- QuestionFull (internal) and QuestionPublic (transport-safe). The `handler` and `hub` packages are architecturally forbidden from importing QuestionFull. Enforced by go-arch-lint in CI.
**Rationale:** Runtime filtering (stripping fields from a single type) can be bypassed by a serialization change. A structural type boundary cannot be bypassed without a compile error or a lint violation.
**ADR:** ADR-004

---

### DEC-019: Go Explicit State Machine (resolves OQ-02)

**Date:** 2026-03-29
**Decision:** Implement game state machine as a Go struct with named transition methods on GameSession. No FSM library.
**Rationale:** The state machine has 7 states and ~12 transitions -- well within the range where a hand-rolled machine is more readable than a library abstraction. The transition table is documented in data-models.md. A library would add a dependency with no clarity gain for this scale.

---

### DEC-020: URL Query Token for Quizmaster Auth (resolves OQ-03)

**Date:** 2026-03-29
**Decision:** /host requires `?token=<HOST_TOKEN>` URL parameter. Token set via environment variable at container startup. Validated by Auth Guard middleware on every request.
**Rationale:** Adequate for local network personal use. No login form complexity. Quizmaster bookmarks the full URL. Prevents accidental access by other LAN devices.

---

### DEC-021: Modular Monolith Architecture

**Date:** 2026-03-29
**Decision:** Single deployable binary. Internal module separation via Go packages with enforced dependency rules.
**Rationale:** Solo developer, simple deployment requirement, no need for independent deployability of sub-components. Modular monolith with ports-and-adapters provides testability and maintainability without microservices overhead.
**Rejected alternative:** Microservices (game state service + WebSocket gateway service) -- no justification for a solo developer personal tool.

---

### DEC-022: go:embed for Frontend Asset Serving

**Date:** 2026-03-29
**Decision:** Vite-compiled frontend `dist/` is embedded into the Go binary using go:embed directive. Go serves the SPA from memory.
**Rationale:** Single-binary deployment. No nginx or separate static server needed. No runtime filesystem dependency for frontend assets. Eliminates the "/static files not found" class of deployment error.

---

### DEC-023: Exponential Backoff Reconnection

**Date:** 2026-03-29
**Decision:** Client-side WebSocket reconnection uses exponential backoff: 1s base, 2x multiplier, 30s max, max 10 attempts.
**Rationale:** Prevents thundering herd on server restart. 10 attempts over ~5 minutes is adequate for a local-network session where brief disconnections are common (device sleeps, WiFi handoff). After 10 attempts, user sees a "disconnected" banner with a manual reconnect button.

---

## Open Questions Resolved

| OQ | Question | Resolution | Decision |
|----|----------|-----------|---------|
| OQ-01 | WebSocket library | nhooyr.io/websocket | DEC-016, ADR-002 |
| OQ-02 | State machine implementation | Hand-rolled Go struct with named transition methods | DEC-019 |
| OQ-03 | Quizmaster session protection | URL query token from env var | DEC-020 |
| OQ-04 | Reconnection backoff | Exponential: 1s base, 2x, 30s max, 10 attempts | DEC-023 |
| OQ-05 | Web framework and language | Go net/http + TypeScript/React | DEC-015, ADR-001 |
| OQ-06 | Static media file serving | net/http FileServer on /media/* backed by bind-mount volume | DEC-017, ADR-003 |

---

## Deferred to Release 3+

| Item | Status |
|------|--------|
| Multi-part answers (answers: []) | Schema parsed, validation deferred |
| Multiple choice questions | Schema parsed, validation deferred |
| Media file rendering (image/audio/video) | Architecture supports (media server component), UI deferred |
| Draft answer cross-device sync | localStorage only in Release 1; server-side draft store deferred |
