package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
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
	query := `
		INSERT INTO users (id, username, email, password_hash, display_name, avatar_url, bio, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		RETURNING id, username, email, password_hash, display_name, avatar_url, bio, status, last_seen_at, created_at, updated_at`

	var savedUser models.User
	err := r.db.QueryRowxContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.DisplayName, user.AvatarURL, user.Bio, user.Status,
		user.CreatedAt, user.UpdatedAt).StructScan(&savedUser)

	if err != nil {
		return nil, err
	}

	return &savedUser, nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, bio, status, last_seen_at, created_at, updated_at 
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

func (r *postgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, bio, status, last_seen_at, created_at, updated_at 
		FROM users 
		WHERE username = $1`

	var user models.User
	err := r.db.QueryRowxContext(ctx, query, username).StructScan(&user)
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
		SELECT id, username, email, password_hash, display_name, avatar_url, bio, status, last_seen_at, created_at, updated_at 
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

func (r *postgresUserRepository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		UPDATE users 
		SET username = $2, email = $3, password_hash = $4, display_name = $5, avatar_url = $6, bio = $7, status = $8, updated_at = $9
		WHERE id = $1 
		RETURNING id, username, email, password_hash, display_name, avatar_url, bio, status, last_seen_at, created_at, updated_at`

	var updatedUser models.User
	err := r.db.QueryRowxContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.DisplayName, user.AvatarURL, user.Bio, user.Status,
		user.UpdatedAt).StructScan(&updatedUser)

	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

func (r *postgresUserRepository) UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error {
	query := `UPDATE users SET status = $2, updated_at = $3 WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID, status, time.Now())
	return err
}

func (r *postgresUserRepository) UpdateUserLastSeen(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_seen_at = $2, updated_at = $3 WHERE id = $1`
	
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, now, now)
	return err
}

func (r *postgresUserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgresUserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, bio, status, last_seen_at, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.StructScan(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

func (r *postgresUserRepository) GetUsersByIDs(ctx context.Context, ids []string) ([]*models.User, error) {
	if len(ids) == 0 {
		return []*models.User{}, nil
	}

	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, bio, status, last_seen_at, created_at, updated_at 
		FROM users 
		WHERE id = ANY($1)`

	rows, err := r.db.QueryxContext(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.StructScan(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}