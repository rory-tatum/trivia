# Journey Gherkin -- Trivia Game
# Feature ID: trivia
# Phase: DISCUSS -- Phase 2 (Journey Design)
# Date: 2026-03-29
# Covers: all three interfaces (/host, /play, /display) -- full journey scenarios

Feature: Trivia Night -- Complete Game Journey

  Background:
    Given Marcus has prepared a YAML quiz file "friday-march-2026.yaml" containing 4 rounds
    And the YAML file includes text questions, one image question, one audio question, and one multiple choice question
    And Priya leads "Team Awesome" with 3 players
    And Jordan is a member of "The Brainiacs" with 2 players
    And "Quiz Killers" has 4 players

  # =========================================================================
  # HOST JOURNEY -- /host interface
  # =========================================================================

  # H1 -- Load Quiz
  # -------------------------------------------------------------------------

  Scenario: Quizmaster loads a valid YAML quiz file
    Given Marcus opens /host in his browser
    When Marcus provides the path to "friday-march-2026.yaml"
    Then the system validates the YAML structure and all referenced media files
    And Marcus sees "Friday Night Trivia -- March 2026 | 4 rounds | 32 questions"
    And the lobby screen shows two shareable URLs: one for /play and one for /display
    And no teams are yet connected

  Scenario: Quizmaster loads a YAML file with a missing required field
    Given Marcus opens /host and provides "quiz-with-error.yaml"
    When the system validates the file
    Then Marcus sees the error: "Round 2, Question 3: missing required field 'answer'"
    And no game session is created
    And Marcus can correct the file and reload without losing the page context

  Scenario: Quizmaster loads a YAML file referencing a missing media file
    Given the YAML file references "round1-image.jpg" in Round 1, Question 4
    And "round1-image.jpg" is not present in the expected directory
    When Marcus loads the YAML
    Then Marcus sees the error: "Round 1, Question 4: image file 'round1-image.jpg' not found"
    And the game session is not created until the error is resolved

  # H2 -- Lobby
  # -------------------------------------------------------------------------

  Scenario: Teams join and quizmaster sees them in the lobby
    Given Marcus has loaded "friday-march-2026.yaml" successfully
    And Marcus has shared the /play URL with his guests
    When Team Awesome, The Brainiacs, and Quiz Killers each connect to /play and register their team names
    Then Marcus sees all three teams listed in the lobby with their connection timestamps
    And the "Start Game" button is available

  Scenario: Quizmaster starts game with all teams present
    Given three teams are connected in the lobby
    When Marcus clicks "Start Game"
    Then all /play clients transition to the round active state for Round 1
    And the /display transitions from the holding screen to the first question view
    And Marcus sees the reveal panel for Round 1

  # H3 -- Question Reveal
  # -------------------------------------------------------------------------

  Scenario: Quizmaster reveals questions one at a time
    Given Marcus is on the Round 1 reveal panel
    And no questions have been revealed yet
    When Marcus clicks "Reveal Q1"
    Then "What is the capital of France?" appears on all /play clients
    And "What is the capital of France?" appears on /display as the current question
    And the question counter on /host shows "1 of 8 revealed"

  Scenario: Quizmaster reveals a second question while players answer the first
    Given Marcus has revealed Q1 ("What is the capital of France?")
    And Priya has entered "Paris" in the Q1 answer field
    When Marcus clicks "Reveal Q2"
    Then Q2 ("Name the three primary colors.") appears in Priya's answer form below Q1
    And Priya's Q1 answer "Paris" is preserved
    And /display now shows Q2 as the current question

  Scenario: Quizmaster reveals an image question
    Given Marcus is on the reveal panel for Round 1
    When Marcus clicks "Reveal Q3" (which has an associated image "eiffel.jpg")
    Then /play shows the image "eiffel.jpg" above the question text "Name this landmark."
    And /display shows the image prominently with the question text below
    And the image is served from the same directory as the YAML file

  Scenario: Quizmaster reveals an audio question
    Given Marcus is on the reveal panel for Round 2
    When Marcus clicks "Reveal Q1" (which has associated audio "mystery-track.mp3")
    Then /display shows "Now playing..." and begins playing the audio file
    And /play shows "Name this song and artist." with a multi-part answer field

  Scenario: Quizmaster reveals a multiple choice question
    Given Marcus clicks "Reveal Q4" (which has choices: Venus, Mercury, Mars, Earth)
    Then /play shows four radio buttons labeled A) Venus, B) Mercury, C) Mars, D) Earth
    And /display shows the question with all four choices labeled A through D
    And players can select exactly one choice

  # H4 -- Submission Monitoring
  # -------------------------------------------------------------------------

  Scenario: Quizmaster monitors team submission status after round ends
    Given all 8 questions in Round 1 have been revealed
    When Marcus clicks "End Round"
    Then /host shows the submission status panel
    And Team Awesome and The Brainiacs are listed as "submitted"
    And Quiz Killers is listed as "waiting"
    And the "Open Scoring" button is greyed out with a note "1 team has not yet submitted"

  Scenario: Quizmaster overrides and opens scoring before all teams submit
    Given Quiz Killers has not submitted after 5 minutes
    When Marcus clicks "Open Scoring Anyway"
    Then the scoring interface opens with answers from Team Awesome and The Brainiacs
    And Quiz Killers shows blank submissions for all questions

  # H5 -- Scoring Interface
  # -------------------------------------------------------------------------

  Scenario: Quizmaster marks a correct text answer
    Given Marcus is on the scoring interface for Round 1, Q1 ("capital of France?"; answer: "Paris")
    And Team Awesome submitted "paris"
    When Marcus clicks "correct" for Team Awesome's answer "paris"
    Then Team Awesome's score increments by 1 point
    And the answer is marked green in the scoring grid
    And the running total for Team Awesome updates automatically

  Scenario: Quizmaster marks an incorrect answer
    Given Quiz Killers submitted "Lyon" for Q1 ("capital of France?"; answer: "Paris")
    When Marcus clicks "wrong" for Quiz Killers' answer "Lyon"
    Then Quiz Killers' score does not increment
    And the answer is marked red in the scoring grid

  Scenario: Quizmaster scores a multi-part unordered answer
    Given Q2 requires answers "Red", "Blue", "Yellow" in any order
    And The Brainiacs submitted "Yellow", "Red", "Blue"
    When Marcus clicks "correct" for The Brainiacs' Q2 answer
    Then The Brainiacs receive 1 point for Q2
    And the scoring interface confirms the answer as correct

  Scenario: Quizmaster scores a multi-part ordered answer
    Given Q7 requires answers in the specific order "1st", "2nd", "3rd"
    And Team Awesome submitted the parts in the wrong order
    When Marcus reviews Team Awesome's ordered answer
    Then the scoring interface highlights that the order does not match
    And Marcus can choose to mark it correct or wrong based on judgment

  # H6 -- Answer Ceremony
  # -------------------------------------------------------------------------

  Scenario: Quizmaster starts the answer ceremony
    Given all answers for Round 1 have been marked correct or incorrect
    When Marcus clicks "Save Scores & Start Ceremony"
    Then /display transitions to the ceremony mode showing Q1 question text
    And /play clients show a ceremony view
    And /host shows the ceremony control panel with Q1 selected

  Scenario: Quizmaster advances through ceremony answers
    Given the ceremony is in progress at Q2
    When Marcus clicks "Next Answer"
    Then /display shows Q3 question text with the correct answer revealed below it
    And the ceremony progress indicator shows "3 of 8 answers revealed"
    And Marcus can see on /host which teams got Q3 correct

  Scenario: Quizmaster ends ceremony and displays round scores
    Given Marcus has stepped through all 8 answers in Round 1
    When Marcus clicks "End Ceremony & Show Scores"
    Then /display shows the Round 1 scores in rank order with running totals
    And /host shows an option to start Round 2

  # =========================================================================
  # PLAYER JOURNEY -- /play interface
  # =========================================================================

  # P1 -- Join
  # -------------------------------------------------------------------------

  Scenario: Player joins game for the first time
    Given Priya opens /play on her iPhone
    When Priya enters "Team Awesome" in the team name field and clicks "Join Game"
    Then a team identity token is written to Priya's browser localStorage
    And Priya sees the lobby screen showing "Team Awesome -- 1 player connected"
    And "Other teams: The Brainiacs, Quiz Killers" are listed

  Scenario: Player returns after browser refresh during the game
    Given Priya has previously joined as "Team Awesome"
    And Priya accidentally refreshes the page during Round 1 while she has entered 3 answers
    When the page reloads
    Then the app reads Priya's localStorage token
    And automatically rejoins "Team Awesome" to the current game session
    And Priya sees "Welcome back, Team Awesome!" and her 3 entered answers are restored

  Scenario: Team name is already taken
    Given "Team Awesome" has already been registered by another device
    When Jordan tries to register "Team Awesome" from a different device
    Then Jordan sees the error "That name is taken -- try a different team name"
    And Jordan can enter a new name without reloading the page

  Scenario: Late joiner opens /play after game has started
    Given Round 1 is in progress and Q1 and Q2 have been revealed
    When a new player opens /play and registers as "Last Minute Squad"
    Then the player sees Q1 and Q2 already revealed in their answer form
    And the player can begin entering answers immediately
    And their late registration is visible to Marcus on /host

  # P3 -- Answer Entry
  # -------------------------------------------------------------------------

  Scenario: Player enters and edits a text answer before submission
    Given Q1 "What is the capital of France?" has been revealed
    And Priya has entered "Paris" in the Q1 answer field
    When Priya changes her answer to "paris" (lowercase)
    Then the answer field shows "paris"
    And the change is preserved in localStorage
    And the change is synced to the server as a draft

  Scenario: Player answers a multiple choice question
    Given Q4 has choices A) Venus, B) Mercury, C) Mars, D) Earth
    When Priya taps "B) Mercury"
    Then Mercury is selected and highlighted
    And the other options are deselected
    And Priya can change her selection at any time before submission

  Scenario: Player fills a multi-part answer
    Given Q2 asks for three primary colors and shows three answer fields
    When Priya enters "Red" in field 1, "Blue" in field 2, "Yellow" in field 3
    Then all three values are saved as Priya's Q2 answer
    And the answer is displayed as "Red / Blue / Yellow" on the review screen

  # P4 -- Submit
  # -------------------------------------------------------------------------

  Scenario: Player reviews answers before submitting with a blank question
    Given all 8 questions in Round 1 have been revealed
    And Priya has answered Q1-Q4 and Q6-Q8 but left Q5 blank
    When the round ends and Priya opens the submit review screen
    Then Priya sees a review list showing all 8 answers
    And Q5 is flagged with "!" and the note "no answer entered"
    And a "Go Back & Edit" button is prominent alongside "Submit Answers"

  Scenario: Player confirms submission via the confirmation dialog
    Given Priya is on the submit review screen with all answers entered
    When Priya clicks "Submit Answers"
    Then a confirmation dialog appears: "Once submitted, you cannot change your answers."
    And Priya clicks "Yes, Submit"
    Then the submission is locked and confirmed on the server
    And Priya sees "Your answers are locked in." with the team submission status panel

  Scenario: Player cancels submission after seeing the confirmation dialog
    Given the confirmation dialog is open
    When Priya clicks "Go Back"
    Then the dialog closes and Priya returns to the answer form
    And all previously entered answers are intact

  Scenario: Player cannot edit answers after submitting
    Given Team Awesome has submitted their Round 1 answers
    When Priya tries to tap on the Q1 answer field
    Then the field is read-only and non-interactive
    And a banner reads "Answers are locked for Round 1"

  # P5 -- Ceremony View
  # -------------------------------------------------------------------------

  Scenario: Player sees answer ceremony on their device
    Given Marcus has started the Round 1 answer ceremony
    When Marcus advances to Q3
    Then Priya's /play view shows the ceremony screen for Q3
    And the correct answer "Eiffel Tower" is visible after Marcus advances to it
    And Priya can see her team's submitted answer for Q3 alongside the correct answer

  # =========================================================================
  # DISPLAY JOURNEY -- /display interface
  # =========================================================================

  # D1 -- Holding Screen
  # -------------------------------------------------------------------------

  Scenario: Display shows holding screen before game starts
    Given Marcus has loaded the quiz and the lobby is open
    When the /display URL is opened on the room TV
    Then the TV shows "Friday Night Trivia -- March 2026"
    And the join URL "http://trivia.local/play" is visible
    And no question content or quizmaster controls are visible

  # D2 -- Question View
  # -------------------------------------------------------------------------

  Scenario: Display updates immediately when quizmaster reveals a question
    Given the game is in Round 1 and Q2 was previously shown
    When Marcus clicks "Reveal Q3" (with image "eiffel.jpg")
    Then the /display TV updates within 1 second to show the image and "Name this landmark."
    And the round name "Round 1: General Knowledge" and "Question 3 of 8" are shown in the header
    And no answer, no upcoming questions, and no quizmaster controls are visible

  Scenario: Display shows no private information
    Given Marcus's /host screen shows the correct answer "Eiffel Tower" for Q3
    When /display is showing Q3
    Then "Eiffel Tower" is NOT visible on /display
    And upcoming question content for Q4-Q8 is NOT visible on /display

  # D3 -- Waiting for Submissions
  # -------------------------------------------------------------------------

  Scenario: Display shows submission waiting state with anonymous team count
    Given Round 1 has ended and teams are submitting answers
    When 2 of 3 teams have submitted
    Then /display shows "2 of 3 teams submitted"
    And no team names are shown in the submission count
    And the display does not show any answer content

  # D4 -- Answer Ceremony
  # -------------------------------------------------------------------------

  Scenario: Display shows correct answer during ceremony
    Given the ceremony is in progress for Round 1
    And Marcus has advanced to Q1
    When Marcus clicks "Next Answer" to reveal Q1's answer
    Then /display shows Q1 question text: "What is the capital of France?"
    And below it: "Answer: Paris"
    And which teams got it right is NOT shown on /display (quizmaster announces this verbally)

  # D5 -- Round Scores
  # -------------------------------------------------------------------------

  Scenario: Display shows round scores in rank order after ceremony
    Given the Round 1 ceremony is complete
    When Marcus clicks "End Ceremony & Show Scores"
    Then /display shows:
      | Rank | Team | Round Points | Running Total |
      | 1 | Team Awesome | 8 | 8 |
      | 2 | The Brainiacs | 6 | 6 |
      | 3 | Quiz Killers | 4 | 4 |
    And the display shows "Round 2 starts shortly..."

  # =========================================================================
  # RESILIENCE & EDGE CASES
  # =========================================================================

  Scenario: Player reconnects after connection loss mid-round
    Given Priya is in Round 2 with 4 answers entered
    And Priya loses internet connection for 30 seconds
    When Priya's connection is restored
    Then the /play client reconnects automatically
    And Priya's 4 entered answers are restored (from localStorage)
    And the current revealed question set is re-synced from the server
    And Priya sees a brief "Reconnecting..." banner that dismisses on success

  Scenario: Display reconnects after connection loss
    Given /display is showing Q3 during Round 1
    And the TV's browser loses its WebSocket connection
    When the connection is restored
    Then /display reconnects and shows the current game state (whatever Marcus has since revealed)
    And the TV shows "Reconnecting..." overlay briefly, then clears

  Scenario: Quizmaster accidentally closes /host tab and reopens it
    Given the game is in progress during Round 2 scoring
    When Marcus reopens /host in a new tab
    Then Marcus is prompted to re-authenticate (or use the session token from the original tab)
    And the game state is fully restored to the scoring interface

  Scenario: Game runs to completion with final scores
    Given all 4 rounds have been played, scored, and ceremonialized
    When Marcus clicks "End Game" on the /host final screen
    Then /display shows "Trivia Night Complete" with the final standings in rank order
    And Team Awesome is highlighted as the winner
    And /play clients show the final scores screen
    And all answer entry fields are permanently locked
