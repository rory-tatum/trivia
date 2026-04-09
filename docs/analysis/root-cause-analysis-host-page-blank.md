# Root Cause Analysis: Host Page Shows Nothing When Token Provided via Query String

**Date:** 2026-04-09
**Analyst:** Rex (RCA Specialist — Toyota 5 Whys)
**Scope:** Frontend host route and backend WebSocket auth path for `?token=<value>`

---

## Problem Statement

When a user navigates to the application root (`/`) with `?token=1234` appended, the host page renders
a near-empty UI — only a static heading ("Quizmaster Panel") and a connection status line. No
quiz-management controls, no game state, and no meaningful content appear. The page is not blank in
the HTML sense (the shell renders), but it is functionally blank: it carries no interactive content
and does not reflect what the host is supposed to be able to do.

---

## Scope Definition

**In scope:**
- Frontend host route (`frontend/src/routes/Host.tsx`)
- React router configuration (`frontend/src/main.tsx`)
- Backend WebSocket routing and auth guard (`cmd/server/main.go`, `internal/handler/auth.go`)
- Static file serving (`internal/static/embed.go`)

**Out of scope:**
- Play and Display routes (structurally identical issue but separate feature)
- Quiz-loading business logic (not reachable if UI is empty)
- Network-layer connectivity failures (separate investigation)

---

## Evidence Gathered

| # | File | Lines | Observation |
|---|------|-------|-------------|
| E1 | `frontend/src/main.tsx` | 11-14 | React Router defines three routes: `/` → `<Host>`, `/play` → `<Play>`, `/display` → `<Display>`. There is no `/host` route. |
| E2 | `frontend/src/routes/Host.tsx` | 6-9 | `getHostWsUrl()` reads `?token` from `window.location.search` and appends it to `/ws?token=<value>`. Token is correctly forwarded to the WebSocket URL. |
| E3 | `frontend/src/routes/Host.tsx` | 37-43 | `return` block renders only: an `<h1>` ("Quizmaster Panel"), a static status paragraph, and conditionally a "Last event" paragraph. There are no controls, no quiz-loading form, no game-state display — the JSX is a skeleton. |
| E4 | `frontend/src/ws/client.ts` | 28 | `setConnected(true)` is called immediately after `client.connect()`, before the WebSocket `onopen` fires. Status shows "connected" even if the WS handshake fails. |
| E5 | `cmd/server/main.go` | 59-70 | Backend `/ws` route: if `?token` is non-empty, `authGuard(hostHandler)` runs. The token is validated against `HOST_TOKEN` env var (via `config.Load()`). |
| E6 | `internal/handler/auth.go` | 16-19 | Auth guard performs exact string equality: `token != hostToken`. If `?token=1234` does not exactly match the server's `HOST_TOKEN`, the WebSocket upgrade returns HTTP 403 Forbidden. |
| E7 | `config/config.go` | 18-20 | `HOST_TOKEN` is read from the environment. The server refuses to start if it is empty, but its value is never communicated to the user/frontend. |
| E8 | `frontend/src/ws/client.ts` | 78-80 | `ws.onerror` closes the socket; `ws.onclose` triggers `scheduleReconnect()`. A 403 during WebSocket upgrade will land here, causing silent reconnect loops — no error is surfaced to the UI. |
| E9 | `frontend/src/routes/Host.tsx` | 21-24 | `client.onMessage` only sets `lastEvent` to `msg.event`. Even if the WS connects successfully and the server sends events, the UI renders only the event name string — no structured content. |
| E10 | `internal/handler/host.go` | 158-164 | Server sends `quiz_loaded` event after `host_load_quiz` command. But `host_load_quiz` requires the host to first *send* a message — the UI has no control to trigger this. |

---

## Toyota 5 Whys — Multi-Causal Analysis

### Branch A: Host UI is a Skeleton (Primary Cause of "Nothing Showing")

**WHY 1 — What is the observable symptom?**
The host page renders only a heading, status line, and optional last-event string. No quiz-management
controls, game-state display, or interactive elements appear.
Evidence: E3 — Host.tsx JSX contains exactly three elements; none are interactive UI.

**WHY 2 — Why does the UI contain no controls?**
The `Host` component was never implemented beyond a structural skeleton. It mirrors the same
boilerplate as `Play.tsx` and `Display.tsx` (which are also skeletons), suggesting all three were
scaffolded at the same time as connection stubs, with UI implementation deferred.
Evidence: E3, E2 — Host.tsx and Play.tsx are near-identical 44-line files. No conditional rendering
based on game state. No form elements. No event-driven UI updates beyond `lastEvent` string.

**WHY 3 — Why was the host UI left as a scaffold?**
The component handles the WebSocket connection layer (reads token, builds URL, connects) but has no
state model for game phases, quiz metadata, or host actions. The message handler (`onMessage`)
discards all payload data, keeping only the event name string. There is no state to drive a UI from.
Evidence: E9 — `setLastEvent(msg.event)` discards `msg.payload` entirely.

**WHY 4 — Why is there no state model for host game phases?**
The frontend has the full message type definitions (`messages.ts`) — `quiz_loaded`, `round_started`,
`question_revealed`, etc. — but `Host.tsx` does not import or consume them. The type system exists
but the component was not wired to it.
Evidence: E9, E3 — messages.ts (E7 in messages file) defines `SessionCreatedMsg`, `RoundStartedMsg`,
etc. Host.tsx useState is `string | null`, not a typed game-state union.

**WHY 5 — Root Cause A**
The Host component was built as a connection-layer proof-of-concept (verifying the WS auth path works)
but UI implementation — state machine, event-to-state mapping, render branches per game phase — was
never written. The skeleton was committed as if complete.

---

### Branch B: Connection Status is Misleading (Masks Real Failures)

**WHY 1 — What is the observable symptom?**
The page shows "Status: connected" even when the WebSocket connection fails (e.g., auth rejected).
Evidence: E4 — `setConnected(true)` executes synchronously after `client.connect()`, before
`ws.onopen` fires.

**WHY 2 — Why does the status show "connected" before the handshake completes?**
`client.connect()` is synchronous — it calls `openSocket()` which creates the WebSocket object, but
the connection is asynchronous. `setConnected(true)` is called on the next line, not inside
`ws.onopen`.
Evidence: E4 — `client.connect(); setConnected(true);` in Host.tsx lines 28-29. WsClient.openSocket()
sets `ws.onopen` as a callback that runs later.

**WHY 3 — Why is there no `onopen` callback exposed by WsClient?**
`WsClient` exposes `onMessage` and `on(event, handler)` for named events (e.g., `reconnect_failed`),
but no `onOpen` or `onConnected` hook. The internal `ws.onopen` resets backoff counters but emits no
external signal.
Evidence: E6 (client.ts lines 57-60) — `ws.onopen` sets `this.attempt = 0` and `this.backoffMs`,
never calls `this.emit(...)`.

**WHY 4 — Why was no onOpen hook provided?**
The `reconnect_failed` event was implemented as the only named lifecycle event — the failure path was
considered (E8: `ws.onerror` → `ws.close()` → `scheduleReconnect()` → emit `reconnect_failed`), but
the success path was not. Auth-failure (HTTP 403 during upgrade) causes `onerror` + `onclose` without
reaching `onopen`, and this failure is silent beyond initiating a reconnect loop.
Evidence: E8 — `ws.onerror` closes socket; `ws.onclose` calls `scheduleReconnect`. No error state
propagated to the component.

**WHY 5 — Root Cause B**
WsClient's lifecycle model is incomplete: it handles the terminal failure case (`reconnect_failed`)
but has no mechanism to signal successful connection establishment or to distinguish a transport error
(network) from an auth rejection (HTTP 403). The `connected` state in `Host.tsx` is therefore
untrustworthy, and auth failures are silent.

---

### Branch C: Token Mismatch Silently Fails (Auth Path)

**WHY 1 — What is the observable symptom?**
If the user provides `?token=1234` but the server's `HOST_TOKEN` env var is set to a different value,
the WebSocket connection silently fails. The page still shows "connected."
Evidence: E6 — `token != hostToken` returns 403 Forbidden. E4+E8 — 403 causes `onerror/onclose`,
which triggers silent reconnect loop. UI never shows an error.

**WHY 2 — Why does a token mismatch not produce a visible error?**
The backend returns HTTP 403, which the WebSocket client receives as a connection error. `ws.onerror`
fires, closes the socket, and `scheduleReconnect` queues another attempt. The error is not
distinguished from a transient network failure.
Evidence: E8 — identical handling for all error types in `ws.onerror`.

**WHY 3 — Why is there no distinction between auth failure and network failure?**
The WebSocket API does not expose the HTTP status code after a failed upgrade in the `onerror` event.
However, the `onclose` event does receive a `CloseEvent` with a `code` and `reason`. The WsClient
`ws.onclose` handler ignores these fields entirely.
Evidence: client.ts line 73 — `ws.onclose = () => { if (this.closed) return; this.scheduleReconnect(); }`.
`CloseEvent.code` and `CloseEvent.reason` are not read.

**WHY 4 — Why are CloseEvent codes not inspected?**
The `scheduleReconnect` design assumes all closes are transient and recoverable. No policy was
defined for permanent failures (auth rejection, bad URL). There is no error state fed back to the
component other than the terminal `reconnect_failed` event after 10 attempts.
Evidence: client.ts lines 83-97 — only `attempt >= MAX_RECONNECT_ATTEMPTS` triggers `emit(RECONNECT_FAILED)`.

**WHY 5 — Root Cause C**
The reconnect strategy has no concept of non-recoverable failures. Auth rejections (which will never
succeed on retry with the same token) are retried up to 10 times before `reconnect_failed` is
emitted, and even then the component only sets `connected = false` — it never shows the user *why*
the connection failed (wrong token, server down, etc.).

---

### Branch D: No `/host` Named Route (Navigation Issue)

**WHY 1 — What is the observable symptom?**
A user navigating to `/host?token=1234` would not see the Host component — they would see a blank
page (no route match).
Evidence: E1 — `main.tsx` defines routes: `/` → Host, `/play` → Play, `/display` → Display. There
is no `/host` route.

**WHY 2 — Why is the host accessible only at `/`?**
The root path `/` is assigned to `<Host>`. This is a deliberate design choice: the host URL is the
app root. The token in the query string is the access control mechanism, not the path.
Evidence: E1 — `<Route path="/" element={<Host />} />`.

**WHY 3 — Why is this a problem?**
It is only a problem if the host is expected to navigate to `/host?token=...` (a natural assumption
from the problem description "the host page"). Navigating to `/host?token=1234` would hit the static
file server fallback, which serves `index.html` (E13 — embed.go lines 28-32: fallback to index.html).
React Router receives `/host` as the path, matches no route, and renders nothing.
Evidence: E1 (no `/host` route), embed.go lines 28-32 (SPA fallback serves index.html for unknown
paths), React Router renders null for unmatched paths with no catch-all `*` route defined.

**WHY 4 — Why is there no catch-all or 404 route?**
`main.tsx` defines only the three known routes. No `<Route path="*">` exists. React Router v6 with
no wildcard route silently renders nothing for unmatched paths.
Evidence: E1 — three explicit routes, no wildcard.

**WHY 5 — Root Cause D**
The host route is at `/` not `/host`, but this is undocumented and counter-intuitive. Combined with
no 404/wildcard route, navigating to `/host` instead of `/?token=...` yields a completely blank page
with no feedback to the user. The host entry URL is not communicated anywhere in the frontend.

---

## Backwards Chain Validation

**Root Cause A** → Host component has no state model or interactive UI → `Host.tsx` JSX renders
only skeleton → user sees heading + status + last event string only → "host page shows nothing."
Chain validates.

**Root Cause B** → `setConnected(true)` called before `ws.onopen` + no `onOpen` hook → `connected`
state is set to `true` before connection is confirmed → UI always shows "connected" regardless of
actual WS state → misleading status masks the real problem.
Chain validates.

**Root Cause C** → `HOST_TOKEN` env value != user-supplied token → server returns 403 → `ws.onerror`
fires → `scheduleReconnect()` with no error propagation → component receives no error signal →
page shows "connected" and nothing else → "host page shows nothing."
Chain validates.

**Root Cause D** → user navigates to `/host?token=1234` → static server falls back to `index.html`
→ React Router receives path `/host` → no matching route → renders nothing → completely blank page.
Chain validates (distinct blank scenario from A/B/C).

---

## Root Causes Summary

| ID | Root Cause | Nature |
|----|-----------|--------|
| A | `Host.tsx` UI is an unimplemented scaffold — no state model, no controls, no event-to-render mapping | Missing implementation |
| B | `WsClient` has no `onOpen` lifecycle hook; `connected` state is set optimistically before the handshake | Design gap in WsClient lifecycle model |
| C | Auth failures (HTTP 403) are indistinguishable from transient errors; silently retried; no error surfaced to UI | Missing error classification in reconnect strategy |
| D | Host route is at `/` not `/host`; no wildcard/404 route means `/host?token=...` renders a blank page | Undocumented routing + missing catch-all route |

---

## Recommended Fixes

### Immediate Mitigations (restore observable function, do not fix the root)

**M1 (Root Cause D):** Add a `/host` route alias that redirects to `/?token=<token>`, or add a
wildcard `<Route path="*">` that renders a "not found" message. This stops the completely blank page
at wrong URLs. Effort: low.

**M2 (Root Cause C):** In `WsClient.onclose`, inspect `CloseEvent.code`. Codes 1006 (abnormal
closure, typical of failed HTTP upgrade) vs 1000/1001 (normal close) allow distinguishing auth
failure from a clean disconnect. Emit a distinct `auth_failed` event that `Host.tsx` can display.
Effort: low.

### Permanent Fixes (address root causes, prevent recurrence)

**F1 (Root Cause A — primary fix):** Implement the Host component UI:
- State: `gamePhase` (idle | quiz_loaded | round_active | scoring | ceremony | game_over), `quizMeta`,
  `revealedQuestions`, `scores`.
- Event handlers: map each `OutgoingMessage` event type to state transitions using the typed union in
  `messages.ts`.
- Controls: file-path input + "Load Quiz" button (sends `host_load_quiz`); round controls
  (`host_start_round`, `host_reveal_question`); scoring controls (`host_mark_answer`,
  `host_publish_scores`); game-end control (`host_end_game`).
- Render branches: one per `gamePhase` value.
This is the core missing work. All other root causes contribute to a bad experience; this one is why
the page is functionally empty.

**F2 (Root Cause B):** Add `onOpen(handler: () => void): void` to `WsClient`. Inside `ws.onopen`,
call all registered open handlers after resetting backoff. In `Host.tsx`, set `connected = true`
only inside this callback, not synchronously after `connect()`.

**F3 (Root Cause C):** Extend `WsClient.scheduleReconnect` to accept a close code. For codes
indicating permanent failure (4001 for application-level auth, or detected 403 pattern via 1006 on
first attempt), emit a typed `connection_refused` event rather than retrying. `Host.tsx` handles
this by setting an `authError` state and rendering a "Invalid token — check HOST_TOKEN" message.

**F4 (Root Cause D):** Either:
(a) Add `<Route path="/host" element={<Host />} />` alongside `/` in `main.tsx` (supports both
entry points), or
(b) Document that the host URL is `http://<host>/?token=<HOST_TOKEN>` and add a fallback `<Route
path="*" element={<NotFound />}>` so unmatched routes show a message rather than silence.

---

## Fix Priority

| Priority | Fix | Root Cause | Effort |
|----------|-----|-----------|--------|
| 1 | F1 — Implement Host UI | A | High (core feature work) |
| 2 | M2 / F3 — Surface auth errors | C | Low/Medium |
| 3 | F2 — Fix `connected` state timing | B | Low |
| 4 | F4 — Add `/host` route or catch-all | D | Low |
