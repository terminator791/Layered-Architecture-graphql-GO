package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

type redisPublisher struct {
	client *redis.Client
}

// NewRedisPublisher creates a new Redis publisher
func NewRedisPublisher(client *redis.Client) Publisher {
	return &redisPublisher{
		client: client,
	}
}

func (p *redisPublisher) PublishMessage(ctx context.Context, room string, message *models.Message) error {
	// Serialize message to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish to room-specific channel
	channel := fmt.Sprintf("chat_room:%s", room)
	err = p.client.Publish(ctx, channel, messageJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to publish message to channel %s: %w", channel, err)
	}

	return nil
}

func (p *redisPublisher) Close() error {
	return p.client.Close()
}

type redisSubscriber struct {
	client *redis.Client
}

// NewRedisSubscriber creates a new Redis subscriber
func NewRedisSubscriber(client *redis.Client) Subscriber {
	return &redisSubscriber{
		client: client,
	}
}

func (s *redisSubscriber) SubscribeToRoom(ctx context.Context, room string) (<-chan *models.Message, error) {
	channel := fmt.Sprintf("chat_room:%s", room)
	
	// Subscribe to the channel
	pubsub := s.client.Subscribe(ctx, channel)
	
	// Create a channel to send messages
	messageCh := make(chan *models.Message)
	
	// Start a goroutine to listen for messages
	go func() {
		defer close(messageCh)
		defer pubsub.Close()
		
		// Listen for messages
		ch := pubsub.Channel()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				
				// Deserialize message
				var message models.Message
				if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
					// Log error but continue listening
					continue
				}
				
				// Send message to channel
				select {
				case messageCh <- &message:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return messageCh, nil
}

func (s *redisSubscriber) Close() error {
	return s.client.Close()
}