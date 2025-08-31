package models

import (
	"time"
)

// RoomType represents the type of a chat room
type RoomType string

const (
	RoomTypePublic  RoomType = "public"
	RoomTypePrivate RoomType = "private"
	RoomTypeDirect  RoomType = "direct"
)

// RoomRole represents the role of a user in a room
type RoomRole string

const (
	RoomRoleAdmin     RoomRole = "admin"
	RoomRoleModerator RoomRole = "moderator"
	RoomRoleMember    RoomRole = "member"
)

// Room represents a chat room
type Room struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description" db:"description"`
	RoomType     RoomType  `json:"roomType" db:"room_type"`
	PasswordHash *string   `json:"-" db:"password_hash"` // Never expose password hash
	CreatorID    string    `json:"creatorId" db:"creator_id"`
	AvatarURL    *string   `json:"avatarUrl" db:"avatar_url"`
	MaxMembers   int       `json:"maxMembers" db:"max_members"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	
	// Populated fields (not from DB directly)
	Creator      *User         `json:"creator"`
	Members      []*RoomMember `json:"members"`
	MemberCount  int           `json:"memberCount"`
	OnlineCount  int           `json:"onlineCount"`
}

// RoomMember represents a user's membership in a room
type RoomMember struct {
	ID         string     `json:"id" db:"id"`
	RoomID     string     `json:"roomId" db:"room_id"`
	UserID     string     `json:"userId" db:"user_id"`
	Role       RoomRole   `json:"role" db:"role"`
	JoinedAt   time.Time  `json:"joinedAt" db:"joined_at"`
	LastReadAt *time.Time `json:"lastReadAt" db:"last_read_at"`
	
	// Populated fields
	User *User `json:"user"`
}

// CreateRoomInput represents the input for creating a new room
type CreateRoomInput struct {
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	RoomType    RoomType  `json:"roomType"`
	Password    *string   `json:"password"`
	AvatarURL   *string   `json:"avatarUrl"`
	MaxMembers  *int      `json:"maxMembers"`
}

// UpdateRoomInput represents the input for updating a room
type UpdateRoomInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	AvatarURL   *string `json:"avatarUrl"`
	MaxMembers  *int    `json:"maxMembers"`
}

// JoinRoomInput represents the input for joining a room
type JoinRoomInput struct {
	RoomID   string  `json:"roomId"`
	Password *string `json:"password"`
}

// UpdateRoomMemberInput represents the input for updating a room member
type UpdateRoomMemberInput struct {
	RoomID string    `json:"roomId"`
	UserID string    `json:"userId"`
	Role   *RoomRole `json:"role"`
}