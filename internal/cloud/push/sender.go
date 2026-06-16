package push

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/rksdevs/sleepguard/internal/cloud/store"
	"github.com/rksdevs/sleepguard/internal/domain"
)

// Sender delivers Web Push notifications to paired clients.
type Sender struct {
	publicKey  string
	privateKey string
	subject    string
	store      *store.Postgres
	log        *slog.Logger
}

// NewSender creates a push sender. Returns nil if VAPID keys are not configured.
func NewSender(publicKey, privateKey, subject string, st *store.Postgres, log *slog.Logger) *Sender {
	if publicKey == "" || privateKey == "" {
		return nil
	}
	if subject == "" {
		subject = "mailto:sleepguard@rksdevs.in"
	}
	return &Sender{
		publicKey:  publicKey,
		privateKey: privateKey,
		subject:    subject,
		store:      st,
		log:        log,
	}
}

// Enabled reports whether push is configured.
func (s *Sender) Enabled() bool {
	return s != nil
}

// PublicKey returns the VAPID public key for browser subscription.
func (s *Sender) PublicKey() string {
	if s == nil {
		return ""
	}
	return s.publicKey
}

// NotifyCycleAlert sends a push when sustained motion cycles hit the threshold.
func (s *Sender) NotifyCycleAlert(ctx context.Context, deviceID string, cycles int, event domain.Event) {
	if s == nil {
		return
	}
	targets, err := s.store.ListPushTargets(ctx, deviceID)
	if err != nil {
		s.log.Error("list push targets failed", "error", err)
		return
	}
	if len(targets) == 0 {
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"title": "SleepGuard — sustained motion",
		"body":  fmt.Sprintf("%d motion cycles — open the app to capture an image (%s)", cycles, event.Source),
		"url":   "/",
	})

	s.send(ctx, targets, payload)
}

// NotifyMotion sends a push for a single motion rise (legacy; prefer cycle rules).
func (s *Sender) NotifyMotion(ctx context.Context, deviceID string, event domain.Event) {
	if s == nil {
		return
	}
	targets, err := s.store.ListPushTargets(ctx, deviceID)
	if err != nil {
		s.log.Error("list push targets failed", "error", err)
		return
	}
	if len(targets) == 0 {
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"title": "SleepGuard — motion",
		"body":  fmt.Sprintf("%s: %s detected", event.Source, event.Pattern),
		"url":   "/",
	})

	s.send(ctx, targets, payload)
}

func (s *Sender) send(ctx context.Context, targets []store.PushTarget, payload []byte) {
	for _, target := range targets {
		sub := &webpush.Subscription{
			Endpoint: target.Endpoint,
			Keys: webpush.Keys{
				P256dh: target.P256dh,
				Auth:   target.Auth,
			},
		}
		resp, err := webpush.SendNotificationWithContext(ctx, payload, sub, &webpush.Options{
			Subscriber:      s.subject,
			VAPIDPublicKey:  s.publicKey,
			VAPIDPrivateKey: s.privateKey,
			TTL:             60,
			Urgency:         webpush.UrgencyHigh,
		})
		if err != nil {
			s.log.Warn("push send failed", "error", err, "client", target.Name)
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
			_ = s.store.RevokePairingByEndpoint(ctx, target.Endpoint)
		}
		if resp.StatusCode >= 300 {
			s.log.Warn("push rejected", "status", resp.StatusCode, "client", target.Name)
		}
	}
}
