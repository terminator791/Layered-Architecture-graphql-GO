package redis

import (
	"context"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

// Publisher defines the interface for publishing messages to Redis
type Publisher interface {
	// PublishMessage publishes a message to the specified room channel
	PublishMessage(ctx context.Context, room string, message *models.Message) error
	
	// Close closes the Redis connection
	Close() error
}

// Subscriber defines the interface for subscribing to Redis channels
type Subscriber interface {
	// SubscribeToRoom subscribes to messages for a specific room
	SubscribeToRoom(ctx context.Context, room string) (<-chan *models.Message, error)
	
	// Close closes the Redis connection
	Close() error
}