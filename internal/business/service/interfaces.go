package service

import (
	"context"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

// MessageService defines the interface for message business logic
type MessageService interface {
	// CreateMessage creates a new message and publishes it for real-time updates
	CreateMessage(ctx context.Context, input *models.CreateMessageInput) (*models.Message, error)
	
	// GetMessagesByRoom retrieves all messages for a specific room
	GetMessagesByRoom(ctx context.Context, room string) ([]*models.Message, error)
	
	// SubscribeToRoom subscribes to real-time messages for a specific room
	SubscribeToRoom(ctx context.Context, room string) (<-chan *models.Message, error)
}