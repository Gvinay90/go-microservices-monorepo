package service

import (
	"context"

	"github.com/vinay/splitwise-grpc/internal/expense/domain"
)

// Service defines the business logic interface for expense operations
type Service interface {
	// CreateExpense creates a new expense
	CreateExpense(ctx context.Context, description string, amount float64, paidBy string, splits []domain.Split) (*domain.Expense, error)

	// GetExpense retrieves an expense by ID
	GetExpense(ctx context.Context, id string) (*domain.Expense, error)

	// ListExpenses retrieves all expenses
	ListExpenses(ctx context.Context, userID string) ([]*domain.Expense, error)

	// GetUserBalance calculates the balance for a user
	GetUserBalance(ctx context.Context, userID string) (float64, error)

	// SettleBalance settles a balance between users
	SettleBalance(ctx context.Context, fromUserID, toUserID string, amount float64) error

	// UpdateExpense updates an expense (description + total_amount)
	UpdateExpense(ctx context.Context, expenseID string, description string, totalAmount float64) (*domain.Expense, error)

	// DeleteExpense deletes an expense
	DeleteExpense(ctx context.Context, expenseID string) error
}
