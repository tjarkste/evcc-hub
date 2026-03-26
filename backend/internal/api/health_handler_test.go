package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
)

func setupHealthTestDB(t *testing.T) *storage.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://evcc:evcc@localhost:5432/evcc_hub_test?sslmode=disable"
	}
	db, err := storage.Open(databaseURL)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.TruncateAll()
	t.Cleanup(func() {
		db.TruncateAll()
		db.Close()
	})
	return db
}

func TestHealthHandler_Healthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHealthTestDB(t)

	h := &healthHandler{db: db}
	r := gin.New()
	r.GET("/health", h.Health)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	if body["status"] != "ok" {
		t.Errorf("status = %v, want ok", body["status"])
	}
	checks := body["checks"].(map[string]interface{})
	if checks["database"] != "ok" {
		t.Errorf("database = %v, want ok", checks["database"])
	}
}

func TestHealthHandler_DBClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://evcc:evcc@localhost:5432/evcc_hub_test?sslmode=disable"
	}
	db, err := storage.Open(databaseURL)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Close()

	h := &healthHandler{db: db}
	r := gin.New()
	r.GET("/health", h.Health)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	if body["status"] != "degraded" {
		t.Errorf("status = %v, want degraded", body["status"])
	}
}
