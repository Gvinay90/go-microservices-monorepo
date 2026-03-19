package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
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
		Select("COALESCE(SUM(total_amount), 0)").
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

	// Apply settlements ledger:
	// When `fromUserID` pays `toUserID`, `fromUserID` net balance increases by `amount`
	// and `toUserID` net balance decreases by `amount`.
	var settledIn float64
	if err := r.db.WithContext(ctx).
		Model(&storage.Settlement{}).
		Where("from_user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&settledIn).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate settled in: %w", err)
	}

	var settledOut float64
	if err := r.db.WithContext(ctx).
		Model(&storage.Settlement{}).
		Where("to_user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&settledOut).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate settled out: %w", err)
	}

	balance = paidTotal - owedTotal + settledIn - settledOut
	return balance, nil
}

func (r *gormRepository) SettleBalance(ctx context.Context, fromUserID, toUserID string, amount float64) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		// Make sure we don't leave the transaction open.
		_ = tx.Rollback()
	}()

	// Validate users exist.
	var userCount int64
	if err := tx.Model(&storage.User{}).
		Where("id IN ?", []string{fromUserID, toUserID}).
		Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to validate users: %w", err)
	}
	if userCount != 2 {
		return domain.ErrUserNotFound
	}

	// Compute outstanding debt from -> to:
	// sum(splits.amount) where splits.user_id=from AND expenses.paid_by=to
	var owedFromTo float64
	if err := tx.Table("splits").
		Select("COALESCE(SUM(splits.amount), 0)").
		Joins("JOIN expenses ON expenses.id = splits.expense_id").
		Where("splits.user_id = ? AND expenses.paid_by = ?", fromUserID, toUserID).
		Scan(&owedFromTo).Error; err != nil {
		return fmt.Errorf("failed to calculate outstanding owed amount: %w", err)
	}

	// Compute how much has already been settled from -> to.
	var settledFromTo float64
	if err := tx.Model(&storage.Settlement{}).
		Where("from_user_id = ? AND to_user_id = ?", fromUserID, toUserID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&settledFromTo).Error; err != nil {
		return fmt.Errorf("failed to calculate settled amount: %w", err)
	}

	outstanding := owedFromTo - settledFromTo
	// Allow a tiny floating point tolerance.
	if amount > outstanding+0.01 {
		return domain.ErrInsufficientBalance
	}

	if err := tx.Create(&storage.Settlement{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		CreatedAt:  time.Now().Unix(),
	}).Error; err != nil {
		return fmt.Errorf("failed to record settlement: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit settlement transaction: %w", err)
	}

	return nil
}

func (r *gormRepository) Update(ctx context.Context, expenseID string, description string, totalAmount float64) (*domain.Expense, error) {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var storageExpense storage.Expense
	if err := tx.Preload("Splits").First(&storageExpense, "id = ?", expenseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrExpenseNotFound
		}
		return nil, fmt.Errorf("failed to find expense: %w", err)
	}

	if len(storageExpense.Splits) == 0 {
		return nil, domain.ErrNoSplits
	}

	storageExpense.Description = description
	storageExpense.TotalAmount = totalAmount

	// Recalculate equal split across the same set of participants.
	// Distribute rounding cents so sum(splits) == totalAmount (to cents).
	n := int64(len(storageExpense.Splits))
	totalCents := int64(math.Round(totalAmount * 100))
	per := totalCents / n
	rem := totalCents % n

	for i := range storageExpense.Splits {
		additional := int64(0)
		if int64(i) < rem {
			additional = 1
		}
		amtCents := per + additional
		storageExpense.Splits[i].Amount = float64(amtCents) / 100
	}

	// Persist expense fields.
	if err := tx.Model(&storage.Expense{}).
		Where("id = ?", expenseID).
		Updates(map[string]any{
			"description":  storageExpense.Description,
			"total_amount": storageExpense.TotalAmount,
		}).Error; err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	// Persist split amounts.
	for _, split := range storageExpense.Splits {
		if err := tx.Model(&storage.Split{}).
			Where("id = ?", split.ID).
			Update("amount", split.Amount).Error; err != nil {
			return nil, fmt.Errorf("failed to update split: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit expense update transaction: %w", err)
	}

	return toDomainExpense(&storageExpense), nil
}

func (r *gormRepository) Delete(ctx context.Context, expenseID string) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete splits first to keep FK-like constraints happy.
	if err := tx.Where("expense_id = ?", expenseID).Delete(&storage.Split{}).Error; err != nil {
		return fmt.Errorf("failed to delete splits: %w", err)
	}

	res := tx.Where("id = ?", expenseID).Delete(&storage.Expense{})
	if res.Error != nil {
		return fmt.Errorf("failed to delete expense: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrExpenseNotFound
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit expense delete transaction: %w", err)
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
