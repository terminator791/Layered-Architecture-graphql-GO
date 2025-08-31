package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/business/service"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/config"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/infrastructure/database"
	redisInfra "github.com/terminator791/Layered-Architecture-graphql-GO/internal/infrastructure/redis"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/persistence/repository"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/presentation/graphql/generated"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/presentation/graphql/resolvers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseConnection(db)

	// Create database tables
	if err := database.CreateTables(db); err != nil {
		log.Fatalf("Failed to create database tables: %v", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.RedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize repositories
	messageRepo := repository.NewPostgresMessageRepository(db)
	userRepo := repository.NewPostgresUserRepository(db)
	roomRepo := repository.NewPostgresRoomRepository(db)
	typingRepo := repository.NewPostgresTypingRepository(db)
	
	// Initialize Redis pub/sub
	redisPublisher := redisInfra.NewRedisPublisher(redisClient)
	redisSubscriber := redisInfra.NewRedisSubscriber(redisClient)
	
	// Initialize services
	messageService := service.NewMessageService(messageRepo, roomRepo, userRepo, typingRepo, redisPublisher, redisSubscriber)

	// Initialize GraphQL resolver
	resolver := &resolvers.Resolver{
		MessageService: messageService,
	}

	// Create GraphQL server
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	// Add WebSocket support for subscriptions
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in development
				// In production, you should implement proper CORS validation
				return true
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Setup HTTP routes
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on http://localhost%s", serverAddr)
	log.Printf("GraphQL Playground available at http://localhost%s", serverAddr)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: nil,
	}

	// Handle graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}