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
func (h *mqttAuthHandler) MQTTACL(c *gin.Context) {
	var req models.MQTTACLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.GetUserByMQTTUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "unknown user"})
		return
	}

	if !CheckACL(user.TopicPrefix, req.Topic, req.Acc) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.Status(http.StatusOK)
}

// CheckACL decides whether a given topic access is allowed for the user's topicPrefix.
//
// ACL rules:
//   - acc 1 (read):  topic must be exactly the prefix or start with prefix+"/"
//   - acc 2 (write): topic must start with prefix+"/" and end with "/set"
//   - acc 3 (read+write): both read and write rules must be satisfied
//
// Cross-user access is always denied.
func CheckACL(topicPrefix, topic string, acc int) bool {
	// Ensure the topic belongs to this user's prefix.
	if topic != topicPrefix && !strings.HasPrefix(topic, topicPrefix+"/") {
		return false
	}

	switch acc {
	case 1: // read
		return true
	case 2: // write — topic must end with "/set"
		return strings.HasSuffix(topic, "/set")
	case 3: // read+write
		return strings.HasSuffix(topic, "/set")
	default:
		return false
	}
}
