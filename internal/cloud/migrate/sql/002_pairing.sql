CREATE TABLE IF NOT EXISTS paired_clients (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES devices (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    push_endpoint TEXT NOT NULL,
    push_p256dh TEXT NOT NULL,
    push_auth TEXT NOT NULL,
    notify_on_rise BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_paired_clients_device
    ON paired_clients (device_id)
    WHERE revoked_at IS NULL;
