# ADR-004: Answer-Stripping Boundary Pattern

## Status

Accepted

## Date

2026-03-29

## Context

DEC-010 (DISCUSS wave) established a critical security invariant: correct answers in the quiz content tree must never be sent to `/play` or `/display` clients until the ceremony reveal for that specific question.

This is not a feature preference -- it is a game integrity invariant. A single serialization mistake that exposes answer fields to the player-facing WebSocket room would destroy the game. Players could trivially inspect their browser's WebSocket traffic to see correct answers.

The question is: what enforcement mechanism makes this invariant impossible to accidentally violate?

## Decision

**Structural dual-type boundary:** Maintain two distinct Go struct types for question data.

- `QuestionFull` (and `QuizFull`, `RoundFull`): Server-internal types in the `game` package. Contain `Answer string` and `Answers []string` fields. Exported only within the `game` package.
- `QuestionPublic` (and `QuizPublic`, `RoundPublic`): Transport-safe types. Identical structure to Full types except `Answer` and `Answers` fields are absent by design.

The `boundary.go` file in the `game` package contains the single conversion function `StripAnswers(QuestionFull) QuestionPublic`.

**Package dependency rule:** The `handler` and `hub` packages must never import `QuestionFull` or `QuizFull`. This is enforced by `go-arch-lint` in CI. A violation causes the CI pipeline to fail at the lint stage, not at runtime.

**Ceremony exception:** `ceremony_answer_revealed` events include `answer: string` in the payload, but these events are only broadcast to the `/display` room (not `/play`). The answer field in this event is a separate field on the event struct, not a `QuestionFull` reference.

## Alternatives Considered

### Alternative A: Runtime field stripping (single type with omit logic)

Use a single `Question` type with `Answer` and `Answers` fields. Add serialization logic that omits these fields when serializing for transport to /play or /display clients.

- **Pro:** Simpler type hierarchy. One Question type everywhere.
- **Con:** The omission logic is a runtime behavior that can be bypassed by:
  - Adding a new serialization path that forgets to apply the omit logic
  - A future developer adding a new event type that accidentally includes the full struct
  - A JSON struct tag change that re-enables the field
- **Rejected:** Runtime filtering is fragile. The invariant must be enforced at compile time.

### Alternative B: JSON struct tags (`json:"-"` on answer fields)

Mark `Answer` and `Answers` with `json:"-"` on the single Question type so they are never serialized.

- **Pro:** Simplest approach. No separate type hierarchy.
- **Con:** `json:"-"` prevents serialization in ALL contexts -- the scoring interface also needs to read answer fields to display them to the quizmaster. Using `json:"-"` globally breaks the scoring workflow, or requires a parallel type just for scoring (which is effectively the dual-type approach anyway, but less clear).
- **Rejected:** Insufficient for the use case; solving this properly requires either the dual-type approach or complex JSON marshalling customization.

### Alternative C: Separate serialization wrapper types per endpoint

Define per-endpoint DTOs (Data Transfer Objects) that explicitly map from domain types.

- **Pro:** Very explicit about what each endpoint receives.
- **Con:** More types than needed. For this application, the distinction is binary: full (server-internal) vs public (transport). Two types is the right granularity.
- **Partially accepted:** The dual-type approach IS a form of this -- QuestionPublic is the single transport DTO for all client-facing payloads. The ceremony answer is a separate event field, not a DTO.

## Consequences

### Positive

- Answer leakage is a **compile-time or CI-lint error**, not a runtime bug
- Any developer adding a new WebSocket event that accidentally includes QuestionFull will get a lint failure before the code merges
- The `boundary.go` function is the single, easily-audited location for the Full->Public conversion
- TypeScript event types on the frontend mirror this boundary: the `QuestionPublic` interface in `ws/events.ts` has no answer fields, so TypeScript will not compile code that tries to access `question.answer` from a revealed question event

### Negative

- Two parallel type hierarchies (Full and Public) must be kept in sync as the schema evolves
- When adding new question fields (e.g., for Release 3+ media), developer must remember to add to both QuestionFull and QuestionPublic (or consciously exclude from Public)
- Mitigation: a unit test that confirms QuestionPublic contains no answer fields (asserting via reflection that the struct has no field named `Answer` or `Answers`) catches accidental additions

## Enforcement

`go-arch-lint` configuration (`.go-arch-lint.yml`) includes:
```yaml
rules:
  - name: "answer-boundary"
    deny:
      - package: "internal/game"
        types: ["QuestionFull", "QuizFull", "RoundFull"]
    in:
      - "internal/handler"
      - "internal/hub"
```

This rule is checked in the CI pipeline on every pull request and on every commit to main.
