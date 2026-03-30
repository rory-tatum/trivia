# Outcome KPIs -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DISCUSS -- Phase 4 (Requirements)
- Date: 2026-03-29

---

## KPI Framework

Each KPI captures: who changes their behavior, what they do differently, how much, and how we know.

---

### KPI-01: Round Scoring Time

- **Who:** Marcus (quizmaster)
- **Does what:** Scores a complete round (5 teams x 10 questions) by marking each answer correct or incorrect
- **By how much:** Completes scoring in under 3 minutes
- **Measured by:** End-to-end timing from "Open Scoring" click to "Start Ceremony" click in a controlled test game session
- **Baseline:** Current paper-based scoring: 10-15 minutes per round (counting, re-checking, re-writing)
- **Target:** Under 3 minutes (5x improvement)

---

### KPI-02: Player Answer Loss Rate on Refresh

- **Who:** Priya (team captain), Jordan (casual player)
- **Does what:** Accidentally refreshes the browser during a round
- **By how much:** 0% of refreshes result in lost answers
- **Measured by:** Automated test -- write N answers, refresh, verify N answers restored
- **Baseline:** Current paper: 0% loss (paper doesn't refresh), but 100% loss in current web apps with no persistence
- **Target:** 0% answer loss on refresh (localStorage + server draft)

---

### KPI-03: Player Onboarding Time (Join to First Answer)

- **Who:** Jordan (casual player, first time using the app)
- **Does what:** Opens the /play URL and enters their first answer
- **By how much:** Under 60 seconds from URL open to first answer entered
- **Measured by:** Stopwatch timing with a new user who has never seen the app
- **Baseline:** Paper: ~2-3 minutes (explain the process verbally, hand out sheets, write team name)
- **Target:** Under 60 seconds (2-3x improvement)

---

### KPI-04: Quizmaster Setup Time

- **Who:** Marcus (quizmaster)
- **Does what:** Loads YAML file, shares URLs, confirms teams are connected, starts game
- **By how much:** Under 2 minutes from opening /host to clicking "Start Game"
- **Measured by:** End-to-end timing in a real or simulated game session
- **Baseline:** Current: 5-10 minutes (explaining paper rules, distributing sheets, counting teams)
- **Target:** Under 2 minutes

---

### KPI-05: Answer Submission Completeness

- **Who:** All teams playing
- **Does what:** Submit their round answers before scoring begins
- **By how much:** 100% of teams submit before scoring opens (in normal play, no override needed)
- **Measured by:** Submission status panel -- count of rounds where all teams submitted before quizmaster opened scoring
- **Baseline:** Paper: variable -- teams sometimes forget to hand in; quizmaster must chase
- **Target:** 100% of teams submit on their own (override never needed)

---

### KPI-06: Quizmaster Satisfaction (Net Outcome)

- **Who:** Marcus (quizmaster)
- **Does what:** States they would use the app again for their next trivia night
- **By how much:** 100% (binary for personal use -- either he uses it or he doesn't)
- **Measured by:** Post-game verbal feedback / informal check ("Was this better than paper?")
- **Baseline:** Current: mixed (YAML authoring is good; delivery is painful)
- **Target:** "Yes, this is better than paper. I'm using it for every trivia night from now on."

---

### KPI-07: Display Information Security (Zero Leaks)

- **Who:** Marcus (quizmaster)
- **Does what:** Opens /display on the room TV and plays the game normally
- **By how much:** 0 instances of answer or upcoming question content appearing on /display
- **Measured by:** Manual verification test: play full game session; check /display never shows answer fields or questions not yet revealed
- **Baseline:** Current shared-screen approach: information leakage happens nearly every game (wrong window, wrong tab)
- **Target:** Zero leaks in any tested game session

---

## KPI Summary Table

| KPI | Who | Behavior Change | Target | Measurement |
|-----|-----|-----------------|--------|-------------|
| KPI-01 | Quizmaster | Scores round | < 3 min | Timed test |
| KPI-02 | Players | Refresh without loss | 0% loss | Automated test |
| KPI-03 | New player | Join to first answer | < 60 sec | Timed test |
| KPI-04 | Quizmaster | Setup time | < 2 min | Timed test |
| KPI-05 | All teams | Submit before override | 100% | Session log |
| KPI-06 | Quizmaster | Reuse intent | 100% | Post-game check |
| KPI-07 | Quizmaster | Zero display leaks | 0 leaks | Manual verification |
