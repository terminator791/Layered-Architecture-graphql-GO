package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/mocks"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockUserRepository{}
	
	service := NewUserService(mockRepo)
	
	input := &models.CreateUserInput{
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	expectedUser := &models.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}
	
	// Mock expectations - no existing user, successful creation
	mockRepo.On("GetUserByEmail", ctx, "test@example.com").Return(nil, nil)
	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(expectedUser, nil)
	
	// Act
	result, err := service.CreateUser(ctx, input)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Email, result.Email)
	assert.Equal(t, expectedUser.Name, result.Name)
	
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_EmailAlreadyExists(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockUserRepository{}
	
	service := NewUserService(mockRepo)
	
	input := &models.CreateUserInput{
		Email: "existing@example.com",
		Name:  "Test User",
	}
	
	existingUser := &models.User{
		ID:        "existing-user-123",
		Email:     "existing@example.com",
		Name:      "Existing User",
		CreatedAt: time.Now(),
	}
	
	// Mock expectations - existing user found
	mockRepo.On("GetUserByEmail", ctx, "existing@example.com").Return(existingUser, nil)
	
	// Act
	result, err := service.CreateUser(ctx, input)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user with email existing@example.com already exists")
	
	mockRepo.AssertExpectations(t)
	// CreateUser should not be called since email already exists
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestUserService_CreateUser_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockUserRepository{}
	
	service := NewUserService(mockRepo)
	
	input := &models.CreateUserInput{
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	repositoryError := fmt.Errorf("database connection error")
	
	// Mock expectations
	mockRepo.On("GetUserByEmail", ctx, "test@example.com").Return(nil, nil)
	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil, repositoryError)
	
	// Act
	result, err := service.CreateUser(ctx, input)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save user")
	
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_ValidationErrors(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockUserRepository{}
	
	service := NewUserService(mockRepo)
	
	testCases := []struct {
		name        string
		input       *models.CreateUserInput
		expectedErr string
	}{
		{
			name:        "nil input",
			input:       nil,
			expectedErr: "input cannot be nil",
		},
		{
			name: "empty email",
			input: &models.CreateUserInput{
				Email: "",
				Name:  "Test User",
			},
			expectedErr: "email cannot be empty",
		},
		{
			name: "empty name",
			input: &models.CreateUserInput{
				Email: "test@example.com",
				Name:  "",
			},
			expectedErr: "name cannot be empty",
		},
		{
			name: "invalid email format",
			input: &models.CreateUserInput{
				Email: "invalid-email",
				Name:  "Test User",
			},
			expectedErr: "invalid email format",
		},
		{
			name: "name too long",
			input: &models.CreateUserInput{
				Email: "test@example.com",
				Name:  string(make([]byte, 101)), // 101 characters
			},
			expectedErr: "name too long",
		},
		{
			name: "email too long",
			input: &models.CreateUserInput{
				Email: string(make([]byte, 256)) + "@example.com", // 256+ characters
				Name:  "Test User",
			},
			expectedErr: "email too long",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := service.CreateUser(ctx, tc.input)
			
			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
	
	// Ensure no external calls were made during validation failures
	mockRepo.AssertNotCalled(t, "GetUserByEmail")
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockUserRepository{}
	
	service := NewUserService(mockRepo)
	
	userID := "user-123"
	expectedUser := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}
	
	// Mock expectations
	mockRepo.On("GetUserByID", ctx, userID).Return(expectedUser, nil)
	
	// Act
	result, err := service.GetUserByID(ctx, userID)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Email, result.Email)
	assert.Equal(t, expectedUser.Name, result.Name)
	
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_EmptyID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockUserRepository{}
	
	service := NewUserService(mockRepo)
	
	// Act
	result, err := service.GetUserByID(ctx, "")
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
	
	mockRepo.AssertNotCalled(t, "GetUserByID")
}

func TestUserService_GetUserByID_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockUserRepository{}
	
	service := NewUserService(mockRepo)
	
	userID := "user-123"
	repositoryError := fmt.Errorf("database connection error")
	
	// Mock expectations
	mockRepo.On("GetUserByID", ctx, userID).Return(nil, repositoryError)
	
	// Act
	result, err := service.GetUserByID(ctx, userID)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user")
	
	mockRepo.AssertExpectations(t)
}