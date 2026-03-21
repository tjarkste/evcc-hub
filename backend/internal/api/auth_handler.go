package api

import (
	"net/http"

	"evcc-cloud/backend/internal/auth"
	"evcc-cloud/backend/internal/models"
	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
)

type authHandler struct {
	db        *storage.DB
	jwtSecret string
}

// Register handles POST /api/auth/register.
func (h *authHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.CreateUser(req.Email, req.Password)
	if err != nil {
		// Duplicate email results in a UNIQUE constraint error.
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	// Fetch default site created during registration
	sites, _ := h.db.GetSitesByUserID(user.ID)
	var defaultSite *models.Site
	if len(sites) > 0 {
		full, _ := h.db.GetSiteByMQTTUsername(sites[0].MQTTUsername)
		defaultSite = full
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token:        token,
		MQTTUsername: user.MQTTUsername,
		MQTTPassword: user.MQTTPassword,
		UserID:       user.ID,
		DefaultSite:  defaultSite,
	})
}

// Login handles POST /api/auth/login.
func (h *authHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token:        token,
		MQTTUsername: user.MQTTUsername,
		MQTTPassword: user.MQTTPassword,
		UserID:       user.ID,
	})
}
