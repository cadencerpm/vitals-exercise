# Cadence Vitals Interview (Node/Fastify)

Minimal Node.js app for ingesting blood pressure vitals and creating alerts.
This is a purposefully simplified toy implementation for interview use.

## Repo map

- `src/server.ts`: Fastify server entrypoint and wiring.
- `src/app.ts`: server builder and background worker lifecycle.
- `src/api/`: Fastify routes and JSON serializers.
- `src/domain/`: domain logic (service, store, pub/sub, alert worker, models).
- `test/`: node:test unit tests for domain logic.

## Flow

High-level vital lifecycle:
- A vital arrives via HTTP JSON.
- The service validates and stores the vital with a server-side received timestamp.
- An event is published and the background worker evaluates thresholds.
- If abnormal, an alert is created and stored; `GET /alerts` reads from that store.

Note: Everything is in-memory, so restarting the server clears vitals/alerts.

## Running the App

From the `node/` directory:

**1. Install dependencies:**
```bash
npm install
```

**2. Start the server:**
```bash
npm run dev
```

**3. Example requests:**
```bash
# Insert a normal vital (120/80)
curl -s -X POST http://127.0.0.1:3000/vitals \
  -H 'content-type: application/json' \
  -d '{"patientId":"patient-1","systolic":120,"diastolic":80,"takenAt":1700000000}'

# Insert an abnormal vital (190/130) - this will trigger an alert
curl -s -X POST http://127.0.0.1:3000/vitals \
  -H 'content-type: application/json' \
  -d '{"patientId":"patient-1","systolic":190,"diastolic":130,"takenAt":1700000000}'

# List vitals for a patient
curl -s "http://127.0.0.1:3000/vitals?patientId=patient-1"

# List alerts for a patient
curl -s "http://127.0.0.1:3000/alerts?patientId=patient-1"
```

## Testing

```bash
npm test
```
