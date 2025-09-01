package resolvers

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/business/service"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/infrastructure/auth"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	MessageService service.MessageService
	UserService    service.UserService
	RoomService    service.RoomService
	JWTService     *auth.JWTService
}

// Helper function to get user ID from context
func (r *Resolver) getUserIDFromContext(ctx context.Context) (string, error) {
	// Get the GraphQL operation context
	gc := graphql.GetOperationContext(ctx)

	// Get Authorization header
	authHeader := gc.Headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header not found")
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]

	// Validate token and extract user ID
	claims, err := r.JWTService.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	return claims.UserID, nil
}
