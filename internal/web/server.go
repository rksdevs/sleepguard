package web

import (
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/rksdevs/sleepguard/internal/alert"
	"github.com/rksdevs/sleepguard/internal/config"
	"github.com/rksdevs/sleepguard/internal/store"
)

//go:embed templates/dashboard.html
var templateFS embed.FS

// Server exposes HTTP endpoints for health, status, events, and the dashboard.
type Server struct {
	cfg     config.Config
	store   *store.Memory
	alerts  *alert.Manager
	started time.Time
	tmpl    *template.Template
}

// New creates an HTTP server wired to the event store and alert manager.
func New(cfg config.Config, mem *store.Memory, alerts *alert.Manager) (*Server, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/dashboard.html")
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:     cfg,
		store:   mem,
		alerts:  alerts,
		started: time.Now(),
		tmpl:    tmpl,
	}, nil
}

// ListenAndServe starts the HTTP server and shuts down when ctx is cancelled.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /status", s.handleStatus)
	mux.HandleFunc("GET /events", s.handleEvents)
	mux.HandleFunc("GET /config", s.handleConfig)
	mux.HandleFunc("GET /", s.handleDashboard)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	motionCount, alertCount, uptime, last, hasLast := s.store.Stats()

	payload := map[string]any{
		"device":       s.cfg.DeviceName,
		"uptime":       uptime.String(),
		"motion_count": motionCount,
		"alert_count":  alertCount,
		"alert_state":  s.alerts.CurrentState(),
		"gpio_pin":     s.cfg.GPIOPin,
	}

	if hasLast {
		payload["last_event"] = last
	}

	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	limit := 50
	events := s.store.Recent(limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"count":  len(events),
		"events": events,
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"device":          s.cfg.DeviceName,
		"http_addr":       s.cfg.HTTPAddr,
		"alert_cooldown":  s.cfg.AlertCooldown.String(),
		"report_interval": s.cfg.ReportInterval.String(),
		"gpio_pin":        s.cfg.GPIOPin,
		"mock_sensor":     s.cfg.MockSensor,
	})
}

type dashboardEvent struct {
	Timestamp string
	State     string
	Pattern   string
	Source    string
}

func (s *Server) handleDashboard(w http.ResponseWriter, _ *http.Request) {
	events := s.store.Recent(25)
	viewEvents := make([]dashboardEvent, len(events))
	for i, event := range events {
		viewEvents[i] = dashboardEvent{
			Timestamp: event.Timestamp.Local().Format("15:04:05"),
			State:     event.State,
			Pattern:   event.Pattern,
			Source:    event.Source,
		}
	}

	motionCount, alertCount, uptime, _, _ := s.store.Stats()

	data := map[string]any{
		"Device":      s.cfg.DeviceName,
		"Uptime":      uptime.Round(time.Second).String(),
		"MotionCount": motionCount,
		"AlertCount":  alertCount,
		"EventCount":  len(events),
		"AlertState":  s.alerts.CurrentState(),
		"Events":      viewEvents,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
