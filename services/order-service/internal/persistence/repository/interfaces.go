package repository

import (
	"context"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/domain/models"
)

// OrderRepository defines the interface for order data access
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	GetOrderByID(ctx context.Context, id string) (*models.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*models.Order, error)
}