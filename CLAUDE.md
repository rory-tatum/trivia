## Development Paradigm

This project follows the **object-oriented** paradigm with ports-and-adapters (hexagonal) architecture.

- **Backend**: Go 1.23 — use `@nw-software-crafter` for implementation
- **Frontend**: TypeScript/React 18 — use `@nw-software-crafter` for implementation
- Domain core (`game` package) must have zero infrastructure imports
- `QuestionFull`/`QuizFull` must never appear in `handler` or `hub` packages

## Mutation Testing Strategy

This project uses **per-feature** mutation testing. Runs after refactoring during each delivery, scoped to modified files. Kill rate gate: >= 80%.
