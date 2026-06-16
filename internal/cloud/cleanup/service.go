package cleanup

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/rksdevs/sleepguard/internal/cloud/store"
)

// Result reports what was removed.
type Result struct {
	EventsDeleted   int64 `json:"events_deleted"`
	SnapshotsDeleted int  `json:"snapshots_deleted"`
	Cutoff          time.Time `json:"cutoff"`
}

// Service purges old events and snapshot files.
type Service struct {
	store        *store.Postgres
	snapshotDir  string
	retention    time.Duration
	log          *slog.Logger
}

// New creates a cleanup service.
func New(st *store.Postgres, snapshotDir string, retention time.Duration, log *slog.Logger) *Service {
	return &Service{
		store:       st,
		snapshotDir: snapshotDir,
		retention:   retention,
		log:         log,
	}
}

// Run deletes data older than the retention window. Devices and pairings are kept.
func (s *Service) Run(ctx context.Context) (Result, error) {
	cutoff := time.Now().UTC().Add(-s.retention)
	result := Result{Cutoff: cutoff}

	events, err := s.store.DeleteEventsOlderThan(ctx, cutoff)
	if err != nil {
		return result, err
	}
	result.EventsDeleted = events

	snaps, err := deleteOldSnapshots(s.snapshotDir, cutoff)
	if err != nil {
		return result, err
	}
	result.SnapshotsDeleted = snaps

	s.log.Info("cleanup complete",
		"events_deleted", result.EventsDeleted,
		"snapshots_deleted", result.SnapshotsDeleted,
		"cutoff", cutoff.Format(time.RFC3339),
	)
	return result, nil
}

func deleteOldSnapshots(dir string, cutoff time.Time) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	var deleted int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().UTC().Before(cutoff) {
			if err := os.Remove(path); err == nil {
				deleted++
			}
		}
	}
	return deleted, nil
}

// StartScheduler runs cleanup on an interval until ctx is cancelled.
func StartScheduler(ctx context.Context, svc *Service, interval time.Duration, log *slog.Logger) {
	if interval <= 0 {
		return
	}
	go func() {
		run := func() {
			if _, err := svc.Run(ctx); err != nil {
				log.Error("scheduled cleanup failed", "error", err)
			}
		}
		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
