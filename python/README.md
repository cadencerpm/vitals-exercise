# Cadence Vitals Interview (Python)

Minimal Python app for ingesting blood pressure vitals and creating alerts.

## Quick Start

```bash
python -m venv .venv
source .venv/bin/activate  # or: source .venv/bin/activate.fish
pip install -e .
python -m app.server
```

A dashboard opens at http://127.0.0.1:5000 where you can insert vitals and see alerts.

## How It Works

1. A vital is submitted (via dashboard or API)
2. The service validates and stores it
3. A `VITAL_RECEIVED` event is published
4. The alert worker checks if it's abnormal (systolic > 180 or diastolic > 120)
5. If abnormal, an alert is created and logged to the console

Everything is in-memory—restarting the server clears all data.

## Repo Structure

```
app/
├── server.py          # Entrypoint, wires everything together
├── service.py         # Business logic for ingesting vitals
├── store.py           # In-memory storage
├── pubsub.py          # Simple pub/sub for events
├── alert_worker.py    # Background worker that creates alerts
├── message_queue.py   # Message queue with simulated delays
├── message_worker.py  # Background worker that processes messages
├── models.py          # Data models (Vital, Alert, Message)
└── api/
    ├── vitals.py      # POST /vitals, GET /vitals
    ├── alerts.py      # GET /alerts
    └── dashboard.py   # Web UI at /
```

## CLI

You can also use the CLI instead of the dashboard:

```bash
# Insert vitals
python -m app.cli insert-vital --patient patient-1 --systolic 120 --diastolic 80
python -m app.cli insert-vital --patient patient-1 --systolic 190 --diastolic 130

# List data
python -m app.cli list-vitals --patient patient-1
python -m app.cli list-alerts --patient patient-1
```

## Testing

```bash
pip install -e ".[dev]"
python -m pytest
```
