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
	source          string
	pin             gpio.PinIO
	pollInterval    time.Duration
	cooldown        time.Duration
	log             *slog.Logger
	lastEmit        time.Time
	previouslyHigh  bool
}

// NewPIR configures GPIO and returns a PIR reader.
func NewPIR(source string, pinNumber int, pollInterval, cooldown time.Duration, log *slog.Logger) (*PIR, error) {
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
		source:       source,
		pin:          pin,
		pollInterval: pollInterval,
		cooldown:     cooldown,
		log:          log,
	}, nil
}

// Run polls the PIR pin and logs motion with cooldown debouncing.
func (p *PIR) Run(ctx context.Context) error {
	p.log.Info("pir sensor started",
		"source", p.source,
		"poll_interval", p.pollInterval.String(),
		"cooldown", p.cooldown.String(),
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

	if active {
		if !p.previouslyHigh {
			p.onRisingEdge()
		}
	} else if p.previouslyHigh {
		p.log.Debug("motion_idle", "source", p.source)
	}

	p.previouslyHigh = active
}

func (p *PIR) onRisingEdge() {
	now := time.Now()
	if !p.lastEmit.IsZero() && now.Sub(p.lastEmit) < p.cooldown {
		p.log.Debug("cooldown_skipped",
			"source", p.source,
			"since_last", now.Sub(p.lastEmit).String(),
			"cooldown", p.cooldown.String(),
		)
		return
	}

	event := NewEvent(EventMotion, p.source, StateActive)
	p.lastEmit = now
	p.log.Info("motion_detected", "event", event.String())
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
