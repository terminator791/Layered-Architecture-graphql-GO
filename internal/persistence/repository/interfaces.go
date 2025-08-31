package repository

import (
	"context"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

// MessageRepository defines the interface for message persistence operations
type MessageRepository interface {
	// CreateMessage saves a new message to the database
	CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error)
	
	// GetMessagesByRoom retrieves all messages for a specific room
	GetMessagesByRoom(ctx context.Context, room string) ([]*models.Message, error)
	
	// GetMessageByID retrieves a message by its ID
	GetMessageByID(ctx context.Context, id string) (*models.Message, error)
}