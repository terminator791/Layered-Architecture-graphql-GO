package database

import (
	"fmt"

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
	// Create messages table
	createMessagesTable := `
		CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY,
			room VARCHAR(50) NOT NULL,
			"user" VARCHAR(50) NOT NULL,
			text TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_messages_room ON messages(room);
		CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	`

	_, err := db.Exec(createMessagesTable)
	if err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
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