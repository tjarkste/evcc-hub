package api

import "github.com/gin-gonic/gin"

// apiError sends a JSON error response with a machine-readable code and human-readable message.
func apiError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": message,
		"code":  code,
	})
}
