package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	user "github.com/terminator791/Layered-Architecture-graphql-GO/proto"
)

// UserServiceClient interface for dependency injection and testing
type UserServiceClient interface {
	GetUser(ctx context.Context, userID string) (*user.User, error)
	Close() error
}

type userServiceClient struct {
	conn   *grpc.ClientConn
	client user.UserServiceClient
}

// NewUserServiceClient creates a new gRPC client for user service
func NewUserServiceClient(address string) (UserServiceClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	client := user.NewUserServiceClient(conn)

	return &userServiceClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *userServiceClient) GetUser(ctx context.Context, userID string) (*user.User, error) {
	resp, err := c.client.GetUser(ctx, &user.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		return nil, err
	}

	return resp.User, nil
}

func (c *userServiceClient) Close() error {
	return c.conn.Close()
}