package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/domain/models"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/persistence/repository"
)

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context, input *models.CreateUserInput) (*models.User, error) {
	// Validate input
	if err := s.validateCreateUserInput(input); err != nil {
		return nil, err
	}

	// Check if user with email already exists
	existingUser, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", input.Email)
	}

	// Create user model
	user := &models.User{
		Email: input.Email,
		Name:  input.Name,
	}

	// Save user to database
	savedUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return savedUser, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *userService) validateCreateUserInput(input *models.CreateUserInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Basic email validation
	if !strings.Contains(input.Email, "@") {
		return fmt.Errorf("invalid email format")
	}

	if len(input.Name) > 100 {
		return fmt.Errorf("name too long (max 100 characters)")
	}

	if len(input.Email) > 255 {
		return fmt.Errorf("email too long (max 255 characters)")
	}

	return nil
}