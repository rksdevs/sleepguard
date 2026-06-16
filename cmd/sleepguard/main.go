package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rksdevs/sleepguard/internal/config"
	"github.com/rksdevs/sleepguard/internal/sensor"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	log := config.NewLogger(cfg.Debug)
	log.Info("SleepGuard starting", "phase", 1)
	log.Info("configuration loaded", cfg.LogAttrs()...)

	reader, err := sensor.Open(cfg, log)
	if err != nil {
		log.Error("failed to open sensor", "error", err)
		os.Exit(1)
	}
	defer reader.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("sensor running — Ctrl+C to stop")

	if err := reader.Run(ctx); err != nil {
		log.Error("sensor stopped with error", "error", err)
		os.Exit(1)
	}

	log.Info("SleepGuard stopped")
}
