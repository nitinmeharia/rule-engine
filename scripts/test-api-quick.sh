#!/bin/bash

# Generic Rule Engine - Quick API Testing Script
# This script runs essential API tests for quick validation

set -e

# Configuration
BASE_URL="http://localhost:8080"
JWT_SECRET="dev-secret-key-change-in-production"
CLIENT_ID="test-client"

# Use conda Python that has PyJWT installed
PYTHON="/opt/anaconda3/bin/python3"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS")
            echo -e "${GREEN}✓ PASS${NC}: $message"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            TOTAL_TESTS=$((TOTAL_TESTS + 1))
            ;;
        "FAIL")
            echo -e "${RED}✗ FAIL${NC}: $message"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            TOTAL_TESTS=$((TOTAL_TESTS + 1))
            ;;
        "INFO")
            echo -e "${BLUE}ℹ INFO${NC}: $message"
            ;;
    esac
}

# Function to generate JWT token using conda Python
generate_jwt_token() {
    local client_id=$1
    local role=$2
    
    if [ -x "$PYTHON" ]; then
        $PYTHON -c "
import jwt
import datetime

secret = '$JWT_SECRET'
client_id = '$client_id'
role = '$role'

# Use timezone-aware datetime to avoid deprecation warnings
now = datetime.datetime.now(datetime.timezone.utc)

claims = {
    'clientId': client_id,
    'role': role,
    'exp': now + datetime.timedelta(hours=24),
    'iat': now,
    'nbf': now
}

token = jwt.encode(claims, secret, algorithm='HS256')
print(token)
"
    else
        echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRJZCI6InRlc3QtY2xpZW50Iiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzUwNzM5MjAwMCwiaWF0IjoxNzUwNzM1NjAwMCwibmJmIjoxNzUwNzM1NjAwMH0.test-signature"
    fi
}

# Function to make HTTP requests
make_request() {
    local method=$1
    local endpoint=$2
    local token=$3
    local data=$4
    local expected_status=$5
    local test_name=$6
    
    local url="$BASE_URL$endpoint"
    local curl_cmd="curl -s -w '\n%{http_code}'"
    
    # Add method
    if [ "$method" != "GET" ]; then
        curl_cmd="$curl_cmd -X '$method'"
    fi
    
    # Add headers
    if [ -n "$token" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $token'"
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json'"
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    # Add URL
    curl_cmd="$curl_cmd '$url'"
    
    # Make the request and capture response
    local response
    local status_code
    
    response=$(eval $curl_cmd 2>/dev/null || echo -e "\n000")
    
    # Extract status code (last line) - macOS compatible
    status_code=$(echo "$response" | tail -n 1)
    # Extract response body (all lines except last) - macOS compatible
    response_body=$(echo "$response" | sed '$d')
    
    # Validate status code
    if [ "$status_code" = "$expected_status" ]; then
        print_status "PASS" "$test_name"
    else
        print_status "FAIL" "$test_name (Expected: $expected_status, Got: $status_code)"
    fi
}

# Function to check if server is running
check_server() {
    if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to clean up test data (namespaces, fields, functions, rules)
cleanup_test_data() {
    print_status "INFO" "Cleaning up test data"
    psql -U postgres -d rule_engine_dev -c "DELETE FROM rules WHERE namespace IN ('quick-test-ns', 'fields-quick-test'); DELETE FROM functions WHERE namespace IN ('quick-test-ns', 'fields-quick-test'); DELETE FROM fields WHERE namespace IN ('quick-test-ns', 'fields-quick-test'); DELETE FROM namespaces WHERE id IN ('quick-test-ns', 'fields-quick-test');" > /dev/null 2>&1 || true
}

# Minimal logical functions test (Rules API not implemented yet)
quick_functions_test() {
    print_status "INFO" "Quick functions test"
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "admin")

    # Create namespace
    local ns_data='{"id": "quick-test-ns", "description": "Quick test namespace"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$ns_data" "201" "Create quick-test-ns"

    # Create field
    local field_data='{"fieldId": "score", "type": "number", "description": "User score"}'
    make_request "POST" "/v1/namespaces/quick-test-ns/fields" "$admin_token" "$field_data" "201" "Create field: score"

    # Create function using the field (supported type: max)
    local fn_data='{"id": "max_score", "type": "max", "args": ["score"]}'
    make_request "POST" "/v1/namespaces/quick-test-ns/functions" "$admin_token" "$fn_data" "201" "Create function: max_score"
    make_request "POST" "/v1/namespaces/quick-test-ns/functions/max_score/publish" "$admin_token" "" "200" "Publish function: max_score"

    # Test function operations
    make_request "GET" "/v1/namespaces/quick-test-ns/functions" "$admin_token" "" "200" "List functions"
    make_request "GET" "/v1/namespaces/quick-test-ns/functions/max_score" "$admin_token" "" "200" "Get function: max_score"

    # Test edge case
    make_request "POST" "/v1/namespaces/quick-test-ns/functions" "$admin_token" '{"id":"max_score","type":"max","args":["score"]}' "409" "Create duplicate function"

    # Clean up
    make_request "DELETE" "/v1/namespaces/quick-test-ns" "$admin_token" "" "204" "Delete quick-test-ns"
}

# Main test execution
main() {
    echo "=========================================="
    echo "    Generic Rule Engine - Quick Tests"
    echo "=========================================="
    echo "Base URL: $BASE_URL"
    echo ""
    
    # Clean up any existing test data
    bash ./scripts/cleanup-test-data.sh
    
    # Check if server is running
    if ! check_server; then
        print_status "FAIL" "Server is not running at $BASE_URL"
        echo "Please start the server first: make run"
        exit 1
    fi
    
    print_status "INFO" "Server is running"
    
    # Generate tokens
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "admin")
    local viewer_token=$(generate_jwt_token "$CLIENT_ID" "viewer")
    
    # Quick health check
    make_request "GET" "/health" "" "" "200" "Health check"
    
    # Quick authentication test
    make_request "GET" "/v1/namespaces" "" "" "401" "Missing auth header"
    make_request "GET" "/v1/namespaces" "$admin_token" "" "200" "Valid admin token"
    
    # Quick namespace test
    local create_data='{"id": "quick-test-ns", "description": "Quick test namespace"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$create_data" "201" "Create namespace"
    make_request "GET" "/v1/namespaces/quick-test-ns" "$admin_token" "" "200" "Get namespace"
    make_request "DELETE" "/v1/namespaces/quick-test-ns" "$admin_token" "" "204" "Delete namespace"
    
    # Quick fields test
    local create_ns_data='{"id": "fields-quick-test", "description": "Namespace for quick fields test"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$create_ns_data" "201" "Create namespace for fields"
    
    local create_field_data='{"fieldId": "test_field", "type": "string", "description": "Test field"}'
    make_request "POST" "/v1/namespaces/fields-quick-test/fields" "$admin_token" "$create_field_data" "201" "Create field"
    make_request "GET" "/v1/namespaces/fields-quick-test/fields" "$admin_token" "" "200" "List fields"
    
    # Test RBAC
    make_request "POST" "/v1/namespaces/fields-quick-test/fields" "$viewer_token" "$create_field_data" "403" "Viewer cannot create field"
    make_request "GET" "/v1/namespaces/fields-quick-test/fields" "$viewer_token" "" "200" "Viewer can list fields"
    
    # Cleanup
    make_request "DELETE" "/v1/namespaces/fields-quick-test" "$admin_token" "" "204" "Cleanup test namespace"
    
    # Quick logical functions test
    quick_functions_test
    
    # Print summary
    echo ""
    echo "=========================================="
    echo "           QUICK TEST SUMMARY"
    echo "=========================================="
    echo "Total Tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}All quick tests passed! 🎉${NC}"
        exit 0
    else
        echo -e "${RED}Some quick tests failed! ❌${NC}"
        exit 1
    fi
}

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is required but not installed${NC}"
    exit 1
fi

# Run the main function
main "$@" 