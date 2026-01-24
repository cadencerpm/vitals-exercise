# Live Coding Exercise

You have a small service that ingests blood pressure readings, stores them, and generates alerts.

Starting the server according to the README will also provide you with a GUI for interacting with the service.

**Important:** You may use whatever AI agents you want for *implementation*, but please walk us through your *reasoning and planning* before coding each step. Please do not use planning or thinking mode when leveraging AI. Our goal here is not to see if AI can one-shot a solution, but rather to see that you understand how the pieces interact, and can use AI effectively to *build* the pieces.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                     HTTP Layer                                       │
│                                                                                      │
│   POST /vitals ─────► vitals.py ─────► Service.ingest_vital()                        │
│   GET  /vitals ─────► vitals.py ─────► Service.list_vitals()                         │
│   GET  /alerts ─────► alerts.py ─────► Service.list_alerts()                         │
│                                                                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘
                                           │
                                           │ uses
                                           ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                      Service                                         │
│                                                                                      │
│   ingest_vital():                                                                    │
│     1. Validate input                                                                │
│     2. Store vital ──────────────────────────────► Store.add_vital()                 │
│     3. Publish event ────────────────────────────► PubSub.publish(VITAL_RECEIVED)    │
│                                                                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘
                                           │
                         ┌─────────────────┴─────────────────┐
                         │                                   │
                         ▼                                   ▼
┌─────────────────────────────────┐         ┌─────────────────────────────────┐
│             Store               │         │             PubSub              │
│         (InMemoryStore)         │         │                                 │
│                                 │         │  publish(event) ───────────┐    │
│  • add_vital(vital) -> Vital    │         │                            │    │
│  • add_alert(alert) -> Alert    │         │  subscribe()               │    |
│  • list_vitals() -> [Vital]     │         │                            │    │
│  • list_alerts() -> [Alert]     │         └────────────────────────────│────┘
│                                 │                                      │
└─────────────────────────────────┘                                      │
         ▲                                                               │
         │                                                               │
         │ reads/writes                                     subscribes to events
         │                                                               │
         │                                                               ▼
         │                                  ┌─────────────────────────────────┐
         │                                  │         AlertWorker             │
         │                                  │       (background thread)       │
         └──────────────────────────────────│                                 │
                                            │  Listens for VITAL_RECEIVED     │
                                            │  If vital is abnormal:          │
                                            │    • Creates Alert (ACTIVE)     │
                                            │    • Stores in Store            │
                                            │                                 │
                                            │  Has: store, message_queue      │
                                            │(hint: message_queue is unused!) |
                                            |   (your job is to fix that)     │
                                            └─────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────────────────────┐
│                              Message System (independent)                            │
│                                                                                      │
│  ┌─────────────────────────────────┐         ┌─────────────────────────────────┐     │
│  │         MessageQueue            │         │        MessageWorker            │     │
│  │                                 │ ◄────── │       (background thread)       │     │
│  │  • enqueue(patient_id, content) │ process │                                 │     │
│  │    -> Message (QUEUED)          │  next   │  Continuously calls             │     │
│  │                                 │         │  MessageQueue.process_next()    │     │
│  │  • process_next()               │         │                                 │     │
│  │    -> QUEUED -> PROCESSING      │         └─────────────────────────────────┘     │
│  │    -> (5-20 sec delay)          │                                                 │
│  │    -> SENT                      │                                                 │
│  │                                 │                                                 │
│  │  • list_messages() -> [Message] │                                                 │
│  └─────────────────────────────────┘                                                 │
│                                                                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                   Data Models                                        │
│                                                                                      │
│  Vital                          Alert                        Message                 │
│  ├─ id                          ├─ id                        ├─ id                   │
│  ├─ patient_id                  ├─ patient_id                ├─ patient_id           │
│  ├─ systolic                    ├─ vital_id                  ├─ content              │
│  ├─ diastolic                   ├─ systolic, diastolic       ├─ status (QUEUED/      │
│  ├─ taken_at                    ├─ taken_at, received_at     │          PROCESSING/  │
│  └─ received_at                 ├─ reason                    │          SENT)        │
│                                 ├─ status (see below)        ├─ queued_at            │
│                                 └─ created                   └─ sent_at              │
│                                                                                      │
│  AlertStatus:                                                                        │
│    ACTIVE             - Abnormal reading received, notification queued               │
│    AUTO_RESOLVED      - Patient submitted normal reading BEFORE notification sent    │
│    RESOLVED_BY_RETAKE - Patient was notified, retook vitals, reading was normal      │
│    CONFIRMED_ABNORMAL - Patient was notified, retook, reading still abnormal         │
│                                                                                      │
│  Abnormal = systolic > 180 OR diastolic > 120                                        │
│                                                                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Part 1: Bug Fix

A test is failing. Fix the bug so all tests pass.

### Steps:
1. Run tests, identify the failing test
2. Read the test to understand what behavior it expects
3. Explain your diagnosis: what's the bug and why does it happen?
4. Propose your fix (discuss tradeoffs if multiple approaches exist)
5. Implement and verify tests pass

**Tip:** Pay attention to what the test name tells you about expected behavior.

---

## Part 2: Feature

Patients with abnormal readings should be notified to retake their vitals.

### Steps:
1. Explore the codebase. What existing infrastructure could you use?
2. Explain your design before implementing
3. Implement basic notification on abnormal reading (it will appear in the GUI if successful)

### Follow-ups

4. **Avoid spamming:** What if 5 abnormal readings come in quickly?

5. **Handle resolution:** What happens when the patient retakes and the reading is normal? What if it's abnormal?
   - Hint: What state transitions does an alert need? Are we handling all of them?

**Tip:** Look at `AlertStatus` in `models.py` — the docstring helps clarify these states.

---

## Part 3: Extension

The notification system intentionally has a 5-20 second delay. What if a patient submits a normal reading *before* the notification is actually sent? Do we still want to send it?

### The Problem:
- Patient submits an abnormal reading → Alert created, notification queued
- Patient immediately retakes on their own → normal reading comes in
- But notification hasn't been sent yet (still in queue)
- Should we still send "please retake your vitals" when they already did?

### Steps:
1. Explain the race condition
2. Propose a solution (multiple solutions exist so be prepared to discuss tradeoffs)
3. Implement the fix

**Tip:** Think about the relationship between Alert and Message.

---

## Evaluation Criteria

- **Correctness:** Does it work? Are there tests? (If time permits)
- **Design:** Is the code readable and maintainable?
- **Systems thinking:** Do you understand how the async pieces interact?
- **Product thinking:** Do you consider edge cases and user experience?
- **AI fluency:** Do you use AI tools effectively while demonstrating understanding?
