package api

import (
	"net/http"
	"strings"

	"evcc-cloud/backend/internal/models"
	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
)

type mqttAuthHandler struct {
	db *storage.DB
}

// MQTTAuth handles POST /api/mqtt/auth (Mosquitto HTTP auth plugin).
func (h *mqttAuthHandler) MQTTAuth(c *gin.Context) {
	var req models.MQTTAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.AuthenticateMQTT(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid mqtt credentials"})
		return
	}

	c.Status(http.StatusOK)
}

// MQTTACL handles POST /api/mqtt/acl (Mosquitto HTTP ACL plugin).
// NOTE: Mosquitto ACL requests have username + topic + acc but NO password.
// We use LookupMQTTCredentialByUsername instead of AuthenticateMQTT.
func (h *mqttAuthHandler) MQTTACL(c *gin.Context) {
	var req models.MQTTACLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResult, err := h.db.LookupMQTTCredentialByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "unknown user"})
		return
	}

	if !CheckACL(authResult.CredType, authResult.TopicPrefix, authResult.UserID, req.Topic, req.Acc) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.Status(http.StatusOK)
}

// CheckACL decides whether a given topic access is allowed.
//
// ACL rules:
//   - Site credentials: full read+write to their own topic prefix subtree
//   - User credentials: read across all user's sites, write only to /set topics
func CheckACL(credType storage.MQTTCredentialType, topicPrefix, userID, topic string, acc int) bool {
	switch credType {
	case storage.MQTTCredSite:
		if !strings.HasPrefix(topic, topicPrefix+"/") && topic != topicPrefix {
			return false
		}
		return true

	case storage.MQTTCredUser:
		userPrefix := "user/" + userID + "/site/"
		if !strings.HasPrefix(topic, userPrefix) {
			return false
		}
		switch acc {
		case 1: // read
			return true
		case 2: // write
			return strings.HasSuffix(topic, "/set")
		case 3: // read+write
			return strings.HasSuffix(topic, "/set")
		default:
			return false
		}

	default:
		return false
	}
}
