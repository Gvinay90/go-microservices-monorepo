package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vinay/splitwise-grpc/api-gateway/clients"
	"github.com/vinay/splitwise-grpc/pkg/config"
	"github.com/vinay/splitwise-grpc/pkg/keycloak"
)

type AuthHandler struct {
	userClient     *clients.UserClient
	keycloakClient *keycloak.Client
	config         *config.Config
	logger         *slog.Logger
}

func NewAuthHandler(userClient *clients.UserClient, cfg *config.Config, logger *slog.Logger) *AuthHandler {
	// Initialize Keycloak client
	kcClient := keycloak.NewClient(
		cfg.Auth.KeycloakURL,
		cfg.Auth.Realm,
		cfg.Auth.ClientID,
		cfg.Auth.ClientSecret,
		logger,
	)

	return &AuthHandler{
		userClient:     userClient,
		keycloakClient: kcClient,
		config:         cfg,
		logger:         logger,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// Register creates a new user account in both Keycloak and our database
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user in Keycloak first
	keycloakUserID, err := h.keycloakClient.CreateUser(req.Email, req.Name, req.Password)
	if err != nil {
		h.logger.Error("Failed to create user in Keycloak", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	// Create user in our database with Keycloak ID
	resp, err := h.userClient.RegisterUser(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to create user in database", "error", err, "keycloak_id", keycloakUserID)
		// Note: User exists in Keycloak but not in our DB - this is a partial failure
		// In production, implement compensation logic or use distributed transactions
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":    resp.User.Id,
			"name":  resp.User.Name,
			"email": resp.User.Email,
		},
		"message": "User registered successfully. You can now login.",
	})
}

// Login authenticates user via Keycloak and returns JWT token
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate with Keycloak
	tokenResp, err := h.keycloakClient.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to authenticate", "error", err, "email", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenResp.AccessToken,
		"refresh_token": tokenResp.RefreshToken,
		"token_type":    tokenResp.TokenType,
		"expires_in":    tokenResp.ExpiresIn,
	})
}

// GetMe returns current user information
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	email, _ := c.Get("email")
	roles, _ := c.Get("roles")

	// Ensure user exists in our database (sync from Keycloak)
	userIDStr := userID.(string)
	emailStr, _ := email.(string)

	// Try to get user from our database
	_, err := h.userClient.GetUser(c.Request.Context(), "", userIDStr)
	if err != nil {
		// User doesn't exist in our DB, create them
		h.logger.Info("User not found in database, creating from Keycloak", "user_id", userIDStr, "email", emailStr)

		// Extract name from email (before @)
		name := emailStr
		if atIndex := len(emailStr); atIndex > 0 {
			for i, ch := range emailStr {
				if ch == '@' {
					name = emailStr[:i]
					break
				}
			}
		}

		// Register user with Keycloak ID
		_, err := h.userClient.RegisterUserWithID(c.Request.Context(), userIDStr, name, emailStr, "keycloak-managed")
		if err != nil {
			h.logger.Error("Failed to sync user to database", "error", err, "user_id", userIDStr)
			// Continue anyway - user can still use the app
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"email":   email,
		"roles":   roles,
	})
}

// ForgotPassword sends password reset email via Keycloak
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send password reset email via Keycloak
	if err := h.keycloakClient.SendPasswordResetEmail(req.Email); err != nil {
		h.logger.Error("Failed to send password reset email", "error", err, "email", req.Email)
		// Don't reveal if user exists or not for security
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent successfully"})
}

// RefreshToken refreshes the user's access token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenResp, err := h.keycloakClient.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Failed to refresh token", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}
