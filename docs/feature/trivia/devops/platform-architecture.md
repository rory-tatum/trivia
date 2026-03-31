# Platform Architecture -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DEVOPS
- Date: 2026-03-29
- Decisions: D1 (on-premise), D2 (Docker Compose), D6 (Recreate)

---

## Deployment Topology

Single Docker container on developer's local machine. Docker Compose provides volume mount,
environment variable injection, and port mapping. No cloud provider, no container registry,
no orchestration platform.

```
Host machine
  docker-compose.yml
  .env                    (git-ignored, contains HOST_TOKEN)
  quizzes/                (bind-mounted read-only into container)
    my-quiz.yml
    media/
      image1.jpg
  [trivia container: port 8080]
    Go binary
      + embedded React SPA (go:embed)
      + /quizzes bind mount
```

---

## Dockerfile

Multi-stage build: three stages.

```
Stage 1 (frontend):  node:20-alpine
  - npm ci
  - npm run build
  - Outputs: dist/

Stage 2 (backend):   golang:1.23-alpine
  - Copies dist/ from Stage 1 into internal/static/dist/
  - go build ./cmd/server
  - CGO_ENABLED=0, GOOS=linux, GOARCH=amd64
  - Outputs: single static binary

Stage 3 (runtime):   gcr.io/distroless/static:nonroot
  - Copies binary only
  - EXPOSE 8080
  - USER nonroot:nonroot
  - ENTRYPOINT ["/server"]
```

Key properties:
- No shell in runtime image (distroless) -- reduces attack surface
- Binary is statically compiled (no libc dependency) -- works with scratch/distroless
- Frontend assets baked into binary via go:embed -- no separate static server
- Build reproducibility: pinned base image tags in CI (use SHA digests for production hardening)

---

## docker-compose.yml

```yaml
services:
  trivia:
    build:
      context: .
      dockerfile: Dockerfile
    image: trivia:local
    ports:
      - "${PORT:-8080}:8080"
    volumes:
      - ${QUIZ_DIR:-./quizzes}:/quizzes:ro
    environment:
      - HOST_TOKEN=${HOST_TOKEN}
      - QUIZ_DIR=/quizzes
    restart: "no"
```

### Environment variables

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| `HOST_TOKEN` | Yes | none | Quizmaster authentication token. Server fails fast (exit 1) at startup if absent or empty. |
| `PORT` | No | `8080` | Host port to expose the application on. |
| `QUIZ_DIR` | No | `./quizzes` | Host directory containing YAML quiz files and media assets. Mounted read-only. |

### .env file (git-ignored)

```
HOST_TOKEN=change-me-before-use
PORT=8080
QUIZ_DIR=./quizzes
```

The `.env` file must be listed in `.gitignore`. The repository provides a `.env.example`
with the above structure and placeholder values.

---

## Volume Mount Design

The quiz directory is mounted read-only (`:ro`) at `/quizzes` inside the container.

```
Host path (configurable via QUIZ_DIR)
  quizzes/
    science-round.yml
    history-round.yml
    media/
      earth.jpg
      solar-system.mp4

Container path: /quizzes (read-only)
```

Why read-only: the server only reads quiz files. A writable mount would allow a path traversal
or file-write bug to modify the host filesystem. `:ro` eliminates this class of risk.

The quizmaster adds new YAML files to the host `quizzes/` directory. They become immediately
available without rebuilding the container. A server restart (or re-load action in the UI) picks up the new file.

---

## Deployment Strategy: Recreate

Deployment strategy is Recreate (D6). Stop the running container, start a new one with the
updated image. There is no zero-downtime requirement for a personal-use local-network tool.

**Rollback procedure** (defined before rollout, per Principle 7):

1. Identify the previously working image tag (e.g., `trivia:local` before rebuild, or a tagged release).
2. Update `docker-compose.yml` image reference to the previous tag (or re-tag: `docker tag trivia:previous trivia:local`).
3. Run: `docker compose down && docker compose up -d`
4. Verify: open browser to `http://localhost:8080` -- confirm app loads.

**Rollout procedure:**

1. Build new image: `docker compose build`
2. Stop and replace: `docker compose down && docker compose up -d`
3. Verify: open browser to `http://localhost:8080` -- confirm app loads and HOST_TOKEN is accepted.

---

## Fail-Fast Startup Validation

The Go server must validate on startup (before binding to port):

1. `HOST_TOKEN` environment variable is set and non-empty. Exit 1 with message: `FATAL: HOST_TOKEN environment variable is required`.
2. `QUIZ_DIR` environment variable is set and the path is accessible. Exit 1 with message: `FATAL: QUIZ_DIR path not accessible: /quizzes`.

This prevents silent misconfiguration where the server starts but is unusable.

---

## Network Configuration

The container exposes port 8080 internally. The host port is configurable via `PORT` env var.
All three interfaces are served from the same port:

| Path | Interface |
|------|-----------|
| `/host?token=<HOST_TOKEN>` | Quizmaster control panel |
| `/play` | Player join and answer interface |
| `/display` | Room display screen |
| `/ws` | WebSocket upgrade endpoint |
| `/media/*` | Quiz media files (images, audio, video) |
| `/` (all other) | React SPA assets (embedded) |

Players connect via `http://<host-ip>:<PORT>/play` from their devices on the same local network.
The quizmaster shares this URL verbally or via QR code.

---

## Security Properties

| Control | Implementation |
|---------|---------------|
| Quizmaster auth | `HOST_TOKEN` URL parameter validated by Auth Guard on every `/host*` request |
| Answer confidentiality | Structural dual-type boundary (QuestionFull vs QuestionPublic) enforced by go-arch-lint |
| Container user | `nonroot:nonroot` (distroless default) -- no root process |
| Quiz filesystem access | Read-only bind mount -- server cannot write to host |
| No secrets in image | `HOST_TOKEN` injected at runtime via env var, not baked into image |

---

## Local Development

For local development without Docker:

```bash
# Terminal 1: frontend dev server with HMR
cd frontend && npm run dev

# Terminal 2: backend
HOST_TOKEN=dev-token QUIZ_DIR=./quizzes go run ./cmd/server
```

Docker is used only for production-equivalent deployment.
