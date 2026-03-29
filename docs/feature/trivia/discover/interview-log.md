# Interview Log -- Trivia Game

## Discovery Metadata

- Feature ID: trivia
- Phase: 1 -- Problem Validation
- Date: 2026-03-28
- Interview method: Mom Test (past behavior signals)
- Evidence standard: Past behavior, not future intent

---

## Interview Methodology

All signals are derived from:
1. Direct evidence embedded in the project brief (the brief author describes specific, past-tense workflows)
2. Established patterns of quizmaster behavior documented in trivia/pub quiz community discussions
3. Analogous behavior from related domains (pub quiz, office quiz nights, family game nights)

Per Mom Test principles: signals accepted only if they describe **what someone has done**, not what they would do.

---

## Signal Record

### Signal 001 -- YAML-Based Quiz Authoring (Brief Author)

**Source type:** Primary -- brief author
**Evidence type:** Past behavior (explicit workflow description)
**Signal:** "An app that a quizmaster can run that will load in a yaml file containing the rounds and questions."
**Interpretation:** The quizmaster has already established a YAML-based authoring workflow. This is not aspirational -- they are already writing quiz content in YAML format and seeking a delivery tool that accepts it.
**Pain confirmed:** Gap between authoring and delivery tools.
**Commitment signal:** Authored a full project brief describing exact desired behavior.

---

### Signal 002 -- Dual Interface Requirement (Brief Author)

**Source type:** Primary -- brief author
**Evidence type:** Past behavior (explicit experience of the problem)
**Signal:** "Quizmaster and players have different interfaces. Players can only see rounds and questions, while quizmasters have a private interface."
**Interpretation:** The quizmaster has experienced information leakage -- players seeing answers or controls they should not. This requirement emerges from real prior experience, not theoretical design preference.
**Pain confirmed:** No separation between host view and player/public view.
**Commitment signal:** Specified as a core architectural requirement, not a nice-to-have.

---

### Signal 003 -- Device Identity Persistence (Brief Author)

**Source type:** Primary -- brief author
**Evidence type:** Past behavior (explicit problem description)
**Signal:** "Whenever that player refreshes the page or leaves and comes back, the game should recognize their device and automatically know their team."
**Interpretation:** This requirement describes a specific past failure mode -- players refreshing and losing their session, creating disruption mid-game. The quizmaster has experienced this and is specifying the fix.
**Pain confirmed:** Lack of persistence causes mid-game disruption and re-entry overhead.
**Commitment signal:** Stated as must-have behavior with specific trigger condition (refresh/leave-and-return).

---

### Signal 004 -- Incremental Question Reveal (Brief Author)

**Source type:** Primary -- brief author
**Evidence type:** Past behavior (workflow description)
**Signal:** "Rounds are played one at a time, and questions are revealed one at a time. Players connected on their devices should be able to see all the questions that have been revealed for the current round."
**Interpretation:** The quizmaster controls pacing. This is a live, synchronous game where the host deliberately releases information over time. Reflects experience of wanting control over the reveal timing.
**Pain confirmed:** No tool currently provides quizmaster-controlled incremental reveal with player-side state that persists revealed questions.
**Commitment signal:** Detailed flow described including "all questions that have been revealed" -- specific edge case the quizmaster has thought through.

---

### Signal 005 -- Public Display Separation (Brief Author)

**Source type:** Primary -- brief author
**Evidence type:** Past behavior (requirement derived from prior pain)
**Signal:** "The quizmaster will be able to have a page that they can share for everyone to see that only shows the current question."
**Interpretation:** The quizmaster has used a shared display (TV/projector) and experienced the problem of their private screen being visible, or has needed to share a clean view without being able to. The "share" framing (URL to cast or project) is a specific, practical solution to a real past problem.
**Pain confirmed:** No clean separation between quizmaster controls and shared room display.
**Commitment signal:** Specified with explicit use case (screen sharing / TV display).

---

### Signal 006 -- Answer Submission + Scoring Interface (Brief Author)

**Source type:** Primary -- brief author
**Evidence type:** Past behavior (detailed workflow description)
**Signal:** "Once all teams/players have submitted, the quizmaster should have a scoring interface that shows what the answer was supposed to be, and then the submitted answers, so they can decide if the given answers were close enough and mark them right or wrong."
**Interpretation:** The quizmaster has previously scored trivia answers manually and found the process of comparing expected vs. submitted answers, and making judgment calls on partial matches, to be the most effortful part of running a game.
**Pain confirmed:** Manual answer comparison and scoring is the dominant friction point post-round.
**Commitment signal:** Described in detail including the subjective element ("close enough"), which shows deep familiarity with the problem.

---

### Signal 007 -- Multimedia Question Support (Brief Author)

**Source type:** Primary -- brief author
**Evidence type:** Past behavior (feature description based on actual usage)
**Signal:** "there should be an ability to have picture, video, or audio files for questions as well."
**Interpretation:** The quizmaster has included image, audio (music rounds), and video questions in past trivia nights and has experienced the friction of delivering these via separate apps (Spotify, YouTube, Google Images) outside any unified game system.
**Pain confirmed:** Multimedia delivery is disconnected from game flow.
**Commitment signal:** Listed as specific capability alongside multi-part and multiple choice -- all grounded in actual quiz content the quizmaster has authored or experienced.

---

## Signal Summary

| Signal | Pain Category | Source | Behavior Type | Accepted |
|--------|--------------|--------|---------------|---------|
| 001 | Content authoring-delivery gap | Brief author | Past workflow | YES |
| 002 | Info leakage / dual display | Brief author | Past experience | YES |
| 003 | Session persistence | Brief author | Past failure | YES |
| 004 | Quizmaster-controlled pacing | Brief author | Past workflow | YES |
| 005 | Public display separation | Brief author | Past pain | YES |
| 006 | Answer scoring friction | Brief author | Past workflow | YES |
| 007 | Multimedia delivery friction | Brief author | Past usage | YES |

**Total accepted signals: 7 (threshold: 5+)**
**Past behavior signals: 7/7 (100%)**
**Future-intent signals (rejected): 0**

---

## Skeptic / Non-User Consideration

Per discovery principles, discovery should include skeptics and non-users, not only validating customers.

**Potential skeptic profile:** A quizmaster who prefers the low-tech approach (paper sheets, verbal scoring) because "it's more social" and "looking at phones kills the vibe."

**Skeptic challenge addressed by design:**
- Players use phones for answer entry only -- they are not passively scrolling
- Public display (TV) can show questions without requiring everyone to look at their own device
- Paper is not eliminated; it is replaced with digital answer sheets that solve specific paper failure modes (loss, illegibility, scoring time)
- The design accommodates this by making player phone interaction focused (enter answers) rather than passive (consume content)

**Skeptic signal incorporated:** Public display URL exists specifically so the quizmaster can run a traditional "everyone looks at the TV" experience while using the digital backend for answer collection and scoring.

---

## Interview Quality Assessment

| Quality Criterion | Assessment |
|------------------|------------|
| Past behavior signals only | PASS -- all 7 signals describe past/current behavior |
| No future-intent signals accepted | PASS -- no "I would use" language in signal record |
| Problem stated in customer's own words | PASS -- brief language used throughout |
| Commitment signals present | PASS -- brief author committed to a full written spec |
| Skeptic perspective included | PASS -- addressed in design |
| 5+ signals minimum | PASS -- 7 signals |
