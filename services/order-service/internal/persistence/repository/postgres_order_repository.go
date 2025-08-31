package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/domain/models"
)

type postgresOrderRepository struct {
	db *sqlx.DB
}

// NewPostgresOrderRepository creates a new PostgreSQL order repository
func NewPostgresOrderRepository(db *sqlx.DB) OrderRepository {
	return &postgresOrderRepository{
		db: db,
	}
}

func (r *postgresOrderRepository) CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	// Generate UUID if not provided
	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	
	// Set created time if not provided
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}

	// Set default status if not provided
	if order.Status == "" {
		order.Status = "pending"
	}

	query := `
		INSERT INTO orders (id, user_id, product_id, quantity, amount, status, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, user_id, product_id, quantity, amount, status, created_at`

	var savedOrder models.Order
	err := r.db.QueryRowxContext(ctx, query, 
		order.ID, order.UserID, order.ProductID, order.Quantity, order.Amount, order.Status, order.CreatedAt).StructScan(&savedOrder)
	
	if err != nil {
		return nil, err
	}

	return &savedOrder, nil
}

func (r *postgresOrderRepository) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	query := `
		SELECT id, user_id, product_id, quantity, amount, status, created_at 
		FROM orders 
		WHERE id = $1`

	var order models.Order
	err := r.db.QueryRowxContext(ctx, query, id).StructScan(&order)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &order, nil
}

func (r *postgresOrderRepository) GetOrdersByUserID(ctx context.Context, userID string) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, product_id, quantity, amount, status, created_at 
		FROM orders 
		WHERE user_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.StructScan(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}