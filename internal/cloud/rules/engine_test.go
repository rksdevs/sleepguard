package rules

import (
	"log/slog"
	"testing"
	"time"

	"github.com/rksdevs/sleepguard/internal/domain"
)

func testEngine() *Engine {
	return New(Config{
		NotifyCycles: 3,
		IdleReset:    time.Hour,
	}, slog.Default())
}

func completeCycle(e *Engine, device string, t time.Time) Result {
	e.Process(device, domain.PatternRise, t)
	return e.Process(device, domain.PatternFall, t.Add(time.Second))
}

func TestEngine_NotifyEveryThreeCycles(t *testing.T) {
	e := testEngine()
	base := time.Now()

	for i := 1; i <= 9; i++ {
		r := completeCycle(e, "nursery", base.Add(time.Duration(i)*time.Minute))
		shouldNotify := i%3 == 0
		if r.Notify != shouldNotify {
			t.Fatalf("cycle %d: notify=%v, want %v", i, r.Notify, shouldNotify)
		}
		if r.Cycles != i {
			t.Fatalf("cycle %d: count=%d", i, r.Cycles)
		}
	}
}

func TestEngine_IgnoresHoldAndInitial(t *testing.T) {
	e := testEngine()
	now := time.Now()
	e.Process("nursery", domain.PatternInitial, now)
	e.Process("nursery", domain.PatternHold, now.Add(time.Second))
	r := completeCycle(e, "nursery", now.Add(2*time.Second))
	if r.Cycles != 1 {
		t.Fatalf("cycles=%d, want 1", r.Cycles)
	}
}

func TestEngine_IdleReset(t *testing.T) {
	e := testEngine()
	base := time.Now()
	completeCycle(e, "nursery", base)
	completeCycle(e, "nursery", base.Add(time.Minute))

	r := completeCycle(e, "nursery", base.Add(2*time.Hour))
	if r.Cycles != 1 {
		t.Fatalf("after idle reset cycles=%d, want 1", r.Cycles)
	}
}
