# SleepGuard Implementation Plan

Phase-by-phase plan for **edge (Pi) + cloud (Hetzner) + PWA**.  
Pi-specific steps: [checklist.md](checklist.md). Architecture: [architecture.md](architecture.md). Ops: [deploy/README.md](../deploy/README.md).

**Architecture shift (2026):** PIR verified locally. Primary UX is **cloud API + PWA**. Pi runs a lightweight **agent** that uploads events and camera stills on command. Pattern rules, pairing, push, and cleanup live on the **server**.

---

## Status summary (June 2026)

| Phase | Status | Verified |
|-------|--------|----------|
| Local Pi PIR (0–1) | Done | GPIO17, HC-SR501 |
| **A** Cloud API + Postgres | Done | Hetzner |
| **B** PWA live log | Done | sleepguard.rksdevs.in |
| **C** Pi agent → cloud | Done | systemd, WAF skip |
| **D** Web Push + cleanup | Done | Phone + desktop |
| **E** Pattern rules | Done | Push at 3, 6, 9… cycles |
| **F** Manual camera capture | Done | ~5–10s capture-to-PWA |
| **G** Polish + portfolio | **RTC — next** | — |
| **H** Backlog | Future | WebRTC, ntfy, multi-device |

---

## RTC — Resume To Continue (Phase G)

When picking up development in a new session, start here.

### What works today

- Pi agent: PIR events + heartbeat + 5s command poll + `rpicam-still` capture
- Cloud: ingest, rules (rise→fall = 1 cycle), Web Push, snapshot store, 24h retention
- PWA: live log, device status, enable notifications, capture button, 5-snapshot carousel, download, manual cleanup

### Phase G tasks (recommended order)

1. **Portfolio assets** — screenshots (PWA log, push notification, capture carousel), short demo GIF
2. **README polish** — done in root README; add images when captured
3. **Per-client settings UI** — override `SLEEPGUARD_RULE_NOTIFY_CYCLES` per paired phone (DB + PWA)
4. **Push actions** — optional deep-link URL to open PWA on alert tap
5. **Legacy cleanup** — mark `cmd/sleepguard` dev-only in docs; no production deploy path
6. **Checklist audit** — keep [checklist.md](checklist.md) in sync after G

### Deploy quick reference

```bash
# Hetzner (cloud or PWA changes)
cd /data/sleepguard && git pull
sudo chown -R 10001:10001 /data/sleepguard/data/snapshots   # once per server
docker compose -f deploy/docker-compose.yml up -d --build
bash deploy/build-pwa.sh

# Pi (agent/camera changes only)
cd ~/sleepguard && git pull
go build -o ~/sleepguard/bin/sleepguard-agent ./cmd/agent
sudo systemctl restart sleepguard-agent
```

### Key env vars (`deploy/.env`)

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | Postgres `sleepguard` on `:5433` |
| `SLEEPGUARD_READ_API_KEY` | PWA + admin API |
| `SLEEPGUARD_BOOTSTRAP_DEVICE_TOKEN` | Pi agent auth |
| `SLEEPGUARD_VAPID_*` | Web Push |
| `SLEEPGUARD_EVENT_RETENTION` | Default `24h` |
| `SLEEPGUARD_RULE_NOTIFY_CYCLES` | Default `3` |
| `SLEEPGUARD_RULE_IDLE_RESET` | Default `10m` |

### Pi agent env (`deploy/agent.env`)

| Variable | Purpose |
|----------|---------|
| `SLEEPGUARD_CLOUD_URL` | `https://sleepguard.rksdevs.in` |
| `SLEEPGUARD_DEVICE_TOKEN` | From Hetzner `.env` |
| `SLEEPGUARD_COMMAND_POLL_INTERVAL` | Default `5s` |

---

## Completed — Local Pi foundation

| Phase | Status | Notes |
|-------|--------|-------|
| 0 — Bootstrap | Done | Go module, config, slog |
| 1 — PIR / GPIO | Done | GPIO17, HC-SR501 verified on Pi |
| 2 — Local HTTP dashboard | Done | Dev fallback only (`cmd/sleepguard`) |

Production path: `cmd/agent` + `cmd/cloud` + PWA.

---

## Target architecture

```text
┌──────────────────┐         HTTPS (events, heartbeats, images)
│  Pi — agent      │ ───────────────────────────────────────────►
│  PIR · camera    │         ◄── poll commands ~5s ──────────────
└──────────────────┘                              ┌─────────────▼────────────┐
                                                    │  Hetzner — cloud API     │
                                                    │  Postgres (sleepguard DB)│
                                                    │  disk: /data/snapshots   │
                                                    └─────────────┬────────────┘
                                                                  │ HTTPS
                                                    ┌─────────────▼────────────┐
                                                    │  PWA (static build)      │
                                                    │  log · push · capture    │
                                                    └──────────────────────────┘
```

Modular monorepo: two Go binaries + static frontend. One clone on Pi and Hetzner.

### Repo layout

```text
sleepguard/
├── cmd/agent/              # Pi edge (sensor + upload + camera)
├── cmd/cloud/              # Hetzner API
├── internal/camera/        # rpicam-still / libcamera-still
├── internal/cloud/
│   ├── rules/              # Cycle counting, push triggers
│   ├── commands/           # Capture queue
│   ├── push/               # Web Push (VAPID)
│   └── cleanup/            # Retention
├── web/pwa/                # Vite + TypeScript PWA
└── deploy/                 # Docker, systemd, env examples
```

---

## Where pattern rules live

**Server (`internal/cloud/rules`).** Pi sends raw events; server counts cycles and triggers push.

| Factor | Pi | Server |
|--------|-----|--------|
| Rule timing | — | Cycles seconds apart; EU RTT irrelevant |
| Web Push / pairing | Cannot | Natural home |
| Threshold changes | Needs redeploy | Env or future PWA settings |

**Cycle definition:** `rise` → `fall` = one cycle. Push at 3, 6, 9… (not per-rise, not once-only).  
**No auto-snapshot on motion** — capture is **manual** from PWA (Phase F).

---

## Phase A — Cloud API + Postgres

**Status:** Done.

Events ingest, device auth, status, Docker on Hetzner `:8090`, migrations `001`–`003`.

---

## Phase B — PWA v1

**Status:** Done.

Live event log, device status, read API key auth, nginx static + `/api/` proxy.

---

## Phase C — Pi agent → cloud

**Status:** Done.

`cmd/agent`, offline JSONL queue, heartbeat, Cloudflare WAF skip for `SleepGuard-Agent`, systemd unit.

---

## Phase D — Web Push + cleanup

**Status:** Done.

| Deliverable | Notes |
|-------------|-------|
| `paired_clients` table | Pair/unpair via PWA |
| VAPID Web Push | `scripts/gen-vapid-keys.go` or `npx web-push generate-vapid-keys` |
| Service worker | `web/pwa/public/sw.js` |
| Retention | 24h default; scheduler + PWA **Run cleanup now** |

---

## Phase E — Pattern rules

**Status:** Done (revised scope).

| Original plan | Shipped |
|---------------|---------|
| Push at 3 cycles; snapshot at 5 | Push at 3, 6, 9… only |
| Auto capture on rule | **Removed** — manual capture only |

`internal/cloud/rules/engine.go` — state machine per device, idle reset configurable.

---

## Phase F — Manual camera capture

**Status:** Done.

```text
PWA → POST /api/v1/devices/{id}/capture
    → commands.CaptureQueue
    → Pi polls GET /api/v1/agent/commands (~5s)
    → rpicam-still (fallback: libcamera-still, raspistill)
    → POST /api/v1/snapshots (multipart)
    → PWA carousel (5 recent) + lightbox + download
```

**Pi:** `rpicam-apps` on Bookworm+; camera `imx219` on CSI.  
**Hetzner:** `chown 10001:10001` on snapshot volume (container uid).

---

## Phase G — Polish + portfolio (RTC)

**Status:** Next.

| Task | Deliverable |
|------|-------------|
| Screenshots / demo | Portfolio + README |
| Per-client settings | PWA-editable notify threshold |
| Push deep links | Open PWA from notification |
| Legacy dashboard | Document `cmd/sleepguard` as dev-only |
| Final doc pass | Keep checklist + deploy README current |

**Done when:** Portfolio-ready story with visuals; optional settings UI; docs complete.

---

## Phase H — Future backlog

| Item | Notes |
|------|-------|
| **WebRTC live stream** | Signaling + coturn on Hetzner; Pi CPU heavy |
| Telegram / ntfy fallback | If Web Push insufficient on some phones |
| Multi-room devices | Multiple `device_id`s, same PWA |
| Grafana on wow-logs | Separate DB; do not mix |

---

## Phase dependencies

```text
[A] ──► [B]
 │
 └──► [C] ──► [D] ──► [E] ──► [F] ──► [G]
```

---

## Definition of done (cloud MVP)

- [x] PIR motion on Pi (GPIO17)
- [x] Events ingested on Hetzner Postgres
- [x] PWA shows live log from India
- [x] Pi agent auto-starts via systemd
- [x] Phones paired; Web Push notifications work
- [x] Server rules: push every 3 cycles (6, 9…)
- [x] Manual snapshots on disk + visible in PWA
- [x] Cleanup removes logs/images; keeps devices + settings
- [x] Single-repo deploy docs for Pi + Hetzner
- [ ] Portfolio screenshots + Phase G polish

---

## How we work with the checklist

1. Implement code for the current phase.
2. Update [checklist.md](checklist.md) with Pi **and** Hetzner steps.
3. Deploy and verify “done when”.
4. Move to next phase; update this file’s **RTC** section.
