package service

import (
	"context"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

// MessageService defines the interface for message business logic
type MessageService interface {
	// Legacy methods for backward compatibility
	CreateMessage(ctx context.Context, input *models.CreateMessageInput) (*models.Message, error)
	GetMessagesByRoom(ctx context.Context, room string) ([]*models.Message, error)
	SubscribeToRoom(ctx context.Context, room string) (<-chan *models.Message, error)
	
	// Enhanced message operations
	SendMessageToRoom(ctx context.Context, userID, roomID, text string, messageType *models.MessageType, replyToID *string, metadata *models.MessageMetadata) (*models.Message, error)
	GetMessagesByRoomID(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error)
	GetMessageByID(ctx context.Context, messageID string) (*models.Message, error)
	UpdateMessage(ctx context.Context, userID, messageID string, input *models.UpdateMessageInput) (*models.Message, error)
	DeleteMessage(ctx context.Context, userID, messageID string) error
	SearchMessages(ctx context.Context, query string, roomID *string, limit, offset int) ([]*models.Message, error)
	
	// Message reactions
	AddReaction(ctx context.Context, userID string, input *models.AddReactionInput) (*models.MessageReaction, error)
	RemoveReaction(ctx context.Context, userID string, input *models.RemoveReactionInput) error
	GetReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error)
	
	// Enhanced subscriptions
	SubscribeToRoomMessages(ctx context.Context, roomID string) (<-chan *models.Message, error)
	SubscribeToRoomReactions(ctx context.Context, roomID string) (<-chan *models.MessageReaction, error)
	
	// Typing indicators
	StartTyping(ctx context.Context, userID string, input *models.StartTypingInput) error
	StopTyping(ctx context.Context, userID string, input *models.StopTypingInput) error
	GetTypingUsers(ctx context.Context, roomID string) ([]*models.TypingIndicator, error)
	
	// Thread operations
	StartThread(ctx context.Context, userID string, input *models.StartThreadInput) (*models.Message, error)
	ReplyToThread(ctx context.Context, userID string, input *models.ReplyToThreadInput) (*models.Message, error)
	GetThreadReplies(ctx context.Context, threadID string, limit, offset int) ([]*models.Message, error)
}