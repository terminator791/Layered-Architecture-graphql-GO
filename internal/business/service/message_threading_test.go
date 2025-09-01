package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/mocks"
)

func TestMessageService_StartThread_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	userID := "user_123"
	messageID := "message_456"
	roomID := "room_789"
	
	input := &models.StartThreadInput{
		MessageID: messageID,
		Text:      "Starting a thread on this message",
	}
	
	originalMessage := &models.Message{
		ID:           messageID,
		RoomID:       &roomID,
		Text:         "Original message",
		IsThreadRoot: false,
		CreatedAt:    time.Now(),
	}
	
	member := &models.RoomMember{
		UserID: userID,
		RoomID: roomID,
		Role:   models.RoomRoleMember,
	}
	
	user := &models.User{
		ID:       userID,
		Username: "testuser",
	}
	
	expectedReply := &models.Message{
		ID:       "reply_123",
		UserID:   &userID,
		RoomID:   &roomID,
		Text:     input.Text,
		ThreadID: &messageID,
		User:     user.Username,
	}
	
	// Mock expectations
	mockRepo.On("GetMessageByID", ctx, messageID).Return(originalMessage, nil)
	mockRoomRepo.On("GetRoomMember", ctx, roomID, userID).Return(member, nil)
	mockRepo.On("MarkAsThreadRoot", ctx, messageID).Return(nil)
	mockUserRepo.On("GetUserByID", ctx, userID).Return(user, nil)
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*models.Message")).Return(expectedReply, nil)
	mockPublisher.On("PublishMessage", ctx, roomID, expectedReply).Return(nil)
	
	// Act
	result, err := service.StartThread(ctx, userID, input)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, input.Text, result.Text)
	assert.Equal(t, &messageID, result.ThreadID)
	assert.Equal(t, user, result.UserInfo)
	
	// Verify all expectations were met
	mockRepo.AssertExpectations(t)
	mockRoomRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestMessageService_StartThread_InvalidInput(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	userID := "user_123"
	
	testCases := []struct {
		name  string
		input *models.StartThreadInput
	}{
		{
			name:  "nil input",
			input: nil,
		},
		{
			name: "empty message ID",
			input: &models.StartThreadInput{
				MessageID: "",
				Text:      "Thread text",
			},
		},
		{
			name: "empty text",
			input: &models.StartThreadInput{
				MessageID: "message_123",
				Text:      "",
			},
		},
		{
			name: "text too long",
			input: &models.StartThreadInput{
				MessageID: "message_123",
				Text:      string(make([]byte, 1001)), // > 1000 characters
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := service.StartThread(ctx, userID, tc.input)
			
			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestMessageService_StartThread_MessageNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	userID := "user_123"
	messageID := "nonexistent_message"
	
	input := &models.StartThreadInput{
		MessageID: messageID,
		Text:      "Thread text",
	}
	
	// Mock expectations
	mockRepo.On("GetMessageByID", ctx, messageID).Return(nil, nil)
	
	// Act
	result, err := service.StartThread(ctx, userID, input)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "original message not found")
	
	mockRepo.AssertExpectations(t)
}

func TestMessageService_StartThread_UserNotMember(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	userID := "user_123"
	messageID := "message_456"
	roomID := "room_789"
	
	input := &models.StartThreadInput{
		MessageID: messageID,
		Text:      "Thread text",
	}
	
	originalMessage := &models.Message{
		ID:     messageID,
		RoomID: &roomID,
		Text:   "Original message",
	}
	
	// Mock expectations
	mockRepo.On("GetMessageByID", ctx, messageID).Return(originalMessage, nil)
	mockRoomRepo.On("GetRoomMember", ctx, roomID, userID).Return(nil, nil)
	
	// Act
	result, err := service.StartThread(ctx, userID, input)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user is not a member of this room")
	
	mockRepo.AssertExpectations(t)
	mockRoomRepo.AssertExpectations(t)
}

func TestMessageService_ReplyToThread_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	userID := "user_123"
	threadID := "thread_456"
	roomID := "room_789"
	
	input := &models.ReplyToThreadInput{
		ThreadID: threadID,
		Text:     "Reply to thread",
	}
	
	threadRoot := &models.Message{
		ID:           threadID,
		RoomID:       &roomID,
		Text:         "Thread root message",
		IsThreadRoot: true,
		CreatedAt:    time.Now(),
	}
	
	member := &models.RoomMember{
		UserID: userID,
		RoomID: roomID,
		Role:   models.RoomRoleMember,
	}
	
	user := &models.User{
		ID:       userID,
		Username: "testuser",
	}
	
	expectedReply := &models.Message{
		ID:       "reply_123",
		UserID:   &userID,
		RoomID:   &roomID,
		Text:     input.Text,
		ThreadID: &threadID,
		User:     user.Username,
	}
	
	// Mock expectations
	mockRepo.On("GetMessageByID", ctx, threadID).Return(threadRoot, nil)
	mockRoomRepo.On("GetRoomMember", ctx, roomID, userID).Return(member, nil)
	mockUserRepo.On("GetUserByID", ctx, userID).Return(user, nil)
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*models.Message")).Return(expectedReply, nil)
	mockPublisher.On("PublishMessage", ctx, roomID, expectedReply).Return(nil)
	
	// Act
	result, err := service.ReplyToThread(ctx, userID, input)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, input.Text, result.Text)
	assert.Equal(t, &threadID, result.ThreadID)
	assert.Equal(t, user, result.UserInfo)
	
	// Verify all expectations were met
	mockRepo.AssertExpectations(t)
	mockRoomRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestMessageService_ReplyToThread_NotThreadRoot(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	userID := "user_123"
	threadID := "message_456"
	roomID := "room_789"
	
	input := &models.ReplyToThreadInput{
		ThreadID: threadID,
		Text:     "Reply to thread",
	}
	
	notThreadRoot := &models.Message{
		ID:           threadID,
		RoomID:       &roomID,
		Text:         "Regular message",
		IsThreadRoot: false, // Not a thread root
		CreatedAt:    time.Now(),
	}
	
	// Mock expectations
	mockRepo.On("GetMessageByID", ctx, threadID).Return(notThreadRoot, nil)
	
	// Act
	result, err := service.ReplyToThread(ctx, userID, input)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "message is not a thread root")
	
	mockRepo.AssertExpectations(t)
}

func TestMessageService_GetThreadReplies_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	threadID := "thread_123"
	limit := 20
	offset := 0
	
	replies := []*models.Message{
		{
			ID:       "reply_1",
			ThreadID: &threadID,
			Text:     "First reply",
			UserID:   stringPtr("user_1"),
		},
		{
			ID:       "reply_2",
			ThreadID: &threadID,
			Text:     "Second reply",
			UserID:   stringPtr("user_2"),
		},
	}
	
	users := []*models.User{
		{ID: "user_1", Username: "user1"},
		{ID: "user_2", Username: "user2"},
	}
	
	// Mock expectations
	mockRepo.On("GetThreadReplies", ctx, threadID, limit, offset).Return(replies, nil)
	mockUserRepo.On("GetUsersByIDs", ctx, []string{"user_1", "user_2"}).Return(users, nil)
	mockRepo.On("GetMessageReactions", ctx, "reply_1").Return([]*models.MessageReaction{}, nil)
	mockRepo.On("GetReactionCounts", ctx, "reply_1").Return(map[string]int{}, nil)
	mockRepo.On("GetMessageReactions", ctx, "reply_2").Return([]*models.MessageReaction{}, nil)
	mockRepo.On("GetReactionCounts", ctx, "reply_2").Return(map[string]int{}, nil)
	
	// Act
	result, err := service.GetThreadReplies(ctx, threadID, limit, offset)
	
	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "First reply", result[0].Text)
	assert.Equal(t, "Second reply", result[1].Text)
	
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestMessageService_GetThreadReplies_EmptyThreadID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	// Act
	result, err := service.GetThreadReplies(ctx, "", 20, 0)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "thread ID cannot be empty")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}