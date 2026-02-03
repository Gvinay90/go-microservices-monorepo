package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
	logger   *slog.Logger
}

type visitor struct {
	tokens     int
	lastUpdate time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration, logger *slog.Logger) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		logger:   logger,
	}

	// Cleanup old visitors every minute
	go rl.cleanup()

	return rl
}

// Middleware returns a Gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get identifier (user ID or IP)
		identifier := rl.getIdentifier(c)

		if !rl.allow(identifier) {
			rl.logger.Warn("Rate limit exceeded", "identifier", identifier, "path", c.Request.URL.Path)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if request is allowed
func (rl *RateLimiter) allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[identifier]

	if !exists {
		rl.visitors[identifier] = &visitor{
			tokens:     rl.rate - 1,
			lastUpdate: now,
		}
		return true
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(v.lastUpdate)
	tokensToAdd := int(elapsed / rl.window * time.Duration(rl.rate))

	if tokensToAdd > 0 {
		v.tokens = min(rl.rate, v.tokens+tokensToAdd)
		v.lastUpdate = now
	}

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// getIdentifier returns user ID from context or IP address
func (rl *RateLimiter) getIdentifier(c *gin.Context) string {
	// Try to get user ID from context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return "user:" + uid
		}
	}

	// Fall back to IP address
	return "ip:" + c.ClientIP()
}

// cleanup removes old visitors
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for id, v := range rl.visitors {
			if now.Sub(v.lastUpdate) > rl.window*2 {
				delete(rl.visitors, id)
			}
		}
		rl.mu.Unlock()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
