# Trivia

A real-time trivia game server. The quizmaster loads a YAML quiz file, players join on their devices, and the game runs through question reveal, answer submission, scoring, and ceremony.

Three browser interfaces share a single server:
- `/host` — quizmaster panel (token-authenticated)
- `/play` — player answer entry
- `/display` — big-screen scoreboard and ceremony view

---

## Prerequisites

- **Go 1.23+** — `go version`
- **Node.js 18+** and **npm** — `node --version`, `npm --version`

---

## Building from scratch

The build has two steps: compile the frontend, then compile the Go binary. The frontend must be built first because the Go binary embeds the compiled assets via `go:embed`.

### 1. Build the frontend

```sh
cd frontend
npm install
npm run build
cd ..
```

This compiles the TypeScript/React app and writes the output to `internal/static/dist/`, which is where the Go embed directive reads from.

### 2. Build the Go binary

```sh
go build -o trivia ./cmd/server
```

This produces a single self-contained binary (`trivia`) with the frontend assets embedded.

---

## Running

The server requires two environment variables:

| Variable | Description |
|----------|-------------|
| `HOST_TOKEN` | Secret token that authenticates the quizmaster panel |
| `QUIZ_DIR` | Path to the directory containing your YAML quiz files |

```sh
HOST_TOKEN=mysecret QUIZ_DIR=/path/to/quizzes ./trivia
```

The server listens on port **8080**.

### Opening the interfaces

| Interface | URL | Auth |
|-----------|-----|------|
| Quizmaster panel | `http://localhost:8080/host?token=mysecret` | `?token=` query param |
| Player join | `http://localhost:8080/play` | None |
| Display screen | `http://localhost:8080/display` | None |

Share the player join URL with participants and put the display URL on a projector.

---

## Quiz file format

Quiz files are YAML and live in `QUIZ_DIR`. Example:

```yaml
title: "Friday Night Trivia"
rounds:
  - name: "Round 1 — Geography"
    questions:
      - text: "What is the capital of France?"
        answer: "Paris"
      - text: "Which country has the most natural lakes?"
        answer: "Canada"
  - name: "Round 2 — Science"
    questions:
      - text: "What is the atomic number of carbon?"
        answer: "6"
```

---

## Running the tests

```sh
# Unit tests
go test ./internal/...

# Acceptance tests (starts a real in-process server)
go test ./tests/acceptance/trivia/... -run "^TestAcceptance$" -v
```

---

## Development workflow

To iterate on the frontend without rebuilding the binary each time, run Vite's dev server alongside the Go server:

```sh
# Terminal 1 — Go backend
HOST_TOKEN=dev QUIZ_DIR=./testdata go run ./cmd/server

# Terminal 2 — Vite dev server (proxies API calls to Go)
cd frontend && npm run dev
```

To rebuild just the frontend assets and pick them up on the next Go build:

```sh
cd frontend && npm run build
```
