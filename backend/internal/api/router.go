package api

import (
	"net/http"

	"evcc-cloud/backend/internal/storage"

	"github.com/gin-gonic/gin"
	sentrygin "github.com/getsentry/sentry-go/gin"
)

// Config holds the runtime configuration passed to the router.
type Config struct {
	JWTSecret      string
	DevMode        bool
	CORSOrigin     string // Allowed origin in production (e.g. "https://cloud.evcc.io")
	MQTTBrokerAddr string // optional, for health check
}

// NewRouter builds and returns the gin engine with all routes registered.
func NewRouter(db *storage.DB, cfg Config) *gin.Engine {
	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	r.Use(gin.Recovery())
	r.Use(SecurityHeaders())

	if cfg.DevMode {
		r.Use(corsMiddleware("")) // reflect any origin in dev
	} else if cfg.CORSOrigin != "" {
		r.Use(corsMiddleware(cfg.CORSOrigin)) // restrict in production
	}

	// Health check.
	hh := &healthHandler{db: db, mqttBrokerAddr: cfg.MQTTBrokerAddr}
	r.GET("/health", hh.Health)

	ah := &authHandler{db: db, jwtSecret: cfg.JWTSecret}
	mh := &mqttAuthHandler{db: db}
	sh := &siteHandler{db: db}

	apiGroup := r.Group("/api")
	{
		authRateLimiter := RateLimiter(0.083, 5) // ~5 requests per minute per IP, burst of 5
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/register", authRateLimiter, ah.Register)
			authGroup.POST("/login", authRateLimiter, ah.Login)
			authGroup.POST("/refresh", ah.Refresh)
			authGroup.POST("/logout", ah.Logout)
		}

		mqttGroup := apiGroup.Group("/mqtt")
		{
			mqttGroup.POST("/auth", mh.MQTTAuth)
			mqttGroup.POST("/acl", mh.MQTTACL)
		}

		profileGroup := apiGroup.Group("/auth")
		profileGroup.Use(JWTAuthMiddleware(cfg.JWTSecret))
		{
			profileGroup.GET("/profile", ah.GetProfile)
			profileGroup.PUT("/password", ah.ChangePassword)
		}

		sitesGroup := apiGroup.Group("/sites")
		sitesGroup.Use(JWTAuthMiddleware(cfg.JWTSecret))
		{
			sitesGroup.POST("", sh.CreateSite)
			sitesGroup.GET("", sh.ListSites)
			sitesGroup.PUT("/:id", sh.UpdateSite)
			sitesGroup.DELETE("/:id", sh.DeleteSite)
			sitesGroup.GET("/:id/credentials", sh.GetSiteCredentials)
		}
	}

	return r
}

// corsMiddleware returns CORS middleware. In dev mode, the requesting origin is reflected.
// In production, only the configured origin is allowed.
func corsMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := allowedOrigin
		if origin == "" {
			// Dev: reflect the requesting origin
			origin = c.GetHeader("Origin")
		}
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
