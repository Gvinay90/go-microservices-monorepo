package service

import (
	"context"

	"github.com/vinay/splitwise-grpc/internal/user/domain"
)

// Service defines the business logic interface for user operations
type Service interface {
	// Register creates a new user
	Register(ctx context.Context, id, name, email, password string) (*domain.User, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id string) (*domain.User, error)

	// AddFriend creates a friendship between two users
	AddFriend(ctx context.Context, userID, friendID string) error

	// AddFriendByEmail adds a friend by email, inviting them if they don't exist
	AddFriendByEmail(ctx context.Context, userID, email string) error

	// GetFriends retrieves all friends for a user
	GetFriends(ctx context.Context, userID string) ([]*domain.User, error)

	// ListUsers retrieves all users
	ListUsers(ctx context.Context) ([]*domain.User, error)
}
