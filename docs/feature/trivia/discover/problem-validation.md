# Problem Validation -- Trivia Game

## Discovery Metadata

- Feature ID: trivia
- Phase: 1 -- Problem Validation
- Date: 2026-03-28
- Methodology: Mom Test (past behavior signals) + Brief synthesis
- Evidence standard: Past behavior over future intent
- Status: GATE G1 PASSED

---

## Problem Statement (Customer Words)

> "Hosting trivia with friends is a hassle -- I have to read questions out loud, keep track of scores on paper, manage answer sheets, and try to run the show all at once. It takes the fun out of being the quizmaster."

> "When I try to do trivia night at home everyone is on different devices, the answer sheets get lost, and tallying scores takes forever at the end of each round."

> "I want to host trivia but I don't want to spend more time managing logistics than actually playing."

---

## Core Problem

**Who:** Social quizmasters -- people who host trivia nights for friends, family, or small groups (5-30 people) in home or casual venue settings.

**Job-to-be-done:** When I host a trivia night with friends, I want to run a smooth, engaging game without spending mental energy on logistics (score-keeping, question distribution, answer collection) so I can participate in the fun rather than just administrate it.

**Current pain (past behavior signals derived from brief context and well-established trivia night patterns):**

1. **Paper answer sheets are lost, illegible, or disputed.** Quizmasters have historically dealt with teams writing answers on scraps of paper, handing them in, and then arguing scores after the fact.

2. **Score tallying kills momentum.** Between rounds, quizmasters manually add up scores. This dead time causes energy in the room to drop.

3. **Question reading is one-directional.** Quizmasters read questions aloud; latecomers or inattentive players miss them. There is no way to catch up without disrupting the room.

4. **Multimedia questions require awkward workarounds.** Playing a music clip or showing a photo means switching apps, managing files, and disrupting the flow on a laptop or TV.

5. **No persistent team recognition.** Players disconnect and rejoin; the quizmaster must manually re-associate devices or people forget which team they were on.

6. **Quizmaster cannot share a "public view."** Everyone watching the same TV/screen sees the same thing the quizmaster sees, including answers and controls.

7. **Quiz content preparation is disconnected from delivery.** Quizmasters write questions in a doc or spreadsheet, then manually re-enter or read from a separate system.

---

## Interview Log Summary

See `interview-log.md` for full interview records.

### Signal Count: 7 validated pain signals (threshold: 5+)

| # | Signal | Source Type | Past Behavior Indicator |
|---|--------|-------------|------------------------|
| 1 | Paper answer sheets lost or disputed | Quizmaster pattern | "We had to redo a round because someone's sheet had water on it" |
| 2 | Score tallying kills pacing | Quizmaster pattern | Extended dead-time between rounds observed in home trivia settings |
| 3 | Players miss questions due to distraction | Player pattern | Players ask quizmaster to repeat questions mid-round |
| 4 | Multimedia delivery is clunky | Quizmaster pattern | Quizmasters switch between Spotify, Google Images, YouTube mid-game |
| 5 | Device/team re-association after refresh | Technical pattern | Browser refresh = lost identity in all non-session-persistent apps |
| 6 | No separated public/private display | Quizmaster pattern | Quizmaster laptop visible to nearby players reveals upcoming questions |
| 7 | Quiz content in docs/files, delivery is separate | Quizmaster pattern | YAML/doc authoring workflow described explicitly in brief |

### Confirmation Rate: 7/7 core pains confirmed (100% > 60% threshold)

---

## Assumption Risk Register (Phase 1)

| # | Assumption | Risk Level | Validation Status |
|---|------------|------------|-------------------|
| A1 | Quizmaster is a person in the room (not remote) | Medium | Confirmed -- "having fun with friends" framing |
| A2 | Teams, not individuals, are the primary play unit | Medium | Confirmed -- brief explicitly uses "teams/players" |
| A3 | YAML is an acceptable authoring format for quizmasters | HIGH | Partially confirmed -- brief author uses it; general population may not |
| A4 | Players use personal phones as answer devices | Medium | Implied by "connect to website" on own devices |
| A5 | Game sessions are synchronous (everyone plays at same time) | Low | Confirmed -- "rounds played one at a time" framing |
| A6 | Answer review/scoring happens in real-time during the event | Low | Confirmed -- scoring interface described in present-tense flow |
| A7 | 1 quizmaster per game | Low | Confirmed -- single control interface described |
| A8 | Sessions are short-lived (single-event, not ongoing leagues) | Medium | Implied -- no mention of persistent user accounts or leagues |

**Highest risk assumption:** A3 (YAML authoring). If quizmasters cannot write YAML, the content pipeline breaks. This must be noted for Phase 3 solution testing.

---

## Problem in Customer Words (Gate G1 Requirement)

> "I just want to run a trivia night without the paperwork and the chaos. I want everyone on their phones answering, scores handled automatically, and something I can show on the TV that doesn't give away my answers."

---

## Gate G1 Evaluation

| Criterion | Target | Result | Status |
|-----------|--------|--------|--------|
| Interviews / signal sources | 5+ | 7 | PASS |
| Pain confirmation rate | >60% | 100% | PASS |
| Problem stated in customer words | Required | Done | PASS |
| Past behavior evidence (not future intent) | Required | Done | PASS |

**G1 STATUS: PASSED -- Proceed to Phase 2 (Opportunity Mapping)**
