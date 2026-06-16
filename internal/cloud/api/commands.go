package api

import (
	"net/http"

	"github.com/rksdevs/sleepguard/internal/domain"
)

func (s *Server) handleAgentCommands(w http.ResponseWriter, r *http.Request) {
	deviceID, err := s.authenticateDevice(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	capture := false
	if s.captures != nil {
		capture = s.captures.Take(deviceID)
	}

	writeJSON(w, http.StatusOK, domain.AgentCommands{
		CaptureSnapshot: capture,
	})
}
