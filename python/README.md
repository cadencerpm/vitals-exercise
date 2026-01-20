# Cadence Vitals Interview (Python)

Minimal Python app for ingesting blood pressure vitals and creating alerts.
This is a purposefully simplified toy implementation for interview use.

## Repo map

- `app/`: domain logic (service, store, pub/sub, alert worker, models)
- `app/server.py`: Flask server entrypoint and wiring
- `app/cli.py`: small CLI for inserting vitals and listing alerts
- `app/api/`: Flask routes and JSON mappings
- `tests/`: pytest tests for domain logic

## Flow

High-level vital lifecycle:
- A vital arrives via HTTP (the CLI is just a thin client).
- The service validates and stores the vital with a server-side received timestamp.
- An event is published and the background worker evaluates thresholds.
- If abnormal, an alert is created and stored; `GET /alerts` reads from that store.

Note: Everything is in-memory, so restarting the server clears vitals/alerts.

## Running the App

From the `python/` directory:

**1. Install dependencies:**
```bash
python -m venv .venv
source .venv/bin/activate
python -m pip install -e .
```

**2. Start the server:**
```bash
python -m app.server
```

**3. In another terminal, use the CLI commands:**
```bash
# Insert a normal vital (120/80)
python -m app.cli insert-vital --patient patient-1 --systolic 120 --diastolic 80

# Insert an abnormal vital (190/130) - this will trigger an alert
python -m app.cli insert-vital --patient patient-1 --systolic 190 --diastolic 130

# List all vitals for a patient
python -m app.cli list-vitals --patient patient-1

# List all alerts for a patient
python -m app.cli list-alerts --patient patient-1
```

## Testing

```bash
python -m pip install -e ".[dev]"
python -m pytest
```
