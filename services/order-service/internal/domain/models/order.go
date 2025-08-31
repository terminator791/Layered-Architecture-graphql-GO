package models

import (
	"time"
)

// Order represents an order in the system
type Order struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	ProductID string    `db:"product_id" json:"product_id"`
	Quantity  int       `db:"quantity" json:"quantity"`
	Amount    float64   `db:"amount" json:"amount"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CreateOrderInput represents the input for creating an order
type CreateOrderInput struct {
	UserID    string  `json:"user_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Amount    float64 `json:"amount"`
}