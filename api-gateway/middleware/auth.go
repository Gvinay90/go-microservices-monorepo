package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vinay/splitwise-grpc/pkg/auth"
	"github.com/vinay/splitwise-grpc/pkg/config"
)

// AuthRequired validates JWT token and adds user info to context
func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token if auth is enabled
		if cfg.Auth.Enabled {
			validator := auth.NewValidator(cfg.Auth.KeycloakURL, cfg.Auth.Realm)
			claims, err := validator.ValidateToken(c.Request.Context(), token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "invalid or expired token",
				})
				c.Abort()
				return
			}

			// Add claims to context (preferred_username fallback when email is empty)
			c.Set("user_id", claims.Subject)
			c.Set("email", claims.Email)
			c.Set("preferred_username", claims.PreferredUsername)
			c.Set("roles", claims.RealmAccess.Roles)
		}

		// Add token to context for forwarding to gRPC services
		c.Set("jwt_token", token)

		c.Next()
	}
}

// OptionalAuth validates token if present but doesn't require it
func OptionalAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			c.Set("jwt_token", token)

			if cfg.Auth.Enabled {
				validator := auth.NewValidator(cfg.Auth.KeycloakURL, cfg.Auth.Realm)
				if claims, err := validator.ValidateToken(c.Request.Context(), token); err == nil {
					c.Set("user_id", claims.Subject)
					c.Set("email", claims.Email)
					c.Set("preferred_username", claims.PreferredUsername)
					c.Set("roles", claims.RealmAccess.Roles)
				}
			}
		}

		c.Next()
	}
}
