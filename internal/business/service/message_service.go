package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/infrastructure/redis"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/persistence/repository"
)

type messageService struct {
	repo           repository.MessageRepository
	roomRepo       repository.RoomRepository
	userRepo       repository.UserRepository
	typingRepo     repository.TypingRepository
	publisher      redis.Publisher
	subscriber     redis.Subscriber
}

// NewMessageService creates a new message service
func NewMessageService(
	repo repository.MessageRepository,
	roomRepo repository.RoomRepository,
	userRepo repository.UserRepository,
	typingRepo repository.TypingRepository,
	publisher redis.Publisher,
	subscriber redis.Subscriber,
) MessageService {
	return &messageService{
		repo:       repo,
		roomRepo:   roomRepo,
		userRepo:   userRepo,
		typingRepo: typingRepo,
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

// Enhanced message operations

func (s *messageService) SendMessageToRoom(ctx context.Context, userID, roomID, text string, messageType *models.MessageType, replyToID *string, metadata *models.MessageMetadata) (*models.Message, error) {
	// Validate inputs
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if roomID == "" {
		return nil, fmt.Errorf("room ID cannot be empty")
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("message text cannot be empty")
	}
	if len(text) > 1000 {
		return nil, fmt.Errorf("message text cannot exceed 1000 characters")
	}

	// Verify user exists and can access room
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is a member of the room
	member, err := s.roomRepo.GetRoomMember(ctx, roomID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room membership: %w", err)
	}
	if member == nil {
		return nil, fmt.Errorf("user is not a member of this room")
	}

	// Set default message type
	msgType := models.MessageTypeText
	if messageType != nil {
		msgType = *messageType
	}

	// Create message
	message := &models.Message{
		ID:          uuid.New().String(),
		UserID:      &userID,
		RoomID:      &roomID,
		Text:        text,
		MessageType: msgType,
		ReplyToID:   replyToID,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		// Legacy fields for backward compatibility
		Room: roomID,
		User: user.Username,
	}

	// Save message
	savedMessage, err := s.repo.CreateMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Publish for real-time updates
	if err := s.publisher.PublishMessage(ctx, roomID, savedMessage); err != nil {
		fmt.Printf("Warning: failed to publish message for real-time updates: %v\n", err)
	}

	// Populate user info
	savedMessage.UserInfo = user

	return savedMessage, nil
}

func (s *messageService) GetMessagesByRoomID(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room ID cannot be empty")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	messages, err := s.repo.GetMessagesByRoomID(ctx, roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages for room %s: %w", roomID, err)
	}

	// Populate user info for messages
	return s.populateMessageUserInfo(ctx, messages)
}

func (s *messageService) GetMessageByID(ctx context.Context, messageID string) (*models.Message, error) {
	if messageID == "" {
		return nil, fmt.Errorf("message ID cannot be empty")
	}

	message, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	// Populate user info
	messages := []*models.Message{message}
	populatedMessages, err := s.populateMessageUserInfo(ctx, messages)
	if err != nil {
		return nil, err
	}

	return populatedMessages[0], nil
}

func (s *messageService) UpdateMessage(ctx context.Context, userID, messageID string, input *models.UpdateMessageInput) (*models.Message, error) {
	if err := s.validateUpdateMessageInput(input); err != nil {
		return nil, err
	}

	// Get existing message
	message, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	// Check if user owns the message
	if message.UserID == nil || *message.UserID != userID {
		return nil, fmt.Errorf("user does not have permission to update this message")
	}

	// Check if message was deleted
	if message.DeletedAt != nil {
		return nil, fmt.Errorf("cannot update deleted message")
	}

	// Update fields
	if input.Text != nil {
		message.Text = *input.Text
	}
	if input.Metadata != nil {
		message.Metadata = input.Metadata
	}

	// Save updated message
	updatedMessage, err := s.repo.UpdateMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Publish update for real-time
	if message.RoomID != nil {
		if err := s.publisher.PublishMessage(ctx, *message.RoomID, updatedMessage); err != nil {
			fmt.Printf("Warning: failed to publish message update: %v\n", err)
		}
	}

	return updatedMessage, nil
}

func (s *messageService) DeleteMessage(ctx context.Context, userID, messageID string) error {
	if userID == "" || messageID == "" {
		return fmt.Errorf("user ID and message ID cannot be empty")
	}

	// Get existing message
	message, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return fmt.Errorf("message not found")
	}

	// Check if user owns the message or is room moderator
	canDelete := false
	if message.UserID != nil && *message.UserID == userID {
		canDelete = true
	} else if message.RoomID != nil {
		// Check if user is moderator or admin of the room
		member, err := s.roomRepo.GetRoomMember(ctx, *message.RoomID, userID)
		if err == nil && member != nil && (member.Role == models.RoomRoleAdmin || member.Role == models.RoomRoleModerator) {
			canDelete = true
		}
	}

	if !canDelete {
		return fmt.Errorf("user does not have permission to delete this message")
	}

	// Delete message
	if err := s.repo.DeleteMessage(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	// Publish deletion for real-time
	if message.RoomID != nil {
		message.DeletedAt = &time.Time{}
		if err := s.publisher.PublishMessage(ctx, *message.RoomID, message); err != nil {
			fmt.Printf("Warning: failed to publish message deletion: %v\n", err)
		}
	}

	return nil
}

func (s *messageService) SearchMessages(ctx context.Context, query string, roomID *string, limit, offset int) ([]*models.Message, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	messages, err := s.repo.SearchMessages(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	return s.populateMessageUserInfo(ctx, messages)
}

// Message reactions

func (s *messageService) AddReaction(ctx context.Context, userID string, input *models.AddReactionInput) (*models.MessageReaction, error) {
	if err := s.validateAddReactionInput(input); err != nil {
		return nil, err
	}

	// Verify user exists
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Verify message exists
	message, err := s.repo.GetMessageByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	// Check if user can access the room
	if message.RoomID != nil {
		member, err := s.roomRepo.GetRoomMember(ctx, *message.RoomID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check room membership: %w", err)
		}
		if member == nil {
			return nil, fmt.Errorf("user is not a member of this room")
		}
	}

	// Create reaction
	reaction := &models.MessageReaction{
		ID:        uuid.New().String(),
		MessageID: input.MessageID,
		UserID:    userID,
		Emoji:     input.Emoji,
		CreatedAt: time.Now(),
	}

	// Save reaction
	savedReaction, err := s.repo.AddMessageReaction(ctx, reaction)
	if err != nil {
		return nil, fmt.Errorf("failed to add reaction: %w", err)
	}

	// Populate user info
	savedReaction.User = user

	// Publish for real-time
	if message.RoomID != nil {
		// We could create a separate channel for reactions
		fmt.Printf("Reaction added: %s by %s\n", input.Emoji, user.Username)
	}

	return savedReaction, nil
}

func (s *messageService) RemoveReaction(ctx context.Context, userID string, input *models.RemoveReactionInput) error {
	if err := s.validateRemoveReactionInput(input); err != nil {
		return err
	}

	// Verify message exists and user has access
	message, err := s.repo.GetMessageByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return fmt.Errorf("message not found")
	}

	// Check if user can access the room
	if message.RoomID != nil {
		member, err := s.roomRepo.GetRoomMember(ctx, *message.RoomID, userID)
		if err != nil {
			return fmt.Errorf("failed to check room membership: %w", err)
		}
		if member == nil {
			return fmt.Errorf("user is not a member of this room")
		}
	}

	// Remove reaction
	if err := s.repo.RemoveMessageReaction(ctx, input.MessageID, userID, input.Emoji); err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	return nil
}

func (s *messageService) GetReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error) {
	if messageID == "" {
		return nil, fmt.Errorf("message ID cannot be empty")
	}

	reactions, err := s.repo.GetMessageReactions(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reactions: %w", err)
	}

	return reactions, nil
}

// Enhanced subscriptions

func (s *messageService) SubscribeToRoomMessages(ctx context.Context, roomID string) (<-chan *models.Message, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room ID cannot be empty")
	}

	messageCh, err := s.subscriber.SubscribeToRoom(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to room %s: %w", roomID, err)
	}

	return messageCh, nil
}

func (s *messageService) SubscribeToRoomReactions(ctx context.Context, roomID string) (<-chan *models.MessageReaction, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room ID cannot be empty")
	}

	// For now, return a channel that waits for context cancellation
	// In a full implementation, this would subscribe to Redis channels for reaction events
	reactionCh := make(chan *models.MessageReaction)
	
	go func() {
		defer close(reactionCh)
		// Placeholder: In real implementation, subscribe to Redis "room:{roomID}:reactions" channel
		// and forward reaction events to reactionCh
		<-ctx.Done()
	}()

	return reactionCh, nil
}

// Typing indicators

func (s *messageService) StartTyping(ctx context.Context, userID string, input *models.StartTypingInput) error {
	if err := s.validateStartTypingInput(input); err != nil {
		return err
	}

	// Check if user can access the room
	member, err := s.roomRepo.GetRoomMember(ctx, input.RoomID, userID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}
	if member == nil {
		return fmt.Errorf("user is not a member of this room")
	}

	// Start typing indicator
	if err := s.typingRepo.StartTyping(ctx, input.RoomID, userID); err != nil {
		return fmt.Errorf("failed to start typing indicator: %w", err)
	}

	return nil
}

func (s *messageService) StopTyping(ctx context.Context, userID string, input *models.StopTypingInput) error {
	if err := s.validateStopTypingInput(input); err != nil {
		return err
	}

	// Stop typing indicator
	if err := s.typingRepo.StopTyping(ctx, input.RoomID, userID); err != nil {
		return fmt.Errorf("failed to stop typing indicator: %w", err)
	}

	return nil
}

func (s *messageService) GetTypingUsers(ctx context.Context, roomID string) ([]*models.TypingIndicator, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room ID cannot be empty")
	}

	indicators, err := s.typingRepo.GetTypingUsers(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get typing users: %w", err)
	}

	return indicators, nil
}

// Helper functions

func (s *messageService) populateMessageUserInfo(ctx context.Context, messages []*models.Message) ([]*models.Message, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	// Collect unique user IDs
	userIDMap := make(map[string]bool)
	for _, message := range messages {
		if message.UserID != nil {
			userIDMap[*message.UserID] = true
		}
	}

	// Get all users
	userIDs := make([]string, 0, len(userIDMap))
	for userID := range userIDMap {
		userIDs = append(userIDs, userID)
	}

	users, err := s.userRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Create user lookup map
	userMap := make(map[string]*models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// Populate user info in messages
	for _, message := range messages {
		if message.UserID != nil {
			if user, exists := userMap[*message.UserID]; exists {
				message.UserInfo = user
			}
		}

		// Get reactions for message
		reactions, err := s.repo.GetMessageReactions(ctx, message.ID)
		if err == nil {
			message.Reactions = reactions
		}

		// Get reaction counts
		reactionCounts, err := s.repo.GetReactionCounts(ctx, message.ID)
		if err == nil {
			message.ReactionCount = reactionCounts
		}
	}

	return messages, nil
}

// Additional validation functions

func (s *messageService) validateUpdateMessageInput(input *models.UpdateMessageInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.MessageID) == "" {
		return fmt.Errorf("message ID cannot be empty")
	}

	if input.Text != nil && strings.TrimSpace(*input.Text) == "" {
		return fmt.Errorf("message text cannot be empty")
	}

	if input.Text != nil && len(*input.Text) > 1000 {
		return fmt.Errorf("message text cannot exceed 1000 characters")
	}

	return nil
}

func (s *messageService) validateAddReactionInput(input *models.AddReactionInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.MessageID) == "" {
		return fmt.Errorf("message ID cannot be empty")
	}

	if strings.TrimSpace(input.Emoji) == "" {
		return fmt.Errorf("emoji cannot be empty")
	}

	if len(input.Emoji) > 10 {
		return fmt.Errorf("emoji cannot exceed 10 characters")
	}

	return nil
}

func (s *messageService) validateRemoveReactionInput(input *models.RemoveReactionInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.MessageID) == "" {
		return fmt.Errorf("message ID cannot be empty")
	}

	if strings.TrimSpace(input.Emoji) == "" {
		return fmt.Errorf("emoji cannot be empty")
	}

	return nil
}

func (s *messageService) validateStartTypingInput(input *models.StartTypingInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.RoomID) == "" {
		return fmt.Errorf("room ID cannot be empty")
	}

	return nil
}

func (s *messageService) validateStopTypingInput(input *models.StopTypingInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.RoomID) == "" {
		return fmt.Errorf("room ID cannot be empty")
	}

	return nil
}
// Thread operations

func (s *messageService) StartThread(ctx context.Context, userID string, input *models.StartThreadInput) (*models.Message, error) {
	if err := s.validateStartThreadInput(input); err != nil {
		return nil, err
	}
	
	// Get the original message to start thread from
	originalMessage, err := s.repo.GetMessageByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original message: %w", err)
	}
	if originalMessage == nil {
		return nil, fmt.Errorf("original message not found")
	}
	
	// Check if user has access to the room
	if originalMessage.RoomID == nil {
		return nil, fmt.Errorf("cannot start thread on message without room")
	}
	
	member, err := s.roomRepo.GetRoomMember(ctx, *originalMessage.RoomID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room membership: %w", err)
	}
	if member == nil {
		return nil, fmt.Errorf("user is not a member of this room")
	}
	
	// Check if original message is already deleted
	if originalMessage.DeletedAt != nil {
		return nil, fmt.Errorf("cannot start thread on deleted message")
	}
	
	// Mark original message as thread root if not already
	if !originalMessage.IsThreadRoot {
		if err := s.repo.MarkAsThreadRoot(ctx, input.MessageID); err != nil {
			return nil, fmt.Errorf("failed to mark message as thread root: %w", err)
		}
	}
	
	// Create the first reply in the thread
	threadReply := &models.Message{
		ID:          uuid.New().String(),
		UserID:      &userID,
		RoomID:      originalMessage.RoomID,
		Text:        input.Text,
		MessageType: models.MessageTypeText,
		ThreadID:    &input.MessageID, // Thread ID is the original message ID
		CreatedAt:   time.Now(),
		// Legacy fields for backward compatibility
		Room: *originalMessage.RoomID,
		User: "", // Will be populated by user info
	}
	
	// Get user info for legacy field
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user != nil {
		threadReply.User = user.Username
	}
	
	// Save the thread reply
	savedReply, err := s.repo.CreateMessage(ctx, threadReply)
	if err != nil {
		return nil, fmt.Errorf("failed to save thread reply: %w", err)
	}
	
	// Populate user info
	savedReply.UserInfo = user
	
	// Publish real-time event for thread creation
	if err := s.publisher.PublishMessage(ctx, *originalMessage.RoomID, savedReply); err != nil {
		fmt.Printf("Warning: failed to publish thread reply: %v\n", err)
	}
	
	return savedReply, nil
}

func (s *messageService) ReplyToThread(ctx context.Context, userID string, input *models.ReplyToThreadInput) (*models.Message, error) {
	if err := s.validateReplyToThreadInput(input); err != nil {
		return nil, err
	}
	
	// Get the thread root message to validate thread exists
	threadRoot, err := s.repo.GetMessageByID(ctx, input.ThreadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread root: %w", err)
	}
	if threadRoot == nil {
		return nil, fmt.Errorf("thread not found")
	}
	
	// Verify it's actually a thread root
	if !threadRoot.IsThreadRoot {
		return nil, fmt.Errorf("message is not a thread root")
	}
	
	// Check if user has access to the room
	if threadRoot.RoomID == nil {
		return nil, fmt.Errorf("cannot reply to thread without room")
	}
	
	member, err := s.roomRepo.GetRoomMember(ctx, *threadRoot.RoomID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room membership: %w", err)
	}
	if member == nil {
		return nil, fmt.Errorf("user is not a member of this room")
	}
	
	// Check if thread root is deleted
	if threadRoot.DeletedAt != nil {
		return nil, fmt.Errorf("cannot reply to deleted thread")
	}
	
	// Create the thread reply
	threadReply := &models.Message{
		ID:          uuid.New().String(),
		UserID:      &userID,
		RoomID:      threadRoot.RoomID,
		Text:        input.Text,
		MessageType: models.MessageTypeText,
		ThreadID:    &input.ThreadID,
		CreatedAt:   time.Now(),
		// Legacy fields for backward compatibility
		Room: *threadRoot.RoomID,
		User: "", // Will be populated by user info
	}
	
	// Get user info for legacy field
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user != nil {
		threadReply.User = user.Username
	}
	
	// Save the thread reply
	savedReply, err := s.repo.CreateMessage(ctx, threadReply)
	if err != nil {
		return nil, fmt.Errorf("failed to save thread reply: %w", err)
	}
	
	// Populate user info
	savedReply.UserInfo = user
	
	// Publish real-time event for thread reply
	if err := s.publisher.PublishMessage(ctx, *threadRoot.RoomID, savedReply); err != nil {
		fmt.Printf("Warning: failed to publish thread reply: %v\n", err)
	}
	
	return savedReply, nil
}

func (s *messageService) GetThreadReplies(ctx context.Context, threadID string, limit, offset int) ([]*models.Message, error) {
	if threadID == "" {
		return nil, fmt.Errorf("thread ID cannot be empty")
	}
	
	// Validate pagination
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	// Get thread replies
	replies, err := s.repo.GetThreadReplies(ctx, threadID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread replies: %w", err)
	}
	
	// Populate user info for replies
	return s.populateMessageUserInfo(ctx, replies)
}

// Validation functions for thread operations

func (s *messageService) validateStartThreadInput(input *models.StartThreadInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	
	if strings.TrimSpace(input.MessageID) == "" {
		return fmt.Errorf("message ID cannot be empty")
	}
	
	if strings.TrimSpace(input.Text) == "" {
		return fmt.Errorf("thread text cannot be empty")
	}
	
	if len(input.Text) > 1000 {
		return fmt.Errorf("thread text cannot exceed 1000 characters")
	}
	
	return nil
}

func (s *messageService) validateReplyToThreadInput(input *models.ReplyToThreadInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	
	if strings.TrimSpace(input.ThreadID) == "" {
		return fmt.Errorf("thread ID cannot be empty")
	}
	
	if strings.TrimSpace(input.Text) == "" {
		return fmt.Errorf("reply text cannot be empty")
	}
	
	if len(input.Text) > 1000 {
		return fmt.Errorf("reply text cannot exceed 1000 characters")
	}
	
	return nil
}