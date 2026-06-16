# SleepGuard — Hetzner deploy (Phase A)

Cloud API on Docker. Postgres on host port **5433** (same server as wow-logs, **separate database**). nginx on 80/443 (configured once on the server; not in this repo).

## wow-logs server layout

| Item | Value |
|------|--------|
| Clone path | `/data/sleepguard` |
| Cloud port | `127.0.0.1:8090` (8091 = prabhujee) |
| Postgres | host `127.0.0.1:5433` — wow-logs uses DB `parser-v2` |
| Public URL | `https://sleepguard.rksdevs.in` |

Other ports in use: 8091 prabhujee, 8081 invenzo, 5434/5435 other Postgres.

### Create SleepGuard database

```bash
sudo -u postgres psql -p 5433
```

```sql
CREATE USER sleepguard WITH PASSWORD 'your-strong-password';
CREATE DATABASE sleepguard OWNER sleepguard;
\q
```

### Clone repo

```bash
mkdir -p /data/sleepguard
cd /data/sleepguard
git clone https://github.com/rksdevs/sleepguard.git .
```

### Configure and run Docker

```bash
cd /data/sleepguard/deploy
cp env.example .env
nano .env
```

```bash
DATABASE_URL=postgres://sleepguard:PASSWORD@127.0.0.1:5433/sleepguard?sslmode=disable
```

Docker uses `network_mode: host` so the container can reach Postgres on `127.0.0.1:5433` (host Postgres does not listen on the Docker bridge IP).

Generate secrets:

```bash
openssl rand -hex 32   # SLEEPGUARD_READ_API_KEY
openssl rand -hex 32   # SLEEPGUARD_BOOTSTRAP_DEVICE_TOKEN
```

```bash
docker compose up -d --build
docker compose logs -f cloud
```

Local health check:

```bash
curl -s http://127.0.0.1:8090/health
```

### nginx (one-time on server)

Mirror `prabhujee.rksdevs.in` (proxy to localhost). Create `/etc/nginx/sites-available/sleepguard.rksdevs.in` with `proxy_pass http://127.0.0.1:8090`, enable site, certbot, reload.

```bash
sudo cat /etc/nginx/sites-enabled/prabhujee.rksdevs.in   # reference
sudo certbot --nginx -d sleepguard.rksdevs.in
sudo nginx -t && sudo systemctl reload nginx
curl -s https://sleepguard.rksdevs.in/health
```

Phase B: serve PWA from `/data/sleepguard/web/pwa/dist` and proxy `/api/` to 8090 (edit existing nginx site after certbot).

#### Phase B — PWA + nginx update

Build on server:

```bash
cd /data/sleepguard
git pull
bash deploy/build-pwa.sh
```

Edit `/etc/nginx/sites-available/sleepguard.rksdevs.in` — inside the `server { listen 443 ssl; ... }` block, replace the single `location /` proxy with:

```nginx
client_max_body_size 25M;
root /data/sleepguard/web/pwa/dist;
index index.html;

location /api/ {
    proxy_pass http://127.0.0.1:8090/api/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

location /health {
    proxy_pass http://127.0.0.1:8090/health;
}

location / {
    try_files $uri $uri/ /index.html;
}
```

```bash
sudo nginx -t && sudo systemctl reload nginx
```

Open `https://sleepguard.rksdevs.in` in a browser (Cloudflare challenge once), enter read API key from `.env`, device id `nursery`.

**Cloudflare (orange proxy):** Browsers pass the challenge; the **Pi agent cannot**. Before running the agent, add a WAF custom rule in Cloudflare:

- **If** URI Path starts with `/api/v1/` **and** User Agent contains `SleepGuard-Agent` → **Skip** all security rules (or skip Managed Challenge)

Or set a **DNS-only (grey cloud)** record for a separate hostname like `api.sleepguard.rksdevs.in` used only by the Pi.

Snapshots (phase F): `/data/sleepguard/data/snapshots`

---

## Phase C — Pi agent

On the **Raspberry Pi**:

```bash
cd ~/sleepguard
git pull
go build -o bin/sleepguard-agent ./cmd/agent

cp deploy/agent.env.example deploy/agent.env
nano deploy/agent.env   # SLEEPGUARD_DEVICE_TOKEN from Hetzner deploy/.env
```

Test manually (wave at PIR after ~60s warm-up):

```bash
go run ./cmd/agent -debug
```

Install systemd (edit paths in unit if your home dir differs):

```bash
mkdir -p ~/sleepguard/bin
go build -o ~/sleepguard/bin/sleepguard-agent ./cmd/agent
sudo cp deploy/systemd/sleepguard-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now sleepguard-agent
sudo journalctl -u sleepguard-agent -f
```

Verify on PWA: status **ONLINE**, new events when motion detected.

---

## Create a device (alternative to bootstrap)

```bash
cd /data/sleepguard/deploy
docker compose run --rm cloud \
  -database-url "postgres://sleepguard:PASSWORD@127.0.0.1:5433/sleepguard?sslmode=disable" \
  -create-device \
  -device-id nursery \
  -device-name "Nursery Pi" \
  -device-token "YOUR_DEVICE_TOKEN"
```

## Smoke test

```bash
curl -s -X POST https://sleepguard.rksdevs.in/api/v1/events \
  -H "Authorization: Bearer DEVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"motion","source":"nursery","state":"active","pattern":"rise"}'

curl -s "https://sleepguard.rksdevs.in/api/v1/events?device_id=nursery&limit=10" \
  -H "Authorization: Bearer READ_KEY"
```

## API summary

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/health` | none | Liveness |
| POST | `/api/v1/events` | device token | Ingest motion event |
| POST | `/api/v1/heartbeat` | device token | Update last_seen |
| GET | `/api/v1/agent/commands` | device token | Fast poll for capture requests (~5s) |
| GET | `/api/v1/events?device_id=&limit=` | read key or device token | List events |
| GET | `/api/v1/devices/{id}/status` | read key or device token | Device status |
| POST | `/api/v1/devices/{id}/capture` | read key | Queue manual snapshot (Pi must be online) |
| GET | `/api/v1/snapshots?device_id=` | read key | List recent snapshots |
| GET | `/api/v1/snapshots/{id}/image` | read key | JPEG bytes |
| POST | `/api/v1/snapshots` | device token | Upload snapshot from Pi |
| POST | `/api/v1/admin/cleanup` | read key | Purge old events/snapshots |
| GET | `/api/v1/push/vapid-key` | none | VAPID public key for PWA |
| POST | `/api/v1/pair` | read key | Register phone for push |
| GET | `/api/v1/pair?device_id=` | read key | List paired phones |
| DELETE | `/api/v1/pair/{id}` | read key | Unpair a phone |

---

## Phase D — Push notifications + cleanup

### VAPID keys (one-time)

On the server or your dev PC (Node):

```bash
npx web-push generate-vapid-keys
```

Or with Go from the repo root:

```bash
go run ./scripts/gen-vapid-keys.go
```

Add to `deploy/.env`:

```bash
SLEEPGUARD_VAPID_PUBLIC_KEY=...
SLEEPGUARD_VAPID_PRIVATE_KEY=...
SLEEPGUARD_EVENT_RETENTION=24h
SLEEPGUARD_CLEANUP_INTERVAL=24h
```

Rebuild and restart:

```bash
cd /data/sleepguard
git pull
docker compose -f deploy/docker-compose.yml up -d --build
bash deploy/build-pwa.sh
sudo nginx -t && sudo systemctl reload nginx
```

### Test notifications

1. Open PWA on your phone (Add to Home Screen).
2. Tap **Enable notifications** and allow permission.
3. After **3 motion cycles** (rise→fall pairs), you get a push; again at 6, 9, etc.

Manual cleanup from PWA: **Run cleanup now** (or `curl -X POST .../api/v1/admin/cleanup -H "Authorization: Bearer READ_KEY"`).

---

## Phase E — Pattern rules (push every 3 cycles)

Server counts **rise → fall** as one motion cycle (hold/initial ignored).

| Cycles | Action |
|--------|--------|
| 3, 6, 9… | Web Push: sustained motion alert |

Optional `deploy/.env`:

```bash
SLEEPGUARD_RULE_NOTIFY_CYCLES=3
SLEEPGUARD_RULE_IDLE_RESET=10m
```

---

## Phase F — Manual camera capture

User taps **Capture image** in the PWA (anytime; Pi must show **online**).

```text
PWA → POST /api/v1/devices/{id}/capture
    → cloud queues command
    → Pi polls GET /api/v1/agent/commands every ~5s
    → rpicam-still on Pi
    → POST /api/v1/snapshots (multipart)
    → PWA shows latest image
```

**Pi prerequisites:** Pi Camera v2 on CSI; install `rpicam-apps` (or `libcamera-apps` on older images). On recent Pi OS use **`rpicam-still`**, not `libcamera-still`:

```bash
sudo apt install -y rpicam-apps
rpicam-hello --list-cameras    # should show imx219
rpicam-still -o ~/test.jpg -n -t 2000
```

Deploy:

```bash
# Hetzner — Docker cloud runs as uid 10001; snapshot volume must be writable
sudo mkdir -p /data/sleepguard/data/snapshots
sudo chown -R 10001:10001 /data/sleepguard/data/snapshots

cd /data/sleepguard && git pull
docker compose -f deploy/docker-compose.yml up -d --build
bash deploy/build-pwa.sh

# Pi
cd ~/sleepguard && git pull
go build -o ~/sleepguard/bin/sleepguard-agent ./cmd/agent
sudo systemctl restart sleepguard-agent
```

Test capture: open PWA → **Capture image** → image should appear within **~5–10 seconds** (command poll + camera + upload).

Smoke test cycles (push):

```bash
for i in 1 2 3; do
  curl -s -X POST https://sleepguard.rksdevs.in/api/v1/events \
    -H "Authorization: Bearer DEVICE_TOKEN" -H "Content-Type: application/json" \
    -d '{"type":"motion","state":"active","pattern":"rise"}'
  curl -s -X POST https://sleepguard.rksdevs.in/api/v1/events \
    -H "Authorization: Bearer DEVICE_TOKEN" -H "Content-Type: application/json" \
    -d '{"type":"motion","state":"idle","pattern":"fall"}'
done
# → push on 3rd cycle
```

## Local dev without Docker

```bash
export DATABASE_URL='postgres://sleepguard:pass@localhost:5433/sleepguard?sslmode=disable'
export SLEEPGUARD_READ_API_KEY='dev-read-key'

go run ./cmd/cloud -create-device -device-id nursery -device-name Nursery -device-token dev-device-token
go run ./cmd/cloud -debug
```
