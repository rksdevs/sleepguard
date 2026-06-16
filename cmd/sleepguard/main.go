package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rksdevs/sleepguard/internal/alert"
	"github.com/rksdevs/sleepguard/internal/config"
	"github.com/rksdevs/sleepguard/internal/sensor"
	"github.com/rksdevs/sleepguard/internal/store"
	"github.com/rksdevs/sleepguard/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	log := config.NewLogger(cfg.Debug)
	log.Info("SleepGuard starting", "phase", 2)
	log.Info("configuration loaded", cfg.LogAttrs()...)

	mem := store.NewMemory(cfg.StoreCapacity)

	notifier := alertNotifier(cfg, log)
	alerts := alert.NewManager(notifier, cfg.AlertCooldown, log, func() {
		mem.RecordAlert()
	})

	handleEvent := func(event sensor.Event) {
		mem.Append(event)
		alerts.Handle(context.Background(), event)
	}

	reader, err := sensor.Open(cfg, log, handleEvent)
	if err != nil {
		log.Error("failed to open sensor", "error", err)
		os.Exit(1)
	}
	defer reader.Close()

	httpServer, err := web.New(cfg, mem, alerts)
	if err != nil {
		log.Error("failed to create web server", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("http server starting", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(ctx, cfg.HTTPAddr); err != nil {
			log.Error("http server stopped with error", "error", err)
			stop()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("sensor running")
		if err := reader.Run(ctx); err != nil {
			log.Error("sensor stopped with error", "error", err)
			stop()
		}
	}()

	wg.Wait()
	log.Info("SleepGuard stopped")
}

func alertNotifier(cfg config.Config, log *slog.Logger) alert.Notifier {
	if cfg.AlertCommand != "" {
		return alert.NewExecNotifier(cfg.AlertCommand, log)
	}
	return alert.NewLogNotifier(log)
}
