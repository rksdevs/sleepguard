package rules

import (
	"log/slog"
	"sync"
	"time"

	"github.com/rksdevs/sleepguard/internal/domain"
)

// Config holds cycle rule thresholds and timing.
type Config struct {
	// NotifyCycles triggers a push every time the cycle count hits a multiple of this value.
	NotifyCycles int
	// IdleReset clears cycle state after this long without rise/fall events.
	IdleReset time.Duration
}

// DefaultConfig returns production defaults (notify every 3 cycles).
func DefaultConfig() Config {
	return Config{
		NotifyCycles: 3,
		IdleReset:    10 * time.Minute,
	}
}

// Result is produced when an event is evaluated.
type Result struct {
	Cycles int
	Notify bool
}

// Engine tracks per-device motion cycles on the server.
//
// Cycle definition: one complete rise → fall pair from the PIR agent.
// Hold and initial events are ignored for counting.
type Engine struct {
	cfg  Config
	log  *slog.Logger
	mu   sync.Mutex
	byID map[string]*deviceState
}

type deviceState struct {
	inMotion      bool
	cycles        int
	lastPatternAt time.Time
}

// New creates a rules engine.
func New(cfg Config, log *slog.Logger) *Engine {
	if cfg.NotifyCycles <= 0 {
		cfg.NotifyCycles = 3
	}
	if cfg.IdleReset <= 0 {
		cfg.IdleReset = 10 * time.Minute
	}
	return &Engine{
		cfg:  cfg,
		log:  log,
		byID: make(map[string]*deviceState),
	}
}

// Process evaluates a motion pattern and returns any triggered actions.
func (e *Engine) Process(deviceID, pattern string, at time.Time) Result {
	if pattern != domain.PatternRise && pattern != domain.PatternFall {
		return Result{}
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	st := e.state(deviceID)
	e.maybeResetIdle(st, at)

	var result Result

	switch pattern {
	case domain.PatternRise:
		if !st.inMotion {
			st.inMotion = true
		}
	case domain.PatternFall:
		if st.inMotion {
			st.inMotion = false
			st.cycles++
			result.Cycles = st.cycles

			if st.cycles%e.cfg.NotifyCycles == 0 {
				result.Notify = true
				e.log.Info("rule triggered",
					"device_id", deviceID,
					"cycles", st.cycles,
					"notify", true,
				)
			}
		}
	}

	st.lastPatternAt = at
	return result
}

func (e *Engine) state(deviceID string) *deviceState {
	st, ok := e.byID[deviceID]
	if !ok {
		st = &deviceState{}
		e.byID[deviceID] = st
	}
	return st
}

func (e *Engine) maybeResetIdle(st *deviceState, at time.Time) {
	if st.lastPatternAt.IsZero() {
		return
	}
	if at.Sub(st.lastPatternAt) >= e.cfg.IdleReset {
		st.inMotion = false
		st.cycles = 0
	}
}
