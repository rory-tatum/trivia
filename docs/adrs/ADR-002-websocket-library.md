# ADR-002: WebSocket Library -- nhooyr.io/websocket

## Status

Accepted

## Date

2026-03-29

## Context

The trivia game requires server-side WebSocket support for real-time synchronization between the quizmaster and 2-10 player/display clients (DEC-008). The backend is Go 1.23.

The two primary WebSocket libraries for Go are:

1. `github.com/gorilla/websocket` -- the historically dominant choice
2. `nhooyr.io/websocket` -- a newer library with an idiomatic Go API

This resolves open question OQ-01 from the DISCUSS wave.

## Decision

Use `nhooyr.io/websocket` v1.x.

## Rationale

### gorilla/websocket: archived

The Gorilla web toolkit organization was archived in December 2022. The `gorilla/websocket` repository has received no new releases since then. While the code remains functional, choosing an archived library introduces a maintenance liability: known CVEs will not receive official patches, and the API will not evolve with Go versions.

At the time of writing, there is a community fork under the `gorilla` org umbrella, but it has not reached stable release.

### nhooyr.io/websocket: actively maintained

`nhooyr.io/websocket` is actively maintained. It was designed with Go idioms from the start:

- Context-aware: all read/write operations accept `context.Context` for cancellation and timeout control
- Does not require a separate goroutine per write (uses a mutex internally)
- `io.Reader`/`io.Writer` interface compatible
- Supports HTTP/2 (via WebSocket over HTTP/2) in addition to HTTP/1.1
- Tested against Go's race detector

## Alternatives Considered

### Alternative A: gorilla/websocket

- **Pro:** Most tutorials and examples use it. Large install base means more StackOverflow answers.
- **Con:** Archived. No official patch path for future CVEs. API design predates context propagation in Go stdlib.
- **Rejected:** Archived status is a maintenance liability for a project intended to be maintained and extended.

### Alternative B: gobwas/ws

- **Pro:** Very low-level, high-performance WebSocket library used in production at scale.
- **Con:** Low-level API requires more boilerplate. Designed for high-throughput servers; complexity is not justified for 2-10 connections.
- **Rejected:** Unnecessary complexity for this use case.

### Alternative C: net/http + manual WebSocket upgrade (RFC 6455 from scratch)

- **Pro:** Zero dependency.
- **Con:** WebSocket framing, masking, ping/pong, and close handshake are non-trivial to implement correctly. Reinventing this has no benefit.
- **Rejected:** Use a well-tested library.

## Consequences

### Positive

- Context-aware API integrates cleanly with Go's request lifecycle
- Active maintenance reduces long-term security risk
- Idiomatic Go reduces learning curve for the developer

### Negative

- Less tutorial material compared to gorilla/websocket; developer may need to read library source
- nhooyr.io/websocket has fewer GitHub stars than gorilla/websocket (community familiarity lower)

## Compliance

- License: MIT (permissive, free)
