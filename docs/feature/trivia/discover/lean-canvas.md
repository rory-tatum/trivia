# Lean Canvas -- Trivia Game

## Discovery Metadata

- Feature ID: trivia
- Phase: 4 -- Market Viability
- Date: 2026-03-28
- Evidence basis: Validated through Phases 1-3
- Status: GATE G4 PASSED

---

## Lean Canvas

### Problem (Validated -- Phase 1)

**Top 3 problems:**

1. Collecting and scoring paper answer sheets is time-consuming, error-prone, and kills event momentum between rounds.
2. Quizmasters have no way to display questions publicly on a shared screen without exposing their private controls and upcoming content.
3. There is no persistent, device-aware team identity -- players lose their answers on refresh or reconnect.

**Existing alternatives:**
- Paper answer sheets + verbal scoring (universal default, high friction)
- Kahoot (buzzer-style, not round-based trivia, no team answer sheets)
- Jackbox Party Packs (pre-built content only, no custom quiz loading)
- Trivia Crack / pub quiz apps (async, not synchronous host-led)
- Google Forms (answer collection only, no game pacing or scoring interface)

**Gap:** No tool exists that combines: host-controlled pacing + team answer sheets + quizmaster scoring interface + dual display (private/public) + custom YAML content.

---

### Customer Segments (Validated -- Phase 1)

**Primary segment -- Job-to-be-done framing:**

> "When I host a trivia night with friends, I want to run a smooth game without logistics overhead so I can enjoy the evening rather than just administrate it."

**Profile:**
- Socially organized adults who host regular game nights (monthly or more)
- Comfortable with technology (can write/edit a YAML file or learn to)
- Groups of 4-30 players, typically 2-8 teams
- Venue: home, casual pub, office social event
- Device access: quizmaster on laptop, players on smartphones

**Early adopter signal (brief author):** Already using YAML as mental model for quiz structure. Wants digital delivery to replace paper workflow they currently manage manually.

---

### Unique Value Proposition

**For quizmasters:**
> "Load your quiz, control the room, score answers in minutes -- without a single sheet of paper."

**For players:**
> "Answer from your phone, edit until you submit, never lose your answers to a browser refresh."

**Differentiator:** The only tool that separates quizmaster controls, player answer sheets, and public display into three purpose-built interfaces driven by a single YAML file.

---

### Solution (Validated -- Phase 3)

**Three-interface architecture:**

1. **Quizmaster Interface (`/host`)** -- Load YAML, start game, control question/round reveal, view submission status, score answers, trigger public answer ceremony
2. **Player Interface (`/play`)** -- Join game, create team, see revealed questions, enter/edit answers, submit at round end
3. **Public Display (`/display`)** -- Current question only (safe for TV/cast), answer ceremony walkthrough post-round

**Key features validated in Phase 3:**
- YAML schema supporting text, image, audio, video, multi-part (ordered/unordered), and multiple choice questions
- Persistent team identity via browser localStorage
- Side-by-side scoring view (expected answer vs. all team answers)
- Auto-calculated round and running scores
- Post-round answer reveal ceremony on public display

---

### Channels

**Primary (personal use context):**
- Direct deployment by quizmaster on personal hosting (VPS, home server, Raspberry Pi)
- Self-hosted via Docker or simple Node/Python server
- Shared via group chat link at game time

**Secondary (if distributed):**
- GitHub open source release (social reach through developer community)
- Word of mouth at events ("What app is this? Where can I get it?")

**Channel risk:** Low -- this is a personal-use tool. The quizmaster IS the distribution channel. No sales motion required.

---

### Revenue Streams

**Context:** Personal social app. Primary success metric is quizmaster enjoyment and reduced friction, not revenue.

**Revenue model: Not applicable for personal use.**

If ever distributed:
- Open source with optional hosted version (Patreon / one-time purchase)
- No freemium/subscription complexity warranted for this use case

**Viability note:** Zero revenue pressure. Success = quizmaster runs better trivia nights. Unit economics are irrelevant at this scale.

---

### Cost Structure

**Personal deployment costs:**
- Development time (one-time)
- Hosting: $0-5/month (self-hosted or minimal VPS)
- Domain: $10-15/year (optional)
- Media storage: negligible (local files served from same directory)

**No ongoing operational cost pressure.**

---

### Key Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Time to score a round (5 teams, 10 questions) | < 3 minutes | Quizmaster timing |
| Player answer loss rate on refresh | 0% | localStorage persistence test |
| Quizmaster setup time (YAML load to game start) | < 2 minutes | End-to-end timing |
| Answer submission rate (% teams submitting before scoring) | 100% | Game session log |
| Quizmaster satisfaction | "I'd use this again" | Post-game feedback |

---

### Unfair Advantage

- Custom YAML schema purpose-built for trivia night structure (rounds + questions + media + multi-part)
- Three-interface separation (quizmaster / player / display) is not available in any consumer tool today
- Designed by someone who actually hosts trivia nights -- the brief reflects genuine, lived pain

---

## Four Big Risks Assessment

### 1. Value Risk -- ACCEPTABLE

**Risk:** Does the solution actually remove the pain?
**Evidence:** All top-scored opportunities (Phase 2) have direct feature counterparts in the validated solution (Phase 3). Paper answer sheets are replaced. Scoring interface eliminates manual tallying. Dual display eliminates info leakage.
**Residual risk:** Low. YAML literacy is a minor barrier but the target user (brief author) is already YAML-fluent.

### 2. Usability Risk -- ACCEPTABLE

**Risk:** Can the quizmaster and players actually use the interfaces without confusion?
**Evidence:** 22/22 usability scenarios passed in Phase 3. Three separate interfaces each have a single, focused job. Player interface is mobile-first. Quizmaster interface has clear state machine (setup > reveal > submission > scoring > ceremony).
**Residual risk:** Low. Media file management and YAML errors need clear in-app messaging.

### 3. Feasibility Risk -- ACCEPTABLE

**Risk:** Can this be built with reasonable effort?
**Evidence:**
- Real-time state sync: WebSocket (Socket.io or native WS) -- well-established pattern
- YAML parsing: mature libraries in all major languages
- localStorage persistence: browser native, no backend required
- Media serving: static file serving from same directory as YAML
- No external APIs, no auth system, no database required for basic operation
**Residual risk:** Low. The most complex piece is real-time state sync between quizmaster actions and player/display views. This is a solved problem with WebSockets.

### 4. Viability Risk -- ACCEPTABLE

**Risk:** Is this worth building and maintaining?
**Evidence:**
- Personal tool -- success is personal enjoyment, not commercial return
- Zero ongoing infrastructure cost
- Problem is chronic (every trivia night) not one-time -- tool gets used repeatedly
- No competitive pressure; existing tools don't address this exact workflow
**Residual risk:** None material for a personal-use application.

---

## Go / No-Go Decision

| Dimension | Assessment | Recommendation |
|-----------|------------|----------------|
| Problem validated | YES -- 7 signals, 100% confirmation | GO |
| Opportunity prioritized | YES -- top opportunities score 14-19 | GO |
| Solution concept tested | YES -- 22 scenarios, 100% pass rate | GO |
| All 4 risks acceptable | YES -- value, usability, feasibility, viability all low risk | GO |
| Business/personal viability | YES -- personal use, zero revenue pressure | GO |

**OVERALL DECISION: GO**

Proceed to requirements definition with product-owner.

---

## Gate G4 Evaluation

| Criterion | Target | Result | Status |
|-----------|--------|--------|--------|
| Lean Canvas complete | Required | Done (all 9 sections) | PASS |
| All 4 risks addressed | Required | Value + Usability + Feasibility + Viability | PASS |
| All risks rated acceptable | Required | All 4 rated acceptable | PASS |
| Channels validated | Required | Self-hosted personal tool | PASS |
| Unit economics / viability confirmed | Required | Personal use -- N/A, confirmed viable | PASS |
| Go/No-Go documented | Required | GO | PASS |

**G4 STATUS: PASSED -- Ready for peer review and handoff to product-owner**
