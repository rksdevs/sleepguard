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
	EventRetention       time.Duration
	CleanupInterval      time.Duration
	VAPIDPublicKey       string
	VAPIDPrivateKey      string
	VAPIDSubject         string
	RuleNotifyCycles int
	RuleIdleReset    time.Duration
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
	flag.DurationVar(&cfg.EventRetention, "event-retention", envDuration("SLEEPGUARD_EVENT_RETENTION", 24*time.Hour), "delete events older than this")
	flag.DurationVar(&cfg.CleanupInterval, "cleanup-interval", envDuration("SLEEPGUARD_CLEANUP_INTERVAL", 24*time.Hour), "automatic cleanup interval (0 disables)")
	flag.StringVar(&cfg.VAPIDPublicKey, "vapid-public-key", os.Getenv("SLEEPGUARD_VAPID_PUBLIC_KEY"), "Web Push VAPID public key")
	flag.StringVar(&cfg.VAPIDPrivateKey, "vapid-private-key", os.Getenv("SLEEPGUARD_VAPID_PRIVATE_KEY"), "Web Push VAPID private key")
	flag.StringVar(&cfg.VAPIDSubject, "vapid-subject", envOr("SLEEPGUARD_VAPID_SUBJECT", "mailto:sleepguard@rksdevs.in"), "Web Push VAPID subject (mailto: or https:)")
	flag.IntVar(&cfg.RuleNotifyCycles, "rule-notify-cycles", envInt("SLEEPGUARD_RULE_NOTIFY_CYCLES", 3), "push alert every N rise→fall cycles")
	flag.DurationVar(&cfg.RuleIdleReset, "rule-idle-reset", envDuration("SLEEPGUARD_RULE_IDLE_RESET", 10*time.Minute), "reset cycle count after idle period")

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

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n <= 0 {
		return fallback
	}
	return n
}
