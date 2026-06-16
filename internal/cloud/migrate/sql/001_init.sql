-- SleepGuard cloud schema (phase A)

CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    api_key_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES devices (id) ON DELETE CASCADE,
    recorded_at TIMESTAMPTZ NOT NULL,
    type TEXT NOT NULL,
    source TEXT NOT NULL,
    state TEXT NOT NULL,
    pattern TEXT NOT NULL DEFAULT '',
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_events_device_recorded
    ON events (device_id, recorded_at DESC);

CREATE INDEX IF NOT EXISTS idx_events_received
    ON events (received_at DESC);
