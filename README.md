# Advanced Real-Time Chat Application with Layered Architecture

A comprehensive, production-ready real-time chat application built with Go, GraphQL, and Redis. This application showcases modern software architecture patterns with extensive features for user management, room administration, advanced messaging, and real-time communication.

## 🚀 Features

### 🔐 User Management & Authentication
- **User Registration & Login** with JWT authentication
- **User Profiles** with avatars, display names, and bio
- **User Status Management** (online, offline, away, busy)
- **Secure Password Hashing** using bcrypt
- **User Presence Tracking** with last seen timestamps

### 🏠 Advanced Room Management
- **Multiple Room Types**: Public, Private, and Direct message rooms
- **Room Metadata**: Names, descriptions, avatars, and member limits
- **Member Role System**: Admins, Moderators, and Members
- **Password-Protected Rooms** for enhanced privacy
- **Room Permissions**: Join/leave, kick members, update settings
- **Real-time Member Counts** and online status tracking

### 💬 Rich Messaging System
- **Multiple Message Types**: Text, images, files, and system notifications
- **Message Threading**: Reply to specific messages
- **Message Editing & Deletion** with history tracking
- **Message Reactions**: Emoji reactions with real-time counts
- **Message Search**: Full-text search across rooms
- **Rich Metadata Support**: Image dimensions, file information
- **Soft Deletion**: Messages are archived, not permanently deleted

### ⚡ Real-Time Features
- **WebSocket Subscriptions** for instant message delivery
- **Typing Indicators** showing who's currently typing
- **User Presence Broadcasting** for online/offline status
- **Live Room Statistics** (member counts, online users)
- **Real-time Reactions** and message updates

### 🏗️ Architecture & Technical Excellence
- **Layered Architecture** with clear separation of concerns
- **Domain-Driven Design** with rich domain models
- **Repository Pattern** for data access abstraction
- **Dependency Injection** for testable, maintainable code
- **GraphQL API** with type-safe operations
- **Redis Pub/Sub** for scalable real-time messaging
- **PostgreSQL** with optimized indexes and relationships

## 🛠️ Technology Stack

| Component | Technology | Version | Purpose |
|-----------|------------|---------|---------|
| **Language** | Go | 1.21+ | Backend development |
| **API Framework** | GraphQL (gqlgen) | v0.17.78 | Type-safe API |
| **Database** | PostgreSQL | 15-alpine | Data persistence |
| **Cache/Pub-Sub** | Redis | 7-alpine | Real-time messaging |
| **Authentication** | JWT | v5.3.0 | Secure user auth |
| **Password Hashing** | bcrypt | Latest | Secure passwords |
| **Database Driver** | sqlx + pq | Latest | Database operations |
| **Testing** | testify | Latest | Unit testing |
| **Containerization** | Docker & Docker Compose | Latest | Deployment |

## 🏗️ Architecture Overview

This application follows a strict layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│              (GraphQL Resolvers & Schema)                  │
├─────────────────────────────────────────────────────────────┤
│                     Business Layer                         │
│           (Services, Domain Logic, Validation)             │
├─────────────────────────────────────────────────────────────┤
│                   Persistence Layer                        │
│              (Repository Pattern & Interfaces)             │
├─────────────────────────────────────────────────────────────┤
│                  Infrastructure Layer                      │
│         (Database, Redis, JWT, External Services)          │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

- **Presentation Layer**: GraphQL schema, resolvers, input validation, and API endpoints
- **Business Layer**: Core business logic, domain rules, service orchestration, and validation
- **Persistence Layer**: Data access abstractions, repository implementations, and data mapping
- **Infrastructure Layer**: External dependencies (database, Redis, JWT, configuration)

## 🏃‍♂️ Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (if running locally)
- Redis 7+ (if running locally)

### Using Docker (Recommended)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/terminator791/Layered-Architecture-graphql-GO.git
   cd Layered-Architecture-graphql-GO
   ```

2. **Start the application**:
   ```bash
   docker-compose up -d
   ```

3. **Access the application**:
   - GraphQL Playground: http://localhost:8080/
   - GraphQL API: http://localhost:8080/query

### Local Development

1. **Setup environment variables**:
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=chatuser
   export DB_PASSWORD=chatpass
   export DB_NAME=chatdb
   export REDIS_HOST=localhost
   export REDIS_PORT=6379
   export JWT_SECRET=your-super-secret-jwt-key
   ```

2. **Run database migrations**:
   ```bash
   # Database will be created automatically on first run
   go run cmd/server/main.go
   ```

3. **Start the server**:
   ```bash
   go run cmd/server/main.go
   ```

## 🎯 GraphQL API

### 🔐 Authentication Operations

```graphql
# Register a new user
mutation Register {
  register(input: {
    username: "john_doe"
    email: "john@example.com"
    password: "securepassword"
    displayName: "John Doe"
    bio: "Software developer"
  }) {
    token
    user {
      id
      username
      email
      displayName
      status
      createdAt
    }
  }
}

# Login existing user
mutation Login {
  login(input: {
    username: "john_doe"
    password: "securepassword"
  }) {
    token
    user {
      id
      username
      status
      lastSeenAt
    }
  }
}
```

### 🏠 Room Management

```graphql
# Create a new room
mutation CreateRoom {
  createRoom(input: {
    name: "Tech Discussions"
    description: "General technology discussions"
    roomType: PUBLIC
    maxMembers: 100
  }) {
    id
    name
    description
    creator {
      username
    }
    memberCount
    onlineCount
  }
}

# Join a room
mutation JoinRoom {
  joinRoom(input: {
    roomId: "room-uuid"
    password: "optional-password"
  }) {
    id
    role
    joinedAt
    user {
      username
      status
    }
  }
}
```

### 💬 Advanced Messaging

```graphql
# Send a message with rich metadata
mutation SendRichMessage {
  sendMessageToRoom(
    roomId: "room-uuid"
    text: "Check out this image!"
    messageType: IMAGE
    metadata: {
      imageUrl: "https://example.com/image.jpg"
      imageWidth: 800
      imageHeight: 600
    }
  ) {
    id
    text
    messageType
    metadata {
      imageUrl
      imageWidth
      imageHeight
    }
    userInfo {
      username
      avatarUrl
    }
    reactions {
      emoji
      user {
        username
      }
    }
    reactionCount {
      emoji
      count
    }
  }
}

# Add reaction to message
mutation AddReaction {
  addReaction(input: {
    messageId: "message-uuid"
    emoji: "👍"
  }) {
    id
    emoji
    user {
      username
    }
  }
}

# Search messages
query SearchMessages {
  searchMessages(
    query: "important announcement"
    roomId: "room-uuid"
    limit: 20
  ) {
    id
    text
    userInfo {
      username
    }
    createdAt
  }
}
```

### ⚡ Real-Time Subscriptions

```graphql
# Subscribe to room messages
subscription RoomMessages {
  messageAddedToRoom(roomId: "room-uuid") {
    id
    text
    messageType
    userInfo {
      username
      avatarUrl
      status
    }
    reactions {
      emoji
      user {
        username
      }
    }
    createdAt
  }
}

# Subscribe to typing indicators
subscription TypingIndicators {
  typingStarted(roomId: "room-uuid") {
    user {
      username
      avatarUrl
    }
    startedAt
  }
}
```

### 📊 Advanced Queries

```graphql
# Get user's rooms with statistics
query MyRooms {
  myRooms {
    id
    name
    description
    roomType
    memberCount
    onlineCount
    creator {
      username
    }
    members {
      role
      user {
        username
        status
        lastSeenAt
      }
    }
  }
}

# Get room messages with pagination
query RoomMessages {
  messagesByRoom(
    roomId: "room-uuid"
    limit: 50
    offset: 0
  ) {
    id
    text
    messageType
    userInfo {
      username
      avatarUrl
    }
    replyTo {
      id
      text
      userInfo {
        username
      }
    }
    reactions {
      emoji
      user {
        username
      }
    }
    editedAt
    createdAt
  }
}
```

## 🔄 Real-Time Flow

The application implements a comprehensive real-time messaging flow:

1. **Client Authentication**: User logs in and receives JWT token
2. **Room Access**: User joins rooms based on permissions and room type
3. **Message Creation**: User sends message with optional metadata and threading
4. **Business Validation**: Service layer validates message content and user permissions
5. **Database Persistence**: Message saved to PostgreSQL with relations
6. **Real-time Broadcasting**: Redis Pub/Sub distributes to all room subscribers
7. **Live Updates**: WebSocket subscriptions deliver instant updates
8. **Reaction System**: Users can react with emojis in real-time
9. **Typing Indicators**: Show live typing status with automatic cleanup
10. **Presence Management**: Track and broadcast user online/offline status

## 🗄️ Database Schema

### Core Tables

- **users**: User accounts with profiles and authentication
- **rooms**: Chat rooms with metadata and settings  
- **room_members**: User-room relationships with roles
- **messages**: Chat messages with threading and metadata
- **message_reactions**: Emoji reactions to messages
- **typing_indicators**: Real-time typing status (auto-expiring)

### Key Relationships

- Users can be members of multiple rooms
- Rooms have creators and members with different roles
- Messages belong to rooms and users, can reply to other messages
- Reactions link users to messages with emoji data
- Typing indicators track temporary user activity

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests in verbose mode
go test -v ./...

# Run specific service tests
go test ./internal/business/service/...
```

### Test Coverage

- **Unit Tests**: Service layer business logic
- **Repository Tests**: Data access layer validation
- **Integration Tests**: End-to-end workflow testing
- **Mock Testing**: Isolated component testing

## 🚀 Production Considerations

### Security

- [x] **JWT Authentication** with secure token validation
- [x] **Password Hashing** using bcrypt with proper cost
- [x] **Input Validation** at service and GraphQL layers
- [x] **SQL Injection Prevention** using parameterized queries
- [ ] Rate limiting for API endpoints
- [ ] HTTPS enforcement in production
- [ ] Content filtering and moderation tools
- [ ] User blocking and reporting system

### Performance

- [x] **Database Indexing** on frequently queried columns
- [x] **Connection Pooling** for database connections
- [x] **Efficient Queries** with proper JOIN strategies
- [x] **Redis Caching** for real-time message distribution
- [ ] GraphQL query complexity analysis
- [ ] Message pagination with cursor-based navigation
- [ ] Database query optimization and monitoring
- [ ] CDN integration for file uploads

### Scalability

- [x] **Microservice-Ready Architecture** with clear boundaries
- [x] **Horizontal Scaling** support through stateless design
- [x] **Redis Pub/Sub** for distributed real-time messaging
- [ ] Database read replicas for query scaling
- [ ] Redis clustering for high availability
- [ ] Load balancing strategies
- [ ] Auto-scaling configuration

### Monitoring & Observability

- [ ] Structured logging with correlation IDs
- [ ] Metrics collection (Prometheus/Grafana)
- [ ] Health check endpoints
- [ ] Distributed tracing (Jaeger/Zipkin)
- [ ] Error tracking and alerting
- [ ] Performance monitoring and profiling

### Deployment

- [x] **Docker containerization** with multi-stage builds
- [x] **Docker Compose** for local development
- [ ] Kubernetes manifests and Helm charts
- [ ] CI/CD pipeline with automated testing
- [ ] Database migration automation
- [ ] Environment-specific configurations
- [ ] Blue-green deployment strategy

## 🏗️ Architecture Decisions

### Why Layered Architecture?

1. **Separation of Concerns**: Each layer has a single, well-defined responsibility
2. **Testability**: Business logic can be tested in isolation with mocked dependencies
3. **Maintainability**: Changes in one layer don't affect others
4. **Scalability**: Layers can be scaled independently as microservices

### Why GraphQL?

1. **Type Safety**: Strong typing with schema-first development approach
2. **Real-time Subscriptions**: Built-in support for WebSocket subscriptions
3. **Flexible Queries**: Clients can request exactly the data they need
4. **Introspection**: Self-documenting API with interactive playground
5. **Single Endpoint**: Reduces complexity compared to multiple REST endpoints

### Why Redis Pub/Sub?

1. **Performance**: In-memory operations for fast message delivery
2. **Scalability**: Horizontal scaling with Redis Cluster support
3. **Reliability**: Persistent connections with automatic reconnection
4. **Simplicity**: Straightforward pub/sub pattern implementation
5. **Real-time**: Sub-millisecond message delivery for live features

### Why Repository Pattern?

1. **Data Access Abstraction**: Hide database implementation details
2. **Testability**: Easy to mock for unit testing
3. **Flexibility**: Can switch between different data sources
4. **Consistency**: Standardized data access patterns

## 🔧 Development

### Project Structure

```
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── business/        # Business logic layer
│   │   └── service/     # Service implementations
│   ├── domain/          # Domain models and interfaces
│   │   └── models/      # Rich domain models
│   ├── infrastructure/  # Infrastructure layer
│   │   ├── auth/        # JWT authentication
│   │   ├── database/    # Database configuration
│   │   └── redis/       # Redis pub/sub implementation
│   ├── persistence/     # Persistence layer
│   │   └── repository/  # Repository interfaces and implementations
│   └── presentation/    # Presentation layer
│       └── graphql/     # GraphQL schema and resolvers
├── migrations/          # Database migration files
├── docker-compose.yml   # Docker services configuration
└── Dockerfile          # Application container definition
```

### Adding New Features

1. **Define Domain Models**: Add new rich models in `internal/domain/models/`
2. **Create Repository Interface**: Define data access interface in `internal/persistence/repository/`
3. **Implement Repository**: Create PostgreSQL implementation with proper indexing
4. **Create Service**: Implement business logic with validation in `internal/business/service/`
5. **Update GraphQL Schema**: Add new types, inputs, and operations
6. **Implement Resolvers**: Connect GraphQL operations to services
7. **Write Tests**: Add comprehensive unit and integration tests
8. **Update Documentation**: Document new API operations and features

## 🐳 Docker Support

### Development Environment

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

### Production Build

```bash
# Build production image
docker build -t chat-app:latest .

# Run production container
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e REDIS_HOST=your-redis-host \
  -e JWT_SECRET=your-jwt-secret \
  chat-app:latest
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write comprehensive unit tests for new features
- Update documentation for API changes
- Use conventional commits for clear history
- Ensure all tests pass before submitting PR

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [gqlgen](https://gqlgen.com/) for excellent GraphQL tooling and code generation
- [testify](https://github.com/stretchr/testify) for comprehensive testing utilities
- [sqlx](https://github.com/jmoiron/sqlx) for enhanced SQL operations and scanning
- [go-redis](https://github.com/redis/go-redis) for robust Redis client library
- [jwt-go](https://github.com/golang-jwt/jwt) for JWT authentication implementation
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) for secure password hashing

---

**Built with ❤️ using Go, GraphQL, PostgreSQL, and Redis**

*This project demonstrates production-ready Go development with modern architecture patterns, comprehensive testing, and scalable real-time features.*