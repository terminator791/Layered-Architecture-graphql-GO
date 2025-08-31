package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

// MockPublisher is a mock implementation of redis.Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) PublishMessage(ctx context.Context, room string, message *models.Message) error {
	args := m.Called(ctx, room, message)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockSubscriber is a mock implementation of redis.Subscriber
type MockSubscriber struct {
	mock.Mock
}

func (m *MockSubscriber) SubscribeToRoom(ctx context.Context, room string) (<-chan *models.Message, error) {
	args := m.Called(ctx, room)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan *models.Message), args.Error(1)
}

func (m *MockSubscriber) Close() error {
	args := m.Called()
	return args.Error(0)
}