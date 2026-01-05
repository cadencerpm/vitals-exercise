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
- Everything is in-memory, so restarting the server clears vitals/alerts.

## Setup

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

**2. Install Go tools and generate proto files:**
```bash
make setup
make proto
```

## Run

```bash
make server
make cli insert-vital --patient patient-1 --systolic 120 --diastolic 80
make cli list-alerts
```

## Test

```bash
make test
```
