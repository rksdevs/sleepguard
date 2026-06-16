# SleepGuard — Raspberry Pi 4 Checklist

Action items to run on the **Pi 4** after code is ready for each phase.  
**You** complete hardware and Pi-side steps; **we** update this file when new code needs Pi verification.

**Legend:** `[ ]` todo · `[x]` done · `[-]` skipped / N/A

---

## One-time Pi setup

| # | Task | Status | Notes |
|---|------|--------|-------|
| 0.1 | Pi 4 imaged with Raspberry Pi OS (64-bit recommended) | [ ] | |
| 0.2 | Pi connected to home Wi‑Fi | [ ] | Ethernet N/A — Wi‑Fi confirmed on hand |
| 0.3 | SSH enabled (`ssh pi@<pi-ip>` works) | [ ] | |
| 0.4 | Hostname or static IP noted (e.g. `192.168.1.x`) | [ ] | |
| 0.5 | Go 1.24+ installed on Pi (`go version`) | [ ] | Or cross-compile from dev PC |
| 0.6 | Git clone repo on Pi (or `scp` binary) | [ ] | `git clone ...` |
| 0.7 | PIR sensor wired per [electronics.md](electronics.md) | [ ] | Do not power GPIO until wiring checked |
| 0.8 | Pi Camera Module v2 connected (ribbon seated) | [ ] | Phase 4 — hardware ready, enable in raspi-config later |

---

## Phase 0 — Project bootstrap

*Code: flags, module, basic main. No GPIO yet.*

| # | Task | Status | Notes |
|---|------|--------|-------|
| 1.1 | `go run ./cmd/sleepguard` on Pi succeeds | [ ] | May work same as laptop |
| 1.2 | `go run ./cmd/sleepguard -device nursery -alert-cooldown 45s -debug=false` | [ ] | Verify flags parse |
| 1.3 | Note any Pi-specific Go module issues | [ ] | e.g. proxy, disk space |

**Phase 0 gate:** App runs on Pi with expected flag output.

---

## Phase 1 — Sensor and GPIO

*Code: PIR reader, debounce, structured logs.*

| # | Task | Status | Notes |
|---|------|--------|-------|
| 2.1 | Install periph dependencies build on Pi (`go build ./...`) | [ ] | First GPIO code pull |
| 2.2 | Confirm PIR wiring: VCC, GND, OUT → GPIO (see electronics) | [ ] | |
| 2.3 | Run app with real PIR reader (not mock) | [ ] | |
| 2.4 | Wave hand — log shows `motion_detected` (or equivalent) | [ ] | |
| 2.5 | Hold still — no repeated spam within cooldown window | [ ] | |
| 2.6 | Unplug OUT wire — app logs sensor error, does not crash | [ ] | |
| 2.7 | Document actual GPIO pin used in Notes below | [ ] | Default plan: GPIO17 |

**Phase 1 gate:** Reliable motion logs on Pi without event spam.

### Your notes (Phase 1)

```
Pi IP:
GPIO pin used:
Cooldown tested (seconds):
Issues:
```

---

## Phase 2 — Web dashboard and alerts

*Code: HTTP server, `/health`, `/events`, dashboard HTML, notifier.*

| # | Task | Status | Notes |
|---|------|--------|-------|
| 3.1 | Build and run app; confirm it listens on `:8080` | [ ] | `ss -tlnp` or `curl localhost:8080/health` |
| 3.2 | From phone (same LAN): open `http://<pi-ip>:8080/health` | [ ] | |
| 3.3 | Trigger motion — `/events` returns JSON with new event | [ ] | |
| 3.4 | Dashboard `/` shows readable event list | [ ] | |
| 3.5 | Alert fires (sound/exec) on motion | [ ] | Describe method used |
| 3.6 | Firewall: port 8080 allowed if ufw enabled | [ ] | `sudo ufw allow 8080` if needed |

**Phase 2 gate:** Phone dashboard + working alert on motion.

### Your notes (Phase 2)

```
Alert method (aplay / exec / other):
Phone browser tested:
Screenshot saved for portfolio (y/n):
```

---

## Phase 3 — Concurrency, persistence, telemetry

*Code: channels, goroutines, JSONL/SQLite, graceful shutdown, metrics on UI.*

| # | Task | Status | Notes |
|---|------|--------|-------|
| 4.1 | Pull latest; `go build -o sleepguard ./cmd/sleepguard` | [ ] | |
| 4.2 | Run app; trigger several motion events | [ ] | |
| 4.3 | Ctrl+C — app exits cleanly (no hang) | [ ] | |
| 4.4 | Restart app — previous events still on dashboard | [ ] | |
| 4.5 | Dashboard shows motion count / uptime / last event | [ ] | |
| 4.6 | Check data file path on disk (JSONL or DB) | [ ] | Note path below |
| 4.7 | Optional: run as `systemd` service for auto-start | [ ] | Post-MVP polish |

**Phase 3 gate:** Persistent history + clean shutdown + telemetry visible.

### Your notes (Phase 3)

```
Store file path:
Event count after restart:
systemd unit created (y/n):
```

---

## Phase 4 — Camera, polish, Docker

*Code: snapshot on motion, `/snapshot/latest`, Dockerfile, tuning.*

| # | Task | Status | Notes |
|---|------|--------|-------|
| 5.1 | Enable camera in `raspi-config` if using Pi Camera Module | [ ] | Interface → Camera |
| 5.2 | Test `libcamera-still` or `raspistill` manually | [ ] | One JPEG saved |
| 5.3 | Run app; motion triggers snapshot file on disk | [ ] | |
| 5.4 | Dashboard or `/snapshot/latest` shows image on phone | [ ] | |
| 5.5 | Tune cooldown — system feels calm, not noisy | [ ] | Record final values |
| 5.6 | Build Docker image on Pi or dev machine | [ ] | GPIO may need `--privileged` / device mount |
| 5.7 | Capture portfolio screenshots / short demo video | [ ] | |

**Phase 4 gate:** End-to-end demo ready for portfolio.

### Your notes (Phase 4)

```
Camera type (Pi module / USB):
Snapshot directory:
Final alertInterval:
Docker run command used:
```

---

## Troubleshooting quick reference

| Symptom | Things to check |
|---------|-----------------|
| `permission denied` on GPIO | Run as user in `gpio` group, or check periph docs |
| No motion events | PIR warm-up ~60 s; wrong pin; 3.3 V vs 5 V logic |
| Too many events | Increase cooldown; check debounce in code |
| Cannot reach dashboard | Pi IP, firewall, same subnet, correct port |
| Camera fails | `raspi-config` camera on; cable seated; test CLI first |
| `go build` fails on Pi | `go mod tidy`; enough RAM; swap if needed |

---

## Changelog (updated by development sessions)

| Date | Phase | What was added / changed |
|------|-------|--------------------------|
| 2026-06-16 | — | Initial checklist created. Phase 0 code exists (flags only). |
| 2026-06-16 | — | Kit confirmed: Pi 4 4GB, HC-SR501, Camera v2, breadboard/M-F wires, 300Ω+LEDs, Wi‑Fi. |
| 2026-06-16 | 0 | Phase 0 coded: `internal/config`, `internal/sensor`, slog startup. Pull repo on Pi and run checklist Phase 0. |

---

*When you finish a task, change `[ ]` to `[x]` and add notes. Tell the agent when a phase gate is passed so we start the next phase.*
