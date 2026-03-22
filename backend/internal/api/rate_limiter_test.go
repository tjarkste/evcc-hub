package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_AllowsNormalTraffic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimiter(5, 5))
	r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: got %d, want 200", i, w.Code)
		}
	}
}

func TestRateLimiter_BlocksExcessiveTraffic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimiter(1, 1))
	r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	// First request should pass
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first request: got %d, want 200", w.Code)
	}

	// Immediate second request should be rate-limited
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/test", nil)
	req2.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: got %d, want 429", w2.Code)
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimiter(1, 1))
	r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Exhaust first IP
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/test", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w1, req1)

	// Second IP should still be allowed
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/test", nil)
	req2.RemoteAddr = "5.6.7.8:1234"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("different IP: got %d, want 200", w2.Code)
	}
}
