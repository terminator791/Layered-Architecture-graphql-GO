package models

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of a message
type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

// Message represents a chat message in the domain
type Message struct {
	ID          string       `json:"id" db:"id"`
	Room        string       `json:"room" db:"room"`          // Legacy field for backward compatibility
	User        string       `json:"user" db:"user"`          // Legacy field for backward compatibility
	Text        string       `json:"text" db:"text"`
	UserID      *string      `json:"userId" db:"user_id"`
	RoomID      *string      `json:"roomId" db:"room_id"`
	MessageType MessageType  `json:"messageType" db:"message_type"`
	ReplyToID   *string      `json:"replyToId" db:"reply_to_id"`
	EditedAt    *time.Time   `json:"editedAt" db:"edited_at"`
	DeletedAt   *time.Time   `json:"deletedAt" db:"deleted_at"`
	Metadata    *MessageMetadata `json:"metadata"`
	CreatedAt   time.Time    `json:"createdAt" db:"created_at"`
	
	// Populated fields (not from DB directly)
	UserInfo      *User              `json:"userInfo"`
	RoomInfo      *Room              `json:"roomInfo"`
	ReplyTo       *Message           `json:"replyTo"`
	Reactions     []*MessageReaction `json:"reactions"`
	ReactionCount map[string]int     `json:"reactionCount"`
}

// MessageMetadata contains type-specific metadata for messages
type MessageMetadata struct {
	// For image messages
	ImageWidth  *int    `json:"imageWidth,omitempty"`
	ImageHeight *int    `json:"imageHeight,omitempty"`
	ImageURL    *string `json:"imageUrl,omitempty"`
	
	// For file messages
	FileName *string `json:"fileName,omitempty"`
	FileSize *int64  `json:"fileSize,omitempty"`
	FileURL  *string `json:"fileUrl,omitempty"`
	MimeType *string `json:"mimeType,omitempty"`
	
	// For system messages
	SystemType    *string                `json:"systemType,omitempty"`
	SystemData    map[string]interface{} `json:"systemData,omitempty"`
}

// MessageReaction represents a reaction to a message
type MessageReaction struct {
	ID        string    `json:"id" db:"id"`
	MessageID string    `json:"messageId" db:"message_id"`
	UserID    string    `json:"userId" db:"user_id"`
	Emoji     string    `json:"emoji" db:"emoji"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	
	// Populated fields
	User *User `json:"user"`
}

// TypingIndicator represents a user typing in a room
type TypingIndicator struct {
	ID        string    `json:"id" db:"id"`
	RoomID    string    `json:"roomId" db:"room_id"`
	UserID    string    `json:"userId" db:"user_id"`
	StartedAt time.Time `json:"startedAt" db:"started_at"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	
	// Populated fields
	User *User `json:"user"`
}

// CreateMessageInput represents the input for creating a new message
type CreateMessageInput struct {
	Room        string           `json:"room"`        // Legacy field for backward compatibility
	User        string           `json:"user"`        // Legacy field for backward compatibility
	Text        string           `json:"text"`
	RoomID      *string          `json:"roomId"`
	MessageType *MessageType     `json:"messageType"`
	ReplyToID   *string          `json:"replyToId"`
	Metadata    *MessageMetadata `json:"metadata"`
}

// UpdateMessageInput represents the input for updating a message
type UpdateMessageInput struct {
	MessageID string           `json:"messageId"`
	Text      *string          `json:"text"`
	Metadata  *MessageMetadata `json:"metadata"`
}

// AddReactionInput represents the input for adding a reaction to a message
type AddReactionInput struct {
	MessageID string `json:"messageId"`
	Emoji     string `json:"emoji"`
}

// RemoveReactionInput represents the input for removing a reaction from a message
type RemoveReactionInput struct {
	MessageID string `json:"messageId"`
	Emoji     string `json:"emoji"`
}

// StartTypingInput represents the input for starting typing indicator
type StartTypingInput struct {
	RoomID string `json:"roomId"`
}

// StopTypingInput represents the input for stopping typing indicator
type StopTypingInput struct {
	RoomID string `json:"roomId"`
}

// Custom JSON marshaling for MessageMetadata to handle JSONB in database
func (m *Message) SetMetadataFromJSON(data []byte) error {
	if data == nil {
		return nil
	}
	var metadata MessageMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}
	m.Metadata = &metadata
	return nil
}

func (m *Message) GetMetadataAsJSON() ([]byte, error) {
	if m.Metadata == nil {
		return nil, nil
	}
	return json.Marshal(m.Metadata)
}