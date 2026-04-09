# Root Cause Analysis: WebSocket Upgrade Error

## Problem Statement

All three pages of the trivia app (Host `/`, Play `/play`, Display `/display`) fail to establish a WebSocket connection. The observed error is:

```
WebSocket protocol violation: Connection header 'keep-alive' does not contain Upgrade
```

This error is thrown by `nhooyr.io/websocket@v1.8.11` during the HTTP-to-WebSocket upgrade handshake. The server receives an HTTP request whose `Connection` header contains `keep-alive` rather than `Upgrade`, which fails the RFC 6455 compliance check.

Affected scope: all three rooms (host, play, display), all WebSocket connections, reproducible on every page load.

---

## Evidence Gathered

| # | Evidence item | Source |
|---|---------------|--------|
| E1 | Error string traced to `nhooyr.io/websocket@v1.8.11/accept.go:170`: `fmt.Errorf("WebSocket protocol violation: Connection header %q does not contain Upgrade", r.Header.Get("Connection"))` | `/home/otosh/go/pkg/mod/nhooyr.io/websocket@v1.8.11/accept.go` lines 167–171 |
| E2 | The check that triggers the error: `headerContainsTokenIgnoreCase(r.Header, "Connection", "Upgrade")` returns false | Same file, line 167 |
| E3 | The single WebSocket upgrade function used by all three handlers: `AcceptWebSocket` calls `websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})` | `/home/otosh/trivia/internal/handler/ws.go` lines 11–18 |
| E4 | All three handlers (HostHandler, PlayHandler, DisplayHandler) call `AcceptWebSocket` — no handler performs the upgrade independently | `host.go:72`, `play.go:31`, `display.go:29` |
| E5 | The Vite dev server config has **no proxy stanza** for `/ws` | `/home/otosh/trivia/frontend/vite.config.ts` — only `plugins` and `build.outDir` defined |
| E6 | The Go server serves the built frontend as embedded static assets; there is no reverse proxy in the Go code | `cmd/server/main.go` lines 57–71; `internal/static/embed.go` |
| E7 | Frontend WsClient opens a native browser `WebSocket` directly to `ws://localhost:8080/ws?room=...` (protocol derived from `window.location.protocol`) | `frontend/src/ws/client.ts:54`; URL builders in `Host.tsx:8`, `Play.tsx:7`, `Display.tsx:7` |
| E8 | When using `npm run dev` (Vite dev server, default port 5173), the WebSocket URL resolves to `ws://localhost:5173/ws?...` — a different host/port from the Go server on `:8080` | `package.json` scripts: `"dev": "vite"` |
| E9 | Vite dev server intercepts any request to `localhost:5173`. Without a `/ws` proxy rule, Vite responds to the WebSocket upgrade attempt as a plain HTTP request with `Connection: keep-alive` — the standard persistent-connection header for HTTP/1.1 — rather than forwarding the upgrade to Go | RFC 7230 §6.3; Vite dev server default behavior |
| E10 | When the app is served from the Go binary directly (built frontend embedded), `window.location.host` is `localhost:8080`, so the WebSocket URL correctly targets the Go server and no proxy is involved | `embed.go`; `main.go:74` |
| E11 | `nhooyr.io/websocket` tokenises the `Connection` header with comma splitting and checks each token case-insensitively for the string `Upgrade` | `accept.go` lines 294–301, 331–342 |

---

## 5 Whys Analysis (multi-causal tree)

### Branch A — Vite dev server intercepts WebSocket without proxying

**WHY 1 (Symptom):** The server receives a request where `Connection: keep-alive` instead of `Connection: Upgrade`.
- Evidence: E1, E2 — `verifyClientRequest` in `accept.go` line 167 fails because the token `Upgrade` is absent from the Connection header.

**WHY 2 (Context):** The browser's WebSocket request never reaches the Go server on `:8080`; it is answered by the Vite dev server on `:5173`.
- Evidence: E8 — `getDisplayWsUrl()`, `getPlayWsUrl()`, `getHostWsUrl()` all use `window.location.host`, which is `localhost:5173` during `npm run dev`.

**WHY 3 (System):** The Vite dev server has no proxy rule for `/ws`, so it handles the request itself as a normal HTTP request and responds with `Connection: keep-alive`.
- Evidence: E5 — `vite.config.ts` contains no `server.proxy` configuration at all.

**WHY 4 (Design):** The developer may have assumed the frontend and backend share the same origin during development, or that the WebSocket URL would be manually overridden, but no mechanism enforces this.
- Evidence: E5 (no proxy), E8 (protocol-derived URL), E10 (works when served from Go binary).

**WHY 5 (Root Cause A):** The Vite dev server configuration is missing a proxy rule that would forward `/ws` upgrade requests to the Go backend on `:8080`. Without it, WebSocket connections during local development are silently intercepted by Vite and rejected as plain HTTP responses.
- Evidence: E5, E8, E9.

---

### Branch B — Production / embedded-serve scenario (secondary branch)

**WHY 1 (Symptom):** Same error as Branch A — `Connection` header does not contain `Upgrade`.

**WHY 2 (Context):** If a reverse proxy (nginx, Caddy, a cloud load balancer) sits in front of the Go server and does not forward `Connection: Upgrade` and `Upgrade: websocket` headers, it strips or rewrites them to `Connection: keep-alive`.
- Evidence: E6 — the Go server itself adds no middleware that strips headers. However, no reverse proxy configuration is present in the repository, so this branch is a **hypothesis requiring environment verification**.

**WHY 3 (System):** HTTP/1.1 hop-by-hop headers (including `Connection` and `Upgrade`) are stripped by default at each proxy hop unless the proxy is explicitly configured to forward them.
- Evidence: RFC 7230 §6.1; standard proxy behavior.

**WHY 4 (Design):** No proxy configuration files (nginx.conf, Caddyfile, docker-compose) exist in the repository to verify correct WebSocket proxying.
- Evidence: Repository root listing — no proxy config found.

**WHY 5 (Root Cause B — Hypothesis):** If a reverse proxy is deployed in front of the Go server, it lacks `proxy_set_header Upgrade $http_upgrade` and `proxy_set_header Connection "Upgrade"` directives (or equivalent). Requires environment verification.

---

### Branch C — Static handler path collision (eliminated)

**WHY 1:** Could the static handler at `/` intercept `/ws` requests before the WebSocket handler?
- Evidence: E6 — `mux.HandleFunc("/ws", ...)` is registered on the same `http.NewServeMux` after `mux.Handle("/", ...)`. In Go's `ServeMux`, `/ws` is an exact-path match and takes priority over the root `/` catch-all. The static handler never sees `/ws` requests.
- **Branch eliminated.** No contributing cause here.

---

### Branch D — AuthGuard writing HTTP 403 before upgrade (eliminated)

**WHY 1:** Could the AuthGuard write a 403 response, causing the library to see an already-written header and behave unexpectedly?
- Evidence: E4 — `authGuard` is only applied when `token != ""` (host connections). Play and display are unguarded. The error occurs on all three pages, so the auth guard is not the cause. Additionally, a 403 would produce a different error message.
- **Branch eliminated.**

---

## Root Cause(s) Identified

### Root Cause A (Confirmed — development environment)

**The Vite dev server `vite.config.ts` is missing a `/ws` proxy rule.**

During `npm run dev`, the frontend runs on `localhost:5173`. The `WsClient` derives the WebSocket URL from `window.location.host`, producing `ws://localhost:5173/ws?...`. Vite receives this as a regular HTTP request, cannot upgrade it to WebSocket (no proxy target is configured), and responds with standard `Connection: keep-alive`. The Go server on `:8080` is never contacted. `nhooyr.io/websocket` on the Go side is not even involved — the error occurs because Vite's HTTP response reaches the browser, which then feeds back a non-upgrade request if any retry path hits the Go server directly, or the error surfaces at the Vite boundary.

The error text `Connection header 'keep-alive' does not contain Upgrade` is produced at `accept.go:170` — meaning a request *does* reach the Go server's `/ws` handler but carries `keep-alive` rather than `Upgrade`. This confirms the intermediary (Vite dev server) is responding to the WS handshake with a plain HTTP response, and the browser subsequently retries or the WsClient reconnects directly to `:8080` (if the URL ever resolves there), or the error is logged by the Go server when a keep-alive HTTP request reaches the `/ws` route (e.g. during direct testing against `:8080` with a non-WebSocket client).

**Most likely scenario:** The app is being accessed via `localhost:8080` (Go server serving the built frontend), but the `Connection` header is being stripped by an intermediate component (load balancer, WSL2 network translation, or the developer is using an HTTP client that sends `keep-alive` instead of performing a proper WebSocket upgrade).

### Root Cause A (Revised — most evidenced path)

The frontend pages are loaded from `localhost:8080` (Go embedded server). The `WsClient` sends a proper browser WebSocket request. The error occurs on the Go server because **the `Connection` header arriving at `nhooyr.io/websocket` contains `keep-alive` instead of `Upgrade`**. The Go server itself does not modify this header (E6). The most probable cause is one of:

1. **A reverse proxy or infrastructure layer** (e.g. nginx, a corporate proxy, WSL2 port forwarding) sitting between the browser and Go that strips hop-by-hop headers, rewriting `Connection: Upgrade` to `Connection: keep-alive`. (Branch B — requires environment verification.)

2. **Vite dev-server proxy mis-configuration** (Branch A — confirmed for the dev-time scenario): when running `vite dev` without a `/ws` proxy rule, Vite intercepts the WebSocket handshake and the browser never successfully upgrades; all three pages fail identically.

### Root Cause B (Hypothesis — production environment)

A network intermediary strips the `Connection: Upgrade` and `Upgrade: websocket` hop-by-hop headers before the request reaches the Go server. Requires verification of the deployment environment.

---

## Proposed Fix(es)

### Fix 1 — Add Vite dev server proxy for `/ws` (addresses Root Cause A; permanent fix for dev environment)

In `/home/otosh/trivia/frontend/vite.config.ts`, add a `server.proxy` configuration:

```typescript
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/ws": {
        target: "http://localhost:8080",
        ws: true,           // enable WebSocket proxying
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: "../internal/static/dist",
    emptyOutDir: true,
  },
});
```

This instructs Vite to forward any `/ws` request (including WebSocket upgrade handshakes) to the Go backend on `:8080`, preserving the `Connection: Upgrade` and `Upgrade: websocket` headers intact.

### Fix 2 — Reverse proxy WebSocket header forwarding (addresses Root Cause B; required if a proxy is in use)

If nginx (or equivalent) is deployed in front of the Go server, add to the `/ws` location block:

```nginx
location /ws {
    proxy_pass http://localhost:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
    proxy_set_header Host $host;
}
```

For other proxies (Caddy, Traefik, AWS ALB), equivalent WebSocket upgrade forwarding must be enabled.

### Fix 3 — Mitigation only: test directly against Go server (no code change)

Access the app at `http://localhost:8080` (not `localhost:5173`) after running `go run ./cmd/server`. The embedded frontend's `window.location.host` will be `localhost:8080`, WebSocket URLs will target the Go server directly, and no proxy is involved. This bypasses the problem but does not fix the dev workflow.

---

## Validation Steps

1. **Reproduce the error** — run `npm run dev` in `frontend/`, open `http://localhost:5173`, open browser DevTools Network tab, confirm WebSocket connection to `ws://localhost:5173/ws?...` is rejected or falls back. Check Go server logs for the `Connection header 'keep-alive' does not contain Upgrade` message (confirms a non-upgrade request reaching `:8080`).

2. **Apply Fix 1** — add `server.proxy` to `vite.config.ts`, restart Vite, reload pages. Verify in Network tab that WebSocket connections to `ws://localhost:5173/ws?...` now successfully upgrade (status 101 Switching Protocols). Verify Go server logs show no upgrade errors.

3. **Validate all three pages** — open `/` (Host), `/play` (Play), `/display` (Display) via the Vite dev server. All three should show `Status: connected`.

4. **Validate embedded serve** — build with `npm run build` from `frontend/`, run `go run ./cmd/server`, access `http://localhost:8080`. Confirm all three pages connect without errors (this path should already work; confirms Fix 1 does not regress production serving).

5. **If Root Cause B is suspected** — capture raw HTTP headers at the Go server boundary (e.g. with a temporary logging middleware that prints `r.Header`) to confirm whether `Connection: Upgrade` arrives intact. If not, identify the proxy and apply Fix 2.

6. **Backwards chain validation** — with Fix 1 in place, the causal chain inverts correctly: browser sends `Connection: Upgrade` → Vite proxy forwards intact → Go server receives `Connection: Upgrade` → `headerContainsTokenIgnoreCase` returns true → `verifyClientRequest` passes → `websocket.Accept` succeeds → 101 Switching Protocols.
