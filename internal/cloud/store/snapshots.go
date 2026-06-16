package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/rksdevs/sleepguard/internal/domain"
)

// InsertSnapshot records a snapshot metadata row with the given id and file name.
func (s *Postgres) InsertSnapshot(ctx context.Context, id, deviceID, fileName string, sizeBytes int64, capturedAt time.Time) (domain.Snapshot, error) {
	if capturedAt.IsZero() {
		capturedAt = time.Now().UTC()
	}

	var snap domain.Snapshot
	err := s.pool.QueryRow(ctx, `
		INSERT INTO snapshots (id, device_id, captured_at, file_name, size_bytes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, device_id, captured_at, size_bytes
	`, id, deviceID, capturedAt.UTC(), fileName, sizeBytes).Scan(
		&snap.ID, &snap.DeviceID, &snap.CapturedAt, &snap.SizeBytes,
	)
	if err != nil {
		return domain.Snapshot{}, fmt.Errorf("insert snapshot: %w", err)
	}
	return snap, nil
}

// ListSnapshots returns recent snapshots for a device, newest first.
func (s *Postgres) ListSnapshots(ctx context.Context, deviceID string, limit int) ([]domain.Snapshot, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, device_id, captured_at, size_bytes
		FROM snapshots
		WHERE device_id = $1
		ORDER BY captured_at DESC, id DESC
		LIMIT $2
	`, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	defer rows.Close()

	var out []domain.Snapshot
	for rows.Next() {
		var snap domain.Snapshot
		if err := rows.Scan(&snap.ID, &snap.DeviceID, &snap.CapturedAt, &snap.SizeBytes); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		out = append(out, snap)
	}
	return out, rows.Err()
}

// SnapshotFile returns the on-disk file name for a snapshot id.
func (s *Postgres) SnapshotFile(ctx context.Context, snapshotID string) (deviceID, fileName string, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT device_id, file_name FROM snapshots WHERE id = $1
	`, snapshotID).Scan(&deviceID, &fileName)
	if err == pgx.ErrNoRows {
		return "", "", ErrNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("snapshot file: %w", err)
	}
	return deviceID, fileName, nil
}

// DeleteSnapshotsOlderThan removes snapshot rows captured before cutoff.
func (s *Postgres) DeleteSnapshotsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM snapshots WHERE captured_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete snapshots: %w", err)
	}
	return tag.RowsAffected(), nil
}
