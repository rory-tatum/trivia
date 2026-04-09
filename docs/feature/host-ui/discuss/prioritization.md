# Prioritization — Host UI (Quizmaster Panel)

**Feature ID**: host-ui
**Date**: 2026-04-09

---

## Delivery Sequence

Stories are sequenced by user outcome dependency: each story must be usable before the next one becomes meaningful.

### Walking Skeleton (deliver first — proves the whole pipe works)

| Priority | Story | Why first |
|---|---|---|
| 1 | US-01: Reliable connection status + auth errors | Every other story depends on knowing the connection is real |
| 2 | US-02: Load quiz | No game can start without a loaded quiz |
| 3 | US-03: Run a round (start + reveal + end) | Core game loop — players need questions to answer |
| 4 | US-04: Score a round (mark + publish) | Scoring is the point of the game |
| 5 | US-06: End game + final leaderboard | Gives the session a definitive end state |

**Walking skeleton outcome**: Marcus can run a complete one-round trivia game end-to-end. The full pipe — connect → load → round → score → end — is verified and usable.

---

### Release 1 (adds ceremony — the answer walkthrough experience)

| Priority | Story | Why here |
|---|---|---|
| 6 | US-05: Run answer ceremony on display | Enhances experience; depends on scoring being done first; display route must be functional |

**Release 1 outcome**: Marcus can walk the room through answers after each round, making the experience feel complete.

---

## Splitting Rationale

The walking skeleton and Release 1 are separated because:
- The ceremony (US-05) requires the display room (`/display` route) to be functional — a separate concern
- A game is fully playable without ceremony; scores can be verbally announced if needed
- Ceremony involves a different broadcast room (display-only for answers) — lower integration risk to defer

---

## Out of Scope (host-ui feature)

These items emerged from discovery but belong to separate features:

| Item | Reason |
|---|---|
| Play route UI (player answer submission) | Separate feature; different persona (player, not quizmaster) |
| Display route UI (audience view) | Separate feature; display persona, different route |
| Quiz authoring (creating YAML files) | Separate feature; quizmaster as author, not runner |
| Multi-game session management | Future; current architecture is single-session |

---

## Risk Register

| Risk | Mitigation |
|---|---|
| WsClient onOpen hook missing (IC-1) | US-01 explicitly requires this fix as a technical note; blocks all other stories |
| Auth failure silent retry loop (IC-2) | US-01 requires CloseEvent.code inspection; blocks error UX |
| Scoring panel needs team answers from server | IC-4 — verify score_updated includes team_name; if not, require a state_snapshot on scoring_opened |
| Ceremony display boundary (answers to display-only, not play) | IC verified in handler.go line 296 — answers sent to RoomDisplay only |
