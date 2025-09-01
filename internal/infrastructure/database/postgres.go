package database

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/config"
)

// NewConnection creates a new database connection
func NewConnection(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CreateTables creates the necessary database tables
func CreateTables(db *sqlx.DB) error {
	// Create users table first (no dependencies)
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			display_name VARCHAR(100),
			avatar_url TEXT,
			bio TEXT,
			status VARCHAR(20) DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'away', 'busy')),
			last_seen_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
		CREATE INDEX IF NOT EXISTS idx_users_last_seen_at ON users(last_seen_at);
	`

	_, err := db.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create rooms table (depends on users)
	createRoomsTable := `
		CREATE TABLE IF NOT EXISTS rooms (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			description TEXT,
			room_type VARCHAR(20) DEFAULT 'public' CHECK (room_type IN ('public', 'private', 'direct')),
			password_hash VARCHAR(255),
			creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			avatar_url VARCHAR(500),
			max_members INTEGER DEFAULT 1000,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_rooms_name ON rooms(name);
		CREATE INDEX IF NOT EXISTS idx_rooms_creator_id ON rooms(creator_id);
		CREATE INDEX IF NOT EXISTS idx_rooms_room_type ON rooms(room_type);
	`

	_, err = db.Exec(createRoomsTable)
	if err != nil {
		return fmt.Errorf("failed to create rooms table: %w", err)
	}

	// Create room_members table (depends on rooms and users)
	createRoomMembersTable := `
		CREATE TABLE IF NOT EXISTS room_members (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'moderator', 'member')),
			joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			last_read_at TIMESTAMP WITH TIME ZONE,
			UNIQUE(room_id, user_id)
		);

		CREATE INDEX IF NOT EXISTS idx_room_members_room_id ON room_members(room_id);
		CREATE INDEX IF NOT EXISTS idx_room_members_user_id ON room_members(user_id);
	`

	_, err = db.Exec(createRoomMembersTable)
	if err != nil {
		return fmt.Errorf("failed to create room_members table: %w", err)
	}

	// Create messages table (depends on users and rooms)
	createMessagesTable := `
		CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
			text TEXT NOT NULL,
			message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file', 'system')),
			reply_to_id UUID,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE,
			-- Legacy fields for backward compatibility
			room VARCHAR(50),
			"user" VARCHAR(50)
		);

		CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id);
		CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);
		CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	`

	_, err = db.Exec(createMessagesTable)
	if err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	// Add foreign key constraint for reply_to_id after table creation
	addReplyToConstraint := `
		ALTER TABLE messages ADD CONSTRAINT fk_messages_reply_to_id 
		FOREIGN KEY (reply_to_id) REFERENCES messages(id) ON DELETE SET NULL;
	`

	_, err = db.Exec(addReplyToConstraint)
	if err != nil {
		// Ignore error if constraint already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to add reply_to_id constraint: %w", err)
		}
	}

	// Add index for reply_to_id
	addReplyToIndex := `CREATE INDEX IF NOT EXISTS idx_messages_reply_to_id ON messages(reply_to_id);`
	_, err = db.Exec(addReplyToIndex)
	if err != nil {
		return fmt.Errorf("failed to create reply_to_id index: %w", err)
	}

	// Create message_reactions table (depends on messages and users)
	createMessageReactionsTable := `
		CREATE TABLE IF NOT EXISTS message_reactions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			emoji VARCHAR(10) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE(message_id, user_id, emoji)
		);

		CREATE INDEX IF NOT EXISTS idx_message_reactions_message_id ON message_reactions(message_id);
		CREATE INDEX IF NOT EXISTS idx_message_reactions_user_id ON message_reactions(user_id);
	`

	_, err = db.Exec(createMessageReactionsTable)
	if err != nil {
		return fmt.Errorf("failed to create message_reactions table: %w", err)
	}

	// Create typing_indicators table (depends on rooms and users)
	createTypingIndicatorsTable := `
		CREATE TABLE IF NOT EXISTS typing_indicators (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() + INTERVAL '10 seconds'),
			UNIQUE(room_id, user_id)
		);

		CREATE INDEX IF NOT EXISTS idx_typing_indicators_room_id ON typing_indicators(room_id);
		CREATE INDEX IF NOT EXISTS idx_typing_indicators_expires_at ON typing_indicators(expires_at);
	`

	_, err = db.Exec(createTypingIndicatorsTable)
	if err != nil {
		return fmt.Errorf("failed to create typing_indicators table: %w", err)
	}

	return nil
}

// CloseConnection closes the database connection
func CloseConnection(db *sqlx.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}