package storage

import (
	"strings"
	"testing"
	"unicode"
)

func TestGenerateMQTTUsername_Format(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	username := GenerateMQTTUsername(userID)

	if !strings.HasPrefix(username, "user_") {
		t.Errorf("expected username to start with 'user_', got %q", username)
	}

	suffix := strings.TrimPrefix(username, "user_")
	if len(suffix) != 16 {
		t.Errorf("expected 16-char suffix, got %d chars: %q", len(suffix), suffix)
	}

	// The suffix must not contain dashes (UUID dashes stripped).
	if strings.Contains(suffix, "-") {
		t.Errorf("expected no dashes in suffix, got %q", suffix)
	}
}

func TestGenerateMQTTUsername_Deterministic(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	u1 := GenerateMQTTUsername(userID)
	u2 := GenerateMQTTUsername(userID)
	if u1 != u2 {
		t.Errorf("expected deterministic result, got %q and %q", u1, u2)
	}
}

func TestGenerateRandomPassword_Length(t *testing.T) {
	for _, n := range []int{16, 24, 32} {
		pwd, err := GenerateRandomPassword(n)
		if err != nil {
			t.Fatalf("unexpected error for n=%d: %v", n, err)
		}
		if len(pwd) != n {
			t.Errorf("expected length %d, got %d", n, len(pwd))
		}
	}
}

func TestGenerateRandomPassword_Alphanumeric(t *testing.T) {
	pwd, err := GenerateRandomPassword(64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range pwd {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			t.Errorf("non-alphanumeric character %q found in password", c)
		}
	}
}

func TestGenerateRandomPassword_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 20; i++ {
		pwd, err := GenerateRandomPassword(24)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[pwd] {
			t.Errorf("duplicate password generated: %q", pwd)
		}
		seen[pwd] = true
	}
}
