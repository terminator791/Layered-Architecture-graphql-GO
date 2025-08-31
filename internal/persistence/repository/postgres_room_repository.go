package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

type postgresRoomRepository struct {
	db *sqlx.DB
}

// NewPostgresRoomRepository creates a new PostgreSQL room repository
func NewPostgresRoomRepository(db *sqlx.DB) RoomRepository {
	return &postgresRoomRepository{
		db: db,
	}
}

func (r *postgresRoomRepository) CreateRoom(ctx context.Context, room *models.Room) (*models.Room, error) {
	query := `
		INSERT INTO rooms (id, name, description, room_type, password_hash, creator_id, avatar_url, max_members, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		RETURNING id, name, description, room_type, password_hash, creator_id, avatar_url, max_members, created_at, updated_at`

	var savedRoom models.Room
	err := r.db.QueryRowxContext(ctx, query,
		room.ID, room.Name, room.Description, room.RoomType,
		room.PasswordHash, room.CreatorID, room.AvatarURL,
		room.MaxMembers, room.CreatedAt, room.UpdatedAt).StructScan(&savedRoom)

	if err != nil {
		return nil, err
	}

	return &savedRoom, nil
}

func (r *postgresRoomRepository) GetRoomByID(ctx context.Context, id string) (*models.Room, error) {
	query := `
		SELECT id, name, description, room_type, password_hash, creator_id, avatar_url, max_members, created_at, updated_at 
		FROM rooms 
		WHERE id = $1`

	var room models.Room
	err := r.db.QueryRowxContext(ctx, query, id).StructScan(&room)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &room, nil
}

func (r *postgresRoomRepository) UpdateRoom(ctx context.Context, room *models.Room) (*models.Room, error) {
	query := `
		UPDATE rooms 
		SET name = $2, description = $3, room_type = $4, password_hash = $5, avatar_url = $6, max_members = $7, updated_at = $8
		WHERE id = $1 
		RETURNING id, name, description, room_type, password_hash, creator_id, avatar_url, max_members, created_at, updated_at`

	var updatedRoom models.Room
	err := r.db.QueryRowxContext(ctx, query,
		room.ID, room.Name, room.Description, room.RoomType,
		room.PasswordHash, room.AvatarURL, room.MaxMembers,
		room.UpdatedAt).StructScan(&updatedRoom)

	if err != nil {
		return nil, err
	}

	return &updatedRoom, nil
}

func (r *postgresRoomRepository) DeleteRoom(ctx context.Context, id string) error {
	query := `DELETE FROM rooms WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgresRoomRepository) ListRooms(ctx context.Context, limit, offset int) ([]*models.Room, error) {
	query := `
		SELECT id, name, description, room_type, password_hash, creator_id, avatar_url, max_members, created_at, updated_at 
		FROM rooms 
		WHERE room_type = 'public'
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*models.Room
	for rows.Next() {
		var room models.Room
		if err := rows.StructScan(&room); err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}

	return rooms, rows.Err()
}

func (r *postgresRoomRepository) GetRoomsByUserID(ctx context.Context, userID string) ([]*models.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.room_type, r.password_hash, r.creator_id, r.avatar_url, r.max_members, r.created_at, r.updated_at 
		FROM rooms r
		JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1
		ORDER BY rm.joined_at DESC`

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*models.Room
	for rows.Next() {
		var room models.Room
		if err := rows.StructScan(&room); err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}

	return rooms, rows.Err()
}

func (r *postgresRoomRepository) GetRoomMembers(ctx context.Context, roomID string) ([]*models.RoomMember, error) {
	query := `
		SELECT id, room_id, user_id, role, joined_at, last_read_at 
		FROM room_members 
		WHERE room_id = $1
		ORDER BY joined_at ASC`

	rows, err := r.db.QueryxContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.RoomMember
	for rows.Next() {
		var member models.RoomMember
		if err := rows.StructScan(&member); err != nil {
			return nil, err
		}
		members = append(members, &member)
	}

	return members, rows.Err()
}

func (r *postgresRoomRepository) GetRoomMember(ctx context.Context, roomID, userID string) (*models.RoomMember, error) {
	query := `
		SELECT id, room_id, user_id, role, joined_at, last_read_at 
		FROM room_members 
		WHERE room_id = $1 AND user_id = $2`

	var member models.RoomMember
	err := r.db.QueryRowxContext(ctx, query, roomID, userID).StructScan(&member)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &member, nil
}

func (r *postgresRoomRepository) AddRoomMember(ctx context.Context, member *models.RoomMember) (*models.RoomMember, error) {
	query := `
		INSERT INTO room_members (id, room_id, user_id, role, joined_at, last_read_at) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, room_id, user_id, role, joined_at, last_read_at`

	var savedMember models.RoomMember
	err := r.db.QueryRowxContext(ctx, query,
		member.ID, member.RoomID, member.UserID, member.Role,
		member.JoinedAt, member.LastReadAt).StructScan(&savedMember)

	if err != nil {
		return nil, err
	}

	return &savedMember, nil
}

func (r *postgresRoomRepository) UpdateRoomMember(ctx context.Context, member *models.RoomMember) (*models.RoomMember, error) {
	query := `
		UPDATE room_members 
		SET role = $3, last_read_at = $4
		WHERE room_id = $1 AND user_id = $2 
		RETURNING id, room_id, user_id, role, joined_at, last_read_at`

	var updatedMember models.RoomMember
	err := r.db.QueryRowxContext(ctx, query,
		member.RoomID, member.UserID, member.Role,
		member.LastReadAt).StructScan(&updatedMember)

	if err != nil {
		return nil, err
	}

	return &updatedMember, nil
}

func (r *postgresRoomRepository) RemoveRoomMember(ctx context.Context, roomID, userID string) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	
	_, err := r.db.ExecContext(ctx, query, roomID, userID)
	return err
}

func (r *postgresRoomRepository) GetRoomMemberCount(ctx context.Context, roomID string) (int, error) {
	query := `SELECT COUNT(*) FROM room_members WHERE room_id = $1`
	
	var count int
	err := r.db.QueryRowxContext(ctx, query, roomID).Scan(&count)
	return count, err
}

func (r *postgresRoomRepository) GetOnlineMemberCount(ctx context.Context, roomID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM room_members rm
		JOIN users u ON rm.user_id = u.id
		WHERE rm.room_id = $1 AND u.status = 'online'`
	
	var count int
	err := r.db.QueryRowxContext(ctx, query, roomID).Scan(&count)
	return count, err
}

func (r *postgresRoomRepository) UpdateMemberLastRead(ctx context.Context, roomID, userID string) error {
	query := `UPDATE room_members SET last_read_at = $3 WHERE room_id = $1 AND user_id = $2`
	
	_, err := r.db.ExecContext(ctx, query, roomID, userID, time.Now())
	return err
}

// PostgreSQL typing repository implementation
type postgresTypingRepository struct {
	db *sqlx.DB
}

// NewPostgresTypingRepository creates a new PostgreSQL typing repository
func NewPostgresTypingRepository(db *sqlx.DB) TypingRepository {
	return &postgresTypingRepository{
		db: db,
	}
}

func (r *postgresTypingRepository) StartTyping(ctx context.Context, roomID, userID string) error {
	query := `
		INSERT INTO typing_indicators (id, room_id, user_id, started_at, expires_at) 
		VALUES (gen_random_uuid(), $1, $2, $3, $4) 
		ON CONFLICT (room_id, user_id) 
		DO UPDATE SET started_at = $3, expires_at = $4`

	now := time.Now()
	expiresAt := now.Add(10 * time.Second)
	
	_, err := r.db.ExecContext(ctx, query, roomID, userID, now, expiresAt)
	return err
}

func (r *postgresTypingRepository) StopTyping(ctx context.Context, roomID, userID string) error {
	query := `DELETE FROM typing_indicators WHERE room_id = $1 AND user_id = $2`
	
	_, err := r.db.ExecContext(ctx, query, roomID, userID)
	return err
}

func (r *postgresTypingRepository) GetTypingUsers(ctx context.Context, roomID string) ([]*models.TypingIndicator, error) {
	query := `
		SELECT ti.id, ti.room_id, ti.user_id, ti.started_at, ti.expires_at,
		       u.id, u.username, u.display_name, u.avatar_url
		FROM typing_indicators ti
		JOIN users u ON ti.user_id = u.id
		WHERE ti.room_id = $1 AND ti.expires_at > $2
		ORDER BY ti.started_at ASC`

	rows, err := r.db.QueryxContext(ctx, query, roomID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indicators []*models.TypingIndicator
	for rows.Next() {
		var indicator models.TypingIndicator
		var user models.User
		
		err := rows.Scan(
			&indicator.ID, &indicator.RoomID, &indicator.UserID, &indicator.StartedAt, &indicator.ExpiresAt,
			&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		
		indicator.User = &user
		indicators = append(indicators, &indicator)
	}

	return indicators, rows.Err()
}

func (r *postgresTypingRepository) CleanExpiredTypingIndicators(ctx context.Context) error {
	query := `DELETE FROM typing_indicators WHERE expires_at <= $1`
	
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}