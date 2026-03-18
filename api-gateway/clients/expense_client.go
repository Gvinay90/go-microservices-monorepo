package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/vinay/splitwise-grpc/proto/expense"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ExpenseClient struct {
	conn   *grpc.ClientConn
	client pb.ExpenseServiceClient
}

func NewExpenseClient(address string) (*ExpenseClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to expense service: %w", err)
	}

	return &ExpenseClient{
		conn:   conn,
		client: pb.NewExpenseServiceClient(conn),
	}, nil
}

func (c *ExpenseClient) Close() error {
	return c.conn.Close()
}

func (c *ExpenseClient) contextWithAuth(ctx context.Context, token string) context.Context {
	if token != "" {
		md := metadata.Pairs("authorization", "Bearer "+token)
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

func (c *ExpenseClient) CreateExpense(ctx context.Context, token string, req *pb.CreateExpenseRequest) (*pb.CreateExpenseResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	return c.client.CreateExpense(ctx, req)
}

func (c *ExpenseClient) GetExpense(ctx context.Context, token, expenseID string) (*pb.GetExpenseResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	req := &pb.GetExpenseRequest{ExpenseId: expenseID}
	return c.client.GetExpense(ctx, req)
}

func (c *ExpenseClient) ListExpenses(ctx context.Context, token, userID string) (*pb.ListExpensesResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	req := &pb.ListExpensesRequest{
		UserId: userID,
	}
	return c.client.ListExpenses(ctx, req)
}

func (c *ExpenseClient) UpdateExpense(ctx context.Context, token string, req *pb.UpdateExpenseRequest) (*pb.UpdateExpenseResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	return c.client.UpdateExpense(ctx, req)
}

func (c *ExpenseClient) DeleteExpense(ctx context.Context, token, expenseID string) (*pb.DeleteExpenseResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	req := &pb.DeleteExpenseRequest{ExpenseId: expenseID}
	return c.client.DeleteExpense(ctx, req)
}
