package api

import "github.com/gin-gonic/gin"

// SecurityHeaders adds security-related HTTP headers to every response.
// Note: HSTS (Strict-Transport-Security) is set in nginx (TLS terminator), not here.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
