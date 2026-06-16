package sensor

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	EventMotion = "motion"

	StateActive = "active"
	StateIdle   = "idle"
)

// Event represents a single sensor reading or state change.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	State     string    `json:"state"`
}

// NewEvent builds an event with the current timestamp.
func NewEvent(eventType, source, state string) Event {
	return Event{
		Timestamp: time.Now().UTC(),
		Type:      eventType,
		Source:    source,
		State:     state,
	}
}

// String returns a short human-readable summary.
func (e Event) String() string {
	return fmt.Sprintf(
		"%s %s from %s at %s",
		e.Type,
		e.State,
		e.Source,
		e.Timestamp.Format(time.RFC3339),
	)
}

// JSON returns the event encoded as JSON.
func (e Event) JSON() ([]byte, error) {
	return json.Marshal(e)
}
