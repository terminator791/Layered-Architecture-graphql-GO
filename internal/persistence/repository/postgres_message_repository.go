package repository

import (
	"context"
	"database/sql"
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