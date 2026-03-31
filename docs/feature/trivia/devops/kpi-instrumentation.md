# KPI Instrumentation Design -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DEVOPS
- Date: 2026-03-29
- Decision: D5 (observability deferred -- this document is future-readiness design only)
- Source KPIs: docs/feature/trivia/discuss/outcome-kpis.md

---

## Purpose

Observability setup is deferred for this personal tool (D5). This document records WHAT would
be instrumented for each outcome KPI if observability were enabled. It serves as:

1. A future-readiness spec: when the developer decides to add instrumentation, the design is already done.
2. An acceptance test guide: the measurement methods described for each KPI map directly to manual validation tests that can be run without observability tooling.

Nothing in this document requires action during the current delivery.

---

## Instrumentation Approach (When Enabled)

If observability is added in a future release, the recommended approach for a personal tool
running on Docker Compose is:

- **Structured JSON logs** from the Go server (using `log/slog` stdlib, Go 1.21+)
- **No external metrics backend** needed -- log-based measurement is sufficient for personal use
- **Event emission pattern**: key game events emit structured log entries that can be parsed for timing analysis

This is the simplest viable instrumentation: no Prometheus, no Grafana, no OpenTelemetry
agent required. Log files + manual analysis or simple shell scripts suffice.

---

## KPI-01: Round Scoring Time

**KPI:** Quizmaster scores a complete round in under 3 minutes (target: < 3 min vs 10-15 min baseline).

**Measurement method (manual, no instrumentation needed):**
Stopwatch from "Open Scoring" click to "Start Ceremony" click in a test game session.

**Future instrumentation design:**

Events to emit:
```json
{ "event": "scoring_opened",    "game_id": "abc", "round_index": 0, "ts": "2026-03-29T20:00:00Z" }
{ "event": "scoring_completed", "game_id": "abc", "round_index": 0, "ts": "2026-03-29T20:02:45Z",
  "duration_seconds": 165, "teams_scored": 5, "questions_scored": 10 }
```

Derived metric: `scoring_completed.duration_seconds` per round.
Alert threshold (future): log a WARNING if `duration_seconds > 180`.

Dashboard (future): Table of round scoring durations per game session.

---

## KPI-02: Player Answer Loss Rate on Refresh

**KPI:** 0% of page refreshes result in lost answers (target: 0% loss).

**Measurement method (automated test, no production instrumentation needed):**
Existing acceptance test: write N answers, refresh, verify N answers restored from localStorage + server draft.

**Future instrumentation design:**

Events to emit:
```json
{ "event": "team_rejoin",    "team_id": "t1", "device_token": "...", "draft_restored": true,
  "draft_answer_count": 7, "ts": "..." }
{ "event": "draft_answer",   "team_id": "t1", "round_index": 0, "question_index": 3,
  "source": "localStorage_restore", "ts": "..." }
```

Derived metric: `team_rejoin.draft_restored` = true for all rejoin events = 100% restore rate.
Loss event: `team_rejoin` where `draft_answer_count` < expected answers for current round state.

Alert threshold (future): log an ERROR if `draft_restored = false` on any rejoin event.

---

## KPI-03: Player Onboarding Time (Join to First Answer)

**KPI:** New player opens /play and enters first answer in under 60 seconds.

**Measurement method (manual, no instrumentation needed):**
Stopwatch with a new user who has never seen the app.

**Future instrumentation design:**

Events to emit:
```json
{ "event": "team_registered", "team_id": "t1", "ts": "2026-03-29T20:05:00Z" }
{ "event": "draft_answer",    "team_id": "t1", "round_index": 0, "question_index": 0,
  "ts": "2026-03-29T20:05:42Z" }
```

Derived metric: time delta between `team_registered` and first `draft_answer` for `team_id`.
Alert threshold (future): log a WARNING if first-answer latency > 60 seconds (indicates UX friction).

---

## KPI-04: Quizmaster Setup Time

**KPI:** Quizmaster completes setup (load YAML, share URLs, start game) in under 2 minutes.

**Measurement method (manual, no instrumentation needed):**
Stopwatch from opening /host to clicking "Start Game".

**Future instrumentation design:**

Events to emit:
```json
{ "event": "quiz_loaded",   "file_path": "science.yml", "round_count": 3,
  "question_count": 30, "ts": "2026-03-29T20:04:00Z" }
{ "event": "game_started",  "round_index": 0, "team_count": 5,
  "ts": "2026-03-29T20:05:45Z" }
```

Derived metric: time delta between `quiz_loaded` and `game_started`.
Note: this metric captures YAML-load-to-start, not the full setup flow (URL sharing is offline).

---

## KPI-05: Answer Submission Completeness

**KPI:** 100% of teams submit before scoring opens (override never needed).

**Measurement method:**
The /host submission status panel already shows this. Log-based measurement is a complement.

**Future instrumentation design:**

Events to emit:
```json
{ "event": "scoring_opened", "game_id": "abc", "round_index": 0,
  "total_teams": 5, "submitted_teams": 5, "forced": false, "ts": "..." }
{ "event": "scoring_opened", "game_id": "abc", "round_index": 1,
  "total_teams": 5, "submitted_teams": 4, "forced": true, "ts": "..." }
```

Derived metric: `forced = true` on any `scoring_opened` event = KPI-05 not met for that round.
Alert threshold (future): log a WARNING when `forced = true` (quizmaster had to chase a team).

Dashboard (future): Per-round submission completeness rate across sessions.

---

## KPI-06: Quizmaster Satisfaction

**KPI:** Quizmaster states they would use the app again (100% binary -- personal use).

**Measurement method:** Post-game verbal feedback ("Was this better than paper?").

**Instrumentation design:** Not instrumentable via software. This is a qualitative outcome.

The proxy metric is continued usage: if the quizmaster runs another game, the tool succeeded.
A proxy instrumentation could log:

```json
{ "event": "game_over", "game_id": "abc", "session_count_lifetime": 3,
  "total_rounds": 4, "ts": "..." }
```

`session_count_lifetime` incrementing over time is a proxy for quizmaster satisfaction.
Requires persistent storage (currently out of scope -- in-memory only, DEC-004).

---

## KPI-07: Display Information Security (Zero Leaks)

**KPI:** /display never shows answer content before ceremony reveal (0 leaks).

**Measurement method (manual verification):**
Play a full game session; check /display never shows answer fields or unrevealed questions.
The boundary_test.go reflection test also provides automated assurance.

**Future instrumentation design:**

The structural type boundary (QuestionPublic has no answer fields) makes this instrumentation
almost unnecessary -- a violation would be a compile error or go-arch-lint failure, not a
runtime event.

However, for defense-in-depth, the answer boundary enforcer could log:

```json
{ "event": "answer_stripped",  "from_type": "QuestionFull", "to_type": "QuestionPublic",
  "destination_room": "play", "question_index": 3, "ts": "..." }
{ "event": "ceremony_reveal",  "destination_room": "display", "question_index": 3,
  "answer_included": true, "ts": "..." }
```

A log line where `answer_included = true` AND `destination_room = "play"` would be a CRITICAL
security violation alert. This event should never occur.

---

## Instrumentation Implementation Notes (When Enabling)

### Recommended: structured JSON logging with log/slog

```go
// Go 1.21+ stdlib -- no external dependency
import "log/slog"

slog.Info("scoring_opened",
    "game_id", gameID,
    "round_index", roundIndex,
    "total_teams", totalTeams,
    "submitted_teams", submittedTeams,
    "forced", forced,
)
```

Configure at startup to output JSON to stdout:

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
slog.SetDefault(slog.New(handler))
```

Docker Compose captures stdout by default. Logs are accessible via `docker compose logs trivia`.

### Log retention (future)

If log history is needed across sessions, mount a log volume:

```yaml
volumes:
  - ./logs:/logs
```

And configure slog to write to `/logs/trivia.log` with rotation.

### No external services required

All instrumentation described above uses only:
- Go stdlib `log/slog` (zero new dependencies)
- `docker compose logs` for access
- Optional: `jq` for shell-based log analysis

No Prometheus, no Grafana, no OpenTelemetry, no external logging service needed
for a personal tool at this scale.
