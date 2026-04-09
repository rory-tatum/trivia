# Outcome KPIs — Host UI (Quizmaster Panel)

**Feature ID**: host-ui
**Date**: 2026-04-09

---

## Summary

The host UI feature succeeds when Marcus can run a complete trivia game session without confusion, frustration, or incorrect connection signals. These KPIs measure observable behavior changes, not vanity metrics.

---

## KPI-01: Connection Status Accuracy

- **Who**: Quizmasters opening the host page (any token, any network condition)
- **Does what**: Trust the connection status indicator — acts on it (load quiz, start round) only when it reads "Connected"
- **By how much**: Zero instances of status reading "Connected" while the WebSocket handshake has not completed or has failed
- **Measured by**: Automated test verifying `connected` state is only set inside `onOpen` callback; QA session with network throttling
- **Baseline**: Currently, `setConnected(true)` is called synchronously — "Connected" shows on every page load before handshake, including when auth fails
- **Story**: US-01

---

## KPI-02: Auth Failure Clarity

- **Who**: Quizmasters who supply a wrong token
- **Does what**: Understand immediately why the page is not working and know the corrective action
- **By how much**: 100% of auth failures result in a clear error message within 2 seconds of page load; 0 silent retry loops on auth failure
- **Measured by**: QA session with wrong token — count seconds to error display; inspect WS frames for absence of retry attempts
- **Baseline**: Currently, wrong token triggers silent retry loop; user sees "Connected" then eventually nothing
- **Story**: US-01

---

## KPI-03: Quiz Load Success Rate

- **Who**: Quizmasters attempting to load a quiz after connecting
- **Does what**: Load their quiz file successfully on the first attempt (path is correct)
- **By how much**: Load success rate >= 95% of attempts where the file path is valid; failed attempts with invalid path show inline error within 1 second
- **Measured by**: QA session: load quiz with valid path, observe confirmation; load with invalid path, observe error
- **Baseline**: Currently impossible — no load quiz form exists
- **Story**: US-02

---

## KPI-04: Round Control Accuracy

- **Who**: Quizmasters running an active round
- **Does what**: Reveal questions in the correct sequence without skipping or double-revealing
- **By how much**: Zero out-of-order reveals across a complete game session; question counter matches actual reveals at all times
- **Measured by**: QA session: run a full round tracking question_index values in WS frames; verify counter display matches
- **Baseline**: Currently impossible — no round controls exist
- **Story**: US-03

---

## KPI-05: Scoring Panel Completeness

- **Who**: Quizmasters during the scoring phase
- **Does what**: Mark every team's answer for every question before publishing scores
- **By how much**: Zero published rounds where any team-answer pair has not been verdicted (or Marcus explicitly skips)
- **Measured by**: QA session: verify all team-answer rows show a verdict indicator before "Publish Scores" is clicked
- **Baseline**: Currently impossible — no scoring interface exists
- **Story**: US-04

---

## KPI-06: Ceremony Display Accuracy

- **Who**: Quizmasters running the answer ceremony
- **Does what**: Show questions and reveal answers on the display screen in correct sequence; answers visible to display room only
- **By how much**: Display screen shows correct content at each ceremony step; play screen never shows an answer during ceremony
- **Measured by**: QA session with display and play screens open simultaneously; verify content per step
- **Baseline**: Currently impossible — no ceremony controls exist
- **Story**: US-05

---

## KPI-07: Game Session Completion Rate

- **Who**: Quizmasters who connect and load a quiz
- **Does what**: Complete a full game session (all rounds scored + game ended)
- **By how much**: >= 90% of sessions that start a round reach the "game_over" state without host-side errors
- **Measured by**: Session observation in QA; track game_over event received
- **Baseline**: Currently 0% — no controls exist to advance through any game phase
- **Story**: US-06

---

## Feature-Level Outcome Statement

> Marcus opens the host page, connects within 3 seconds, loads a quiz, runs all rounds (start → reveal → score → publish), walks through the ceremony, and ends the game — all without consulting documentation, without seeing misleading connection status, and without a single silent failure.

**Measurement moment**: A live game session with 2+ teams, 2+ rounds, all phases completed.
**Target**: Feature is considered done when a complete session runs without Marcus encountering any of the root causes (A, B, C, D) identified in the RCA.
