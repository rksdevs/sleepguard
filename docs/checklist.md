# SleepGuard — Raspberry Pi 4 Checklist

Action items to run on the **Pi** and verify on **Hetzner** after each phase.  
**You** complete hardware and deploy steps; **we** update this file when new code needs verification.

**Legend:** `[x]` done · `[ ]` todo · `[-]` skipped / N/A

**Production URL:** https://sleepguard.rksdevs.in · **Pi:** `rksdevs@192.168.0.103` · **Device id:** `nursery`

---

## One-time Pi setup

| # | Task | Status | Notes |
|---|------|--------|-------|
| 0.1 | Pi 4 imaged with Raspberry Pi OS (64-bit) | [x] | |
| 0.2 | Pi connected to home Wi‑Fi | [x] | IP may change per network |
| 0.3 | SSH enabled | [x] | `ssh rksdevs@192.168.0.103` |
| 0.4 | Hostname or static IP noted | [x] | `192.168.0.103` |
| 0.5 | Go 1.24+ installed (`go version`) | [x] | |
| 0.6 | Git clone repo on Pi | [x] | `~/sleepguard` |
| 0.7 | PIR wired per [electronics.md](electronics.md) | [x] | GPIO17 (pin 11) |
| 0.8 | Pi Camera v2 on CSI | [x] | `imx219`; ribbon toward HDMI |

---

## Phase A — Cloud API (Hetzner)

| # | Task | Status | Notes |
|---|------|--------|-------|
| A.1 | Postgres database `sleepguard` on `:5433` | [x] | Separate from wow-logs |
| A.2 | Clone `/data/sleepguard`, `deploy/.env` configured | [x] | |
| A.3 | `docker compose up -d --build` | [x] | `:8090` |
| A.4 | `curl localhost:8090/health` | [x] | |
| A.5 | POST test event with device token | [x] | |
| A.6 | GET events with read API key | [x] | |

**Gate:** Event ingested and readable via API. **Passed.**

---

## Phase B — PWA (Hetzner)

| # | Task | Status | Notes |
|---|------|--------|-------|
| B.1 | `git pull` + `bash deploy/build-pwa.sh` | [x] | Node 18+ |
| B.2 | nginx: `root` → `web/pwa/dist`, proxy `/api/` | [x] | |
| B.3 | Open `https://sleepguard.rksdevs.in` | [x] | Cloudflare challenge once |
| B.4 | Enter read API key, device `nursery` — see events | [x] | |

**Gate:** PWA shows live log. **Passed.**

---

## Phase C — Pi agent → cloud

| # | Task | Status | Notes |
|---|------|--------|-------|
| C.1 | Cloudflare WAF skip: `/api/v1/*` + `SleepGuard-Agent` | [x] | Orange proxy |
| C.2 | `go build -o bin/sleepguard-agent ./cmd/agent` | [x] | |
| C.3 | `deploy/agent.env` with device token | [x] | |
| C.4 | Test `go run ./cmd/agent -debug` | [x] | PIR warm-up ~60s |
| C.5 | PWA **ONLINE** + live events | [x] | |
| C.6 | `systemctl enable --now sleepguard-agent` | [x] | |

**Gate:** Real PIR motion in PWA without manual curl. **Passed.**

---

## Phase D — Web Push + cleanup

| # | Task | Status | Notes |
|---|------|--------|-------|
| D.1 | VAPID keys in `deploy/.env` | [x] | `gen-vapid-keys.go` |
| D.2 | Rebuild cloud + PWA on Hetzner | [x] | |
| D.3 | PWA **Enable notifications** on phone | [x] | Add to Home Screen |
| D.4 | Push after 3 motion cycles | [x] | 6, 9… repeat |
| D.5 | **Run cleanup now** or wait for scheduler | [x] | 24h retention |

**Gate:** Push received on phone; cleanup works. **Passed.**

---

## Phase E — Pattern rules

| # | Task | Status | Notes |
|---|------|--------|-------|
| E.1 | Cycle = rise → fall | [x] | hold/initial ignored |
| E.2 | Push at 3, 6, 9… cycles | [x] | Not once-only |
| E.3 | No auto-snapshot on motion | [x] | Manual capture only |

**Gate:** Sustained motion triggers recurring push. **Passed.**

---

## Phase F — Manual camera capture

| # | Task | Status | Notes |
|---|------|--------|-------|
| F.1 | `sudo apt install -y rpicam-apps` | [x] | Bookworm+ |
| F.2 | `rpicam-hello --list-cameras` shows imx219 | [x] | |
| F.3 | `rpicam-still -o ~/test.jpg -n -t 2000` | [x] | |
| F.4 | Hetzner: `chown 10001:10001` on snapshots dir | [x] | Fixes 500 store |
| F.5 | Rebuild agent + restart systemd | [x] | |
| F.6 | PWA **Capture image** → image in ~5–10s | [x] | 5s command poll |
| F.7 | Carousel shows 5 recent; download works | [x] | |

**Gate:** End-to-end manual capture from India. **Passed.**

---

## Phase G — Polish (RTC — next)

| # | Task | Status | Notes |
|---|------|--------|-------|
| G.1 | Portfolio screenshots (PWA, push, capture) | [ ] | |
| G.2 | Short demo video or GIF | [ ] | Optional |
| G.3 | Per-client notify threshold in PWA | [ ] | Phase G code |
| G.4 | Document legacy `cmd/sleepguard` as dev-only | [x] | In implementation plan |

---

## Legacy — local Pi dashboard (`cmd/sleepguard`)

Dev/bench only. Not used in production.

| # | Task | Status | Notes |
|---|------|--------|-------|
| 2.1 | Local dashboard on `:8080` | [x] | Phase 2 code |
| 3.1 | Persistence / telemetry | [-] | Superseded by cloud |

---

## Troubleshooting quick reference

| Symptom | Things to check |
|---------|-----------------|
| GPIO always HIGH | Use GPIO17, not GPIO2 (pin 3) |
| PIR stuck HIGH | Time Delay CCW; jumper on H |
| Agent 403/challenge | Cloudflare WAF skip for agent UA |
| Capture 500 | `chown 10001:10001 /data/sleepguard/data/snapshots` |
| `libcamera-still` not found | Use `rpicam-still` (install `rpicam-apps`) |
| Slow capture (~60s) | Agent must poll `/api/v1/agent/commands` (5s default) |
| Camera fails | CSI ribbon seated; `rpicam-hello --list-cameras` |

### Pi notes

```
Pi IP: 192.168.0.103
GPIO: GPIO17 (BCM) / pin 11
Camera: imx219, rpicam-still
Agent: sleepguard-agent.service
```

---

## Changelog

| Date | Phase | What changed |
|------|-------|--------------|
| 2026-06-16 | 0–2 | Local PIR + dashboard |
| 2026-06-16 | A–C | Cloud, PWA, Pi agent deployed |
| 2026-06-16 | D | Web Push, pairing, 24h cleanup |
| 2026-06-16 | E | Rules: push every 3 cycles; no auto-snapshot |
| 2026-06-16 | F | Manual capture, fast poll, carousel, download |
| 2026-06-16 | G | RTC — portfolio polish next |

---

*When a phase gate passes, mark tasks `[x]` and tell the agent to start the next phase.*
