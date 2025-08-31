package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	user "github.com/terminator791/Layered-Architecture-graphql-GO/proto"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/business/service"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/domain/models"
)

// UserServiceServer implements the gRPC UserService
type UserServiceServer struct {
	user.UnimplementedUserServiceServer
	userService service.UserService
}

// NewUserServiceServer creates a new gRPC user service server
func NewUserServiceServer(userService service.UserService) *UserServiceServer {
	return &UserServiceServer{
		userService: userService,
	}
}

// GetUser retrieves a user by ID
func (s *UserServiceServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	userModel, err := s.userService.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user: %v", err))
	}

	if userModel == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Convert domain model to proto
	protoUser := &user.User{
		Id:        userModel.ID,
		Email:     userModel.Email,
		Name:      userModel.Name,
		CreatedAt: userModel.CreatedAt.Unix(),
	}

	return &user.GetUserResponse{
		User: protoUser,
	}, nil
}

// CreateUser creates a new user
func (s *UserServiceServer) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	input := &models.CreateUserInput{
		Email: req.Email,
		Name:  req.Name,
	}

	userModel, err := s.userService.CreateUser(ctx, input)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create user: %v", err))
	}

	// Convert domain model to proto
	protoUser := &user.User{
		Id:        userModel.ID,
		Email:     userModel.Email,
		Name:      userModel.Name,
		CreatedAt: userModel.CreatedAt.Unix(),
	}

	return &user.CreateUserResponse{
		User: protoUser,
	}, nil
}