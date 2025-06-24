# Generic Rule Engine

A high-performance, scalable rule engine built in Go with comprehensive API support, JWT authentication, and PostgreSQL backend.

## 🚀 Features

### ✅ Implemented Features

- **🔐 JWT Authentication & Authorization**
  - Secure token-based authentication
  - Role-based access control (admin, viewer, executor)
  - Configurable JWT secrets and expiration

- **📁 Namespace Management**
  - Create, read, list, and delete namespaces
  - Hierarchical organization of rules and configurations
  - Validation and error handling

- **🏷️ Fields API**
  - Create and list fields within namespaces
  - Support for "number" and "string" field types
  - Optional descriptions for fields
  - Proper validation and error handling

- **🗄️ Database Integration**
  - PostgreSQL with sqlc for type-safe queries
  - Comprehensive database migrations
  - Connection pooling and optimization

- **🛡️ Error Handling**
  - Standardized error contract across all APIs
  - Unique error codes for each operation
  - Proper HTTP status codes and error messages

- **📊 Logging & Monitoring**
  - Structured logging with zerolog
  - Request/response logging
  - Performance metrics and monitoring

- **🧪 Testing**
  - Comprehensive unit tests for all layers
  - Mock-based testing for services and handlers
  - Integration test support

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
- **Viewer**: Read-only access to namespaces and fields
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
    type         text CHECK (type IN ('number','string')),
    description  text,
    created_by   text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (namespace, field_id)
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
```bash
# Start the server
go run ./cmd/api &

# Test with curl
curl -X POST http://localhost:8080/v1/namespaces \
  -H "Authorization: Bearer $(go run ./cmd/jwt-generator -client-id test -role admin | grep -o 'Bearer [^ ]*' | tail -1)" \
  -H "Content-Type: application/json" \
  -d '{"id": "test-ns", "description": "Test namespace"}'
```

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
│   ├── domain/             # Domain models and errors
│   ├── execution/          # Rule execution engine
│   ├── handlers/           # HTTP request handlers
│   ├── infra/              # Infrastructure (DB, logging)
│   ├── models/             # Database models (sqlc generated)
│   ├── repository/         # Data access layer
│   ├── server/             # HTTP server and middleware
│   └── service/            # Business logic layer
├── migrations/             # Database migrations
├── queries/                # SQL queries for sqlc
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

1. **Database Layer**: Add SQL queries to `queries/` directory
2. **Models**: Run `sqlc generate` to update models
3. **Repository**: Implement data access methods
4. **Service**: Add business logic
5. **Handler**: Create HTTP endpoints
6. **Tests**: Add comprehensive test coverage

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
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Code Style
- Follow Go conventions and `gofmt`
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions small and focused

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For issues and questions:
1. Check the documentation
2. Review existing issues
3. Create a new issue with detailed information

## 🔮 Roadmap

### Planned Features
- [ ] Functions API (CRUD operations)
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

---

**Generic Rule Engine** - A powerful, scalable rule engine for modern applications.
