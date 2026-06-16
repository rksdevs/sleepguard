package sensor

import (
	"context"
	"log/slog"
	"time"
)

// Mock simulates motion events for development without GPIO hardware.
type Mock struct {
	source        string
	interval      time.Duration
	log           *slog.Logger
	onEvent       EventHandler
	active        bool
	firstEmission bool
}

// NewMock returns a reader that emits alternating motion states.
func NewMock(source string, interval time.Duration, log *slog.Logger, onEvent EventHandler) *Mock {
	return &Mock{
		source:   source,
		interval: interval,
		log:      log,
		onEvent:  onEvent,
	}
}

// Run emits simulated motion on a timer until the context is cancelled.
func (m *Mock) Run(ctx context.Context) error {
	m.log.Info("mock sensor started", "source", m.source, "interval", m.interval.String())

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// First event soon after start so dev runs are easy to verify.
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(3 * time.Second):
		m.emit(true)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			m.emit(m.active)
		}
	}
}

func (m *Mock) emit(active bool) {
	var (
		state   string
		pattern string
		msg     string
	)

	if active {
		state = StateActive
		if !m.firstEmission || !m.active {
			pattern = PatternRise
			msg = "motion_detected"
		} else {
			pattern = PatternHold
			msg = "motion_active"
		}
	} else {
		state = StateIdle
		if m.active {
			pattern = PatternFall
			msg = "motion_ended"
		} else {
			pattern = PatternHold
			msg = "motion_idle"
		}
	}

	event := NewObservedEvent(EventMotion, m.source, state, pattern)
	m.log.Info(msg, "event", event.String())
	if m.onEvent != nil {
		m.onEvent(event)
	}
	m.firstEmission = true
	m.active = !m.active
}

// Close is a no-op for the mock reader.
func (m *Mock) Close() error {
	return nil
}
