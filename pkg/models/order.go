package models

import "time"

// CreateOrderInput represents the input for creating an order
type CreateOrderInput struct {
	UserID    string  `json:"user_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Amount    float64 `json:"amount"`
}

// Order represents an order in the system
type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}