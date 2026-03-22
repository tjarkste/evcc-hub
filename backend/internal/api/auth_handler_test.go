package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"evcc-cloud/backend/internal/auth"
	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
)

func setupAuthTestRouter(t *testing.T) (*gin.Engine, *storage.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := storage.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	cfg := Config{JWTSecret: "test-secret", DevMode: true}
	r := NewRouter(db, cfg)
	return r, db
}

func TestRegister_PasswordTooLong(t *testing.T) {
	r, db := setupAuthTestRouter(t)
	defer db.Close()

	longPass := make([]byte, 200)
	for i := range longPass {
		longPass[i] = 'a'
	}
	body, _ := json.Marshal(map[string]string{
		"email":    "long@test.de",
		"password": string(longPass),
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400 for 200-char password", w.Code)
	}
}

func TestRefreshEndpoint_Success(t *testing.T) {
	r, db := setupAuthTestRouter(t)
	defer db.Close()

	user, _ := db.CreateUser("test@test.de", "testpass123")
	rawToken, _ := auth.GenerateRefreshToken()
	tokenHash := auth.HashRefreshToken(rawToken)
	db.CreateRefreshToken(user.ID, tokenHash)

	body, _ := json.Marshal(map[string]string{"refreshToken": rawToken})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected new access token")
	}
	if resp["refreshToken"] == nil || resp["refreshToken"] == "" {
		t.Error("expected new refresh token")
	}
	// Refresh should NOT return MQTT credentials
	if pw, ok := resp["mqttPassword"].(string); ok && pw != "" {
		t.Error("refresh should not return mqttPassword")
	}
}

func TestRefreshEndpoint_OldTokenRejected(t *testing.T) {
	r, db := setupAuthTestRouter(t)
	defer db.Close()

	user, _ := db.CreateUser("test2@test.de", "testpass123")
	rawToken, _ := auth.GenerateRefreshToken()
	tokenHash := auth.HashRefreshToken(rawToken)
	db.CreateRefreshToken(user.ID, tokenHash)

	body, _ := json.Marshal(map[string]string{"refreshToken": rawToken})

	// Use token once
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Use same token again — should fail
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w2.Code)
	}
}

func TestLogoutEndpoint_InvalidatesToken(t *testing.T) {
	r, db := setupAuthTestRouter(t)
	defer db.Close()

	user, _ := db.CreateUser("test3@test.de", "testpass123")
	rawToken, _ := auth.GenerateRefreshToken()
	tokenHash := auth.HashRefreshToken(rawToken)
	db.CreateRefreshToken(user.ID, tokenHash)

	body, _ := json.Marshal(map[string]string{"refreshToken": rawToken})

	// Logout
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("logout: got %d, want 200", w.Code)
	}

	// Refresh should now fail
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("refresh after logout: got %d, want 401", w2.Code)
	}
}

func TestRefreshEndpoint_InvalidToken(t *testing.T) {
	r, db := setupAuthTestRouter(t)
	defer db.Close()

	body, _ := json.Marshal(map[string]string{"refreshToken": "totally-bogus-token"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}
