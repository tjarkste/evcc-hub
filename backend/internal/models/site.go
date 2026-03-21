package models

import "time"

// Site represents a single evcc instance belonging to a user.
type Site struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Name         string    `json:"name"`
	MQTTUsername string    `json:"mqttUsername"`
	MQTTPassword string    `json:"mqttPassword,omitempty"`
	TopicPrefix  string    `json:"topicPrefix"`
	Timezone     *string   `json:"timezone"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// CreateSiteRequest is the JSON body for POST /api/sites.
type CreateSiteRequest struct {
	Name     string  `json:"name" binding:"required,min=1,max=100"`
	Timezone *string `json:"timezone"`
}

// UpdateSiteRequest is the JSON body for PUT /api/sites/:id.
type UpdateSiteRequest struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=100"`
	Timezone *string `json:"timezone"`
}

// SiteResponse wraps a single site for API responses.
type SiteResponse struct {
	Site Site `json:"site"`
}

// SiteListResponse wraps a list of sites for API responses.
type SiteListResponse struct {
	Sites []Site `json:"sites"`
}
