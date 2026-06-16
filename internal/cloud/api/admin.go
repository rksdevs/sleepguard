package api

import (
	"net/http"
)

func (s *Server) handleAdminCleanup(w http.ResponseWriter, r *http.Request) {
	if err := s.authenticateAdmin(r); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.cleanup == nil {
		writeError(w, http.StatusServiceUnavailable, "cleanup not configured")
		return
	}

	result, err := s.cleanup.Run(r.Context())
	if err != nil {
		s.log.Error("cleanup failed", "error", err)
		writeError(w, http.StatusInternalServerError, "cleanup failed")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) authenticateAdmin(r *http.Request) error {
	token, err := authBearer(r)
	if err != nil {
		return err
	}
	if s.cfg.ReadAPIKey == "" || token != s.cfg.ReadAPIKey {
		return errUnauthorized
	}
	return nil
}
