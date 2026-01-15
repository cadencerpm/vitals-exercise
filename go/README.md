# Cadence Vitals Interview

Minimal Go app for ingesting blood pressure vitals and creating alerts.
This is a purposefully simplified toy implementation for interview use.

## Repo map

- `cmd/server`: gRPC server entrypoint and wiring.
- `cmd/cli`: small CLI for inserting vitals and listing alerts.
- `internal/api`: gRPC handlers + proto mappings.
- `internal/app`: domain logic (service, store, pub/sub, alert worker, models).
- `proto/vitals/v1`: protobuf definitions and generated code.

## Flow

High-level vital lifecycle:
- A vital arrives via gRPC (the CLI is just a thin client).
- The service validates and stores the vital with a server-side received timestamp.
- An event is published and the background worker evaluates thresholds.
- If abnormal, an alert is created and stored; `ListAlerts` reads from that store.

Note: Everything is in-memory, so restarting the server clears vitals/alerts.

## Running the App

**1. Start the server:**
```bash
make server
```

**2. In another terminal, use the CLI commands:**
```bash
# Insert a normal vital (120/80)
go run ./cmd/cli insert-vital --patient patient-1 --systolic 120 --diastolic 80

# Insert an abnormal vital (190/130) - this will trigger an alert
go run ./cmd/cli insert-vital --patient patient-1 --systolic 190 --diastolic 130

# List all vitals for a patient
go run ./cmd/cli list-vitals --patient patient-1

# List all alerts for a patient
go run ./cmd/cli list-alerts --patient patient-1
```

## Testing

```bash
make test
```

## Development

If you need to modify the protobuf definitions in `proto/vitals/v1/*.proto`, you'll need to regenerate the Go code.

**1. Install Protocol Buffers compiler (protoc):**

macOS:
```bash
brew install protobuf
```

Linux:
```bash
# Ubuntu/Debian
sudo apt install protobuf-compiler

# Fedora
sudo dnf install protobuf-compiler
```

**2. Install Go protobuf tools:**
```bash
make setup
```

**3. Regenerate proto files after making changes:**
```bash
make proto
```
