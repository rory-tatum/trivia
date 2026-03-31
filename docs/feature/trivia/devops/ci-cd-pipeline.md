# CI/CD Pipeline -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DEVOPS
- Date: 2026-03-29
- Decisions: D3 (GitHub Actions), D8 (trunk-based), D9 (per-feature mutation testing)

---

## Pipeline Overview

Trunk-based development: every push to `main` triggers the full pipeline.
No feature branches in CI -- all integration happens on `main`.

```
push to main
  |
  +-- [frontend-checks] -----> npm ci -> type-check -> lint -> build
  |
  +-- [backend-checks] ------> go vet -> go test -> go-arch-lint
  |
  (both pass)
  |
  +-- [build-image] ----------> docker build (multi-stage)
  |
  (image built)
  |
  +-- [mutation-test] ---------> gremlins (scoped to changed files, kill rate >= 80%)
```

Frontend checks and backend checks run in parallel. The build-image job waits for both.
Mutation testing runs after the image builds (all tests must pass first).

---

## Stage Details

### Stage 1a: Frontend Checks (parallel)

**Runs on:** ubuntu-latest
**Trigger:** push to main, pull_request to main

Steps:
1. Checkout
2. Setup Node.js 20
3. `npm ci` (deterministic install from package-lock.json)
4. `npx tsc --noEmit` (TypeScript type-check, no output files)
5. `npm run lint` (ESLint with @typescript-eslint, enforces no-explicit-any)
6. `npm run build` (Vite production build -- validates the embed target builds clean)

Quality gates enforced:
- Zero TypeScript type errors
- Zero ESLint errors (warnings allowed, errors fail the build)
- Vite build succeeds (dist/ produced)

### Stage 1b: Backend Checks (parallel)

**Runs on:** ubuntu-latest
**Trigger:** push to main, pull_request to main

Steps:
1. Checkout
2. Setup Go 1.23
3. `go vet ./...` (static analysis -- catches common bugs)
4. `go test ./... -race -count=1` (unit + integration tests with race detector)
5. Install go-arch-lint: `go install github.com/fe3dback/go-arch-lint@latest`
6. `go-arch-lint check` (enforces package dependency rules from `.go-arch-lint.yml`)

Quality gates enforced:
- Zero `go vet` findings
- All tests pass
- Race detector finds no data races
- go-arch-lint finds no boundary violations (QuestionFull never in handler/hub)

#### go-arch-lint enforcement

go-arch-lint reads `.go-arch-lint.yml` at repository root. The critical rules:

```yaml
# .go-arch-lint.yml (design reference -- actual file written by software-crafter)
version: 2
workdir: .
allow:
  depOnAnyVendor: false

components:
  game:
    in: internal/game/**
  hub:
    in: internal/hub/**
  handler:
    in: internal/handler/**
  quiz:
    in: internal/quiz/**
  media:
    in: internal/media/**
  config:
    in: config/**
  cmd:
    in: cmd/**

deps:
  game:
    mayDependOn: []
  hub:
    mayDependOn: [game]
  handler:
    mayDependOn: [game, hub, quiz]
  quiz:
    mayDependOn: [game]
  media:
    mayDependOn: []
  config:
    mayDependOn: []
  cmd:
    mayDependOn: [game, hub, handler, quiz, media, config]

# The critical rule: handler and hub must not reference QuestionFull or QuizFull.
# go-arch-lint enforces at the package level; the boundary_test.go reflection test
# enforces at the type level within the game package.
```

A go-arch-lint violation is a hard CI failure. The pipeline does not proceed to build-image.

### Stage 2: Build Image

**Runs on:** ubuntu-latest
**Needs:** frontend-checks AND backend-checks (both must pass)

Steps:
1. Checkout
2. Set image tag from git SHA: `IMAGE_TAG=trivia:${GITHUB_SHA::8}`
3. `docker build -t $IMAGE_TAG .` (multi-stage build: node -> golang -> distroless)
4. Save image as artifact: `docker save $IMAGE_TAG | gzip > trivia-image.tar.gz`
5. Upload artifact (retained 7 days)

The image is not pushed to a registry (personal tool, local deployment). The tarball artifact
can be downloaded and loaded with `docker load < trivia-image.tar.gz` if needed.

### Stage 3: Mutation Testing

**Runs on:** ubuntu-latest
**Needs:** build-image (all prior gates passed)

Tool: `gremlins` (https://github.com/go-gremlins/gremlins) -- Go mutation testing tool.
License: Apache 2.0.

Strategy: per-feature, scoped to files changed in this push (D9 decision).

Steps:
1. Checkout with full history (needed to compute changed files)
2. Setup Go 1.23
3. Install gremlins: `go install github.com/go-gremlins/gremlins/cmd/gremlins@latest`
4. Compute changed Go files: `git diff --name-only HEAD~1 HEAD -- '*.go'`
5. Run gremlins on changed packages only
6. Assert kill rate >= 80%

Kill rate gate: if escaped mutants exceed 20% of surviving mutants, the job fails.
This is a soft gate for this pipeline stage -- a failure is reported but does not block
the artifact. The developer is expected to add tests before the next delivery.

Note on gate severity: per D9 (per-feature strategy, personal tool), mutation testing is
a quality signal, not a deployment blocker. The developer reviews the report and addresses
escaped mutants in the same or next commit.

---

## GitHub Actions Workflow File

File location: `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  frontend-checks:
    name: Frontend Checks
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: "20"
          cache: "npm"
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        run: npm ci

      - name: TypeScript type-check
        run: npx tsc --noEmit

      - name: Lint
        run: npm run lint

      - name: Build
        run: npm run build

  backend-checks:
    name: Backend Checks
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.23"
          cache: true

      - name: go vet
        run: go vet ./...

      - name: go test
        run: go test ./... -race -count=1

      - name: Install go-arch-lint
        run: go install github.com/fe3dback/go-arch-lint@latest

      - name: go-arch-lint
        run: go-arch-lint check

  build-image:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: [frontend-checks, backend-checks]
    steps:
      - uses: actions/checkout@v4

      - name: Set image tag
        id: tag
        run: echo "tag=trivia:${GITHUB_SHA::8}" >> $GITHUB_OUTPUT

      - name: Build Docker image
        run: docker build -t ${{ steps.tag.outputs.tag }} .

      - name: Export image artifact
        run: docker save ${{ steps.tag.outputs.tag }} | gzip > trivia-image.tar.gz

      - name: Upload image artifact
        uses: actions/upload-artifact@v4
        with:
          name: trivia-image-${{ github.sha }}
          path: trivia-image.tar.gz
          retention-days: 7

  mutation-test:
    name: Mutation Testing
    runs-on: ubuntu-latest
    needs: [build-image]
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 2

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.23"
          cache: true

      - name: Install gremlins
        run: go install github.com/go-gremlins/gremlins/cmd/gremlins@latest

      - name: Compute changed Go packages
        id: changed
        run: |
          CHANGED=$(git diff --name-only HEAD~1 HEAD -- '*.go' \
            | xargs -I{} dirname {} \
            | sort -u \
            | sed 's|^|./|' \
            | tr '\n' ' ')
          echo "packages=${CHANGED}" >> $GITHUB_OUTPUT

      - name: Run mutation tests
        if: steps.changed.outputs.packages != ''
        run: |
          gremlins unleash \
            --threshold-efficacy 80 \
            ${{ steps.changed.outputs.packages }}

      - name: No Go changes -- skip mutation testing
        if: steps.changed.outputs.packages == ''
        run: echo "No Go files changed. Skipping mutation testing."
```

---

## Quality Gate Summary

| Gate | Tool | Failure behavior |
|------|------|-----------------|
| TypeScript type errors | `tsc --noEmit` | Hard fail -- blocks image build |
| ESLint errors | `npm run lint` | Hard fail -- blocks image build |
| Vite build failure | `npm run build` | Hard fail -- blocks image build |
| Go vet findings | `go vet` | Hard fail -- blocks image build |
| Test failures | `go test -race` | Hard fail -- blocks image build |
| Race conditions | `-race` flag | Hard fail -- blocks image build |
| Architecture violations | `go-arch-lint` | Hard fail -- blocks image build |
| Mutation kill rate < 80% | `gremlins` | Soft fail -- reported, does not block artifact |

---

## DORA Metrics Design

This pipeline is designed to support the following DORA targets for a solo developer personal tool:

| Metric | Target | How achieved |
|--------|--------|-------------|
| Deployment Frequency | On-demand (after each feature) | Trunk-based: every main push produces a deployable image |
| Lead Time for Changes | Under 15 minutes | Parallel frontend/backend checks; fast Go build |
| Change Failure Rate | < 5% | go-arch-lint prevents the most critical class of regressions (answer leakage) |
| Time to Restore | Under 30 minutes | Recreate strategy + Docker image artifact for rollback |

---

## Pipeline Execution Time Estimates

| Stage | Estimated duration |
|-------|--------------------|
| frontend-checks | 2-3 min (npm ci + tsc + lint + build) |
| backend-checks | 1-2 min (go test + go-arch-lint) |
| build-image | 3-5 min (multi-stage Docker build, layer caching helps) |
| mutation-test | 5-10 min (scoped to changed files) |
| Total (parallel) | ~10-15 min end-to-end |
