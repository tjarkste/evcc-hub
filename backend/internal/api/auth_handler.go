package api

import (
	"errors"
	"log"
	"net/http"

	"evcc-cloud/backend/internal/auth"
	"evcc-cloud/backend/internal/models"
	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type authHandler struct {
	db        *storage.DB
	jwtSecret string
}

// Register handles POST /api/auth/register.
func (h *authHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiError(c, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	user, err := h.db.CreateUser(req.Email, req.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			apiError(c, http.StatusConflict, "duplicate_email", "email already registered")
		} else {
			log.Printf("CreateUser error: %v", err)
			apiError(c, http.StatusInternalServerError, "registration_failed", "could not create account")
		}
		return
	}

	log.Printf("[NEWSIGNUP] new user registered (id=%s)", user.ID) // email omitted for PII

	token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "token_generation_failed", "could not generate token")
		return
	}

	rawRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		apiError(c, http.StatusInternalServerError, "refresh_token_failed", "could not generate refresh token")
		return
	}
	refreshHash := auth.HashRefreshToken(rawRefresh)
	if _, err := h.db.CreateRefreshToken(user.ID, refreshHash); err != nil {
		apiError(c, http.StatusInternalServerError, "refresh_token_store_failed", "could not store refresh token")
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
		RefreshToken: rawRefresh,
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
		apiError(c, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	user, err := h.db.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		apiError(c, http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "token_generation_failed", "could not generate token")
		return
	}

	rawRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		apiError(c, http.StatusInternalServerError, "refresh_token_failed", "could not generate refresh token")
		return
	}
	refreshHash := auth.HashRefreshToken(rawRefresh)
	if _, err := h.db.CreateRefreshToken(user.ID, refreshHash); err != nil {
		apiError(c, http.StatusInternalServerError, "refresh_token_store_failed", "could not store refresh token")
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token:        token,
		RefreshToken: rawRefresh,
		MQTTUsername: user.MQTTUsername,
		MQTTPassword: user.MQTTPassword,
		UserID:       user.ID,
	})
}

// Refresh handles POST /api/auth/refresh.
// Rotates the refresh token: invalidates the old one, issues a new access+refresh pair.
// Does NOT return MQTT credentials — the client already has them from login.
func (h *authHandler) Refresh(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiError(c, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	oldHash := auth.HashRefreshToken(req.RefreshToken)
	rt, err := h.db.GetRefreshTokenByHash(oldHash)
	if err != nil {
		apiError(c, http.StatusUnauthorized, "invalid_refresh_token", "invalid or expired refresh token")
		return
	}

	// Rotate: delete old token first
	_ = h.db.DeleteRefreshToken(oldHash)

	user, err := h.db.GetUserByID(rt.UserID)
	if err != nil {
		apiError(c, http.StatusUnauthorized, "user_not_found", "user not found")
		return
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "token_generation_failed", "could not generate token")
		return
	}

	newRawToken, err := auth.GenerateRefreshToken()
	if err != nil {
		apiError(c, http.StatusInternalServerError, "refresh_token_failed", "could not generate refresh token")
		return
	}
	newHash := auth.HashRefreshToken(newRawToken)
	if _, err := h.db.CreateRefreshToken(user.ID, newHash); err != nil {
		apiError(c, http.StatusInternalServerError, "refresh_token_store_failed", "could not store refresh token")
		return
	}

	c.JSON(http.StatusOK, models.RefreshResponse{
		Token:        accessToken,
		RefreshToken: newRawToken,
		UserID:       user.ID,
	})
}

// Logout handles POST /api/auth/logout.
// Invalidates the provided refresh token server-side.
func (h *authHandler) Logout(c *gin.Context) {
	var req models.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiError(c, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	tokenHash := auth.HashRefreshToken(req.RefreshToken)
	_ = h.db.DeleteRefreshToken(tokenHash)

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
