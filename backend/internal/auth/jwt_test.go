package auth

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key-for-unit-tests"

func TestGenerateAndValidateToken(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	email := "test@example.com"

	token, err := GenerateToken(userID, email, testSecret)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("expected subject %q, got %q", userID, claims.Subject)
	}
	if claims.Email != email {
		t.Errorf("expected email %q, got %q", email, claims.Email)
	}
}

func TestToken_ExpiresIn24Hours(t *testing.T) {
	token, err := GenerateToken("uid", "user@example.com", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	claims, err := ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}

	expiry := claims.ExpiresAt.Time
	issued := claims.IssuedAt.Time
	duration := expiry.Sub(issued)

	// Allow a small margin (1 second) for test execution time.
	if duration < 23*time.Hour+59*time.Minute {
		t.Errorf("expected ~24h expiry, got %v", duration)
	}
	if duration > 24*time.Hour+time.Second {
		t.Errorf("expected ~24h expiry, got %v", duration)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, err := GenerateToken("uid", "user@example.com", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	_, err = ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("expected error when validating with wrong secret")
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	_, err := ValidateToken("not.a.valid.jwt", testSecret)
	if err == nil {
		t.Error("expected error for malformed token")
	}
}

func TestValidateToken_Empty(t *testing.T) {
	_, err := ValidateToken("", testSecret)
	if err == nil {
		t.Error("expected error for empty token")
	}
}
