package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vinay/splitwise-grpc/internal/expense/domain"
	"github.com/vinay/splitwise-grpc/storage"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, expense *domain.Expense) error {
	storageExpense := toStorageExpense(expense)

	if err := r.db.WithContext(ctx).Create(&storageExpense).Error; err != nil {
		return fmt.Errorf("failed to create expense: %w", err)
	}

	return nil
}

func (r *gormRepository) FindByID(ctx context.Context, id string) (*domain.Expense, error) {
	var storageExpense storage.Expense

	if err := r.db.WithContext(ctx).Preload("Splits").First(&storageExpense, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrExpenseNotFound
		}
		return nil, fmt.Errorf("failed to find expense: %w", err)
	}

	return toDomainExpense(&storageExpense), nil
}

func (r *gormRepository) List(ctx context.Context, userID string) ([]*domain.Expense, error) {
	var storageExpenses []storage.Expense

	// Find expenses where:
	// 1. User paid for it (paid_by = userID)
	// OR
	// 2. User is part of the splits (joined table)
	err := r.db.WithContext(ctx).
		Preload("Splits").
		Distinct("expenses.*").
		Joins("LEFT JOIN splits ON splits.expense_id = expenses.id").
		Where("expenses.paid_by = ? OR splits.user_id = ?", userID, userID).
		Find(&storageExpenses).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}

	expenses := make([]*domain.Expense, len(storageExpenses))
	for i, se := range storageExpenses {
		expenses[i] = toDomainExpense(&se)
	}

	return expenses, nil
}

func (r *gormRepository) GetUserBalance(ctx context.Context, userID string) (float64, error) {
	var balance float64

	// Calculate what user paid
	var paidTotal float64
	if err := r.db.WithContext(ctx).
		Model(&storage.Expense{}).
		Where("paid_by = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&paidTotal).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate paid total: %w", err)
	}

	// Calculate what user owes
	var owedTotal float64
	if err := r.db.WithContext(ctx).
		Model(&storage.Split{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&owedTotal).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate owed total: %w", err)
	}

	balance = paidTotal - owedTotal
	return balance, nil
}

func (r *gormRepository) SettleBalance(ctx context.Context, fromUserID, toUserID string, amount float64) error {
	// In a real system, you'd record this settlement in a settlements table
	// For now, we'll just validate the users exist
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&storage.Expense{}).
		Where("paid_by IN ?", []string{fromUserID, toUserID}).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to validate users: %w", err)
	}

	return nil
}

func toStorageExpense(expense *domain.Expense) storage.Expense {
	splits := make([]storage.Split, len(expense.Splits))
	for i, s := range expense.Splits {
		splits[i] = storage.Split{
			UserID: s.UserID,
			Amount: s.Amount,
		}
	}

	return storage.Expense{
		ID:          expense.ID,
		Description: expense.Description,
		TotalAmount: expense.Amount,
		PaidBy:      expense.PaidBy,
		Splits:      splits,
		CreatedAt:   expense.CreatedAt.Unix(),
	}
}

func toDomainExpense(expense *storage.Expense) *domain.Expense {
	splits := make([]domain.Split, len(expense.Splits))
	for i, s := range expense.Splits {
		splits[i] = domain.Split{
			UserID: s.UserID,
			Amount: s.Amount,
		}
	}

	return &domain.Expense{
		ID:          expense.ID,
		Description: expense.Description,
		Amount:      expense.TotalAmount,
		PaidBy:      expense.PaidBy,
		Splits:      splits,
		CreatedAt:   time.Unix(expense.CreatedAt, 0),
	}
}
