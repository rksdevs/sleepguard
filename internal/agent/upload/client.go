package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

// Heartbeat updates device last_seen and returns optional commands from the cloud.
func (c *Client) Heartbeat(ctx context.Context) (domain.HeartbeatResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/heartbeat", nil)
	if err != nil {
		return domain.HeartbeatResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "SleepGuard-Agent/1.0")

	res, err := c.http.Do(req)
	if err != nil {
		return domain.HeartbeatResponse{}, err
	}
	defer res.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return domain.HeartbeatResponse{}, fmt.Errorf("cloud returned %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}

	var resp domain.HeartbeatResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return domain.HeartbeatResponse{}, fmt.Errorf("decode heartbeat response: %w", err)
	}
	return resp, nil
}

// PostSnapshot uploads a JPEG still image to the cloud.
func (c *Client) PostSnapshot(ctx context.Context, imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/snapshots", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
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
