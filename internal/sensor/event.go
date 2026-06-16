package sensor

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	EventMotion = "motion"

	PatternInitial = "initial"
	PatternRise    = "rise"
	PatternFall    = "fall"
	PatternHold    = "hold"

	StateActive = "active"
	StateIdle   = "idle"
)

// Event represents a single sensor reading or state change.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	State     string    `json:"state"`
	Pattern   string    `json:"pattern,omitempty"`
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

// NewObservedEvent builds an event and records the detection pattern.
func NewObservedEvent(eventType, source, state, pattern string) Event {
	event := NewEvent(eventType, source, state)
	event.Pattern = pattern
	return event
}

// String returns a short human-readable summary.
func (e Event) String() string {
	if e.Pattern != "" {
		return fmt.Sprintf(
			"%s %s (%s) from %s at %s",
			e.Type,
			e.State,
			e.Pattern,
			e.Source,
			e.Timestamp.Format(time.RFC3339),
		)
	}

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
