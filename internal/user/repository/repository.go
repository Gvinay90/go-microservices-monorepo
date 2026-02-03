package repository

import (
	"context"

	"github.com/vinay/splitwise-grpc/internal/user/domain"
)

// Repository defines the interface for user data access
type Repository interface {
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error

	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id string) (*domain.User, error)

	// FindByEmail retrieves a user by email
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// List retrieves all users
	List(ctx context.Context) ([]*domain.User, error)

	// AddFriend creates a bidirectional friendship
	AddFriend(ctx context.Context, userID, friendID string) error

	// GetFriends retrieves all friends for a user
	GetFriends(ctx context.Context, userID string) ([]*domain.User, error)
}
