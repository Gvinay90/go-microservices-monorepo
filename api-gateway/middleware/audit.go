package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogger logs security-relevant events
type AuditLogger struct {
	logger *slog.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *slog.Logger) *AuditLogger {
	return &AuditLogger{
		logger: logger,
	}
}

// Middleware returns a Gin middleware for audit logging
func (al *AuditLogger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get user info if available
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")

		// Process request
		c.Next()

		// Log after request completes
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Log security-relevant events
		if shouldAudit(c.Request.Method, c.Request.URL.Path, statusCode) {
			al.logger.Info("audit",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", statusCode,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
				"user_id", userID,
				"email", email,
				"user_agent", c.Request.UserAgent(),
			)
		}
	}
}

// shouldAudit determines if an event should be logged
func shouldAudit(method, path string, status int) bool {
	// Always log authentication endpoints
	if path == "/api/v1/auth/login" || path == "/api/v1/auth/register" {
		return true
	}

	// Log failed requests
	if status >= 400 {
		return true
	}

	// Log write operations
	if method == "POST" || method == "PUT" || method == "DELETE" {
		return true
	}

	return false
}
