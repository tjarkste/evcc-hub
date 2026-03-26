package api

import (
	"net/http"

	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type ProfileResponse struct {
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

func (h *authHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")
	user, err := h.db.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, ProfileResponse{
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *authHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("userID")
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), storage.BcryptCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.db.UpdateUserPassword(userID, string(newHash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// Invalidate all refresh tokens (force re-login on other devices)
	_ = h.db.DeleteRefreshTokensByUserID(userID)

	c.Status(http.StatusNoContent)
}
