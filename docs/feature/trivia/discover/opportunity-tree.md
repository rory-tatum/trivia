# Opportunity Tree -- Trivia Game

## Discovery Metadata

- Feature ID: trivia
- Phase: 2 -- Opportunity Mapping
- Date: 2026-03-28
- Framework: Opportunity Solution Tree (OST)
- Status: GATE G2 PASSED

---

## Desired Outcome (Root)

**Quizmaster Outcome:** Run an engaging, low-friction trivia night where I control the pacing and scoring without losing the fun of being with my friends.

**Player Outcome:** Participate in trivia from my phone, submit answers with my team, and see scores unfold in real-time without confusion.

---

## Job-to-Be-Done Map

### Quizmaster JTBD

| Job Step | Current Solution | Friction Level |
|----------|-----------------|----------------|
| 1. Author quiz content | Google Docs, Word, or spreadsheet | Medium -- no standard format |
| 2. Load quiz into delivery tool | Manual copy-paste or none | HIGH -- no dedicated tool |
| 3. Onboard players at game start | Verbal instruction + paper | HIGH -- chaos at start |
| 4. Reveal questions at right pace | Read aloud from paper/screen | Medium -- no control without shouting |
| 5. Display public question view | Share laptop screen / TV cast | HIGH -- leaks private info |
| 6. Collect answers at round end | Paper sheets handed in | HIGH -- sorting, illegibility |
| 7. Score submitted answers | Manual review per sheet | HIGH -- time-consuming, subjective |
| 8. Display round scores | Write on whiteboard / announce verbally | Medium -- not visible to all |
| 9. Advance to next round | Verbal | Low -- easy but can be cleaner |
| 10. End game, announce winner | Verbal + manual tally | Medium -- anticlimactic without ceremony |

### Player JTBD

| Job Step | Current Solution | Friction Level |
|----------|-----------------|----------------|
| 1. Join a game | Given URL/code verbally | Low -- straightforward |
| 2. Register as a team | Paper + verbal | Medium -- redundant on refresh |
| 3. Read/hear current question | Listen to quizmaster | HIGH -- miss questions easily |
| 4. Discuss with team | In person | Low -- social, fine |
| 5. Enter answer | Write on paper | HIGH -- hard to edit, messy |
| 6. Submit answers for round | Hand in paper | HIGH -- irreversible, no confirmation |
| 7. See results | Listen to quizmaster announce | Medium -- no persistent record |

---

## Opportunity Tree (OST)

```
ROOT OUTCOME
  Run a smooth, fun trivia night (quizmaster)
  Participate easily from phone (player)
        |
        +-- OPP-01: Game Setup & Content Loading
        |         |
        |         +-- OPP-01a: YAML-based quiz authoring
        |         +-- OPP-01b: Quiz file validation and preview
        |         +-- OPP-01c: Game session initialization
        |
        +-- OPP-02: Player Onboarding & Identity
        |         |
        |         +-- OPP-02a: Team creation on join
        |         +-- OPP-02b: Persistent device recognition (no re-entry on refresh)
        |         +-- OPP-02c: Rejoin after disconnect
        |
        +-- OPP-03: Question Delivery & Pacing
        |         |
        |         +-- OPP-03a: Quizmaster-controlled question reveal
        |         +-- OPP-03b: Player view shows all revealed questions for round
        |         +-- OPP-03c: Public display view (shareable, quizmaster-controlled)
        |         +-- OPP-03d: Multimedia question support (image/audio/video)
        |
        +-- OPP-04: Answer Entry & Submission
        |         |
        |         +-- OPP-04a: Team answers editable until round submission
        |         +-- OPP-04b: Multi-part answer support (ordered/unordered)
        |         +-- OPP-04c: Multiple choice question support
        |         +-- OPP-04d: Round submission with confirmation
        |
        +-- OPP-05: Scoring & Results
        |         |
        |         +-- OPP-05a: Quizmaster scoring interface (expected vs submitted)
        |         +-- OPP-05b: Fuzzy/subjective answer acceptance (quizmaster decides)
        |         +-- OPP-05c: Per-round score tallying (auto-calculated)
        |         +-- OPP-05d: Running scoreboard visible to quizmaster
        |
        +-- OPP-06: Answer Review Ceremony
                  |
                  +-- OPP-06a: Public display walks through answers post-round
                  +-- OPP-06b: Reveal correct answer per question on public display
                  +-- OPP-06c: Score announcement at round end
```

---

## Opportunity Scoring (Opportunity Algorithm)

Scoring formula: **Importance + (Importance - Satisfaction)** where both are rated 1-10.
Score > 8 = underserved opportunity (threshold for prioritization).

| OPP ID | Opportunity | Importance | Satisfaction (current) | Score | Priority |
|--------|-------------|------------|----------------------|-------|----------|
| OPP-03c | Public display view (TV-shareable, no private info) | 10 | 1 | 19 | P1 |
| OPP-05a | Scoring interface (expected vs submitted answers) | 10 | 1 | 19 | P1 |
| OPP-04a | Team answers editable until submission | 9 | 2 | 16 | P1 |
| OPP-02b | Persistent device recognition | 9 | 2 | 16 | P1 |
| OPP-04d | Round submission with confirmation | 9 | 3 | 15 | P1 |
| OPP-03a | Quizmaster-controlled question reveal | 9 | 3 | 15 | P1 |
| OPP-03b | Player view of all revealed questions | 8 | 3 | 13 | P1 |
| OPP-05c | Per-round auto score tallying | 8 | 2 | 14 | P1 |
| OPP-01a | YAML-based quiz authoring | 9 | 5 | 13 | P1 |
| OPP-03d | Multimedia question support (image/audio/video) | 7 | 2 | 12 | P2 |
| OPP-04b | Multi-part answer support (ordered/unordered) | 7 | 2 | 12 | P2 |
| OPP-04c | Multiple choice question support | 7 | 3 | 11 | P2 |
| OPP-06a | Public display post-round answer walkthrough | 8 | 2 | 14 | P1 |
| OPP-06b | Correct answer reveal on public display | 8 | 2 | 14 | P1 |
| OPP-02a | Team creation on join | 8 | 5 | 11 | P2 |
| OPP-05b | Fuzzy answer acceptance (quizmaster decides) | 9 | 1 | 17 | P1 |
| OPP-01c | Game session initialization | 7 | 4 | 10 | P2 |
| OPP-05d | Running scoreboard for quizmaster | 7 | 3 | 11 | P2 |
| OPP-02c | Rejoin after disconnect | 8 | 3 | 13 | P1 |
| OPP-01b | Quiz file validation and preview | 6 | 3 | 9 | P2 |

---

## Top Prioritized Opportunities (Score > 8, P1)

### Tier 1 -- Critical Path (Score 15+)

1. **OPP-03c: Public display view** (Score: 19) -- Separate shareable URL that shows only current question, no quizmaster controls or upcoming content. Fundamental to the experience.

2. **OPP-05a: Scoring interface** (Score: 19) -- The hardest manual work today. Side-by-side view of expected answer vs. each team's submission with accept/reject controls. Removes the biggest time sink.

3. **OPP-05b: Fuzzy answer acceptance** (Score: 17) -- Quizmaster judgment call per answer. Not algorithmic -- quizmaster sees the answer and clicks correct/incorrect. Removes disputes.

4. **OPP-04a: Team answers editable until submission** (Score: 16) -- Players can change any answer until they click submit at round end. Replaces the irreversibility of paper.

5. **OPP-02b: Persistent device recognition** (Score: 16) -- Local storage or cookie-based team identity. Survive refresh and brief disconnects without quizmaster intervention.

6. **OPP-04d + OPP-05c: Round submission + auto tallying** (Score: 15/14) -- Submission gates scoring; tallying should be automatic once quizmaster grades answers.

7. **OPP-06a/b: Post-round answer ceremony** (Score: 14) -- Public display walks through answers after all teams submit. Turns scoring into an entertainment moment.

### Tier 2 -- High Value (Score 9-14)

- OPP-03a: Quizmaster-controlled reveal pacing
- OPP-03b: Player view of all revealed questions
- OPP-03d: Multimedia support
- OPP-04b/c: Multi-part and multiple choice questions
- OPP-01a: YAML authoring pipeline
- OPP-02c: Reconnect after disconnect

---

## Underserved Need Summary (Top 3 for Phase 3)

1. **The Scoring Bottleneck** -- No good tool exists for quizmasters to quickly compare expected vs. submitted answers and mark right/wrong. This is the most painful, time-consuming step today.

2. **The Dual-Display Problem** -- Quizmasters need one view for themselves (controls + all info) and a separate clean view for the room (current question only). No casual trivia tool separates these well.

3. **The Answer Sheet Replacement** -- Teams need to enter, edit, and submit answers digitally with persistence across refreshes. Paper is the current standard and it fails constantly.

---

## Gate G2 Evaluation

| Criterion | Target | Result | Status |
|-----------|--------|--------|--------|
| OST complete with all job steps mapped | Required | Done | PASS |
| Top opportunities scored | >8 | 19, 19, 17, 16, 16 | PASS |
| Opportunities identified | 5+ | 20 | PASS |
| Job step coverage | 80% | 10/10 quizmaster, 7/7 player = 100% | PASS |
| Team alignment documented | Required | See wave-decisions.md | PASS |
| Top 2-3 underserved needs identified | Required | Done | PASS |

**G2 STATUS: PASSED -- Proceed to Phase 3 (Solution Testing)**
