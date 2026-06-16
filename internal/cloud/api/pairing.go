package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rksdevs/sleepguard/internal/cloud/store"
	"github.com/rksdevs/sleepguard/internal/domain"
)

func (s *Server) handleVAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
	key := ""
	if s.push != nil {
		key = s.push.PublicKey()
	}
	if key == "" {
		writeError(w, http.StatusServiceUnavailable, "push notifications not configured")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"public_key": key})
}

func (s *Server) handlePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	s.handleCreatePairing(w, r)
}

func (s *Server) handleCreatePairing(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticateRead(r); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body domain.PairingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if body.DeviceID == "" || body.Subscription.Endpoint == "" {
		writeError(w, http.StatusBadRequest, "device_id and subscription are required")
		return
	}
	if body.Name == "" {
		body.Name = "Phone"
	}

	client, err := s.store.CreatePairing(r.Context(), body.DeviceID, body.Name, body.Subscription)
	if err != nil {
		s.log.Error("create pairing failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to pair")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"client": client})
}

func (s *Server) handleListPairings(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticateRead(r); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "device_id is required")
		return
	}

	clients, err := s.store.ListPairings(r.Context(), deviceID)
	if err != nil {
		s.log.Error("list pairings failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list pairings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"device_id": deviceID,
		"count":     len(clients),
		"clients":   clients,
	})
}

func (s *Server) handleUnpair(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticateRead(r); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	pairingID := r.PathValue("id")
	if pairingID == "" {
		writeError(w, http.StatusBadRequest, "pairing id is required")
		return
	}

	if err := s.store.RevokePairing(r.Context(), pairingID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "pairing not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to unpair")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}
