package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	agentcfg "github.com/rksdevs/sleepguard/internal/agent/config"
	"github.com/rksdevs/sleepguard/internal/agent/queue"
	"github.com/rksdevs/sleepguard/internal/agent/upload"
	"github.com/rksdevs/sleepguard/internal/domain"
	"github.com/rksdevs/sleepguard/internal/sensor"
)

func main() {
	cfg, err := agentcfg.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	log := agentcfg.NewLogger(cfg.Debug)
	log.Info("SleepGuard agent starting", "phase", "C")
	log.Info("configuration",
		"cloud_url", cfg.CloudURL,
		"device", cfg.DeviceName,
		"gpio_pin", cfg.GPIOPin,
		"queue_path", cfg.QueuePath,
		"heartbeat_interval", cfg.HeartbeatInterval.String(),
		"mock_sensor", cfg.MockSensor,
	)

	q, err := queue.NewFile(cfg.QueuePath)
	if err != nil {
		log.Error("failed to open queue", "error", err)
		os.Exit(1)
	}

	client := upload.NewClient(cfg.CloudURL, cfg.DeviceToken, cfg.UploadTimeout, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	post := func(event domain.IngestEvent) error {
		return client.PostEvent(ctx, event)
	}

	flushQueue := func() {
		sent, err := q.Flush(post)
		if err != nil {
			log.Error("queue flush failed", "error", err)
			return
		}
		if sent > 0 {
			log.Info("queued events uploaded", "count", sent)
		}
	}

	flushQueue()

	handleEvent := func(event sensor.Event) {
		ingest := upload.IngestFromSensor(event)
		if err := client.PostEvent(ctx, ingest); err != nil {
			log.Warn("cloud upload failed, queuing event", "error", err, "pattern", event.Pattern)
			if qerr := q.Append(ingest); qerr != nil {
				log.Error("failed to queue event", "error", qerr)
			}
			return
		}
		log.Debug("event uploaded", "pattern", event.Pattern)
	}

	reader, err := sensor.Open(cfg.SensorConfig(), log, handleEvent)
	if err != nil {
		log.Error("failed to open sensor", "error", err)
		os.Exit(1)
	}
	defer reader.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(cfg.HeartbeatInterval)
		defer ticker.Stop()

		sendHeartbeat := func() {
			if err := client.Heartbeat(ctx); err != nil {
				log.Warn("heartbeat failed", "error", err)
				return
			}
			log.Debug("heartbeat sent")
		}

		sendHeartbeat()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sendHeartbeat()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				flushQueue()
			}
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
	flushQueue()
	log.Info("SleepGuard agent stopped")
}
