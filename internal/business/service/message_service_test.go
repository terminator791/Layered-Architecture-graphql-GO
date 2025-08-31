package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/mocks"
)

func TestMessageService_CreateMessage_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	input := &models.CreateMessageInput{
		Room: "general",
		User: "testuser",
		Text: "Hello, world!",
	}
	
	expectedMessage := &models.Message{
		ID:        "message-id",
		Room:      "general",
		User:      "testuser",
		Text:      "Hello, world!",
		CreatedAt: time.Now(),
	}
	
	// Mock expectations
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*models.Message")).Return(expectedMessage, nil)
	mockPublisher.On("PublishMessage", ctx, "general", expectedMessage).Return(nil)
	
	// Act
	result, err := service.CreateMessage(ctx, input)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMessage.ID, result.ID)
	assert.Equal(t, expectedMessage.Room, result.Room)
	assert.Equal(t, expectedMessage.User, result.User)
	assert.Equal(t, expectedMessage.Text, result.Text)
	
	// Verify all expectations were met
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestMessageService_CreateMessage_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	input := &models.CreateMessageInput{
		Room: "general",
		User: "testuser",
		Text: "Hello, world!",
	}
	
	expectedError := fmt.Errorf("database error")
	
	// Mock expectations
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*models.Message")).Return(nil, expectedError)
	
	// Act
	result, err := service.CreateMessage(ctx, input)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save message")
	
	// Verify all expectations were met
	mockRepo.AssertExpectations(t)
}

func TestMessageService_CreateMessage_ValidationError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	testCases := []struct {
		name  string
		input *models.CreateMessageInput
	}{
		{
			name:  "nil input",
			input: nil,
		},
		{
			name: "empty room",
			input: &models.CreateMessageInput{
				Room: "",
				User: "testuser",
				Text: "Hello, world!",
			},
		},
		{
			name: "empty user",
			input: &models.CreateMessageInput{
				Room: "general",
				User: "",
				Text: "Hello, world!",
			},
		},
		{
			name: "empty text",
			input: &models.CreateMessageInput{
				Room: "general",
				User: "testuser",
				Text: "",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := service.CreateMessage(ctx, tc.input)
			
			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestMessageService_GetMessagesByRoom_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	room := "general"
	expectedMessages := []*models.Message{
		{
			ID:        "message-1",
			Room:      "general",
			User:      "user1",
			Text:      "Hello",
			CreatedAt: time.Now(),
		},
		{
			ID:        "message-2",
			Room:      "general",
			User:      "user2",
			Text:      "Hi there",
			CreatedAt: time.Now(),
		},
	}
	
	// Mock expectations
	mockRepo.On("GetMessagesByRoom", ctx, room).Return(expectedMessages, nil)
	
	// Act
	result, err := service.GetMessagesByRoom(ctx, room)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedMessages, result)
	
	// Verify all expectations were met
	mockRepo.AssertExpectations(t)
}

func TestMessageService_GetMessagesByRoom_EmptyRoom(t *testing.T) {
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
	result, err := service.GetMessagesByRoom(ctx, "")
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "room cannot be empty")
}

func TestMessageService_GetMessagesByRoom_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	room := "general"
	expectedError := fmt.Errorf("database error")
	
	// Mock expectations
	mockRepo.On("GetMessagesByRoom", ctx, room).Return(nil, expectedError)
	
	// Act
	result, err := service.GetMessagesByRoom(ctx, room)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get messages for room")
	
	// Verify all expectations were met
	mockRepo.AssertExpectations(t)
}

func TestMessageService_SubscribeToRoom_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	room := "general"
	messageCh := make(chan *models.Message, 1)
	
	// Mock expectations
	mockSubscriber.On("SubscribeToRoom", ctx, room).Return((<-chan *models.Message)(messageCh), nil)
	
	// Act
	result, err := service.SubscribeToRoom(ctx, room)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// Verify all expectations were met
	mockSubscriber.AssertExpectations(t)
}

func TestMessageService_SubscribeToRoom_EmptyRoom(t *testing.T) {
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
	result, err := service.SubscribeToRoom(ctx, "")
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "room cannot be empty")
}

func TestMessageService_SubscribeToRoom_SubscriberError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockRoomRepo := &mocks.MockRoomRepository{}
	mockUserRepo := &mocks.MockUserRepository{}
	mockTypingRepo := &mocks.MockTypingRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
	
	room := "general"
	expectedError := fmt.Errorf("redis error")
	
	// Mock expectations
	mockSubscriber.On("SubscribeToRoom", ctx, room).Return(nil, expectedError)
	
	// Act
	result, err := service.SubscribeToRoom(ctx, room)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to subscribe to room")
	
	// Verify all expectations were met
	mockSubscriber.AssertExpectations(t)
}