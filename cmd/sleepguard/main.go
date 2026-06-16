package main

import (
	"log/slog"
	"os"

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
	log.Info("SleepGuard starting", "phase", 0)
	log.Info("configuration loaded", cfg.LogAttrs()...)

	sample := sensor.NewEvent(sensor.EventMotion, cfg.DeviceName, sensor.StateActive)
	jsonEvent, err := sample.JSON()
	if err != nil {
		log.Error("failed to serialize sample event", "error", err)
		os.Exit(1)
	}

	log.Info("sample event", "summary", sample.String(), "json", string(jsonEvent))
	log.Info("SleepGuard phase 0 bootstrap complete")
}
