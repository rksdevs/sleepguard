package domain

import "time"

const (
	EventMotion = "motion"

	PatternInitial = "initial"
	PatternRise    = "rise"
	PatternFall    = "fall"
	PatternHold    = "hold"

	StateActive = "active"
	StateIdle   = "idle"
)

// Event is the shared motion event shape for agent ingest and cloud storage.
type Event struct {
	ID         int64     `json:"id,omitempty"`
	DeviceID   string    `json:"device_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Type       string    `json:"type"`
	Source     string    `json:"source"`
	State      string    `json:"state"`
	Pattern    string    `json:"pattern,omitempty"`
	ReceivedAt time.Time `json:"received_at,omitempty"`
}

// IngestEvent is the request body for POST /api/v1/events.
type IngestEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	State     string    `json:"state"`
	Pattern   string    `json:"pattern,omitempty"`
}

// DeviceStatus is returned by GET /api/v1/devices/{id}/status.
type DeviceStatus struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
	EventCount  int64      `json:"event_count"`
	Online      bool       `json:"online"`
	OnlineAfter time.Duration `json:"-"`
}
