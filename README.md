# Generic Rule Engine

[![CI](https://img.shields.io/github/actions/workflow/status/your-org/rule-engine/ci.yml?branch=main)](https://github.com/your-org/rule-engine/actions)
[![Coverage](https://img.shields.io/codecov/c/github/your-org/rule-engine)](https://codecov.io/gh/your-org/rule-engine)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## High-Level Summary

Generic Rule Engine is a high-performance, stateless rule evaluation service for real-time decisioning. It enables organizations to define, manage, and execute complex business rules and workflows with transactional integrity, hot in-memory caching, and robust API support. The engine is designed for scalability, reliability, and ease of integration in modern cloud-native environments.

---

## Key Architectural Features

- **Stateless Execution:** All reads are served from a hot in-memory cache for maximum throughput and low latency.
- **Polling-Based Cache Coherency:** Each node polls the database for checksum changes and atomically reloads only changed namespaces—no message bus required.
- **Transactional Integrity:** All state changes (especially publish operations) are atomic and managed by the database, ensuring consistency.
- **Clean Architecture:** Strict separation between Domain (business logic), Service (orchestration), and Handler (HTTP/API) layers for testability and maintainability.

---

## Prerequisites

- Go 1.24+
- PostgreSQL 15+
- Docker (optional)
- make

---

## Quick Start (5-Minute Setup)

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd rule-engine
   ```
2. **Copy example config:**
   ```bash
   cp configs/config.development.yaml.example configs/config.development.yaml
   ```
3. **Setup the database:**
   ```bash
   make db-setup
   ```
4. **Run the service:**
   ```bash
   make run
   ```
5. **Check health:**
   ```bash
   curl http://localhost:8080/health
   ```

---

## Docker Quick Start

**Using Docker Compose (Recommended):**
```bash
# Start the entire stack (PostgreSQL + Rule Engine)
docker-compose up -d

# Check logs
docker-compose logs -f rule-engine

# Stop the stack
docker-compose down
```

**Using Docker directly:**
```bash
# Build the image
docker build -t rule-engine .

# Run with environment variables
docker run -p 8080:8080 \
  -e RULE_ENGINE_DATABASE__HOST=your-db-host \
  -e RULE_ENGINE_DATABASE__PORT=5432 \
  -e RULE_ENGINE_DATABASE__NAME=rule_engine \
  -e RULE_ENGINE_DATABASE__USER=your-user \
  -e RULE_ENGINE_DATABASE__PASSWORD=your-password \
  rule-engine
```

**Docker Features:**
- Multi-stage build for optimized image size
- Non-root user for security
- Health checks for monitoring
- Alpine Linux base for minimal footprint
- Proper handling of timezone and certificates

---

## Core Concepts

- **Namespace:** An isolated container for all other entities (fields, functions, rules, workflows, etc.).
- **Field:** A typed data definition used as input for rules and functions.
- **Function:** A reusable computation (e.g., max, sum, avg, in) that operates on fields or values.
- **Rule:** A set of conditions (using fields/functions) that evaluates to true or false.
- **Terminal:** A declared end state for a workflow (e.g., "accept", "reject").
- **Workflow:** A directed graph of rules and terminals that guides execution to a final outcome.

---

## API Usage Overview & Documentation

- **Base URL:** `/v1`
- **Authentication:** All requests require `Authorization: Bearer <token>` header (JWT).

**Example: Create a Namespace**
```bash
curl -X POST http://localhost:8080/v1/namespaces \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"id": "my-namespace", "description": "Test namespace"}'
```

- **Full API documentation:** See [docs/API_DOCUMENTATION.md](docs/API_DOCUMENTATION.md)

---

## Configuration

- Configuration is loaded from the `configs/` directory (e.g., `configs/config.development.yaml`).
- Any config value can be overridden by environment variables prefixed with `RULE_ENGINE_` (e.g., `RULE_ENGINE_DATABASE__HOST=localhost`).

---

## Running Tests

- **Unit tests:**
  ```bash
  make test
  ```
- **End-to-end tests:**
  ```bash
  make test-e2e
  ```

---

## Test Case Coverage

The codebase includes comprehensive unit and integration tests across all major layers with **51.6% overall coverage**:

- **Domain Layer:** Extensive validation tests for all core models (namespace, field, function, rule, workflow, terminal) and workflow graph logic. **47.0% coverage**.
- **Repository Layer:** Full CRUD and edge case tests for all repositories (field, function, rule, workflow, terminal, cache), using mocks for database isolation. **83.0% coverage**.
- **Service Layer:** Unit tests for orchestration logic, including error handling, versioning, and business rules. **52.8% coverage**.
- **Handler Layer:** Complete coverage of all API endpoints, including success, validation, not found, and internal error scenarios. Input validation and error-wrapping are consistent and tested for all POST/PUT endpoints. **58.3% coverage**.
- **Admin & Execution:** Specialized tests for cache stats, cache refresh, and rule/workflow execution (including trace and error paths). **47.8% coverage**.
- **Metrics & Logging:** Unit tests for custom metrics and logging utilities. **87.3% coverage**.
- **Configuration & DB:** Tests for config loading, environment overrides, and database connection logic. **76.1% coverage** for config, **92.9% coverage** for database layer.

All tests are isolated (using mocks or in-memory config), fast, and CI-friendly. The API contract for error responses is consistent and predictable across all endpoints.

For more details, see the `internal/handlers/`, `internal/service/`, `internal/repository/`, and `internal/domain/` test files.

---

**Generic Rule Engine** — Fast, reliable, and maintainable rule execution for modern applications.
