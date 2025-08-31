package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CreateUserInput represents the input for creating a user
type CreateUserInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}