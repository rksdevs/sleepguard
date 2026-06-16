package domain

import "time"

// HeartbeatResponse is returned to the Pi agent on POST /api/v1/heartbeat.
type HeartbeatResponse struct {
	DeviceID   string    `json:"device_id"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

// AgentCommands is returned by GET /api/v1/agent/commands (fast poll for capture).
type AgentCommands struct {
	CaptureSnapshot bool `json:"capture_snapshot,omitempty"`
}
