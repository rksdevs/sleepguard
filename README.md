# SleepGuard

A **baby sleep motion monitor** — Pi edge agent + Hetzner cloud API + PWA. Detects motion via PIR, streams events to your server, and serves a family dashboard from anywhere.

## Why this project

SleepGuard is designed as a portfolio piece that demonstrates:

- **Embedded / IoT**: GPIO sensor integration on Raspberry Pi
- **Edge + cloud**: Pi agent uploads events; Hetzner hosts API and PWA
- **Backend engineering**: HTTP API, Postgres, Docker, structured logging
- **Systems thinking**: Event-driven architecture, device auth, retention

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
cmd/sleepguard/                 # Legacy local Pi app (dev / bench)
cmd/cloud/                      # Hetzner cloud API
internal/domain/                # Shared event types
internal/sensor/                # PIR reader (Pi)
internal/cloud/                 # API, auth, Postgres store, migrations
internal/agent/                 # Pi uploader (phase C)
web/pwa/                        # PWA (phase B)
deploy/                         # Docker, env example, Hetzner guide
docs/                           # Architecture, plan, checklist, electronics
```

## Current status

| Phase | Status |
|-------|--------|
| Pi PIR (local) | Complete on GPIO17 |
| **A — Cloud API + Postgres** | Complete on Hetzner |
| **B — PWA live log** | Code complete — build + nginx on server |
| C — Pi agent → cloud | Not started |
| D+ — Pairing, rules, camera | Planned |

See [docs/implementation-plan.md](docs/implementation-plan.md) and [deploy/README.md](deploy/README.md).

## Quick start — cloud (Phase A)

```bash
# Local dev (requires Postgres database `sleepguard`)
export DATABASE_URL='postgres://sleepguard:pass@localhost:5432/sleepguard?sslmode=disable'
go run ./cmd/cloud -debug
```

Hetzner: follow [deploy/README.md](deploy/README.md).

## Quick start — local Pi (dev)

With flags (mock sensor on dev machine — no GPIO):

```bash
go run ./cmd/sleepguard -mock-sensor -device nursery -report-interval 5s -debug
```

## Quick start (Raspberry Pi 4)

1. Install Go 1.24+ on the Pi (or cross-compile from your dev machine).
2. Wire the PIR sensor per [docs/electronics.md](docs/electronics.md) — **GPIO17 (pin 11)**.
3. Follow [docs/checklist.md](docs/checklist.md) for each phase.
4. Run the app and open the dashboard from a phone or laptop on the same LAN:

```bash
cd ~/sleepguard && git pull
go run ./cmd/sleepguard -device nursery -report-interval 5s -debug
# Dashboard: http://<pi-ip>:8080
# Health:     curl localhost:8080/health
```

Optional audible alert via shell command:

```bash
go run ./cmd/sleepguard -device nursery -alert-cmd "aplay /path/to/beep.wav"
```

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
