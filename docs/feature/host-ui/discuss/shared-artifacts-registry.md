# Shared Artifacts Registry — host-ui

**Feature ID**: host-ui
**Date**: 2026-04-09

This registry is the single source of truth for all runtime artifacts shared across steps in the Host UI journey. Every `${variable}` used in mockups and every field in state maps to one entry here.

---

## Registry

### `token`
- **Type**: `string`
- **Source**: URL query param `?token=` read in `getHostWsUrl()` in `Host.tsx`
- **Populated at**: Step 1 (page load)
- **Consumed by**: WsClient (appended to WS URL `/ws?token=<value>`), Host.tsx (build WS URL)
- **Integration checkpoint**: IC-1 — If absent, WS upgrade proceeds but auth guard returns 403
- **Validation rule**: Exact equality match against `HOST_TOKEN` env var on server

---

### `connected`
- **Type**: `boolean`
- **Source**: Set to `true` ONLY inside WsClient `onOpen` callback (not synchronously after `connect()`)
- **Populated at**: Step 2 (onOpen fires)
- **Cleared at**: reconnect_failed or auth failure
- **Consumed by**: Connection status indicator in all render branches
- **Integration checkpoint**: IC-1 — requires WsClient to expose an `onOpen` hook

---

### `gamePhase`
- **Type**: enum — `idle | quiz_loaded | round_active | scoring | ceremony | game_over`
- **Source**: Derived from WS events
  - `idle`: set at Step 2 (connected)
  - `quiz_loaded`: set on receiving `quiz_loaded` event
  - `round_active`: set on receiving `round_started` event
  - `scoring`: set on receiving `scoring_opened` event (after `host_begin_scoring`)
  - `ceremony`: set when host initiates ceremony (after `host_publish_scores`)
  - `game_over`: set on receiving `game_over` event
- **Consumed by**: Host.tsx render branch selector (all panels)
- **Note**: Only one phase is active at a time; transitions are linear with no branching back

---

### `quizMeta`
- **Type**: `QuizLoadedMeta` — `{ title, round_count, question_count, player_url, display_url, confirmation, session_id }`
- **Source**: `quiz_loaded` event payload from server
- **Populated at**: Step 3 (successful `host_load_quiz`)
- **Consumed by**:
  - Confirmation string display (e.g., "Pub Night Vol. 3 | 3 rounds | 15 questions")
  - Round selector (derives how many rounds exist)
  - Player and display URL display panels
- **Integration checkpoint**: IC-5 — server populates `player_url` and `display_url` with actual base URL

---

### `roundIndex`
- **Type**: `number` (0-based)
- **Source**: `round_started` event payload `{ round_index }`
- **Populated at**: Step 4 (Start Round)
- **Updated**: Each time a new round starts
- **Consumed by**: Round active panel, scoring panel, ceremony panel, all `host_*_round` commands

---

### `revealedQuestions`
- **Type**: `Array<{ roundIndex: number, questionIndex: number, text: string }>`
- **Source**: Accumulated `question_revealed` event payloads
- **Populated at**: Step 5 (each Reveal Next Question click)
- **Reset**: On new round start
- **Consumed by**:
  - Round active panel question list
  - Question counter ("N of M revealed")
  - End Round button visibility condition (all questions revealed)

---

### `teamAnswers`
- **Type**: `Array<{ teamId: string, teamName: string, questionIndex: number, answer: string, verdict?: string }>`
- **Source**: Provided in scoring panel state — populated from server state snapshot or via scoring panel data when scoring phase begins
- **Populated at**: Step 6 (begin scoring)
- **Consumed by**: Scoring panel (per-question, per-team rows)
- **Integration checkpoint**: IC-4 — `score_updated` events include `team_name` for display

---

### `scores`
- **Type**: `Map<teamId: string, { teamName: string, score: number }>`
- **Source**: `score_updated` events (incremental updates during scoring)
- **Populated at**: Step 7 (mark answer clicks)
- **Consumed by**:
  - Scoring panel running totals display
  - `game_over` leaderboard (final scores from `game_over` payload)
- **Integration checkpoint**: IC-4 — `score_updated` must include `team_name`

---

### `playerURL`
- **Type**: `string`
- **Source**: `quiz_loaded` payload field `player_url` (server populates as `baseURL + "/play"`)
- **Populated at**: Step 3 (quiz loaded)
- **Consumed by**: Idle/loaded panel URL display for Marcus to share with participants

---

### `displayURL`
- **Type**: `string`
- **Source**: `quiz_loaded` payload field `display_url` (server populates as `baseURL + "/display"`)
- **Populated at**: Step 3 (quiz loaded)
- **Consumed by**: Idle/loaded panel URL display for Marcus to share with the display screen

---

### `authError`
- **Type**: `boolean | string`
- **Source**: WsClient emits a distinct event (e.g., `connection_refused`) when CloseEvent.code is 1006 on first attempt
- **Populated at**: Step 1 error path (auth failure)
- **Consumed by**: Auth error message render branch
- **Integration checkpoint**: IC-2 — requires WsClient to inspect CloseEvent.code and stop retrying on auth failure

---

### `ceremonyCursor`
- **Type**: `number` (0-based question index)
- **Source**: Incremented with each `host_ceremony_show_question` command
- **Populated at**: Step 9 (ceremony start)
- **Consumed by**: Ceremony panel progress counter ("Question N of M shown")

---

## Integration Checkpoints

| ID | Description | Affects Stories |
|----|-------------|-----------------|
| IC-1 | WsClient must expose `onOpen(handler)` hook; `connected` state set only in callback | US-01 |
| IC-2 | WsClient inspects CloseEvent.code on first close; distinguishes auth failure from transient drop | US-01 |
| IC-3 | `host_end_round` must be followed by `host_begin_scoring` (two commands) to enter scoring phase | US-05 |
| IC-4 | `score_updated` events must carry `team_name` (not only `team_id`) for scoring panel display | US-06 |
| IC-5 | `quiz_loaded` payload provides `player_url` and `display_url` for Marcus to share | US-03 |
