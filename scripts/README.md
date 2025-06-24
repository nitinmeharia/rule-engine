# API Testing Scripts

This directory contains comprehensive testing scripts for the Generic Rule Engine APIs.

## Scripts Overview

### 1. `test-api-e2e.sh` - Comprehensive End-to-End Tests
The main testing script that covers all API functionality with detailed validation.

**Features:**
- Health endpoint testing
- Authentication and authorization testing
- Namespaces API (CRUD operations)
- Fields API (CRUD operations)
- Role-based access control (RBAC) testing
- Error handling validation
- Edge cases testing
- Performance testing
- Colored output with detailed test results

**Usage:**
```bash
# Run with server already running
./scripts/test-api-e2e.sh

# Or use Makefile target
make test-api
```

### 2. `test-api-quick.sh` - Quick Validation Tests
A lightweight script for fast validation during development.

**Features:**
- Essential functionality only
- Fast execution
- Basic health checks
- Core API validation
- RBAC testing

**Usage:**
```bash
# Run quick tests
./scripts/test-api-quick.sh

# Or use Makefile target
make test-api-quick
```

**Example Output:**
```
==========================================
    Generic Rule Engine - Quick Tests
==========================================
Base URL: http://localhost:8080

ℹ INFO: Server is running
✓ PASS: Health check
✓ PASS: Missing auth header
✓ PASS: Valid admin token
✓ PASS: Create namespace
✓ PASS: Get namespace
✓ PASS: Delete namespace
✓ PASS: Create namespace for fields
✓ PASS: Create field
✓ PASS: List fields
✓ PASS: Viewer cannot create field
✓ PASS: Viewer can list fields
✓ PASS: Cleanup test namespace

==========================================
           QUICK TEST SUMMARY
==========================================
Total Tests: 12
Passed: 12
Failed: 0

All quick tests passed! 🎉
```

### 3. `generate-jwt.py` - JWT Token Generator
Python script for generating JWT tokens for testing.

**Features:**
- Generate tokens with different roles
- Decode and verify existing tokens
- Multiple output formats
- Configurable expiration

**Usage:**
```bash
# Generate admin token
python3 scripts/generate-jwt.py --role admin

# Generate viewer token
python3 scripts/generate-jwt.py --role viewer

# Generate token with custom client ID
python3 scripts/generate-jwt.py --client-id my-client --role admin

# Get curl command with token
python3 scripts/generate-jwt.py --role admin --format curl

# Decode existing token
python3 scripts/generate-jwt.py --decode <token>

# Or use Makefile target
make generate-jwt
```

## Prerequisites

### Required Tools
- `curl` - For HTTP requests
- `bash` - For shell scripts
- `python3` - For JWT token generation (optional, fallback available)

### Python Dependencies
If using Python for JWT generation:
```bash
pip install PyJWT
```

> **Note:** If you see `Error: PyJWT library is required. Install it with: pip install PyJWT`, ensure you are using the correct Python environment (e.g., your virtualenv or conda environment) and that `python3` points to the Python where PyJWT is installed.

### Server Setup
Ensure the rule engine server is running:
```bash
# Start the server
make run

# Or run directly
go run ./cmd/api
```

## Makefile Targets

The following Makefile targets are available for easy testing:

```bash
# Run comprehensive API tests (server must be running)
make test-api

# Run full end-to-end tests (starts server automatically)
make test-api-e2e

# Run quick API tests
make test-api-quick

# Generate JWT token
make generate-jwt
```

## Test Coverage

### Health Endpoint
- ✅ Health check without authentication
- ✅ Response format validation

### Authentication
- ✅ Missing authorization header
- ✅ Invalid token format
- ✅ Invalid JWT signature
- ✅ Expired tokens (when available)

### Namespaces API
- ✅ List namespaces (all roles)
- ✅ Create namespace (admin only)
- ✅ Get specific namespace
- ✅ Delete namespace (admin only)
- ✅ Duplicate namespace creation
- ✅ Invalid namespace data
- ✅ Non-existent namespace access

### Fields API
- ✅ List fields in namespace
- ✅ Create field (admin only)
- ✅ Duplicate field creation
- ✅ Invalid field data
- ✅ Fields in non-existent namespace
- ✅ Multiple fields in namespace

### Role-Based Access Control (RBAC)
- ✅ Admin role permissions
- ✅ Viewer role permissions
- ✅ Executor role permissions
- ✅ Forbidden operations validation

### Error Handling
- ✅ Malformed JSON
- ✅ Missing required fields
- ✅ Invalid field types
- ✅ Validation errors

### Edge Cases
- ✅ Very long input values
- ✅ Special characters
- ✅ Empty values
- ✅ Boundary conditions

### Performance
- ✅ Response time validation
- ✅ Concurrent request handling

## Configuration

### Environment Variables
The scripts use the following configuration (can be modified in the scripts):

```bash
BASE_URL="http://localhost:8080"
JWT_SECRET="dev-secret-key-change-in-production"
CLIENT_ID="test-client"
```

### Customization
You can modify the scripts to:
- Change the base URL for different environments
- Use different JWT secrets
- Add custom test cases
- Modify expected responses

## Output Format

### Success Output
```
✓ PASS: Health check (no auth)
✓ PASS: Create namespace (admin)
✓ PASS: List fields (viewer)
```

### Failure Output
```
✗ FAIL: Create namespace (viewer - forbidden) (Expected: 403, Got: 200)
  Response: {"success":true,"namespace":{...}}
```

### Summary
```
==========================================
           TEST SUMMARY
==========================================
Total Tests: 45
Passed: 43
Failed: 2

Some tests failed! ❌
```

## Troubleshooting

### Common Issues

1. **Server not running**
   ```
   Error: Server does not appear to be running at http://localhost:8080
   ```
   **Solution:** Start the server with `make run`

2. **Database connection issues**
   ```
   Error: Failed to connect to database
   ```
   **Solution:** Ensure PostgreSQL is running and migrations are applied

3. **JWT token issues**
   ```
   Error: Invalid JWT token
   ```
   **Solution:** Check JWT secret configuration matches server config

4. **Permission denied**
   ```
   Error: Permission denied
   ```
   **Solution:** Make scripts executable: `chmod +x scripts/*.sh`

5. **PyJWT not found or Python environment issues**
   ```
   Error: PyJWT library is required. Install it with: pip install PyJWT
   ```
   **Solution:**
   - Run `pip install PyJWT` (or `pip3 install PyJWT`)
   - Make sure you are using the correct Python environment (e.g., activate your virtualenv or conda environment)
   - Check that `python3` points to the Python where PyJWT is installed: `python3 -m pip show PyJWT`
   - If using Anaconda, try `conda install pyjwt`

### Debug Mode
For debugging, you can modify the scripts to show more verbose output:
- Add `set -x` at the beginning of bash scripts
- Use `--verbose` flag with curl commands
- Enable debug logging in the server

## Continuous Integration

These scripts can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions step
- name: Run API Tests
  run: |
    make run &
    sleep 10
    make test-api
```

## Contributing

When adding new API endpoints, update the testing scripts to include:
1. Happy path testing
2. Error case testing
3. RBAC validation
4. Edge case testing

Follow the existing patterns in the scripts for consistency. 