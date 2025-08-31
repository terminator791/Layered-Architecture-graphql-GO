package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/persistence/repository"
)

// RoomService defines the interface for room business logic
type RoomService interface {
	// Room management
	CreateRoom(ctx context.Context, userID string, input *models.CreateRoomInput) (*models.Room, error)
	GetRoomByID(ctx context.Context, roomID string) (*models.Room, error)
	UpdateRoom(ctx context.Context, userID, roomID string, input *models.UpdateRoomInput) (*models.Room, error)
	DeleteRoom(ctx context.Context, userID, roomID string) error
	ListRooms(ctx context.Context, limit, offset int) ([]*models.Room, error)
	GetUserRooms(ctx context.Context, userID string) ([]*models.Room, error)
	
	// Room membership
	JoinRoom(ctx context.Context, userID string, input *models.JoinRoomInput) (*models.RoomMember, error)
	LeaveRoom(ctx context.Context, userID, roomID string) error
	GetRoomMembers(ctx context.Context, roomID string) ([]*models.RoomMember, error)
	UpdateRoomMember(ctx context.Context, userID string, input *models.UpdateRoomMemberInput) (*models.RoomMember, error)
	KickRoomMember(ctx context.Context, userID, roomID, targetUserID string) error
	
	// Room permissions
	CanUserAccessRoom(ctx context.Context, userID, roomID string) (bool, error)
	CanUserModerateRoom(ctx context.Context, userID, roomID string) (bool, error)
	CanUserAdminRoom(ctx context.Context, userID, roomID string) (bool, error)
}

type roomService struct {
	roomRepo repository.RoomRepository
	userRepo repository.UserRepository
}

// NewRoomService creates a new room service
func NewRoomService(roomRepo repository.RoomRepository, userRepo repository.UserRepository) RoomService {
	return &roomService{
		roomRepo: roomRepo,
		userRepo: userRepo,
	}
}

func (s *roomService) CreateRoom(ctx context.Context, userID string, input *models.CreateRoomInput) (*models.Room, error) {
	// Validate input
	if err := s.validateCreateRoomInput(input); err != nil {
		return nil, err
	}

	// Verify user exists
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Hash password if provided
	var passwordHash *string
	if input.Password != nil && *input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		hashStr := string(hash)
		passwordHash = &hashStr
	}

	maxMembers := 1000
	if input.MaxMembers != nil {
		maxMembers = *input.MaxMembers
	}

	// Create room
	room := &models.Room{
		ID:           uuid.New().String(),
		Name:         input.Name,
		Description:  input.Description,
		RoomType:     input.RoomType,
		PasswordHash: passwordHash,
		CreatorID:    userID,
		AvatarURL:    input.AvatarURL,
		MaxMembers:   maxMembers,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save room
	savedRoom, err := s.roomRepo.CreateRoom(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// Add creator as admin member
	member := &models.RoomMember{
		ID:       uuid.New().String(),
		RoomID:   savedRoom.ID,
		UserID:   userID,
		Role:     models.RoomRoleAdmin,
		JoinedAt: time.Now(),
	}

	_, err = s.roomRepo.AddRoomMember(ctx, member)
	if err != nil {
		// Try to clean up the room if member addition fails
		s.roomRepo.DeleteRoom(ctx, savedRoom.ID)
		return nil, fmt.Errorf("failed to add creator as room member: %w", err)
	}

	// Populate room with creator info
	savedRoom.Creator = user
	savedRoom.MemberCount = 1
	savedRoom.OnlineCount = 0
	if user.Status == models.UserStatusOnline {
		savedRoom.OnlineCount = 1
	}

	return savedRoom, nil
}

func (s *roomService) GetRoomByID(ctx context.Context, roomID string) (*models.Room, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room ID cannot be empty")
	}

	room, err := s.roomRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("room not found")
	}

	// Populate creator info
	if room.CreatorID != "" {
		creator, err := s.userRepo.GetUserByID(ctx, room.CreatorID)
		if err == nil && creator != nil {
			room.Creator = creator
		}
	}

	// Get member and online counts
	memberCount, err := s.roomRepo.GetRoomMemberCount(ctx, roomID)
	if err == nil {
		room.MemberCount = memberCount
	}

	onlineCount, err := s.roomRepo.GetOnlineMemberCount(ctx, roomID)
	if err == nil {
		room.OnlineCount = onlineCount
	}

	return room, nil
}

func (s *roomService) UpdateRoom(ctx context.Context, userID, roomID string, input *models.UpdateRoomInput) (*models.Room, error) {
	if err := s.validateUpdateRoomInput(input); err != nil {
		return nil, err
	}

	// Check if user can admin the room
	canAdmin, err := s.CanUserAdminRoom(ctx, userID, roomID)
	if err != nil {
		return nil, err
	}
	if !canAdmin {
		return nil, fmt.Errorf("user does not have permission to update this room")
	}

	// Get existing room
	room, err := s.roomRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("room not found")
	}

	// Update fields
	if input.Name != nil {
		room.Name = *input.Name
	}
	if input.Description != nil {
		room.Description = input.Description
	}
	if input.AvatarURL != nil {
		room.AvatarURL = input.AvatarURL
	}
	if input.MaxMembers != nil {
		room.MaxMembers = *input.MaxMembers
	}
	room.UpdatedAt = time.Now()

	// Save room
	_, err = s.roomRepo.UpdateRoom(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	return s.GetRoomByID(ctx, roomID) // Return room with populated fields
}

func (s *roomService) DeleteRoom(ctx context.Context, userID, roomID string) error {
	// Check if user can admin the room
	canAdmin, err := s.CanUserAdminRoom(ctx, userID, roomID)
	if err != nil {
		return err
	}
	if !canAdmin {
		return fmt.Errorf("user does not have permission to delete this room")
	}

	if err := s.roomRepo.DeleteRoom(ctx, roomID); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	return nil
}

func (s *roomService) ListRooms(ctx context.Context, limit, offset int) ([]*models.Room, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rooms, err := s.roomRepo.ListRooms(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	// Populate creator info for each room
	for _, room := range rooms {
		if room.CreatorID != "" {
			creator, err := s.userRepo.GetUserByID(ctx, room.CreatorID)
			if err == nil && creator != nil {
				room.Creator = creator
			}
		}

		// Get counts
		memberCount, err := s.roomRepo.GetRoomMemberCount(ctx, room.ID)
		if err == nil {
			room.MemberCount = memberCount
		}

		onlineCount, err := s.roomRepo.GetOnlineMemberCount(ctx, room.ID)
		if err == nil {
			room.OnlineCount = onlineCount
		}
	}

	return rooms, nil
}

func (s *roomService) GetUserRooms(ctx context.Context, userID string) ([]*models.Room, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	rooms, err := s.roomRepo.GetRoomsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}

	// Populate creator info and counts for each room
	for _, room := range rooms {
		if room.CreatorID != "" {
			creator, err := s.userRepo.GetUserByID(ctx, room.CreatorID)
			if err == nil && creator != nil {
				room.Creator = creator
			}
		}

		memberCount, err := s.roomRepo.GetRoomMemberCount(ctx, room.ID)
		if err == nil {
			room.MemberCount = memberCount
		}

		onlineCount, err := s.roomRepo.GetOnlineMemberCount(ctx, room.ID)
		if err == nil {
			room.OnlineCount = onlineCount
		}
	}

	return rooms, nil
}

func (s *roomService) JoinRoom(ctx context.Context, userID string, input *models.JoinRoomInput) (*models.RoomMember, error) {
	// Validate input
	if err := s.validateJoinRoomInput(input); err != nil {
		return nil, err
	}

	// Check if user exists
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get room
	room, err := s.roomRepo.GetRoomByID(ctx, input.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("room not found")
	}

	// Check if already a member
	existingMember, _ := s.roomRepo.GetRoomMember(ctx, input.RoomID, userID)
	if existingMember != nil {
		return existingMember, nil // Already a member
	}

	// Check room capacity
	memberCount, err := s.roomRepo.GetRoomMemberCount(ctx, input.RoomID)
	if err == nil && memberCount >= room.MaxMembers {
		return nil, fmt.Errorf("room is at maximum capacity")
	}

	// Check password for private rooms
	if room.RoomType == models.RoomTypePrivate && room.PasswordHash != nil {
		if input.Password == nil {
			return nil, fmt.Errorf("password required for private room")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*room.PasswordHash), []byte(*input.Password)); err != nil {
			return nil, fmt.Errorf("invalid room password")
		}
	}

	// Add member
	member := &models.RoomMember{
		ID:       uuid.New().String(),
		RoomID:   input.RoomID,
		UserID:   userID,
		Role:     models.RoomRoleMember,
		JoinedAt: time.Now(),
	}

	savedMember, err := s.roomRepo.AddRoomMember(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to join room: %w", err)
	}

	// Populate user info
	savedMember.User = user

	return savedMember, nil
}

func (s *roomService) LeaveRoom(ctx context.Context, userID, roomID string) error {
	if userID == "" || roomID == "" {
		return fmt.Errorf("user ID and room ID cannot be empty")
	}

	// Check if user is a member
	member, err := s.roomRepo.GetRoomMember(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to get room member: %w", err)
	}
	if member == nil {
		return fmt.Errorf("user is not a member of this room")
	}

	// Remove member
	if err := s.roomRepo.RemoveRoomMember(ctx, roomID, userID); err != nil {
		return fmt.Errorf("failed to leave room: %w", err)
	}

	return nil
}

func (s *roomService) GetRoomMembers(ctx context.Context, roomID string) ([]*models.RoomMember, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room ID cannot be empty")
	}

	members, err := s.roomRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room members: %w", err)
	}

	// Populate user info for each member
	userIDs := make([]string, len(members))
	for i, member := range members {
		userIDs[i] = member.UserID
	}

	users, err := s.userRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get member user info: %w", err)
	}

	// Create user map for quick lookup
	userMap := make(map[string]*models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// Populate user info in members
	for _, member := range members {
		if user, exists := userMap[member.UserID]; exists {
			member.User = user
		}
	}

	return members, nil
}

func (s *roomService) UpdateRoomMember(ctx context.Context, userID string, input *models.UpdateRoomMemberInput) (*models.RoomMember, error) {
	if err := s.validateUpdateRoomMemberInput(input); err != nil {
		return nil, err
	}

	// Check if user can moderate the room
	canModerate, err := s.CanUserModerateRoom(ctx, userID, input.RoomID)
	if err != nil {
		return nil, err
	}
	if !canModerate {
		return nil, fmt.Errorf("user does not have permission to update room members")
	}

	// Get existing member
	member, err := s.roomRepo.GetRoomMember(ctx, input.RoomID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room member: %w", err)
	}
	if member == nil {
		return nil, fmt.Errorf("user is not a member of this room")
	}

	// Update role if provided
	if input.Role != nil {
		member.Role = *input.Role
	}

	// Save member
	updatedMember, err := s.roomRepo.UpdateRoomMember(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to update room member: %w", err)
	}

	// Populate user info
	user, err := s.userRepo.GetUserByID(ctx, input.UserID)
	if err == nil && user != nil {
		updatedMember.User = user
	}

	return updatedMember, nil
}

func (s *roomService) KickRoomMember(ctx context.Context, userID, roomID, targetUserID string) error {
	// Check if user can moderate the room
	canModerate, err := s.CanUserModerateRoom(ctx, userID, roomID)
	if err != nil {
		return err
	}
	if !canModerate {
		return fmt.Errorf("user does not have permission to kick room members")
	}

	// Cannot kick yourself
	if userID == targetUserID {
		return fmt.Errorf("cannot kick yourself")
	}

	// Remove member
	if err := s.roomRepo.RemoveRoomMember(ctx, roomID, targetUserID); err != nil {
		return fmt.Errorf("failed to kick room member: %w", err)
	}

	return nil
}

func (s *roomService) CanUserAccessRoom(ctx context.Context, userID, roomID string) (bool, error) {
	member, err := s.roomRepo.GetRoomMember(ctx, roomID, userID)
	if err != nil {
		return false, err
	}
	return member != nil, nil
}

func (s *roomService) CanUserModerateRoom(ctx context.Context, userID, roomID string) (bool, error) {
	member, err := s.roomRepo.GetRoomMember(ctx, roomID, userID)
	if err != nil {
		return false, err
	}
	if member == nil {
		return false, nil
	}
	return member.Role == models.RoomRoleAdmin || member.Role == models.RoomRoleModerator, nil
}

func (s *roomService) CanUserAdminRoom(ctx context.Context, userID, roomID string) (bool, error) {
	member, err := s.roomRepo.GetRoomMember(ctx, roomID, userID)
	if err != nil {
		return false, err
	}
	if member == nil {
		return false, nil
	}
	return member.Role == models.RoomRoleAdmin, nil
}

// Validation functions
func (s *roomService) validateCreateRoomInput(input *models.CreateRoomInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("room name cannot be empty")
	}

	if len(input.Name) > 100 {
		return fmt.Errorf("room name cannot exceed 100 characters")
	}

	if input.Description != nil && len(*input.Description) > 500 {
		return fmt.Errorf("room description cannot exceed 500 characters")
	}

	if input.MaxMembers != nil && (*input.MaxMembers < 1 || *input.MaxMembers > 10000) {
		return fmt.Errorf("max members must be between 1 and 10000")
	}

	return nil
}

func (s *roomService) validateUpdateRoomInput(input *models.UpdateRoomInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return fmt.Errorf("room name cannot be empty")
	}

	if input.Name != nil && len(*input.Name) > 100 {
		return fmt.Errorf("room name cannot exceed 100 characters")
	}

	if input.Description != nil && len(*input.Description) > 500 {
		return fmt.Errorf("room description cannot exceed 500 characters")
	}

	if input.MaxMembers != nil && (*input.MaxMembers < 1 || *input.MaxMembers > 10000) {
		return fmt.Errorf("max members must be between 1 and 10000")
	}

	return nil
}

func (s *roomService) validateJoinRoomInput(input *models.JoinRoomInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.RoomID) == "" {
		return fmt.Errorf("room ID cannot be empty")
	}

	return nil
}

func (s *roomService) validateUpdateRoomMemberInput(input *models.UpdateRoomMemberInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.RoomID) == "" {
		return fmt.Errorf("room ID cannot be empty")
	}

	if strings.TrimSpace(input.UserID) == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	return nil
}