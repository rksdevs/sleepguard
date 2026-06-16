package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// Config holds runtime settings loaded from flags and environment.
type Config struct {
	DeviceName    string
	HTTPAddr      string
	AlertCooldown time.Duration
	ReportInterval time.Duration
	Debug         bool
	GPIOPin       int
	PollInterval  time.Duration
	StorePath     string
	CameraEnabled bool
	SnapshotDir   string
	MockSensor    bool
	AlertCommand  string
	StoreCapacity int
}

// Load parses flags and returns the application configuration.
func Load() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.DeviceName, "device", "nursery", "device name (e.g. nursery)")
	flag.StringVar(&cfg.HTTPAddr, "http-addr", ":8080", "HTTP listen address")
	flag.DurationVar(&cfg.AlertCooldown, "alert-cooldown", 30*time.Second, "minimum time between motion alerts")
	flag.DurationVar(&cfg.ReportInterval, "report-interval", 5*time.Second, "interval for repeating steady-state PIR logs")
	flag.BoolVar(&cfg.Debug, "debug", false, "enable debug logging")
	flag.IntVar(&cfg.GPIOPin, "gpio-pin", 17, "BCM GPIO pin for PIR OUT signal")
	flag.DurationVar(&cfg.PollInterval, "poll-interval", 200*time.Millisecond, "PIR poll interval")
	flag.StringVar(&cfg.StorePath, "store-path", "data/events.jsonl", "event history file path")
	flag.BoolVar(&cfg.CameraEnabled, "camera", false, "enable camera snapshots (phase 4)")
	flag.StringVar(&cfg.SnapshotDir, "snapshot-dir", "data/snapshots", "directory for motion snapshots")
	flag.BoolVar(&cfg.MockSensor, "mock-sensor", false, "use simulated sensor (dev machine without GPIO)")
	flag.StringVar(&cfg.AlertCommand, "alert-cmd", "", "optional shell command to run on motion alert (e.g. aplay)")
	flag.IntVar(&cfg.StoreCapacity, "store-capacity", 100, "maximum in-memory events to retain")

	flag.Parse()

	if cfg.DeviceName == "" {
		return Config{}, fmt.Errorf("device name must not be empty")
	}
	if cfg.AlertCooldown < 0 {
		return Config{}, fmt.Errorf("alert-cooldown must not be negative")
	}
	if cfg.ReportInterval < 0 {
		return Config{}, fmt.Errorf("report-interval must not be negative")
	}
	if cfg.GPIOPin < 0 {
		return Config{}, fmt.Errorf("gpio-pin must not be negative")
	}

	return cfg, nil
}

// NewLogger returns a structured logger configured for the given debug level.
func NewLogger(debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}

// LogAttrs returns stable key/value pairs for logging the loaded config.
func (c Config) LogAttrs() []any {
	return []any{
		"device", c.DeviceName,
		"http_addr", c.HTTPAddr,
		"alert_cooldown", c.AlertCooldown.String(),
		"report_interval", c.ReportInterval.String(),
		"debug", c.Debug,
		"gpio_pin", c.GPIOPin,
		"poll_interval", c.PollInterval.String(),
		"store_path", c.StorePath,
		"camera_enabled", c.CameraEnabled,
		"snapshot_dir", c.SnapshotDir,
		"mock_sensor", c.MockSensor,
		"alert_cmd", c.AlertCommand,
		"store_capacity", c.StoreCapacity,
	}
}
