package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

type postgresMessageRepository struct {
	db *sqlx.DB
}

// NewPostgresMessageRepository creates a new PostgreSQL message repository
func NewPostgresMessageRepository(db *sqlx.DB) MessageRepository {
	return &postgresMessageRepository{
		db: db,
	}
}

func (r *postgresMessageRepository) CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error) {
	// Generate UUID if not provided
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	
	// Set created time if not provided
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO messages (id, room, "user", text, created_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, room, "user", text, created_at`

	var savedMessage models.Message
	err := r.db.QueryRowxContext(ctx, query, 
		message.ID, message.Room, message.User, message.Text, message.CreatedAt).StructScan(&savedMessage)
	
	if err != nil {
		return nil, err
	}

	return &savedMessage, nil
}

func (r *postgresMessageRepository) GetMessagesByRoom(ctx context.Context, room string) ([]*models.Message, error) {
	query := `
		SELECT id, room, "user", text, created_at 
		FROM messages 
		WHERE room = $1 
		ORDER BY created_at ASC`

	rows, err := r.db.QueryxContext(ctx, query, room)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var message models.Message
		if err := rows.StructScan(&message); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}

	return messages, rows.Err()
}

func (r *postgresMessageRepository) GetMessageByID(ctx context.Context, id string) (*models.Message, error) {
	query := `
		SELECT id, room, "user", text, created_at 
		FROM messages 
		WHERE id = $1`

	var message models.Message
	err := r.db.QueryRowxContext(ctx, query, id).StructScan(&message)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &message, nil
}

// GetMessagesByRoomID retrieves messages by room ID with pagination
func (r *postgresMessageRepository) GetMessagesByRoomID(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT id, room, "user", text, user_id, room_id, message_type, reply_to_id, edited_at, deleted_at, metadata, created_at 
		FROM messages 
		WHERE room_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryxContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var message models.Message
		var metadataBytes []byte
		
		err := rows.Scan(
			&message.ID, &message.Room, &message.User, &message.Text,
			&message.UserID, &message.RoomID, &message.MessageType,
			&message.ReplyToID, &message.EditedAt, &message.DeletedAt,
			&metadataBytes, &message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse metadata if present
		if metadataBytes != nil {
			if err := message.SetMetadataFromJSON(metadataBytes); err != nil {
				return nil, err
			}
		}
		
		messages = append(messages, &message)
	}

	return messages, rows.Err()
}

// UpdateMessage updates a message
func (r *postgresMessageRepository) UpdateMessage(ctx context.Context, message *models.Message) (*models.Message, error) {
	metadataBytes, err := message.GetMetadataAsJSON()
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE messages 
		SET text = $2, metadata = $3, edited_at = $4
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, room, "user", text, user_id, room_id, message_type, reply_to_id, edited_at, deleted_at, metadata, created_at`

	var updatedMessage models.Message
	var returnedMetadataBytes []byte
	
	err = r.db.QueryRowxContext(ctx, query, message.ID, message.Text, metadataBytes, time.Now()).Scan(
		&updatedMessage.ID, &updatedMessage.Room, &updatedMessage.User, &updatedMessage.Text,
		&updatedMessage.UserID, &updatedMessage.RoomID, &updatedMessage.MessageType,
		&updatedMessage.ReplyToID, &updatedMessage.EditedAt, &updatedMessage.DeletedAt,
		&returnedMetadataBytes, &updatedMessage.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	// Parse returned metadata
	if returnedMetadataBytes != nil {
		if err := updatedMessage.SetMetadataFromJSON(returnedMetadataBytes); err != nil {
			return nil, err
		}
	}

	return &updatedMessage, nil
}

// DeleteMessage soft deletes a message
func (r *postgresMessageRepository) DeleteMessage(ctx context.Context, id string) error {
	query := `UPDATE messages SET deleted_at = $2 WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

// SearchMessages searches for messages
func (r *postgresMessageRepository) SearchMessages(ctx context.Context, query string, roomID *string, limit, offset int) ([]*models.Message, error) {
	sqlQuery := `
		SELECT id, room, "user", text, user_id, room_id, message_type, reply_to_id, edited_at, deleted_at, metadata, created_at 
		FROM messages 
		WHERE deleted_at IS NULL AND text ILIKE $1`
	
	args := []interface{}{"%" + query + "%"}
	argPos := 2
	
	if roomID != nil {
		sqlQuery += " AND room_id = $" + fmt.Sprintf("%d", argPos)
		args = append(args, *roomID)
		argPos++
	}
	
	sqlQuery += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", argPos) + " OFFSET $" + fmt.Sprintf("%d", argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryxContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var message models.Message
		var metadataBytes []byte
		
		err := rows.Scan(
			&message.ID, &message.Room, &message.User, &message.Text,
			&message.UserID, &message.RoomID, &message.MessageType,
			&message.ReplyToID, &message.EditedAt, &message.DeletedAt,
			&metadataBytes, &message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse metadata if present
		if metadataBytes != nil {
			if err := message.SetMetadataFromJSON(metadataBytes); err != nil {
				return nil, err
			}
		}
		
		messages = append(messages, &message)
	}

	return messages, rows.Err()
}

// GetMessageReactions gets reactions for a message
func (r *postgresMessageRepository) GetMessageReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error) {
	query := `
		SELECT r.id, r.message_id, r.user_id, r.emoji, r.created_at,
		       u.id, u.username, u.display_name, u.avatar_url
		FROM message_reactions r
		JOIN users u ON r.user_id = u.id
		WHERE r.message_id = $1
		ORDER BY r.created_at ASC`

	rows, err := r.db.QueryxContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []*models.MessageReaction
	for rows.Next() {
		var reaction models.MessageReaction
		var user models.User
		
		err := rows.Scan(
			&reaction.ID, &reaction.MessageID, &reaction.UserID, &reaction.Emoji, &reaction.CreatedAt,
			&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		
		reaction.User = &user
		reactions = append(reactions, &reaction)
	}

	return reactions, rows.Err()
}

// AddMessageReaction adds a reaction to a message
func (r *postgresMessageRepository) AddMessageReaction(ctx context.Context, reaction *models.MessageReaction) (*models.MessageReaction, error) {
	// Generate UUID if not provided
	if reaction.ID == "" {
		reaction.ID = uuid.New().String()
	}
	
	// Set created time if not provided
	if reaction.CreatedAt.IsZero() {
		reaction.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO message_reactions (id, message_id, user_id, emoji, created_at) 
		VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
		RETURNING id, message_id, user_id, emoji, created_at`

	var savedReaction models.MessageReaction
	err := r.db.QueryRowxContext(ctx, query,
		reaction.ID, reaction.MessageID, reaction.UserID, reaction.Emoji, reaction.CreatedAt).StructScan(&savedReaction)

	if err != nil {
		return nil, err
	}

	return &savedReaction, nil
}

// RemoveMessageReaction removes a reaction from a message
func (r *postgresMessageRepository) RemoveMessageReaction(ctx context.Context, messageID, userID, emoji string) error {
	query := `DELETE FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3`
	
	_, err := r.db.ExecContext(ctx, query, messageID, userID, emoji)
	return err
}

// GetReactionCounts gets reaction counts for a message
func (r *postgresMessageRepository) GetReactionCounts(ctx context.Context, messageID string) (map[string]int, error) {
	query := `
		SELECT emoji, COUNT(*) as count
		FROM message_reactions 
		WHERE message_id = $1
		GROUP BY emoji`

	rows, err := r.db.QueryxContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var emoji string
		var count int
		
		err := rows.Scan(&emoji, &count)
		if err != nil {
			return nil, err
		}
		
		counts[emoji] = count
	}

	return counts, rows.Err()
}