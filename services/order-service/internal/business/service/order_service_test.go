package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	user "github.com/terminator791/Layered-Architecture-graphql-GO/proto"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/mocks"
)

func TestOrderService_CreateOrder_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockOrderRepository{}
	mockUserClient := &mocks.MockUserServiceClient{}
	
	service := NewOrderService(mockRepo, mockUserClient)
	
	input := &models.CreateOrderInput{
		UserID:    "user-123",
		ProductID: "product-456",
		Quantity:  2,
		Amount:    99.99,
	}
	
	expectedUser := &user.User{
		Id:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now().Unix(),
	}
	
	expectedOrder := &models.Order{
		ID:        "order-789",
		UserID:    "user-123",
		ProductID: "product-456",
		Quantity:  2,
		Amount:    99.99,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	
	// Mock expectations
	mockUserClient.On("GetUser", ctx, "user-123").Return(expectedUser, nil)
	mockRepo.On("CreateOrder", ctx, mock.AnythingOfType("*models.Order")).Return(expectedOrder, nil)
	
	// Act
	result, err := service.CreateOrder(ctx, input)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedOrder.ID, result.ID)
	assert.Equal(t, expectedOrder.UserID, result.UserID)
	assert.Equal(t, expectedOrder.ProductID, result.ProductID)
	assert.Equal(t, expectedOrder.Quantity, result.Quantity)
	assert.Equal(t, expectedOrder.Amount, result.Amount)
	
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_CreateOrder_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockOrderRepository{}
	mockUserClient := &mocks.MockUserServiceClient{}
	
	service := NewOrderService(mockRepo, mockUserClient)
	
	input := &models.CreateOrderInput{
		UserID:    "non-existent-user",
		ProductID: "product-456",
		Quantity:  2,
		Amount:    99.99,
	}
	
	// Mock user service to return NOT_FOUND error
	notFoundError := status.Error(codes.NotFound, "user not found")
	mockUserClient.On("GetUser", ctx, "non-existent-user").Return(nil, notFoundError)
	
	// Act
	result, err := service.CreateOrder(ctx, input)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user with ID non-existent-user not found")
	
	mockUserClient.AssertExpectations(t)
	// Repository should not be called since user validation failed
	mockRepo.AssertNotCalled(t, "CreateOrder")
}

func TestOrderService_CreateOrder_UserServiceError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockOrderRepository{}
	mockUserClient := &mocks.MockUserServiceClient{}
	
	service := NewOrderService(mockRepo, mockUserClient)
	
	input := &models.CreateOrderInput{
		UserID:    "user-123",
		ProductID: "product-456",
		Quantity:  2,
		Amount:    99.99,
	}
	
	// Mock user service to return internal error
	internalError := status.Error(codes.Internal, "internal server error")
	mockUserClient.On("GetUser", ctx, "user-123").Return(nil, internalError)
	
	// Act
	result, err := service.CreateOrder(ctx, input)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to validate user")
	
	mockUserClient.AssertExpectations(t)
	// Repository should not be called since user validation failed
	mockRepo.AssertNotCalled(t, "CreateOrder")
}

func TestOrderService_CreateOrder_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockOrderRepository{}
	mockUserClient := &mocks.MockUserServiceClient{}
	
	service := NewOrderService(mockRepo, mockUserClient)
	
	input := &models.CreateOrderInput{
		UserID:    "user-123",
		ProductID: "product-456",
		Quantity:  2,
		Amount:    99.99,
	}
	
	expectedUser := &user.User{
		Id:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now().Unix(),
	}
	
	repositoryError := fmt.Errorf("database connection error")
	
	// Mock expectations
	mockUserClient.On("GetUser", ctx, "user-123").Return(expectedUser, nil)
	mockRepo.On("CreateOrder", ctx, mock.AnythingOfType("*models.Order")).Return(nil, repositoryError)
	
	// Act
	result, err := service.CreateOrder(ctx, input)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save order")
	
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_CreateOrder_ValidationErrors(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockOrderRepository{}
	mockUserClient := &mocks.MockUserServiceClient{}
	
	service := NewOrderService(mockRepo, mockUserClient)
	
	testCases := []struct {
		name        string
		input       *models.CreateOrderInput
		expectedErr string
	}{
		{
			name:        "nil input",
			input:       nil,
			expectedErr: "input cannot be nil",
		},
		{
			name: "empty user ID",
			input: &models.CreateOrderInput{
				UserID:    "",
				ProductID: "product-456",
				Quantity:  2,
				Amount:    99.99,
			},
			expectedErr: "user ID cannot be empty",
		},
		{
			name: "empty product ID",
			input: &models.CreateOrderInput{
				UserID:    "user-123",
				ProductID: "",
				Quantity:  2,
				Amount:    99.99,
			},
			expectedErr: "product ID cannot be empty",
		},
		{
			name: "invalid quantity",
			input: &models.CreateOrderInput{
				UserID:    "user-123",
				ProductID: "product-456",
				Quantity:  0,
				Amount:    99.99,
			},
			expectedErr: "quantity must be greater than 0",
		},
		{
			name: "invalid amount",
			input: &models.CreateOrderInput{
				UserID:    "user-123",
				ProductID: "product-456",
				Quantity:  2,
				Amount:    0,
			},
			expectedErr: "amount must be greater than 0",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := service.CreateOrder(ctx, tc.input)
			
			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
	
	// Ensure no external calls were made during validation failures
	mockUserClient.AssertNotCalled(t, "GetUser")
	mockRepo.AssertNotCalled(t, "CreateOrder")
}

func TestOrderService_GetOrderByID_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockOrderRepository{}
	mockUserClient := &mocks.MockUserServiceClient{}
	
	service := NewOrderService(mockRepo, mockUserClient)
	
	orderID := "order-789"
	expectedOrder := &models.Order{
		ID:        orderID,
		UserID:    "user-123",
		ProductID: "product-456",
		Quantity:  2,
		Amount:    99.99,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	
	// Mock expectations
	mockRepo.On("GetOrderByID", ctx, orderID).Return(expectedOrder, nil)
	
	// Act
	result, err := service.GetOrderByID(ctx, orderID)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedOrder.ID, result.ID)
	assert.Equal(t, expectedOrder.UserID, result.UserID)
	
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetOrderByID_EmptyID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &mocks.MockOrderRepository{}
	mockUserClient := &mocks.MockUserServiceClient{}
	
	service := NewOrderService(mockRepo, mockUserClient)
	
	// Act
	result, err := service.GetOrderByID(ctx, "")
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "order ID cannot be empty")
	
	mockRepo.AssertNotCalled(t, "GetOrderByID")
}