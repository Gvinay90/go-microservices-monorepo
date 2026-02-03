package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vinay/splitwise-grpc/api-gateway/clients"
)

type UserHandler struct {
	userClient *clients.UserClient
	logger     *slog.Logger
}

func NewUserHandler(userClient *clients.UserClient, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userClient: userClient,
		logger:     logger,
	}
}

// ListUsers returns all users
func (h *UserHandler) ListUsers(c *gin.Context) {
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)

	resp, err := h.userClient.ListUsers(c.Request.Context(), tokenStr)
	if err != nil {
		h.logger.Error("Failed to list users", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	users := make([]gin.H, 0, len(resp.Users))
	for _, user := range resp.Users {
		users = append(users, gin.H{
			"id":    user.Id,
			"name":  user.Name,
			"email": user.Email,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUser returns a specific user
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)

	resp, err := h.userClient.GetUser(c.Request.Context(), tokenStr, userID)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err, "user_id", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    resp.User.Id,
			"name":  resp.User.Name,
			"email": resp.User.Email,
		},
	})
}

// UpdateUser updates user information
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.userClient.UpdateUser(c.Request.Context(), tokenStr, userID, req.Name, req.Email)
	if err != nil {
		h.logger.Error("Failed to update user", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    resp.User.Id,
			"name":  resp.User.Name,
			"email": resp.User.Email,
		},
	})
}

// AddFriend adds a friend relationship
func (h *UserHandler) AddFriend(c *gin.Context) {
	userID := c.Param("id")
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)

	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the user client with email (the gRPC layer handles email now)
	_, err := h.userClient.AddFriendByEmail(c.Request.Context(), tokenStr, userID, req.Email)
	if err != nil {
		h.logger.Error("Failed to add friend", "error", err, "user_id", userID, "email", req.Email)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "friend added successfully"})
}

// GetFriends returns user's friends
func (h *UserHandler) GetFriends(c *gin.Context) {
	userID := c.Param("id")
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)

	resp, err := h.userClient.GetFriends(c.Request.Context(), tokenStr, userID)
	if err != nil {
		h.logger.Error("Failed to get friends", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get friends"})
		return
	}

	friends := make([]gin.H, 0, len(resp.Friends))
	for _, friend := range resp.Friends {
		friends = append(friends, gin.H{
			"id":    friend.Id,
			"name":  friend.Name,
			"email": friend.Email,
		})
	}

	c.JSON(http.StatusOK, gin.H{"friends": friends})
}
