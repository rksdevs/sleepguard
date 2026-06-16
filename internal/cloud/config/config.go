package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds cloud server settings from flags and environment.
type Config struct {
	HTTPAddr          string
	DatabaseURL       string
	ReadAPIKey        string
	SnapshotDir       string
	Debug             bool
	OnlineAfter       time.Duration
	MigrateOnly       bool
	CreateDevice        bool
	CreateDeviceID      string
	CreateDeviceName    string
	CreateDeviceToken   string
	BootstrapDeviceID   string
	BootstrapDeviceName string
	BootstrapDeviceToken string
}

// Load parses flags and environment variables.
func Load() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.HTTPAddr, "http-addr", envOr("SLEEPGUARD_HTTP_ADDR", ":8090"), "HTTP listen address")
	flag.StringVar(&cfg.DatabaseURL, "database-url", os.Getenv("DATABASE_URL"), "Postgres connection URL")
	flag.StringVar(&cfg.ReadAPIKey, "read-api-key", os.Getenv("SLEEPGUARD_READ_API_KEY"), "API key for read-only PWA/admin access")
	flag.StringVar(&cfg.SnapshotDir, "snapshot-dir", envOr("SLEEPGUARD_SNAPSHOT_DIR", "/data/snapshots"), "directory for snapshot files (phase F)")
	flag.BoolVar(&cfg.Debug, "debug", envBool("SLEEPGUARD_DEBUG"), "enable debug logging")
	flag.DurationVar(&cfg.OnlineAfter, "online-after", 2*time.Minute, "device considered online if last_seen within this duration")
	flag.BoolVar(&cfg.MigrateOnly, "migrate-only", false, "run migrations and exit")
	flag.BoolVar(&cfg.CreateDevice, "create-device", false, "create a device and exit")
	flag.StringVar(&cfg.CreateDeviceID, "device-id", "", "device id for -create-device")
	flag.StringVar(&cfg.CreateDeviceName, "device-name", "", "display name for -create-device")
	flag.StringVar(&cfg.CreateDeviceToken, "device-token", "", "raw API token for -create-device")
	flag.StringVar(&cfg.BootstrapDeviceID, "bootstrap-device-id", os.Getenv("SLEEPGUARD_BOOTSTRAP_DEVICE_ID"), "create device on startup if missing")
	flag.StringVar(&cfg.BootstrapDeviceName, "bootstrap-device-name", os.Getenv("SLEEPGUARD_BOOTSTRAP_DEVICE_NAME"), "name for bootstrap device")
	flag.StringVar(&cfg.BootstrapDeviceToken, "bootstrap-device-token", os.Getenv("SLEEPGUARD_BOOTSTRAP_DEVICE_TOKEN"), "token for bootstrap device")

	flag.Parse()

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("database-url or DATABASE_URL is required")
	}
	if cfg.CreateDevice {
		if cfg.CreateDeviceID == "" || cfg.CreateDeviceName == "" || cfg.CreateDeviceToken == "" {
			return Config{}, fmt.Errorf("-create-device requires -device-id, -device-name, and -device-token")
		}
	}
	if cfg.BootstrapDeviceToken != "" && (cfg.BootstrapDeviceID == "" || cfg.BootstrapDeviceName == "") {
		return Config{}, fmt.Errorf("bootstrap device requires id and name when token is set")
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string) bool {
	v := strings.ToLower(os.Getenv(key))
	return v == "1" || v == "true" || v == "yes"
}
