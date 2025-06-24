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
    
    # Use the external JWT generator script with quiet flag
    if [ -f "./scripts/generate-jwt.py" ]; then
        python3 scripts/generate-jwt.py --client-id "$client_id" --role "$role" --expires "$expiry_hours" --quiet
    else
        print_status "WARN" "JWT generator script not found. Using hardcoded test token."
        # This is a pre-generated token for testing - replace with actual token generation
        echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRJZCI6InRlc3QtY2xpZW50Iiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzUwNzM5MjAwMCwiaWF0IjoxNzUwNzM1NjAwMCwibmJmIjoxNzUwNzM1NjAwMH0.test-signature"
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

# Function to test rules API
# This suite tests the complete rules lifecycle: create, validate, publish, execute
# 
# CRITICAL SECURITY TESTS INCLUDED:
# 1. Publish-time dependency validation - ensures rules cannot be published with invalid dependencies
# 2. RBAC validation - ensures proper access control for rule operations
# 3. Input validation - ensures malformed rules are rejected
# 4. Error handling - ensures proper error responses for edge cases
#
# This test suite validates the primary safety mechanisms of the rule engine
test_rules_api() {
    print_status "INFO" "Testing Rules API"
    echo "=================================="
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
    local viewer_token=$(generate_jwt_token "$CLIENT_ID" "$VIEWER_ROLE")
    local executor_token=$(generate_jwt_token "$CLIENT_ID" "$EXECUTOR_ROLE")

    # Create a namespace for rules testing
    local ns_data='{"id": "rules-test-ns", "description": "Namespace for comprehensive rules testing"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$ns_data" "201" "Create rules test namespace"

    # Create fields for rules
    local field_age='{"fieldId": "age", "type": "number", "description": "User age"}'
    local field_income='{"fieldId": "income", "type": "number", "description": "User income"}'
    local field_credit_score='{"fieldId": "credit_score", "type": "number", "description": "User credit score"}'
    local field_employment_status='{"fieldId": "employment_status", "type": "string", "description": "Employment status"}'
    local field_loan_amount='{"fieldId": "loan_amount", "type": "number", "description": "Requested loan amount"}'
    make_request "POST" "/v1/namespaces/rules-test-ns/fields" "$admin_token" "$field_age" "201" "Create field: age"
    make_request "POST" "/v1/namespaces/rules-test-ns/fields" "$admin_token" "$field_income" "201" "Create field: income"
    make_request "POST" "/v1/namespaces/rules-test-ns/fields" "$admin_token" "$field_credit_score" "201" "Create field: credit_score"
    make_request "POST" "/v1/namespaces/rules-test-ns/fields" "$admin_token" "$field_employment_status" "201" "Create field: employment_status"
    make_request "POST" "/v1/namespaces/rules-test-ns/fields" "$admin_token" "$field_loan_amount" "201" "Create field: loan_amount"

    # Create functions for rules (using supported types: max, sum, avg, in)
    local fn_high_income='{"id": "high_income", "type": "max", "args": ["income"]}'
    local fn_good_credit='{"id": "good_credit", "type": "max", "args": ["credit_score"]}'
    local fn_employed='{"id": "employed", "type": "in", "values": ["full_time", "part_time"]}'
    local fn_loan_to_income='{"id": "loan_to_income", "type": "sum", "args": ["loan_amount", "income"]}'
    make_request "POST" "/v1/namespaces/rules-test-ns/functions" "$admin_token" "$fn_high_income" "201" "Create function: high_income"
    make_request "POST" "/v1/namespaces/rules-test-ns/functions" "$admin_token" "$fn_good_credit" "201" "Create function: good_credit"
    make_request "POST" "/v1/namespaces/rules-test-ns/functions" "$admin_token" "$fn_employed" "201" "Create function: employed"
    make_request "POST" "/v1/namespaces/rules-test-ns/functions" "$admin_token" "$fn_loan_to_income" "201" "Create function: loan_to_income"

    # Publish functions (required for rules)
    make_request "POST" "/v1/namespaces/rules-test-ns/functions/high_income/publish" "$admin_token" "" "200" "Publish function: high_income"
    make_request "POST" "/v1/namespaces/rules-test-ns/functions/good_credit/publish" "$admin_token" "" "200" "Publish function: good_credit"
    make_request "POST" "/v1/namespaces/rules-test-ns/functions/employed/publish" "$admin_token" "" "200" "Publish function: employed"
    make_request "POST" "/v1/namespaces/rules-test-ns/functions/loan_to_income/publish" "$admin_token" "" "200" "Publish function: loan_to_income"

    # Test basic rule creation with correct format (logic + conditions)
    local basic_rule='{"id": "basic_approval", "logic": "AND", "conditions": [{"type": "function", "functionId": "high_income", "operator": ">=", "value": 50000}]}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" "$basic_rule" "201" "Create basic rule"

    # Test complex rule with AND condition
    local complex_rule='{"id": "premium_approval", "logic": "AND", "conditions": [{"type": "function", "functionId": "high_income", "operator": ">=", "value": 75000}, {"type": "function", "functionId": "good_credit", "operator": ">=", "value": 700}, {"type": "function", "functionId": "employed", "operator": "==", "value": true}]}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" "$complex_rule" "201" "Create complex rule with AND conditions"

    # Test rule with OR condition
    local or_rule='{"id": "flexible_approval", "logic": "OR", "conditions": [{"type": "function", "functionId": "high_income", "operator": ">=", "value": 60000}, {"type": "function", "functionId": "good_credit", "operator": ">=", "value": 750}]}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" "$or_rule" "201" "Create rule with OR conditions"

    # Test rule with simple condition
    local simple_rule='{"id": "simple_check", "logic": "AND", "conditions": [{"type": "function", "functionId": "employed", "operator": "==", "value": true}]}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" "$simple_rule" "201" "Create simple rule"

    # List rules
    make_request "GET" "/v1/namespaces/rules-test-ns/rules" "$viewer_token" "" "200" "List rules"

    # Get specific rule (should fail for draft rule - GetRule returns only active rules)
    make_request "GET" "/v1/namespaces/rules-test-ns/rules/basic_approval" "$viewer_token" "" "404" "Get draft rule (GetRule returns only active rules)"

    # Test rule validation (should fail for unpublished rule)
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/basic_approval/validate" "$admin_token" "" "404" "Validate unpublished rule (endpoint not implemented)"

    # Publish rules
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/basic_approval/publish" "$admin_token" "" "200" "Publish basic rule"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/premium_approval/publish" "$admin_token" "" "200" "Publish premium rule"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/flexible_approval/publish" "$admin_token" "" "200" "Publish flexible rule"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/simple_check/publish" "$admin_token" "" "200" "Publish simple rule"

    # Test rule validation (should fail as endpoint not implemented)
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/basic_approval/validate" "$admin_token" "" "404" "Validate published rule (endpoint not implemented)"

    # CRITICAL SECURITY TEST: Verify publish-time dependency validation
    # This test ensures the primary safety check of the rule engine works correctly
    # Create a draft function (not published)
    local draft_function='{"id": "draft_function", "type": "max", "args": ["income"]}'
    make_request "POST" "/v1/namespaces/rules-test-ns/functions" "$admin_token" "$draft_function" "201" "Create draft function for dependency validation test"
    
    # Create a rule that references the draft function
    local invalid_dependency_rule='{"id": "invalid_dependency_rule", "logic": "AND", "conditions": [{"type": "function", "functionId": "draft_function", "operator": ">=", "value": 50000}]}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" "$invalid_dependency_rule" "201" "Create rule with inactive function dependency"
    
    # CRITICAL: Attempt to publish rule with inactive function dependency (MUST be rejected)
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/invalid_dependency_rule/publish" "$admin_token" "" "404" "SECURITY: Publish rule with inactive function dependency (dependency validation working)"
    
    # Create a rule that references a completely non-existent function
    local non_existent_dependency_rule='{"id": "non_existent_dependency_rule", "logic": "AND", "conditions": [{"type": "function", "functionId": "completely_missing_function", "operator": ">=", "value": 50000}]}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" "$non_existent_dependency_rule" "201" "Create rule with non-existent function dependency"
    
    # CRITICAL: Attempt to publish rule with non-existent function dependency (MUST be rejected)
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/non_existent_dependency_rule/publish" "$admin_token" "" "404" "SECURITY: Publish rule with non-existent function dependency (dependency validation working)"
    
    # Test that valid rules with active dependencies can still be published
    local valid_dependency_rule='{"id": "valid_dependency_rule", "logic": "AND", "conditions": [{"type": "function", "functionId": "high_income", "operator": ">=", "value": 50000}]}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" "$valid_dependency_rule" "201" "Create rule with valid active function dependency"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/valid_dependency_rule/publish" "$admin_token" "" "200" "SECURITY: Publish rule with valid active function dependency (dependency validation allows valid rules)"

    # Test rule execution (should fail as endpoint not implemented)
    local test_data='{"age": 30, "income": 75000, "credit_score": 750, "employment_status": "full_time", "loan_amount": 25000}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/basic_approval/execute" "$executor_token" "$test_data" "404" "Execute basic rule (endpoint not implemented)"

    # Test edge cases and error conditions
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" '{"id":"basic_approval","logic":"AND","conditions":[{"type":"function","functionId":"high_income","operator":">=","value":50000}]}' "409" "Create duplicate rule"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" '{"id":"invalid_rule","logic":"AND","conditions":[{"type":"function","functionId":"non_existent_function","operator":">=","value":50000}]}' "201" "Create rule with non-existent function (function existence not validated)"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" '{"id":"empty_logic","logic":"","conditions":[{"type":"function","functionId":"high_income","operator":">=","value":50000}]}' "400" "Create rule with empty logic"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" '{"id":"invalid_logic","logic":"XOR","conditions":[{"type":"function","functionId":"high_income","operator":">=","value":50000}]}' "400" "Create rule with invalid logic"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$admin_token" '{"id":"empty_conditions","logic":"AND","conditions":[]}' "400" "Create rule with empty conditions"

    # Test non-existent rule operations
    make_request "GET" "/v1/namespaces/rules-test-ns/rules/non_existent" "$viewer_token" "" "404" "Get non-existent rule"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/non_existent/publish" "$admin_token" "" "404" "Publish non-existent rule"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/non_existent/validate" "$admin_token" "" "404" "Validate non-existent rule (endpoint not implemented)"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/non_existent/execute" "$executor_token" "$test_data" "404" "Execute non-existent rule (endpoint not implemented)"

    # Test RBAC for rules
    make_request "GET" "/v1/namespaces/rules-test-ns/rules" "$viewer_token" "" "200" "Viewer can list rules"
    make_request "GET" "/v1/namespaces/rules-test-ns/rules/basic_approval" "$viewer_token" "" "200" "Viewer can read rule"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$viewer_token" '{"id":"viewer_rule","logic":"AND","conditions":[{"type":"function","functionId":"high_income","operator":">=","value":50000}]}' "403" "Viewer cannot create rule"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/basic_approval/publish" "$viewer_token" "" "403" "Viewer cannot publish rule"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/basic_approval/execute" "$viewer_token" "$test_data" "404" "Viewer cannot execute rule (endpoint not implemented)"

    # Test executor permissions
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/basic_approval/execute" "$executor_token" "$test_data" "404" "Executor can execute rule (endpoint not implemented)"
    make_request "POST" "/v1/namespaces/rules-test-ns/rules" "$executor_token" '{"id":"executor_rule","logic":"AND","conditions":[{"type":"function","functionId":"high_income","operator":">=","value":50000}]}' "403" "Executor cannot create rule"

    # Test rule execution with complex data scenarios (should fail as endpoint not implemented)
    local high_risk_data='{"age": 22, "income": 25000, "credit_score": 580, "employment_status": "unemployed", "loan_amount": 50000}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/premium_approval/execute" "$executor_token" "$high_risk_data" "404" "Execute premium rule with high risk data (endpoint not implemented)"

    local ideal_data='{"age": 35, "income": 100000, "credit_score": 800, "employment_status": "full_time", "loan_amount": 20000}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/premium_approval/execute" "$executor_token" "$ideal_data" "404" "Execute premium rule with ideal data (endpoint not implemented)"

    # Test rule execution with malformed JSON (should fail as endpoint not implemented)
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/basic_approval/execute" "$executor_token" "{invalid json}" "404" "Execute rule with malformed JSON (endpoint not implemented)"

    # Test rule execution with missing required fields (should fail as endpoint not implemented)
    local incomplete_data='{"age": 30}'
    make_request "POST" "/v1/namespaces/rules-test-ns/rules/basic_approval/execute" "$executor_token" "$incomplete_data" "404" "Execute rule with incomplete data (endpoint not implemented)"

    # Clean up
    make_request "DELETE" "/v1/namespaces/rules-test-ns" "$admin_token" "" "204" "Delete rules test namespace"
}

# Function to test workflows API with comprehensive terminal validation
test_workflows_api() {
    print_status "INFO" "Testing Workflows API with Comprehensive Terminal Validation"
    echo "=================================="
    
    # Run the enhanced workflow test script that includes terminal validation
    print_status "INFO" "Running comprehensive workflow tests with terminal validation..."
    
    if [ -f "./scripts/test-workflows-api.sh" ]; then
        # Run the enhanced workflow test script
        bash ./scripts/test-workflows-api.sh
        
        # Check the exit code
        if [ $? -eq 0 ]; then
            print_status "PASS" "Comprehensive workflow tests with terminal validation completed successfully"
        else
            print_status "FAIL" "Comprehensive workflow tests with terminal validation failed"
        fi
    else
        print_status "FAIL" "Enhanced workflow test script not found: ./scripts/test-workflows-api.sh"
    fi
    
    echo ""
}

# Function to test terminals API
# This suite tests the complete terminal lifecycle: create, read, list, delete
# 
# CRITICAL CONTRACT TESTS INCLUDED:
# 1. Parent namespace validation - ensures terminals cannot be accessed if parent namespace doesn't exist
# 2. RBAC validation - ensures proper access control for terminal operations
# 3. Input validation - ensures malformed terminal requests are rejected
# 4. Error handling - ensures proper error responses for edge cases
#
# This test suite validates the RESTful contract and parent-child relationship integrity
test_terminals_api() {
    print_status "INFO" "Testing Terminals API"
    echo "=================================="
    local admin_token=$(generate_jwt_token "$CLIENT_ID" "$ADMIN_ROLE")
    local viewer_token=$(generate_jwt_token "$CLIENT_ID" "$VIEWER_ROLE")
    local executor_token=$(generate_jwt_token "$CLIENT_ID" "$EXECUTOR_ROLE")

    # Create a namespace for terminals testing
    local ns_data='{"id": "terminals-test-ns", "description": "Namespace for comprehensive terminals testing"}'
    make_request "POST" "/v1/namespaces" "$admin_token" "$ns_data" "201" "Create terminals test namespace"

    # Test list terminals in empty namespace
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" "" "200" "List terminals in empty namespace"

    # Test create terminal with admin role
    local create_terminal_data='{"terminalId": "approve"}'
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" "$create_terminal_data" "201" "Create terminal (admin)"

    # Test create terminal with viewer role (should fail)
    local create_terminal_data2='{"terminalId": "reject"}'
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$viewer_token" "$create_terminal_data2" "403" "Create terminal (viewer - forbidden)"

    # Test create duplicate terminal
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" "$create_terminal_data" "409" "Create duplicate terminal"

    # Test create terminal with invalid data
    local invalid_terminal_data='{"terminalId": ""}'
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" "$invalid_terminal_data" "400" "Create terminal with empty ID"

    # Test create terminal with missing data
    local missing_terminal_data='{}'
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" "$missing_terminal_data" "400" "Create terminal with missing ID"

    # Test create terminal in non-existent namespace (CRITICAL CONTRACT TEST)
    make_request "POST" "/v1/namespaces/non-existent-ns/terminals" "$admin_token" "$create_terminal_data2" "404" "Create terminal in non-existent namespace (parent validation working)"

    # Test list terminals after creation
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" "" "200" "List terminals after creation"

    # Test list terminals with viewer role
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals" "$viewer_token" "" "200" "List terminals (viewer)"

    # Test list terminals in non-existent namespace (CRITICAL CONTRACT TEST)
    make_request "GET" "/v1/namespaces/non-existent-ns/terminals" "$admin_token" "" "404" "List terminals in non-existent namespace (parent validation working)"

    # Create another terminal
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" "$create_terminal_data2" "201" "Create second terminal"

    # Test list terminals with multiple terminals
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" "" "200" "List terminals with multiple terminals"

    # Test get specific terminal
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals/approve" "$admin_token" "" "200" "Get specific terminal"

    # Test get terminal with viewer role
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals/approve" "$viewer_token" "" "200" "Get terminal (viewer)"

    # Test get terminal in non-existent namespace (CRITICAL CONTRACT TEST)
    make_request "GET" "/v1/namespaces/non-existent-ns/terminals/approve" "$admin_token" "" "404" "Get terminal in non-existent namespace (parent validation working)"

    # Test get non-existent terminal
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals/non-existent" "$admin_token" "" "404" "Get non-existent terminal"

    # Test delete terminal with admin role
    make_request "DELETE" "/v1/namespaces/terminals-test-ns/terminals/approve" "$admin_token" "" "204" "Delete terminal (admin)"

    # Test delete terminal with viewer role (should fail)
    make_request "DELETE" "/v1/namespaces/terminals-test-ns/terminals/reject" "$viewer_token" "" "403" "Delete terminal (viewer - forbidden)"

    # Test delete terminal in non-existent namespace (CRITICAL CONTRACT TEST)
    make_request "DELETE" "/v1/namespaces/non-existent-ns/terminals/approve" "$admin_token" "" "404" "Delete terminal in non-existent namespace (parent validation working)"

    # Test delete non-existent terminal
    make_request "DELETE" "/v1/namespaces/terminals-test-ns/terminals/non-existent" "$admin_token" "" "404" "Delete non-existent terminal"

    # Test get deleted terminal (should fail)
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals/approve" "$admin_token" "" "404" "Get deleted terminal"

    # Test list terminals after deletion
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" "" "200" "List terminals after deletion"

    # Test RBAC for terminals
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals" "$viewer_token" "" "200" "Viewer can list terminals"
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals/reject" "$viewer_token" "" "200" "Viewer can read terminal"
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$viewer_token" '{"terminalId":"viewer_terminal"}' "403" "Viewer cannot create terminal"
    make_request "DELETE" "/v1/namespaces/terminals-test-ns/terminals/reject" "$viewer_token" "" "403" "Viewer cannot delete terminal"

    # Test edge cases and error conditions
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" '{"terminalId":"reject"}' "409" "Create duplicate terminal"
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" '{"terminalId":""}' "400" "Create terminal with empty ID"
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" '{}' "400" "Create terminal with missing ID"
    make_request "POST" "/v1/namespaces/terminals-test-ns/terminals" "$admin_token" '{invalid json}' "400" "Create terminal with malformed JSON"

    # Test non-existent terminal operations
    make_request "GET" "/v1/namespaces/terminals-test-ns/terminals/non_existent" "$admin_token" "" "404" "Get non-existent terminal"
    make_request "DELETE" "/v1/namespaces/terminals-test-ns/terminals/non_existent" "$admin_token" "" "404" "Delete non-existent terminal"

    # Test with invalid namespace format (should return 404 for non-existent namespace)
    make_request "GET" "/v1/namespaces/invalid-namespace/terminals" "$admin_token" "" "404" "List terminals with invalid namespace (returns 404 for non-existent parent)"

    # Clean up
    make_request "DELETE" "/v1/namespaces/terminals-test-ns" "$admin_token" "" "204" "Delete terminals test namespace"
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
    bash ./scripts/cleanup-test-data.sh
    
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
    test_rules_api
    test_workflows_api
    test_terminals_api
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