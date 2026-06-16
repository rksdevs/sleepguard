package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/rksdevs/sleepguard/internal/domain"
	"github.com/rksdevs/sleepguard/internal/sensor"
)

// Client posts events and heartbeats to the cloud API.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
	log     *slog.Logger
}

// NewClient creates a cloud upload client.
func NewClient(baseURL, token string, timeout time.Duration, log *slog.Logger) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http: &http.Client{
			Timeout: timeout,
		},
		log: log,
	}
}

// PostEvent sends one motion event to the cloud.
func (c *Client) PostEvent(ctx context.Context, event domain.IngestEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/events", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SleepGuard-Agent/1.0")

	return c.do(req)
}

// Heartbeat updates device last_seen on the cloud.
func (c *Client) Heartbeat(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/heartbeat", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "SleepGuard-Agent/1.0")

	return c.do(req)
}

func (c *Client) do(req *http.Request) error {
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	b, _ := io.ReadAll(io.LimitReader(res.Body, 512))
	return fmt.Errorf("cloud returned %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
}

// IngestFromSensor converts a sensor event for cloud ingest.
func IngestFromSensor(event sensor.Event) domain.IngestEvent {
	return domain.IngestEvent{
		Timestamp: event.Timestamp,
		Type:      event.Type,
		Source:    event.Source,
		State:     event.State,
		Pattern:   event.Pattern,
	}
}
