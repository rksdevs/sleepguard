package domain

import "time"

// Snapshot is a camera image captured by the Pi agent.
type Snapshot struct {
	ID         string    `json:"id"`
	DeviceID   string    `json:"device_id"`
	CapturedAt time.Time `json:"captured_at"`
	SizeBytes  int64     `json:"size_bytes"`
}
