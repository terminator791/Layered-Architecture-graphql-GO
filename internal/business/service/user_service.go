package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/infrastructure/auth"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/persistence/repository"
)

// UserService defines the interface for user business logic
type UserService interface {
	// Authentication
	Register(ctx context.Context, input *models.CreateUserInput) (*models.AuthPayload, error)
	Login(ctx context.Context, input *models.LoginInput) (*models.AuthPayload, error)
	
	// User management
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetCurrentUser(ctx context.Context, userID string) (*models.User, error)
	UpdateUser(ctx context.Context, userID string, input *models.UpdateUserInput) (*models.User, error)
	UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) (*models.User, error)
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	
	// User presence
	UpdateUserPresence(ctx context.Context, userID string, isOnline bool) error
}

type userService struct {
	userRepo   repository.UserRepository
	jwtService *auth.JWTService
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, jwtService *auth.JWTService) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *userService) Register(ctx context.Context, input *models.CreateUserInput) (*models.AuthPayload, error) {
	// Validate input
	if err := s.validateCreateUserInput(input); err != nil {
		return nil, err
	}

	// Check if username already exists
	existingUser, _ := s.userRepo.GetUserByUsername(ctx, input.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	existingUser, _ = s.userRepo.GetUserByEmail(ctx, input.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  input.DisplayName,
		AvatarURL:    input.AvatarURL,
		Bio:          input.Bio,
		Status:       models.UserStatusOffline,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user
	savedUser, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token
	token, err := s.jwtService.GenerateToken(savedUser)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthPayload{
		Token: token,
		User:  savedUser,
	}, nil
}

func (s *userService) Login(ctx context.Context, input *models.LoginInput) (*models.AuthPayload, error) {
	// Validate input
	if err := s.validateLoginInput(input); err != nil {
		return nil, err
	}

	// Get user by username
	user, err := s.userRepo.GetUserByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}
	if user == nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Update user status to online
	if err := s.userRepo.UpdateUserStatus(ctx, user.ID, models.UserStatusOnline); err != nil {
		// Log warning but don't fail login
		fmt.Printf("Warning: failed to update user status: %v\n", err)
	}

	// Update last seen
	if err := s.userRepo.UpdateUserLastSeen(ctx, user.ID); err != nil {
		// Log warning but don't fail login
		fmt.Printf("Warning: failed to update user last seen: %v\n", err)
	}

	// Generate token
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *userService) GetCurrentUser(ctx context.Context, userID string) (*models.User, error) {
	return s.GetUserByID(ctx, userID)
}

func (s *userService) UpdateUser(ctx context.Context, userID string, input *models.UpdateUserInput) (*models.User, error) {
	if err := s.validateUpdateUserInput(input); err != nil {
		return nil, err
	}

	// Get existing user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	if input.DisplayName != nil {
		user.DisplayName = input.DisplayName
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}
	if input.Bio != nil {
		user.Bio = input.Bio
	}
	if input.Status != nil {
		user.Status = *input.Status
	}
	user.UpdatedAt = time.Now()

	// Save user
	updatedUser, err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (s *userService) UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) (*models.User, error) {
	if err := s.userRepo.UpdateUserStatus(ctx, userID, status); err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	// Update last seen if going offline
	if status == models.UserStatusOffline {
		if err := s.userRepo.UpdateUserLastSeen(ctx, userID); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: failed to update user last seen: %v\n", err)
		}
	}

	return s.GetUserByID(ctx, userID)
}

func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if err := s.userRepo.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	users, err := s.userRepo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

func (s *userService) UpdateUserPresence(ctx context.Context, userID string, isOnline bool) error {
	status := models.UserStatusOffline
	if isOnline {
		status = models.UserStatusOnline
	}

	if err := s.userRepo.UpdateUserStatus(ctx, userID, status); err != nil {
		return fmt.Errorf("failed to update user presence: %w", err)
	}

	if !isOnline {
		if err := s.userRepo.UpdateUserLastSeen(ctx, userID); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: failed to update user last seen: %v\n", err)
		}
	}

	return nil
}

// Validation functions
func (s *userService) validateCreateUserInput(input *models.CreateUserInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.Username) == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(input.Username) < 3 || len(input.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}

	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if len(input.Email) > 255 {
		return fmt.Errorf("email cannot exceed 255 characters")
	}

	if strings.TrimSpace(input.Password) == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(input.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	if input.DisplayName != nil && len(*input.DisplayName) > 100 {
		return fmt.Errorf("display name cannot exceed 100 characters")
	}

	if input.Bio != nil && len(*input.Bio) > 500 {
		return fmt.Errorf("bio cannot exceed 500 characters")
	}

	return nil
}

func (s *userService) validateLoginInput(input *models.LoginInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.Username) == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if strings.TrimSpace(input.Password) == "" {
		return fmt.Errorf("password cannot be empty")
	}

	return nil
}

func (s *userService) validateUpdateUserInput(input *models.UpdateUserInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if input.DisplayName != nil && len(*input.DisplayName) > 100 {
		return fmt.Errorf("display name cannot exceed 100 characters")
	}

	if input.Bio != nil && len(*input.Bio) > 500 {
		return fmt.Errorf("bio cannot exceed 500 characters")
	}

	return nil
}