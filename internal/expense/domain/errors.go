package domain

import "errors"

var (
	ErrInvalidDescription = errors.New("description is required")
	ErrInvalidAmount      = errors.New("amount must be greater than 0")
	ErrInvalidPaidBy      = errors.New("paidBy user ID is required")
	ErrNoSplits           = errors.New("at least one split is required")
	ErrSplitMismatch      = errors.New("splits do not add up to total amount")
	ErrExpenseNotFound    = errors.New("expense not found")
	ErrUserNotFound       = errors.New("user not found")
)
