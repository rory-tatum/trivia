# Data Models -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DESIGN
- Date: 2026-03-29

---

## 1. YAML Schema (OQ-02 input format)

### Release 1 schema (text questions only)

```yaml
title: "Friday Night Trivia -- March 2026"
rounds:
  - name: "Round 1: History"
    questions:
      - text: "What year did World War II end?"
        answer: "1945"
      - text: "Who was the first person to walk on the moon?"
        answer: "Neil Armstrong"
```

### Field specification

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `title` | string | yes | Displayed in lobby and on /display |
| `rounds` | array | yes | Min 1 element |
| `rounds[].name` | string | yes | Displayed as round header |
| `rounds[].questions` | array | yes | Min 1 element |
| `rounds[].questions[].text` | string | yes | The question text |
| `rounds[].questions[].answer` | string | yes* | *One of `answer` or `answers` required |
| `rounds[].questions[].answers` | []string | yes* | *Multi-part; Release 3+ |
| `rounds[].questions[].choices` | []string | no | Multiple choice options; Release 3+ |
| `rounds[].questions[].media` | object | no | Media attachment; Release 3+ |
| `rounds[].questions[].media.type` | enum | no | `image`, `audio`, `video` |
| `rounds[].questions[].media.file` | string | no | Relative path from YAML location |

---

## 2. Go Domain Types

### Server-internal types (game package -- never serialized to client)

```
QuizFull
  Title    string
  Rounds   []RoundFull

RoundFull
  Name      string
  Questions []QuestionFull

QuestionFull
  Index    int
  Text     string
  Answer   string    // single answer -- NEVER sent to /play or /display
  Answers  []string  // multi-part -- NEVER sent to /play or /display
```

### Client-safe types (game package -- safe for transport)

```
QuizPublic
  Title    string
  Rounds   []RoundPublic

RoundPublic
  Name      string
  Questions []QuestionPublic

QuestionPublic
  Index    int
  Text     string
  // Answer and Answers deliberately absent
```

### Game state types

```
GameState (enum string)
  LOBBY
  ROUND_ACTIVE
  ROUND_ENDED
  SCORING
  CEREMONY
  ROUND_SCORES
  GAME_OVER

GameSession
  ID              string        // UUID
  Quiz            QuizFull      // full content, server-internal
  State           GameState
  CurrentRound    int           // 0-indexed
  CurrentQuestion int           // 0-indexed within round; -1 = none yet
  Teams           map[TeamID]Team
  RevealedQuestions []int       // question indices revealed in current round
  Submissions     map[TeamID]RoundSubmission
  ScoredAnswers   map[TeamID]map[RoundIndex]map[QuestionIndex]ScoredAnswer
  RoundScores     map[TeamID][]RoundScore

Team
  ID          TeamID    // UUID
  Name        string
  DeviceToken string    // stored in player localStorage for rejoin

RoundSubmission
  TeamID      TeamID
  RoundIndex  int
  Answers     map[int]string  // question_index -> answer text
  SubmittedAt time.Time

ScoredAnswer
  Answer   string
  Verdict  Verdict   // correct | wrong
  Points   int       // 1 for correct, 0 for wrong (Release 1)

RoundScore
  RoundIndex    int
  RoundPoints   int
  RunningTotal  int
```

---

## 3. WebSocket Message Types

### Outbound -- Server to Client

All messages share the envelope:

```
{
  "event": "<event_name>",
  "payload": { ... }
}
```

#### state_snapshot

Sent on connection/reconnection. Payload varies by game state.

```json
{
  "event": "state_snapshot",
  "payload": {
    "session_id": "uuid",
    "game_state": "ROUND_ACTIVE",
    "current_round": 0,
    "quiz_title": "Friday Night Trivia",
    "teams": [
      { "team_id": "uuid", "team_name": "Team A" }
    ],
    "revealed_questions": [
      { "index": 0, "text": "What year did WW2 end?" }
    ],
    "submission_status": [
      { "team_id": "uuid", "submitted": false }
    ],
    "round_scores": []
  }
}
```

#### question_revealed

```json
{
  "event": "question_revealed",
  "payload": {
    "round_index": 0,
    "question_index": 2,
    "question": {
      "index": 2,
      "text": "Who painted the Mona Lisa?"
    }
  }
}
```

#### submission_ack

```json
{
  "event": "submission_ack",
  "payload": {
    "team_id": "uuid",
    "round_index": 0
  }
}
```

#### ceremony_answer_revealed

```json
{
  "event": "ceremony_answer_revealed",
  "payload": {
    "question_index": 2,
    "answer": "Leonardo da Vinci"
  }
}
```

#### round_scores_published

```json
{
  "event": "round_scores_published",
  "payload": {
    "round_index": 0,
    "scores": [
      { "team_id": "uuid", "team_name": "Team A", "round_score": 6, "running_total": 6 },
      { "team_id": "uuid", "team_name": "Team B", "round_score": 4, "running_total": 4 }
    ]
  }
}
```

#### error

```json
{
  "event": "error",
  "payload": {
    "code": "DUPLICATE_TEAM_NAME",
    "message": "A team named 'The Brains' already exists"
  }
}
```

### Inbound -- Client to Server

```
{
  "event": "<event_name>",
  "payload": { ... }
}
```

#### team_register

```json
{
  "event": "team_register",
  "payload": {
    "team_name": "The Brains",
    "device_token": "random-uuid-from-localstorage"
  }
}
```

#### submit_answers

```json
{
  "event": "submit_answers",
  "payload": {
    "team_id": "uuid",
    "round_index": 0,
    "answers": [
      { "question_index": 0, "answer": "1945" },
      { "question_index": 1, "answer": "Neil Armstrong" }
    ]
  }
}
```

#### host_mark_answer

```json
{
  "event": "host_mark_answer",
  "payload": {
    "team_id": "uuid",
    "round_index": 0,
    "question_index": 1,
    "verdict": "correct"
  }
}
```

---

## 4. Game State Machine (OQ-02)

The state machine is implemented as a struct method on `GameSession`. Each transition is a named method that validates the current state before mutating.

### Valid transitions

| Current State | Event | Next State | Side Effects |
|--------------|-------|-----------|--------------|
| LOBBY | StartRound(roundIndex) | ROUND_ACTIVE | Clears revealed questions for round; broadcasts round_started |
| ROUND_ACTIVE | RevealQuestion(qIndex) | ROUND_ACTIVE | Appends to RevealedQuestions; broadcasts question_revealed |
| ROUND_ACTIVE | EndRound() | ROUND_ENDED | Broadcasts round_ended |
| ROUND_ENDED | OpenScoring() | SCORING | Broadcasts scoring_opened to host room |
| SCORING | MarkAnswer(teamID, qIdx, verdict) | SCORING | Updates ScoredAnswers; checks if all marked; broadcasts mark_applied to host |
| SCORING | StartCeremony() | CEREMONY | Broadcasts ceremony_started |
| CEREMONY | ShowCeremonyQuestion(qIdx) | CEREMONY | Broadcasts ceremony_question_shown (no answer) to display |
| CEREMONY | RevealCeremonyAnswer(qIdx) | CEREMONY | Broadcasts ceremony_answer_revealed (with answer) to display only |
| CEREMONY | CompleteCeremony() | ROUND_SCORES | Calculates RoundScores; broadcasts round_scores_published |
| ROUND_SCORES | StartRound(nextRoundIndex) | ROUND_ACTIVE | (same as first StartRound) |
| ROUND_SCORES | EndGame() | GAME_OVER | Broadcasts game_over with final scores |
| GAME_OVER | (terminal -- no transitions) | -- | -- |

### Invalid transition behavior

Any transition method called in an invalid state returns a domain error. The handler layer converts this to a WebSocket `error` event sent to the requesting client only.

---

## 5. localStorage Schema (Browser)

Players store team identity in localStorage to survive page refresh.

```json
{
  "trivia_team_id": "uuid",
  "trivia_device_token": "uuid",
  "trivia_draft_answers": {
    "round_0": {
      "0": "1945",
      "1": "Neil Armstrong"
    }
  }
}
```

Keys are prefixed `trivia_` to avoid collisions with other apps on the same origin.

---

## 6. Error Code Catalogue

| Code | Trigger | Client action |
|------|---------|--------------|
| `INVALID_TOKEN` | Host request with wrong token | Show 403 page |
| `DUPLICATE_TEAM_NAME` | Register with existing name | Show inline error, allow retry |
| `TEAM_NOT_FOUND` | Rejoin with unknown team_id | Prompt fresh registration |
| `INVALID_STATE_TRANSITION` | Any action invalid for current state | Show transient error banner |
| `ALREADY_SUBMITTED` | Submit twice for same round | Re-send submission_ack (idempotent) |
| `YAML_VALIDATION_ERROR` | Load invalid YAML | Return detailed field-level errors |
| `FILE_NOT_FOUND` | YAML path does not exist | Return path in error message |
| `MEDIA_NOT_FOUND` | Media ref in YAML missing | Return round/question/path in error |
