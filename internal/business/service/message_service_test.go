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
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
	input := &models.CreateMessageInput{
		Room: "general",
		User: "testuser",
		Text: "Hello, world!",
	}
	
	expectedMessage := &models.Message{
		ID:        "test-id",
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
	
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestMessageService_CreateMessage_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
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
	
	mockRepo.AssertExpectations(t)
	// Publisher should not be called if repository fails
	mockPublisher.AssertNotCalled(t, "PublishMessage")
}

func TestMessageService_CreateMessage_PublisherError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
	input := &models.CreateMessageInput{
		Room: "general",
		User: "testuser",
		Text: "Hello, world!",
	}
	
	expectedMessage := &models.Message{
		ID:        "test-id",
		Room:      "general",
		User:      "testuser",
		Text:      "Hello, world!",
		CreatedAt: time.Now(),
	}
	
	publishError := fmt.Errorf("redis connection error")
	
	// Mock expectations
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*models.Message")).Return(expectedMessage, nil)
	mockPublisher.On("PublishMessage", ctx, "general", expectedMessage).Return(publishError)
	
	// Act
	result, err := service.CreateMessage(ctx, input)
	
	// Assert
	// Even if publishing fails, the service should still return the saved message
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMessage.ID, result.ID)
	
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestMessageService_CreateMessage_ValidationErrors(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
	testCases := []struct {
		name        string
		input       *models.CreateMessageInput
		expectedErr string
	}{
		{
			name:        "nil input",
			input:       nil,
			expectedErr: "input cannot be nil",
		},
		{
			name: "empty room",
			input: &models.CreateMessageInput{
				Room: "",
				User: "testuser",
				Text: "Hello, world!",
			},
			expectedErr: "room cannot be empty",
		},
		{
			name: "empty user",
			input: &models.CreateMessageInput{
				Room: "general",
				User: "",
				Text: "Hello, world!",
			},
			expectedErr: "user cannot be empty",
		},
		{
			name: "empty text",
			input: &models.CreateMessageInput{
				Room: "general",
				User: "testuser",
				Text: "",
			},
			expectedErr: "text cannot be empty",
		},
		{
			name: "text too long",
			input: &models.CreateMessageInput{
				Room: "general",
				User: "testuser",
				Text: string(make([]byte, 1001)), // 1001 characters
			},
			expectedErr: "message text cannot exceed 1000 characters",
		},
		{
			name: "username too long",
			input: &models.CreateMessageInput{
				Room: "general",
				User: string(make([]byte, 51)), // 51 characters
				Text: "Hello, world!",
			},
			expectedErr: "username cannot exceed 50 characters",
		},
		{
			name: "room name too long",
			input: &models.CreateMessageInput{
				Room: string(make([]byte, 51)), // 51 characters
				User: "testuser",
				Text: "Hello, world!",
			},
			expectedErr: "room name cannot exceed 50 characters",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := service.CreateMessage(ctx, tc.input)
			
			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
	
	// Repository and Publisher should not be called for validation errors
	mockRepo.AssertNotCalled(t, "CreateMessage")
	mockPublisher.AssertNotCalled(t, "PublishMessage")
}

func TestMessageService_GetMessagesByRoom_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
	room := "general"
	expectedMessages := []*models.Message{
		{
			ID:        "msg1",
			Room:      "general",
			User:      "user1",
			Text:      "Hello",
			CreatedAt: time.Now(),
		},
		{
			ID:        "msg2",
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
	assert.Equal(t, len(expectedMessages), len(result))
	assert.Equal(t, expectedMessages[0].ID, result[0].ID)
	assert.Equal(t, expectedMessages[1].ID, result[1].ID)
	
	mockRepo.AssertExpectations(t)
}

func TestMessageService_GetMessagesByRoom_EmptyRoom(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
	// Act
	result, err := service.GetMessagesByRoom(ctx, "")
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "room cannot be empty")
	
	mockRepo.AssertNotCalled(t, "GetMessagesByRoom")
}

func TestMessageService_GetMessagesByRoom_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
	room := "general"
	expectedError := fmt.Errorf("database connection error")
	
	// Mock expectations
	mockRepo.On("GetMessagesByRoom", ctx, room).Return(nil, expectedError)
	
	// Act
	result, err := service.GetMessagesByRoom(ctx, room)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get messages for room")
	
	mockRepo.AssertExpectations(t)
}

func TestMessageService_SubscribeToRoom_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
	room := "general"
	messageCh := make(chan *models.Message)
	
	// Mock expectations
	mockSubscriber.On("SubscribeToRoom", ctx, room).Return((<-chan *models.Message)(messageCh), nil)
	
	// Act
	result, err := service.SubscribeToRoom(ctx, room)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	mockSubscriber.AssertExpectations(t)
}

func TestMessageService_SubscribeToRoom_EmptyRoom(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
	// Act
	result, err := service.SubscribeToRoom(ctx, "")
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "room cannot be empty")
	
	mockSubscriber.AssertNotCalled(t, "SubscribeToRoom")
}

func TestMessageService_SubscribeToRoom_SubscriberError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockMessageRepository{}
	mockPublisher := &mocks.MockPublisher{}
	mockSubscriber := &mocks.MockSubscriber{}
	
	service := NewMessageService(mockRepo, mockPublisher, mockSubscriber)
	
	room := "general"
	expectedError := fmt.Errorf("redis subscription error")
	
	// Mock expectations
	mockSubscriber.On("SubscribeToRoom", ctx, room).Return(nil, expectedError)
	
	// Act
	result, err := service.SubscribeToRoom(ctx, room)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to subscribe to room")
	
	mockSubscriber.AssertExpectations(t)
}