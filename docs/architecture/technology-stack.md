# Technology Stack -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DESIGN
- Date: 2026-03-29

---

## Backend

### Language: Go 1.23

- License: BSD 3-Clause (free)
- Rationale: Chosen by user (Q2). Excellent fit for concurrent WebSocket handling via goroutines. Strong interface system maps cleanly to ports-and-adapters. Single compiled binary simplifies Docker packaging.
- Static binary compilation enables distroless runtime image.

### HTTP Server: `net/http` (stdlib)

- License: BSD 3-Clause (Go stdlib)
- Rationale: Go's stdlib HTTP server is production-grade and sufficient for a local-network personal tool with 2-10 devices. No external framework dependency. `http.ServeMux` provides adequate routing.
- Rejected alternative: `gin` / `echo` -- unnecessary complexity for this scale. Adds a dependency without material benefit.

### WebSocket Library: `nhooyr.io/websocket` v1.x

- License: MIT
- Repository: https://github.com/nhooyr/websocket
- Rationale: Context-aware, idiomatic Go API. Does not require a separate goroutine per connection (uses `io.Reader`/`io.Writer`). Well-maintained. No CGO dependency.
- Rejected alternative: `gorilla/websocket` -- the gorilla organization is archived (no new releases since 2022). nhooyr/websocket is the actively maintained successor with a cleaner API.
- See ADR-002 for full decision.

### YAML Parser: `gopkg.in/yaml.v3`

- License: Apache 2.0 + MIT
- Repository: https://gopkg.in/yaml.v3
- Rationale: The canonical YAML library for Go. Stable, widely used, supports struct tags for clean unmarshalling.

### Architecture Enforcement: `go-arch-lint` v2.x

- License: MIT
- Repository: https://github.com/fe3dback/go-arch-lint
- Rationale: Enforces package dependency rules declared in `.go-arch-lint.yml`. Prevents answer-boundary violations at CI time.

---

## Frontend

### Language: TypeScript 5.x

- License: Apache 2.0
- Rationale: Chosen by user (Q2). Strong typing enables safe WebSocket message handling and prevents accidental use of answer fields in client code (complementary to server-side structural enforcement).

### UI Framework: React 18

- License: MIT
- Repository: https://react.dev
- Rationale: Widely understood, strong TypeScript support, sufficient for three-route SPA with real-time state. The project does not require fine-grained reactivity (Solid, Svelte) or SSR (Next.js) -- React is the simplest well-understood choice.
- Rejected alternative: Vanilla TypeScript with DOM manipulation -- manageable for a small project but React's component model makes the three-route SPA easier to maintain and extend.
- Rejected alternative: Next.js -- SSR adds complexity with no benefit for a local-network app with no SEO or initial-load performance requirements.

### Build Tool: Vite 5.x

- License: MIT
- Repository: https://vitejs.dev
- Rationale: Fast HMR for development, simple production build outputting a `dist/` directory that Go embeds. Zero-config TypeScript + React support. Replaces CRA (deprecated) and Webpack (heavier configuration).

### WebSocket Client: Native browser `WebSocket` API

- No library dependency needed. TypeScript interfaces typed to match server event protocol.

### Linting: ESLint + `@typescript-eslint`

- License: MIT
- Enforces `no-explicit-any` on WebSocket message handlers, preventing silent answer-field leakage in client code.

---

## Infrastructure / Docker

### Go Runtime Image: `gcr.io/distroless/static:nonroot`

- License: Apache 2.0 (Google distroless)
- Rationale: Minimal attack surface. Contains no shell, no package manager. Go static binaries run without libc. Non-root user by default.

### Frontend Build Image: `node:20-alpine`

- License: MIT (Alpine Linux) + MIT (Node.js)
- Rationale: Minimal Node image for the frontend Vite build stage. Alpine keeps the layer small.

### Go Build Image: `golang:1.23-alpine`

- License: BSD 3-Clause + MIT (Alpine)
- Rationale: Standard Go build image. Alpine minimizes layer size.

### Orchestration: Docker Compose v2

- License: Apache 2.0
- Rationale: Single-service compose provides volume mount configuration, environment variable injection, and port mapping without requiring Kubernetes or other orchestration overhead.

---

## Summary Table

| Component | Technology | Version | License |
|-----------|-----------|---------|---------|
| Backend language | Go | 1.23 | BSD 3-Clause |
| HTTP server | net/http (stdlib) | Go 1.23 | BSD 3-Clause |
| WebSocket (server) | nhooyr.io/websocket | v1.x | MIT |
| YAML parser | gopkg.in/yaml.v3 | v3 | Apache 2.0 / MIT |
| Architecture lint | go-arch-lint | v2.x | MIT |
| Frontend language | TypeScript | 5.x | Apache 2.0 |
| Frontend framework | React | 18 | MIT |
| Build tool | Vite | 5.x | MIT |
| WebSocket (client) | Browser native | N/A | N/A |
| Frontend lint | ESLint + @typescript-eslint | latest | MIT |
| Runtime image | distroless/static | nonroot | Apache 2.0 |
| Build image (node) | node:20-alpine | 20 | MIT |
| Build image (go) | golang:1.23-alpine | 1.23 | BSD 3-Clause |
| Orchestration | Docker Compose v2 | 2.x | Apache 2.0 |

All choices are open source. No proprietary dependencies.
