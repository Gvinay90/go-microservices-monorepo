package domain

import "time"

// Expense represents an expense in the system
type Expense struct {
	ID          string
	Description string
	Amount      float64
	PaidBy      string // User ID
	Splits      []Split
	CreatedAt   time.Time
}

// Split represents how an expense is split among users
type Split struct {
	UserID string
	Amount float64
}

// NewExpense creates a new Expense with validation
func NewExpense(id, description string, amount float64, paidBy string, splits []Split) (*Expense, error) {
	if description == "" {
		return nil, ErrInvalidDescription
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if paidBy == "" {
		return nil, ErrInvalidPaidBy
	}
	if len(splits) == 0 {
		return nil, ErrNoSplits
	}

	return &Expense{
		ID:          id,
		Description: description,
		Amount:      amount,
		PaidBy:      paidBy,
		Splits:      splits,
		CreatedAt:   time.Now(),
	}, nil
}

// ValidateSplits ensures splits add up to the total amount
func (e *Expense) ValidateSplits() error {
	var total float64
	for _, split := range e.Splits {
		total += split.Amount
	}

	// Allow small floating point differences
	if abs(total-e.Amount) > 0.01 {
		return ErrSplitMismatch
	}

	return nil
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
