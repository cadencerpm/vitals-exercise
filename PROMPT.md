# Live Coding Prompt

You have a small service that ingests blood pressure readings, stores them, and generates alerts.

## Part 1 — Bug fix
Run tests. At least one test fails. Fix the bug so tests pass.

## Part 2 — Extension
Extend the service to support re-taking vital signs for abnormal readings:

**Requirements:**
1. When an abnormal blood pressure reading is detected, create an alert and prompt the patient to re-take the vital
2. The system should prompt the patient at most once per alert
3. If the patient re-takes the vital on their own (before being prompted), the system should:
   - Skip sending the prompt
   - Mark the alert as confirmed (indicating the abnormality was verified, not a measurement error)
4. If any re-taken value (prompted or unprompted) is within the normal range, automatically resolve the alert

## Goals:
- correctness + tests
- readable, maintainable design
- discuss tradeoffs + product thinking
- show how you use AI tools effectively - use whichever coding agents you want!
