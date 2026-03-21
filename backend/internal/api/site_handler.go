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

// CreateSite handles POST /api/sites.
func (h *siteHandler) CreateSite(c *gin.Context) {
	var req models.CreateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	site, err := h.db.CreateSite(userID, req.Name, req.Timezone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create site"})
		return
	}

	c.JSON(http.StatusCreated, models.SiteResponse{Site: *site})
}

// ListSites handles GET /api/sites.
func (h *siteHandler) ListSites(c *gin.Context) {
	userID := c.GetString("userID")
	sites, err := h.db.GetSitesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch sites"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	site, err := h.db.UpdateSite(siteID, userID, req.Name, req.Timezone)
	if err != nil {
		if err.Error() == "site not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update site"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete site"})
		return
	}

	c.Status(http.StatusNoContent)
}
