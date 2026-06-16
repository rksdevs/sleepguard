package sensor

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/rksdevs/sleepguard/internal/config"
)

// EventHandler receives sensor events as they are observed.
type EventHandler func(Event)

// MotionReader emits debounced motion events from hardware or a mock source.
type MotionReader interface {
	Run(ctx context.Context) error
	Close() error
}

// Open creates a motion reader from configuration.
func Open(cfg config.Config, log *slog.Logger, onEvent EventHandler) (MotionReader, error) {
	if cfg.MockSensor {
		return NewMock(cfg.DeviceName, cfg.ReportInterval, log, onEvent), nil
	}
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("GPIO sensor requires linux (use -mock-sensor on other platforms)")
	}
	return NewPIR(cfg.DeviceName, cfg.GPIOPin, cfg.PollInterval, cfg.ReportInterval, log, onEvent)
}
