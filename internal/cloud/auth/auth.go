package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
)

// ErrUnauthorized is returned when credentials are missing or invalid.
var ErrUnauthorized = errors.New("unauthorized")

// HashToken returns a hex-encoded SHA-256 hash of the raw API token.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// TokenMatches reports whether raw matches the stored hash.
func TokenMatches(raw, storedHash string) bool {
	got := HashToken(raw)
	return subtle.ConstantTimeCompare([]byte(got), []byte(storedHash)) == 1
}

// BearerToken extracts the token from an Authorization header.
func BearerToken(header string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", ErrUnauthorized
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", ErrUnauthorized
	}
	return token, nil
}
