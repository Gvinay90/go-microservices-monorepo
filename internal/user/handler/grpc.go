package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/vinay/splitwise-grpc/internal/user/domain"
	"github.com/vinay/splitwise-grpc/internal/user/service"
	pb "github.com/vinay/splitwise-grpc/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandler implements the gRPC UserServiceServer
type GRPCHandler struct {
	pb.UnimplementedUserServiceServer
	svc service.Service
	log *slog.Logger
}

// NewGRPCHandler creates a new gRPC handler
func NewGRPCHandler(svc service.Service, log *slog.Logger) *GRPCHandler {
	return &GRPCHandler{
		svc: svc,
		log: log,
	}
}

// RegisterUser handles user registration
func (h *GRPCHandler) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	// Use req.Id when syncing from Keycloak (subject id); otherwise service auto-generates
	user, err := h.svc.Register(ctx, req.Id, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, h.mapError(err)
	}

	return &pb.RegisterUserResponse{
		User: toPBUser(user),
	}, nil
}

// GetUser retrieves a user by ID
func (h *GRPCHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.svc.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, h.mapError(err)
	}

	return &pb.GetUserResponse{
		User: toPBUser(user),
	}, nil
}

// UpdateUser updates user information
func (h *GRPCHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// Get existing user
	user, err := h.svc.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, h.mapError(err)
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	// Note: In a real implementation, you'd have an Update method in the service
	// For now, we'll just return the user as-is since we don't have Update in the service layer
	// TODO: Add Update method to user service

	return &pb.UpdateUserResponse{
		User: toPBUser(user),
	}, nil
}

// AddFriend creates a friendship
func (h *GRPCHandler) AddFriend(ctx context.Context, req *pb.AddFriendRequest) (*pb.AddFriendResponse, error) {
	if req.Email != "" {
		if err := h.svc.AddFriendByEmail(ctx, req.UserId, req.Email); err != nil {
			return nil, h.mapError(err)
		}
		return &pb.AddFriendResponse{
			Success: true,
			Message: fmt.Sprintf("Friend request processed for %s", req.Email),
		}, nil
	}

	// Add by ID
	if err := h.svc.AddFriend(ctx, req.UserId, req.FriendId); err != nil {
		return nil, h.mapError(err)
	}

	// Get updated user info for response message
	user, _ := h.svc.GetUser(ctx, req.UserId)
	friend, _ := h.svc.GetUser(ctx, req.FriendId)

	message := "Friendship created"
	if user != nil && friend != nil {
		message = fmt.Sprintf("%s and %s are now friends", user.Name, friend.Name)
	}

	return &pb.AddFriendResponse{
		Success: true,
		Message: message,
	}, nil
}

// GetFriends retrieves all friends for a user
func (h *GRPCHandler) GetFriends(ctx context.Context, req *pb.GetFriendsRequest) (*pb.GetFriendsResponse, error) {
	friends, err := h.svc.GetFriends(ctx, req.UserId)
	if err != nil {
		return nil, h.mapError(err)
	}

	pbFriends := make([]*pb.User, len(friends))
	for i, f := range friends {
		pbFriends[i] = toPBUser(f)
	}

	return &pb.GetFriendsResponse{
		Friends: pbFriends,
	}, nil
}

// ListUsers retrieves all users
func (h *GRPCHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := h.svc.ListUsers(ctx)
	if err != nil {
		return nil, h.mapError(err)
	}

	pbUsers := make([]*pb.User, len(users))
	for i, u := range users {
		pbUsers[i] = toPBUser(u)
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
	}, nil
}

// toPBUser converts domain.User to pb.User
func toPBUser(user *domain.User) *pb.User {
	return &pb.User{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		FriendIds: user.Friends,
	}
}

// mapError maps domain errors to gRPC status codes
func (h *GRPCHandler) mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, domain.ErrInvalidName):
		return status.Error(codes.InvalidArgument, "name is required")
	case errors.Is(err, domain.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, "email is required")
	case errors.Is(err, domain.ErrCannotFriendSelf):
		return status.Error(codes.InvalidArgument, "cannot add yourself as a friend")
	case errors.Is(err, domain.ErrAlreadyFriends):
		return status.Error(codes.AlreadyExists, "already friends")
	default:
		h.log.Error("Unexpected error", "error", err)
		return status.Error(codes.Internal, "internal server error")
	}
}
