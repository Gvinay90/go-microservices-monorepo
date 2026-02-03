package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vinay/splitwise-grpc/internal/expense/domain"
	"github.com/vinay/splitwise-grpc/internal/expense/repository"
)

type expenseService struct {
	repo repository.Repository
	log  *slog.Logger
}

func NewExpenseService(repo repository.Repository, log *slog.Logger) Service {
	return &expenseService{
		repo: repo,
		log:  log,
	}
}

func (s *expenseService) CreateExpense(ctx context.Context, description string, amount float64, paidBy string, splits []domain.Split) (*domain.Expense, error) {
	s.log.Info("Creating expense", "description", description, "amount", amount)

	// Generate ID
	id := generateExpenseID(description)

	// Create domain expense (with validation)
	expense, err := domain.NewExpense(id, description, amount, paidBy, splits)
	if err != nil {
		return nil, err
	}

	// Validate splits add up
	if err := expense.ValidateSplits(); err != nil {
		return nil, err
	}

	// Persist
	if err := s.repo.Create(ctx, expense); err != nil {
		s.log.Error("Failed to create expense", "error", err)
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	s.log.Info("Expense created successfully", "id", expense.ID)
	return expense, nil
}

func (s *expenseService) GetExpense(ctx context.Context, id string) (*domain.Expense, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *expenseService) ListExpenses(ctx context.Context, userID string) ([]*domain.Expense, error) {
	return s.repo.List(ctx, userID)
}

func (s *expenseService) GetUserBalance(ctx context.Context, userID string) (float64, error) {
	s.log.Info("Calculating balance", "user_id", userID)

	balance, err := s.repo.GetUserBalance(ctx, userID)
	if err != nil {
		s.log.Error("Failed to calculate balance", "error", err)
		return 0, err
	}

	return balance, nil
}

func (s *expenseService) SettleBalance(ctx context.Context, fromUserID, toUserID string, amount float64) error {
	s.log.Info("Settling balance", "from", fromUserID, "to", toUserID, "amount", amount)

	if amount <= 0 {
		return domain.ErrInvalidAmount
	}

	if err := s.repo.SettleBalance(ctx, fromUserID, toUserID, amount); err != nil {
		s.log.Error("Failed to settle balance", "error", err)
		return err
	}

	s.log.Info("Balance settled", "from", fromUserID, "to", toUserID)
	return nil
}

func generateExpenseID(description string) string {
	return fmt.Sprintf("exp_%d", time.Now().UnixNano())
}
