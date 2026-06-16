package alert

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/rksdevs/sleepguard/internal/sensor"
)

// State represents the alert state machine phase.
type State string

const (
	StateIdle            State = "idle"
	StateMotionDetected  State = "motion_detected"
	StateAlertSent       State = "alert_sent"
	StateCooldown        State = "cooldown"
)

// Manager applies cooldown and dispatches alerts for motion rises.
type Manager struct {
	notifier Notifier
	cooldown time.Duration
	log      *slog.Logger

	mu        sync.Mutex
	state     State
	lastAlert time.Time
	onAlert   func()
}

// NewManager creates an alert manager with the given notifier and cooldown.
func NewManager(notifier Notifier, cooldown time.Duration, log *slog.Logger, onAlert func()) *Manager {
	return &Manager{
		notifier: notifier,
		cooldown: cooldown,
		log:      log,
		state:    StateIdle,
		onAlert:  onAlert,
	}
}

// Handle processes a sensor event and may trigger an alert.
func (m *Manager) Handle(ctx context.Context, event sensor.Event) {
	if event.Pattern != sensor.PatternRise {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if !m.lastAlert.IsZero() && now.Sub(m.lastAlert) < m.cooldown {
		m.log.Debug("cooldown_skipped",
			"since_last", now.Sub(m.lastAlert).String(),
			"cooldown", m.cooldown.String(),
		)
		m.state = StateCooldown
		return
	}

	m.state = StateMotionDetected

	if err := m.notifier.Notify(ctx, event); err != nil {
		m.log.Error("alert_failed", "error", err)
		return
	}

	m.lastAlert = now
	m.state = StateAlertSent
	if m.onAlert != nil {
		m.onAlert()
	}

	m.log.Info("alert_sent", "event", event.String())
}

// CurrentState returns the latest alert state.
func (m *Manager) CurrentState() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}
