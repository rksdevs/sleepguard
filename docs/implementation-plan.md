# SleepGuard Implementation Plan

Phase-by-phase development plan for portfolio delivery. Each phase ends with a **demoable milestone**. Pi-specific steps live in [checklist.md](checklist.md).

**Time budget:** ~2 hours/day (adjust as needed for portfolio urgency).

**Priority order:** motion alert → dashboard → persistence → camera → packaging.

---

## Phase 0 — Project bootstrap

**Goal:** Runnable Go module with config flags and a clear package layout.

| Task | Package / area | Deliverable |
|------|----------------|-------------|
| Initialize module | `go.mod` | `github.com/rksdevs/sleepguard` |
| Entry point | `cmd/sleepguard/main.go` | Startup message, flag parsing |
| Event model | `internal/sensor` or migrate `internals/helpers` | `Event` struct + JSON marshal |
| Config skeleton | `internal/config` | Struct for device name, cooldown, debug, HTTP addr |
| Structured logging | `main` | `log/slog` for startup and config dump |

**Done when:** `go run ./cmd/sleepguard -device nursery -alertInterval 30s` prints config and exits cleanly.

**Status:** Complete on Pi.

---

## Phase 1 — Sensor and GPIO (MVP core)

**Goal:** Read PIR motion on the Pi and emit debounced events to stdout/logs.

**Status:** Code complete — Pi wiring + verification pending.

| Task | Package / area | Deliverable |
|------|----------------|-------------|
| Sensor interface | `internal/sensor/sensor.go` | `Reader` interface + `Event` type |
| Mock reader | `internal/sensor/mock.go` | Dev machine testing without hardware |
| PIR reader | `internal/sensor/pir.go` | GPIO poll via `periph.io` |
| Debounce + cooldown | `internal/sensor/pir.go` | One event per motion burst |
| Wire in main | `cmd/sleepguard/main.go` | Loop or goroutine reading sensor |
| Logging | all | `motion_detected`, `cooldown_skipped`, `sensor_error` |

**Dependencies:** `periph.io/x/host/v3`, `periph.io/x/conn/v3/gpio`

**Done when:** On the Pi, waving in front of the PIR produces clear log lines without spam.

**Pi checklist section:** Phase 1

---

## Phase 2 — Local alerting and HTTP dashboard

**Goal:** LAN-accessible web UI and at least one real alert channel.

| Task | Package / area | Deliverable |
|------|----------------|-------------|
| HTTP server | `internal/web/server.go` | Listen on `:8080` (configurable) |
| Health endpoints | `internal/web/handlers.go` | `/health`, `/status` |
| Events API | `internal/web/handlers.go` | `GET /events` → JSON |
| In-memory store | `internal/store/memory.go` | Thread-safe ring buffer |
| Connect sensor → store | `main` | Events visible via API |
| HTML dashboard | `internal/web/templates/` | Table of recent events |
| Alert notifier | `internal/alert/notifier.go` | Local sound or exec command |
| Alert state machine | `internal/alert/alert.go` | `idle` → `motion` → `alert_sent` → `cooldown` |

**Done when:** Phone on same Wi‑Fi opens `http://<pi-ip>:8080`, sees events, and hears/sees an alert on motion.

**Pi checklist section:** Phase 2

---

## Phase 3 — Concurrency, persistence, and telemetry

**Goal:** Production-shaped architecture — goroutines, channels, disk persistence, metrics.

| Task | Package / area | Deliverable |
|------|----------------|-------------|
| Event channel | `main` | `chan sensor.Event` between workers |
| Goroutine split | `main` | Sensor, alert, store workers |
| Context shutdown | `main` | SIGINT/SIGTERM → graceful stop |
| JSONL persistence | `internal/store/jsonl.go` | Events survive restart |
| Load on startup | `internal/store` | Hydrate memory from file |
| Telemetry | `internal/telemetry/metrics.go` | Counters + last event time |
| Dashboard metrics | `internal/web` | Show counts on UI |
| Refactor + cleanup | all | Package names, remove dead code |

**Optional:** SQLite instead of JSONL if you want SQL in the portfolio story.

**Done when:** Restarting the app keeps history; dashboard shows motion/alert counts; Ctrl+C shuts down cleanly.

**Pi checklist section:** Phase 3

---

## Phase 4 — Camera, polish, and packaging

**Goal:** Snapshot on motion, tuned reliability, Docker, portfolio-ready docs.

| Task | Package / area | Deliverable |
|------|----------------|-------------|
| Camera capture | `internal/camera/capture.go` | `libcamera-still` / `raspistill` on motion |
| Link snapshot to event | `internal/store` | `SnapshotPath` on event |
| Latest image route | `internal/web` | `GET /snapshot/latest` |
| Dashboard image | templates | Show latest capture |
| Rate limits | `internal/alert`, `sensor` | Tune cooldowns, document defaults |
| Metrics endpoint | `internal/web` | `/metrics` or telemetry page |
| Dockerfile | repo root | Multi-stage or simple copy binary |
| README polish | `README.md` | Screenshots, demo GIF, architecture link |

**Done when:** Motion triggers snapshot; dashboard shows image; Docker builds; project is portfolio-ready.

**Pi checklist section:** Phase 4

---

## Phase 5 — Future (post-portfolio backlog)

Not required for MVP. Track as issues or a backlog section.

| Item | When to consider |
|------|------------------|
| Push notifications (Pushover, ntfy) | After local alert works |
| MQTT integration | Smart home users |
| Cloud sync | Multiple caregivers / locations |
| Prometheus exporter | Homelab / SRE story |
| Basic auth on dashboard | If exposed beyond trusted LAN |
| Kubernetes | Multiple services, not for single-binary Pi app |
| AI vision on snapshots | After camera pipeline is stable |

---

## Suggested timeline

| Phase | Focus | Approx. duration |
|-------|--------|----------------|
| 0 | Bootstrap | 1–2 days |
| 1 | GPIO + sensor | 3–5 days |
| 2 | Web + alerts | 5–7 days |
| 3 | Concurrency + storage | 5–7 days |
| 4 | Camera + Docker | 5–7 days |

**Total MVP:** ~3–4 weeks at 2 h/day. Compress phases 2–3 if portfolio deadline is tight (e.g. skip HTML polish until phase 4).

---

## Phase dependencies

```text
Phase 0 ──► Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4
              │            │            │
              │            └────────────┴── can parallelize camera late
              └── requires Pi wiring (checklist)
```

---

## Definition of done (full MVP)

- [ ] PIR motion detected on Pi with debouncing
- [ ] Alert fires on motion (audible or visible)
- [ ] LAN dashboard lists events and status
- [ ] Events persist across restart
- [ ] Goroutine + channel architecture with graceful shutdown
- [ ] Telemetry visible on dashboard
- [ ] Optional: snapshot on motion
- [ ] Dockerfile builds and runs
- [ ] README + architecture docs complete

---

## Mapping from original learning plan

The [motion-sensor-go.md](motion-sensor-go.md) day-by-day plan maps roughly as:

| Original weeks | This plan |
|----------------|-----------|
| Week 1 (days 1–7) | Phase 0 + Phase 1 |
| Week 2 (days 8–14) | Phase 2 |
| Week 3 (days 15–21) | Phase 3 |
| Week 4 (days 22–28) | Phase 4 |

Use the original doc for Go concept references; use this doc for delivery milestones.

---

## How we work with the checklist

1. Implement code for the current phase on dev machine (or directly on Pi).
2. Update [checklist.md](checklist.md) with new Pi steps for that phase.
3. You complete Pi steps and mark them done.
4. Verify "done when" criteria, then move to the next phase.
