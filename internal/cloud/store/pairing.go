package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/rksdevs/sleepguard/internal/domain"
)

// CreatePairing stores a push subscription for a device.
func (s *Postgres) CreatePairing(ctx context.Context, deviceID, name string, sub domain.PushSubscription) (domain.PairedClient, error) {
	id, err := randomID()
	if err != nil {
		return domain.PairedClient{}, err
	}

	var client domain.PairedClient
	err = s.pool.QueryRow(ctx, `
		INSERT INTO paired_clients (id, device_id, name, push_endpoint, push_p256dh, push_auth)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, device_id, name, notify_on_rise, created_at
	`, id, deviceID, name, sub.Endpoint, sub.Keys.P256dh, sub.Keys.Auth).Scan(
		&client.ID,
		&client.DeviceID,
		&client.Name,
		&client.NotifyOnRise,
		&client.CreatedAt,
	)
	if err != nil {
		return domain.PairedClient{}, fmt.Errorf("insert pairing: %w", err)
	}
	return client, nil
}

// ListPairings returns active pairings for a device.
func (s *Postgres) ListPairings(ctx context.Context, deviceID string) ([]domain.PairedClient, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, device_id, name, notify_on_rise, created_at
		FROM paired_clients
		WHERE device_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("list pairings: %w", err)
	}
	defer rows.Close()

	var out []domain.PairedClient
	for rows.Next() {
		var c domain.PairedClient
		if err := rows.Scan(&c.ID, &c.DeviceID, &c.Name, &c.NotifyOnRise, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan pairing: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// RevokePairing marks a client as unpaired.
func (s *Postgres) RevokePairing(ctx context.Context, pairingID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE paired_clients SET revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL
	`, pairingID)
	if err != nil {
		return fmt.Errorf("revoke pairing: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// PushTarget is a subscription used to send notifications.
type PushTarget struct {
	ID       string
	DeviceID string
	Name     string
	Endpoint string
	P256dh   string
	Auth     string
}

// ListPushTargets returns active subscriptions for a device with notify_on_rise.
func (s *Postgres) ListPushTargets(ctx context.Context, deviceID string) ([]PushTarget, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, device_id, name, push_endpoint, push_p256dh, push_auth
		FROM paired_clients
		WHERE device_id = $1 AND revoked_at IS NULL AND notify_on_rise = TRUE
	`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("list push targets: %w", err)
	}
	defer rows.Close()

	var out []PushTarget
	for rows.Next() {
		var t PushTarget
		if err := rows.Scan(&t.ID, &t.DeviceID, &t.Name, &t.Endpoint, &t.P256dh, &t.Auth); err != nil {
			return nil, fmt.Errorf("scan push target: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// RevokePairingByEndpoint disables a subscription that returned 410 Gone.
func (s *Postgres) RevokePairingByEndpoint(ctx context.Context, endpoint string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE paired_clients SET revoked_at = now()
		WHERE push_endpoint = $1 AND revoked_at IS NULL
	`, endpoint)
	return err
}

// GetPairing returns one pairing if it exists.
func (s *Postgres) GetPairing(ctx context.Context, pairingID string) (domain.PairedClient, error) {
	var c domain.PairedClient
	var revoked *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT id, device_id, name, notify_on_rise, created_at, revoked_at
		FROM paired_clients WHERE id = $1
	`, pairingID).Scan(&c.ID, &c.DeviceID, &c.Name, &c.NotifyOnRise, &c.CreatedAt, &revoked)
	if err == pgx.ErrNoRows {
		return domain.PairedClient{}, ErrNotFound
	}
	if err != nil {
		return domain.PairedClient{}, err
	}
	c.RevokedAt = revoked
	return c, nil
}

func randomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
