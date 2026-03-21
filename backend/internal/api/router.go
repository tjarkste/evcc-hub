package api

import (
	"net/http"

	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
)

// Config holds the runtime configuration passed to the router.
type Config struct {
	JWTSecret string
	DevMode   bool
}

// NewRouter builds and returns the gin engine with all routes registered.
func NewRouter(db *storage.DB, cfg Config) *gin.Engine {
	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	if cfg.DevMode {
		r.Use(corsMiddleware())
	}

	// Health check.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	ah := &authHandler{db: db, jwtSecret: cfg.JWTSecret}
	mh := &mqttAuthHandler{db: db}
	sh := &siteHandler{db: db}

	apiGroup := r.Group("/api")
	{
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/register", ah.Register)
			authGroup.POST("/login", ah.Login)
		}

		mqttGroup := apiGroup.Group("/mqtt")
		{
			mqttGroup.POST("/auth", mh.MQTTAuth)
			mqttGroup.POST("/acl", mh.MQTTACL)
		}

		sitesGroup := apiGroup.Group("/sites")
		sitesGroup.Use(JWTAuthMiddleware(cfg.JWTSecret))
		{
			sitesGroup.POST("", sh.CreateSite)
			sitesGroup.GET("", sh.ListSites)
			sitesGroup.PUT("/:id", sh.UpdateSite)
			sitesGroup.DELETE("/:id", sh.DeleteSite)
		}
	}

	return r
}

// corsMiddleware adds permissive CORS headers for development.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
