package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

// MockMessageRepository is a mock implementation of repository.MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error) {
	args := m.Called(ctx, message)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) GetMessagesByRoom(ctx context.Context, room string) ([]*models.Message, error) {
	args := m.Called(ctx, room)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageRepository) GetMessageByID(ctx context.Context, id string) (*models.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}