# 🏗️ Architecture Deep Dive

This document provides a comprehensive understanding of the layered architecture implementation in this real-time chat application. It serves as both a learning resource and reference guide for understanding clean architecture principles in Go.

## 📋 Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Layer-by-Layer Analysis](#layer-by-layer-analysis)
3. [Design Patterns](#design-patterns)
4. [Data Flow](#data-flow)
5. [Dependency Management](#dependency-management)
6. [Scalability Considerations](#scalability-considerations)
7. [Best Practices Demonstrated](#best-practices-demonstrated)

## 🎯 Architecture Overview

### The Four-Layer Architecture

This application implements a strict layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                   📱 Presentation Layer                     │
│              (GraphQL Schema & Resolvers)                  │
│  • GraphQL schema definition                               │
│  • Resolver implementations                                │
│  • Input validation                                        │
│  • Response formatting                                     │
├─────────────────────────────────────────────────────────────┤
│                   🧠 Business Logic Layer                   │
│              (Services & Domain Logic)                     │
│  • Core business rules                                     │
│  • Service orchestration                                   │
│  • Domain validation                                       │
│  • Use case implementation                                 │
├─────────────────────────────────────────────────────────────┤
│                  💾 Persistence Layer                       │
│            (Repository Pattern & Interfaces)               │
│  • Data access abstraction                                 │
│  • Repository implementations                              │
│  • Query optimization                                      │
│  • Database-specific logic                                 │
├─────────────────────────────────────────────────────────────┤
│                 🔧 Infrastructure Layer                     │
│         (Database, Redis, JWT, External Services)          │
│  • Database connections                                     │
│  • External service integrations                           │
│  • Configuration management                                │
│  • Technical implementations                               │
└─────────────────────────────────────────────────────────────┘
```

### 🎯 Layer Principles

#### 1. **Dependency Rule**
- Higher layers depend on lower layers
- Lower layers never depend on higher layers
- Dependencies point inward (toward the domain)

#### 2. **Single Responsibility**
- Each layer has one primary concern
- Clear boundaries between layers
- Cohesive functionality within layers

#### 3. **Interface Segregation**
- Layers communicate through interfaces
- Enables testability and flexibility
- Supports dependency injection

## 🔍 Layer-by-Layer Analysis

### 📱 Presentation Layer (`internal/presentation/`)

**Purpose**: Handle external communication and user interface concerns

**Components**:
```
presentation/
├── graphql/
│   ├── generated/     # Auto-generated GraphQL code
│   ├── resolvers/     # GraphQL resolver implementations
│   └── schema/        # GraphQL schema definitions
```

**Responsibilities**:
- **Input Validation**: Basic GraphQL schema validation
- **Authentication**: JWT token verification
- **Response Formatting**: Converting domain models to GraphQL types
- **Error Handling**: Translating service errors to GraphQL errors
- **API Contracts**: Defining the external API structure

**Key Files**:
- `schema.graphql`: Complete API definition
- `resolver.go`: Main resolver struct with dependency injection
- Individual resolver files for each domain (messages, users, rooms)

**Example Resolver Pattern**:
```go
func (r *mutationResolver) SendMessage(ctx context.Context, input model.SendMessageInput) (*model.Message, error) {
    // 1. Extract user from context (authentication)
    userID := auth.GetUserIDFromContext(ctx)
    
    // 2. Convert GraphQL input to domain model
    domainInput := &models.CreateMessageInput{
        Room: input.Room,
        User: input.User,
        Text: input.Text,
    }
    
    // 3. Call business layer
    message, err := r.MessageService.CreateMessage(ctx, domainInput)
    if err != nil {
        return nil, fmt.Errorf("failed to create message: %w", err)
    }
    
    // 4. Convert domain model to GraphQL response
    return convertMessageToGraphQL(message), nil
}
```

### 🧠 Business Logic Layer (`internal/business/`)

**Purpose**: Implement core business rules and orchestrate operations

**Components**:
```
business/
└── service/
    ├── interfaces.go        # Service interfaces
    ├── message_service.go   # Message business logic
    ├── user_service.go      # User business logic
    ├── room_service.go      # Room business logic
    └── *_test.go           # Comprehensive unit tests
```

**Responsibilities**:
- **Business Rules Enforcement**: User permissions, room access controls
- **Workflow Orchestration**: Coordinating multiple repository calls
- **Domain Validation**: Ensuring data integrity and business constraints
- **Transaction Management**: Handling complex multi-step operations
- **Real-time Event Publishing**: Publishing events for live updates

**Service Interface Pattern**:
```go
type MessageService interface {
    // Core message operations
    CreateMessage(ctx context.Context, input *models.CreateMessageInput) (*models.Message, error)
    SendMessageToRoom(ctx context.Context, userID, roomID, text string, messageType *models.MessageType, replyToID *string, metadata *models.MessageMetadata) (*models.Message, error)
    
    // Enhanced operations
    UpdateMessage(ctx context.Context, userID, messageID string, input *models.UpdateMessageInput) (*models.Message, error)
    DeleteMessage(ctx context.Context, userID, messageID string) error
    SearchMessages(ctx context.Context, query string, roomID *string, limit, offset int) ([]*models.Message, error)
    
    // Real-time features
    SubscribeToRoomMessages(ctx context.Context, roomID string) (<-chan *models.Message, error)
    
    // Reactions and interactions
    AddReaction(ctx context.Context, userID string, input *models.AddReactionInput) (*models.MessageReaction, error)
    RemoveReaction(ctx context.Context, userID string, input *models.RemoveReactionInput) error
}
```

**Business Logic Examples**:

1. **Permission Checking**:
```go
func (s *messageService) SendMessageToRoom(ctx context.Context, userID, roomID, text string, ...) (*models.Message, error) {
    // Verify user exists and can access room
    member, err := s.roomRepo.GetRoomMember(ctx, roomID, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to check room membership: %w", err)
    }
    if member == nil {
        return nil, fmt.Errorf("user is not a member of this room")
    }
    // Continue with message creation...
}
```

2. **Complex Workflow Orchestration**:
```go
func (s *messageService) UpdateMessage(ctx context.Context, userID, messageID string, input *models.UpdateMessageInput) (*models.Message, error) {
    // 1. Get existing message
    message, err := s.repo.GetMessageByID(ctx, messageID)
    if err != nil {
        return nil, fmt.Errorf("failed to get message: %w", err)
    }
    
    // 2. Check ownership
    if message.UserID == nil || *message.UserID != userID {
        return nil, fmt.Errorf("user does not have permission to update this message")
    }
    
    // 3. Validate business rules
    if message.DeletedAt != nil {
        return nil, fmt.Errorf("cannot update deleted message")
    }
    
    // 4. Update and persist
    updatedMessage, err := s.repo.UpdateMessage(ctx, message)
    if err != nil {
        return nil, fmt.Errorf("failed to update message: %w", err)
    }
    
    // 5. Publish real-time event
    if message.RoomID != nil {
        if err := s.publisher.PublishMessage(ctx, *message.RoomID, updatedMessage); err != nil {
            // Log but don't fail the operation
            fmt.Printf("Warning: failed to publish message update: %v\n", err)
        }
    }
    
    return updatedMessage, nil
}
```

### 💾 Persistence Layer (`internal/persistence/`)

**Purpose**: Abstract data access and provide consistent interfaces

**Components**:
```
persistence/
└── repository/
    ├── interfaces.go                    # Repository interfaces
    ├── postgres_message_repository.go   # PostgreSQL message implementation
    ├── postgres_room_repository.go      # PostgreSQL room implementation
    ├── postgres_user_repository.go      # PostgreSQL user implementation
    └── postgres_typing_repository.go    # PostgreSQL typing indicators
```

**Repository Pattern Benefits**:
1. **Data Source Abstraction**: Easy to switch between databases
2. **Testability**: Mock repositories for unit testing
3. **Consistency**: Standardized data access patterns
4. **Query Optimization**: Database-specific optimizations

**Interface Design**:
```go
type MessageRepository interface {
    // Basic CRUD operations
    CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error)
    GetMessageByID(ctx context.Context, id string) (*models.Message, error)
    UpdateMessage(ctx context.Context, message *models.Message) (*models.Message, error)
    DeleteMessage(ctx context.Context, id string) error
    
    // Advanced querying
    GetMessagesByRoomID(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error)
    SearchMessages(ctx context.Context, query string, roomID *string, limit, offset int) ([]*models.Message, error)
    
    // Reactions
    AddMessageReaction(ctx context.Context, reaction *models.MessageReaction) (*models.MessageReaction, error)
    RemoveMessageReaction(ctx context.Context, messageID, userID, emoji string) error
    GetMessageReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error)
    GetReactionCounts(ctx context.Context, messageID string) (map[string]int, error)
}
```

**Implementation Example**:
```go
func (r *postgresMessageRepository) CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error) {
    query := `
        INSERT INTO messages (id, room, "user", text, created_at) 
        VALUES ($1, $2, $3, $4, $5) 
        RETURNING id, room, "user", text, created_at`

    var savedMessage models.Message
    err := r.db.QueryRowxContext(ctx, query, 
        message.ID, message.Room, message.User, message.Text, message.CreatedAt).StructScan(&savedMessage)
    
    if err != nil {
        return nil, fmt.Errorf("failed to insert message: %w", err)
    }

    return &savedMessage, nil
}
```

### 🔧 Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: Handle technical concerns and external dependencies

**Components**:
```
infrastructure/
├── auth/        # JWT token management
├── database/    # Database connection and configuration
└── redis/       # Redis pub/sub implementation
```

**Key Responsibilities**:
- **Database Management**: Connection pooling, migrations, health checks
- **Authentication**: JWT token generation, validation, middleware
- **Real-time Communication**: Redis pub/sub for WebSocket events
- **Configuration**: Environment-based configuration management

**Redis Pub/Sub Implementation**:
```go
type Publisher interface {
    PublishMessage(ctx context.Context, room string, message *models.Message) error
    Close() error
}

type redisPublisher struct {
    client *redis.Client
}

func (p *redisPublisher) PublishMessage(ctx context.Context, room string, message *models.Message) error {
    data, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }
    
    channel := fmt.Sprintf("room:%s:messages", room)
    return p.client.Publish(ctx, channel, data).Err()
}
```

## 🎨 Design Patterns

### 1. Repository Pattern

**Intent**: Encapsulate data access logic and provide a uniform interface

**Implementation**:
- Interface-based contracts in `repository/interfaces.go`
- PostgreSQL-specific implementations
- Easy mocking for testing

**Benefits**:
- Database independence
- Testability
- Consistency
- Query optimization

### 2. Dependency Injection

**Intent**: Invert dependencies to enable testability and flexibility

**Implementation**:
```go
type messageService struct {
    repo           repository.MessageRepository
    roomRepo       repository.RoomRepository
    userRepo       repository.UserRepository
    typingRepo     repository.TypingRepository
    publisher      redis.Publisher
    subscriber     redis.Subscriber
}

func NewMessageService(
    repo repository.MessageRepository,
    roomRepo repository.RoomRepository,
    userRepo repository.UserRepository,
    typingRepo repository.TypingRepository,
    publisher redis.Publisher,
    subscriber redis.Subscriber,
) MessageService {
    return &messageService{
        repo:       repo,
        roomRepo:   roomRepo,
        userRepo:   userRepo,
        typingRepo: typingRepo,
        publisher:  publisher,
        subscriber: subscriber,
    }
}
```

### 3. Observer Pattern (Pub/Sub)

**Intent**: Enable real-time communication without tight coupling

**Implementation**:
- Redis channels for message broadcasting
- GraphQL subscriptions for client notifications
- Event-driven architecture

### 4. Factory Pattern

**Intent**: Create repository instances with proper configuration

**Implementation**:
```go
func NewPostgresMessageRepository(db *sqlx.DB) MessageRepository {
    return &postgresMessageRepository{
        db: db,
    }
}
```

## 🔄 Data Flow

### Message Creation Flow

```
1. GraphQL Request
   ↓
2. Resolver (Presentation)
   ├── Validate input
   ├── Extract user context
   └── Call service
   ↓
3. Service (Business)
   ├── Validate business rules
   ├── Check permissions
   ├── Create domain model
   └── Call repository
   ↓
4. Repository (Persistence)
   ├── Execute SQL query
   ├── Handle database errors
   └── Return domain model
   ↓
5. Service (Business)
   ├── Publish real-time event
   └── Return to resolver
   ↓
6. Resolver (Presentation)
   ├── Convert to GraphQL type
   └── Return response
```

### Real-time Event Flow

```
1. Message Created
   ↓
2. Service publishes to Redis
   ↓
3. Redis broadcasts to subscribers
   ↓
4. WebSocket connections receive event
   ↓
5. Client UI updates automatically
```

## 🔗 Dependency Management

### Dependency Direction

```
Presentation Layer (GraphQL)
    ↓ depends on
Business Layer (Services)
    ↓ depends on
Persistence Layer (Repositories)
    ↓ depends on
Infrastructure Layer (Database, Redis)
```

### Interface-based Communication

```go
// Business layer defines what it needs
type MessageRepository interface {
    CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error)
}

// Infrastructure layer implements what business needs
type postgresMessageRepository struct {
    db *sqlx.DB
}

func (r *postgresMessageRepository) CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error) {
    // Implementation details...
}
```

## 📈 Scalability Considerations

### Horizontal Scaling

1. **Stateless Services**: All business logic is stateless
2. **Database Connection Pooling**: Efficient database usage
3. **Redis Clustering**: Scalable real-time messaging
4. **Load Balancing**: Multiple service instances

### Performance Optimizations

1. **Database Indexing**: Optimized queries for frequent operations
2. **Connection Pooling**: Reuse database connections
3. **Pagination**: Limit data transfer with offset/limit patterns
4. **Caching**: Redis for frequently accessed data

### Microservice Preparation

The layered architecture naturally supports microservice decomposition:

- **User Service**: User management and authentication
- **Room Service**: Room creation and membership
- **Message Service**: Message handling and real-time features
- **Notification Service**: Push notifications and alerts

## ✨ Best Practices Demonstrated

### 1. Error Handling

```go
func (s *messageService) CreateMessage(ctx context.Context, input *models.CreateMessageInput) (*models.Message, error) {
    // Validate input
    if err := s.validateCreateMessageInput(input); err != nil {
        return nil, err // Return validation error directly
    }

    // Call repository
    savedMessage, err := s.repo.CreateMessage(ctx, message)
    if err != nil {
        return nil, fmt.Errorf("failed to save message: %w", err) // Wrap with context
    }

    return savedMessage, nil
}
```

### 2. Testing Strategy

```go
func TestMessageService_CreateMessage_Success(t *testing.T) {
    // Arrange: Set up mocks and test data
    ctx := context.Background()
    mockRepo := &mocks.MockMessageRepository{}
    service := NewMessageService(mockRepo, ...)
    
    // Set expectations
    mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*models.Message")).Return(expectedMessage, nil)
    
    // Act: Execute the operation
    result, err := service.CreateMessage(ctx, input)
    
    // Assert: Verify results and expectations
    assert.NoError(t, err)
    assert.Equal(t, expectedMessage.ID, result.ID)
    mockRepo.AssertExpectations(t)
}
```

### 3. Configuration Management

```go
type Config struct {
    Database DatabaseConfig `json:"database"`
    Redis    RedisConfig    `json:"redis"`
    JWT      JWTConfig      `json:"jwt"`
    Server   ServerConfig   `json:"server"`
}

func LoadConfig() (*Config, error) {
    // Load from environment variables with defaults
    // Support multiple configuration sources
    // Validate configuration completeness
}
```

### 4. Graceful Shutdown

```go
func main() {
    // Set up signal handling
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    // Start server in goroutine
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()

    // Wait for shutdown signal
    <-c
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
    }
}
```

## 🎓 Learning Outcomes

By studying this architecture, developers will understand:

1. **Clean Architecture Principles**: Separation of concerns and dependency inversion
2. **Domain-Driven Design**: Rich domain models and business logic encapsulation
3. **Testable Code**: Dependency injection and interface-based design
4. **Real-time Systems**: WebSocket subscriptions and event-driven architecture
5. **Database Design**: Efficient queries, indexing, and repository patterns
6. **API Design**: GraphQL schema design and resolver implementation
7. **Error Handling**: Proper error propagation and user-friendly messages
8. **Performance**: Caching, pagination, and optimization strategies

This architecture serves as a blueprint for building scalable, maintainable, and testable applications in Go while demonstrating industry best practices and modern development patterns.