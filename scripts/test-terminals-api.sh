#!/bin/bash

# Terminal API Test Script
# Tests all Terminal endpoints with proper authentication and error handling

set -e

# Configuration
API_BASE_URL="http://localhost:8080"
NAMESPACE_ID="test-namespace-$(date +%s)"
TERMINAL_ID="test-terminal-$(date +%s)"
JWT_TOKEN=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test function
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local test_name=$5
    
    log_info "Testing: $test_name"
    log_info "Endpoint: $method $endpoint"
    
    if [ -n "$data" ]; then
        log_info "Data: $data"
    fi
    
    # Make the request
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -d "$data" \
            "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$API_BASE_URL$endpoint")
    fi
    
    # Extract status code and body
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    log_info "Response Status: $http_code"
    log_info "Response Body: $body"
    
    # Check if status code matches expected
    if [ "$http_code" -eq "$expected_status" ]; then
        log_success "✓ Test passed: Expected $expected_status, got $http_code"
    else
        log_error "✗ Test failed: Expected $expected_status, got $http_code"
        return 1
    fi
    
    echo ""
}

# Setup: Generate JWT token
setup() {
    log_info "Setting up test environment..."
    
    # Generate JWT token
    JWT_TOKEN=$(python3 scripts/generate-jwt.py --role admin --client-id test-client)
    if [ -z "$JWT_TOKEN" ]; then
        log_error "Failed to generate JWT token"
        exit 1
    fi
    
    log_success "JWT token generated successfully"
    log_info "Namespace ID: $NAMESPACE_ID"
    log_info "Terminal ID: $TERMINAL_ID"
    echo ""
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test data..."
    
    # Delete terminal if it exists
    curl -s -X DELETE \
        -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_BASE_URL/v1/namespaces/$NAMESPACE_ID/terminals/$TERMINAL_ID" > /dev/null 2>&1 || true
    
    # Delete namespace if it exists
    curl -s -X DELETE \
        -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_BASE_URL/v1/namespaces/$NAMESPACE_ID" > /dev/null 2>&1 || true
    
    log_success "Cleanup completed"
}

# Test Terminal API endpoints
test_terminals_api() {
    log_info "Starting Terminal API tests..."
    echo ""
    
    # 1. Create namespace first (required for terminal operations)
    log_info "=== Step 1: Create namespace ==="
    test_endpoint "POST" "/v1/namespaces" \
        "{\"id\":\"$NAMESPACE_ID\",\"description\":\"Test namespace for terminal testing\"}" \
        201 "Create namespace"
    
    # 2. List terminals (should be empty initially)
    log_info "=== Step 2: List terminals (empty) ==="
    test_endpoint "GET" "/v1/namespaces/$NAMESPACE_ID/terminals" \
        "" 200 "List terminals (empty)"
    
    # 3. Create terminal
    log_info "=== Step 3: Create terminal ==="
    test_endpoint "POST" "/v1/namespaces/$NAMESPACE_ID/terminals" \
        "{\"terminalId\":\"$TERMINAL_ID\"}" \
        201 "Create terminal"
    
    # 4. Get terminal
    log_info "=== Step 4: Get terminal ==="
    test_endpoint "GET" "/v1/namespaces/$NAMESPACE_ID/terminals/$TERMINAL_ID" \
        "" 200 "Get terminal"
    
    # 5. List terminals (should now contain the created terminal)
    log_info "=== Step 5: List terminals (with data) ==="
    test_endpoint "GET" "/v1/namespaces/$NAMESPACE_ID/terminals" \
        "" 200 "List terminals (with data)"
    
    # 6. Create duplicate terminal (should fail)
    log_info "=== Step 6: Create duplicate terminal (should fail) ==="
    test_endpoint "POST" "/v1/namespaces/$NAMESPACE_ID/terminals" \
        "{\"terminalId\":\"$TERMINAL_ID\"}" \
        409 "Create duplicate terminal"
    
    # 7. Get non-existent terminal (should fail)
    log_info "=== Step 7: Get non-existent terminal (should fail) ==="
    test_endpoint "GET" "/v1/namespaces/$NAMESPACE_ID/terminals/non-existent-terminal" \
        "" 404 "Get non-existent terminal"
    
    # 8. Create terminal with invalid data (should fail)
    log_info "=== Step 8: Create terminal with invalid data (should fail) ==="
    test_endpoint "POST" "/v1/namespaces/$NAMESPACE_ID/terminals" \
        "{\"terminalId\":\"\"}" \
        400 "Create terminal with empty ID"
    
    # 9. Create terminal with missing data (should fail)
    log_info "=== Step 9: Create terminal with missing data (should fail) ==="
    test_endpoint "POST" "/v1/namespaces/$NAMESPACE_ID/terminals" \
        "{}" \
        400 "Create terminal with missing ID"
    
    # 10. Delete terminal
    log_info "=== Step 10: Delete terminal ==="
    test_endpoint "DELETE" "/v1/namespaces/$NAMESPACE_ID/terminals/$TERMINAL_ID" \
        "" 204 "Delete terminal"
    
    # 11. Get deleted terminal (should fail)
    log_info "=== Step 11: Get deleted terminal (should fail) ==="
    test_endpoint "GET" "/v1/namespaces/$NAMESPACE_ID/terminals/$TERMINAL_ID" \
        "" 404 "Get deleted terminal"
    
    # 12. Delete non-existent terminal (should fail)
    log_info "=== Step 12: Delete non-existent terminal (should fail) ==="
    test_endpoint "DELETE" "/v1/namespaces/$NAMESPACE_ID/terminals/non-existent-terminal" \
        "" 404 "Delete non-existent terminal"
    
    # 13. List terminals (should be empty again)
    log_info "=== Step 13: List terminals (empty after deletion) ==="
    test_endpoint "GET" "/v1/namespaces/$NAMESPACE_ID/terminals" \
        "" 200 "List terminals (empty after deletion)"
    
    # 14. Test with invalid namespace
    log_info "=== Step 14: Test with invalid namespace ==="
    test_endpoint "GET" "/v1/namespaces/invalid-namespace/terminals" \
        "" 404 "List terminals with invalid namespace"
    
    # 15. Test without authentication
    log_info "=== Step 15: Test without authentication ==="
    response=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/v1/namespaces/$NAMESPACE_ID/terminals")
    http_code=$(echo "$response" | tail -n1)
    if [ "$http_code" -eq 401 ]; then
        log_success "✓ Authentication test passed: Expected 401, got $http_code"
    else
        log_error "✗ Authentication test failed: Expected 401, got $http_code"
    fi
    echo ""
    
    # 16. Test with invalid JWT token
    log_info "=== Step 16: Test with invalid JWT token ==="
    response=$(curl -s -w "\n%{http_code}" -X GET \
        -H "Authorization: Bearer invalid-token" \
        "$API_BASE_URL/v1/namespaces/$NAMESPACE_ID/terminals")
    http_code=$(echo "$response" | tail -n1)
    if [ "$http_code" -eq 401 ]; then
        log_success "✓ Invalid JWT test passed: Expected 401, got $http_code"
    else
        log_error "✗ Invalid JWT test failed: Expected 401, got $http_code"
    fi
    echo ""
}

# Main execution
main() {
    log_info "Starting Terminal API End-to-End Tests"
    log_info "API Base URL: $API_BASE_URL"
    echo ""
    
    # Check if server is running
    if ! curl -s "$API_BASE_URL/health" > /dev/null; then
        log_error "Server is not running at $API_BASE_URL"
        log_info "Please start the server first: make run"
        exit 1
    fi
    
    # Setup test environment
    setup
    
    # Run tests
    test_terminals_api
    
    # Cleanup
    cleanup
    
    log_success "All Terminal API tests completed successfully!"
}

# Run main function
main "$@" 