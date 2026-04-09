# Root Cause Analysis: WebSocket Upgrade Failure

**Error**: `WebSocket protocol violation: Connection header "keep-alive" does not contain Upgrade`
**Date**: 2026-04-06
**Analyst**: Rex (Toyota 5 Whys RCA)
**Scope**: Go backend (`cmd/server/main.go`), TypeScript/React frontend (`frontend/src/routes/`)

---

## Problem Statement

Pages (`/play`, `/display`) do not function. When the browser navigates to these routes and the
React SPA attempts to open a WebSocket connection, the server returns a protocol violation error
because the HTTP request arriving at `websocket.Accept` carries `Connection: keep-alive` instead
of `Connection: Upgrade`. The upgrade handshake is therefore rejected.

---

## Evidence Collected

| Evidence ID | Source | Finding |
|-------------|--------|---------|
| E1 | `cmd/server/main.go` (uncommitted diff) | Previous committed version had `mux` and `cfg` stubbed out (`_ = mux`; `_ = cfg`). All routing logic is new in the working tree. |
| E2 | `cmd/server/main.go` lines 59-71 | `/ws` handler dispatches on `?token` (host) or `?room=play`/`?room=display`. Play and display paths require `?room=` param. |
| E3 | `frontend/src/routes/Play.tsx` line 7 | `getPlayWsUrl()` returns `` `${protocol}//${window.location.host}/ws` `` — no `?room=` param. |
| E4 | `frontend/src/routes/Display.tsx` line 7 | `getDisplayWsUrl()` returns `` `${protocol}//${window.location.host}/ws` `` — no `?room=` param. |
| E5 | `cmd/server/main.go` line 68 | When neither `?token` nor `?room` match, handler calls `http.Error(w, "invalid room", http.StatusBadRequest)` — this is a plain HTTP response, not a WebSocket upgrade. |
| E6 | `frontend/src/routes/Host.tsx` line 9 | Host correctly appends `?token=<token>` — host route is unaffected. |
| E7 | `internal/static/embed.go` lines 21-24 | Static handler only calls `http.ServeFileFS(..., "index.html")` for `r.URL.Path == "/"`. All other paths pass to `http.FileServer`. |
| E8 | `internal/static/dist/` contents | `dist/` contains only `index.html` and `assets/`. No `play` or `display` files exist. |
| E9 | `cmd/server/main.go` line 58 | `mux.Handle("/", static.NewStaticHandler())` — the `/` catch-all is registered on the mux. |
| E10 | `frontend/src/main.tsx` lines 10-16 | `BrowserRouter` with `<Route path="/play">` and `<Route path="/display">` — HTML5 client-side routing. Server must serve `index.html` for these paths. |
| E11 | `tests/acceptance/trivia/steps/driver.go` lines 110, 122 | Test driver correctly appends `?room=display` and `?room=play` — tests pass but production frontend does not. |
| E12 | `internal/handler/play_test.go` line 39 | Unit tests also use `?room=play` — test-to-production divergence. |
| E13 | Binary `server` and `trivia` in repo root | Both are identical (`md5sum` match). The deployed binary predates the new routing code in `main.go` (uncommitted changes). |
| E14 | `go build ./cmd/server/` succeeds | New code compiles without error; the binary discrepancy is a deployment gap, not a compile error. |

---

## Causal Branches

Two independent causal branches produce the failure:

- **Branch A**: `Play.tsx` / `Display.tsx` omit `?room=` query param → server's `/ws` handler falls through to `http.Error` → plain HTTP 400 response arrives at browser's `WebSocket` constructor → browser reports Connection header violation.
- **Branch B**: Static handler returns 404 for `/play` and `/display` (no SPA fallback) → SPA never loads → no WebSocket connection attempt is even possible for a cold navigation.

Branch A is the primary cause of the error message. Branch B is a compounding cause that blocks cold navigation entirely.

---

## Toyota 5 Whys Analysis

### Branch A — Missing `?room=` Parameter

**WHY 1 (Symptom)**: Browser reports `Connection header "keep-alive" does not contain Upgrade`.
Evidence: E5 — when neither `?token` nor `?room` are present, the Go handler responds with `http.Error(..., 400)`. The browser's `WebSocket` constructor receives a plain HTTP 400. The browser interprets this as the server having responded without upgrading, yielding the Connection-header protocol violation.

**WHY 2 (Context)**: The Go `/ws` handler receives a request without `?room=play` or `?room=display`.
Evidence: E3, E4 — `getPlayWsUrl()` and `getDisplayWsUrl()` both return `/ws` with no query parameters. The browser's `WebSocket(url)` call sends that URL verbatim.

**WHY 3 (System)**: The frontend URL-builder functions were written without knowledge of the server's room-dispatch contract.
Evidence: E2, E11 — the server contract (dispatch on `?room=`) is implemented in `cmd/server/main.go` and is exercised correctly by acceptance tests and unit tests. The frontend functions in `Play.tsx` and `Display.tsx` were never wired to match that contract.

**WHY 4 (Design)**: No shared specification or type-safe contract exists between frontend and backend for the WebSocket URL shape.
Evidence: E11 vs E3/E4 — test driver and frontend independently construct URLs. The test driver gets it right; the SPA gets it wrong. There is no single source of truth (no OpenAPI spec, no generated client, no documented constant) that both reference.

**WHY 5 (Root Cause A)**: The WebSocket connection protocol (URL parameters, room dispatch) is implicit, undocumented, and not enforced at the boundary between backend and frontend. Each implementation point authors the URL from memory, with no compilation or test gate to catch divergence.

---

### Branch B — SPA Fallback Missing from Static Handler

**WHY 1 (Symptom)**: Navigating directly to `/play` or `/display` returns 404 (or serves a file-server 404 page) instead of `index.html`.
Evidence: E7, E8 — static handler only returns `index.html` for `r.URL.Path == "/"`. For `/play`, the `http.FileServer` attempts to serve `dist/play`, which does not exist.

**WHY 2 (Context)**: `NewStaticHandler` delegates non-root paths to `http.FileServer` without a fallback.
Evidence: E7 — the handler has an explicit `"/"` case but no catch-all for unrecognised paths that should resolve to `index.html`.

**WHY 3 (System)**: The handler was implemented as a traditional file server rather than as an SPA host.
Evidence: E10 — `main.tsx` uses `BrowserRouter`, which requires the server to serve `index.html` for any path the SPA owns, so client-side routing can take over. A traditional file server cannot satisfy this requirement.

**WHY 4 (Design)**: No test asserts that `/play` or `/display` returns a 200 with HTML content.
Evidence: `embed_test.go` — only `TestStaticHandler_ServesRootWithHTML` exists; it tests `/` only. The contract that SPA sub-routes must be served was never captured as a failing test.

**WHY 5 (Root Cause B)**: The SPA hosting requirement (serve `index.html` for all client-side routes) was not defined as a testable acceptance criterion when the static handler was written, allowing the implementation gap to survive undetected.

---

## Backwards Chain Validation

**Branch A forward trace**: Root Cause A (no shared URL contract) → Play.tsx and Display.tsx build `/ws` with no `?room=` → Go handler hits `else` branch → `http.Error(400)` returned → browser WebSocket receives HTTP 400 → reports "Connection header does not contain Upgrade". Validates against E2, E3, E4, E5. Chain is consistent.

**Branch B forward trace**: Root Cause B (no SPA fallback test) → static handler only maps `/` to index.html → `/play` request hits `http.FileServer` → no `dist/play` file → 404 response → React SPA never loads → no WebSocket connection made. Validates against E7, E8, E10. Chain is consistent.

The two chains are independent. Branch A explains the specific error message. Branch B explains why cold navigation to `/play` or `/display` produces a blank page with no WebSocket attempt at all.

---

## Solution Map

Every solution is labelled: **[IMMEDIATE]** = restores service without architectural change; **[PERMANENT]** = prevents recurrence.

### Fix A1 — Add `?room=` to Play and Display WebSocket URLs [IMMEDIATE + PERMANENT]

**File**: `frontend/src/routes/Play.tsx`

Change `getPlayWsUrl`:
```typescript
function getPlayWsUrl(): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/ws?room=play`;
}
```

**File**: `frontend/src/routes/Display.tsx`

Change `getDisplayWsUrl`:
```typescript
function getDisplayWsUrl(): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/ws?room=display`;
}
```

This directly resolves the Connection-header error. The Go handler will route correctly to `playHandler` and `displayHandler`.

---

### Fix A2 — Centralise WebSocket URL Construction [PERMANENT]

Create `frontend/src/ws/urls.ts` as the single source of truth for all WebSocket endpoint URLs:

```typescript
export function hostWsUrl(token: string): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/ws?token=${token}`;
}

export function playWsUrl(): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/ws?room=play`;
}

export function displayWsUrl(): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/ws?room=display`;
}
```

Import from this module in `Host.tsx`, `Play.tsx`, `Display.tsx`. Any future change to URL shape is made once and propagates everywhere.

---

### Fix B1 — Add SPA Fallback to Static Handler [IMMEDIATE + PERMANENT]

**File**: `internal/static/embed.go`

Replace the handler body with a fallback that serves `index.html` for any path not found in the embedded filesystem:

```go
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Try to serve the exact file. If it does not exist, serve index.html
    // so that client-side routes (BrowserRouter) are handled by the SPA.
    if r.URL.Path != "/" {
        _, err := distFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
        if err == nil {
            fileServer.ServeHTTP(w, r)
            return
        }
    }
    http.ServeFileFS(w, r, distFS, "index.html")
})
```

Add `"strings"` to the import block. This ensures `/play` and `/display` serve `index.html`, allowing the React Router to render the correct component.

---

### Fix B2 — Add Tests for SPA Sub-Route Serving [PERMANENT]

**File**: `internal/static/embed_test.go`

Add test cases asserting that `/play` and `/display` each return HTTP 200 with HTML content:

```go
func TestStaticHandler_ServesPlayRouteWithHTML(t *testing.T) {
    handler := static.NewStaticHandler()
    req := httptest.NewRequest(http.MethodGet, "/play", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Fatalf("expected HTTP 200 for /play, got %d", rec.Code)
    }
    if !strings.Contains(rec.Body.String(), "<html") {
        t.Errorf("expected HTML for /play")
    }
}

func TestStaticHandler_ServesDisplayRouteWithHTML(t *testing.T) {
    handler := static.NewStaticHandler()
    req := httptest.NewRequest(http.MethodGet, "/display", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Fatalf("expected HTTP 200 for /display, got %d", rec.Code)
    }
    if !strings.Contains(rec.Body.String(), "<html") {
        t.Errorf("expected HTML for /display")
    }
}
```

These tests would have caught Root Cause B at development time.

---

### Fix C — Rebuild and Redeploy Binary [IMMEDIATE]

**Evidence**: E13, E14 — the `server` and `trivia` binaries in the repo root have identical hashes and predate the new routing code in `cmd/server/main.go`. The current working-tree code compiles successfully but has not been built into a deployable binary.

After applying fixes A1 and B1:

```sh
# 1. Rebuild frontend (incorporates ?room= fix)
cd frontend && npm run build && cd ..

# 2. Rebuild Go binary (embeds updated frontend, includes new routing)
go build -o trivia ./cmd/server

# 3. Run with required env vars
HOST_TOKEN=<secret> QUIZ_DIR=<path> ./trivia
```

The repo-root binaries should be removed or gitignored; they represent a stale build artefact that will silently hide regression if reused.

---

## Prioritised Action Order

| Priority | Fix | Effort | Impact |
|----------|-----|--------|--------|
| 1 | Fix A1: Add `?room=` params to Play.tsx and Display.tsx | Minutes | Eliminates the WebSocket upgrade error immediately |
| 2 | Fix B1: Add SPA fallback to static handler | 10 min | Enables cold navigation to `/play` and `/display` |
| 3 | Fix C: Rebuild binary | Minutes | Ensures running binary includes both fixes |
| 4 | Fix B2: Add static handler sub-route tests | 15 min | Prevents regression of Branch B |
| 5 | Fix A2: Centralise WebSocket URL module | 30 min | Prevents future URL divergence (structural prevention) |

---

## Prevention Strategy

**For Root Cause A** (implicit URL contract): Introduce a `frontend/src/ws/urls.ts` module (Fix A2) so that all WebSocket URL construction has a single location. Pair this with an acceptance test that asserts each room (`play`, `display`, `host`) connects successfully — the test driver already models this correctly and could be used as the specification template.

**For Root Cause B** (no SPA fallback): The static handler must be extended (Fix B1) and the new behaviour pinned with tests (Fix B2) before the fix is complete. The existing test in `embed_test.go` should be treated as incomplete until it covers all SPA-owned routes.

**Structural observation**: The test driver (`driver.go`) correctly specifies the WebSocket protocol (room param, token param) but this specification is not referenced by or enforced against the frontend. A Vite dev-proxy configuration that mirrors the production routing (including the room-dispatch logic) would surface the missing `?room=` params during development, before integration.
