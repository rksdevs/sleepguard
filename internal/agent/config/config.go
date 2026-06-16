package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	appconfig "github.com/rksdevs/sleepguard/internal/config"
)

// Config holds Pi agent settings.
type Config struct {
	CloudURL           string
	DeviceToken        string
	DeviceName         string
	QueuePath          string
	HeartbeatInterval  time.Duration
	UploadTimeout      time.Duration
	Debug              bool
	MockSensor         bool
	GPIOPin            int
	PollInterval       time.Duration
	ReportInterval     time.Duration
}

// Load parses flags and environment for the edge agent.
func Load() (Config, error) {
	loadDefaultEnvFiles()

	cfg := Config{}

	flag.StringVar(&cfg.CloudURL, "cloud-url", envOr("SLEEPGUARD_CLOUD_URL", "https://sleepguard.rksdevs.in"), "cloud API base URL")
	flag.StringVar(&cfg.DeviceToken, "device-token", os.Getenv("SLEEPGUARD_DEVICE_TOKEN"), "device API token")
	flag.StringVar(&cfg.DeviceName, "device", envOr("SLEEPGUARD_DEVICE", "nursery"), "device / sensor source name")
	flag.StringVar(&cfg.QueuePath, "queue-path", envOr("SLEEPGUARD_QUEUE_PATH", "data/agent-queue.jsonl"), "offline event queue file")
	flag.DurationVar(&cfg.HeartbeatInterval, "heartbeat-interval", 60*time.Second, "cloud heartbeat interval")
	flag.DurationVar(&cfg.UploadTimeout, "upload-timeout", 15*time.Second, "HTTP timeout per upload request")
	flag.BoolVar(&cfg.Debug, "debug", envBool("SLEEPGUARD_DEBUG"), "enable debug logging")
	flag.BoolVar(&cfg.MockSensor, "mock-sensor", false, "use simulated sensor")
	flag.IntVar(&cfg.GPIOPin, "gpio-pin", 17, "BCM GPIO pin for PIR OUT")
	flag.DurationVar(&cfg.PollInterval, "poll-interval", 200*time.Millisecond, "PIR poll interval")
	flag.DurationVar(&cfg.ReportInterval, "report-interval", 5*time.Second, "steady-state log interval")

	flag.Parse()

	if cfg.DeviceToken == "" {
		cfg.DeviceToken = os.Getenv("SLEEPGUARD_DEVICE_TOKEN")
	}

	if cfg.CloudURL == "" {
		return Config{}, fmt.Errorf("cloud-url is required")
	}
	if cfg.DeviceToken == "" {
		return Config{}, fmt.Errorf("device-token or SLEEPGUARD_DEVICE_TOKEN is required")
	}
	if cfg.DeviceName == "" {
		return Config{}, fmt.Errorf("device name must not be empty")
	}

	return cfg, nil
}

// SensorConfig maps agent settings to the shared sensor package config.
func (c Config) SensorConfig() appconfig.Config {
	return appconfig.Config{
		DeviceName:     c.DeviceName,
		GPIOPin:        c.GPIOPin,
		PollInterval:   c.PollInterval,
		ReportInterval: c.ReportInterval,
		MockSensor:     c.MockSensor,
		Debug:          c.Debug,
	}
}

// NewLogger returns a logger for the agent.
func NewLogger(debug bool) *slog.Logger {
	return appconfig.NewLogger(debug)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string) bool {
	v := os.Getenv(key)
	return v == "1" || v == "true" || v == "yes"
}
