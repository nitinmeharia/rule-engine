# Generic Rule Engine

A high-performance, scalable rule engine built in Go with comprehensive API support, JWT authentication, and PostgreSQL backend. Built following Clean Architecture principles with rich domain models and comprehensive testing.

## 🚀 Features

### ✅ Implemented Features

- **🔐 JWT Authentication & Authorization**
  - Secure token-based authentication
  - Role-based access control (admin, viewer, executor)
  - Configurable JWT secrets and expiration

- **📁 Namespace Management**
  - Create, read, list, and delete namespaces
  - Hierarchical organization of rules and configurations
  - Rich domain model validation
  - Comprehensive error handling

- **🏷️ Fields API**
  - Create and list fields within namespaces
  - Support for multiple field types (number, string, boolean, date)
  - Optional descriptions for fields
  - Domain-driven validation

- **⚙️ Functions API**
  - Create, read, update, and delete functions
  - Support for multiple function types (max, sum, avg, in)
  - Draft and published versions with lifecycle management
  - Function validation and dependency checking
  - Role-based access control

- **🏗️ Clean Architecture**
  - Proper separation of concerns
  - Framework-agnostic business logic
  - Rich domain models with encapsulated validation
  - Dependency inversion with interfaces
  - Testable and maintainable codebase

- **🗄️ Database Integration**
  - PostgreSQL with sqlc for type-safe queries
  - Comprehensive database migrations
  - Connection pooling and optimization

- **🛡️ Error Handling**
  - Standardized error contract across all APIs
  - Unique error codes for each operation
  - Proper HTTP status codes and error messages
  - Centralized response handling

- **📊 Logging & Monitoring**
  - Structured logging with zerolog
  - Request/response logging
  - Performance metrics and monitoring

- **🧪 Comprehensive Testing**
  - Unit tests for all layers (handlers, services, repositories)
  - Integration tests with real database
  - End-to-end API testing
  - Mock-based testing for services
  - Consolidated test scripts with shared cleanup

## 🏗️ Architecture

The project follows Clean Architecture principles with clear separation of concerns:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Handlers      │    │    Services     │    │   Domain        │
│   (HTTP Layer)  │───▶│   (Orchestration)│───▶│   (Business     │
│                 │    │                 │    │    Logic)       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   DTOs          │    │  Repositories   │    │   Models        │
│   (Data Transfer)│    │   (Data Access) │    │   (Rich Domain) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Architectural Features

- **Framework Independence**: Business logic is completely decoupled from HTTP framework
- **Rich Domain Models**: Validation logic encapsulated within domain entities
- **Dependency Inversion**: Services depend on interfaces, not concrete implementations
- **Testability**: Each layer can be tested independently
- **Maintainability**: Clear separation of concerns and consistent patterns

## 📋 Prerequisites

- **Go 1.21+**
- **PostgreSQL 12+**
- **Make** (for build automation)

## 🛠️ Installation & Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd rule-engine
```

### 2. Database Setup

#### Install PostgreSQL
```bash
# macOS (using Homebrew)
brew install postgresql
brew services start postgresql

# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib
sudo systemctl start postgresql
```

#### Create Database
```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE rule_engine_dev;
\q
```

#### Run Migrations
```bash
# Install goose (if not already installed)
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
goose -dir migrations postgres "postgresql://postgres:postgres@localhost:5432/rule_engine_dev?sslmode=disable" up
```

### 3. Configuration

The application uses environment-specific configuration files:

- `configs/config.yaml` - Default configuration
- `configs/config.development.yaml` - Development overrides
- `configs/config.production.yaml` - Production overrides

Key configuration options:
```yaml
database:
  host: "localhost"
  port: 5432
  name: "rule_engine_dev"
  user: "postgres"
  password: "postgres"
  sslMode: "disable"

jwt:
  secret: "dev-secret-key-change-in-production"
  tokenExpiration: 24h
  requiredClaims: ["clientId", "role"]

server:
  host: "0.0.0.0"
  port: 8080
```

### 4. Build & Run

#### Development Mode
```bash
# Run the API server
go run ./cmd/api

# Generate JWT tokens for testing
go run ./cmd/jwt-generator -client-id test-user -role admin
```

#### Production Build
```bash
# Build the application
make build

# Run the binary
./bin/rule-engine
```

## 🔌 API Documentation

### Authentication

All API endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

#### Generate JWT Token
```bash
go run ./cmd/jwt-generator -client-id your-client -role admin -secret "your-secret"
```

### Namespaces API

#### List Namespaces
```http
GET /v1/namespaces
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "my-namespace",
    "description": "My test namespace",
    "createdAt": "2025-06-24T09:58:11.634627+05:30",
    "createdBy": "test-client"
  }
]
```

#### Create Namespace
```http
POST /v1/namespaces
Authorization: Bearer <token>
Content-Type: application/json

{
  "id": "new-namespace",
  "description": "A new namespace"
}
```

**Response:**
```json
{
  "success": true,
  "namespace": {
    "id": "new-namespace",
    "description": "A new namespace",
    "createdAt": "2025-06-24T09:58:11.634627+05:30",
    "createdBy": "test-client"
  }
}
```

#### Get Namespace
```http
GET /v1/namespaces/{id}
Authorization: Bearer <token>
```

#### Delete Namespace
```http
DELETE /v1/namespaces/{id}
Authorization: Bearer <token>
```

### Fields API

#### List Fields
```http
GET /v1/namespaces/{namespace-id}/fields
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "fieldId": "user_age",
    "type": "number",
    "description": "Age of the user",
    "createdAt": "2025-06-24T09:58:00.903324+05:30",
    "createdBy": "test-client"
  }
]
```

#### Create Field
```http
POST /v1/namespaces/{namespace-id}/fields
Authorization: Bearer <token>
Content-Type: application/json

{
  "fieldId": "user_name",
  "type": "string",
  "description": "User's full name"
}
```

**Response:**
```json
{
  "success": true,
  "field": {
    "fieldId": "user_name",
    "type": "string",
    "description": "User's full name",
    "createdAt": "2025-06-24T09:58:29.146389+05:30",
    "createdBy": "test-client"
  }
}
```

### Functions API

#### List Functions
```http
GET /v1/namespaces/{namespace-id}/functions
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "max_income",
    "version": 1,
    "status": "active",
    "type": "max",
    "args": ["salary", "bonus"],
    "values": null,
    "returnType": "number",
    "createdAt": "2025-06-24T09:58:29.146389+05:30",
    "createdBy": "test-client",
    "publishedAt": "2025-06-24T10:00:00.000000+05:30",
    "publishedBy": "test-client"
  }
]
```

#### Create Function
```http
POST /v1/namespaces/{namespace-id}/functions
Authorization: Bearer <token>
Content-Type: application/json

{
  "id": "max_income",
  "type": "max",
  "args": ["salary", "bonus"]
}
```

**Response:**
```json
{
  "function": {
    "id": "max_income",
    "version": 1,
    "status": "draft",
    "type": "max",
    "args": ["salary", "bonus"],
    "values": null,
    "returnType": "number",
    "createdAt": "2025-06-24T09:58:29.146389+05:30",
    "createdBy": "test-client",
    "publishedAt": null,
    "publishedBy": null
  }
}
```

#### Get Function
```http
GET /v1/namespaces/{namespace-id}/functions/{function-id}
Authorization: Bearer <token>
```

#### Update Function Draft
```http
PUT /v1/namespaces/{namespace-id}/functions/{function-id}/versions/draft
Authorization: Bearer <token>
Content-Type: application/json

{
  "type": "max",
  "args": ["salary", "bonus", "commission"]
}
```

#### Publish Function
```http
POST /v1/namespaces/{namespace-id}/functions/{function-id}/publish
Authorization: Bearer <token>
```

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-06-24T04:06:20.058354Z",
  "version": "1.0.0"
}
```

## 🔐 Authorization

The API supports role-based access control:

- **Admin**: Full access to all operations
- **Viewer**: Read-only access to namespaces, fields, and functions
- **Executor**: Read access + execution capabilities (future)

### Required Permissions

| Operation | Admin | Viewer | Executor |
|-----------|-------|--------|----------|
| List Namespaces | ✅ | ✅ | ✅ |
| Create Namespace | ✅ | ❌ | ❌ |
| Get Namespace | ✅ | ✅ | ✅ |
| Delete Namespace | ✅ | ❌ | ❌ |
| List Fields | ✅ | ✅ | ✅ |
| Create Field | ✅ | ❌ | ❌ |
| List Functions | ✅ | ✅ | ✅ |
| Create Function | ✅ | ❌ | ❌ |
| Update Function | ✅ | ❌ | ❌ |
| Publish Function | ✅ | ❌ | ❌ |

## 🗄️ Database Schema

### Namespaces Table
```sql
CREATE TABLE namespaces (
    id          text PRIMARY KEY,
    description text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    created_by  text NOT NULL
);
```

### Fields Table
```sql
CREATE TABLE fields (
    namespace    text REFERENCES namespaces(id) ON DELETE CASCADE,
    field_id     text,
    type         text CHECK (type IN ('number','string','boolean','date')),
    description  text,
    created_by   text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (namespace, field_id)
);
```

### Functions Table
```sql
CREATE TABLE functions (
    namespace     text REFERENCES namespaces(id) ON DELETE CASCADE,
    function_id   text,
    version       integer NOT NULL,
    status        text CHECK (status IN ('draft','active','inactive')),
    type          text NOT NULL,
    args          text[],
    values        text[],
    return_type   text NOT NULL,
    created_by    text NOT NULL,
    published_by  text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    published_at  timestamptz,
    PRIMARY KEY (namespace, function_id, version)
);
```

## 🧪 Testing

### Run All Tests
```bash
go test ./...
```

### Run Specific Test Suites
```bash
# Handler tests
go test ./internal/handlers -v

# Service tests
go test ./internal/service -v

# Repository tests
go test ./internal/repository -v
```

### Integration Testing

#### Quick Tests
```bash
# Run quick API tests
./scripts/test-api-quick.sh
```

#### End-to-End Tests
```bash
# Run comprehensive E2E tests
./scripts/test-api-e2e.sh
```

#### Functions API Tests
```bash
# Run functions-specific tests
./scripts/test-functions-api.sh
```

### Test Architecture

The testing strategy follows the Clean Architecture principles:

- **Unit Tests**: Test each layer in isolation with mocks
- **Integration Tests**: Test service layer with real repositories
- **End-to-End Tests**: Test complete API workflows
- **Shared Test Infrastructure**: Consolidated cleanup scripts and utilities

## 🏗️ Project Structure

```
rule-engine/
├── cmd/
│   ├── api/                 # Main API server
│   └── jwt-generator/       # JWT token generator
├── configs/                 # Configuration files
├── docs/                    # Documentation
├── internal/
│   ├── auth/               # Authentication utilities
│   ├── bootstrap/          # Application initialization
│   ├── config/             # Configuration management
│   ├── domain/             # Domain models, DTOs, and errors
│   ├── execution/          # Rule execution engine
│   ├── handlers/           # HTTP request handlers
│   ├── infra/              # Infrastructure (DB, logging)
│   ├── models/             # Database models (sqlc generated)
│   ├── repository/         # Data access layer
│   ├── server/             # HTTP server and middleware
│   └── service/            # Business logic layer
├── migrations/             # Database migrations
├── queries/                # SQL queries for sqlc
├── scripts/                # Test scripts and utilities
│   ├── cleanup-test-data.sh # Shared test cleanup
│   ├── test-api-e2e.sh     # End-to-end tests
│   ├── test-api-quick.sh   # Quick tests
│   └── test-functions-api.sh # Functions API tests
└── Makefile               # Build automation
```

## 🔧 Development

### Code Generation

#### Generate SQL Models
```bash
sqlc generate
```

#### Generate Migrations
```bash
# Create new migration
goose -dir migrations postgres "postgresql://postgres:postgres@localhost:5432/rule_engine_dev?sslmode=disable" create migration_name sql
```

### Adding New Features

1. **Domain Layer**: Define rich domain models with validation
2. **Database Layer**: Add SQL queries to `queries/` directory
3. **Models**: Run `sqlc generate` to update models
4. **Repository**: Implement data access methods
5. **Service**: Add business logic (orchestration only)
6. **Handler**: Create HTTP endpoints with DTOs
7. **Tests**: Add comprehensive test coverage

### Architectural Guidelines

#### Domain Models
```go
// Rich domain model with encapsulated validation
type Namespace struct {
    ID          string    `json:"id"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"createdAt"`
    CreatedBy   string    `json:"createdBy"`
}

func (n *Namespace) Validate() error {
    // Validation logic encapsulated within the domain model
    if n == nil {
        return ErrValidationError
    }
    // ... validation rules
    return nil
}
```

#### Service Layer
```go
// Service depends on interfaces, not concrete types
type NamespaceService struct {
    namespaceRepo domain.NamespaceRepository
}

func (s *NamespaceService) CreateNamespace(ctx context.Context, namespace *domain.Namespace) error {
    // Use domain validation
    if err := namespace.Validate(); err != nil {
        return err
    }
    // Business logic orchestration
    return s.namespaceRepo.Create(ctx, namespace)
}
```

#### Handler Layer
```go
// Handler uses DTOs and response handler for consistency
type NamespaceHandler struct {
    namespaceService service.NamespaceServiceInterface
    responseHandler  *ResponseHandler
}

func (h *NamespaceHandler) CreateNamespace(c *gin.Context) {
    var req domain.CreateNamespaceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.responseHandler.BadRequest(c, "Invalid request body")
        return
    }
    // ... handler logic
}
```

### Error Handling

The application uses a standardized error contract:

```go
type APIError struct {
    Code      string `json:"code"`
    ErrorType string `json:"error"`
    Message   string `json:"message"`
}
```

All errors are mapped to appropriate HTTP status codes and include unique error codes for client handling.

## 🚀 Deployment

### Docker (Recommended)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o rule-engine ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/rule-engine .
EXPOSE 8080
CMD ["./rule-engine"]
```

### Environment Variables

```bash
# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=rule_engine_prod
DATABASE_USER=postgres
DATABASE_PASSWORD=secure_password
DATABASE_SSL_MODE=require

# JWT
JWT_SECRET=your-production-secret
JWT_TOKEN_EXPIRATION=24h

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
```

## 📊 Monitoring & Logging

### Log Levels
- `debug`: Detailed debugging information
- `info`: General application information
- `warn`: Warning messages
- `error`: Error conditions

### Metrics
The application exposes Prometheus metrics at `/metrics` (when enabled).

### Health Checks
Use the `/health` endpoint for load balancer health checks.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes following Clean Architecture principles
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Code Style
- Follow Go conventions and `gofmt`
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions small and focused
- Follow Clean Architecture principles
- Use rich domain models with encapsulated validation

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For issues and questions:
1. Check the documentation
2. Review existing issues
3. Create a new issue with detailed information

## 🔮 Roadmap

### Planned Features
- [x] Functions API (CRUD operations) ✅
- [ ] Rules API (rule definition and management)
- [ ] Workflows API (workflow orchestration)
- [ ] Terminals API (execution endpoints)
- [ ] Rule execution engine
- [ ] Caching layer
- [ ] Rate limiting
- [ ] API versioning
- [ ] OpenAPI/Swagger documentation
- [ ] GraphQL support
- [ ] WebSocket support for real-time updates

### Performance Improvements
- [ ] Database query optimization
- [ ] Connection pooling improvements
- [ ] Caching strategies
- [ ] Load balancing support
- [ ] Horizontal scaling

### Architectural Enhancements
- [x] Clean Architecture implementation ✅
- [x] Rich domain models ✅
- [x] Framework-agnostic business logic ✅
- [x] Comprehensive testing strategy ✅
- [ ] Event sourcing
- [ ] CQRS pattern
- [ ] Microservices architecture

---

**Generic Rule Engine** - A powerful, scalable rule engine built with Clean Architecture principles for modern applications.
