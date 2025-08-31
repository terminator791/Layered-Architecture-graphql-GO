# gRPC Microservices Architecture - User & Order Services

A complete multi-service Go project demonstrating best-practice microservices architecture with gRPC inter-service communication.

## 🏗️ Architecture Overview

This project implements two microservices:

- **`user-service`**: Manages user data with its own PostgreSQL database and exposes gRPC API
- **`order-service`**: Manages orders with its own PostgreSQL database. Validates users via gRPC calls to user-service

### Core gRPC Communication Flow

1. **User Service** exposes gRPC methods: `GetUser` and `CreateUser`
2. **Order Service** acts as gRPC client to validate users during order creation
3. When creating an order, order-service calls user-service via gRPC to verify the user exists
4. If user validation fails (NOT_FOUND), order creation is rejected
5. If user exists, order is created in order-service's database

## 🛠️ Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.24+ |
| RPC Framework | gRPC | v1.75.0 |
| Interface Definition | Protocol Buffers v3 | Latest |
| Database | PostgreSQL | 15-alpine |
| Database Driver | sqlx + pq | Latest |
| Testing | testify | Latest |
| Containerization | Docker & Docker Compose | Latest |

## 📁 Project Structure

```
.
├── proto/                          # Protocol Buffer definitions
│   ├── user.proto                 # User service gRPC definition
│   ├── user.pb.go                 # Generated protobuf code
│   └── user_grpc.pb.go            # Generated gRPC code
├── services/
│   ├── user-service/              # User microservice
│   │   ├── cmd/server/            # Service entry point
│   │   ├── internal/
│   │   │   ├── business/service/  # Business logic
│   │   │   ├── domain/models/     # Domain models
│   │   │   ├── grpc/              # gRPC server implementation
│   │   │   ├── persistence/repository/ # Data access layer
│   │   │   ├── config/            # Configuration
│   │   │   └── mocks/             # Test mocks
│   │   └── Dockerfile
│   └── order-service/             # Order microservice
│       ├── cmd/server/            # Service entry point
│       ├── internal/
│       │   ├── business/service/  # Business logic with gRPC client
│       │   ├── domain/models/     # Domain models
│       │   ├── grpc/              # gRPC client for user service
│       │   ├── http/              # HTTP REST API handlers
│       │   ├── persistence/repository/ # Data access layer
│       │   ├── config/            # Configuration
│       │   └── mocks/             # Test mocks including gRPC client
│       └── Dockerfile
├── cmd/cli/                       # CLI client for testing
└── docker-compose.microservices.yml # Multi-service Docker setup
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Protocol Buffers compiler (`protoc`)

### Running with Docker (Recommended)

```bash
# Start all services (user-service, order-service, and their databases)
docker-compose -f docker-compose.microservices.yml up --build

# Services will be available at:
# - User Service (gRPC): localhost:9001
# - Order Service (HTTP): localhost:8080
# - User DB: localhost:5433
# - Order DB: localhost:5434
```

### Local Development

1. **Start databases:**
```bash
# Start only the databases
docker-compose -f docker-compose.microservices.yml up user-db order-db
```

2. **Run user service:**
```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5433
export DB_NAME=userdb
export SERVER_PORT=9001

# Run user service
go run ./services/user-service/cmd/server
```

3. **Run order service:**
```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5434
export DB_NAME=orderdb
export SERVER_PORT=8080
export USER_SERVICE_ADDRESS=localhost:9001

# Run order service
go run ./services/order-service/cmd/server
```

## 🧪 Testing

### Unit Tests with Mocked gRPC Client

The project includes comprehensive unit tests that mock the gRPC client for complete isolation:

```bash
# Run all tests
go test ./services/... -v

# Run order service tests (includes mocked gRPC client tests)
go test ./services/order-service/internal/business/service -v

# Run user service tests
go test ./services/user-service/internal/business/service -v
```

### Key Test Coverage

- **Order Service Tests**: Mock gRPC client to test order creation logic in isolation
- **User validation scenarios**: NOT_FOUND, internal errors, successful validation
- **Input validation**: All edge cases and error conditions
- **Repository error handling**: Database failure scenarios

## 🎯 API Usage

### Using the CLI Client

```bash
# Build CLI client
go build -o cli ./cmd/cli

# Create a user
./cli create-user "john@example.com" "John Doe"
# Output: User created with ID: abc-123

# Get user by ID
./cli get-user "abc-123"

# Create an order (will validate user via gRPC)
./cli create-order "abc-123" "product-456" 2 99.99

# Try to create order with non-existent user (will fail)
./cli create-order "non-existent" "product-456" 1 50.00
```

### Direct API Calls

**User Service (gRPC):**
```bash
# Using grpcurl (install separately)
grpcurl -plaintext -d '{"email":"test@example.com","name":"Test User"}' \
  localhost:9001 user.UserService/CreateUser

grpcurl -plaintext -d '{"id":"user-id-here"}' \
  localhost:9001 user.UserService/GetUser
```

**Order Service (HTTP REST):**
```bash
# Create order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":"abc-123","product_id":"product-456","quantity":2,"amount":99.99}'

# Get order
curl "http://localhost:8080/orders/get?id=order-id-here"
```

## 🔄 gRPC Communication Flow Example

1. **Client requests order creation:**
   ```json
   POST /orders
   {
     "user_id": "abc-123",
     "product_id": "product-456", 
     "quantity": 2,
     "amount": 99.99
   }
   ```

2. **Order service validates user via gRPC:**
   ```protobuf
   GetUserRequest { id: "abc-123" }
   → user-service:9001
   ```

3. **User service responds:**
   ```protobuf
   GetUserResponse { 
     user: { id: "abc-123", email: "john@example.com", name: "John Doe" }
   }
   ```

4. **Order service creates order:**
   ```json
   {
     "id": "order-789",
     "user_id": "abc-123",
     "product_id": "product-456",
     "quantity": 2,
     "amount": 99.99,
     "status": "pending"
   }
   ```

## 🏗️ Architecture Decisions

### gRPC for Inter-Service Communication
- **Performance**: Binary protocol, HTTP/2, efficient serialization
- **Type Safety**: Strong typing with Protocol Buffers
- **Code Generation**: Automatic client/server code generation
- **Streaming**: Built-in support for streaming RPCs

### Separate Databases per Service
- **Data Isolation**: Each service owns its data
- **Independent Scaling**: Scale databases independently
- **Technology Flexibility**: Use different DB technologies per service

### Layered Architecture within Services
- **Separation of Concerns**: Business logic separated from data access
- **Testability**: Easy to mock dependencies
- **Maintainability**: Clear boundaries between layers

## 🔍 Testing Strategy

### Unit Tests with Mocks
- **gRPC Client Mocking**: Test order service without running user service
- **Repository Mocking**: Test business logic without database
- **Comprehensive Coverage**: All error scenarios and edge cases

### Integration Testing
- **End-to-End Flow**: CLI client tests the complete user creation → order creation flow
- **Service Interaction**: Verify gRPC communication works correctly
- **Database Integration**: Test with real PostgreSQL instances

## 🚀 Production Considerations

### Service Discovery
- Consider using Consul, etcd, or Kubernetes service discovery
- Implement health checks and service registration

### Load Balancing
- Use gRPC load balancing for user service calls
- Consider server-side or client-side load balancing

### Monitoring & Observability
- Add gRPC interceptors for logging and metrics
- Implement distributed tracing (OpenTelemetry)
- Health check endpoints

### Security
- Implement TLS for gRPC communication
- Add authentication/authorization middleware
- Validate and sanitize all inputs

### Resilience
- Circuit breaker pattern for gRPC calls
- Retry logic with exponential backoff
- Timeout configuration

## 📝 Environment Variables

### User Service
```bash
DB_HOST=localhost           # Database host
DB_PORT=5432               # Database port  
DB_USER=postgres           # Database user
DB_PASSWORD=postgres       # Database password
DB_NAME=userdb            # Database name
DB_SSLMODE=disable        # SSL mode
SERVER_PORT=9001          # gRPC server port
```

### Order Service
```bash
DB_HOST=localhost           # Database host
DB_PORT=5432               # Database port
DB_USER=postgres           # Database user  
DB_PASSWORD=postgres       # Database password
DB_NAME=orderdb           # Database name
DB_SSLMODE=disable        # SSL mode
SERVER_PORT=8080          # HTTP server port
USER_SERVICE_ADDRESS=localhost:9001  # User service gRPC address
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.