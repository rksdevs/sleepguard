# SleepGuard

A **baby sleep motion monitor** — Pi edge agent + Hetzner cloud API + installable PWA. PIR on the nursery Pi streams motion events to your server; family checks the live log, gets push alerts on sustained motion, and captures camera stills on demand from anywhere.

**Live:** [https://sleepguard.rksdevs.in](https://sleepguard.rksdevs.in)

## Why this project

SleepGuard is designed as a portfolio piece that demonstrates:

- **Embedded / IoT**: PIR + Pi Camera on Raspberry Pi 4
- **Edge + cloud**: Pi agent uploads events; Hetzner hosts API, Postgres, and PWA
- **Backend engineering**: HTTP API, device auth, rules engine, Web Push, Docker
- **Systems thinking**: Server-side pattern rules, manual capture UX, retention

## Current status (June 2026)

| Phase | Scope | Status |
|-------|--------|--------|
| 0–1 | Pi PIR local verify (GPIO17) | Done |
| **A** | Cloud API + Postgres on Hetzner | Done |
| **B** | PWA live log | Done |
| **C** | Pi agent → cloud (systemd) | Done |
| **D** | Web Push pairing + 24h cleanup | Done |
| **E** | Server rules: push every 3 motion cycles | Done |
| **F** | Manual camera capture + snapshot carousel | Done |
| **G** | Portfolio polish, legacy cleanup | **Next (RTC)** |

See [docs/implementation-plan.md](docs/implementation-plan.md) for the full roadmap and [deploy/README.md](deploy/README.md) for ops.

## Architecture (production)

```text
Pi (nursery)                    Hetzner (wowlogs)
┌─────────────────┐            ┌──────────────────────────┐
│ PIR → agent     │──HTTPS───►│ cloud API (:8090)        │
│ rpicam-still    │           │ Postgres (sleepguard DB) │
│ systemd agent   │◄─poll 5s──│ snapshots on disk        │
└─────────────────┘            │ nginx + PWA dist       │
                               └───────────┬──────────────┘
                                           │ HTTPS
                               ┌───────────▼──────────────┐
                               │ Family phones (PWA)      │
                               │ log · push · capture     │
                               └──────────────────────────┘
```

**Pi sends:** raw `rise` / `fall` / `hold` events + heartbeats + JPEG uploads.  
**Server does:** cycle counting, Web Push (3, 6, 9… cycles), capture queue, retention cleanup.  
**User triggers:** camera still via PWA **Capture image** (not automatic on motion).

## Repository layout

```text
cmd/agent/          # Pi edge agent (PIR + camera + cloud upload)
cmd/cloud/          # Hetzner API server
cmd/sleepguard/     # Legacy local Pi app (dev / bench only)
internal/domain/    # Shared event + pairing types
internal/sensor/    # PIR reader
internal/camera/    # rpicam-still / libcamera-still capture
internal/agent/     # Upload client, offline queue, config
internal/cloud/     # API, auth, store, rules, push, cleanup, commands
web/pwa/            # Vite + TypeScript PWA
deploy/             # Docker, env examples, Hetzner + Pi guides
scripts/            # gen-vapid-keys.go
docs/               # Architecture, plan, checklist, electronics
```

## Quick start — deploy

| Target | Guide |
|--------|--------|
| **Hetzner** (cloud + PWA) | [deploy/README.md](deploy/README.md) |
| **Pi** (agent + camera) | [deploy/README.md](deploy/README.md) § Phase C / F |
| **Local dev** (mock sensor) | `go run ./cmd/agent -mock-sensor -mock-camera` |

## Hardware

- Raspberry Pi 4 (4GB), hostname `sleepguard`
- PIR HC-SR501 on **GPIO17 (pin 11)** — not GPIO2
- Pi Camera Module v2 on CSI (`imx219`, use `rpicam-still` on Bookworm+)

Wiring: [docs/electronics.md](docs/electronics.md). Pi checklist: [docs/checklist.md](docs/checklist.md).

## Documentation

| Document | Purpose |
|----------|---------|
| [docs/implementation-plan.md](docs/implementation-plan.md) | Phases A–H, definition of done, **RTC** |
| [docs/architecture.md](docs/architecture.md) | HLD/LLD (legacy local + production addendum) |
| [docs/checklist.md](docs/checklist.md) | Pi + Hetzner verification checklist |
| [deploy/README.md](deploy/README.md) | Deploy, API reference, smoke tests |
| [docs/electronics.md](docs/electronics.md) | Parts and wiring |

## Tech stack

- **Go 1.25+**, **Postgres**, **Docker** (cloud), **Vite/TS** (PWA)
- **GPIO:** periph.io · **Camera:** rpicam-apps · **Push:** Web Push (VAPID)
- **Infra:** Hetzner, nginx, Cloudflare, Pi systemd

## License

TBD.

## Resume / portfolio summary

> Built a full-stack baby motion monitor: Raspberry Pi edge agent (PIR + camera) streams to a Go cloud API on Hetzner with Postgres, server-side motion-cycle rules, Web Push alerts, and an installable PWA for live logs and on-demand nursery snapshots — deployed end-to-end from India to EU with Docker, nginx, and systemd.
