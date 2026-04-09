# Requirements — Host UI (Quizmaster Panel)

**Feature ID**: host-ui
**Date**: 2026-04-09
**Persona**: Marcus, quizmaster (runs pub trivia for friends from a browser)

---

## Problem Statement

Marcus navigates to `/?token=HOST_SECRET` to run his trivia night. The page currently shows only a heading and a status line. There are no controls, no game state, no way to load a quiz or advance the game. The connection status is misleading (shows "Connected" before the WebSocket handshake completes and even after auth failures). Marcus cannot run a game.

The root causes are:
- **A**: Host.tsx UI is an unimplemented scaffold — no state model, no controls, no event-to-render mapping
- **B**: WsClient has no `onOpen` hook; `connected` set optimistically before handshake
- **C**: Auth failures silently retry; no error shown to user
- **D**: Host route is at `/` not `/host`; no catch-all route for wrong URLs

---

## Functional Requirements

### FR-1: Connection status accuracy
The `connected` state must reflect actual WebSocket handshake success. Status must only become "Connected" after `ws.onopen` fires. Before that, status must show "Connecting...".

### FR-2: Auth failure detection and display
When the WebSocket upgrade is rejected (HTTP 403), the page must detect the failure (via CloseEvent.code 1006 on first attempt without prior onOpen success) and display a clear, permanent error. The client must not enter an infinite retry loop on auth failure.

### FR-3: Load quiz form
When connected, Marcus must see a form with a file path text input and a "Load Quiz" button. Submitting sends `host_load_quiz`. On success, the confirmation string is shown and round controls appear. On failure, an inline error is shown and the form remains editable.

### FR-4: Round control — start and reveal
Marcus must be able to start a round and reveal questions one by one. Each reveal sends `host_reveal_question`. A counter shows how many questions have been revealed. The "End Round" button appears only when all questions are revealed.

### FR-5: Scoring panel
After ending a round (and beginning scoring), Marcus must see each question with its correct answer and each team's submitted answer. Each answer row has "Correct" and "Wrong" buttons. Clicking sends `host_mark_answer`. Running totals update on each verdict.

### FR-6: Publish scores
After marking all answers, Marcus must be able to publish the scores for the round. Sending `host_publish_scores` triggers a `scores_published` broadcast. Controls then offer "Start Ceremony" or "Start Next Round".

### FR-7: Ceremony (answer walkthrough)
Marcus must be able to walk the room through answers question by question. Each "Show Next Question" sends `host_ceremony_show_question` to the display + play rooms. "Reveal Answer" sends `host_ceremony_reveal_answer` to the display room only.

### FR-8: End game
Marcus must be able to end the game after all rounds. Sending `host_end_game` triggers a `game_over` broadcast. The host panel shows the final leaderboard.

### FR-9: Reconnect handling
When the WebSocket drops mid-game (after a prior successful connection), the client must reconnect with exponential backoff. Status shows "Reconnecting...". Client-side game state is preserved. If 10 attempts fail, a "Could not reconnect. Please reload." message appears with a reload button.

---

## Non-Functional Requirements

### NFR-1: Connection status must be trusted
The host must never see "Connected" while actually disconnected. A wrong signal breaks the trust loop and makes the tool unreliable.

### NFR-2: Controls visible only when meaningful
"Start Round" must not appear before a quiz is loaded. "End Round" must not appear before all questions are revealed. Preventing invalid actions reduces errors and cognitive load.

### NFR-3: Error messages must state what to do
"Connection refused — invalid token. Check HOST_TOKEN and reload." is acceptable. "Error" alone is not. Every error shown to Marcus must tell him the next action.

### NFR-4: No technical jargon in UI
"WebSocket", "HTTP 403", "JSON", "payload" must not appear in user-facing text. The UI speaks in game domain language.

---

## Constraints (from architecture)

- `Host.tsx` may only communicate with the backend via WsClient (WebSocket)
- No REST endpoints for the host — all commands are WebSocket messages
- `QuizFull`/`QuestionFull` types must never appear in the handler package — the frontend receives `QuizLoadedMeta` shape only
- The host auth is token-based via URL query param; no login form required
- The existing event schema in `messages.ts` / `events.ts` is the contract; no new events may be added by the frontend unilaterally

---

## Open Questions (resolved)

| Question | Resolution |
|---|---|
| Does scoring panel need team answers server-sent, or does client accumulate from events? | The `score_updated` event carries running totals but not submitted answers. A state snapshot or dedicated scoring data event is needed for the scoring panel. Flagged as IC-4. |
| Is `host_end_round` + `host_begin_scoring` two separate commands or one? | Two commands in sequence (confirmed in handler.go lines 207-231). |
| Are answers only sent to display room or play room during ceremony? | Answers (`ceremony_answer_revealed`) go to display room only. Questions go to display + play. Confirmed in handler.go lines 281-298. |
