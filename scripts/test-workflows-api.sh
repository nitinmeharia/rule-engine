#!/bin/bash

# Test script for Workflows API with Comprehensive Terminal Validation
# This script tests the complete workflow lifecycle including CRUD operations,
# publishing, deactivation, dependency validation, and mandatory terminal validation.

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

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

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

# Enhanced test function with error message validation
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local test_name=$5
    local expected_error_contains=$6

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
        
        # Check error message if expected
        if [ -n "$expected_error_contains" ]; then
            if echo "$body" | grep -q "$expected_error_contains"; then
                log_success "Expected error message found: $expected_error_contains"
                ((TESTS_PASSED++))
            else
                log_error "Expected error message not found. Expected: $expected_error_contains"
                ((TESTS_FAILED++))
                return 1
            fi
        else
            ((TESTS_PASSED++))
        fi
    else
        log_error "Status code mismatch. Expected: $expected_status, Got: $status_code"
        ((TESTS_FAILED++))
        return 1
    fi

    echo
}

# Setup: Generate JWT token and create test environment
setup() {
    log_info "Setting up test environment..."
    
    # Generate JWT token
    JWT_TOKEN=$(python3 scripts/generate-jwt.py --role admin --client-id test-client --quiet)
    log_info "Generated JWT token: ${JWT_TOKEN:0:50}..."

    # Create namespace
    log_info "Creating namespace: $NAMESPACE_ID"
    namespace_data="{\"id\":\"$NAMESPACE_ID\",\"description\":\"Test namespace for workflows\"}"
    test_endpoint "POST" "/namespaces" "$namespace_data" "201" "Create namespace"

    # Create fields for rule conditions
    log_info "Creating fields for rule conditions"
    field_data="{\"fieldId\":\"status\",\"type\":\"string\",\"description\":\"Application status field\"}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/fields" "$field_data" "201" "Create status field"

    field_data="{\"fieldId\":\"salary\",\"type\":\"number\",\"description\":\"Monthly salary\"}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/fields" "$field_data" "201" "Create salary field"

    field_data="{\"fieldId\":\"credit_score\",\"type\":\"number\",\"description\":\"Credit score\"}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/fields" "$field_data" "201" "Create credit score field"

    # Create functions for workflow dependencies
    log_info "Creating functions for workflow dependencies"
    function_data="{\"id\":\"eligibility-check\",\"type\":\"in\",\"values\":[\"approved\",\"pending\"]}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/functions" "$function_data" "201" "Create eligibility function"

    function_data="{\"id\":\"income-check\",\"type\":\"max\",\"args\":[\"salary\",\"bonus\"]}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/functions" "$function_data" "201" "Create income function"

    # Publish the functions
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/functions/eligibility-check/publish" "" "200" "Publish eligibility function"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/functions/income-check/publish" "" "200" "Publish income function"

    # Create rules for workflow dependencies
    log_info "Creating rules for workflow dependencies"
    rule_data='{"id":"basic-approval","logic":"AND","conditions":[{"type":"field","fieldId":"status","operator":"==","value":"approved"}]}'
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/rules" "$rule_data" "201" "Create basic approval rule"

    rule_data='{"id":"income-rule","logic":"AND","conditions":[{"type":"field","fieldId":"salary","operator":">=","value":50000}]}'
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/rules" "$rule_data" "201" "Create income rule"

    rule_data='{"id":"credit-rule","logic":"AND","conditions":[{"type":"field","fieldId":"credit_score","operator":">=","value":700}]}'
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/rules" "$rule_data" "201" "Create credit rule"

    # Publish the rules
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/rules/basic-approval/publish" "" "200" "Publish basic approval rule"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/rules/income-rule/publish" "" "200" "Publish income rule"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/rules/credit-rule/publish" "" "200" "Publish credit rule"

    # Create terminals for workflow
    log_info "Creating terminals for workflow"
    terminal_data="{\"terminalId\":\"approve\"}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/terminals" "$terminal_data" "201" "Create approve terminal"

    terminal_data="{\"terminalId\":\"reject\"}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/terminals" "$terminal_data" "201" "Create reject terminal"

    terminal_data="{\"terminalId\":\"manual_review\"}"
    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/terminals" "$terminal_data" "201" "Create manual review terminal"

    log_success "Setup completed successfully"
    echo
}

# Test workflow creation with valid workflows
test_workflow_creation() {
    log_info "=== Testing Workflow Creation (Valid Cases) ==="

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

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$workflow_data" "201" "Create simple workflow"

    # Create a complex workflow with multiple branches
    complex_workflow_data="{
        \"id\":\"complex-approval-workflow\",
        \"startAt\":\"initial-check\",
        \"steps\":{
            \"initial-check\":{
                \"type\":\"rule\",
                \"ruleId\":\"income-rule\",
                \"onTrue\":\"credit-check\",
                \"onFalse\":\"reject\"
            },
            \"credit-check\":{
                \"type\":\"rule\",
                \"ruleId\":\"credit-rule\",
                \"onTrue\":\"approve\",
                \"onFalse\":\"manual_review\"
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
            },
            \"manual_review\":{
                \"type\":\"terminal\",
                \"terminalId\":\"manual_review\",
                \"result\":{\"status\":\"manual_review\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$complex_workflow_data" "201" "Create complex workflow"

    # Create a single terminal workflow
    single_terminal_data="{
        \"id\":\"simple-terminal-workflow\",
        \"startAt\":\"approve\",
        \"steps\":{
            \"approve\":{
                \"type\":\"terminal\",
                \"terminalId\":\"approve\",
                \"result\":{\"status\":\"approved\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$single_terminal_data" "201" "Create single terminal workflow"
}

# Test comprehensive terminal validation
test_terminal_validation() {
    log_info "=== Testing Terminal Validation (Invalid Cases) ==="

    # Test 1: Missing onTrue path
    log_info "Test 1: Missing onTrue path"
    invalid_workflow_1="{
        \"id\":\"invalid-workflow-1\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onFalse\":\"reject\"
            },
            \"reject\":{
                \"type\":\"terminal\",
                \"terminalId\":\"reject\",
                \"result\":{\"status\":\"rejected\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$invalid_workflow_1" "400" "Create workflow with missing onTrue path" "Validation Error: The 'onTrue' path for step 'check-eligibility' does not lead to a terminal"

    # Test 2: Missing onFalse path
    log_info "Test 2: Missing onFalse path"
    invalid_workflow_2="{
        \"id\":\"invalid-workflow-2\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"approve\"
            },
            \"approve\":{
                \"type\":\"terminal\",
                \"terminalId\":\"approve\",
                \"result\":{\"status\":\"approved\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$invalid_workflow_2" "400" "Create workflow with missing onFalse path" "Validation Error: The 'onFalse' path for step 'check-eligibility' does not lead to a terminal"

    # Test 3: onTrue path leads to non-terminal
    log_info "Test 3: onTrue path leads to non-terminal"
    invalid_workflow_3="{
        \"id\":\"invalid-workflow-3\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"credit-check\",
                \"onFalse\":\"reject\"
            },
            \"credit-check\":{
                \"type\":\"rule\",
                \"ruleId\":\"credit-rule\",
                \"onTrue\":\"approve\",
                \"onFalse\":\"missing_terminal\"
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

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$invalid_workflow_3" "400" "Create workflow with onTrue path leading to non-terminal" "Validation Error: The 'onTrue' path for step 'check-eligibility' does not lead to a terminal"

    # Test 4: onFalse path leads to non-terminal
    log_info "Test 4: onFalse path leads to non-terminal"
    invalid_workflow_4="{
        \"id\":\"invalid-workflow-4\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"approve\",
                \"onFalse\":\"credit-check\"
            },
            \"credit-check\":{
                \"type\":\"rule\",
                \"ruleId\":\"credit-rule\",
                \"onTrue\":\"approve\"
            },
            \"approve\":{
                \"type\":\"terminal\",
                \"terminalId\":\"approve\",
                \"result\":{\"status\":\"approved\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$invalid_workflow_4" "400" "Create workflow with onFalse path leading to non-terminal" "Validation Error: The 'onFalse' path for step 'check-eligibility' does not lead to a terminal"

    # Test 5: Step leads to non-existent step
    log_info "Test 5: Step leads to non-existent step"
    invalid_workflow_5="{
        \"id\":\"invalid-workflow-5\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"approve\",
                \"onFalse\":\"non_existent_step\"
            },
            \"approve\":{
                \"type\":\"terminal\",
                \"terminalId\":\"approve\",
                \"result\":{\"status\":\"approved\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$invalid_workflow_5" "400" "Create workflow with step leading to non-existent step" "Validation Error: The 'onFalse' path for step 'check-eligibility' does not lead to a terminal"

    # Test 6: Unknown step type
    log_info "Test 6: Unknown step type"
    invalid_workflow_6="{
        \"id\":\"invalid-workflow-6\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":{
                \"type\":\"unknown_type\"
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$invalid_workflow_6" "400" "Create workflow with unknown step type" "Validation Error: Step 'check-eligibility' is invalid or missing and does not lead to a terminal"

    # Test 7: Malformed step data
    log_info "Test 7: Malformed step data"
    invalid_workflow_7="{
        \"id\":\"invalid-workflow-7\",
        \"startAt\":\"check-eligibility\",
        \"steps\":{
            \"check-eligibility\":\"not a map\"
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$invalid_workflow_7" "400" "Create workflow with malformed step data" "Validation Error: Step 'check-eligibility' is invalid or missing and does not lead to a terminal"
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

# Test cyclic dependencies
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
                \"onFalse\":\"reject\"
            },
            \"step2\":{
                \"type\":\"rule\",
                \"ruleId\":\"basic-approval\",
                \"onTrue\":\"step1\",
                \"onFalse\":\"reject\"
            },
            \"reject\":{
                \"type\":\"terminal\",
                \"terminalId\":\"reject\",
                \"result\":{\"status\":\"rejected\"}
            }
        }
    }"

    test_endpoint "POST" "/namespaces/$NAMESPACE_ID/workflows" "$cyclic_workflow_data" "400" "Create workflow with cyclic dependencies" "cyclic dependency detected in workflow steps"
}

# Test workflow deletion
test_workflow_deletion() {
    log_info "=== Testing Workflow Deletion ==="

    # Delete workflow version
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/workflows/loan-approval-workflow/versions/1" "" "204" "Delete workflow version"

    # Try to delete non-existent workflow
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/workflows/non-existent/versions/1" "" "404" "Delete non-existent workflow"
}

# Cleanup
cleanup() {
    log_info "=== Cleaning up test data ==="

    # Delete workflows
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/workflows/complex-approval-workflow/versions/1" "" "204" "Delete complex workflow"
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/workflows/simple-terminal-workflow/versions/1" "" "204" "Delete simple terminal workflow"

    # Delete terminals
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/terminals/approve" "" "204" "Delete approve terminal"
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/terminals/reject" "" "204" "Delete reject terminal"
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/terminals/manual_review" "" "204" "Delete manual review terminal"

    # Delete rules
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/rules/basic-approval/versions/1" "" "204" "Delete basic approval rule"
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/rules/income-rule/versions/1" "" "204" "Delete income rule"
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/rules/credit-rule/versions/1" "" "204" "Delete credit rule"

    # Delete functions
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/functions/eligibility-check/versions/1" "" "204" "Delete eligibility function"
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID/functions/income-check/versions/1" "" "204" "Delete income function"

    # Delete namespace
    test_endpoint "DELETE" "/namespaces/$NAMESPACE_ID" "" "204" "Delete namespace"

    log_success "Cleanup completed successfully"
}

# Print test results
print_results() {
    echo ""
    echo "=========================================="
    log_info "Test Results Summary"
    echo "=========================================="
    log_success "Tests Passed: $TESTS_PASSED"
    if [[ $TESTS_FAILED -gt 0 ]]; then
        log_error "Tests Failed: $TESTS_FAILED"
    else
        log_success "Tests Failed: $TESTS_FAILED"
    fi

    total_tests=$((TESTS_PASSED + TESTS_FAILED))
    if [[ $total_tests -gt 0 ]]; then
        success_rate=$((TESTS_PASSED * 100 / total_tests))
        log_info "Success Rate: $success_rate%"
    fi

    echo ""
    if [[ $TESTS_FAILED -eq 0 ]]; then
        log_success "All workflow API tests passed!"
    else
        log_error "Some tests failed. Please review the output above."
    fi
}

# Main execution
main() {
    log_info "Starting Comprehensive Workflows API tests with Terminal Validation..."
    echo

    setup
    test_workflow_creation
    test_terminal_validation
    test_workflow_retrieval
    test_workflow_publishing
    test_workflow_deactivation
    test_cyclic_dependencies
    test_workflow_deletion
    cleanup
    print_results

    if [[ $TESTS_FAILED -eq 0 ]]; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@" 