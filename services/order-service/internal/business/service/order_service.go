package service

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/grpc"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/persistence/repository"
)

type orderService struct {
	repo           repository.OrderRepository
	userServiceClient grpc.UserServiceClient
}

// NewOrderService creates a new order service
func NewOrderService(repo repository.OrderRepository, userServiceClient grpc.UserServiceClient) OrderService {
	return &orderService{
		repo:           repo,
		userServiceClient: userServiceClient,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, input *models.CreateOrderInput) (*models.Order, error) {
	// Validate input
	if err := s.validateCreateOrderInput(input); err != nil {
		return nil, err
	}

	// Validate user exists by calling user-service via gRPC
	_, err := s.userServiceClient.GetUser(ctx, input.UserID)
	if err != nil {
		// Check if it's a NOT_FOUND error
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return nil, fmt.Errorf("user with ID %s not found", input.UserID)
		}
		return nil, fmt.Errorf("failed to validate user: %w", err)
	}

	// Create order model
	order := &models.Order{
		UserID:    input.UserID,
		ProductID: input.ProductID,
		Quantity:  input.Quantity,
		Amount:    input.Amount,
	}

	// Save order to database
	savedOrder, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return savedOrder, nil
}

func (s *orderService) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order ID cannot be empty")
	}

	order, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

func (s *orderService) GetOrdersByUserID(ctx context.Context, userID string) ([]*models.Order, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	orders, err := s.repo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, nil
}

func (s *orderService) validateCreateOrderInput(input *models.CreateOrderInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.UserID) == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if strings.TrimSpace(input.ProductID) == "" {
		return fmt.Errorf("product ID cannot be empty")
	}

	if input.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	if input.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	return nil
}