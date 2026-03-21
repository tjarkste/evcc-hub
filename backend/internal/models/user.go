package models

import "time"

// User represents a registered user in the database.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	MQTTUsername string    `json:"mqttUsername"`
	MQTTPassword string    `json:"mqttPassword"`
	TopicPrefix  string    `json:"topicPrefix"`
	CreatedAt    time.Time `json:"createdAt"`
}

// RegisterRequest is the JSON body for POST /api/auth/register.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the JSON body for POST /api/auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is the common response for register and login.
type AuthResponse struct {
	Token        string `json:"token"`
	MQTTUsername string `json:"mqttUsername"`
	MQTTPassword string `json:"mqttPassword"`
	UserID       string `json:"userId"`
	DefaultSite  *Site  `json:"defaultSite,omitempty"`
}

// MQTTAuthRequest is the body sent by the Mosquitto auth plugin.
type MQTTAuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// MQTTACLRequest is the body sent by the Mosquitto ACL plugin.
type MQTTACLRequest struct {
	Username string `json:"username" binding:"required"`
	Topic    string `json:"topic" binding:"required"`
	Acc      int    `json:"acc" binding:"required"`
}
