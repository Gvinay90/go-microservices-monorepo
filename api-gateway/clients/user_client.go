package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/vinay/splitwise-grpc/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type UserClient struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

func NewUserClient(address string) (*UserClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	return &UserClient{
		conn:   conn,
		client: pb.NewUserServiceClient(conn),
	}, nil
}

func (c *UserClient) Close() error {
	return c.conn.Close()
}

// Helper to add JWT to context
func (c *UserClient) contextWithAuth(ctx context.Context, token string) context.Context {
	if token != "" {
		md := metadata.Pairs("authorization", "Bearer "+token)
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

func (c *UserClient) RegisterUser(ctx context.Context, name, email, password string) (*pb.RegisterUserResponse, error) {
	return c.RegisterUserWithID(ctx, "", name, email, password)
}

func (c *UserClient) RegisterUserWithID(ctx context.Context, id, name, email, password string) (*pb.RegisterUserResponse, error) {
	req := &pb.RegisterUserRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}
	return c.client.RegisterUser(ctx, req)
}

func (c *UserClient) GetUser(ctx context.Context, token, userID string) (*pb.GetUserResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	req := &pb.GetUserRequest{UserId: userID}
	return c.client.GetUser(ctx, req)
}

func (c *UserClient) ListUsers(ctx context.Context, token string) (*pb.ListUsersResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	req := &pb.ListUsersRequest{}
	return c.client.ListUsers(ctx, req)
}

func (c *UserClient) UpdateUser(ctx context.Context, token, userID, name, email string) (*pb.UpdateUserResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	req := &pb.UpdateUserRequest{
		UserId: userID,
		Name:   name,
		Email:  email,
	}
	return c.client.UpdateUser(ctx, req)
}

func (c *UserClient) AddFriend(ctx context.Context, token, userID, friendID string) (*pb.AddFriendResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	req := &pb.AddFriendRequest{
		UserId:   userID,
		FriendId: friendID,
	}
	return c.client.AddFriend(ctx, req)
}

func (c *UserClient) AddFriendByEmail(ctx context.Context, token, userID, email string) (*pb.AddFriendResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	req := &pb.AddFriendRequest{
		UserId: userID,
		Email:  email,
	}
	return c.client.AddFriend(ctx, req)
}

func (c *UserClient) GetFriends(ctx context.Context, token, userID string) (*pb.GetFriendsResponse, error) {
	ctx = c.contextWithAuth(ctx, token)
	req := &pb.GetFriendsRequest{UserId: userID}
	return c.client.GetFriends(ctx, req)
}
