package store

import (
	"sync"
	"time"

	"github.com/rksdevs/sleepguard/internal/sensor"
)

// Memory keeps a bounded in-memory ring buffer of sensor events.
type Memory struct {
	mu          sync.RWMutex
	events      []sensor.Event
	capacity    int
	motionCount uint64
	alertCount  uint64
	startTime   time.Time
	lastEvent   sensor.Event
	hasLast     bool
}

// NewMemory creates a store that retains up to capacity events.
func NewMemory(capacity int) *Memory {
	if capacity < 1 {
		capacity = 100
	}
	return &Memory{
		capacity:  capacity,
		startTime: time.Now(),
	}
}

// Append records an event and updates counters.
func (m *Memory) Append(event sensor.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.events) >= m.capacity {
		copy(m.events, m.events[1:])
		m.events[len(m.events)-1] = event
	} else {
		m.events = append(m.events, event)
	}

	m.lastEvent = event
	m.hasLast = true

	if event.Pattern == sensor.PatternRise {
		m.motionCount++
	}
}

// RecordAlert increments the alert counter.
func (m *Memory) RecordAlert() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertCount++
}

// Recent returns up to limit most recent events, newest last.
func (m *Memory) Recent(limit int) []sensor.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.events) {
		limit = len(m.events)
	}
	if limit == 0 {
		return nil
	}

	start := len(m.events) - limit
	out := make([]sensor.Event, limit)
	copy(out, m.events[start:])
	return out
}

// Stats returns operational counters for the dashboard.
func (m *Memory) Stats() (motionCount, alertCount uint64, uptime time.Duration, last sensor.Event, hasLast bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.motionCount, m.alertCount, time.Since(m.startTime), m.lastEvent, m.hasLast
}

// StartTime returns when the store was created.
func (m *Memory) StartTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.startTime
}
