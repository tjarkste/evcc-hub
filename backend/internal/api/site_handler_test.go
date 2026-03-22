package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"evcc-cloud/backend/internal/auth"
	"evcc-cloud/backend/internal/models"
	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *storage.DB, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	path := t.TempDir() + "/test.db"
	db, err := storage.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(path)
	})
	secret := "test-secret"
	router := NewRouter(db, Config{JWTSecret: secret, DevMode: true})
	return router, db, secret
}

func createTestUser(t *testing.T, db *storage.DB, secret string) (string, string) {
	t.Helper()
	user, err := db.CreateUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, err := auth.GenerateToken(user.ID, user.Email, secret)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return user.ID, token
}

func TestCreateSiteEndpoint(t *testing.T) {
	router, db, secret := setupTestRouter(t)
	_, token := createTestUser(t, db, secret)

	body, _ := json.Marshal(models.CreateSiteRequest{Name: "Vacation House"})
	req := httptest.NewRequest(http.MethodPost, "/api/sites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.SiteResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Site.Name != "Vacation House" {
		t.Errorf("expected 'Vacation House', got %q", resp.Site.Name)
	}
	if resp.Site.MQTTUsername == "" || resp.Site.MQTTPassword == "" {
		t.Error("expected MQTT credentials in create response")
	}
}

func TestListSitesEndpoint(t *testing.T) {
	router, db, secret := setupTestRouter(t)
	_, token := createTestUser(t, db, secret)

	req := httptest.NewRequest(http.MethodGet, "/api/sites", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.SiteListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Sites) < 1 {
		t.Fatal("expected at least 1 site")
	}
	for _, s := range resp.Sites {
		if s.MQTTPassword != "" {
			t.Error("expected MQTTPassword to be stripped in list response")
		}
	}
}

func TestUpdateSiteEndpoint(t *testing.T) {
	router, db, secret := setupTestRouter(t)
	userID, token := createTestUser(t, db, secret)

	site, _ := db.CreateSite(userID, "Old Name", nil)

	newName := "New Name"
	body, _ := json.Marshal(models.UpdateSiteRequest{Name: &newName})
	req := httptest.NewRequest(http.MethodPut, "/api/sites/"+site.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.SiteResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Site.Name != "New Name" {
		t.Errorf("expected 'New Name', got %q", resp.Site.Name)
	}
}

func TestSitesRequireAuth(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sites", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w.Code)
	}
}

func TestCreateSite_MaxSitesLimit(t *testing.T) {
	r, db, jwtSecret := setupTestRouter(t)
	defer db.Close()

	user, _ := db.CreateUser("limit@test.de", "testpass123")
	token, _ := auth.GenerateToken(user.ID, user.Email, jwtSecret)

	// There's already 1 default site; create 9 more to hit the limit of 10
	for i := 0; i < 9; i++ {
		db.CreateSite(user.ID, fmt.Sprintf("Site %d", i), nil)
	}

	// 11th site creation attempt should fail
	body, _ := json.Marshal(map[string]string{"name": "One Too Many"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("got %d, want 409 for exceeding site limit", w.Code)
	}
}

func TestDeleteSiteEndpoint(t *testing.T) {
	router, db, secret := setupTestRouter(t)
	userID, token := createTestUser(t, db, secret)

	site, _ := db.CreateSite(userID, "To Delete", nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/sites/"+site.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}
