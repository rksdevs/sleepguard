# SleepGuard Implementation Plan

Phase-by-phase plan for **edge (Pi) + cloud (Hetzner) + PWA**.  
Pi-specific steps: [checklist.md](checklist.md). Architecture: [architecture.md](architecture.md).

**Architecture shift (2026):** PIR verified locally. Primary UX moves to **cloud API + PWA**. Pi runs a lightweight **agent** that uploads events (and later snapshots). Pattern rules, pairing, and notifications live on the **server**.

---

## Completed — Local Pi foundation

| Phase | Status | Notes |
|-------|--------|-------|
| 0 — Bootstrap | Done | Go module, config, slog |
| 1 — PIR / GPIO | Done | GPIO17, HC-SR501 verified on Pi |
| 2 — Local HTTP dashboard | Done | Kept as dev fallback; not primary UX |

Local `cmd/sleepguard` remains useful for bench testing without cloud. Production path is `cmd/agent` + `cmd/cloud` + PWA.

---

## Target architecture

```text
┌──────────────────┐         HTTPS (events, heartbeats, images)
│  Pi — agent      │ ───────────────────────────────────────────►
│  PIR · camera    │                                              │
└──────────────────┘                              ┌─────────────▼────────────┐
                                                    │  Hetzner — cloud API     │
                                                    │  Postgres (sleepguard DB)│
                                                    │  disk: /data/snapshots   │
                                                    └─────────────┬────────────┘
                                                                  │ HTTPS
                                                    ┌─────────────▼────────────┐
                                                    │  PWA (static build)      │
                                                    │  logs · pairing · images │
                                                    └──────────────────────────┘
```

**Not full microservices** — a **modular monorepo** with two Go binaries + static frontend. Portfolio-friendly, easy to deploy from one clone on Pi and Hetzner.

### Repo layout (target)

```text
sleepguard/
├── cmd/
│   ├── agent/              # Pi edge process (sensor + upload + camera)
│   └── cloud/              # Hetzner API server
├── internal/
│   ├── domain/             # Shared Event DTOs, API types
│   ├── sensor/             # PIR reader (agent)
│   ├── camera/             # Snapshot capture (agent, phase F)
│   ├── agent/
│   │   ├── config/
│   │   ├── upload/         # HTTP client, retry, offline queue
│   │   └── runtime/        # Wire sensor → uploader
│   └── cloud/
│       ├── api/            # HTTP handlers, routing
│       ├── auth/           # Device tokens, user/session for PWA
│       ├── store/          # Postgres repositories
│       ├── rules/          # Pattern engine (cycle counting)
│       ├── push/           # Web Push (VAPID)
│       └── cleanup/        # Retention jobs
├── web/
│   └── pwa/                # Vite/React or similar → static dist/
├── migrations/             # SQL schema (goose or embed)
├── deploy/
│   ├── docker-compose.yml  # cloud service (+ optional sidecar)
│   ├── Dockerfile.cloud
│   └── systemd/            # Pi agent unit (phase C)
│       └── sleepguard-agent.service
└── docs/
```

### Postgres

- Use existing Hetzner Postgres **instance** with a **separate database** `sleepguard` — do **not** share tables with `wow-logs`.
- ~5k–10k events/day with daily cleanup is trivial for Postgres.
- No TimescaleDB required.

### Images

- Store JPEGs on **local disk** (e.g. `/var/lib/sleepguard/snapshots/`).
- DB holds metadata + file path only.
- Cleanup deletes rows **and** files.

### PWA distribution

- `npm run build` → `web/pwa/dist/`
- Served by Caddy/nginx on Hetzner (same origin as API — avoids CORS pain).
- Family opens `https://sleepguard.yourdomain.com` — bookmark to home screen (installable PWA).

---

## Where pattern rules should live

**Recommendation: server (`internal/cloud/rules`).**

| Factor | Pi (India home) | Server (EU Hetzner) |
|--------|-----------------|---------------------|
| RTT India → EU | N/A | ~150–250 ms typical |
| Event rate | — | Small JSON every few seconds at most |
| Rule timing | Cycles are **seconds apart** | Latency irrelevant |
| Change thresholds | Needs SSH / redeploy | Edit in PWA settings → DB |
| Pairing / Web Push | Cannot own pairing | Natural home |
| Wife / family settings | Per-phone config belongs in cloud | Yes |
| Pi offline | Could alert locally | Buffer + upload later; optional local log-only fallback |

200 Mbps and a premium Hetzner project are more than enough. Bottleneck is not bandwidth — it is **keeping one place for logic** (server) so the Pi stays a **dumb, reliable sensor**.

**Pi agent sends:** raw events (`rise`, `fall`, `hold`, `initial`) + heartbeat.  
**Server does:** cycle counting, notify at 3, request snapshot at 5, Web Push, cleanup.

**Optional later:** Pi plays a local buzzer on `rise` if cloud is unreachable (fail-safe). Not v1.

---

## Phase A — Cloud API + Postgres (Hetzner)

**Goal:** Deployable cloud service that accepts and stores events.

**Status:** Complete on Hetzner.

| Task | Deliverable |
|------|-------------|
| `internal/cloud/migrate/sql/` | `devices`, `events` tables |
| `cmd/cloud` | HTTP server, health, ingest |
| Device auth | API key per Pi (`Authorization: Bearer <device-token>`) |
| `POST /api/v1/events` | Ingest event JSON from agent |
| `GET /api/v1/events` | List events (newest first), filter by device |
| `GET /api/v1/devices/{id}/status` | last_seen, event count |
| `deploy/docker-compose.yml` | Cloud container, env, volume for snapshots dir |
| `deploy/README.md` | Hetzner setup and smoke tests |

**Done when:** `curl` can POST a fake event and GET it back from Hetzner.

**Deploy:** See [deploy/README.md](../deploy/README.md).

---

## Phase B — PWA v1 (read-only live log)

**Goal:** Browser UI on Hetzner showing motion events without touching the Pi dashboard.

**Status:** Code complete — build on server and update nginx.

| Task | Deliverable |
|------|-------------|
| `web/pwa/` | Vite + TypeScript, installable manifest |
| Event log page | Newest first, poll every 3 s |
| Device status | Online badge, event count, last seen |
| Auth v1 | Read API key in UI → sessionStorage |
| `deploy/build-pwa.sh` | Server build script |
| nginx | PWA static + `/api/` proxy (deploy README) |

**Done when:** Open PWA from browser, enter read key, see events.

---

## Phase C — Pi agent → cloud stream

**Goal:** Real PIR events appear in PWA within seconds.

**Status:** Code complete — deploy on Pi.

| Task | Deliverable |
|------|-------------|
| `cmd/agent` | Edge binary: sensor + cloud upload |
| `internal/agent/upload` | POST events + heartbeat with retry |
| `internal/agent/queue` | JSONL offline queue |
| `deploy/agent.env.example` | Pi credentials template |
| `deploy/systemd/sleepguard-agent.service` | Auto-start on boot |

**Done when:** Wave at PIR → event in PWA within seconds; device shows ONLINE.

---

## Phase D — Pairing, settings, Web Push

**Goal:** Phones paired via portal; configurable notifications including alarm-style push.

| Task | Deliverable |
|------|-------------|
| DB | `paired_clients`, `client_settings` |
| PWA pairing flow | Generate code / QR → register push subscription |
| Unpair | Revoke client in portal |
| `internal/cloud/push` | Web Push (VAPID), priority for alarm |
| Settings UI | Per client: enable notify, cycle threshold override, require interaction |
| PWA manifest + service worker | Installable, push permissions |

**Done when:** Wife pairs phone; motion rule (simple: every `rise` or first rule version) triggers push in India.

---

## Phase E — Pattern rules on server

**Goal:** 3 consecutive motion cycles → notify; 5 → trigger snapshot request.

| Task | Deliverable |
|------|-------------|
| `internal/cloud/rules` | State machine on event stream per device |
| Define cycle | e.g. `rise` → `fall` completes one cycle (document in code) |
| Thresholds | Default 3 / 5; overridable in `client_settings` |
| Agent command channel | Heartbeat response `{ "capture_snapshot": true }` when threshold hit |
| Cooldowns | Avoid notification spam |

**Done when:** Simulated or real motion cycles trigger push at 3; agent receives capture command at 5.

---

## Phase F — Camera + snapshots

**Goal:** JPEG on disk, visible in PWA, linked to event.

| Task | Deliverable |
|------|-------------|
| `internal/camera` | `libcamera-still` / `raspistill` on agent |
| `POST /api/v1/snapshots` | Multipart upload from agent |
| Disk storage | Save under `/var/lib/sleepguard/snapshots/` |
| `GET /api/v1/snapshots/{id}` | Serve image (auth required) |
| PWA gallery | Latest + per-event thumbnail; push links to snapshot |

**Done when:** 5-cycle rule produces image in PWA from India.

---

## Phase G — Cleanup, polish, portfolio

**Goal:** Retention, docs, production hardening.

| Task | Deliverable |
|------|-------------|
| `internal/cloud/cleanup` | Delete events + snapshot files older than retention |
| Manual cleanup | PWA button “Clean up now” (keeps devices + settings) |
| Scheduled cleanup | Cron or daily job (e.g. 04:00) |
| README + architecture | Screenshots, demo flow, Hetzner deploy steps |
| Remove or gate legacy Pi dashboard | Document as dev-only |

**Done when:** Cleanup runs; pairing/settings survive; README tells portfolio story.

---

## Phase H — Future backlog

| Item | Notes |
|------|-------|
| WebRTC live stream | Signaling + coturn on Hetzner; heavy on Pi CPU |
| Telegram / ntfy fallback | If Web Push alarm insufficient on some phones |
| Multi-room devices | Multiple `device_id`s, same PWA |
| Grafana on wow-logs | Separate concern; do not mix DBs |

---

## Phase dependencies

```text
[A: Cloud API] ──► [B: PWA read-only]
        │
        └──────────► [C: Pi agent] ──► [D: Pairing + Push] ──► [E: Rules] ──► [F: Camera]
                                                                              │
                                                                              ▼
                                                                        [G: Cleanup + polish]
```

Phases A + B can start **without Pi**. Phase C connects verified PIR. D–F build on cloud.

---

## Suggested timeline (~2 h/day)

| Phase | Focus | Approx. |
|-------|--------|---------|
| A | Cloud + Postgres | 3–5 days |
| B | PWA live log | 3–4 days |
| C | Pi agent stream | 2–4 days |
| D | Pairing + push | 4–6 days |
| E | Pattern rules | 2–3 days |
| F | Camera | 3–5 days |
| G | Cleanup + docs | 2–3 days |

---

## Definition of done (cloud MVP)

- [x] PIR motion on Pi (local verification)
- [ ] Events ingested on Hetzner Postgres
- [ ] PWA shows live log (newest first) from India
- [ ] Pi agent auto-starts via systemd
- [ ] Phones paired; Web Push notifications work
- [ ] Server pattern rules: 3 cycles → notify, 5 → snapshot
- [ ] Snapshots on disk + visible in PWA
- [ ] Cleanup removes logs/images; keeps devices + settings
- [ ] Single-repo deploy docs for Pi + Hetzner

---

## How we work with the checklist

1. Implement code for the current phase.
2. Update [checklist.md](checklist.md) with Pi **and** Hetzner steps.
3. Deploy and verify “done when”.
4. Move to next phase.
