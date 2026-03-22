package api

import (
	"net/http"
	"strings"

	"evcc-cloud/backend/internal/auth"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware validates the Authorization Bearer token and sets userID + email in the context.
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			apiError(c, http.StatusUnauthorized, "missing_token", "authorization header required")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			apiError(c, http.StatusUnauthorized, "invalid_token", "invalid token")
			c.Abort()
			return
		}

		c.Set("userID", claims.Subject)
		c.Set("email", claims.Email)
		c.Next()
	}
}
