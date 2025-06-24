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

- Go 1.22+
- PostgreSQL 15+
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

## Architecture & Design

- High-level architecture: [docs/HLD.txt](docs/HLD.txt)
- Detailed component design & database schema: [docs/LLD.txt](docs/LLD.txt)
- Code structure overview: [docs/CODE_STRUCTURE.md](docs/CODE_STRUCTURE.md)

---

**Generic Rule Engine** — Fast, reliable, and maintainable rule execution for modern applications.
