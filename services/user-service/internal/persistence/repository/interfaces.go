package repository

import (
	"context"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/domain/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}