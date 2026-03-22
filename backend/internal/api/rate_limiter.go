package api

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter returns a Gin middleware that rate-limits by client IP.
// rps = requests per second, burst = max burst size.
// Stale entries (no requests for 3+ minutes) are evicted periodically.
func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	limiters := make(map[string]*ipLimiter)

	go func() {
		for {
			time.Sleep(3 * time.Minute)
			mu.Lock()
			for ip, entry := range limiters {
				if time.Since(entry.lastSeen) > 3*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		if entry, ok := limiters[ip]; ok {
			entry.lastSeen = time.Now()
			return entry.limiter
		}
		lim := rate.NewLimiter(rate.Limit(rps), burst)
		limiters[ip] = &ipLimiter{limiter: lim, lastSeen: time.Now()}
		return lim
	}

	return func(c *gin.Context) {
		ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			ip = c.Request.RemoteAddr
		}
		if !getLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
