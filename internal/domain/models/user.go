package models

import (
	"time"
)

// UserStatus represents the online status of a user
type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusOffline UserStatus = "offline"
	UserStatusAway    UserStatus = "away"
	UserStatusBusy    UserStatus = "busy"
)

// User represents a user in the system
type User struct {
	ID          string     `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	Email       string     `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose password hash
	DisplayName *string    `json:"displayName" db:"display_name"`
	AvatarURL   *string    `json:"avatarUrl" db:"avatar_url"`
	Bio         *string    `json:"bio" db:"bio"`
	Status      UserStatus `json:"status" db:"status"`
	LastSeenAt  *time.Time `json:"lastSeenAt" db:"last_seen_at"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
}

// CreateUserInput represents the input for creating a new user
type CreateUserInput struct {
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	DisplayName *string `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
	Bio         *string `json:"bio"`
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	DisplayName *string    `json:"displayName"`
	AvatarURL   *string    `json:"avatarUrl"`
	Bio         *string    `json:"bio"`
	Status      *UserStatus `json:"status"`
}

// LoginInput represents the input for user login
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthPayload represents the response after successful authentication
type AuthPayload struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}