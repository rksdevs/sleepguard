package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rksdevs/sleepguard/internal/cloud/auth"
	"github.com/rksdevs/sleepguard/internal/domain"
)

// ErrNotFound is returned when a device does not exist.
var ErrNotFound = errors.New("not found")

// ErrDuplicateDevice is returned when creating a device that already exists.
var ErrDuplicateDevice = errors.New("device already exists")

// Postgres persists devices and events.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres connects to the database and returns a store.
func NewPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

// Close releases database connections.
func (s *Postgres) Close() {
	s.pool.Close()
}

// Pool exposes the underlying pool for migrations.
func (s *Postgres) Pool() *pgxpool.Pool {
	return s.pool
}

// CreateDevice inserts a new device with a hashed API token.
func (s *Postgres) CreateDevice(ctx context.Context, id, name, rawToken string) error {
	hash := auth.HashToken(rawToken)
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO devices (id, name, api_key_hash)
		VALUES ($1, $2, $3)
	`, id, name, hash)
	if err != nil {
		return fmt.Errorf("insert device: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDuplicateDevice
	}
	return nil
}

// EnsureDevice creates the device when it does not exist.
func (s *Postgres) EnsureDevice(ctx context.Context, id, name, rawToken string) error {
	var exists bool
	if err := s.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM devices WHERE id = $1)`,
		id,
	).Scan(&exists); err != nil {
		return fmt.Errorf("check device: %w", err)
	}
	if exists {
		return nil
	}
	return s.CreateDevice(ctx, id, name, rawToken)
}

// DeviceByToken returns the device id for a valid raw API token.
func (s *Postgres) DeviceByToken(ctx context.Context, rawToken string) (string, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, api_key_hash FROM devices`)
	if err != nil {
		return "", fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, hash string
		if err := rows.Scan(&id, &hash); err != nil {
			return "", fmt.Errorf("scan device: %w", err)
		}
		if auth.TokenMatches(rawToken, hash) {
			return id, nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate devices: %w", err)
	}
	return "", auth.ErrUnauthorized
}

// TouchHeartbeat updates last_seen_at for a device.
func (s *Postgres) TouchHeartbeat(ctx context.Context, deviceID string, at time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE devices SET last_seen_at = $2 WHERE id = $1
	`, deviceID, at)
	if err != nil {
		return fmt.Errorf("update heartbeat: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// InsertEvent stores a motion event for a device.
func (s *Postgres) InsertEvent(ctx context.Context, deviceID string, in domain.IngestEvent) (domain.Event, error) {
	if in.Timestamp.IsZero() {
		in.Timestamp = time.Now().UTC()
	}

	var out domain.Event
	err := s.pool.QueryRow(ctx, `
		INSERT INTO events (device_id, recorded_at, type, source, state, pattern)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, device_id, recorded_at, type, source, state, pattern, received_at
	`, deviceID, in.Timestamp.UTC(), in.Type, in.Source, in.State, in.Pattern).Scan(
		&out.ID,
		&out.DeviceID,
		&out.Timestamp,
		&out.Type,
		&out.Source,
		&out.State,
		&out.Pattern,
		&out.ReceivedAt,
	)
	if err != nil {
		return domain.Event{}, fmt.Errorf("insert event: %w", err)
	}
	return out, nil
}

// ListEvents returns recent events for a device, newest first.
func (s *Postgres) ListEvents(ctx context.Context, deviceID string, limit int) ([]domain.Event, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, device_id, recorded_at, type, source, state, pattern, received_at
		FROM events
		WHERE device_id = $1
		ORDER BY recorded_at DESC, id DESC
		LIMIT $2
	`, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := make([]domain.Event, 0, limit)
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(
			&event.ID,
			&event.DeviceID,
			&event.Timestamp,
			&event.Type,
			&event.Source,
			&event.State,
			&event.Pattern,
			&event.ReceivedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return events, nil
}

// DeviceStatus returns operational metadata for a device.
func (s *Postgres) DeviceStatus(ctx context.Context, deviceID string, onlineAfter time.Duration) (domain.DeviceStatus, error) {
	var status domain.DeviceStatus
	var lastSeen *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT d.id, d.name, d.last_seen_at,
		       (SELECT COUNT(*)::bigint FROM events e WHERE e.device_id = d.id)
		FROM devices d
		WHERE d.id = $1
	`, deviceID).Scan(&status.ID, &status.Name, &lastSeen, &status.EventCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DeviceStatus{}, ErrNotFound
	}
	if err != nil {
		return domain.DeviceStatus{}, fmt.Errorf("device status: %w", err)
	}

	status.LastSeenAt = lastSeen
	status.OnlineAfter = onlineAfter
	if lastSeen != nil && time.Since(*lastSeen) <= onlineAfter {
		status.Online = true
	}
	return status, nil
}
