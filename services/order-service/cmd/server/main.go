package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/business/service"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/config"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/grpc"
	httpHandler "github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/http"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/persistence/repository"
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

	// Initialize gRPC client for user service
	userServiceClient, err := grpc.NewUserServiceClient(cfg.UserService.Address)
	if err != nil {
		log.Fatalf("Failed to create user service client: %v", err)
	}
	defer userServiceClient.Close()

	// Initialize repository
	orderRepo := repository.NewPostgresOrderRepository(db)

	// Initialize service
	orderService := service.NewOrderService(orderRepo, userServiceClient)

	// Initialize HTTP handler
	orderHandler := httpHandler.NewOrderHandler(orderService)

	// Setup HTTP routes
	http.HandleFunc("/orders", orderHandler.CreateOrder)
	http.HandleFunc("/orders/get", orderHandler.GetOrder)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Order service listening on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Server.Port), nil); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// createTables creates the necessary database tables
func createTables(db *sqlx.DB) error {
	createOrdersTable := `
		CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			product_id VARCHAR(100) NOT NULL,
			quantity INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
		CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
	`

	_, err := db.Exec(createOrdersTable)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	return nil
}