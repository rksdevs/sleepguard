package sensor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

// PIR reads an HC-SR501-style sensor on a GPIO pin.
type PIR struct {
	source         string
	pin            gpio.PinIO
	pollInterval   time.Duration
	reportInterval time.Duration
	log            *slog.Logger
	lastReport     time.Time
	previouslyHigh bool
	initialized    bool
}

// NewPIR configures GPIO and returns a PIR reader.
func NewPIR(source string, pinNumber int, pollInterval, reportInterval time.Duration, log *slog.Logger) (*PIR, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("init periph host: %w", err)
	}

	pinName := fmt.Sprintf("GPIO%d", pinNumber)
	pin := gpioreg.ByName(pinName)
	if pin == nil {
		return nil, fmt.Errorf("gpio pin %s not found", pinName)
	}

	if err := pin.In(gpio.PullDown, gpio.NoEdge); err != nil {
		return nil, fmt.Errorf("configure %s: %w", pinName, err)
	}

	return &PIR{
		source:         source,
		pin:            pin,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		log:            log,
	}, nil
}

// Run polls the PIR pin and reports initial, transition, and steady-state changes.
func (p *PIR) Run(ctx context.Context) error {
	p.log.Info("pir sensor started",
		"source", p.source,
		"poll_interval", p.pollInterval.String(),
		"report_interval", p.reportInterval.String(),
	)

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			p.poll()
		}
	}
}

func (p *PIR) poll() {
	active, err := p.readMotion()
	if err != nil {
		p.log.Error("sensor_error", "error", err)
		return
	}

	now := time.Now()

	if !p.initialized {
		p.initialized = true
		p.previouslyHigh = active
		p.lastReport = now
		p.emit(active, PatternInitial, "motion_state")
		return
	}

	if active != p.previouslyHigh {
		if active {
			p.emit(active, PatternRise, "motion_detected")
		} else {
			p.emit(active, PatternFall, "motion_ended")
		}
		p.previouslyHigh = active
		p.lastReport = now
		return
	}

	if p.reportInterval > 0 && now.Sub(p.lastReport) >= p.reportInterval {
		if active {
			p.emit(active, PatternHold, "motion_active")
		} else {
			p.emit(active, PatternHold, "motion_idle")
		}
		p.lastReport = now
	}
}

func (p *PIR) emit(active bool, pattern, message string) {
	state := StateIdle
	if active {
		state = StateActive
	}

	event := NewObservedEvent(EventMotion, p.source, state, pattern)
	p.log.Info(message, "event", event.String())
	p.previouslyHigh = active
}

func (p *PIR) readMotion() (bool, error) {
	level := p.pin.Read()
	if level == gpio.High {
		return true, nil
	}
	if level == gpio.Low {
		return false, nil
	}
	return false, fmt.Errorf("unexpected gpio level: %v", level)
}

// Close releases GPIO resources.
func (p *PIR) Close() error {
	if p.pin != nil {
		return p.pin.Halt()
	}
	return nil
}
