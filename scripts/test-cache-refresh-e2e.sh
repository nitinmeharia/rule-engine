#!/bin/bash

# E2E Test script for Cache Refresh Loop functionality
# This script tests the complete cache refresh workflow

set -e

echo "🧪 Testing Cache Refresh Loop E2E"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✅ PASS${NC}: $message"
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}❌ FAIL${NC}: $message"
    elif [ "$status" = "INFO" ]; then
        echo -e "${YELLOW}ℹ️  INFO${NC}: $message"
    elif [ "$status" = "STEP" ]; then
        echo -e "${BLUE}🔧 STEP${NC}: $message"
    fi
}

# Configuration
API_HOST="localhost"
API_PORT="8080"
ADMIN_TOKEN=""
TEST_NAMESPACE="cache-refresh-test-$(date +%s)"
SERVICE_PID=""
DB_SETUP_SCRIPT="scripts/setup-test-db.sh"

# Function to cleanup test data
cleanup() {
    print_status "STEP" "Cleaning up test data..."
    
    # Stop the service if it's running
    if [ ! -z "$SERVICE_PID" ]; then
        print_status "INFO" "Stopping service (PID: $SERVICE_PID)..."
        kill $SERVICE_PID 2>/dev/null || true
        wait $SERVICE_PID 2>/dev/null || true
    fi
    
    # Kill any remaining processes on the port
    lsof -ti:$API_PORT | xargs kill -9 2>/dev/null || true
    
    print_status "PASS" "Cleanup completed"
}

# Function to handle script exit
on_exit() {
    cleanup
}

# Set up exit handler
trap on_exit EXIT

# Function to wait for service to be ready
wait_for_service() {
    local max_attempts=30
    local attempt=1
    
    print_status "STEP" "Waiting for service to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "http://$API_HOST:$API_PORT/health" > /dev/null 2>&1; then
            print_status "PASS" "Service is ready"
            return 0
        fi
        
        print_status "INFO" "Attempt $attempt/$max_attempts: Service not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_status "FAIL" "Service failed to start within expected time"
    return 1
}

# Function to generate JWT token
generate_jwt_token() {
    print_status "STEP" "Generating JWT token..."
    
    if [ ! -f "./jwt-generator" ]; then
        print_status "STEP" "Building JWT generator..."
        go build -o jwt-generator ./cmd/jwt-generator
    fi
    
    ADMIN_TOKEN=$(./jwt-generator | tail -n 1)
    
    if [ -z "$ADMIN_TOKEN" ]; then
        print_status "FAIL" "Failed to generate JWT token"
        exit 1
    fi
    
    print_status "PASS" "JWT token generated successfully"
}

# Function to create test namespace
create_test_namespace() {
    print_status "STEP" "Creating test namespace: $TEST_NAMESPACE"
    
    local payload="{
        \"id\": \"$TEST_NAMESPACE\",
        \"description\": \"Test namespace for cache refresh E2E test\"
    }"
    
    local response=$(curl -s -w "%{http_code}" -X POST "http://$API_HOST:$API_PORT/v1/namespaces" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "$payload")
    
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$status_code" = "201" ]; then
        print_status "PASS" "Test namespace created successfully"
    elif [ "$status_code" = "409" ]; then
        print_status "INFO" "Test namespace already exists, continuing..."
    else
        print_status "FAIL" "Failed to create test namespace. Status: $status_code, Response: $body"
        return 1
    fi
}

# Function to create test function
create_test_function() {
    local function_id="$1"
    local function_type="$2"
    local args_or_values="$3"
    
    print_status "STEP" "Creating test function: $function_id"
    
    local payload
    if [ "$function_type" = "in" ]; then
        payload="{
            \"id\": \"$function_id\",
            \"type\": \"$function_type\",
            \"values\": $args_or_values
        }"
    else
        payload="{
            \"id\": \"$function_id\",
            \"type\": \"$function_type\",
            \"args\": $args_or_values
        }"
    fi
    
    local response=$(curl -s -w "%{http_code}" -X POST "http://$API_HOST:$API_PORT/v1/namespaces/$TEST_NAMESPACE/functions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "$payload")
    
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$status_code" = "201" ]; then
        print_status "PASS" "Test function $function_id created successfully"
    else
        print_status "FAIL" "Failed to create test function $function_id. Status: $status_code, Response: $body"
        return 1
    fi
}

# Function to publish function
publish_function() {
    local function_id="$1"
    
    print_status "STEP" "Publishing function: $function_id"
    
    local response=$(curl -s -w "%{http_code}" -X POST "http://$API_HOST:$API_PORT/v1/namespaces/$TEST_NAMESPACE/functions/$function_id/publish" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$status_code" = "200" ]; then
        print_status "PASS" "Function $function_id published successfully"
    else
        print_status "FAIL" "Failed to publish function $function_id. Status: $status_code, Response: $body"
        return 1
    fi
}

# Function to trigger checksum refresh for namespace
trigger_checksum_refresh() {
    local namespace="$1"
    
    print_status "STEP" "Triggering checksum refresh for namespace: $namespace"
    
    # Create a temporary function and publish it to trigger checksum refresh
    local temp_function_id="temp-trigger-$(date +%s)"
    
    local payload="{
        \"id\": \"$temp_function_id\",
        \"type\": \"in\",
        \"values\": [\"trigger\"]
    }"
    
    # Create the function
    local create_response=$(curl -s -w "%{http_code}" -X POST "http://$API_HOST:$API_PORT/v1/namespaces/$namespace/functions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "$payload")
    
    local create_status="${create_response: -3}"
    
    if [ "$create_status" = "201" ]; then
        # Publish the function to trigger checksum refresh
        local publish_response=$(curl -s -w "%{http_code}" -X POST "http://$API_HOST:$API_PORT/v1/namespaces/$namespace/functions/$temp_function_id/publish" \
            -H "Authorization: Bearer $ADMIN_TOKEN")
        
        local publish_status="${publish_response: -3}"
        
        if [ "$publish_status" = "200" ]; then
            print_status "PASS" "Checksum refresh triggered successfully"
            
            # Delete the temporary function
            curl -s -X DELETE "http://$API_HOST:$API_PORT/v1/namespaces/$namespace/functions/$temp_function_id/versions/1" \
                -H "Authorization: Bearer $ADMIN_TOKEN" > /dev/null 2>&1
            
            return 0
        else
            print_status "FAIL" "Failed to publish temporary function. Status: $publish_status"
            return 1
        fi
    else
        print_status "FAIL" "Failed to create temporary function. Status: $create_status"
        return 1
    fi
}

# Function to wait for namespace to be loaded into cache
wait_for_namespace_in_cache() {
    local namespace="$1"
    local max_attempts=30
    local attempt=1
    
    print_status "STEP" "Waiting for namespace $namespace to be loaded into cache..."
    
    while [ $attempt -le $max_attempts ]; do
        local response=$(curl -s -w "%{http_code}" -X GET "http://$API_HOST:$API_PORT/admin/cache/stats/$namespace" \
            -H "Authorization: Bearer $ADMIN_TOKEN")
        
        local status_code="${response: -3}"
        local body="${response%???}"
        
        if [ "$status_code" = "200" ]; then
            local checksum=$(echo "$body" | jq -r '.data.checksum')
            if [ "$checksum" != "null" ] && [ "$checksum" != "" ]; then
                print_status "PASS" "Namespace $namespace loaded into cache with checksum: $checksum"
                return 0
            fi
        fi
        
        print_status "INFO" "Attempt $attempt/$max_attempts: Namespace not in cache yet"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_status "FAIL" "Namespace $namespace was not loaded into cache within expected time"
    return 1
}

# Function to get cache stats
get_cache_stats() {
    local namespace="$1"
    
    local response=$(curl -s -w "%{http_code}" -X GET "http://$API_HOST:$API_PORT/admin/cache/stats/$namespace" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$status_code" = "200" ]; then
        echo "$body"
    else
        print_status "FAIL" "Failed to get cache stats for $namespace. Status: $status_code, Response: $body"
        return 1
    fi
}

# Function to wait for cache refresh
wait_for_cache_refresh() {
    local initial_checksum="$1"
    local max_attempts=30
    local attempt=1
    
    print_status "STEP" "Waiting for cache refresh (initial checksum: $initial_checksum)..."
    
    while [ $attempt -le $max_attempts ]; do
        local stats=$(get_cache_stats "$TEST_NAMESPACE")
        local current_checksum=$(echo "$stats" | jq -r '.data.checksum')
        
        if [ "$current_checksum" != "$initial_checksum" ]; then
            print_status "PASS" "Cache refreshed! New checksum: $current_checksum"
            return 0
        fi
        
        print_status "INFO" "Attempt $attempt/$max_attempts: Cache not refreshed yet (checksum: $current_checksum)"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_status "FAIL" "Cache refresh did not occur within expected time"
    return 1
}

# Function to verify function is in cache
verify_function_in_cache() {
    local function_id="$1"
    
    print_status "STEP" "Verifying function $function_id is in cache..."
    
    local response=$(curl -s -w "%{http_code}" -X GET "http://$API_HOST:$API_PORT/v1/namespaces/$TEST_NAMESPACE/functions/$function_id" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$status_code" = "200" ]; then
        local cached_function_id=$(echo "$body" | jq -r '.data.id')
        if [ "$cached_function_id" = "$function_id" ]; then
            print_status "PASS" "Function $function_id found in cache"
        else
            print_status "FAIL" "Function $function_id not found in cache"
            return 1
        fi
    else
        print_status "FAIL" "Failed to get function $function_id. Status: $status_code, Response: $body"
        return 1
    fi
}

# Main test execution
main() {
    print_status "INFO" "Starting Cache Refresh E2E Test"
    
    # Step 1: Generate JWT token
    generate_jwt_token
    
    # Step 2: Start the service
    print_status "STEP" "Starting the service..."
    go run ./cmd/api/main.go &
    SERVICE_PID=$!
    
    # Step 3: Wait for service to be ready
    wait_for_service
    
    # Step 4: Create test namespace
    create_test_namespace
    
    # Step 4.5: Wait a moment for namespace to be fully committed
    print_status "STEP" "Waiting for namespace to be fully committed..."
    sleep 2
    
    # Step 5: Trigger checksum refresh to ensure namespace is loaded into cache
    trigger_checksum_refresh "$TEST_NAMESPACE"
    
    # Step 6: Wait for namespace to be loaded into cache
    wait_for_namespace_in_cache "$TEST_NAMESPACE"
    
    # Step 7: Get initial cache stats
    print_status "STEP" "Getting initial cache stats..."
    local initial_stats=$(get_cache_stats "$TEST_NAMESPACE")
    local initial_checksum=$(echo "$initial_stats" | jq -r '.data.checksum')
    print_status "INFO" "Initial checksum: $initial_checksum"
    
    # Step 8: Create and publish a function (this should trigger cache refresh)
    create_test_function "eligibility-check" "in" '["US", "CA", "UK"]'
    publish_function "eligibility-check"
    
    # Step 9: Wait for cache refresh
    wait_for_cache_refresh "$initial_checksum"
    
    # Step 10: Verify the function is now in cache
    verify_function_in_cache "eligibility-check"
    
    # Step 11: Create and publish another function
    create_test_function "high-score" "max" '["score"]'
    publish_function "high-score"
    
    # Step 12: Get updated cache stats
    local updated_stats=$(get_cache_stats "$TEST_NAMESPACE")
    local updated_checksum=$(echo "$updated_stats" | jq -r '.data.checksum')
    print_status "INFO" "Updated checksum: $updated_checksum"
    
    # Step 13: Verify both functions are in cache
    verify_function_in_cache "eligibility-check"
    verify_function_in_cache "high-score"
    
    print_status "PASS" "All cache refresh tests completed successfully!"
    
    # Print final cache stats
    echo ""
    print_status "INFO" "Final Cache Stats:"
    echo "$updated_stats" | jq '.'
}

# Run the main function
main

echo ""
echo "🎉 Cache Refresh E2E Test Completed Successfully!"
echo ""
echo "Test Summary:"
echo "============="
echo "✅ Service startup and health check"
echo "✅ JWT token generation"
echo "✅ Test namespace creation"
echo "✅ Function creation and publishing"
echo "✅ Cache refresh detection"
echo "✅ Cache verification"
echo "✅ Admin endpoint functionality"
echo ""
print_status "INFO" "Cache refresh functionality is working correctly!"
