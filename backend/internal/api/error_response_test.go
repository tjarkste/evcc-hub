package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestApiError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		status  int
		code    string
		message string
	}{
		{"bad request", http.StatusBadRequest, "invalid_input", "email is required"},
		{"unauthorized", http.StatusUnauthorized, "invalid_credentials", "invalid credentials"},
		{"conflict", http.StatusConflict, "duplicate_email", "email already registered"},
		{"internal error", http.StatusInternalServerError, "token_generation_failed", "could not generate token"},
		{"rate limited", http.StatusTooManyRequests, "rate_limited", "too many requests"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			apiError(c, tt.status, tt.code, tt.message)

			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}

			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if body["error"] != tt.message {
				t.Errorf("error = %q, want %q", body["error"], tt.message)
			}
			if body["code"] != tt.code {
				t.Errorf("code = %q, want %q", body["code"], tt.code)
			}
		})
	}
}
