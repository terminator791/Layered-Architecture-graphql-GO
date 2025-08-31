# Real-Time Chat Application with Layered Architecture

A complete Go application implementing a classic **Layered Architecture** pattern with GraphQL API and Redis Pub/Sub for real-time messaging. This project demonstrates clean architecture principles, proper separation of concerns, and modern Go development practices.

## 🏗️ Architecture Overview

This application follows a strict layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│                  (GraphQL Resolvers)                       │
├─────────────────────────────────────────────────────────────┤
│                     Business Layer                         │
│                   (Service Logic)                          │
├─────────────────────────────────────────────────────────────┤
│                   Persistence Layer                        │
│                  (Repository Pattern)                      │
├─────────────────────────────────────────────────────────────┤
│                  Infrastructure Layer                      │
│               (Database, Redis, Config)                    │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

- **Presentation Layer**: GraphQL schema, resolvers, and API endpoints
- **Business Layer**: Core business logic, validation, and service orchestration
- **Persistence Layer**: Data access abstractions and repository implementations
- **Infrastructure Layer**: External dependencies (database, Redis, configuration)

## 🚀 Features

- **Real-time messaging** using GraphQL Subscriptions powered by Redis Pub/Sub
- **Layered Architecture** with proper dependency injection
- **GraphQL API** with Queries, Mutations, and Subscriptions
- **PostgreSQL** for persistent message storage
- **Redis Pub/Sub** for real-time message broadcasting
- **Docker support** with multi-stage builds
- **Comprehensive unit tests** with mocked dependencies
- **Clean project structure** following Go best practices

## 🛠️ Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.21+ |
| API Framework | GraphQL (gqlgen) | v0.17.78 |
| Database | PostgreSQL | 15-alpine |
| Cache/Pub-Sub | Redis | 7-alpine |
| Database Driver | sqlx + pq | Latest |
| Testing | testify | Latest |
| Containerization | Docker & Docker Compose | Latest |

## 📁 Project Structure

```
.
├── cmd/
│   └── server/                 # Application entry point
│       └── main.go
├── internal/
│   ├── business/              # Business logic layer
│   │   └── service/
│   │       ├── interfaces.go
│   │       ├── message_service.go
│   │       └── message_service_test.go
│   ├── config/                # Configuration management
│   │   └── config.go
│   ├── domain/                # Domain models
│   │   └── models/
│   │       └── message.go
│   ├── infrastructure/        # External dependencies
│   │   ├── database/
│   │   │   └── postgres.go
│   │   └── redis/
│   │       ├── interfaces.go
│   │       └── redis_pubsub.go
│   ├── mocks/                 # Test mocks
│   │   ├── redis_mock.go
│   │   └── repository_mock.go
│   ├── persistence/           # Data access layer
│   │   └── repository/
│   │       ├── interfaces.go
│   │       └── postgres_message_repository.go
│   └── presentation/          # API layer
│       └── graphql/
│           ├── generated/     # Generated GraphQL code
│           ├── resolvers/     # GraphQL resolvers
│           └── schema/        # GraphQL schema definition
├── migrations/                # Database migrations
├── docker-compose.yml         # Docker services configuration
├── Dockerfile                 # Application container
├── gqlgen.yml                # GraphQL code generation config
├── go.mod                    # Go module definition
└── README.md                 # This file
```

## 🔄 Real-Time Flow

The application implements a complete real-time messaging flow:

1. **Client sends mutation**: `sendMessage(room: "general", user: "john", text: "Hello!")`
2. **Presentation Layer**: GraphQL resolver receives the request
3. **Business Layer**: Validates input and calls repository to save message
4. **Persistence Layer**: Saves message to PostgreSQL database
5. **Business Layer**: Publishes message to Redis channel `chat_room:general`
6. **Real-time Distribution**: Redis Pub/Sub broadcasts to all subscribers
7. **GraphQL Subscription**: Active subscribers receive the new message instantly

## 🎯 GraphQL API

### Mutations

```graphql
# Send a new message to a chat room
mutation SendMessage {
  sendMessage(input: {
    room: "general"
    user: "john_doe"
    text: "Hello, everyone!"
  }) {
    id
    room
    user
    text
    createdAt
  }
}
```

### Queries

```graphql
# Get all messages for a specific room
query GetMessages {
  messages(room: "general") {
    id
    room
    user
    text
    createdAt
  }
}
```

### Subscriptions

```graphql
# Subscribe to real-time messages in a room
subscription MessageAdded {
  messageAdded(room: "general") {
    id
    room
    user
    text
    createdAt
  }
}
```

## 🏃‍♂️ Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/terminator791/Layered-Architecture-graphql-GO.git
   cd Layered-Architecture-graphql-GO
   ```

2. **Start all services**
   ```bash
   docker-compose up --build
   ```

3. **Access the application**
   - GraphQL Playground: http://localhost:8080
   - GraphQL API Endpoint: http://localhost:8080/query

### Local Development

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Start PostgreSQL and Redis**
   ```bash
   docker-compose up postgres redis -d
   ```

3. **Set environment variables**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=chatdb
   export REDIS_HOST=localhost
   export REDIS_PORT=6379
   export SERVER_PORT=8080
   ```

4. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

## 🧪 Testing

The project includes comprehensive unit tests for the business layer with mocked dependencies.

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run business layer tests only
go test ./internal/business/service/ -v

# Run tests with detailed output
go test ./... -v
```

### Test Coverage

The business layer has **100% test coverage** including:

- ✅ Successful message creation and publishing
- ✅ Repository error handling
- ✅ Publisher error handling (graceful degradation)
- ✅ Input validation (nil input, empty fields, length limits)
- ✅ Room subscription functionality
- ✅ Error propagation and logging

## 🐳 Docker Usage

### Build Application Image

```bash
docker build -t chat-app .
```

### Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up --build
```

## 📝 Database Schema

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    room VARCHAR(50) NOT NULL,
    "user" VARCHAR(50) NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for optimal query performance
CREATE INDEX idx_messages_room ON messages(room);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_room_created_at ON messages(room, created_at);
```

## 🔧 Configuration

The application uses environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL username | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_NAME` | PostgreSQL database name | `chatdb` |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | `` |
| `SERVER_PORT` | Application server port | `8080` |

## 🏗️ Architecture Decisions

### Why Layered Architecture?

1. **Separation of Concerns**: Each layer has a single, well-defined responsibility
2. **Testability**: Business logic can be tested in isolation with mocked dependencies
3. **Maintainability**: Changes in one layer don't affect others
4. **Scalability**: Layers can be scaled independently

### Why GraphQL?

1. **Type Safety**: Strong typing with schema-first development
2. **Real-time Subscriptions**: Built-in support for WebSocket subscriptions
3. **Flexible Queries**: Clients can request exactly the data they need
4. **Introspection**: Self-documenting API with playground

### Why Redis Pub/Sub?

1. **Performance**: In-memory operations for fast message delivery
2. **Scalability**: Horizontal scaling with Redis Cluster
3. **Reliability**: Persistent connections with automatic reconnection
4. **Simplicity**: Straightforward pub/sub pattern implementation

## 🔍 Testing Strategy

### Unit Tests

- **Business Layer**: Complete isolation with mocked dependencies
- **Repository Layer**: Integration tests with test database
- **Service Layer**: End-to-end functionality tests

### Test Principles

1. **Arrange-Act-Assert**: Clear test structure
2. **Mock External Dependencies**: Test business logic in isolation
3. **Test Edge Cases**: Error conditions and boundary values
4. **Descriptive Test Names**: Clear intent and expected behavior

## 🚀 Production Considerations

### Security

- [ ] Implement authentication and authorization
- [ ] Add rate limiting for API endpoints
- [ ] Use HTTPS in production
- [ ] Sanitize user inputs to prevent XSS

### Monitoring

- [ ] Add structured logging
- [ ] Implement metrics collection
- [ ] Set up health check endpoints
- [ ] Add distributed tracing

### Performance

- [ ] Database connection pooling
- [ ] Redis connection optimization
- [ ] GraphQL query complexity analysis
- [ ] Implement caching strategies

### Deployment

- [ ] Kubernetes manifests
- [ ] CI/CD pipeline
- [ ] Database migration strategy
- [ ] Environment-specific configurations

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [gqlgen](https://gqlgen.com/) for excellent GraphQL tooling
- [testify](https://github.com/stretchr/testify) for comprehensive testing utilities
- [sqlx](https://github.com/jmoiron/sqlx) for enhanced SQL operations
- [go-redis](https://github.com/redis/go-redis) for Redis client library

---

**Built with ❤️ using Go and GraphQL**