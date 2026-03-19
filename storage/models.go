package storage

// User model
type User struct {
	ID        string `gorm:"primaryKey"` // We use string IDs in proto, so let's stick to string
	Name      string
	Email     string `gorm:"uniqueIndex"`
	CreatedAt int64
	// For simplicity in this demo, we can store friends as a separate relationship or join table
	Friends []*User `gorm:"many2many:user_friends"`
}

// Expense model
type Expense struct {
	ID          string `gorm:"primaryKey"`
	Description string
	TotalAmount float64
	PaidBy      string
	SplitType   int
	CreatedAt   int64
	Splits      []Split `gorm:"foreignKey:ExpenseID"`
}

// Split model
type Split struct {
	ID        uint `gorm:"primaryKey"`
	ExpenseID string
	UserID    string
	Amount    float64
}

// Balance struct (not a table, but used for query results)
type Balance struct {
	FromUserID string
	ToUserID   string
	Amount     float64
}

// Settlement records a payment from one user to another.
// We keep this separate so balances can be updated after "settle up" without mutating expenses.
type Settlement struct {
	ID          uint `gorm:"primaryKey"`
	FromUserID  string `gorm:"index"`
	ToUserID    string `gorm:"index"`
	Amount      float64
	CreatedAt   int64
}
