package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"evcc-cloud/backend/internal/auth"

	"github.com/gin-gonic/gin"
)

func TestJWTMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	token, _ := auth.GenerateToken("user-123", "test@example.com", secret)

	r := gin.New()
	r.Use(JWTAuthMiddleware(secret))
	r.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("userID")
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuthMiddleware("test-secret"))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuthMiddleware("test-secret"))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
