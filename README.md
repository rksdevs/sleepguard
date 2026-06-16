# SleepGuard

A **local-first baby sleep motion monitor** built in Go on a Raspberry Pi 4. SleepGuard detects motion via a PIR sensor, triggers alerts, stores event history, and exposes a LAN dashboard with optional camera snapshots — no cloud dependency required for the MVP.

## Why this project

SleepGuard is designed as a portfolio piece that demonstrates:

- **Embedded / IoT**: GPIO sensor integration on Raspberry Pi
- **Backend engineering**: HTTP API, structured logging, concurrency, graceful shutdown
- **Systems thinking**: Event-driven architecture, state machines, local persistence
- **Pragmatic scope**: Local-first MVP that can grow into cloud sync, metrics, and vision later

## Scope

### In scope (MVP)

| Area | Description |
|------|-------------|
| Motion detection | PIR sensor on Pi GPIO with debouncing and cooldown |
| Alerting | Local alert channel (sound, browser notification, or LAN notification) |
| Event history | In-memory store with JSONL or SQLite persistence |
| Web dashboard | `/health`, `/status`, `/events`, local HTML UI on LAN |
| Telemetry | Motion count, alert count, uptime, last event time |
| Camera (phase 4) | Snapshot on motion, latest image on dashboard |
| Packaging | Docker image for repeatable deployment |

### Out of scope (for now)

- Cloud sync and multi-user accounts
- Kubernetes / multi-service orchestration
- Live video streaming
- AI / computer vision
- Mobile native apps

These are documented as future phases in [docs/implementation-plan.md](docs/implementation-plan.md).

## Hardware

- Raspberry Pi 4 (primary runtime)
- PIR motion sensor (HC-SR501 or equivalent)
- Optional: Pi Camera Module or USB webcam (phase 4)
- Jumper wires, breadboard or soldered connections, 5V power supply

Full parts list and wiring notes: [docs/electronics.md](docs/electronics.md).

## Repository layout

```text
cmd/sleepguard/main.go          # Entry point
internal/config/                # Flags, env, config loading
internal/sensor/                # Sensor interface + GPIO implementation
internal/alert/                 # Alert channels and state machine
internal/store/                 # In-memory and persistent event storage
internal/web/                   # HTTP server, routes, HTML templates
internal/camera/                # Snapshot capture (phase 4)
internal/telemetry/             # Counters and metrics
internals/helpers/              # Legacy / shared types (migrate to internal/)
docs/                           # Architecture, plan, checklist, electronics
```

> **Note:** New packages follow the `internal/` convention from the plan. Existing `internals/helpers/` will be migrated as the project grows.

## Current status

| Phase | Status |
|-------|--------|
| Phase 0 — Project bootstrap | Code complete — Pi verify pending |
| Phase 1 — Sensor + GPIO | Not started |
| Phase 2 — Web dashboard + alerts | Not started |
| Phase 3 — Concurrency + storage | Not started |
| Phase 4 — Camera + polish | Not started |

See [docs/checklist.md](docs/checklist.md) for Pi-specific tasks and [docs/implementation-plan.md](docs/implementation-plan.md) for the full roadmap.

## Quick start (development machine)

```bash
git clone https://github.com/rksdevs/sleepguard.git
cd sleepguard
go run ./cmd/sleepguard
```

With flags:

```bash
go run ./cmd/sleepguard -device nursery -alert-cooldown 30s -debug
```

## Quick start (Raspberry Pi 4)

1. Install Go 1.24+ on the Pi (or cross-compile from your dev machine).
2. Wire the PIR sensor per [docs/electronics.md](docs/electronics.md).
3. Follow [docs/checklist.md](docs/checklist.md) for each phase.
4. Run the binary and open `http://<pi-ip>:8080` from a phone or laptop on the same LAN.

Detailed Pi setup steps are added to the checklist as each phase is implemented.

## Documentation

| Document | Purpose |
|----------|---------|
| [docs/architecture.md](docs/architecture.md) | High-level and low-level design |
| [docs/implementation-plan.md](docs/implementation-plan.md) | Phase-by-phase build plan |
| [docs/checklist.md](docs/checklist.md) | Pi 4 action items (updated per phase) |
| [docs/electronics.md](docs/electronics.md) | Parts list and wiring reference |
| [docs/motion-sensor-go.md](docs/motion-sensor-go.md) | Original learning-oriented day-by-day plan |

## Tech stack

- **Language:** Go 1.24+
- **Hardware:** Raspberry Pi 4, PIR sensor, optional camera
- **GPIO:** [periph.io](https://periph.io/)
- **HTTP:** `net/http`, `html/template`
- **Storage:** JSONL or SQLite (phase 3)
- **Logging:** `log/slog`
- **Packaging:** Docker (phase 4)

## License

TBD.

## Resume / portfolio summary

> Built a local-first baby sleep monitoring system in Go on Raspberry Pi. Detects motion via GPIO, triggers alerts through an event-driven pipeline with goroutines and channels, persists event history, and serves a LAN dashboard with telemetry and optional camera snapshots. Designed for extension to cloud sync and vision features.
