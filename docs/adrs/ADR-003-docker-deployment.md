# ADR-003: Docker Deployment Strategy

## Status

Accepted

## Date

2026-03-29

## Context

The trivia game must run locally on a developer's machine and be accessible to 2-10 devices on the local network. The developer specified Docker from the start for testing and portability (discovery Q1).

Key constraints:
- Solo developer
- Personal use, local network
- Quiz files (YAML + media) live on the host filesystem and must be accessible inside the container
- The quizmaster token must be configurable without rebuilding the image
- Frontend assets are compiled TypeScript/React; they must be built before the Go binary

## Decision

**Single container** running the Go binary (which embeds the compiled frontend via `go:embed`). **Docker Compose** provides volume mount, port mapping, and environment variable injection. **Multi-stage Dockerfile** separates frontend build, Go build, and runtime layers.

### Build stages

**Stage 1 -- Frontend build:** `node:20-alpine`. Runs `vite build`, produces `dist/`.

**Stage 2 -- Go build:** `golang:1.23-alpine`. Copies `dist/` from Stage 1 into `internal/static/`. Runs `go build`, produces a statically linked binary.

**Stage 3 -- Runtime:** `gcr.io/distroless/static:nonroot`. Copies binary only. Exposes port 8080. Runs as non-root.

### Docker Compose service

```yaml
services:
  trivia:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./quizzes:/quizzes:ro
    environment:
      - HOST_TOKEN=${HOST_TOKEN}
      - QUIZ_DIR=/quizzes
```

`.env` file (git-ignored) holds `HOST_TOKEN=<value>`.

## Alternatives Considered

### Alternative A: Multi-container (Go server + nginx for static files)

- **Pro:** Separation of concerns between API server and static file serving. nginx is highly optimized for static content.
- **Con:** Two containers for a personal tool serving 2-10 devices adds orchestration complexity with no measurable performance benefit. Static files are served from Go memory via `go:embed`; this is entirely adequate for the scale.
- **Rejected:** Over-engineering for the use case.

### Alternative B: Run Go binary directly (no Docker)

- **Pro:** Simplest possible setup. No Docker required.
- **Con:** Developer specified Docker from the start for testing and portability. Without Docker, the build environment must be replicated on every machine. go:embed still requires the build step; Docker makes the build reproducible.
- **Rejected:** User specified Docker requirement.

### Alternative C: Multi-container with a separate database container

- **Pro:** Data persistence across server restarts.
- **Con:** DEC-004 explicitly decided in-memory state only. A database container contradicts this decision and adds significant complexity for a personal tool.
- **Rejected:** Contradicts DEC-004.

## Consequences

### Positive

- Single `docker compose up` command to start the game
- Reproducible builds: same Dockerfile on any machine produces identical binary
- Distroless runtime has no shell, no package manager -- minimal attack surface
- Volume mount is read-only (`:ro`) -- server cannot modify quiz files
- `HOST_TOKEN` never baked into the image; always injected at runtime

### Negative

- Build time is longer (frontend Vite build + Go build both in Dockerfile)
- Developer must have Docker installed; cannot `go run .` directly without pre-building the frontend (or using a dev mode that serves frontend separately via Vite's dev server)

### Development workflow note

For local development without Docker: Vite dev server runs on port 5173 (proxying `/api/*` and `/ws` to Go on port 8080). Go server runs directly with `go run ./cmd/server`. The Go binary detects a `DEV_MODE=true` environment variable and does not try to serve embedded static files (defers to Vite).

## Compliance

- All Docker base images use permissive licenses (MIT, Apache 2.0, BSD 3-Clause)
- No proprietary container registry required (all images from Docker Hub / gcr.io public)
