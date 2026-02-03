package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/smtp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/vinay/splitwise-grpc/internal/user/domain"
	"github.com/vinay/splitwise-grpc/internal/user/repository"
)

// userService implements the Service interface
type userService struct {
	repo repository.Repository
	log  *slog.Logger
}

// NewUserService creates a new user service
func NewUserService(repo repository.Repository, log *slog.Logger) Service {
	return &userService{
		repo: repo,
		log:  log,
	}
}

// Register creates a new user
func (s *userService) Register(ctx context.Context, id, name, email, password string) (*domain.User, error) {
	s.log.Info("Registering user", "email", email, "id", id)

	// Check if user already exists
	existingUser, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Use provided ID or generate one if empty
	if id == "" {
		id = generateUserID(email)
	}

	// Create domain user (with validation)
	user, err := domain.NewUser(id, name, email, string(passwordHash))
	if err != nil {
		return nil, err
	}

	// Persist
	if err := s.repo.Create(ctx, user); err != nil {
		s.log.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.log.Info("User registered successfully", "id", user.ID)
	return user, nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// AddFriend creates a friendship between two users
func (s *userService) AddFriend(ctx context.Context, userID, friendID string) error {
	s.log.Info("Adding friend", "user_id", userID, "friend_id", friendID)

	// Business rule: cannot friend yourself
	if userID == friendID {
		return domain.ErrCannotFriendSelf
	}

	// Verify both users exist
	if _, err := s.repo.FindByID(ctx, userID); err != nil {
		return err
	}
	if _, err := s.repo.FindByID(ctx, friendID); err != nil {
		return err
	}

	// Create friendship
	if err := s.repo.AddFriend(ctx, userID, friendID); err != nil {
		s.log.Error("Failed to add friend", "error", err)
		return fmt.Errorf("failed to add friend: %w", err)
	}

	s.log.Info("Friendship created", "user_id", userID, "friend_id", friendID)
	s.log.Info("Friendship created", "user_id", userID, "friend_id", friendID)
	return nil
}

// AddFriendByEmail adds a friend by email, inviting them if they don't exist
func (s *userService) AddFriendByEmail(ctx context.Context, userID, email string) error {
	s.log.Info("Adding friend by email", "user_id", userID, "email", email)

	// Check if user exists
	friend, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if friend != nil {
		// User exists, add as friend
		return s.AddFriend(ctx, userID, friend.ID)
	}

	// User does not exist, send invite
	return s.sendInviteEmail(email)
}

func (s *userService) sendInviteEmail(email string) error {
	// Simple SMTP sender for Mailhog
	// In production, inject an EmailService dependency
	addr := "mailhog:1025"
	from := "no-reply@expense-sharing.com"
	subject := "Join Expense Sharing App!"
	body := fmt.Sprintf("Hello!\n\nYou have been invited to join the Expense Sharing App.\n\nSign up here: http://localhost:3000/register?email=%s\n", email)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", email, subject, body))

	// Note: We use mailhog hostname because this runs inside Docker
	// If running locally (go run), it might fail if mailhog is not in /etc/hosts
	// But since we use Docker Compose, 'mailhog' resolves correctly.
	// For local development with 'bazel run' or 'go run', we might need 'localhost:1025'
	// A robust solution would use config. For now, we try 'mailhog' first, fallback to 'localhost'

	err := smtp.SendMail(addr, nil, from, []string{email}, msg)
	if err != nil {
		// Try localhost fallback
		addr = "localhost:1025"
		err = smtp.SendMail(addr, nil, from, []string{email}, msg)
		if err != nil {
			s.log.Error("Failed to send invite email", "error", err)
			return fmt.Errorf("failed to send invite: %w", err)
		}
	}

	s.log.Info("Invite email sent", "email", email)
	return nil
}

// GetFriends retrieves all friends for a user
func (s *userService) GetFriends(ctx context.Context, userID string) ([]*domain.User, error) {
	return s.repo.GetFriends(ctx, userID)
}

// ListUsers retrieves all users
func (s *userService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.repo.List(ctx)
}

// generateUserID generates a unique user ID
// In production, use github.com/google/uuid
func generateUserID(email string) string {
	return fmt.Sprintf("user_%s_%d", email, time.Now().UnixNano())
}
