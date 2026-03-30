# JTBD Analysis -- Trivia Game

## Metadata

- Feature ID: trivia
- Phase: DISCUSS -- Phase 1 (JTBD)
- Date: 2026-03-29
- Method: Jobs-to-Be-Done (job stories + four forces + opportunity scoring)
- Personas: Marcus (quizmaster), Priya (team captain/player), Jordan (casual player)

---

## Persona Profiles

### Marcus Okafor -- The Quizmaster

**Role:** Host and game master
**Context:** Hosts trivia nights every 4-6 weeks at home or a local pub with 4-6 teams of 2-5 friends
**Technical fluency:** High -- writes YAML, comfortable with terminal, runs a home server
**Current workflow:** Writes quiz in a YAML-like structure in a text editor, reads questions aloud from laptop, collects paper answer sheets, manually tallies scores into a spreadsheet
**Core frustration:** "By the time I finish scoring a round, everyone's already moved on and the momentum is gone. I spend half the night with my head buried in paper."

---

### Priya Nair -- The Team Captain

**Role:** Elected spokesperson for her team of 3, responsible for final answer entry
**Context:** Attends Marcus's trivia nights regularly, uses her iPhone SE
**Technical fluency:** Medium -- comfortable with web apps, never writes code
**Current workflow:** Discusses answers verbally with team, writes on paper sheet, hands in at round end
**Core frustration:** "We crossed out an answer right before submitting and couldn't read what we'd written. The quizmaster couldn't either. We lost a point for a question we definitely knew."

---

### Jordan Kim -- The Casual Player

**Role:** Occasional participant, sometimes plays on a shared team device
**Context:** Joins when invited, may not know the usual routine
**Technical fluency:** Low-medium -- uses apps naturally but won't figure out complex UIs
**Current frustration:** "I refreshed the page by accident and everything was gone. I had to ask the host to re-explain how to rejoin."

---

## Job Stories

### Job Story JS-01: Starting the Game

**When** I arrive at the venue with my YAML quiz file ready and guests are settling in,
**I want to** load the quiz into the app and have everyone connected within 5 minutes,
**so I can** start the game while energy is high rather than losing momentum to technical setup.

**Functional job:** Initialize a game session from a YAML file
**Emotional job:** Feel competent and in control at the start
**Social job:** Appear prepared to guests; set the tone as a good host

---

### Job Story JS-02: Controlling the Reveal

**When** I'm ready for players to think about the next question,
**I want to** reveal one question at a time at my chosen pace,
**so I can** build suspense and keep the group synchronized rather than having faster readers jump ahead.

**Functional job:** Advance question state one at a time
**Emotional job:** Feel like a performer -- the room is in my hands
**Social job:** Maintain group engagement; no one is bored or confused

---

### Job Story JS-03: Sharing to the Room Display

**When** I want the whole room to see the current question on the TV,
**I want to** open a clean URL on the TV browser that shows only the current question,
**so I can** run the game with my laptop open to my private controls without anyone peeking at answers or upcoming rounds.

**Functional job:** Separate quizmaster view from public display
**Emotional job:** Feel professional; the game feels produced, not improvised
**Social job:** Give everyone in the room a shared focal point (the TV)

---

### Job Story JS-04: Scoring a Round Quickly

**When** all teams have submitted their answers and I want to announce scores,
**I want to** see each question with the expected answer beside each team's actual submission,
**so I can** make accept/reject decisions in under 3 minutes and keep the game moving.

**Functional job:** Compare expected vs. submitted answers; mark correct/incorrect
**Emotional job:** Feel efficient; not feel like I'm letting people down by taking too long
**Social job:** Keep the group entertained -- a long scoring break kills the vibe

---

### Job Story JS-05: Running the Answer Ceremony

**When** scoring is complete and I want to reveal how teams did,
**I want to** walk through each question's correct answer on the shared display one at a time,
**so I can** create a moment of suspense and celebration for each correct answer rather than just announcing a number.

**Functional job:** Display correct answers sequentially on the public screen
**Emotional job:** Feel like a showman; recreate the "ohhh!" moment of live trivia
**Social job:** Give all players the satisfaction of seeing how they did question-by-question

---

### Job Story JS-06: Joining as a Team

**When** I arrive at the game URL on my phone and see the join screen,
**I want to** enter my team name once and have the app remember me for the rest of the night,
**so I can** focus on the game rather than re-registering every time I switch tabs or accidentally refresh.

**Functional job:** Register team identity; persist across browser events
**Emotional job:** Feel immediately included; not feel confused or behind
**Social job:** Sit down and start playing, not troubleshoot

---

### Job Story JS-07: Entering and Editing Answers

**When** the quizmaster reveals questions during a round,
**I want to** type answers into a form I can edit freely until I submit,
**so I can** change my mind, incorporate team discussion, and submit the best version of our answers without penalty.

**Functional job:** Enter text answers; edit before final submission
**Emotional job:** Feel confident -- "our answers are saved, we can still fix things"
**Social job:** Discuss as a team without pressure; reduce disputes about what to write

---

### Job Story JS-08: Submitting Answers with Confidence

**When** the round ends and I'm ready to submit,
**I want to** see all our answers clearly before I confirm submission,
**so I can** catch any blanks or typos and submit knowing we gave our best answers.

**Functional job:** Review all answers; confirm submission
**Emotional job:** Feel certain -- "I've committed our best answers, no regrets"
**Social job:** Team agrees on what was submitted; no post-submission disputes

---

## Four Forces Analysis

The Four Forces model analyzes what pushes users toward a new behavior (push/pull) and what holds them back (inertia/anxiety).

### Forces Acting on Marcus (Quizmaster)

#### Push (Friction with Current Solution)
- Collecting paper sheets takes 5-10 minutes per round (sorting, naming teams, checking completeness)
- Manual scoring from paper is error-prone, especially with multi-team events
- Sharing laptop screen with everyone leaks answers and upcoming rounds
- Paper illegibility causes disputes that slow scoring and create awkwardness

#### Pull (Attraction to New Solution)
- Score an entire round in under 3 minutes with the scoring interface
- TV display shows exactly what the room should see -- no private info
- YAML file is already Marcus's mental model -- he writes quiz content this way now
- Auto-tallied running scores remove the spreadsheet entirely

#### Inertia (Attachment to Current Behavior)
- Paper sheets are familiar and require zero setup
- "Everyone already knows how this works" -- no player onboarding needed
- Works even without internet/devices (true offline fallback)

#### Anxiety (Fear About the New)
- What if the app crashes mid-game? (No backup paper system)
- What if older guests can't figure out the player app on their phones?
- What if the WiFi at the pub is unreliable?
- Is setting up the app more work than just running it the old way?

**Net force assessment:** Push and pull forces are strong (scoring bottleneck is chronic pain). Anxiety is real but mitigable: app must be simple enough to onboard all players in under 2 minutes, and the quizmaster needs visible confirmation that all teams are connected before starting.

---

### Forces Acting on Priya (Team Captain)

#### Push
- Crossed-out or illegible answers on paper cost points unfairly
- No record of what was submitted -- disputes happen
- Writing 10 answers per round with editing is messy on paper

#### Pull
- Edit any answer freely until submission -- no more crossing out
- App remembers the team across refreshes -- no re-entry
- Can see all revealed questions at once on the answer form

#### Inertia
- Writing on paper is familiar and social (team huddle around a sheet)
- Phone typing is slower than handwriting for some players

#### Anxiety
- Will the app eat our answers? (localStorage trust)
- What if I accidentally submit early?
- Is my phone compatible?

**Net force assessment:** Pull forces are strong (the "crossed out answer" pain is concrete and recurring). Submission confirmation dialog directly addresses the anxiety about accidental submission.

---

### Forces Acting on Jordan (Casual Player)

#### Push
- Losing session on refresh creates confusion and requires host intervention

#### Pull
- Rejoin and be recognized automatically -- zero friction re-entry
- Clean, focused interface -- just questions and answer fields

#### Inertia
- Familiar with paper; doesn't know what to expect from the app

#### Anxiety
- What do I do when I get to the URL? Will it be obvious?
- What if my team is already mid-round when I join?

**Net force assessment:** Onboarding must be near-instant and self-explanatory. The join flow must work even if the game has already started (late joiners).

---

## Opportunity Scoring (JTBD-Weighted)

Building on the DISCOVER phase opportunity scores, weighted by job story alignment:

| Opportunity | OPP ID | Imp | Sat | OST Score | Job Stories Served | JTBD Priority |
|-------------|--------|-----|-----|-----------|-------------------|---------------|
| Public display view | OPP-03c | 10 | 1 | 19 | JS-03, JS-05 | P1-Critical |
| Scoring interface | OPP-05a | 10 | 1 | 19 | JS-04 | P1-Critical |
| Fuzzy answer acceptance | OPP-05b | 9 | 1 | 17 | JS-04 | P1-Critical |
| Team answers editable | OPP-04a | 9 | 2 | 16 | JS-07 | P1-Critical |
| Persistent device recognition | OPP-02b | 9 | 2 | 16 | JS-06, JS-07 | P1-Critical |
| Round submission + confirm | OPP-04d | 9 | 3 | 15 | JS-08 | P1-High |
| Quizmaster question reveal | OPP-03a | 9 | 3 | 15 | JS-02 | P1-High |
| Player view all revealed Qs | OPP-03b | 8 | 3 | 13 | JS-07 | P1-High |
| Auto score tallying | OPP-05c | 8 | 2 | 14 | JS-04 | P1-High |
| Post-round answer ceremony | OPP-06a/b | 8 | 2 | 14 | JS-05 | P1-High |
| YAML quiz loading | OPP-01a | 9 | 5 | 13 | JS-01 | P1-High |
| Team creation on join | OPP-02a | 8 | 5 | 11 | JS-06 | P2 |
| Multimedia question support | OPP-03d | 7 | 2 | 12 | JS-02 | P2 |
| Multi-part answer support | OPP-04b | 7 | 2 | 12 | JS-07 | P2 |
| Multiple choice support | OPP-04c | 7 | 3 | 11 | JS-07 | P2 |
| Rejoin after disconnect | OPP-02c | 8 | 3 | 13 | JS-06 | P2 |
| YAML validation + preview | OPP-01b | 6 | 3 | 9 | JS-01 | P2 |
| Game session initialization | OPP-01c | 7 | 4 | 10 | JS-01 | P2 |
| Running scoreboard (QM) | OPP-05d | 7 | 3 | 11 | JS-04 | P2 |

---

## JTBD Gate Assessment

| Criterion | Status |
|-----------|--------|
| Job stories captured (all 3 personas) | PASS -- 8 job stories, 3 personas |
| Four forces documented for each persona | PASS -- quizmaster, team captain, casual player |
| Emotional and social jobs identified | PASS -- each JS has functional/emotional/social decomposition |
| Opportunity scores aligned to job stories | PASS -- all P1 opportunities trace to at least one JS |
| Anxiety forces mapped to design mitigations | PASS -- submission confirm, onboarding simplicity, pre-game readiness check |

**JTBD Gate: PASSED -- Proceed to Journey Design**
