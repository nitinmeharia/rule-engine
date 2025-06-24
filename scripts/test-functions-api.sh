#!/bin/bash

# Generic Rule Engine - Functions API Testing Script
# This script tests the Functions API endpoints

set -e

# Configuration
BASE_URL="http://localhost:8080"
JWT_SECRET="dev-secret-key-change-in-production"
CLIENT_ID="test-client"
ADMIN_ROLE="admin"
VIEWER_ROLE="viewer"

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
        echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRJZCI6InRlc3QtY2xpZW50Iiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzUwNzM5MjAwMCwiaWF0IjoxNzUwNzM1NjAwMCwibmJmIjoxNzUwNzM1NjAwfQ.Ej8Ej8Ej8Ej8Ej8Ej8Ej8Ej8Ej8Ej8Ej8Ej8Ej8"
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
    response=$(eval $curl_cmd)
    
    # Extract status code (last line)
    local status_code
    status_code=$(echo "$response" | tail -n1)
    
    # Extract response body (all lines except last)
    local body
    body=$(echo "$response" | sed '$d')
    
    # Check if status code matches expected
    if [ "$status_code" = "$expected_status" ]; then
        print_status "PASS" "$test_name"
    else
        print_status "FAIL" "$test_name (Expected: $expected_status, Got: $status_code)"
        echo "  Response: $body"
    fi
}

# Main test execution
echo "=========================================="
echo "    Functions API Testing"
echo "=========================================="

# Clean up any existing test data
bash ./scripts/cleanup-test-data.sh

# Generate tokens
print_status "INFO" "Generating JWT tokens"
admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
viewer_token=$(generate_jwt_token "$CLIENT_ID" "$VIEWER_ROLE")

# Test namespace creation (required for function tests)
print_status "INFO" "Creating test namespace"
make_request "POST" "/v1/namespaces" "$admin_token" '{"id":"test-functions","description":"Test namespace for functions"}' "201" "Create test namespace"

# Test Functions API
echo ""
echo "=========================================="
echo "           Functions API Tests"
echo "=========================================="

# Test 1: Create numeric function (max)
make_request "POST" "/v1/namespaces/test-functions/functions" "$admin_token" '{"id":"max_income","type":"max","args":["salary","bonus"]}' "201" "Create max function"

# Test 2: Create numeric function (sum)
make_request "POST" "/v1/namespaces/test-functions/functions" "$admin_token" '{"id":"total_income","type":"sum","args":["salary","bonus","commission"]}' "201" "Create sum function"

# Test 3: Create numeric function (avg)
make_request "POST" "/v1/namespaces/test-functions/functions" "$admin_token" '{"id":"avg_salary","type":"avg","args":["salary1","salary2","salary3"]}' "201" "Create avg function"

# Test 4: Create in function
make_request "POST" "/v1/namespaces/test-functions/functions" "$admin_token" '{"id":"valid_occupations","type":"in","values":["salaried","self_employed","business"]}' "201" "Create in function"

# Test 5: List functions
make_request "GET" "/v1/namespaces/test-functions/functions" "$viewer_token" "" "200" "List functions"

# Test 6: Get specific function
make_request "GET" "/v1/namespaces/test-functions/functions/max_income" "$viewer_token" "" "200" "Get max_income function"

# Test 7: Update function draft
make_request "PUT" "/v1/namespaces/test-functions/functions/max_income/versions/draft" "$admin_token" '{"type":"max","args":["salary","bonus","commission","overtime"]}' "200" "Update function draft"

# Test 8: Publish function
make_request "POST" "/v1/namespaces/test-functions/functions/max_income/publish" "$admin_token" "" "200" "Publish function"

# Test 9: Get published function
make_request "GET" "/v1/namespaces/test-functions/functions/max_income" "$viewer_token" "" "200" "Get published function"

# Test 10: Create duplicate function (should fail)
make_request "POST" "/v1/namespaces/test-functions/functions" "$admin_token" '{"id":"max_income","type":"max","args":["salary"]}' "409" "Create duplicate function"

# Test 11: Create function with invalid type
make_request "POST" "/v1/namespaces/test-functions/functions" "$admin_token" '{"id":"invalid_func","type":"invalid","args":["salary"]}' "400" "Create function with invalid type"

# Test 12: Create function without args for numeric type
make_request "POST" "/v1/namespaces/test-functions/functions" "$admin_token" '{"id":"empty_max","type":"max","args":[]}' "400" "Create numeric function without args"

# Test 13: Create function without values for in type
make_request "POST" "/v1/namespaces/test-functions/functions" "$admin_token" '{"id":"empty_in","type":"in","values":[]}' "400" "Create in function without values"

# Test 14: Get non-existent function
make_request "GET" "/v1/namespaces/test-functions/functions/non_existent" "$viewer_token" "" "404" "Get non-existent function"

# Test 15: Update non-existent function
make_request "PUT" "/v1/namespaces/test-functions/functions/non_existent/versions/draft" "$admin_token" '{"type":"max","args":["salary"]}' "404" "Update non-existent function"

# Test 16: Publish non-existent function
make_request "POST" "/v1/namespaces/test-functions/functions/non_existent/publish" "$admin_token" "" "404" "Publish non-existent function"

# Test 17: Access without authentication
make_request "GET" "/v1/namespaces/test-functions/functions" "" "" "401" "Access without authentication"

# Test 18: Access with viewer role (should work for read operations)
make_request "GET" "/v1/namespaces/test-functions/functions/max_income" "$viewer_token" "" "200" "Viewer can read function"

# Test 19: Viewer tries to create function (should fail)
make_request "POST" "/v1/namespaces/test-functions/functions" "$viewer_token" '{"id":"viewer_func","type":"max","args":["salary"]}' "403" "Viewer cannot create function"

# Test 20: Create function in non-existent namespace
make_request "POST" "/v1/namespaces/non_existent/functions" "$admin_token" '{"id":"test_func","type":"max","args":["salary"]}' "500" "Create function in non-existent namespace"

# Print test summary
echo ""
echo "=========================================="
echo "           TEST SUMMARY"
echo "=========================================="
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"

if [ $FAILED_TESTS -eq 0 ]; then
    echo "All tests passed! 🎉"
    exit 0
else
    echo "Some tests failed! ❌"
    exit 1
fi 