package api

import (
	"net"
	"net/http"
	"time"

	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
)

type healthHandler struct {
	db             *storage.DB
	mqttBrokerAddr string // optional, e.g. "mosquitto:1883"
}

func (h *healthHandler) Health(c *gin.Context) {
	checks := gin.H{}
	healthy := true

	if err := h.db.Ping(); err != nil {
		checks["database"] = "error: " + err.Error()
		healthy = false
	} else {
		checks["database"] = "ok"
	}

	if h.mqttBrokerAddr != "" {
		conn, err := net.DialTimeout("tcp", h.mqttBrokerAddr, 3*time.Second)
		if err != nil {
			checks["mqtt_broker"] = "error: " + err.Error()
			healthy = false
		} else {
			conn.Close()
			checks["mqtt_broker"] = "ok"
		}
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !healthy {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status": status,
		"checks": checks,
	})
}
