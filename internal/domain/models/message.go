package models

import (
	"time"
)

// Message represents a chat message in the domain
type Message struct {
	ID        string    `json:"id" db:"id"`
	Room      string    `json:"room" db:"room"`
	User      string    `json:"user" db:"user"`
	Text      string    `json:"text" db:"text"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// CreateMessageInput represents the input for creating a new message
type CreateMessageInput struct {
	Room string `json:"room"`
	User string `json:"user"`
	Text string `json:"text"`
}