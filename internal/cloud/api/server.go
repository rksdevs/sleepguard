package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/rksdevs/sleepguard/internal/cloud/auth"
	cloudcfg "github.com/rksdevs/sleepguard/internal/cloud/config"
	"github.com/rksdevs/sleepguard/internal/cloud/store"
	"github.com/rksdevs/sleepguard/internal/domain"
)

// Server serves the SleepGuard cloud HTTP API.
type Server struct {
	cfg    cloudcfg.Config
	store  *store.Postgres
	log    *slog.Logger
	mux    *http.ServeMux
	server *http.Server
}

// New creates a configured API server.
func New(cfg cloudcfg.Config, st *store.Postgres, log *slog.Logger) *Server {
	s := &Server{
		cfg:   cfg,
		store: st,
		log:   log,
		mux:   http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /api/v1/events", s.handleIngestEvent)
	s.mux.HandleFunc("POST /api/v1/heartbeat", s.handleHeartbeat)
	s.mux.HandleFunc("GET /api/v1/events", s.handleListEvents)
	s.mux.HandleFunc("GET /api/v1/devices/{id}/status", s.handleDeviceStatus)
}

// ListenAndServe starts the HTTP server until ctx is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    s.cfg.HTTPAddr,
		Handler: s.withCORS(s.mux),
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("cloud api listening", "addr", s.cfg.HTTPAddr)
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"service": "sleepguard-cloud",
	})
}

func (s *Server) handleIngestEvent(w http.ResponseWriter, r *http.Request) {
	deviceID, err := s.authenticateDevice(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body domain.IngestEvent
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if body.Type == "" || body.State == "" {
		writeError(w, http.StatusBadRequest, "type and state are required")
		return
	}
	if body.Source == "" {
		body.Source = deviceID
	}

	event, err := s.store.InsertEvent(r.Context(), deviceID, body)
	if err != nil {
		s.log.Error("insert event failed", "error", err, "device_id", deviceID)
		writeError(w, http.StatusInternalServerError, "failed to store event")
		return
	}

	_ = s.store.TouchHeartbeat(r.Context(), deviceID, time.Now().UTC())

	writeJSON(w, http.StatusCreated, map[string]any{
		"event": event,
	})
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	deviceID, err := s.authenticateDevice(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	at := time.Now().UTC()
	if err := s.store.TouchHeartbeat(r.Context(), deviceID, at); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "heartbeat failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"device_id": deviceID,
		"last_seen_at": at,
	})
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	deviceID, err := s.authenticateRead(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	queryDevice := r.URL.Query().Get("device_id")
	if queryDevice != "" {
		deviceID = queryDevice
	}
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "device_id is required")
		return
	}

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	events, err := s.store.ListEvents(r.Context(), deviceID, limit)
	if err != nil {
		s.log.Error("list events failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list events")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"device_id": deviceID,
		"count":     len(events),
		"events":    events,
	})
}

func (s *Server) handleDeviceStatus(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticateRead(r); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	deviceID := r.PathValue("id")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "device id is required")
		return
	}

	status, err := s.store.DeviceStatus(r.Context(), deviceID, s.cfg.OnlineAfter)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load status")
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (s *Server) authenticateDevice(r *http.Request) (string, error) {
	token, err := auth.BearerToken(r.Header.Get("Authorization"))
	if err != nil {
		return "", err
	}
	return s.store.DeviceByToken(r.Context(), token)
}

func (s *Server) authenticateRead(r *http.Request) (string, error) {
	token, err := auth.BearerToken(r.Header.Get("Authorization"))
	if err != nil {
		return "", err
	}

	if s.cfg.ReadAPIKey != "" && token == s.cfg.ReadAPIKey {
		return "", nil
	}

	return s.store.DeviceByToken(r.Context(), token)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
