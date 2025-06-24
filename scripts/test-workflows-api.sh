#!/bin/bash

# Test script for Workflows API
# This script tests the complete workflow lifecycle including CRUD operations,
# publishing, deactivation, and dependency validation.

set -e

# Configuration
API_BASE_URL="http://localhost:8080/v1"
NAMESPACE_ID="test-workflow-ns-$(date +%s)"
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

    # Make the request and capture both body and status code
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" -H "Authorization: Bearer $JWT_TOKEN" "$API_BASE_URL$endpoint")
    elif [ "$method" = "DELETE" ]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X DELETE -H "Authorization: Bearer $JWT_TOKEN" "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" -H "Content-Type: application/json" -H "Authorization: Bearer $JWT_TOKEN" -d "$data" "$API_BASE_URL$endpoint")
    fi

    # Extract status code and body
    status_code=$(echo "$response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
    body=$(echo "$response" | sed 's/HTTPSTATUS:[0-9]*$//')

    log_info "Response Status: $status_code"
    log_info "Response Body: $body"

    # Check status code
    if [ "$status_code" = "$expected_status" ]; then
        log_success "Status code matches expected: $expected_status"
    else
        log_error "Status code mismatch. Expected: $expected_status, Got: $status_code"
        return 1
    fi

    echo
}

# Setup: Generate JWT token
setup() {
    log_info "Setting up test environment..."
    
    # Generate JWT token
    JWT_TOKEN=$(python3 scripts/generate-jwt.py --role admin --client-id test-client)
    log_info "Generated JWT token: ${JWT_TOKEN:0:50}..."

    # Create namespace
    log_info "Creating namespace: $NAMESPACE_ID"
    namespace_data="{\"id\":\"$NAMESPACE_ID\",\"description\":\"Test namespace for workflows\"}"
    test_endpoint "POST" "/namespaces" "$namespace_data" "201" "Create namespace"

    # Create a field for rule conditions
    log_info "Creating field for rule conditions"
    field_data="{\"fieldId\":\"status\",\"type\":\"string\",\"description\":\"Application status field\"}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/fields" "$field_data" "201" "Create field"

    # Create a function for workflow dependencies
    log_info "Creating function for workflow dependencies"
    function_data="{\"id\":\"eligibility-check\",\"type\":\"in\",\"values\":[\"approved\",\"pending\"]}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/functions" "$function_data" "201" "Create function"

    # Publish the function
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/functions/eligibility-check/publish" "" "200" "Publish function"

    # Create a rule for workflow dependencies
    log_info "Creating rule for workflow dependencies"
    rule_data='{"id":"basic-approval","logic":"AND","conditions":[{"type":"field","fieldId":"status","operator":"==","value":"approved"}]}'
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/rules" "$rule_data" "201" "Create rule"

    # Publish the rule
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/rules/basic-approval/publish" "" "200" "Publish rule"

    # Create terminals for workflow
    log_info "Creating terminals for workflow"
    terminal_data="{\"terminalId\":\"approve\"}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/terminals" "$terminal_data" "201" "Create approve terminal"

    terminal_data="{\"terminalId\":\"reject\"}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/terminals" "$terminal_data" "201" "Create reject terminal"

    log_success "Setup completed successfully"
    echo
}

# Test workflow creation
test_workflow_creation() {
    log_info "=== Testing Workflow Creation ==="

    # Create a simple workflow
    workflow_data="{
        \"id\":\"loan-approval-workflow\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"approve\",
                \"onFalse\":\"reject\"
            },
            \"approve\":{
                \"type\":\"terminal\",
                \"terminalId\":\"approve\",
                \"result\":{\"status\":\"approved\",\"message\":\"Loan approved\"}
            },
            \"reject\":{
                \"type\":\"terminal\",
                \"terminalId\":\"reject\",
                \"result\":{\"status\":\"rejected\",\"message\":\"Loan rejected\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$workflow_data" "201" "Create workflow"
}

# Test workflow retrieval
test_workflow_retrieval() {
    log_info "=== Testing Workflow Retrieval ==="

    # Get workflow by ID (should return active version)
    test_endpoint "GET" "/namespaces/$NAMESPACE_ID/workflows/loan-approval-workflow" "" "200" "Get workflow by ID"

    # Get specific version
    test_endpoint "GET" "/namespaces/$NAMESPACE_ID/workflows/loan-approval-workflow/versions/1" "" "200" "Get workflow by version"

    # List all workflows
    test_endpoint "GET" "/namespaces/$NAMESPACE_ID/workflows" "" "200" "List all workflows"

    # List active workflows
    test_endpoint "GET" "/namespaces/$NAMESPACE_ID/workflows/active" "" "200" "List active workflows"

    # List workflow versions
    test_endpoint "GET" "/namespaces/$NAMESPACE_ID/workflows/loan-approval-workflow/versions" "" "200" "List workflow versions"
}

# Test workflow publishing
test_workflow_publishing() {
    log_info "=== Testing Workflow Publishing ==="

    # Publish the workflow
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows/loan-approval-workflow/versions/1/publish" "" "200" "Publish workflow"
}

# Test workflow deactivation
test_workflow_deactivation() {
    log_info "=== Testing Workflow Deactivation ==="

    # Deactivate the workflow
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows/loan-approval-workflow/deactivate" "" "204" "Deactivate workflow"
}

# Test error cases
test_error_cases() {
    log_info "=== Testing Error Cases ==="

    # Try to create workflow with non-existent rule
    invalid_workflow_data="{
        \"id\":\"invalid-workflow\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"rule\",
                \"ruleId\":\"non-existent-rule\",
                \"onTrue\":\"approve\",
                \"onFalse\":\"reject\"
            },
            \"approve\":{
                \"type\":\"terminal\",
                \"terminalId\":\"approve\",
                \"result\":{\"status\":\"approved\"}
            },
            \"reject\":{
                \"type\":\"terminal\",
                \"terminalId\":\"reject\",
                \"result\":{\"status\":\"rejected\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$invalid_workflow_data" "201" "Create workflow with invalid dependencies"

    # Try to publish workflow with invalid dependencies
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows/invalid-workflow/versions/1/publish" "" "400" "Publish workflow with invalid dependencies"

    # Try to get non-existent workflow
    test_endpoint "GET" "/namespaces/$NAMESPACE_ID/workflows/non-existent" "" "404" "Get non-existent workflow"

    # Try to update non-existent workflow
    update_data="{
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"approve\",
                \"onFalse\":\"reject\"
            },
            \"approve\":{
                \"type\":\"terminal\",
                \"terminalId\":\"approve\",
                \"result\":{\"status\":\"approved\"}
            },
            \"reject\":{
                \"type\":\"terminal\",
                \"terminalId\":\"reject\",
                \"result\":{\"status\":\"rejected\"}
            }
        }
    }"

    test_endpoint "PUT" "/namespaces/$NAMESPACE_ID/workflows/non-existent/versions/1" "$update_data" "404" "Update non-existent workflow"
}

# Test workflow with cyclic dependencies
test_cyclic_dependencies() {
    log_info "=== Testing Cyclic Dependencies ==="

    # Create workflow with cyclic dependencies
    cyclic_workflow_data="{
        \"id\":\"cyclic-workflow\",
        \"startAt\":\"step1\",
        \"steps\":{
            \"step1\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"step2\",
                \"onFalse\":\"step3\"
            },
            \"step2\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"step1\",
                \"onFalse\":\"step3\"
            },
            \"step3\":{
                \"type\":\"terminal\",
                \"terminalId\":\"reject\",
                \"result\":{\"status\":\"rejected\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$cyclic_workflow_data" "400" "Create workflow with cyclic dependencies"
}

# Test workflow with invalid terminal
test_invalid_terminal() {
    log_info "=== Testing Invalid Terminal ==="

    # Create workflow with non-existent terminal
    invalid_terminal_workflow_data="{
        \"id\":\"invalid-terminal-workflow\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"approve\",
                \"onFalse\":\"reject\"
            },
            \"approve\":{
                \"type\":\"terminal\",
                \"terminalId\":\"non-existent-terminal\",
                \"result\":{\"status\":\"approved\"}
            },
            \"reject\":{
                \"type\":\"terminal\",
                \"terminalId\":\"reject\",
                \"result\":{\"status\":\"rejected\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$invalid_terminal_workflow_data" "201" "Create workflow with invalid terminal"

    # Try to publish workflow with invalid terminal
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows/invalid-terminal-workflow/versions/1/publish" "" "400" "Publish workflow with invalid terminal"
}

# Test workflow deletion
test_workflow_deletion() {
    log_info "=== Testing Workflow Deletion ==="

    # Delete workflow version
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/workflows/invalid-workflow/versions/1" "" "204" "Delete workflow version"

    # Try to delete non-existent workflow
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/workflows/non-existent/versions/1" "" "404" "Delete non-existent workflow"
}

# Cleanup
cleanup() {
    log_info "=== Cleaning up test data ==="

    # Delete workflows
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/workflows/loan-approval-workflow/versions/1" "" "204" "Delete workflow"
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/workflows/invalid-terminal-workflow/versions/1" "" "204" "Delete invalid terminal workflow"

    # Delete terminals
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/terminals/approve" "" "204" "Delete approve terminal"
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/terminals/reject" "" "204" "Delete reject terminal"

    # Delete rule
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/rules/basic-approval/versions/1" "" "204" "Delete rule"

    # Delete function
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/functions/eligibility-check/versions/1" "" "204" "Delete function"

    # Delete namespace
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID" "" "204" "Delete namespace"

    log_success "Cleanup completed successfully"
}

# Main execution
main() {
    log_info "Starting Workflows API tests..."
    echo

    setup
    test_workflow_creation
    test_workflow_retrieval
    test_workflow_publishing
    test_workflow_deactivation
    test_error_cases
    test_cyclic_dependencies
    test_invalid_terminal
    test_workflow_deletion
    cleanup

    log_success "All workflow tests completed successfully!"
}

# Run main function
main "$@" 