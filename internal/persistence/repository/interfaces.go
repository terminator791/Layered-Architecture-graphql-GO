package repository

import (
	"context"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

// MessageRepository defines the interface for message persistence operations
type MessageRepository interface {
	// CreateMessage saves a new message to the database
	CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error)
	
	// GetMessagesByRoom retrieves all messages for a specific room (legacy)
	GetMessagesByRoom(ctx context.Context, room string) ([]*models.Message, error)
	
	// GetMessageByID retrieves a message by its ID
	GetMessageByID(ctx context.Context, id string) (*models.Message, error)
	
	// Get messages by room ID with pagination
	GetMessagesByRoomID(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error)
	
	// Update message
	UpdateMessage(ctx context.Context, message *models.Message) (*models.Message, error)
	
	// Delete message (soft delete)
	DeleteMessage(ctx context.Context, id string) error
	
	// Search messages
	SearchMessages(ctx context.Context, query string, roomID *string, limit, offset int) ([]*models.Message, error)
	
	// Get message reactions
	GetMessageReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error)
	
	// Add message reaction
	AddMessageReaction(ctx context.Context, reaction *models.MessageReaction) (*models.MessageReaction, error)
	
	// Remove message reaction
	RemoveMessageReaction(ctx context.Context, messageID, userID, emoji string) error
	
	// Get reaction counts for message
	GetReactionCounts(ctx context.Context, messageID string) (map[string]int, error)
	
	// Thread operations
	StartThread(ctx context.Context, messageID, userID string) (*models.Message, error)
	GetThreadReplies(ctx context.Context, threadID string, limit, offset int) ([]*models.Message, error)
	GetThreadCount(ctx context.Context, threadID string) (int, error)
	MarkAsThreadRoot(ctx context.Context, messageID string) error
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create a new user
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	
	// Get user by ID
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	
	// Get user by username
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	
	// Get user by email
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	
	// Update user
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	
	// Update user status
	UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error
	
	// Update user last seen
	UpdateUserLastSeen(ctx context.Context, userID string) error
	
	// Delete user
	DeleteUser(ctx context.Context, id string) error
	
	// List users with pagination
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	
	// Get users by IDs
	GetUsersByIDs(ctx context.Context, ids []string) ([]*models.User, error)
}

// RoomRepository defines the interface for room data access
type RoomRepository interface {
	// Create a new room
	CreateRoom(ctx context.Context, room *models.Room) (*models.Room, error)
	
	// Get room by ID
	GetRoomByID(ctx context.Context, id string) (*models.Room, error)
	
	// Update room
	UpdateRoom(ctx context.Context, room *models.Room) (*models.Room, error)
	
	// Delete room
	DeleteRoom(ctx context.Context, id string) error
	
	// List rooms with pagination
	ListRooms(ctx context.Context, limit, offset int) ([]*models.Room, error)
	
	// Get rooms by user ID
	GetRoomsByUserID(ctx context.Context, userID string) ([]*models.Room, error)
	
	// Get room members
	GetRoomMembers(ctx context.Context, roomID string) ([]*models.RoomMember, error)
	
	// Get room member
	GetRoomMember(ctx context.Context, roomID, userID string) (*models.RoomMember, error)
	
	// Add room member
	AddRoomMember(ctx context.Context, member *models.RoomMember) (*models.RoomMember, error)
	
	// Update room member
	UpdateRoomMember(ctx context.Context, member *models.RoomMember) (*models.RoomMember, error)
	
	// Remove room member
	RemoveRoomMember(ctx context.Context, roomID, userID string) error
	
	// Get room member count
	GetRoomMemberCount(ctx context.Context, roomID string) (int, error)
	
	// Get online members count
	GetOnlineMemberCount(ctx context.Context, roomID string) (int, error)
	
	// Update member last read
	UpdateMemberLastRead(ctx context.Context, roomID, userID string) error
}

// TypingRepository defines the interface for typing indicators
type TypingRepository interface {
	// Start typing indicator
	StartTyping(ctx context.Context, roomID, userID string) error
	
	// Stop typing indicator
	StopTyping(ctx context.Context, roomID, userID string) error
	
	// Get typing users in room
	GetTypingUsers(ctx context.Context, roomID string) ([]*models.TypingIndicator, error)
	
	// Clean expired typing indicators
	CleanExpiredTypingIndicators(ctx context.Context) error
}