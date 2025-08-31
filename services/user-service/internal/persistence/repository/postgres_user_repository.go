package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/domain/models"
)

type postgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sqlx.DB) UserRepository {
	return &postgresUserRepository{
		db: db,
	}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	
	// Set created time if not provided
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO users (id, email, name, created_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, email, name, created_at`

	var savedUser models.User
	err := r.db.QueryRowxContext(ctx, query, 
		user.ID, user.Email, user.Name, user.CreatedAt).StructScan(&savedUser)
	
	if err != nil {
		return nil, err
	}

	return &savedUser, nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, name, created_at 
		FROM users 
		WHERE id = $1`

	var user models.User
	err := r.db.QueryRowxContext(ctx, query, id).StructScan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, created_at 
		FROM users 
		WHERE email = $1`

	var user models.User
	err := r.db.QueryRowxContext(ctx, query, email).StructScan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}