package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	user "github.com/terminator791/Layered-Architecture-graphql-GO/proto"
)

// MockUserServiceClient is a mock implementation of grpc.UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, userID string) (*user.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserServiceClient) Close() error {
	args := m.Called()
	return args.Error(0)
}