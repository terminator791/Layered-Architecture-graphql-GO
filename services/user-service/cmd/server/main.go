package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	user "github.com/terminator791/Layered-Architecture-graphql-GO/proto"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/business/service"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/config"
	grpcServer "github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/grpc"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/user-service/internal/persistence/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := sqlx.Connect("postgres", cfg.Database.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Initialize repository
	userRepo := repository.NewPostgresUserRepository(db)

	// Initialize service
	userService := service.NewUserService(userRepo)

	// Initialize gRPC server
	grpcUserServer := grpcServer.NewUserServiceServer(userService)

	// Create gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	user.RegisterUserServiceServer(s, grpcUserServer)

	log.Printf("User service listening on port %s", cfg.Server.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// createTables creates the necessary database tables
func createTables(db *sqlx.DB) error {
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err := db.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}