# 🛠️ Development Guide

A comprehensive guide for developing new features in the real-time chat application. This guide follows the layered architecture principles and demonstrates best practices for adding functionality.

## 📋 Table of Contents

1. [Development Setup](#development-setup)
2. [Architecture Guidelines](#architecture-guidelines)
3. [Feature Development Workflow](#feature-development-workflow)
4. [Code Style & Standards](#code-style--standards)
5. [Testing Strategy](#testing-strategy)
6. [Database Management](#database-management)
7. [Performance Considerations](#performance-considerations)
8. [Security Guidelines](#security-guidelines)
9. [Deployment Guide](#deployment-guide)

## 🚀 Development Setup

### Prerequisites

- **Go 1.24+**: Latest stable Go version
- **Docker & Docker Compose**: For local development environment
- **PostgreSQL 15**: Primary database
- **Redis 7**: Real-time messaging and caching
- **Git**: Version control

### Local Environment Setup

```bash
# Clone the repository
git clone https://github.com/terminator791/Layered-Architecture-graphql-GO.git
cd Layered-Architecture-graphql-GO

# Copy environment configuration
cp .env.example .env

# Start services with Docker Compose
docker-compose up -d

# Install Go dependencies
go mod tidy

# Run database migrations
go run cmd/migrate/main.go up

# Start the server
go run cmd/server/main.go
```

### Environment Configuration

Create `.env` file with required variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=chatapp
DB_PASSWORD=password
DB_NAME=chatapp_dev
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRY=24h

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost
GRAPHQL_PLAYGROUND=true

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

### Development Tools

```bash
# Install development tools
go install github.com/99designs/gqlgen@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Generate GraphQL code
go generate ./...

# Run linting
golangci-lint run

# Run tests with coverage
go test -cover ./...
```

## 🏗️ Architecture Guidelines

### Layer Responsibilities

1. **Domain Layer** (`internal/domain/`): Pure business entities
2. **Business Layer** (`internal/business/`): Use cases and business logic
3. **Persistence Layer** (`internal/persistence/`): Data access abstractions
4. **Infrastructure Layer** (`internal/infrastructure/`): External concerns
5. **Presentation Layer** (`internal/presentation/`): API endpoints

### Dependency Flow

```
Presentation → Business → Persistence → Infrastructure
```

**Rules**:
- Higher layers can depend on lower layers
- Lower layers should never depend on higher layers
- Use interfaces to invert dependencies
- Keep domain logic pure and testable

## 📝 Feature Development Workflow

### Step-by-Step Process

Let's implement a **Message Pinning** feature as an example:

#### 1. Define Domain Models

Start with the domain layer - define what a pinned message represents:

```go
// internal/domain/models/message.go

// Add to existing Message struct
type Message struct {
    // ... existing fields
    IsPinned    bool       `json:"isPinned" db:"is_pinned"`
    PinnedAt    *time.Time `json:"pinnedAt" db:"pinned_at"`
    PinnedByID  *string    `json:"pinnedById" db:"pinned_by_id"`
    
    // Populated fields
    PinnedBy    *User      `json:"pinnedBy"`
}

// New input types
type PinMessageInput struct {
    MessageID string `json:"messageId"`
}

type UnpinMessageInput struct {
    MessageID string `json:"messageId"`
}
```

#### 2. Update Repository Interface

Define data access needs in the persistence layer:

```go
// internal/persistence/repository/interfaces.go

type MessageRepository interface {
    // ... existing methods
    
    // Pin/Unpin operations
    PinMessage(ctx context.Context, messageID, userID string) error
    UnpinMessage(ctx context.Context, messageID string) error
    GetPinnedMessages(ctx context.Context, roomID string) ([]*models.Message, error)
    IsPinned(ctx context.Context, messageID string) (bool, error)
}
```

#### 3. Implement Repository

Add concrete implementation for PostgreSQL:

```go
// internal/persistence/repository/postgres_message_repository.go

func (r *postgresMessageRepository) PinMessage(ctx context.Context, messageID, userID string) error {
    query := `
        UPDATE messages 
        SET is_pinned = true, pinned_at = NOW(), pinned_by_id = $2 
        WHERE id = $1 AND deleted_at IS NULL`
    
    result, err := r.db.ExecContext(ctx, query, messageID, userID)
    if err != nil {
        return fmt.Errorf("failed to pin message: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get affected rows: %w", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("message not found or already deleted")
    }
    
    return nil
}

func (r *postgresMessageRepository) UnpinMessage(ctx context.Context, messageID string) error {
    query := `
        UPDATE messages 
        SET is_pinned = false, pinned_at = NULL, pinned_by_id = NULL 
        WHERE id = $1`
    
    _, err := r.db.ExecContext(ctx, query, messageID)
    if err != nil {
        return fmt.Errorf("failed to unpin message: %w", err)
    }
    
    return nil
}

func (r *postgresMessageRepository) GetPinnedMessages(ctx context.Context, roomID string) ([]*models.Message, error) {
    query := `
        SELECT m.id, m.room, m.user, m.text, m.user_id, m.room_id, m.message_type,
               m.reply_to_id, m.edited_at, m.deleted_at, m.metadata, m.created_at,
               m.is_pinned, m.pinned_at, m.pinned_by_id
        FROM messages m
        WHERE m.room_id = $1 AND m.is_pinned = true AND m.deleted_at IS NULL
        ORDER BY m.pinned_at DESC`
    
    rows, err := r.db.QueryxContext(ctx, query, roomID)
    if err != nil {
        return nil, fmt.Errorf("failed to query pinned messages: %w", err)
    }
    defer rows.Close()
    
    var messages []*models.Message
    for rows.Next() {
        var message models.Message
        if err := rows.StructScan(&message); err != nil {
            return nil, fmt.Errorf("failed to scan message: %w", err)
        }
        messages = append(messages, &message)
    }
    
    return messages, nil
}
```

#### 4. Update Service Interface

Define business operations:

```go
// internal/business/service/interfaces.go

type MessageService interface {
    // ... existing methods
    
    // Pin operations
    PinMessage(ctx context.Context, userID, messageID string) error
    UnpinMessage(ctx context.Context, userID, messageID string) error
    GetPinnedMessages(ctx context.Context, roomID string) ([]*models.Message, error)
}
```

#### 5. Implement Business Logic

Add business rules and validations:

```go
// internal/business/service/message_service.go

func (s *messageService) PinMessage(ctx context.Context, userID, messageID string) error {
    // Validate inputs
    if userID == "" || messageID == "" {
        return fmt.Errorf("user ID and message ID cannot be empty")
    }
    
    // Get message to validate
    message, err := s.repo.GetMessageByID(ctx, messageID)
    if err != nil {
        return fmt.Errorf("failed to get message: %w", err)
    }
    if message == nil {
        return fmt.Errorf("message not found")
    }
    
    // Check if message is already deleted
    if message.DeletedAt != nil {
        return fmt.Errorf("cannot pin deleted message")
    }
    
    // Check if user has permission to pin in this room
    if message.RoomID == nil {
        return fmt.Errorf("cannot pin message without room")
    }
    
    member, err := s.roomRepo.GetRoomMember(ctx, *message.RoomID, userID)
    if err != nil {
        return fmt.Errorf("failed to check room membership: %w", err)
    }
    if member == nil {
        return fmt.Errorf("user is not a member of this room")
    }
    
    // Only admins and moderators can pin messages
    if member.Role != models.RoomRoleAdmin && member.Role != models.RoomRoleModerator {
        return fmt.Errorf("insufficient permissions to pin messages")
    }
    
    // Check room pin limit (business rule: max 10 pinned messages per room)
    pinnedMessages, err := s.repo.GetPinnedMessages(ctx, *message.RoomID)
    if err != nil {
        return fmt.Errorf("failed to check pinned messages: %w", err)
    }
    if len(pinnedMessages) >= 10 {
        return fmt.Errorf("room has reached maximum number of pinned messages (10)")
    }
    
    // Pin the message
    if err := s.repo.PinMessage(ctx, messageID, userID); err != nil {
        return fmt.Errorf("failed to pin message: %w", err)
    }
    
    // Publish real-time event for pinned message
    if err := s.publisher.PublishRoomEvent(ctx, *message.RoomID, "message_pinned", map[string]interface{}{
        "messageId": messageID,
        "pinnedBy":  userID,
    }); err != nil {
        // Log but don't fail the operation
        fmt.Printf("Warning: failed to publish pin event: %v\n", err)
    }
    
    return nil
}

func (s *messageService) UnpinMessage(ctx context.Context, userID, messageID string) error {
    // Similar validation logic for unpinning
    // ... implementation
}

func (s *messageService) GetPinnedMessages(ctx context.Context, roomID string) ([]*models.Message, error) {
    if roomID == "" {
        return nil, fmt.Errorf("room ID cannot be empty")
    }
    
    messages, err := s.repo.GetPinnedMessages(ctx, roomID)
    if err != nil {
        return nil, fmt.Errorf("failed to get pinned messages: %w", err)
    }
    
    // Populate user info for messages
    return s.populateMessageUserInfo(ctx, messages)
}
```

#### 6. Update GraphQL Schema

Define API operations:

```graphql
# internal/presentation/graphql/schema/schema.graphql

type Message {
  # ... existing fields
  isPinned: Boolean!
  pinnedAt: Time
  pinnedBy: User
}

input PinMessageInput {
  messageId: ID!
}

input UnpinMessageInput {
  messageId: ID!
}

type Mutation {
  # ... existing mutations
  pinMessage(input: PinMessageInput!): Boolean!
  unpinMessage(input: UnpinMessageInput!): Boolean!
}

type Query {
  # ... existing queries
  pinnedMessages(roomId: ID!): [Message!]!
}

type Subscription {
  # ... existing subscriptions
  messagePinned(roomId: ID!): Message!
  messageUnpinned(roomId: ID!): Message!
}
```

#### 7. Implement GraphQL Resolvers

Connect API to business logic:

```go
// internal/presentation/graphql/resolvers/message.go

func (r *mutationResolver) PinMessage(ctx context.Context, input model.PinMessageInput) (bool, error) {
    // Get user from context
    userID := auth.GetUserIDFromContext(ctx)
    if userID == "" {
        return false, fmt.Errorf("authentication required")
    }
    
    // Call service
    if err := r.MessageService.PinMessage(ctx, userID, input.MessageID); err != nil {
        return false, fmt.Errorf("failed to pin message: %w", err)
    }
    
    return true, nil
}

func (r *mutationResolver) UnpinMessage(ctx context.Context, input model.UnpinMessageInput) (bool, error) {
    userID := auth.GetUserIDFromContext(ctx)
    if userID == "" {
        return false, fmt.Errorf("authentication required")
    }
    
    if err := r.MessageService.UnpinMessage(ctx, userID, input.MessageID); err != nil {
        return false, fmt.Errorf("failed to unpin message: %w", err)
    }
    
    return true, nil
}

func (r *queryResolver) PinnedMessages(ctx context.Context, roomID string) ([]*model.Message, error) {
    messages, err := r.MessageService.GetPinnedMessages(ctx, roomID)
    if err != nil {
        return nil, fmt.Errorf("failed to get pinned messages: %w", err)
    }
    
    // Convert domain models to GraphQL models
    return convertMessagesToGraphQL(messages), nil
}
```

#### 8. Create Database Migration

Add database schema changes:

```sql
-- migrations/000010_add_message_pinning.up.sql

ALTER TABLE messages 
ADD COLUMN is_pinned BOOLEAN DEFAULT FALSE,
ADD COLUMN pinned_at TIMESTAMP,
ADD COLUMN pinned_by_id VARCHAR(255);

-- Add index for efficient pinned message queries
CREATE INDEX idx_messages_pinned ON messages(room_id, is_pinned, pinned_at DESC) 
WHERE is_pinned = TRUE AND deleted_at IS NULL;

-- Add foreign key constraint
ALTER TABLE messages 
ADD CONSTRAINT fk_messages_pinned_by 
FOREIGN KEY (pinned_by_id) REFERENCES users(id);
```

```sql
-- migrations/000010_add_message_pinning.down.sql

-- Remove foreign key constraint
ALTER TABLE messages DROP CONSTRAINT IF EXISTS fk_messages_pinned_by;

-- Remove index
DROP INDEX IF EXISTS idx_messages_pinned;

-- Remove columns
ALTER TABLE messages 
DROP COLUMN IF EXISTS is_pinned,
DROP COLUMN IF EXISTS pinned_at,
DROP COLUMN IF EXISTS pinned_by_id;
```

#### 9. Write Comprehensive Tests

Test each layer thoroughly:

```go
// internal/business/service/message_service_test.go

func TestMessageService_PinMessage_Success(t *testing.T) {
    // Arrange
    ctx := context.Background()
    mockRepo := &mocks.MockMessageRepository{}
    mockRoomRepo := &mocks.MockRoomRepository{}
    mockUserRepo := &mocks.MockUserRepository{}
    mockTypingRepo := &mocks.MockTypingRepository{}
    mockPublisher := &mocks.MockPublisher{}
    mockSubscriber := &mocks.MockSubscriber{}
    
    service := NewMessageService(mockRepo, mockRoomRepo, mockUserRepo, mockTypingRepo, mockPublisher, mockSubscriber)
    
    userID := "user_123"
    messageID := "message_456"
    roomID := "room_789"
    
    message := &models.Message{
        ID:     messageID,
        RoomID: &roomID,
        Text:   "Test message",
    }
    
    member := &models.RoomMember{
        UserID: userID,
        RoomID: roomID,
        Role:   models.RoomRoleAdmin,
    }
    
    // Mock expectations
    mockRepo.On("GetMessageByID", ctx, messageID).Return(message, nil)
    mockRoomRepo.On("GetRoomMember", ctx, roomID, userID).Return(member, nil)
    mockRepo.On("GetPinnedMessages", ctx, roomID).Return([]*models.Message{}, nil)
    mockRepo.On("PinMessage", ctx, messageID, userID).Return(nil)
    mockPublisher.On("PublishRoomEvent", ctx, roomID, "message_pinned", mock.Anything).Return(nil)
    
    // Act
    err := service.PinMessage(ctx, userID, messageID)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
    mockRoomRepo.AssertExpectations(t)
    mockPublisher.AssertExpectations(t)
}

func TestMessageService_PinMessage_InsufficientPermissions(t *testing.T) {
    // Test permission validation
    // ... implementation
}

func TestMessageService_PinMessage_MaxPinsReached(t *testing.T) {
    // Test business rule validation
    // ... implementation
}
```

#### 10. Update Documentation

Document the new API operations:

```markdown
### Pin/Unpin Messages

#### Pin Message
Pin an important message to the top of the room.

```graphql
mutation PinMessage {
  pinMessage(input: {
    messageId: "message_123"
  })
}
```

#### Get Pinned Messages
Retrieve all pinned messages for a room.

```graphql
query PinnedMessages($roomId: ID!) {
  pinnedMessages(roomId: $roomId) {
    id
    text
    isPinned
    pinnedAt
    pinnedBy {
      username
      displayName
    }
    userInfo {
      username
      displayName
    }
  }
}
```
```

## 🎨 Code Style & Standards

### Go Code Style

Follow standard Go conventions:

```go
// Package documentation
// Package service implements business logic for the chat application.
package service

import (
    "context"
    "fmt"
    "time"
    
    "github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

// Service interface documentation
// MessageService defines operations for message management.
type MessageService interface {
    // CreateMessage creates a new message in the system.
    // It validates the input, checks permissions, and publishes real-time events.
    CreateMessage(ctx context.Context, input *models.CreateMessageInput) (*models.Message, error)
}

// Implementation documentation
type messageService struct {
    repo      repository.MessageRepository
    publisher redis.Publisher
}

// Constructor documentation
// NewMessageService creates a new message service with the given dependencies.
func NewMessageService(repo repository.MessageRepository, publisher redis.Publisher) MessageService {
    return &messageService{
        repo:      repo,
        publisher: publisher,
    }
}

// Method documentation
func (s *messageService) CreateMessage(ctx context.Context, input *models.CreateMessageInput) (*models.Message, error) {
    // Validate input
    if err := s.validateInput(input); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Business logic here
    message := &models.Message{
        ID:        generateID(),
        Text:      input.Text,
        CreatedAt: time.Now(),
    }
    
    // Save to repository
    savedMessage, err := s.repo.CreateMessage(ctx, message)
    if err != nil {
        return nil, fmt.Errorf("failed to save message: %w", err)
    }
    
    return savedMessage, nil
}
```

### Naming Conventions

- **Packages**: Short, lowercase, single word when possible
- **Types**: PascalCase, descriptive names
- **Functions**: PascalCase for public, camelCase for private
- **Variables**: camelCase, descriptive names
- **Constants**: UPPER_SNAKE_CASE for package-level

### Error Handling

```go
// Good: Wrap errors with context
func (s *service) DoSomething(ctx context.Context, id string) error {
    if id == "" {
        return fmt.Errorf("id cannot be empty")
    }
    
    result, err := s.repo.Get(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to get item %s: %w", id, err)
    }
    
    // ... use result
    return nil
}

// Good: Define custom error types for business logic
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}
```

## 🧪 Testing Strategy

### Test Pyramid

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows

### Unit Test Example

```go
func TestValidateCreateMessageInput(t *testing.T) {
    tests := []struct {
        name    string
        input   *models.CreateMessageInput
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid input",
            input: &models.CreateMessageInput{
                Room: "general",
                User: "john",
                Text: "Hello world",
            },
            wantErr: false,
        },
        {
            name: "empty text",
            input: &models.CreateMessageInput{
                Room: "general",
                User: "john",
                Text: "",
            },
            wantErr: true,
            errMsg:  "text cannot be empty",
        },
        {
            name:    "nil input",
            input:   nil,
            wantErr: true,
            errMsg:  "input cannot be nil",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := &messageService{}
            err := service.validateCreateMessageInput(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Integration Test Example

```go
func TestMessageRepository_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := repository.NewPostgresMessageRepository(db)
    
    // Test data
    message := &models.Message{
        ID:        "test_message_1",
        Room:      "general",
        User:      "john",
        Text:      "Test message",
        CreatedAt: time.Now(),
    }
    
    // Test create
    savedMessage, err := repo.CreateMessage(context.Background(), message)
    require.NoError(t, err)
    assert.Equal(t, message.Text, savedMessage.Text)
    
    // Test get
    retrievedMessage, err := repo.GetMessageByID(context.Background(), message.ID)
    require.NoError(t, err)
    assert.Equal(t, message.Text, retrievedMessage.Text)
}
```

### Test Coverage

```bash
# Run tests with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Target: Maintain >80% test coverage
```

## 🗄️ Database Management

### Migration Best Practices

1. **Always Create Reversible Migrations**:
```sql
-- up migration
ALTER TABLE messages ADD COLUMN is_pinned BOOLEAN DEFAULT FALSE;

-- down migration  
ALTER TABLE messages DROP COLUMN is_pinned;
```

2. **Use Descriptive Migration Names**:
```
000010_add_message_pinning.up.sql
000011_add_user_preferences.up.sql
000012_optimize_message_indexes.up.sql
```

3. **Test Migrations on Copy of Production Data**:
```bash
# Create database copy
pg_dump production_db > backup.sql
createdb test_migration_db
psql test_migration_db < backup.sql

# Test migration
migrate -path ./migrations -database postgres://user:pass@localhost/test_migration_db up
```

### Query Optimization

1. **Add Proper Indexes**:
```sql
-- For frequent queries
CREATE INDEX idx_messages_room_created ON messages(room_id, created_at DESC);
CREATE INDEX idx_messages_user_created ON messages(user_id, created_at DESC);

-- For search functionality
CREATE INDEX idx_messages_text_search ON messages USING gin(to_tsvector('english', text));
```

2. **Use Query Analysis**:
```sql
-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM messages WHERE room_id = $1 ORDER BY created_at DESC LIMIT 50;

-- Monitor slow queries
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
WHERE mean_time > 100
ORDER BY mean_time DESC;
```

## 🚀 Performance Considerations

### Database Optimization

1. **Connection Pooling**:
```go
config := &sqlx.Config{
    MaxOpenConns:    25,                // Maximum connections
    MaxIdleConns:    25,                // Keep connections alive
    ConnMaxLifetime: 5 * time.Minute,  // Recycle connections
    ConnMaxIdleTime: 30 * time.Second, // Close idle connections
}
```

2. **Query Optimization**:
```go
// Good: Use pagination
func (r *repository) GetMessages(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error) {
    query := `
        SELECT id, text, created_at 
        FROM messages 
        WHERE room_id = $1 
        ORDER BY created_at DESC 
        LIMIT $2 OFFSET $3`
    
    return r.queryMessages(ctx, query, roomID, limit, offset)
}

// Good: Use appropriate indexes
func (r *repository) SearchMessages(ctx context.Context, searchTerm string) ([]*models.Message, error) {
    query := `
        SELECT id, text, created_at 
        FROM messages 
        WHERE to_tsvector('english', text) @@ plainto_tsquery('english', $1)
        ORDER BY ts_rank(to_tsvector('english', text), plainto_tsquery('english', $1)) DESC
        LIMIT 100`
    
    return r.queryMessages(ctx, query, searchTerm)
}
```

### Caching Strategy

```go
// Redis caching for frequently accessed data
func (s *messageService) GetRoomMessages(ctx context.Context, roomID string) ([]*models.Message, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("room:messages:%s", roomID)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        var messages []*models.Message
        if json.Unmarshal(cached, &messages) == nil {
            return messages, nil
        }
    }
    
    // Fallback to database
    messages, err := s.repo.GetMessagesByRoomID(ctx, roomID, 50, 0)
    if err != nil {
        return nil, err
    }
    
    // Cache for future requests
    if data, err := json.Marshal(messages); err == nil {
        s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
    }
    
    return messages, nil
}
```

## 🔒 Security Guidelines

### Input Validation

```go
func (s *messageService) validateCreateMessageInput(input *models.CreateMessageInput) error {
    if input == nil {
        return fmt.Errorf("input cannot be nil")
    }
    
    // Validate required fields
    if strings.TrimSpace(input.Text) == "" {
        return fmt.Errorf("message text cannot be empty")
    }
    
    // Validate length limits
    if len(input.Text) > 1000 {
        return fmt.Errorf("message text cannot exceed 1000 characters")
    }
    
    // Sanitize input
    input.Text = sanitize.HTML(input.Text)
    
    // Validate format
    if !isValidRoomID(input.RoomID) {
        return fmt.Errorf("invalid room ID format")
    }
    
    return nil
}
```

### Authentication & Authorization

```go
// JWT middleware
func JWTAuthMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := extractTokenFromHeader(c.GetHeader("Authorization"))
        
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return []byte(secret), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("userID", claims["user_id"])
            c.Set("username", claims["username"])
        }
        
        c.Next()
    }
}

// Permission checking
func (s *messageService) checkRoomPermission(ctx context.Context, userID, roomID string, requiredRole models.RoomRole) error {
    member, err := s.roomRepo.GetRoomMember(ctx, roomID, userID)
    if err != nil {
        return fmt.Errorf("failed to check membership: %w", err)
    }
    
    if member == nil {
        return fmt.Errorf("user is not a member of this room")
    }
    
    if !hasRequiredRole(member.Role, requiredRole) {
        return fmt.Errorf("insufficient permissions")
    }
    
    return nil
}
```

### Rate Limiting

```go
// Rate limiting middleware
func RateLimitMiddleware(requests int, window time.Duration) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(window/time.Duration(requests)), requests)
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 🚀 Deployment Guide

### Docker Production Build

```dockerfile
# Multi-stage build for production
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

CMD ["./main"]
```

### Environment-specific Configuration

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  app:
    build: .
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - LOG_LEVEL=info
      - GRAPHQL_PLAYGROUND=false
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: chatapp
      POSTGRES_USER: chatapp
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    restart: unless-stopped

volumes:
  postgres_data:
```

### Health Checks

```go
// Health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
    // Check database connection
    if err := h.db.Ping(); err != nil {
        c.JSON(503, gin.H{
            "status": "unhealthy",
            "database": "disconnected",
        })
        return
    }
    
    // Check Redis connection
    if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
        c.JSON(503, gin.H{
            "status": "unhealthy",
            "redis": "disconnected",
        })
        return
    }
    
    c.JSON(200, gin.H{
        "status": "healthy",
        "timestamp": time.Now(),
        "version": os.Getenv("APP_VERSION"),
    })
}
```

This development guide provides a comprehensive foundation for building features following the established architecture patterns while maintaining code quality, security, and performance standards.