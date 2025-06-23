package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rule-engine/internal/config"
	"github.com/rule-engine/internal/infra/db"
	"github.com/rule-engine/internal/server"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite provides end-to-end testing
type IntegrationTestSuite struct {
	suite.Suite
	pool   *pgxpool.Pool
	server *server.Server
	cfg    *config.Config
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Load test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Name:            "rule_engine_test",
			User:            "postgres",
			Password:        "postgres",
			SSLMode:         "disable",
			MaxConnections:  10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
			QueryTimeout:    30 * time.Second,
		},
		Cache: config.CacheConfig{
			RefreshIntervalSec: 60,
			RefreshJitterSec:   10,
			RefreshTimeout:     10 * time.Second,
			MaxSizeMB:          100,
		},
		Logger: config.LoggerConfig{
			Level:      "info",
			Format:     "json",
			TimeFormat: "rfc3339",
		},
		Server: config.ServerConfig{
			Host:            "localhost",
			Port:            8080,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
	}

	// Initialize database connection
	pool, err := db.New(cfg.Database)
	suite.Require().NoError(err)

	// Clean up any existing test data
	suite.cleanupTestData(pool)

	// Create test server (without starting it)
	srv, err := server.New(cfg, pool, nil) // logger can be nil for tests
	suite.Require().NoError(err)

	suite.pool = pool
	suite.server = srv
	suite.cfg = cfg
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.pool != nil {
		suite.cleanupTestData(suite.pool)
		suite.pool.Close()
	}
}

func (suite *IntegrationTestSuite) cleanupTestData(pool *pgxpool.Pool) {
	ctx := context.Background()

	// Clean up in reverse dependency order
	_, _ = pool.Exec(ctx, "DELETE FROM active_config_meta WHERE namespace LIKE 'test-%'")
	_, _ = pool.Exec(ctx, "DELETE FROM workflows WHERE namespace LIKE 'test-%'")
	_, _ = pool.Exec(ctx, "DELETE FROM rules WHERE namespace LIKE 'test-%'")
	_, _ = pool.Exec(ctx, "DELETE FROM functions WHERE namespace LIKE 'test-%'")
	_, _ = pool.Exec(ctx, "DELETE FROM terminals WHERE namespace LIKE 'test-%'")
	_, _ = pool.Exec(ctx, "DELETE FROM fields WHERE namespace LIKE 'test-%'")
	_, _ = pool.Exec(ctx, "DELETE FROM namespaces WHERE id LIKE 'test-%'")
}

func (suite *IntegrationTestSuite) TestCompleteRuleEngineWorkflow() {
	// This test demonstrates the complete rule engine workflow:
	// 1. Create namespace
	// 2. Create fields
	// 3. Create rules
	// 4. Execute rules
	// 5. Verify results

	testNamespace := "test-complete-workflow"

	// Step 1: Create namespace
	namespacePayload := map[string]interface{}{
		"id":          testNamespace,
		"description": "Test namespace for complete workflow",
	}

	resp := suite.makeRequest("POST", "/v1/namespaces", namespacePayload)
	suite.Equal(http.StatusCreated, resp.Code)

	// Step 2: Create fields
	fieldPayloads := []map[string]interface{}{
		{
			"fieldId": "age",
			"type":    "number",
		},
		{
			"fieldId": "income",
			"type":    "number",
		},
		{
			"fieldId": "country",
			"type":    "string",
		},
	}

	for _, fieldPayload := range fieldPayloads {
		resp := suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/fields", testNamespace), fieldPayload)
		suite.Equal(http.StatusCreated, resp.Code)
	}

	// Step 3: Create a rule
	rulePayload := map[string]interface{}{
		"ruleId": "eligibility-check",
		"logic":  "AND",
		"conditions": map[string]interface{}{
			"condition1": map[string]interface{}{
				"field":    "age",
				"operator": "gte",
				"value":    18,
			},
			"condition2": map[string]interface{}{
				"field":    "income",
				"operator": "gt",
				"value":    50000,
			},
			"condition3": map[string]interface{}{
				"field":    "country",
				"operator": "in",
				"value":    []string{"US", "CA", "UK"},
			},
		},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/rules", testNamespace), rulePayload)
	suite.Equal(http.StatusCreated, resp.Code)

	// Step 4: Publish the rule
	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/rules/eligibility-check/versions/1/publish", testNamespace), nil)
	suite.Equal(http.StatusOK, resp.Code)

	// Step 5: Execute rule with test data (should pass)
	executionPayload := map[string]interface{}{
		"data": map[string]interface{}{
			"age":     25,
			"income":  75000,
			"country": "US",
		},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/rules/eligibility-check/execute", testNamespace), executionPayload)
	suite.Equal(http.StatusOK, resp.Code)

	var executionResponse map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &executionResponse)
	suite.NoError(err)
	suite.Equal(true, executionResponse["result"])

	// Step 6: Execute rule with test data (should fail)
	executionPayload = map[string]interface{}{
		"data": map[string]interface{}{
			"age":     17, // Under 18
			"income":  75000,
			"country": "US",
		},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/rules/eligibility-check/execute", testNamespace), executionPayload)
	suite.Equal(http.StatusOK, resp.Code)

	err = json.Unmarshal(resp.Body.Bytes(), &executionResponse)
	suite.NoError(err)
	suite.Equal(false, executionResponse["result"])

	// Step 7: Execute with trace enabled
	executionPayload = map[string]interface{}{
		"data": map[string]interface{}{
			"age":     25,
			"income":  75000,
			"country": "US",
		},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/rules/eligibility-check/execute?trace=full", testNamespace), executionPayload)
	suite.Equal(http.StatusOK, resp.Code)

	err = json.Unmarshal(resp.Body.Bytes(), &executionResponse)
	suite.NoError(err)
	suite.Equal(true, executionResponse["result"])
	suite.NotNil(executionResponse["trace"])

	// Verify trace contains expected information
	trace, ok := executionResponse["trace"].(map[string]interface{})
	suite.True(ok)
	suite.NotEmpty(trace["duration"])
	suite.NotEmpty(trace["steps"])
}

func (suite *IntegrationTestSuite) TestWorkflowExecution() {
	// Test workflow execution with multiple steps
	testNamespace := "test-workflow"

	// Create namespace
	namespacePayload := map[string]interface{}{
		"id":          testNamespace,
		"description": "Test namespace for workflow",
	}

	resp := suite.makeRequest("POST", "/v1/namespaces", namespacePayload)
	suite.Equal(http.StatusCreated, resp.Code)

	// Create fields
	fieldPayload := map[string]interface{}{
		"fieldId": "score",
		"type":    "number",
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/fields", testNamespace), fieldPayload)
	suite.Equal(http.StatusCreated, resp.Code)

	// Create rules
	highScoreRule := map[string]interface{}{
		"ruleId": "high-score",
		"logic":  "AND",
		"conditions": map[string]interface{}{
			"condition1": map[string]interface{}{
				"field":    "score",
				"operator": "gte",
				"value":    80,
			},
		},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/rules", testNamespace), highScoreRule)
	suite.Equal(http.StatusCreated, resp.Code)

	// Publish rule
	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/rules/high-score/versions/1/publish", testNamespace), nil)
	suite.Equal(http.StatusOK, resp.Code)

	// Create terminals
	terminalPayloads := []map[string]interface{}{
		{"terminalId": "approved"},
		{"terminalId": "rejected"},
	}

	for _, terminalPayload := range terminalPayloads {
		resp := suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/terminals", testNamespace), terminalPayload)
		suite.Equal(http.StatusCreated, resp.Code)
	}

	// Create workflow
	workflowPayload := map[string]interface{}{
		"workflowId": "score-evaluation",
		"startAt":    "check-score",
		"steps": map[string]interface{}{
			"check-score": map[string]interface{}{
				"type":    "rule",
				"ruleId":  "high-score",
				"onTrue":  "approve",
				"onFalse": "reject",
			},
			"approve": map[string]interface{}{
				"type":       "terminal",
				"terminalId": "approved",
				"result":     "APPROVED",
			},
			"reject": map[string]interface{}{
				"type":       "terminal",
				"terminalId": "rejected",
				"result":     "REJECTED",
			},
		},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/workflows", testNamespace), workflowPayload)
	suite.Equal(http.StatusCreated, resp.Code)

	// Publish workflow
	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/workflows/score-evaluation/versions/1/publish", testNamespace), nil)
	suite.Equal(http.StatusOK, resp.Code)

	// Execute workflow (should approve)
	executionPayload := map[string]interface{}{
		"data": map[string]interface{}{
			"score": 85,
		},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/workflows/score-evaluation/execute", testNamespace), executionPayload)
	suite.Equal(http.StatusOK, resp.Code)

	var executionResponse map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &executionResponse)
	suite.NoError(err)
	suite.Equal("APPROVED", executionResponse["result"])

	// Execute workflow (should reject)
	executionPayload = map[string]interface{}{
		"data": map[string]interface{}{
			"score": 65,
		},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/workflows/score-evaluation/execute", testNamespace), executionPayload)
	suite.Equal(http.StatusOK, resp.Code)

	err = json.Unmarshal(resp.Body.Bytes(), &executionResponse)
	suite.NoError(err)
	suite.Equal("REJECTED", executionResponse["result"])
}

func (suite *IntegrationTestSuite) TestErrorHandling() {
	// Test various error conditions

	// Test namespace not found
	resp := suite.makeRequest("GET", "/v1/namespaces/non-existent", nil)
	suite.Equal(http.StatusNotFound, resp.Code)

	// Test invalid JSON
	req := httptest.NewRequest("POST", "/v1/namespaces", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	suite.server.ServeHTTP(resp, req)
	suite.Equal(http.StatusBadRequest, resp.Code)

	// Test missing required fields
	invalidPayload := map[string]interface{}{
		"description": "Missing ID",
	}
	resp = suite.makeRequest("POST", "/v1/namespaces", invalidPayload)
	suite.Equal(http.StatusBadRequest, resp.Code)
}

func (suite *IntegrationTestSuite) makeRequest(method, path string, payload interface{}) *httptest.ResponseRecorder {
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		suite.Require().NoError(err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp := httptest.NewRecorder()
	suite.server.ServeHTTP(resp, req)

	return resp
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	// Skip integration tests if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(IntegrationTestSuite))
}
