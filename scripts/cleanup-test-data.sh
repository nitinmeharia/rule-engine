#!/bin/bash
# Shared script to clean up test data for all test suites

print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")
            echo -e "\033[1;34mℹ INFO\033[0m: $message"
            ;;
    esac
}

print_status "INFO" "Cleaning up test data (shared script)"
psql -U postgres -d rule_engine_dev -c "DELETE FROM rules WHERE namespace IN ('test-e2e', 'test-functions', 'fields-test-ns', 'rbac-test-ns', 'quick-test-ns', 'fields-quick-test'); DELETE FROM functions WHERE namespace IN ('test-e2e', 'test-functions', 'fields-test-ns', 'rbac-test-ns', 'quick-test-ns', 'fields-quick-test'); DELETE FROM fields WHERE namespace IN ('test-e2e', 'test-functions', 'fields-test-ns', 'rbac-test-ns', 'quick-test-ns', 'fields-quick-test'); DELETE FROM namespaces WHERE id IN ('test-e2e', 'test-functions', 'fields-test-ns', 'rbac-test-ns', 'quick-test-ns', 'fields-quick-test');" > /dev/null 2>&1 || true 