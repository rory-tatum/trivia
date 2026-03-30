# Prioritization -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DISCUSS -- Phase 2.5
- Date: 2026-03-29

---

## Prioritization Criteria

Stories are prioritized by:

1. **Outcome impact** -- does this story unblock the core value proposition (quizmaster runs a full game)?
2. **JTBD score** -- which job stories does this serve? (see jtbd-analysis.md)
3. **Risk** -- does this story resolve a high-risk integration point?
4. **Dependency** -- does another story require this to be built first?

---

## Release 1 Story Priority Order

Release 1 implements the walking skeleton. Within Release 1, stories are ordered by dependency chain:

| Priority | Story | Title | Dependency | JTBD |
|----------|-------|-------|------------|------|
| 1 | US-01 | Load YAML quiz | None | JS-01 |
| 2 | US-02 | Game lobby -- share URLs | US-01 | JS-01 |
| 3 | US-04 | Player joins (first visit) | US-02 | JS-06 |
| 4 | US-05 | Auto-rejoin on refresh | US-04 | JS-06 |
| 5 | US-03 | Start game -- broadcast | US-02, US-04 | JS-01, JS-02 |
| 6 | US-07 | Quizmaster reveals questions | US-03 | JS-02 |
| 7 | US-08 | Player enters/edits answers | US-07 | JS-07 |
| 8 | US-10 | QM monitors submission status | US-07 | JS-04 |
| 9 | US-09 | Player submits answers (confirm + lock) | US-08 | JS-08 |
| 10 | US-12 | Quizmaster scoring interface | US-09, US-10 | JS-04 |
| 11 | US-14 | Auto-tally scores | US-12 | JS-04 |
| 12 | US-15 | Answer ceremony (step through) | US-14 | JS-05 |
| 13 | US-16 | Round scores on /display | US-15 | JS-05 |
| 14 | US-18 | Advance to next round | US-16 | JS-02 |
| 15 | US-19 | Final scores and winner | US-18 | JS-05 |

---

## Release Priority Rationale

| Release | Outcome | Business Value | Effort | Priority |
|---------|---------|---------------|--------|----------|
| R1 | Walking skeleton -- full text game | Delivers core value prop | High | Must-have |
| R2 | Display question on TV in real-time | Eliminates "reading aloud" friction; core to experience | Medium | High |
| R3 | Media questions (image/audio/video) | Music rounds are critical for trivia nights | Medium | High |
| R4 | MC + multi-part question types | Expands question variety; author uses these | Medium | Medium |
| R5 | Resilience (reconnect, late join) | Reduces quizmaster intervention needed | Medium | Medium |

---

## Deferred (Not in MVP Scope)

| Item | Reason Deferred |
|------|----------------|
| Visual YAML editor | Addressed by DEC-002 -- target user is YAML-fluent |
| Multi-game history / persistent scores | Addressed by DEC-004 -- single session, no DB |
| User accounts / auth | Addressed by DEC-004 |
| Keyboard shortcuts for scoring | Enhancement -- not blocking |
| Export scores to CSV | Nice-to-have -- verbal announcement sufficient for personal use |
| Mobile app (iOS/Android) | Web-based is sufficient for in-room use on WiFi |
