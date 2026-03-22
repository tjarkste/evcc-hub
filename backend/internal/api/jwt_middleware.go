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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("userID", claims.Subject)
		c.Set("email", claims.Email)
		c.Next()
	}
}
