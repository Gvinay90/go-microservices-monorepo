package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vinay/splitwise-grpc/api-gateway/clients"
	pb "github.com/vinay/splitwise-grpc/proto/expense"
)

type ExpenseHandler struct {
	expenseClient *clients.ExpenseClient
	logger        *slog.Logger
}

func NewExpenseHandler(expenseClient *clients.ExpenseClient, logger *slog.Logger) *ExpenseHandler {
	return &ExpenseHandler{
		expenseClient: expenseClient,
		logger:        logger,
	}
}

type SplitRequest struct {
	UserID string  `json:"user_id" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type CreateExpenseRequest struct {
	Description string         `json:"description" binding:"required"`
	TotalAmount float64        `json:"total_amount" binding:"required,gt=0"`
	PaidBy      string         `json:"paid_by" binding:"required"`
	Splits      []SplitRequest `json:"splits" binding:"required,min=1"`
}

// CreateExpense creates a new expense
func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)

	var req CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert splits
	splits := make([]*pb.Split, len(req.Splits))
	for i, split := range req.Splits {
		splits[i] = &pb.Split{
			UserId: split.UserID,
			Amount: split.Amount,
		}
	}

	// Convert to gRPC request
	grpcReq := &pb.CreateExpenseRequest{
		Description: req.Description,
		TotalAmount: req.TotalAmount,
		PaidBy:      req.PaidBy,
		SplitType:   pb.SplitType_EQUAL,
		Splits:      splits,
	}

	resp, err := h.expenseClient.CreateExpense(c.Request.Context(), tokenStr, grpcReq)
	if err != nil {
		h.logger.Error("Failed to create expense", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create expense"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"expense": gin.H{
			"id":           resp.Expense.Id,
			"description":  resp.Expense.Description,
			"total_amount": resp.Expense.TotalAmount,
			"paid_by":      resp.Expense.PaidBy,
			"created_at":   resp.Expense.CreatedAt,
		},
	})
}

// ListExpenses returns all expenses
func (h *ExpenseHandler) ListExpenses(c *gin.Context) {
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	resp, err := h.expenseClient.ListExpenses(c.Request.Context(), tokenStr, userIDStr)
	if err != nil {
		h.logger.Error("Failed to list expenses", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list expenses"})
		return
	}

	expenses := make([]gin.H, 0, len(resp.Expenses))
	for _, expense := range resp.Expenses {
		splits := make([]gin.H, 0, len(expense.Splits))
		for _, split := range expense.Splits {
			splits = append(splits, gin.H{
				"user_id": split.UserId,
				"amount":  split.Amount,
			})
		}
		expenses = append(expenses, gin.H{
			"id":           expense.Id,
			"description":  expense.Description,
			"total_amount": expense.TotalAmount,
			"paid_by":      expense.PaidBy,
			"splits":       splits,
			"created_at":   expense.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"expenses": expenses})
}

// GetExpense returns a specific expense
func (h *ExpenseHandler) GetExpense(c *gin.Context) {
	expenseID := c.Param("id")
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)

	resp, err := h.expenseClient.GetExpense(c.Request.Context(), tokenStr, expenseID)
	if err != nil {
		h.logger.Error("Failed to get expense", "error", err, "expense_id", expenseID)
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}

	splits := make([]gin.H, 0, len(resp.Expense.Splits))
	for _, split := range resp.Expense.Splits {
		splits = append(splits, gin.H{
			"user_id": split.UserId,
			"amount":  split.Amount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"expense": gin.H{
			"id":           resp.Expense.Id,
			"description":  resp.Expense.Description,
			"total_amount": resp.Expense.TotalAmount,
			"paid_by":      resp.Expense.PaidBy,
			"splits":       splits,
			"created_at":   resp.Expense.CreatedAt,
		},
	})
}

// UpdateExpense updates an expense
func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
	expenseID := c.Param("id")
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)

	var req struct {
		Description string  `json:"description"`
		TotalAmount float64 `json:"total_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &pb.UpdateExpenseRequest{
		ExpenseId:   expenseID,
		Description: req.Description,
		TotalAmount: req.TotalAmount,
	}

	resp, err := h.expenseClient.UpdateExpense(c.Request.Context(), tokenStr, grpcReq)
	if err != nil {
		h.logger.Error("Failed to update expense", "error", err, "expense_id", expenseID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update expense"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expense": gin.H{
			"id":           resp.Expense.Id,
			"description":  resp.Expense.Description,
			"total_amount": resp.Expense.TotalAmount,
			"paid_by":      resp.Expense.PaidBy,
		},
	})
}

// DeleteExpense deletes an expense
func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
	expenseID := c.Param("id")
	token, _ := c.Get("jwt_token")
	tokenStr, _ := token.(string)

	_, err := h.expenseClient.DeleteExpense(c.Request.Context(), tokenStr, expenseID)
	if err != nil {
		h.logger.Error("Failed to delete expense", "error", err, "expense_id", expenseID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete expense"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "expense deleted successfully"})
}
