package sensor

import (
	"context"
	"log/slog"
	"time"
)

// Mock simulates motion events for development without GPIO hardware.
type Mock struct {
	source   string
	cooldown time.Duration
	log      *slog.Logger
}

// NewMock returns a reader that emits a motion event every cooldown period.
func NewMock(source string, cooldown time.Duration, log *slog.Logger) *Mock {
	return &Mock{
		source:   source,
		cooldown: cooldown,
		log:      log,
	}
}

// Run emits simulated motion on a timer until the context is cancelled.
func (m *Mock) Run(ctx context.Context) error {
	m.log.Info("mock sensor started", "source", m.source, "interval", m.cooldown.String())

	ticker := time.NewTicker(m.cooldown)
	defer ticker.Stop()

	// First event soon after start so dev runs are easy to verify.
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(3 * time.Second):
		m.emit()
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			m.emit()
		}
	}
}

func (m *Mock) emit() {
	event := NewEvent(EventMotion, m.source, StateActive)
	m.log.Info("motion_detected", "event", event.String())
}

// Close is a no-op for the mock reader.
func (m *Mock) Close() error {
	return nil
}
