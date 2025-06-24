# Rule Engine Testing Scripts

This directory contains comprehensive testing scripts for the Rule Engine API, including specialized tests for workflow terminal validation.

## Scripts Overview

### Core Testing Scripts

1. **`test-api-e2e.sh`** - Complete end-to-end API testing suite
   - Tests all API endpoints with authentication and error handling
   - Includes comprehensive workflow testing with terminal validation
   - Validates RBAC, error handling, and edge cases

2. **`test-workflows-api.sh`** - Enhanced workflow testing with terminal validation
   - **NEW**: Comprehensive terminal validation tests
   - Tests workflow lifecycle (create, publish, deactivate, delete)
   - Validates that all workflow paths end with terminal steps
   - Tests cyclic dependencies and error conditions
   - Includes detailed error message validation

3. **`test-rules-api.sh`** - Rules API testing
   - Tests rule creation, validation, and publishing
   - Validates dependency checking and RBAC

4. **`test-functions-api.sh`** - Functions API testing
   - Tests function creation, publishing, and execution
   - Validates supported function types and error handling

5. **`test-terminals-api.sh`** - Terminals API testing
   - Tests terminal CRUD operations
   - Validates parent namespace relationships

### Utility Scripts

6. **`generate-jwt.py`** - JWT token generator
   - Generates authentication tokens for API testing
   - Supports different roles (admin, viewer, executor)

7. **`setup-test-db.sh`** - Database setup
   - Sets up test database with required schema
   - Runs migrations and initializes test data

8. **`cleanup-test-data.sh`** - Test data cleanup
   - Cleans up test data after testing
   - Removes test namespaces and related data

## Prerequisites

1. **Python 3.x** with PyJWT installed:
   ```bash
   pip install PyJWT
   ```

2. **Server running** on `http://localhost:8080`

3. **Database** properly configured and accessible

## Quick Start

### 1. Run Complete E2E Tests
```bash
# Run all tests including enhanced workflow validation
./scripts/test-api-e2e.sh
```

### 2. Run Specific Test Suites
```bash
# Test workflows with comprehensive terminal validation
./scripts/test-workflows-api.sh

# Test rules API
./scripts/test-rules-api.sh

# Test functions API
./scripts/test-functions-api.sh

# Test terminals API
./scripts/test-terminals-api.sh
```

### 3. Generate JWT Tokens
```bash
# Generate admin token
python3 scripts/generate-jwt.py --role admin --client-id test-client

# Generate viewer token
python3 scripts/generate-jwt.py --role viewer --client-id test-client

# Generate executor token
python3 scripts/generate-jwt.py --role executor --client-id test-client
```

## Workflow Terminal Validation

The enhanced workflow testing includes comprehensive validation to ensure all workflow paths end with terminal steps:

### Test Scenarios

1. **Valid Workflows**
   - Simple workflows with direct terminal paths
   - Complex workflows with multiple branches
   - Single terminal workflows

2. **Invalid Workflows (Expected to Fail)**
   - Missing `onTrue` path
   - Missing `onFalse` path
   - Paths leading to non-terminal steps
   - Steps leading to non-existent steps
   - Unknown step types
   - Malformed step data

3. **Error Validation**
   - Validates specific error messages
   - Ensures actionable feedback for users
   - Tests both status codes and error content

### Expected Error Messages

The validation returns specific, actionable error messages:

- `"Validation Error: The 'onTrue' path for step 'step-name' does not lead to a terminal"`
- `"Validation Error: The 'onFalse' path for step 'step-name' does not lead to a terminal"`
- `"Validation Error: Step 'step-name' is invalid or missing and does not lead to a terminal"`

## Test Results

### Success Indicators
- ✅ All tests pass with expected status codes
- ✅ Error messages match expected patterns
- ✅ Terminal validation correctly identifies invalid workflows
- ✅ Valid workflows are accepted

### Failure Indicators
- ❌ Unexpected status codes
- ❌ Missing or incorrect error messages
- ❌ Valid workflows rejected
- ❌ Invalid workflows accepted

## Troubleshooting

### Common Issues

1. **Server Not Running**
   ```bash
   # Start the server first
   go run ./cmd/api
   ```

2. **Database Connection Issues**
   ```bash
   # Check database configuration
   ./scripts/setup-test-db.sh
   ```

3. **JWT Token Issues**
   ```bash
   # Regenerate tokens
   python3 scripts/generate-jwt.py --role admin --client-id test-client
   ```

4. **Permission Issues**
   ```bash
   # Make scripts executable
   chmod +x scripts/*.sh
   ```

### Debug Mode

To run tests with verbose output:
```bash
# Enable debug output
export DEBUG=1
./scripts/test-workflows-api.sh
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: API Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Setup Python
        uses: actions/setup-python@v2
        with:
          python-version: '3.9'
      - name: Install dependencies
        run: |
          pip install PyJWT
          go mod download
      - name: Start server
        run: go run ./cmd/api &
      - name: Wait for server
        run: sleep 10
      - name: Run tests
        run: ./scripts/test-api-e2e.sh
```

### Docker Integration
```bash
# Run tests in Docker container
docker run --rm -v $(pwd):/app -w /app golang:1.21 bash -c "
  go mod download &&
  go run ./cmd/api &
  sleep 10 &&
  ./scripts/test-api-e2e.sh
"
```

## Contributing

When adding new tests:

1. Follow the existing pattern for test functions
2. Include both positive and negative test cases
3. Validate error messages for failure scenarios
4. Update this README with new test descriptions
5. Ensure tests are idempotent and clean up after themselves

## Security Considerations

- Tests use dedicated test namespaces to avoid conflicts
- JWT tokens have limited scope and expiration
- Test data is cleaned up after each run
- No production data is accessed during testing 