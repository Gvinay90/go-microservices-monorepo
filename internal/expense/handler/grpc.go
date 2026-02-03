package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/vinay/splitwise-grpc/internal/expense/domain"
	"github.com/vinay/splitwise-grpc/internal/expense/service"
	pb "github.com/vinay/splitwise-grpc/proto/expense"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedExpenseServiceServer
	svc service.Service
	log *slog.Logger
}

func NewGRPCHandler(svc service.Service, log *slog.Logger) *GRPCHandler {
	return &GRPCHandler{
		svc: svc,
		log: log,
	}
}

func (h *GRPCHandler) CreateExpense(ctx context.Context, req *pb.CreateExpenseRequest) (*pb.CreateExpenseResponse, error) {
	// Convert proto splits to domain splits
	splits := make([]domain.Split, len(req.Splits))
	for i, s := range req.Splits {
		splits[i] = domain.Split{
			UserID: s.UserId,
			Amount: s.Amount,
		}
	}

	expense, err := h.svc.CreateExpense(ctx, req.Description, req.TotalAmount, req.PaidBy, splits)
	if err != nil {
		return nil, h.mapError(err)
	}

	return &pb.CreateExpenseResponse{
		Expense: toPBExpense(expense),
	}, nil
}

func (h *GRPCHandler) GetExpense(ctx context.Context, req *pb.GetExpenseRequest) (*pb.GetExpenseResponse, error) {
	expense, err := h.svc.GetExpense(ctx, req.ExpenseId)
	if err != nil {
		return nil, h.mapError(err)
	}

	return &pb.GetExpenseResponse{
		Expense: toPBExpense(expense),
	}, nil
}

func (h *GRPCHandler) ListExpenses(ctx context.Context, req *pb.ListExpensesRequest) (*pb.ListExpensesResponse, error) {
	expenses, err := h.svc.ListExpenses(ctx, req.UserId)
	if err != nil {
		return nil, h.mapError(err)
	}

	pbExpenses := make([]*pb.Expense, len(expenses))
	for i, e := range expenses {
		pbExpenses[i] = toPBExpense(e)
	}

	return &pb.ListExpensesResponse{
		Expenses: pbExpenses,
	}, nil
}

func (h *GRPCHandler) SettleBalance(ctx context.Context, req *pb.SettleBalanceRequest) (*pb.SettleBalanceResponse, error) {
	if err := h.svc.SettleBalance(ctx, req.FromUserId, req.ToUserId, req.Amount); err != nil {
		return nil, h.mapError(err)
	}

	// Get remaining balance
	remainingBalance, _ := h.svc.GetUserBalance(ctx, req.FromUserId)

	return &pb.SettleBalanceResponse{
		Success: true,
		Message: "Balance settled successfully",
		RemainingBalance: &pb.Balance{
			FromUserId: req.FromUserId,
			ToUserId:   req.ToUserId,
			Amount:     remainingBalance,
		},
	}, nil
}

func toPBExpense(expense *domain.Expense) *pb.Expense {
	splits := make([]*pb.Split, len(expense.Splits))
	for i, s := range expense.Splits {
		splits[i] = &pb.Split{
			UserId: s.UserID,
			Amount: s.Amount,
		}
	}

	return &pb.Expense{
		Id:          expense.ID,
		Description: expense.Description,
		TotalAmount: expense.Amount,
		PaidBy:      expense.PaidBy,
		Splits:      splits,
		CreatedAt:   expense.CreatedAt.Unix(),
	}
}

func (h *GRPCHandler) mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrExpenseNotFound):
		return status.Error(codes.NotFound, "expense not found")
	case errors.Is(err, domain.ErrInvalidDescription):
		return status.Error(codes.InvalidArgument, "description is required")
	case errors.Is(err, domain.ErrInvalidAmount):
		return status.Error(codes.InvalidArgument, "amount must be greater than 0")
	case errors.Is(err, domain.ErrInvalidPaidBy):
		return status.Error(codes.InvalidArgument, "paidBy user ID is required")
	case errors.Is(err, domain.ErrNoSplits):
		return status.Error(codes.InvalidArgument, "at least one split is required")
	case errors.Is(err, domain.ErrSplitMismatch):
		return status.Error(codes.InvalidArgument, "splits do not add up to total amount")
	default:
		h.log.Error("Unexpected error", "error", err)
		return status.Error(codes.Internal, "internal server error")
	}
}
