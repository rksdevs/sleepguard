package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rksdevs/sleepguard/internal/cloud/api"
	"github.com/rksdevs/sleepguard/internal/cloud/cleanup"
	cloudcfg "github.com/rksdevs/sleepguard/internal/cloud/config"
	"github.com/rksdevs/sleepguard/internal/cloud/migrate"
	"github.com/rksdevs/sleepguard/internal/cloud/push"
	"github.com/rksdevs/sleepguard/internal/cloud/store"
	"github.com/rksdevs/sleepguard/internal/config"
)

func main() {
	cfg, err := cloudcfg.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	log := config.NewLogger(cfg.Debug)
	ctx := context.Background()

	st, err := store.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer st.Close()

	if err := migrate.Up(ctx, st.Pool()); err != nil {
		log.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	log.Info("database migrations applied")

	if cfg.CreateDevice {
		if err := st.CreateDevice(ctx, cfg.CreateDeviceID, cfg.CreateDeviceName, cfg.CreateDeviceToken); err != nil {
			log.Error("create device failed", "error", err)
			os.Exit(1)
		}
		log.Info("device created", "device_id", cfg.CreateDeviceID)
		return
	}

	if cfg.MigrateOnly {
		log.Info("migrate-only complete")
		return
	}

	if cfg.BootstrapDeviceToken != "" {
		if err := st.EnsureDevice(ctx, cfg.BootstrapDeviceID, cfg.BootstrapDeviceName, cfg.BootstrapDeviceToken); err != nil {
			log.Error("bootstrap device failed", "error", err)
			os.Exit(1)
		}
		log.Info("bootstrap device ensured", "device_id", cfg.BootstrapDeviceID)
	}

	if err := os.MkdirAll(cfg.SnapshotDir, 0o755); err != nil {
		log.Error("snapshot directory not writable", "dir", cfg.SnapshotDir, "error", err)
		os.Exit(1)
	}

	log.Info("SleepGuard cloud starting", "phase", "D")
	log.Info("configuration",
		"http_addr", cfg.HTTPAddr,
		"snapshot_dir", cfg.SnapshotDir,
		"online_after", cfg.OnlineAfter.String(),
		"read_api_key_set", cfg.ReadAPIKey != "",
		"event_retention", cfg.EventRetention.String(),
		"cleanup_interval", cfg.CleanupInterval.String(),
		"push_enabled", cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != "",
	)

	pusher := push.NewSender(cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject, st, log)
	cleaner := cleanup.New(st, cfg.SnapshotDir, cfg.EventRetention, log)

	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cleanup.StartScheduler(runCtx, cleaner, cfg.CleanupInterval, log)

	server := api.New(cfg, st, pusher, cleaner, log)

	if err := server.ListenAndServe(runCtx); err != nil {
		log.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
	log.Info("SleepGuard cloud stopped")
}
