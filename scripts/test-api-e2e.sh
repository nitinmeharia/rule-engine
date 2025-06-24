#!/bin/bash

# Generic Rule Engine - End-to-End API Testing Script
# This script tests all available APIs with proper authentication and error handling

set -e  # Exit on any error

# Configuration
BASE_URL="http://localhost:8080"
JWT_SECRET="dev-secret-key-change-in-production"
CLIENT_ID="test-client"
ADMIN_ROLE="admin"
VIEWER_ROLE="viewer"
EXECUTOR_ROLE="executor"

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
        "WARN")
            echo -e "${YELLOW}⚠ WARN${NC}: $message"
            ;;
    esac
}

# Function to generate JWT token
generate_jwt_token() {
    local client_id=$1
    local role=$2
    local expiry_hours=${3:-24}
    
    # Create JWT token using conda Python that has PyJWT
    if [ -x "$PYTHON" ]; then
        $PYTHON -c "
import jwt
import datetime

secret = '$JWT_SECRET'
client_id = '$client_id'
role = '$role'
expiry_hours = $expiry_hours

# Use timezone-aware datetime to avoid deprecation warnings
now = datetime.datetime.now(datetime.timezone.utc)

claims = {
    'clientId': client_id,
    'role': role,
    'exp': now + datetime.timedelta(hours=expiry_hours),
    'iat': now,
    'nbf': now
}

token = jwt.encode(claims, secret, algorithm='HS256')
print(token)
"
    else
        # Fallback: use the JWT generator binary if available
        if [ -f "./bin/jwt-generator" ]; then
            ./bin/jwt-generator -secret="$JWT_SECRET" -client-id="$client_id" -role="$role" -expiry="${expiry_hours}h" | head -n 1
        else
            print_status "WARN" "Conda Python not available and JWT generator not found. Using hardcoded test token."
            # This is a pre-generated token for testing - replace with actual token generation
            echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRJZCI6InRlc3QtY2xpZW50Iiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzUwNzM5MjAwMCwiaWF0IjoxNzUwNzM1NjAwMCwibmJmIjoxNzUwNzM1NjAwMH0.test-signature"
        fi
    fi
}

# Function to make HTTP requests and validate responses
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
        print_status "PASS" "$test_name (Status: $status_code)"
        if [ "$status_code" != "204" ]; then
            echo "  Response: $response_body" | head -c 200
            [ ${#response_body} -gt 200 ] && echo "..."
        fi
    else
        print_status "FAIL" "$test_name (Expected: $expected_status, Got: $status_code)"
        echo "  Response: $response_body"
    fi
    
    echo ""
}

# Function to wait for server to be ready
wait_for_server() {
    print_status "INFO" "Waiting for server to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
            print_status "PASS" "Server is ready"
            return 0
        fi
        
        echo -n "."
        sleep 1
        ((attempt++))
    done
    
    print_status "FAIL" "Server failed to start within $max_attempts seconds"
    return 1
}

# Function to test health endpoint
test_health_endpoint() {
    print_status "INFO" "Testing Health Endpoint"
    echo "=================================="
    
    # Test health check without authentication
    make_request "GET" "/health" "" "" "200" "Health check (no auth)"
}

# Function to test authentication
test_authentication() {
    print_status "INFO" "Testing Authentication"
    echo "=================================="
    
    # Test missing authorization header
    make_request "GET" "/v1/namespaces" "" "" "401" "Missing Authorization header"
    
    # Test invalid token format
    make_request "GET" "/v1/namespaces" "invalid-token" "" "401" "Invalid token format"
    
    # Test expired token (if we had one)
    # make_request "GET" "/v1/namespaces" "expired-token" "" "401" "Expired token"
}

# Function to test namespaces API
test_namespaces_api() {
    print_status "INFO" "Testing Namespaces API"
    echo "=================================="
    
    # Generate admin token
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
    local viewer_token=$(generate_jwt_token "$CLIENT_ID" "$VIEWER_ROLE")
    
    # Test list namespaces (should be empty initially)
    make_request "GET" "/v1/namespaces" "$admin_token" "" "200" "List namespaces (admin)"
    
    # Test create namespace with admin role
    local create_data='{"id": "test-namespace-1", "description": "Test namespace for API testing"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$create_data" "201" "Create namespace (admin)"
    
    # Test create namespace with viewer role (should fail)
    make_request "POST" "/v1/namespaces" "$viewer_token" "$create_data" "403" "Create namespace (viewer - forbidden)"
    
    # Test create duplicate namespace
    make_request "POST" "/v1/namespaces" "$admin_token" "$create_data" "409" "Create duplicate namespace"
    
    # Test create namespace with invalid data
    local invalid_data='{"id": "", "description": "Invalid namespace"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$invalid_data" "400" "Create namespace with invalid ID"
    
    # Test get specific namespace
    make_request "GET" "/v1/namespaces/test-namespace-1" "$admin_token" "" "200" "Get specific namespace"
    
    # Test get non-existent namespace
    make_request "GET" "/v1/namespaces/non-existent" "$admin_token" "" "404" "Get non-existent namespace"
    
    # Test list namespaces again (should have one namespace)
    make_request "GET" "/v1/namespaces" "$admin_token" "" "200" "List namespaces after creation"
    
    # Test delete namespace with admin role
    make_request "DELETE" "/v1/namespaces/test-namespace-1" "$admin_token" "" "204" "Delete namespace (admin)"
    
    # Test delete namespace with viewer role (should fail)
    make_request "DELETE" "/v1/namespaces/test-namespace-1" "$viewer_token" "" "403" "Delete namespace (viewer - forbidden)"
    
    # Test delete non-existent namespace
    make_request "DELETE" "/v1/namespaces/non-existent" "$admin_token" "" "404" "Delete non-existent namespace"
}

# Function to test fields API
test_fields_api() {
    print_status "INFO" "Testing Fields API"
    echo "=================================="
    
    # Generate tokens
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
    local viewer_token=$(generate_jwt_token "$CLIENT_ID" "$VIEWER_ROLE")
    
    # Create a namespace first
    local create_ns_data='{"id": "fields-test-ns", "description": "Namespace for fields testing"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$create_ns_data" "201" "Create namespace for fields testing"
    
    # Test list fields in empty namespace
    make_request "GET" "/v1/namespaces/fields-test-ns/fields" "$admin_token" "" "200" "List fields in empty namespace"
    
    # Test create field with admin role
    local create_field_data='{"fieldId": "user_age", "type": "number", "description": "Age of the user"}'
    make_request "POST" "/v1/namespaces/fields-test-ns/fields" "$admin_token" "$create_field_data" "201" "Create field (admin)"
    
    # Test create field with viewer role (should fail)
    make_request "POST" "/v1/namespaces/fields-test-ns/fields" "$viewer_token" "$create_field_data" "403" "Create field (viewer - forbidden)"
    
    # Test create duplicate field
    make_request "POST" "/v1/namespaces/fields-test-ns/fields" "$admin_token" "$create_field_data" "409" "Create duplicate field"
    
    # Test create field with invalid data
    local invalid_field_data='{"fieldId": "", "type": "invalid_type", "description": "Invalid field"}'
    make_request "POST" "/v1/namespaces/fields-test-ns/fields" "$admin_token" "$invalid_field_data" "400" "Create field with invalid data"
    
    # Test create field in non-existent namespace
    make_request "POST" "/v1/namespaces/non-existent/fields" "$admin_token" "$create_field_data" "404" "Create field in non-existent namespace"
    
    # Test list fields after creation
    make_request "GET" "/v1/namespaces/fields-test-ns/fields" "$admin_token" "" "200" "List fields after creation"
    
    # Test list fields with viewer role
    make_request "GET" "/v1/namespaces/fields-test-ns/fields" "$viewer_token" "" "200" "List fields (viewer)"
    
    # Create another field
    local create_field_data2='{"fieldId": "user_name", "type": "string", "description": "Name of the user"}'
    make_request "POST" "/v1/namespaces/fields-test-ns/fields" "$admin_token" "$create_field_data2" "201" "Create second field"
    
    # Test list fields with multiple fields
    make_request "GET" "/v1/namespaces/fields-test-ns/fields" "$admin_token" "" "200" "List fields with multiple fields"
    
    # Clean up - delete the namespace (this will also delete all fields)
    make_request "DELETE" "/v1/namespaces/fields-test-ns" "$admin_token" "" "204" "Delete namespace with fields"
}

# Function to clean up test data (namespaces, fields, functions, rules)
cleanup_test_data() {
    print_status "INFO" "Cleaning up test data"
    psql -U postgres -d rule_engine_dev -c "DELETE FROM rules WHERE namespace IN ('test-e2e', 'test-functions', 'fields-test-ns', 'rbac-test-ns'); DELETE FROM functions WHERE namespace IN ('test-e2e', 'test-functions', 'fields-test-ns', 'rbac-test-ns'); DELETE FROM fields WHERE namespace IN ('test-e2e', 'test-functions', 'fields-test-ns', 'rbac-test-ns'); DELETE FROM namespaces WHERE id IN ('test-e2e', 'test-functions', 'fields-test-ns', 'rbac-test-ns');" > /dev/null 2>&1 || true
}

# Function to test functions API (Rules API not implemented yet)
# This suite creates fields and functions that use those fields
# It tests the complete function lifecycle: create, update, publish

test_functions_api() {
    print_status "INFO" "Testing Functions API"
    echo "=================================="
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
    local viewer_token=$(generate_jwt_token "$CLIENT_ID" "$VIEWER_ROLE")

    # Create a namespace for this suite
    local ns_data='{"id": "test-e2e", "description": "Namespace for E2E logical tests"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$ns_data" "201" "Create E2E test namespace"

    # Create fields
    local field_age='{"fieldId": "age", "type": "number", "description": "User age"}'
    local field_income='{"fieldId": "income", "type": "number", "description": "User income"}'
    local field_occupation='{"fieldId": "occupation", "type": "string", "description": "User occupation"}'
    make_request "POST" "/v1/namespaces/test-e2e/fields" "$admin_token" "$field_age" "201" "Create field: age"
    make_request "POST" "/v1/namespaces/test-e2e/fields" "$admin_token" "$field_income" "201" "Create field: income"
    make_request "POST" "/v1/namespaces/test-e2e/fields" "$admin_token" "$field_occupation" "201" "Create field: occupation"

    # Create functions using supported types
    local fn_max_income='{"id": "max_income", "type": "max", "args": ["age", "income"]}'
    local fn_total_income='{"id": "total_income", "type": "sum", "args": ["income"]}'
    local fn_avg_age='{"id": "avg_age", "type": "avg", "args": ["age"]}'
    local fn_valid_occupation='{"id": "valid_occupation", "type": "in", "values": ["salaried", "self_employed", "business"]}'
    make_request "POST" "/v1/namespaces/test-e2e/functions" "$admin_token" "$fn_max_income" "201" "Create function: max_income"
    make_request "POST" "/v1/namespaces/test-e2e/functions" "$admin_token" "$fn_total_income" "201" "Create function: total_income"
    make_request "POST" "/v1/namespaces/test-e2e/functions" "$admin_token" "$fn_avg_age" "201" "Create function: avg_age"
    make_request "POST" "/v1/namespaces/test-e2e/functions" "$admin_token" "$fn_valid_occupation" "201" "Create function: valid_occupation"

    # List functions
    make_request "GET" "/v1/namespaces/test-e2e/functions" "$viewer_token" "" "200" "List functions"

    # Get specific function
    make_request "GET" "/v1/namespaces/test-e2e/functions/max_income" "$viewer_token" "" "200" "Get max_income function"

    # Update function draft
    local update_data='{"type": "max", "args": ["age", "income", "bonus"]}'
    make_request "PUT" "/v1/namespaces/test-e2e/functions/max_income/versions/draft" "$admin_token" "$update_data" "200" "Update function draft"

    # Publish function
    make_request "POST" "/v1/namespaces/test-e2e/functions/max_income/publish" "$admin_token" "" "200" "Publish function: max_income"
    make_request "POST" "/v1/namespaces/test-e2e/functions/total_income/publish" "$admin_token" "" "200" "Publish function: total_income"
    make_request "POST" "/v1/namespaces/test-e2e/functions/avg_age/publish" "$admin_token" "" "200" "Publish function: avg_age"
    make_request "POST" "/v1/namespaces/test-e2e/functions/valid_occupation/publish" "$admin_token" "" "200" "Publish function: valid_occupation"

    # Get published function
    make_request "GET" "/v1/namespaces/test-e2e/functions/max_income" "$viewer_token" "" "200" "Get published function"

    # Test edge cases
    make_request "POST" "/v1/namespaces/test-e2e/functions" "$admin_token" '{"id":"max_income","type":"max","args":["salary"]}' "409" "Create duplicate function"
    make_request "POST" "/v1/namespaces/test-e2e/functions" "$admin_token" '{"id":"invalid_func","type":"invalid","args":["salary"]}' "400" "Create function with invalid type"
    make_request "POST" "/v1/namespaces/test-e2e/functions" "$admin_token" '{"id":"empty_max","type":"max","args":[]}' "400" "Create numeric function without args"
    make_request "POST" "/v1/namespaces/test-e2e/functions" "$admin_token" '{"id":"empty_in","type":"in","values":[]}' "400" "Create in function without values"
    make_request "GET" "/v1/namespaces/test-e2e/functions/non_existent" "$viewer_token" "" "404" "Get non-existent function"
    make_request "PUT" "/v1/namespaces/test-e2e/functions/non_existent/versions/draft" "$admin_token" '{"type":"max","args":["salary"]}' "404" "Update non-existent function"
    make_request "POST" "/v1/namespaces/test-e2e/functions/non_existent/publish" "$admin_token" "" "404" "Publish non-existent function"

    # Test RBAC
    make_request "GET" "/v1/namespaces/test-e2e/functions" "$viewer_token" "" "200" "Viewer can list functions"
    make_request "GET" "/v1/namespaces/test-e2e/functions/max_income" "$viewer_token" "" "200" "Viewer can read function"
    make_request "POST" "/v1/namespaces/test-e2e/functions" "$viewer_token" '{"id":"viewer_func","type":"max","args":["salary"]}' "403" "Viewer cannot create function"

    # Clean up
    make_request "DELETE" "/v1/namespaces/test-e2e" "$admin_token" "" "204" "Delete E2E test namespace"
}

# Function to test error handling
test_error_handling() {
    print_status "INFO" "Testing Error Handling"
    echo "=================================="
    
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
    
    # Test malformed JSON
    make_request "POST" "/v1/namespaces" "$admin_token" "{invalid json}" "400" "Malformed JSON"
    
    # Test missing required fields
    local missing_fields='{"id": "test-ns"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$missing_fields" "400" "Missing required fields"
    
    # Test invalid field types
    local invalid_types='{"id": "test-ns", "description": 123}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$invalid_types" "400" "Invalid field types"
}

# Function to test role-based access control
test_rbac() {
    print_status "INFO" "Testing Role-Based Access Control"
    echo "================================================"
    
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
    local viewer_token=$(generate_jwt_token "$CLIENT_ID" "$VIEWER_ROLE")
    local executor_token=$(generate_jwt_token "$CLIENT_ID" "$EXECUTOR_ROLE")
    
    # Create a test namespace
    local create_ns_data='{"id": "rbac-test-ns", "description": "Namespace for RBAC testing"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$create_ns_data" "201" "Create namespace for RBAC testing"
    
    # Test admin can do everything
    make_request "GET" "/v1/namespaces" "$admin_token" "" "200" "Admin can list namespaces"
    make_request "GET" "/v1/namespaces/rbac-test-ns" "$admin_token" "" "200" "Admin can get namespace"
    make_request "GET" "/v1/namespaces/rbac-test-ns/fields" "$admin_token" "" "200" "Admin can list fields"
    
    # Test viewer can read but not write
    make_request "GET" "/v1/namespaces" "$viewer_token" "" "200" "Viewer can list namespaces"
    make_request "GET" "/v1/namespaces/rbac-test-ns" "$viewer_token" "" "200" "Viewer can get namespace"
    make_request "GET" "/v1/namespaces/rbac-test-ns/fields" "$viewer_token" "" "200" "Viewer can list fields"
    
    local create_field_data='{"fieldId": "test_field", "type": "string", "description": "Test field"}'
    make_request "POST" "/v1/namespaces/rbac-test-ns/fields" "$viewer_token" "$create_field_data" "403" "Viewer cannot create field"
    make_request "DELETE" "/v1/namespaces/rbac-test-ns" "$viewer_token" "" "403" "Viewer cannot delete namespace"
    
    # Test executor can read but not write
    make_request "GET" "/v1/namespaces" "$executor_token" "" "200" "Executor can list namespaces"
    make_request "GET" "/v1/namespaces/rbac-test-ns" "$executor_token" "" "200" "Executor can get namespace"
    make_request "GET" "/v1/namespaces/rbac-test-ns/fields" "$executor_token" "" "200" "Executor can list fields"
    make_request "POST" "/v1/namespaces/rbac-test-ns/fields" "$executor_token" "$create_field_data" "403" "Executor cannot create field"
    
    # Clean up
    make_request "DELETE" "/v1/namespaces/rbac-test-ns" "$admin_token" "" "204" "Clean up RBAC test namespace"
}

# Function to test edge cases
test_edge_cases() {
    print_status "INFO" "Testing Edge Cases"
    echo "================================"
    
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
    
    # Test very long namespace ID
    local long_id=$(printf 'a%.0s' {1..100})
    local long_id_data="{\"id\": \"$long_id\", \"description\": \"Test with long ID\"}"
    make_request "POST" "/v1/namespaces" "$admin_token" "$long_id_data" "400" "Very long namespace ID"
    
    # Test very long description
    local long_desc=$(printf 'a%.0s' {1..1000})
    local long_desc_data="{\"id\": \"test-long-desc\", \"description\": \"$long_desc\"}"
    make_request "POST" "/v1/namespaces" "$admin_token" "$long_desc_data" "400" "Very long description"
    
    # Test special characters in namespace ID
    local special_chars_data='{"id": "test@namespace#123", "description": "Test with special chars"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$special_chars_data" "400" "Special characters in namespace ID"
    
    # Test empty description
    local empty_desc_data='{"id": "test-empty-desc", "description": ""}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$empty_desc_data" "400" "Empty description"
}

# Function to run performance tests
test_performance() {
    print_status "INFO" "Testing Performance"
    echo "================================="
    
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
    
    # Test health endpoint performance
    local start_time=$(date +%s%N)
    make_request "GET" "/health" "" "" "200" "Health check performance"
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))
    
    if [ $duration -lt 1000 ]; then
        print_status "PASS" "Health endpoint response time: ${duration}ms"
    else
        print_status "FAIL" "Health endpoint response time too slow: ${duration}ms"
    fi
    
    # Test multiple concurrent requests (simple test)
    print_status "INFO" "Testing concurrent requests..."
    for i in {1..5}; do
        make_request "GET" "/health" "" "" "200" "Concurrent health check $i"
    done
}

# Function to print test summary
print_summary() {
    echo ""
    echo "=========================================="
    echo "           TEST SUMMARY"
    echo "=========================================="
    echo "Total Tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}All tests passed! 🎉${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed! ❌${NC}"
        exit 1
    fi
}

# Main test execution
main() {
    echo "=========================================="
    echo "    Generic Rule Engine - E2E Tests"
    echo "=========================================="
    echo "Base URL: $BASE_URL"
    echo "Client ID: $CLIENT_ID"
    echo "JWT Secret: $JWT_SECRET"
    echo ""

    # Clean up any existing test data
    cleanup_test_data

    # Wait for server to be ready
    if ! wait_for_server; then
        print_status "FAIL" "Cannot proceed with tests - server not available"
        exit 1
    fi

    # Run all test suites
    test_health_endpoint
    test_authentication
    test_namespaces_api
    test_fields_api
    test_functions_api
    test_error_handling
    test_rbac
    test_edge_cases
    test_performance

    # Print summary
    print_summary
}

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is required but not installed${NC}"
    exit 1
fi

# Check if server is running
if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    echo -e "${YELLOW}Warning: Server does not appear to be running at $BASE_URL${NC}"
    echo "Please start the server first:"
    echo "  go run ./cmd/api"
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Run the main function
main "$@" 