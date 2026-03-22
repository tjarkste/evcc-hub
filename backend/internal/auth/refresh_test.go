package auth

import (
	"testing"
)

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) < 32 {
		t.Errorf("token too short: got %d chars", len(token))
	}
}

func TestGenerateRefreshToken_Unique(t *testing.T) {
	t1, _ := GenerateRefreshToken()
	t2, _ := GenerateRefreshToken()
	if t1 == t2 {
		t.Error("two generated tokens should not be identical")
	}
}

func TestHashRefreshToken(t *testing.T) {
	token := "test-refresh-token-abc123"
	hash := HashRefreshToken(token)
	if hash == "" {
		t.Error("hash should not be empty")
	}
	if hash == token {
		t.Error("hash should not equal the raw token")
	}
}

func TestHashRefreshToken_Deterministic(t *testing.T) {
	token := "test-refresh-token-abc123"
	h1 := HashRefreshToken(token)
	h2 := HashRefreshToken(token)
	if h1 != h2 {
		t.Error("same token should produce same hash")
	}
}
