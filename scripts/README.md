# Rule Engine Testing Scripts

[![Go](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![Python](https://img.shields.io/badge/Python-3.8+-green.svg)](https://python.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Comprehensive testing suite for the Rule Engine API with end-to-end validation, cache refresh testing, and automated workflow verification.

## 📋 Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Script Categories](#script-categories)
- [Testing Workflows](#testing-workflows)
- [Cache Refresh Testing](#cache-refresh-testing)
- [Troubleshooting](#troubleshooting)
- [Local Development Workflow](#local-development-workflow)

## 🎯 Overview

This directory contains a comprehensive testing framework for the Rule Engine API, designed to validate:

- **API Endpoints**: All CRUD operations with authentication and RBAC
- **Workflow Validation**: Terminal step validation and path verification
- **Cache Refresh**: Real-time cache synchronization and consistency
- **Error Handling**: Comprehensive error scenarios and edge cases
- **Integration Testing**: End-to-end workflow validation

## 🔧 Prerequisites

### Required Software
- **Go 1.24+** - For running the API server
- **Python 3.8+** - For JWT token generation and test utilities
- **PostgreSQL** - Database server
- **Make** - For build automation

### Python Dependencies
```bash
pip install PyJWT
```

### Environment Setup
```bash
# Clone and setup
git clone <repository-url>
cd rule-engine

# Install Go dependencies
go mod download

# Setup test database
./scripts/setup-test-db.sh
```

## 🚀 Quick Start

### 1. Start the API Server
```bash
# Development mode
make run-dev

# Or directly
go run ./cmd/api
```

### 2. Run Complete E2E Test Suite
```bash
# Full end-to-end testing with cache refresh validation
./scripts/test-api-e2e.sh
```

### 3. Run Quick API Tests
```bash
# Fast API validation without cache refresh
./scripts/test-api-quick.sh
```

### 4. Test Specific Components
```bash
# Test cache refresh functionality
./scripts/test-cache-refresh-e2e.sh

# Test individual API modules
./scripts/test-workflows-api.sh
./scripts/test-rules-api.sh
./scripts/test-functions-api.sh
./scripts/test-terminals-api.sh
```

## 📁 Script Categories

### 🔄 Cache Refresh Testing
| Script | Purpose | Duration | Coverage |
|--------|---------|----------|----------|
| `test-cache-refresh-e2e.sh` | Complete cache refresh E2E testing | ~1-2 min | Full cache lifecycle |
| `test-cache-refresh-integration.sh` | Integration cache refresh tests | ~30 sec | Cache consistency |
| `test-cache-refresh.sh` | Basic cache refresh validation | ~15 sec | Core functionality |

### 🧪 API Testing
| Script | Purpose | Duration | Coverage |
|--------|---------|----------|----------|
| `test-api-e2e.sh` | Complete API E2E testing | ~1 min | All endpoints + cache |
| `test-api-quick.sh` | Fast API validation | ~30 sec | Core endpoints |
| `test-workflows-api.sh` | Workflow-specific testing | ~30 sec | Workflow lifecycle |
| `test-rules-api.sh` | Rules API testing | ~15 sec | Rule CRUD operations |
| `test-functions-api.sh` | Functions API testing | ~15 sec | Function operations |
| `test-terminals-api.sh` | Terminals API testing | ~15 sec | Terminal management |

### 🛠️ Utility Scripts
| Script | Purpose | Usage |
|--------|---------|-------|
| `generate-jwt.py` | JWT token generation | Authentication for tests |
| `setup-test-db.sh` | Database initialization | Test environment setup |
| `cleanup-test-data.sh` | Test data cleanup | Post-test cleanup |

## 🔄 Testing Workflows

### Workflow Terminal Validation

The workflow testing includes comprehensive validation to ensure all workflow paths end with terminal steps:

#### Valid Workflow Patterns
```json
{
  "steps": {
    "start": {
      "type": "condition",
      "onTrue": "terminal-success",
      "onFalse": "terminal-failure"
    },
    "terminal-success": {
      "type": "terminal",
      "result": "success"
    },
    "terminal-failure": {
      "type": "terminal", 
      "result": "failure"
    }
  }
}
```

#### Test Scenarios

**✅ Valid Workflows**
- Simple workflows with direct terminal paths
- Complex workflows with multiple branches
- Single terminal workflows
- Nested conditional workflows

**❌ Invalid Workflows (Expected to Fail)**
- Missing `onTrue` or `onFalse` paths
- Paths leading to non-terminal steps
- Steps leading to non-existent steps
- Unknown step types
- Malformed step data

#### Error Message Validation
```bash
# Expected error patterns
"Validation Error: The 'onTrue' path for step 'step-name' does not lead to a terminal"
"Validation Error: The 'onFalse' path for step 'step-name' does not lead to a terminal"
"Validation Error: Step 'step-name' is invalid or missing and does not lead to a terminal"
```

## 🔄 Cache Refresh Testing

### Cache Refresh Lifecycle

The cache refresh testing validates the complete cache synchronization process:

1. **Initial State**: Empty cache
2. **Data Creation**: Create test namespaces, rules, functions
3. **Publishing**: Trigger cache refresh via publish operations
4. **Verification**: Check cache consistency via admin endpoints
5. **Cleanup**: Remove test data

### Cache Refresh Test Flow
```bash
# 1. Start service with cache refresh enabled
make run-dev

# 2. Run cache refresh E2E test
./scripts/test-cache-refresh-e2e.sh

# 3. Verify cache stats via admin endpoint
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/admin/cache/stats
```

### Expected Cache Behavior
- **Before Publishing**: Cache contains only existing data
- **After Publishing**: Cache includes new/updated data
- **Consistency**: Cache matches database state
- **Performance**: Cache refresh completes within expected time

## 🛠️ Troubleshooting

### Common Issues & Solutions

#### 1. Server Connection Issues
```bash
# Check if server is running
curl http://localhost:8080/health

# Start server if not running
make run-dev
```

#### 2. Database Connection Problems
```bash
# Verify database setup
./scripts/setup-test-db.sh

# Check database configuration
cat configs/config.development.yaml
```

#### 3. JWT Token Issues
```bash
# Regenerate tokens
python3 scripts/generate-jwt.py --role admin --client-id test-client

# Verify token format
echo $ADMIN_TOKEN | cut -d'.' -f2 | base64 -d
```

#### 4. Permission Issues
```bash
# Make scripts executable
chmod +x scripts/*.sh

# Check script permissions
ls -la scripts/
```

#### 5. Port Conflicts
```bash
# Check port usage
lsof -i :8080

# Kill conflicting processes
pkill -f "go run"
```

### Debug Mode

Enable verbose output for detailed debugging:
```bash
# Enable debug output
export DEBUG=1
export VERBOSE=1

# Run tests with debug info
./scripts/test-api-e2e.sh
```

### Test Data Cleanup

Clean up test data after debugging:
```bash
# Clean all test data
./scripts/cleanup-test-data.sh

# Or manually clean specific namespaces
curl -X DELETE \
     -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/v1/namespaces/test-namespace
```

## 🔄 Local Development Workflow

### Available Make Targets

#### Development Commands
```bash
# Start development server
make run-dev

# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run API tests (requires server running)
make test-api

# Run quick API tests
make test-api-quick

# Generate JWT token
make generate-jwt
```

#### Code Quality
```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Pre-commit checks
make pre-commit
```

#### Database Management
```bash
# Run migrations
make migrate

# Check migration status
make migrate-status

# Reset database
make db-reset

# Setup local database
make setup-local-db
```

#### Docker Commands
```bash
# Build Docker image
make docker-build

# Run in Docker
make docker-run

# Start with docker-compose
make docker-compose-up

# Stop docker-compose
make docker-compose-down
```

#### Dependencies
```bash
# Download dependencies
make deps-download

# Tidy dependencies
make deps-tidy

# Verify dependencies
make deps-verify

# Upgrade dependencies
make deps-upgrade
```

### Development Workflow Example

```bash
# 1. Start development environment
make run-dev

# 2. Run tests in parallel
make test-api & make test-coverage & wait

# 3. Check test results
echo "API Tests: $?"
echo "Coverage Tests: $?"
```

## 📊 Test Results Interpretation

### Success Indicators
- ✅ All HTTP status codes match expectations
- ✅ Error messages contain actionable information
- ✅ Cache refresh completes successfully
- ✅ Workflow validation correctly identifies issues
- ✅ All API endpoints respond correctly

### Failure Indicators
- ❌ Unexpected HTTP status codes (500, 404, 403)
- ❌ Missing or incorrect error messages
- ❌ Cache refresh timeouts or failures
- ❌ Valid workflows rejected
- ❌ Invalid workflows accepted

### Test Coverage
- **Unit Tests**: Comprehensive coverage of domain logic, services, and repositories
- **Integration Tests**: End-to-end API validation with real database interactions
- **Cache Tests**: Validation of cache refresh and consistency mechanisms
- **Error Handling**: Verification of proper error responses and validation

## 📚 Additional Resources

- [API Documentation](../docs/API_DOCUMENTATION.md) - Complete API reference
- [Main README](../README.md) - Project overview and architecture

---

**Need Help?** Check the [troubleshooting section](#troubleshooting) or create an issue with detailed error logs. 