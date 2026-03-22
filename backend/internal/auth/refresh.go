package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateRefreshToken creates a cryptographically random 32-byte token encoded as hex (64 chars).
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashRefreshToken returns the SHA-256 hex digest of a refresh token.
// We store hashes, not raw tokens, so a DB leak doesn't compromise active sessions.
func HashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
