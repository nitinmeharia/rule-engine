# Generic Rule Engine - Makefile
# Following best practices with no shortcuts

.PHONY: help build test lint clean run migrate seed docker dev-setup test-api test-api-e2e

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Variables
BINARY_NAME=rule-engine
DOCKER_IMAGE=rule-engine
GO_VERSION=1.22
DB_URL?=postgresql://localhost:5432/rule_engine?sslmode=disable
DB_TEST_URL?=postgresql://localhost:5432/rule_engine_test?sslmode=disable

# Build targets
build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "Binary built: bin/$(BINARY_NAME)"

build-linux: ## Build for Linux (for Docker)
	@echo "Building $(BINARY_NAME) for Linux..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux cmd/api/main.go

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

# Development targets
run: ## Run the application locally
	@echo "Running $(BINARY_NAME)..."
	@go run cmd/api/main.go

dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@go mod download
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "Development environment ready!"

# Testing targets
test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v -race ./test/integration/...

# API Testing targets
test-api: ## Run API tests (requires server to be running)
	@echo "Running API tests..."
	@chmod +x scripts/test-api-e2e.sh
	@./scripts/test-api-e2e.sh

test-api-e2e: ## Run full end-to-end API tests with server startup
	@echo "Running end-to-end API tests..."
	@echo "Starting server in background..."
	@make run &
	@SERVER_PID=$$!
	@sleep 5
	@echo "Server started with PID: $$SERVER_PID"
	@chmod +x scripts/test-api-e2e.sh
	@./scripts/test-api-e2e.sh
	@echo "Stopping server..."
	@kill $$SERVER_PID || true
	@echo "API tests completed"

test-api-quick: ## Run quick API tests (health and basic auth only)
	@echo "Running quick API tests..."
	@chmod +x scripts/test-api-e2e.sh
	@./scripts/test-api-e2e.sh --quick

generate-jwt: ## Generate JWT token for testing
	@echo "Generating JWT token..."
	@python3 scripts/generate-jwt.py --role admin --format curl

# Code quality targets
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Database targets
migrate: ## Run database migrations
	@echo "Running migrations..."
	@goose -dir migrations postgres "$(DB_URL)" up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@goose -dir migrations postgres "$(DB_URL)" down

migrate-status: ## Show migration status
	@echo "Migration status:"
	@goose -dir migrations postgres "$(DB_URL)" status

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=migration_name"; exit 1; fi
	@echo "Creating migration: $(NAME)"
	@goose -dir migrations create $(NAME) sql

seed: ## Seed database with sample data
	@echo "Seeding database..."
	@goose -dir migrations/seed postgres "$(DB_URL)" up

db-reset: ## Reset database (drop and recreate)
	@echo "Resetting database..."
	@goose -dir migrations postgres "$(DB_URL)" reset

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image: $(DOCKER_IMAGE)"
	@docker build -t $(DOCKER_IMAGE):latest .

docker-run: ## Run application in Docker
	@echo "Running $(DOCKER_IMAGE) in Docker..."
	@docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):latest

docker-compose-up: ## Start services with docker-compose (if available)
	@if [ -f docker-compose.yml ]; then \
		docker-compose up -d; \
	else \
		echo "docker-compose.yml not found"; \
	fi

docker-compose-down: ## Stop docker-compose services
	@if [ -f docker-compose.yml ]; then \
		docker-compose down; \
	else \
		echo "docker-compose.yml not found"; \
	fi

# Code generation targets
generate: ## Run go generate
	@echo "Running code generation..."
	@go generate ./...

sqlc-generate: ## Generate sqlc code
	@echo "Generating sqlc code..."
	@sqlc generate

# Dependency management
deps-download: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

deps-tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	@go mod tidy

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify

deps-upgrade: ## Upgrade dependencies
	@echo "Upgrading dependencies..."
	@go get -u ./...
	@go mod tidy

# Development workflow
dev: ## Run development workflow (format, lint, test)
	@echo "Running development workflow..."
	@make fmt
	@make lint
	@make test

pre-commit: ## Run pre-commit checks
	@echo "Running pre-commit checks..."
	@make fmt
	@make lint
	@make vet
	@make test

# Production targets
install-tools: ## Install required development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Security
security-scan: ## Run security scan
	@echo "Running security scan..."
	@gosec ./...

# Benchmarks
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Environment setup
setup-local-db: ## Set up local PostgreSQL database
	@echo "Setting up local database..."
	@createdb rule_engine || echo "Database might already exist"
	@createdb rule_engine_test || echo "Test database might already exist"

# All-in-one targets
all: clean deps-download build test ## Clean, download deps, build, and test

ci: deps-verify lint test ## CI pipeline (verify, lint, test)

release: clean build-linux docker-build ## Build release artifacts 