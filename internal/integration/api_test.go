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

	importedJwt "github.com/golang-jwt/jwt/v4"
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
	token  string // JWT token for Authorization header
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
		JWT: config.JWTConfig{
			Secret:          "dev-secret-key-change-in-production",
			TokenExpiration: 24 * time.Hour,
			RequiredClaims:  []string{"clientId", "role"},
			ValidAudiences:  []string{},
			ValidIssuers:    []string{},
			SkipExpiryCheck: false,
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

	// Generate JWT token for tests
	// suite.token = generateTestJWT()
	suite.token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRJZCI6InRlc3QtY2xpZW50Iiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzUwODMzOTgxLCJuYmYiOjE3NTA3NDc1ODEsImlhdCI6MTc1MDc0NzU4MX0.OEbr07R0P28uNNwofltPBc3quvN9QwO-6L2i48zJ0os"
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
	// This test demonstrates the available rule engine workflow:
	// 1. Create namespace
	// 2. Create fields
	// 3. Create functions
	// 4. Update and publish functions

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
			"fieldId":     "age",
			"type":        "number",
			"description": "User age field",
		},
		{
			"fieldId":     "income",
			"type":        "number",
			"description": "User income field",
		},
		{
			"fieldId":     "country",
			"type":        "string",
			"description": "User country field",
		},
	}

	for _, fieldPayload := range fieldPayloads {
		resp := suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/fields", testNamespace), fieldPayload)
		suite.Equal(http.StatusCreated, resp.Code)
	}

	// Step 3: Create a function
	functionPayload := map[string]interface{}{
		"id":     "eligibility-check",
		"type":   "in",
		"values": []string{"US", "CA", "UK"},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/functions", testNamespace), functionPayload)
	suite.Equal(http.StatusCreated, resp.Code)

	// Step 4: Update the function
	updatePayload := map[string]interface{}{
		"type":   "in",
		"values": []string{"US", "CA", "UK", "AU"}, // Updated values
	}

	resp = suite.makeRequest("PUT", fmt.Sprintf("/v1/namespaces/%s/functions/eligibility-check/versions/draft", testNamespace), updatePayload)
	suite.Equal(http.StatusOK, resp.Code)

	// Step 5: Publish the function
	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/functions/eligibility-check/publish", testNamespace), nil)
	suite.Equal(http.StatusOK, resp.Code)

	// Step 6: Get the function to verify it was created and published
	resp = suite.makeRequest("GET", fmt.Sprintf("/v1/namespaces/%s/functions/eligibility-check", testNamespace), nil)
	suite.Equal(http.StatusOK, resp.Code)

	var functionResponse map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &functionResponse)
	suite.NoError(err)

	// Check the nested data structure
	data := functionResponse["data"].(map[string]interface{})
	suite.Equal("eligibility-check", data["id"])
	suite.Equal("in", data["type"])
	suite.Equal("active", data["status"]) // Status should be "active" after publishing, not "draft"
}

func (suite *IntegrationTestSuite) TestWorkflowExecution() {
	// Test basic workflow with available endpoints
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
		"fieldId":     "score",
		"type":        "number",
		"description": "User score field",
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/fields", testNamespace), fieldPayload)
	suite.Equal(http.StatusCreated, resp.Code)

	// Create a function
	functionPayload := map[string]interface{}{
		"id":   "high-score",
		"type": "max",
		"args": []string{"score"},
	}

	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/functions", testNamespace), functionPayload)
	suite.Equal(http.StatusCreated, resp.Code)

	// Publish the function
	resp = suite.makeRequest("POST", fmt.Sprintf("/v1/namespaces/%s/functions/high-score/publish", testNamespace), nil)
	suite.Equal(http.StatusOK, resp.Code)

	// Get the function to verify it was created and published
	resp = suite.makeRequest("GET", fmt.Sprintf("/v1/namespaces/%s/functions/high-score", testNamespace), nil)
	suite.Equal(http.StatusOK, resp.Code)

	var functionResponse map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &functionResponse)
	suite.NoError(err)

	// Check the nested data structure
	data := functionResponse["data"].(map[string]interface{})
	suite.Equal("high-score", data["id"])
	suite.Equal("max", data["type"])
	suite.Equal("active", data["status"]) // Status should be "active" after publishing

	// List all functions in the namespace
	resp = suite.makeRequest("GET", fmt.Sprintf("/v1/namespaces/%s/functions", testNamespace), nil)
	suite.Equal(http.StatusOK, resp.Code)

	var functionsResponse map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &functionsResponse)
	suite.NoError(err)

	// Check the nested data structure for list response
	functionsData := functionsResponse["data"].([]interface{})
	suite.Len(functionsData, 1)

	firstFunction := functionsData[0].(map[string]interface{})
	suite.Equal("high-score", firstFunction["id"])
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
	suite.Equal(http.StatusUnauthorized, resp.Code)

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
	// Always set Authorization header
	req.Header.Set("Authorization", "Bearer "+suite.token)

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

// Helper to generate a JWT token for tests
func generateTestJWT() string {
	// These values should match the defaults in generate-jwt.py
	claims := map[string]interface{}{
		"clientId": "test-client",
		"role":     "admin",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
		"nbf":      time.Now().Unix(),
	}
	secret := "dev-secret-key-change-in-production"
	// Create token
	token := importedJwt.NewWithClaims(importedJwt.SigningMethodHS256, importedJwt.MapClaims(claims))
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic("Failed to sign JWT for tests: " + err.Error())
	}
	return tokenString
}
