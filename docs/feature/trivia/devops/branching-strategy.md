# Branching Strategy -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DEVOPS
- Date: 2026-03-29
- Decision: D8 (trunk-based development)

---

## Strategy: Trunk-Based Development

All development happens on `main`. The branch is always in a releasable state.
There are no long-lived feature branches. Short-lived branches (< 1 day) are acceptable
for work-in-progress that is not yet ready to push, but integration happens to `main` directly.

This is the appropriate strategy for a solo developer personal tool:
- No merge conflict overhead
- Every commit is immediately tested by CI
- The latest `main` is always the authoritative version

---

## Branch Rules

### main branch

- Single integration branch and production branch
- Every push triggers the full CI pipeline (`frontend-checks`, `backend-checks`, `build-image`, `mutation-test`)
- The image artifact produced from `main` is the deployable artifact
- Direct pushes are allowed (solo developer -- no PR review requirement)

### Short-lived local branches (optional)

When working on a change that spans multiple logical steps, a local branch is acceptable:

```
git checkout -b wip/game-state-machine   # local only
# ... multiple commits ...
git checkout main
git merge --squash wip/game-state-machine
git commit -m "feat(game): implement state machine transitions"
git push origin main
```

The short-lived branch is never pushed to origin. Only the squashed commit reaches `main`.
This is optional -- direct commits to `main` are the default.

---

## Commit Conventions

All commits to `main` follow Conventional Commits (https://www.conventionalcommits.org/).
This is enforced by developer discipline, not automated (no commitlint hook required for a solo project).

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | When to use |
|------|-------------|
| `feat` | New user-visible feature (US-xx implemented) |
| `fix` | Bug fix |
| `refactor` | Code change with no behavior change |
| `test` | Adding or fixing tests |
| `ci` | CI pipeline changes |
| `build` | Dockerfile, docker-compose, build script changes |
| `docs` | Documentation only |
| `chore` | Dependency updates, config changes |

### Scopes (aligned to Go package structure)

| Scope | Package/area |
|-------|-------------|
| `game` | internal/game |
| `hub` | internal/hub |
| `handler` | internal/handler |
| `quiz` | internal/quiz |
| `media` | internal/media |
| `config` | config |
| `frontend` | frontend/ |
| `docker` | Dockerfile, compose |
| `ci` | .github/workflows |

### Examples

```
feat(game): implement LOBBY -> ROUND_ACTIVE state transition
fix(hub): prevent nil pointer on WebSocket disconnect during scoring
refactor(game): extract answer stripping to boundary.go
test(game): add reflection test asserting QuestionPublic has no answer fields
ci: add go-arch-lint step to backend-checks job
build(docker): pin distroless image to SHA digest
```

---

## CI Gates on Push to main

Every push to `main` must pass all hard-failure gates before the commit is considered integrated:

| Gate | Failure action |
|------|---------------|
| TypeScript type errors | Developer fixes before next push |
| ESLint errors | Developer fixes before next push |
| Vite build failure | Developer fixes before next push |
| `go vet` findings | Developer fixes before next push |
| `go test -race` failures | Developer fixes before next push (no merges with failing tests) |
| `go-arch-lint` violations | Developer fixes before next push (architecture boundary violation) |

The mutation testing gate is a soft failure -- the developer reviews the gremlins report
and addresses escaped mutants in the next delivery cycle.

---

## Release Process

There is no formal release process. Deployment is manual (D1: on-premise, personal tool).

**Informal release flow:**

1. All CI gates pass on `main`
2. Developer decides the feature is complete and ready to deploy
3. Developer builds locally: `docker compose build`
4. Developer deploys: `docker compose down && docker compose up -d`
5. Developer verifies the app in the browser

**Optional: tagging a stable version**

For future reference, the developer may tag a known-good commit:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Tags are not required and do not trigger additional CI.

---

## git Configuration Recommendations

```bash
# Set author identity
git config user.name "Marcus"
git config user.email "marcus@example.com"

# Default branch name
git config init.defaultBranch main

# Recommended: sign commits (optional for personal project)
# git config commit.gpgsign true
```

### .gitignore additions

```
# Environment secrets
.env

# Docker artifacts
trivia-image.tar.gz

# Frontend build output
frontend/dist/
frontend/node_modules/

# Go build artifacts
server
/bin/
```

---

## Pre-commit Hooks (Shift-Left Quality)

Local pre-commit hooks mirror the remote CI checks, catching issues before they reach GitHub.

Install via `.git/hooks/pre-commit` or via `pre-commit` framework (https://pre-commit.com/).

Recommended hooks:

```bash
#!/bin/bash
# .git/hooks/pre-commit

set -e

echo "Running pre-commit checks..."

# Backend: go vet
echo "  go vet..."
go vet ./...

# Backend: go test (fast subset -- skip -race for speed, full race check is in CI)
echo "  go test..."
go test ./... -count=1 -short

# Backend: go-arch-lint
echo "  go-arch-lint..."
go-arch-lint check

# Frontend: type-check (only if frontend files changed)
if git diff --cached --name-only | grep -q '^frontend/'; then
  echo "  TypeScript type-check..."
  cd frontend && npx tsc --noEmit && cd ..
fi

echo "Pre-commit checks passed."
```

Make the hook executable: `chmod +x .git/hooks/pre-commit`

The pre-push hook runs the race detector and full test suite (slower, only on push):

```bash
#!/bin/bash
# .git/hooks/pre-push

set -e
echo "Running pre-push checks..."
go test ./... -race -count=1
echo "Pre-push checks passed."
```
