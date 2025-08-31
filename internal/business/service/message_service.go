package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/infrastructure/redis"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/persistence/repository"
)

type messageService struct {
	repo      repository.MessageRepository
	publisher redis.Publisher
	subscriber redis.Subscriber
}

// NewMessageService creates a new message service
func NewMessageService(
	repo repository.MessageRepository,
	publisher redis.Publisher,
	subscriber redis.Subscriber,
) MessageService {
	return &messageService{
		repo:       repo,
		publisher:  publisher,
		subscriber: subscriber,
	}
}

func (s *messageService) CreateMessage(ctx context.Context, input *models.CreateMessageInput) (*models.Message, error) {
	// Validate input
	if err := s.validateCreateMessageInput(input); err != nil {
		return nil, err
	}

	// Create message model
	message := &models.Message{
		Room: input.Room,
		User: input.User,
		Text: input.Text,
	}

	// Save message to database
	savedMessage, err := s.repo.CreateMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Publish message for real-time updates
	if err := s.publisher.PublishMessage(ctx, savedMessage.Room, savedMessage); err != nil {
		// Log error but don't fail the request since message was saved
		// In production, you might want to use a more robust error handling strategy
		fmt.Printf("Warning: failed to publish message for real-time updates: %v\n", err)
	}

	return savedMessage, nil
}

func (s *messageService) GetMessagesByRoom(ctx context.Context, room string) ([]*models.Message, error) {
	if room == "" {
		return nil, fmt.Errorf("room cannot be empty")
	}

	messages, err := s.repo.GetMessagesByRoom(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages for room %s: %w", room, err)
	}

	return messages, nil
}

func (s *messageService) SubscribeToRoom(ctx context.Context, room string) (<-chan *models.Message, error) {
	if room == "" {
		return nil, fmt.Errorf("room cannot be empty")
	}

	messageCh, err := s.subscriber.SubscribeToRoom(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to room %s: %w", room, err)
	}

	return messageCh, nil
}

func (s *messageService) validateCreateMessageInput(input *models.CreateMessageInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.Room) == "" {
		return fmt.Errorf("room cannot be empty")
	}

	if strings.TrimSpace(input.User) == "" {
		return fmt.Errorf("user cannot be empty")
	}

	if strings.TrimSpace(input.Text) == "" {
		return fmt.Errorf("text cannot be empty")
	}

	// Additional validation rules can be added here
	if len(input.Text) > 1000 {
		return fmt.Errorf("message text cannot exceed 1000 characters")
	}

	if len(input.User) > 50 {
		return fmt.Errorf("username cannot exceed 50 characters")
	}

	if len(input.Room) > 50 {
		return fmt.Errorf("room name cannot exceed 50 characters")
	}

	return nil
}