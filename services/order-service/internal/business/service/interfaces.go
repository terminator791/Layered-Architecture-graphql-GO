package service

import (
	"context"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/domain/models"
)

// OrderService defines the interface for order business logic
type OrderService interface {
	CreateOrder(ctx context.Context, input *models.CreateOrderInput) (*models.Order, error)
	GetOrderByID(ctx context.Context, id string) (*models.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*models.Order, error)
}