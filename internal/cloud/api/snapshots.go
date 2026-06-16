package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rksdevs/sleepguard/internal/cloud/store"
)

func (s *Server) handleRequestCapture(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if !status.Online {
		writeError(w, http.StatusServiceUnavailable, "device is offline")
		return
	}

	s.captures.Request(deviceID)
	s.log.Info("capture requested", "device_id", deviceID)

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":    "queued",
		"device_id": deviceID,
		"message":   "capture will run on the next agent heartbeat",
	})
}

func (s *Server) handleListSnapshots(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticateRead(r); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "device_id is required")
		return
	}

	limit := 5
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	snaps, err := s.store.ListSnapshots(r.Context(), deviceID, limit)
	if err != nil {
		s.log.Error("list snapshots failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list snapshots")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"device_id":  deviceID,
		"count":      len(snaps),
		"snapshots":  snaps,
	})
}

func (s *Server) handleSnapshotImage(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticateRead(r); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	snapshotID := r.PathValue("id")
	deviceID, fileName, err := s.store.SnapshotFile(r.Context(), snapshotID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "snapshot not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load snapshot")
		return
	}

	path := filepath.Join(s.cfg.SnapshotDir, deviceID, fileName)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "image file missing")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to open image")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, max-age=60")
	_, _ = io.Copy(w, f)
}

func (s *Server) handleUploadSnapshot(w http.ResponseWriter, r *http.Request) {
	deviceID, err := s.authenticateDevice(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	const maxUpload = 12 << 20 // 12 MiB
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "image field is required")
		return
	}
	defer file.Close()

	if header.Size > maxUpload {
		writeError(w, http.StatusRequestEntityTooLarge, "image too large")
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" {
		ext = ".jpg"
	}

	snapID, err := randomHexID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create snapshot id")
		return
	}

	dir := filepath.Join(s.cfg.SnapshotDir, deviceID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		s.log.Error("snapshot mkdir failed", "error", err, "dir", dir)
		writeError(w, http.StatusInternalServerError, "failed to store image")
		return
	}

	fileName := snapID + ext
	destPath := filepath.Join(dir, fileName)

	out, err := os.Create(destPath)
	if err != nil {
		s.log.Error("snapshot create failed", "error", err, "path", destPath)
		writeError(w, http.StatusInternalServerError, "failed to store image")
		return
	}

	written, err := io.Copy(out, file)
	_ = out.Close()
	if err != nil {
		_ = os.Remove(destPath)
		writeError(w, http.StatusInternalServerError, "failed to write image")
		return
	}

	capturedAt := time.Now().UTC()
	if raw := r.FormValue("captured_at"); raw != "" {
		if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
			capturedAt = parsed.UTC()
		}
	}

	snap, err := s.store.InsertSnapshot(r.Context(), snapID, deviceID, fileName, written, capturedAt)
	if err != nil {
		_ = os.Remove(destPath)
		s.log.Error("insert snapshot failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to record snapshot")
		return
	}

	_ = s.store.TouchHeartbeat(r.Context(), deviceID, time.Now().UTC())

	writeJSON(w, http.StatusCreated, map[string]any{"snapshot": snap})
}

func randomHexID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
