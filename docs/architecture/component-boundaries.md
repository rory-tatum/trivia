# Component Boundaries -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DESIGN
- Date: 2026-03-29

---

## Paradigm

Ports-and-adapters (hexagonal architecture) with OOP. Go interfaces define ports; structs implement adapters. The domain core (`game` package) has zero imports from infrastructure packages (`handler`, `hub`, `yaml`, `http`).

Dependency rule: all dependencies point inward toward the domain.

```
[handler] --> [game]
[hub]     --> [game]
[yaml]    --> [game] (produces domain types)
[game]    --> (no external imports)
```

---

## Go Package Structure

```
trivia/
  cmd/
    server/
      main.go          -- entry point: reads env, wires dependencies, starts server
  internal/
    game/
      game.go          -- Game struct: state machine, in-memory store (domain core)
      state.go         -- GameState enum and transition rules
      types.go         -- QuestionFull, QuizFull (server-internal, never exported to transport)
      public_types.go  -- QuestionPublic, QuizPublic (safe for client transport)
      boundary.go      -- StripAnswers(QuestionFull) -> QuestionPublic  [Answer Boundary Enforcer]
      scoring.go       -- scoring logic: tally, verdict application
    hub/
      hub.go           -- WebSocket Hub: connection registry, room management, broadcast
      client.go        -- Client struct: wraps nhooyr/websocket connection, room assignment
      events.go        -- server-side event types (outbound JSON shapes)
      messages.go      -- client-side message types (inbound JSON shapes)
    handler/
      host.go          -- HTTP handlers for /host routes (auth-guarded)
      play.go          -- HTTP handlers for /play routes
      display.go       -- HTTP handlers for /display route
      ws.go            -- WebSocket upgrade handler (routes to hub)
      auth.go          -- Auth Guard middleware: validates HOST_TOKEN on /host routes
    quiz/
      loader.go        -- YAML file reading, parsing, validation
      validator.go     -- structural validation rules (required fields, media path checks)
      schema.go        -- Go structs for YAML unmarshalling (raw, before domain mapping)
    media/
      server.go        -- wraps net/http.FileServer for /media/* route
    static/
      embed.go         -- go:embed directive for compiled frontend dist/
  config/
    config.go          -- reads HOST_TOKEN and QUIZ_DIR from environment
```

### Dependency rules enforced by go-arch-lint

| Package | May import | Must NOT import |
|---------|-----------|-----------------|
| `game` | stdlib only | `handler`, `hub`, `quiz`, `media`, `config` |
| `hub` | `game` (public types only), stdlib, nhooyr/websocket | `handler`, `quiz`, `media` |
| `handler` | `game` (public types only), `hub`, `quiz`, stdlib | `game.QuestionFull`, `game.QuizFull` |
| `quiz` | `game` (domain types), stdlib, yaml.v3 | `handler`, `hub`, `media` |
| `media` | stdlib | all internal packages |
| `config` | stdlib | all internal packages |
| `cmd/server` | all internal packages | (entry point -- no restriction) |

The critical rule: `handler` and `hub` must never reference `game.QuestionFull` or `game.QuizFull`. Violation exposed only `QuestionPublic` shapes.

---

## Ports (Go Interfaces)

These interfaces define the boundaries between domain and infrastructure. The `game` package defines the ports; infrastructure packages implement or consume them.

### GamePort (defined in `game` package, consumed by `handler`)

Describes all state-mutation operations the quizmaster can trigger:

- Load quiz from validated content
- Start a round
- Reveal a question
- Force end of round
- Mark an answer verdict
- Start ceremony
- Advance ceremony (show question / reveal answer)
- Publish round scores
- End game

### StateReader (defined in `game` package, consumed by `hub`)

Describes read operations for generating state snapshots:

- Current game state enum
- Current round and question index
- Team registry
- Revealed question set (as QuestionPublic)
- Submission status per team
- Round scores

### QuizLoader (defined in `quiz` package boundary, consumed by `handler`)

- LoadFromPath(path string) -> (QuizFull, error)
- Returns domain error types for specific validation failures (missing field, missing file, file not found)

---

## TypeScript Module Structure

```
frontend/
  src/
    main.tsx           -- React entry point, router setup
    routes/
      Host.tsx         -- /host route: loading screen, lobby, reveal controls, scoring, ceremony
      Play.tsx         -- /play route: join form, answer form, submission, ceremony view
      Display.tsx      -- /display route: current question display, ceremony, scores
    components/
      host/
        LoadQuiz.tsx
        Lobby.tsx
        RevealPanel.tsx
        ScoringPanel.tsx
        CeremonyControl.tsx
        ScoresPanel.tsx
      play/
        JoinForm.tsx
        AnswerForm.tsx
        SubmitButton.tsx
        CeremonyView.tsx
      display/
        QuestionDisplay.tsx
        CeremonyDisplay.tsx
        ScoresDisplay.tsx
      shared/
        ConnectionStatus.tsx
        ErrorBanner.tsx
    ws/
      client.ts        -- WebSocket connection management, reconnection backoff
      events.ts        -- TypeScript types for all server->client events
      messages.ts      -- TypeScript types for all client->server messages
    store/
      gameState.ts     -- React context or lightweight state: current game state, teams, questions
    hooks/
      useWebSocket.ts  -- hook: connection lifecycle, event dispatch
      useGameState.ts  -- hook: game state reads
```

### TypeScript module dependency rules

| Module | May import | Must NOT import |
|--------|-----------|-----------------|
| `routes/*` | `components/*`, `hooks/*`, `store/*` | `ws/` directly (use hooks) |
| `components/*` | `hooks/*`, `store/*`, `components/shared/*` | `ws/` directly |
| `hooks/*` | `ws/*`, `store/*` | `routes/*`, `components/*` |
| `ws/*` | `ws/events.ts`, `ws/messages.ts` | all others |
| `store/*` | stdlib/react only | `ws/*`, `routes/*`, `components/*` |

Enforced by ESLint `import/no-restricted-paths` plugin.

---

## Answer Field Boundary -- TypeScript Side

The TypeScript event types defined in `ws/events.ts` must declare the `question_revealed` event payload using a `QuestionPublic` interface that has no `answer` or `answers` fields. The `CeremonyRevealEvent` declares a separate `answer: string` field.

This mirrors the Go structural boundary and means TypeScript code cannot accidentally render an answer field that doesn't exist in the type.

---

## WebSocket Room Enumeration

The Hub manages three named rooms. Each WebSocket client is assigned to exactly one room on connection.

| Room name | Constant | Connected by | Receives |
|-----------|----------|-------------|---------|
| `host` | `RoomHost` | Quizmaster (/host route, token validated) | All events + scoring payloads |
| `play` | `RoomPlay` | Players (/play route) | Game state events, question_revealed (public only), submission_ack |
| `display` | `RoomDisplay` | Display screen (/display route) | Game state events, question_revealed (public only), ceremony events |

Broadcasting rules:
- `host_*` events: sent to `host` room only
- `question_revealed`: sent to `play` and `display` rooms, payload is `QuestionPublic`
- `ceremony_answer_revealed`: sent to `display` room ONLY (not `play`)
- `state_snapshot`, `round_scores_published`, `game_over`: sent to all three rooms
- `error`: sent to the originating client connection only

---

## Adapter Map

| Port | Adapter | Package |
|------|---------|---------|
| Quiz content loading | `YAMLLoader` struct | `quiz` |
| WebSocket transport | `nhooyr/websocket` wrapped in `Client` | `hub` |
| Static file serving | `net/http.FileServer` wrapped in `MediaServer` | `media` |
| Frontend asset serving | `net/http` FileServer + `go:embed` | `static` |
| Configuration | `os.Getenv` wrapped in `Config` struct | `config` |
