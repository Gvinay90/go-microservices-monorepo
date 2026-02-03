package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/vinay/splitwise-grpc/internal/user/domain"
	"github.com/vinay/splitwise-grpc/storage"
	"gorm.io/gorm"
)

// gormRepository implements Repository using GORM
type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based repository
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// Create creates a new user
func (r *gormRepository) Create(ctx context.Context, user *domain.User) error {
	storageUser := toStorageUser(user)

	if err := r.db.WithContext(ctx).Create(&storageUser).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by ID
func (r *gormRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var storageUser storage.User

	if err := r.db.WithContext(ctx).Preload("Friends").First(&storageUser, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return toDomainUser(&storageUser), nil
}

// FindByEmail retrieves a user by email
func (r *gormRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var storageUser storage.User

	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&storageUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return toDomainUser(&storageUser), nil
}

// List retrieves all users
func (r *gormRepository) List(ctx context.Context) ([]*domain.User, error) {
	var storageUsers []storage.User

	if err := r.db.WithContext(ctx).Preload("Friends").Find(&storageUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*domain.User, len(storageUsers))
	for i, su := range storageUsers {
		users[i] = toDomainUser(&su)
	}

	return users, nil
}

// AddFriend creates a bidirectional friendship
func (r *gormRepository) AddFriend(ctx context.Context, userID, friendID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user, friend storage.User

		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return domain.ErrUserNotFound
		}

		if err := tx.First(&friend, "id = ?", friendID).Error; err != nil {
			return domain.ErrUserNotFound
		}

		// Add bidirectional friendship
		if err := tx.Model(&user).Association("Friends").Append(&friend); err != nil {
			return fmt.Errorf("failed to add friend: %w", err)
		}

		if err := tx.Model(&friend).Association("Friends").Append(&user); err != nil {
			return fmt.Errorf("failed to add friend: %w", err)
		}

		return nil
	})
}

// GetFriends retrieves all friends for a user
func (r *gormRepository) GetFriends(ctx context.Context, userID string) ([]*domain.User, error) {
	var storageUser storage.User

	if err := r.db.WithContext(ctx).Preload("Friends").First(&storageUser, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}

	friends := make([]*domain.User, len(storageUser.Friends))
	for i, f := range storageUser.Friends {
		friends[i] = &domain.User{
			ID:    f.ID,
			Name:  f.Name,
			Email: f.Email,
		}
	}

	return friends, nil
}

// toStorageUser converts domain.User to storage.User
func toStorageUser(user *domain.User) storage.User {
	return storage.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

// toDomainUser converts storage.User to domain.User
func toDomainUser(user *storage.User) *domain.User {
	friendIDs := make([]string, len(user.Friends))
	for i, f := range user.Friends {
		friendIDs[i] = f.ID
	}

	return &domain.User{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Friends: friendIDs,
	}
}
