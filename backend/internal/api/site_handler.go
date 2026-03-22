package api

import (
	"net/http"

	"evcc-cloud/backend/internal/models"
	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
)

type siteHandler struct {
	db *storage.DB
}

const maxSitesPerUser = 10

// CreateSite handles POST /api/sites.
func (h *siteHandler) CreateSite(c *gin.Context) {
	var req models.CreateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiError(c, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	userID := c.GetString("userID")

	count, err := h.db.CountSitesByUserID(userID)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "site_creation_failed", err.Error())
		return
	}
	if count >= maxSitesPerUser {
		apiError(c, http.StatusConflict, "site_limit_reached", "maximum 10 sites allowed")
		return
	}

	site, err := h.db.CreateSite(userID, req.Name, req.Timezone)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "site_creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, models.SiteResponse{Site: *site})
}

// ListSites handles GET /api/sites.
func (h *siteHandler) ListSites(c *gin.Context) {
	userID := c.GetString("userID")
	sites, err := h.db.GetSitesByUserID(userID)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "site_fetch_failed", err.Error())
		return
	}
	if sites == nil {
		sites = []models.Site{}
	}
	// Strip MQTT passwords from list response
	for i := range sites {
		sites[i].MQTTPassword = ""
	}
	c.JSON(http.StatusOK, models.SiteListResponse{Sites: sites})
}

// UpdateSite handles PUT /api/sites/:id.
func (h *siteHandler) UpdateSite(c *gin.Context) {
	siteID := c.Param("id")
	userID := c.GetString("userID")

	var req models.UpdateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiError(c, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	site, err := h.db.UpdateSite(siteID, userID, req.Name, req.Timezone)
	if err != nil {
		if err.Error() == "site not found" {
			apiError(c, http.StatusNotFound, "site_not_found", "site not found or not owned by user")
			return
		}
		apiError(c, http.StatusInternalServerError, "site_update_failed", err.Error())
		return
	}

	site.MQTTPassword = ""
	c.JSON(http.StatusOK, models.SiteResponse{Site: *site})
}

// DeleteSite handles DELETE /api/sites/:id.
func (h *siteHandler) DeleteSite(c *gin.Context) {
	siteID := c.Param("id")
	userID := c.GetString("userID")

	if err := h.db.DeleteSite(siteID, userID); err != nil {
		if err.Error() == "site not found" {
			apiError(c, http.StatusNotFound, "site_not_found", "site not found or not owned by user")
			return
		}
		apiError(c, http.StatusInternalServerError, "site_delete_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
