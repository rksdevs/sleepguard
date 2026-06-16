CREATE TABLE IF NOT EXISTS snapshots (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES devices (id) ON DELETE CASCADE,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    file_name TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_snapshots_device_captured
    ON snapshots (device_id, captured_at DESC);
