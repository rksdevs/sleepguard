# SleepGuard — Raspberry Pi 4 Checklist

Action items to run on the **Pi 4** after code is ready for each phase.  
**You** complete hardware and Pi-side steps; **we** update this file when new code needs Pi verification.

**Legend:** `[ ]` todo · `[x]` done · `[-]` skipped / N/A

---

## One-time Pi setup

| # | Task | Status | Notes |
|---|------|--------|-------|
| 0.1 | Pi 4 imaged with Raspberry Pi OS (64-bit recommended) | [x] | Re-flashed with English keyboard |
| 0.2 | Pi connected to home Wi‑Fi | [x] | Manual Wi‑Fi connect; IP may change per network |
| 0.3 | SSH enabled (`ssh pi@<pi-ip>` works) | [x] | `ssh rksdevs@192.168.0.103` |
| 0.4 | Hostname or static IP noted (e.g. `192.168.1.x`) | [x] | `192.168.0.103` (current network) |
| 0.5 | Go 1.24+ installed on Pi (`go version`) | [x] | |
| 0.6 | Git clone repo on Pi (or `scp` binary) | [x] | `git clone` OK |
| 0.7 | PIR sensor wired per [electronics.md](electronics.md) | [x] | GPIO17 (pin 11); see Phase 1 notes |
| 0.8 | Pi Camera Module v2 connected (ribbon seated) | [ ] | Phase 4 — hardware ready, enable in raspi-config later |

---

## Phase 0 — Project bootstrap

*Code: flags, module, basic main. No GPIO yet.*

| # | Task | Status | Notes |
|---|------|--------|-------|
| 1.1 | `go run ./cmd/sleepguard` on Pi succeeds | [x] | |
| 1.2 | `go run ./cmd/sleepguard -device nursery -alert-cooldown 45s -debug=false` | [x] | Phase 0 output verified |
| 1.3 | Note any Pi-specific Go module issues | [x] | None |

**Phase 0 gate:** App runs on Pi with expected flag output.

---

## Phase 1 — Sensor and GPIO

*Code: PIR reader, debounce, structured logs.*

| # | Task | Status | Notes |
|---|------|--------|-------|
| 2.1 | Pull latest + `go build ./...` on Pi | [x] | `git pull` then build — periph deps OK |
| 2.2 | Add user to gpio group: `sudo usermod -aG gpio rksdevs` then re-login | [x] | Required for GPIO access |
| 2.3 | Confirm PIR wiring: VCC→5V, GND→GND, OUT→GPIO17 (pin 11) | [x] | Orange→pin 4, Gray→pin 6, White→pin 11 |
| 2.4 | Run: `go run ./cmd/sleepguard -device nursery -report-interval 5s -debug` | [x] | Runs until Ctrl+C — no `-mock-sensor` on Pi |
| 2.5 | Wait ~60 s (PIR warm-up), wave hand — log shows `motion_detected` | [x] | rise/fall/hold patterns logged |
| 2.6 | Hold still — no repeated spam within cooldown window | [x] | `report-interval` controls hold logs only |
| 2.7 | Unplug OUT wire — app logs `sensor_error`, does not crash | [ ] | Optional test |
| 2.8 | Document GPIO pin used in Notes below | [x] | GPIO17 (BCM), physical pin 11 |

**Phase 1 gate:** Reliable motion logs on Pi without event spam. **Passed.**

### Your notes (Phase 1)

```
Pi IP: 192.168.0.103
GPIO pin used: GPIO17 (BCM) / physical pin 11 — do NOT use GPIO2 (pin 3); bad pull-up on this board
Cooldown tested (seconds): 30s default; alert cooldown separate in Phase 2
PIR module: jumper on H (repeat trigger); Time Delay knob fully CCW or sensor stuck HIGH
Issues resolved: breadboard center gap; wrong pin counting; GPIO2 always HIGH
```

---

---

## Phase A — Cloud API (Hetzner)

*Code: `cmd/cloud`, Postgres, Docker. See [deploy/README.md](../deploy/README.md).*

| # | Task | Status | Notes |
|---|------|--------|-------|
| A.1 | Create Postgres database `sleepguard` (separate from wow-logs) | [ ] | |
| A.2 | Clone repo on Hetzner, `cp deploy/env.example deploy/.env` | [ ] | Set DATABASE_URL, keys |
| A.3 | `docker compose up -d --build` in `deploy/` | [ ] | |
| A.4 | `curl localhost:8090/health` returns ok | [ ] | |
| A.5 | POST test event with device token | [ ] | See deploy README |
| A.6 | GET events with read API key — newest first | [ ] | |

**Phase A gate:** Event ingested on Hetzner and readable via API.

### Your notes (Phase A)

```
Hetzner path:
Domain (if any):
Device token stored (y/n):
```

---

## Phase B — PWA (Hetzner)

*Code: `web/pwa`, nginx static + API proxy. See deploy README.*

| # | Task | Status | Notes |
|---|------|--------|-------|
| B.1 | `git pull` + `bash deploy/build-pwa.sh` | [ ] | Needs Node 18+ on server |
| B.2 | Update nginx: `root` → `web/pwa/dist`, proxy `/api/` | [ ] | Keep certbot SSL block |
| B.3 | POST test event via localhost | [ ] | |
| B.4 | Open `https://sleepguard.rksdevs.in` in browser | [ ] | Pass Cloudflare challenge once |
| B.5 | Enter read API key, device `nursery` — see events | [ ] | |

**Phase B gate:** PWA shows live log from cloud API.

---

## Phase C — Pi agent → cloud

| # | Task | Status | Notes |
|---|------|--------|-------|
| C.1 | Cloudflare WAF skip for `SleepGuard-Agent` on `/api/v1/*` | [ ] | Required with orange proxy |
| C.2 | `git pull` on Pi, `go build -o bin/sleepguard-agent ./cmd/agent` | [ ] | |
| C.3 | `deploy/agent.env` with device token | [ ] | `SLEEPGUARD_BOOTSTRAP_DEVICE_TOKEN` |
| C.4 | Test: `go run ./cmd/agent -debug` | [ ] | PIR warm-up ~60s |
| C.5 | PWA **ONLINE** + live events at sleepguard.rksdevs.in | [ ] | |
| C.6 | `sudo systemctl enable --now sleepguard-agent` | [ ] | |

**Phase C gate:** Real PIR motion appears in PWA without manual curl.

---

## Phase 2 — Local web dashboard (legacy)

| # | Task | Status | Notes |
|---|------|--------|-------|
| 3.1 | Run: `go run ./cmd/sleepguard -device nursery -report-interval 5s -debug` | [ ] | HTTP + sensor run together; listens on `:8080` |
| 3.2 | Confirm listen: `curl -s localhost:8080/health` | [ ] | Expect `{"status":"ok"}` |
| 3.3 | From phone (same LAN): open `http://192.168.0.103:8080/` | [ ] | Dashboard auto-refreshes every 5 s |
| 3.4 | Trigger motion — `curl -s localhost:8080/events` shows new JSON events | [ ] | |
| 3.5 | Dashboard `/` shows readable event list + motion/alert counts | [ ] | |
| 3.6 | Alert fires on `motion_detected` (rise only) | [ ] | Default: loud log line; optional `-alert-cmd` |
| 3.7 | Firewall: port 8080 allowed if ufw enabled | [ ] | `sudo ufw allow 8080` if needed |

**Phase 2 gate:** Phone dashboard + working alert on motion.

**Optional alert commands:**

```bash
# Default — structured log alert (no extra hardware)
go run ./cmd/sleepguard -device nursery -report-interval 5s -debug

# Shell command on each alert (if you have a .wav file)
go run ./cmd/sleepguard -device nursery -alert-cmd "aplay /path/to/beep.wav"
```

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
| GPIO always HIGH | **Do not use GPIO2 (pin 3)** on this Pi; use GPIO17. Run `pinctrl get 17` |
| PIR stuck HIGH | Time Delay knob fully **counter-clockwise**; jumper on **H** |
| No motion events | PIR warm-up ~60 s; breadboard center gap; wrong physical pin |
| Too many events | Increase `-alert-cooldown`; check PIR sensitivity pot |
| Cannot reach dashboard | Pi IP, firewall, same subnet, port 8080, app running |
| Camera fails | `raspi-config` camera on; cable seated; test CLI first |
| `go build` fails on Pi | `go mod tidy`; enough RAM; swap if needed |

---

## Changelog (updated by development sessions)

| Date | Phase | What was added / changed |
|------|-------|--------------------------|
| 2026-06-16 | — | Initial checklist created. Phase 0 code exists (flags only). |
| 2026-06-16 | — | Kit confirmed: Pi 4 4GB, HC-SR501, Camera v2, breadboard/M-F wires, 300Ω+LEDs, Wi‑Fi. |
| 2026-06-16 | 0 | Phase 0 verified on Pi (`192.168.0.103`). |
| 2026-06-16 | 1 | Phase 1 coded: PIR reader, mock sensor, pattern logs. Pi verified on GPIO17. |
| 2026-06-16 | 2 | Phase 2 coded: HTTP server, in-memory store, alert manager, dashboard. Pull + verify on Pi. |

---

*When you finish a task, change `[ ]` to `[x]` and add notes. Tell the agent when a phase gate is passed so we start the next phase.*
