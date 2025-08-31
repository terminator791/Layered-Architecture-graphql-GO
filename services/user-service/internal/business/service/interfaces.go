package service

import (
	"context"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/domain/models"
)

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, input *models.CreateUserInput) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}