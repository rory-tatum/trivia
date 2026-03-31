# ADR-001: Go + TypeScript/React Technology Stack

## Status

Accepted

## Date

2026-03-29

## Context

The trivia game requires a backend server that:
- Handles concurrent WebSocket connections from 2-10 devices
- Maintains in-memory game state for a single active session
- Reads and validates YAML files from the local filesystem
- Serves static frontend assets

The frontend requires:
- Three distinct URL-based interfaces (/host, /play, /display)
- Real-time state updates via WebSocket
- Type-safe message handling to enforce the answer-field confidentiality boundary (DEC-010)

The developer (solo) specified Go for the backend and TypeScript for the frontend (discovery Q2). The developer is familiar with Go.

## Decision

**Backend:** Go 1.23 using the standard library (`net/http` for HTTP routing and serving, `nhooyr.io/websocket` for WebSocket, `gopkg.in/yaml.v3` for YAML parsing).

**Frontend:** TypeScript 5 with React 18, built with Vite 5, served as a compiled SPA embedded in the Go binary via `go:embed`.

**Paradigm:** OOP with ports-and-adapters (hexagonal architecture). Go interfaces define ports. Structs implement adapters. Domain core (`game` package) has no imports from infrastructure packages.

## Alternatives Considered

### Alternative A: Node.js/TypeScript fullstack (e.g., Fastify + React)

- **Pro:** Single language across stack. Large ecosystem. Socket.io provides battle-tested WebSocket rooms.
- **Con:** Developer is not specified as experienced with Node. Go was explicitly chosen. Runtime requires Node in the container vs a single static Go binary. TypeScript strictness at the Node layer is opt-in and requires discipline.
- **Rejected:** Developer preference is Go backend. Go's compiled binary + go:embed produces a simpler Docker image.

### Alternative B: Python/FastAPI + React

- **Pro:** Fast prototyping. FastAPI is excellent for APIs.
- **Con:** Developer specified Go. Python async WebSocket handling (via `websockets` or `starlette`) is more complex than Go goroutines for concurrent connections. Runtime container needs Python runtime.
- **Rejected:** Not aligned with developer preference or skills.

### Alternative C: Go backend + Vanilla TypeScript (no React)

- **Pro:** Fewer dependencies. No framework overhead.
- **Con:** Three-route SPA with real-time state (game state machine reflected in UI) is significantly harder to maintain without a component model. React's unidirectional data flow is well-suited to state-machine-driven UIs.
- **Rejected:** React's component model provides clarity for the three-route SPA structure with acceptable complexity for a solo developer.

## Consequences

### Positive

- Single compiled binary with embedded frontend simplifies Docker image (distroless runtime, no extra layers)
- Go's interface system maps directly to ports-and-adapters
- TypeScript's structural typing enables enforcement of answer-field boundary at the type level
- Developer familiarity with Go reduces onboarding friction

### Negative

- Dual build step (Vite then Go) required in Dockerfile; CI must run both
- Developer must maintain TypeScript tooling (Node, npm/pnpm) in addition to Go toolchain

## Compliance

- All components are open source with permissive licenses (BSD-3, MIT, Apache 2.0)
- No proprietary dependencies
