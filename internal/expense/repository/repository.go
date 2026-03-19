package repository

import (
	"context"

	"github.com/vinay/splitwise-grpc/internal/expense/domain"
)

// Repository defines the interface for expense data access
type Repository interface {
	// Create creates a new expense
	Create(ctx context.Context, expense *domain.Expense) error

	// FindByID retrieves an expense by ID
	FindByID(ctx context.Context, id string) (*domain.Expense, error)

	// List retrieves all expenses for a user
	List(ctx context.Context, userID string) ([]*domain.Expense, error)

	// GetUserBalance calculates the balance for a user
	GetUserBalance(ctx context.Context, userID string) (float64, error)

	// SettleBalance records a settlement between users
	SettleBalance(ctx context.Context, fromUserID, toUserID string, amount float64) error

	// Update updates an expense (including recalculating split amounts)
	Update(ctx context.Context, expenseID string, description string, totalAmount float64) (*domain.Expense, error)

	// Delete deletes an expense and its splits
	Delete(ctx context.Context, expenseID string) error
}
