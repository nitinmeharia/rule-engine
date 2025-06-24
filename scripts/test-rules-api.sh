#!/bin/bash

# Test Rules API functionality
set -e

# Configuration
API_BASE_URL="http://localhost:8080/v1"
JWT_TOKEN=""
NAMESPACE_ID="test-rules-ns-$(date +%s)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to generate JWT token
generate_jwt() {
    print_status "Generating JWT token..."
    JWT_TOKEN=$(go run ./cmd/jwt-generator -client-id test-client -role admin | tail -n1)
    if [ -z "$JWT_TOKEN" ]; then
        print_error "Failed to generate JWT token"
        exit 1
    fi
    print_status "JWT token generated successfully"
}

# Function to make API requests
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -z "$JWT_TOKEN" ]; then
        print_error "JWT token not set"
        exit 1
    fi
    
    local response
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
    
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    echo "$status_code"
    echo "$body"
}

# Function to test namespace creation
test_namespace_creation() {
    print_status "Testing namespace creation..."
    
    local namespace_data='{"id":"'$NAMESPACE_ID'","description":"Test namespace for rules"}'
    local result=$(make_request "POST" "/namespaces" "$namespace_data")
    local status_code=$(echo "$result" | head -n1)
    local body=$(echo "$result" | tail -n +2)
    
    if [ "$status_code" -eq 201 ]; then
        print_status "Namespace created successfully"
    else
        print_error "Failed to create namespace: $status_code - $body"
        return 1
    fi
}

# Function to test field creation
test_field_creation() {
    print_status "Testing field creation..."
    
    local field_data='{"fieldId":"salary","type":"number","description":"Annual salary"}'
    local result=$(make_request "POST" "/namespaces/$NAMESPACE_ID/fields" "$field_data")
    local status_code=$(echo "$result" | head -n1)
    local body=$(echo "$result" | tail -n +2)
    
    if [ "$status_code" -eq 201 ]; then
        print_status "Field created successfully"
    else
        print_error "Failed to create field: $status_code - $body"
        return 1
    fi
}

# Function to test function creation
test_function_creation() {
    print_status "Testing function creation..."
    
    local function_data='{"id":"max_income","type":"max","args":["salary","bonus"]}'
    local result=$(make_request "POST" "/namespaces/$NAMESPACE_ID/functions" "$function_data")
    local status_code=$(echo "$result" | head -n1)
    local body=$(echo "$result" | tail -n +2)
    
    if [ "$status_code" -eq 201 ]; then
        print_status "Function created successfully"
    else
        print_error "Failed to create function: $status_code - $body"
        return 1
    fi
    
    # Publish the function so it can be referenced by rules
    print_status "Publishing function..."
    local publish_result=$(make_request "POST" "/namespaces/$NAMESPACE_ID/functions/max_income/publish")
    local publish_status=$(echo "$publish_result" | head -n1)
    local publish_body=$(echo "$publish_result" | tail -n +2)
    
    if [ "$publish_status" -eq 200 ]; then
        print_status "Function published successfully"
    else
        print_error "Failed to publish function: $publish_status - $publish_body"
        return 1
    fi
}

# Function to test rule creation
test_rule_creation() {
    print_status "Testing rule creation..."
    
    local rule_data='{
        "id":"income_check",
        "logic":"AND",
        "conditions":[
            {
                "type":"field",
                "fieldId":"salary",
                "operator":">=",
                "value":50000
            },
            {
                "type":"function",
                "functionId":"max_income",
                "operator":">=",
                "value":60000
            }
        ]
    }'
    
    local result=$(make_request "POST" "/namespaces/$NAMESPACE_ID/rules" "$rule_data")
    local status_code=$(echo "$result" | head -n1)
    local body=$(echo "$result" | tail -n +2)
    
    if [ "$status_code" -eq 201 ]; then
        print_status "Rule created successfully"
        echo "$body"
    else
        print_error "Failed to create rule: $status_code - $body"
        return 1
    fi
}

# Function to test rule retrieval
test_rule_retrieval() {
    print_status "Testing rule retrieval..."
    
    local result=$(make_request "GET" "/namespaces/$NAMESPACE_ID/rules/income_check")
    local status_code=$(echo "$result" | head -n1)
    local body=$(echo "$result" | tail -n +2)
    
    if [ "$status_code" -eq 200 ]; then
        print_status "Rule retrieved successfully"
        echo "$body"
    else
        print_error "Failed to retrieve rule: $status_code - $body"
        return 1
    fi
}

# Function to test rule listing
test_rule_listing() {
    print_status "Testing rule listing..."
    
    local result=$(make_request "GET" "/namespaces/$NAMESPACE_ID/rules")
    local status_code=$(echo "$result" | head -n1)
    local body=$(echo "$result" | tail -n +2)
    
    if [ "$status_code" -eq 200 ]; then
        print_status "Rules listed successfully"
        echo "$body"
    else
        print_error "Failed to list rules: $status_code - $body"
        return 1
    fi
}

# Function to test rule publishing
test_rule_publishing() {
    print_status "Testing rule publishing..."
    
    local result=$(make_request "POST" "/namespaces/$NAMESPACE_ID/rules/income_check/publish")
    local status_code=$(echo "$result" | head -n1)
    local body=$(echo "$result" | tail -n +2)
    
    if [ "$status_code" -eq 200 ]; then
        print_status "Rule published successfully"
        echo "$body"
    else
        print_error "Failed to publish rule: $status_code - $body"
        return 1
    fi
}

# Function to test rule history
test_rule_history() {
    print_status "Testing rule history..."
    
    local result=$(make_request "GET" "/namespaces/$NAMESPACE_ID/rules/income_check/history")
    local status_code=$(echo "$result" | head -n1)
    local body=$(echo "$result" | tail -n +2)
    
    if [ "$status_code" -eq 200 ]; then
        print_status "Rule history retrieved successfully"
        echo "$body"
    else
        print_error "Failed to retrieve rule history: $status_code - $body"
        return 1
    fi
}

# Main test execution
main() {
    print_status "Starting Rules API tests..."
    
    # Check if server is running
    if ! curl -s http://localhost:8080/health > /dev/null; then
        print_error "Server is not running. Please start the server first."
        print_warning "Run: go run ./cmd/api"
        exit 1
    fi
    
    # Generate JWT token
    generate_jwt
    
    # Run tests
    test_namespace_creation
    test_field_creation
    test_function_creation
    test_rule_creation
    test_rule_publishing
    test_rule_retrieval
    test_rule_listing
    test_rule_history
    
    print_status "All Rules API tests completed successfully! 🎉"
}

# Run main function
main "$@" 