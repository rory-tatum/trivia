# Wave Decisions -- DEVOPS Phase

## Metadata

- Feature ID: trivia
- Phase: DEVOPS
- Date: 2026-03-29
- Carries forward: DEC-001 through DEC-023 (DISCOVER + DISCUSS + DESIGN waves)

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
| DEC-010 | Answer fields must never leave the server to /play or /display | CONFIRMED |
| DEC-011 | /display shows only most recently revealed question | CONFIRMED |
| DEC-012 | Submission acknowledgment required before UI locks | CONFIRMED |
| DEC-013 | Ceremony answer reveal is per-question, two-step | CONFIRMED |
| DEC-014 | Release 2+ out of scope for this handoff | CONFIRMED |
| DEC-015 | Go 1.23 + TypeScript/React 18 stack | CONFIRMED |
| DEC-016 | nhooyr.io/websocket as WebSocket library | CONFIRMED |
| DEC-017 | Single container with Docker Compose | CONFIRMED |
| DEC-018 | Structural dual-type answer boundary (QuestionFull / QuestionPublic) | CONFIRMED |
| DEC-019 | Go explicit state machine (no FSM library) | CONFIRMED |
| DEC-020 | URL query token for quizmaster auth | CONFIRMED |
| DEC-021 | Modular monolith architecture | CONFIRMED |
| DEC-022 | go:embed for frontend asset serving | CONFIRMED |
| DEC-023 | Exponential backoff reconnection (1s base, 2x, 30s max, 10 attempts) | CONFIRMED |

---

## New Decisions from DEVOPS Wave

### DEC-024: GitHub Actions as CI/CD Platform

**Date:** 2026-03-29
**Decision:** Use GitHub Actions for the CI/CD pipeline.
**Rationale:** D3 (user decision). The repository is already on GitHub. GitHub Actions requires no separate CI server, has native integration with the repository, and the free tier is sufficient for a solo developer project.
**Rejected alternatives:**
- GitLab CI: requires migrating the repository or using a separate GitLab instance.
- Jenkins: requires a persistent server, ongoing maintenance overhead with zero benefit for a personal tool.
- CircleCI/Travis CI: free tier restrictions make them less suitable than GitHub's native offering.

---

### DEC-025: Trunk-Based Development

**Date:** 2026-03-29
**Decision:** All development on `main`. No long-lived feature branches. Short-lived local branches squash-merged to `main`.
**Rationale:** D8 (user decision). Optimal for a solo developer. Eliminates merge overhead, keeps CI signal immediate, and maintains a single deployable artifact on `main` at all times.
**Rejected alternatives:**
- GitHub Flow (feature branches + PRs): adds PR overhead with no benefit for a solo developer.
- GitFlow: long-lived develop/release/hotfix branches are excessive complexity for a personal project.

---

### DEC-026: On-Premise / Self-Hosted Deployment Target

**Date:** 2026-03-29
**Decision:** The application runs on the developer's own machine (or a local server). No cloud provider.
**Rationale:** D1 (user decision). Personal tool, local network use. No public internet exposure required. No cloud costs.
**Impact:** No container registry push in CI (image is exported as a tar artifact). No cloud deployment step. Manual deployment via `docker compose down && docker compose up -d`.

---

### DEC-027: Recreate Deployment Strategy

**Date:** 2026-03-29
**Decision:** Stop the running container, start a new one. No rolling update, no blue-green, no canary.
**Rationale:** D6 (user decision). Personal tool with no uptime SLA. Brief downtime during deployment is acceptable. Simplest possible deployment with the least failure modes.
**Rollback:** Re-tag the previous Docker image and run `docker compose down && docker compose up -d`. Full rollback procedure documented in `platform-architecture.md`.
**Rejected alternatives:**
- Blue-green: requires two container instances; no uptime justification for a personal tool.
- Canary: requires traffic splitting infrastructure; massively over-engineered for 2-10 devices.
- Rolling update: not applicable to a single-instance Compose service.

---

### DEC-028: Per-Feature Mutation Testing with gremlins

**Date:** 2026-03-29
**Decision:** Mutation testing runs after each feature delivery, scoped to Go files changed in the push. Tool: gremlins (Apache 2.0). Kill rate gate: >= 80%. Gate is a soft failure (reported, does not block the image artifact).
**Rationale:** D9 (user decision). Per-feature strategy balances quality signal with pipeline speed for a sub-50k LOC project. gremlins is the most actively maintained Go mutation testing tool with a clean CLI interface.
**Kill rate rationale:** 80% kill rate means at most 1 in 5 mutants survives. For a game state machine with a rich transition graph, 80% is achievable without excessive test overhead.
**Soft gate rationale:** The developer is the sole contributor. A hard gate would block deployment for their own project. The soft gate provides the signal; the developer decides when to address escaped mutants.

---

### DEC-029: No Container Registry Push

**Date:** 2026-03-29
**Decision:** CI builds the Docker image and exports it as a tarball artifact. No push to Docker Hub, GHCR, or any registry.
**Rationale:** On-premise deployment (DEC-026). The developer runs `docker compose build` locally for deployment. The CI-built tarball is a verification artifact only -- proof that the image builds cleanly from `main`.
**Future path:** If the tool is ever deployed to a remote server, add a `docker push` step targeting GHCR (GitHub Container Registry) with `GITHUB_TOKEN` authentication. This requires only one new CI step and no architectural changes.

---

### DEC-030: Defer Observability

**Date:** 2026-03-29
**Decision:** No metrics backend, no tracing, no log aggregation configured in this delivery.
**Rationale:** D5 (user decision). Personal tool. The developer can use `docker compose logs` for debugging. The KPI instrumentation design is documented in `kpi-instrumentation.md` for future implementation.
**Future path:** Add `log/slog` JSON structured logging to the Go server (zero external dependency). Mount a log volume in Docker Compose. Use `jq` for ad-hoc analysis. Full design in `kpi-instrumentation.md`.

---

### DEC-031: go-arch-lint as Architecture Enforcement CI Gate

**Date:** 2026-03-29
**Decision:** `go-arch-lint check` is a hard-failure CI gate in the `backend-checks` job. A violation blocks image build.
**Rationale:** The answer-boundary invariant (DEC-010, DEC-018) is the most critical security property of the system. Automated enforcement in CI ensures it cannot be violated by an accidental import. Failing fast at CI prevents a vulnerable image from ever being built.
**Tool choice:** go-arch-lint v2.x (MIT license). Lightweight, zero-runtime, reads `.go-arch-lint.yml` at repository root. No alternative with comparable Go-native package-dependency enforcement was found.

---

### DEC-032: Conventional Commits (Developer Discipline Only)

**Date:** 2026-03-29
**Decision:** Commit messages follow Conventional Commits format by developer convention. No automated enforcement (no commitlint, no commit-msg hook).
**Rationale:** Solo developer. Automated enforcement adds tooling overhead with no collaboration benefit. Conventional Commits are documented in `branching-strategy.md` as the expected format.

---

### DEC-033: Pre-commit and Pre-push Local Quality Hooks

**Date:** 2026-03-29
**Decision:** Provide documented `.git/hooks/pre-commit` and `.git/hooks/pre-push` scripts that mirror CI checks locally.
**Rationale:** Shift-left quality (Principle 10). Catching `go vet`, `go test`, and `go-arch-lint` failures locally is faster than waiting for GitHub Actions to report them. Pre-commit runs the fast checks (no race detector); pre-push runs the full suite.
**Installation:** Manual (`chmod +x .git/hooks/pre-commit`). Not enforced via git config (developer opt-in).

---

## Decision Log Summary

| Decision | Category | Scope |
|----------|----------|-------|
| DEC-024 | CI/CD platform | GitHub Actions |
| DEC-025 | Branching | Trunk-based development |
| DEC-026 | Deployment target | On-premise / self-hosted |
| DEC-027 | Deployment strategy | Recreate |
| DEC-028 | Quality | Per-feature mutation testing (gremlins, >= 80% kill rate) |
| DEC-029 | Infrastructure | No container registry push |
| DEC-030 | Observability | Deferred -- KPI instrumentation design documented |
| DEC-031 | Quality gate | go-arch-lint as hard CI gate |
| DEC-032 | Conventions | Conventional Commits by discipline |
| DEC-033 | Quality | Local pre-commit/pre-push hooks |
