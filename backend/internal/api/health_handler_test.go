package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
)

func TestHealthHandler_Healthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := storage.Open(t.TempDir() + "/test.db")
	defer db.Close()

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
	db, _ := storage.Open(t.TempDir() + "/test.db")
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
