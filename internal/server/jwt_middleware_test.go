package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rule-engine/internal/auth"
	"github.com/rule-engine/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.JWTConfig{
		Secret:          "test-secret-key",
		RequiredClaims:  []string{"clientId", "role"},
		SkipExpiryCheck: false,
	}

	tests := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
		expectedBody   map[string]interface{}
		setupToken     func() string
	}{
		{
			name:           "health_check_skipped",
			path:           "/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			setupToken:     func() string { return "" },
		},
		{
			name:           "missing_auth_header",
			path:           "/test",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error":   "UNAUTHORIZED",
				"message": "Missing Authorization header",
				"code":    "MISSING_AUTH_HEADER",
			},
			setupToken: func() string { return "" },
		},
		{
			name:           "invalid_auth_format",
			path:           "/test",
			authHeader:     "InvalidFormat token123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error":   "UNAUTHORIZED",
				"message": "Invalid Authorization header format",
				"code":    "INVALID_AUTH_FORMAT",
			},
			setupToken: func() string { return "" },
		},
		{
			name:           "valid_jwt_token",
			path:           "/test",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusOK,
			setupToken: func() string {
				claims := &JWTClaims{
					ClientID: "test-client",
					Role:     "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.Secret))
				return tokenString
			},
		},
		{
			name:           "expired_jwt_token",
			path:           "/test",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error":   "UNAUTHORIZED",
				"message": "Invalid JWT token",
				"code":    "INVALID_JWT_TOKEN",
			},
			setupToken: func() string {
				claims := &JWTClaims{
					ClientID: "test-client",
					Role:     "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.Secret))
				return tokenString
			},
		},
		{
			name:           "missing_client_id_claim",
			path:           "/test",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error":   "UNAUTHORIZED",
				"message": "Missing required JWT claims",
				"code":    "MISSING_REQUIRED_CLAIMS",
			},
			setupToken: func() string {
				claims := &JWTClaims{
					Role: "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.Secret))
				return tokenString
			},
		},
		{
			name:           "missing_role_claim",
			path:           "/test",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error":   "UNAUTHORIZED",
				"message": "Missing required JWT claims",
				"code":    "MISSING_REQUIRED_CLAIMS",
			},
			setupToken: func() string {
				claims := &JWTClaims{
					ClientID: "test-client",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.Secret))
				return tokenString
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config for this test
			testCfg := *cfg
			if tt.name == "skip_expiry_check" {
				testCfg.SkipExpiryCheck = true
			}

			router := gin.New()
			router.Use(JWTMiddleware(&testCfg))

			router.GET("/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			router.GET("/test", func(c *gin.Context) {
				clientID := auth.GetClientID(c)
				role := auth.GetUserRole(c)
				roles := auth.GetUserRoles(c)

				c.JSON(http.StatusOK, gin.H{
					"client_id": clientID,
					"role":      role,
					"roles":     roles,
				})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)

			if tt.authHeader != "" {
				token := tt.setupToken()
				req.Header.Set("Authorization", tt.authHeader+token)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}

			// For successful requests, verify context values
			if tt.expectedStatus == http.StatusOK && tt.name != "health_check_skipped" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.name == "valid_jwt_token" {
					assert.Equal(t, "test-client", response["client_id"])
					assert.Equal(t, "user", response["role"])
				}
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requiredRole   string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:         "user_has_required_role",
			requiredRole: "user",
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
				c.Set("client_id", "test-client")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "user_has_role_in_roles_array",
			requiredRole: "admin",
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
				c.Set("roles", []string{"admin", "user"})
				c.Set("client_id", "test-client")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "admin_has_access_to_everything",
			requiredRole: "superuser",
			setupContext: func(c *gin.Context) {
				c.Set("role", "admin")
				c.Set("client_id", "test-client")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "admin_in_roles_array_has_access",
			requiredRole: "superuser",
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
				c.Set("roles", []string{"admin", "user"})
				c.Set("client_id", "test-client")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "user_insufficient_permissions",
			requiredRole: "admin",
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
				c.Set("client_id", "test-client")
			},
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]interface{}{
				"error":   "FORBIDDEN",
				"message": "Insufficient permissions",
				"code":    "INSUFFICIENT_PERMISSIONS",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequireRole(tt.requiredRole))

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)

			// Create a new gin context for the request
			ctx := gin.CreateTestContextOnly(w, router)
			ctx.Request = req

			// Setup the context with test data
			if tt.setupContext != nil {
				tt.setupContext(ctx)
			}

			// Call the middleware directly
			RequireRole(tt.requiredRole)(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestRequireAnyRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requiredRoles  []string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:          "user_has_one_of_required_roles",
			requiredRoles: []string{"admin", "user"},
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
				c.Set("client_id", "test-client")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "user_has_role_in_roles_array",
			requiredRoles: []string{"admin", "superuser"},
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
				c.Set("roles", []string{"admin", "user"})
				c.Set("client_id", "test-client")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "admin_has_access_to_everything",
			requiredRoles: []string{"superuser", "moderator"},
			setupContext: func(c *gin.Context) {
				c.Set("role", "admin")
				c.Set("client_id", "test-client")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "user_insufficient_permissions",
			requiredRoles: []string{"admin", "superuser"},
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
				c.Set("client_id", "test-client")
			},
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]interface{}{
				"error":   "FORBIDDEN",
				"message": "Insufficient permissions",
				"code":    "INSUFFICIENT_PERMISSIONS",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequireAnyRole(tt.requiredRoles...))

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)

			// Create a new gin context for the request
			ctx := gin.CreateTestContextOnly(w, router)
			ctx.Request = req

			// Setup the context with test data
			if tt.setupContext != nil {
				tt.setupContext(ctx)
			}

			// Call the middleware directly
			RequireAnyRole(tt.requiredRoles...)(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestValidateRequiredClaims(t *testing.T) {
	tests := []struct {
		name           string
		claims         *JWTClaims
		requiredClaims []string
		expectedError  bool
		errorMessage   string
	}{
		{
			name: "all_required_claims_present",
			claims: &JWTClaims{
				ClientID: "test-client",
				Role:     "user",
			},
			requiredClaims: []string{"clientId", "role"},
			expectedError:  false,
		},
		{
			name: "missing_client_id",
			claims: &JWTClaims{
				Role: "user",
			},
			requiredClaims: []string{"clientId", "role"},
			expectedError:  true,
			errorMessage:   "missing required claim: clientId",
		},
		{
			name: "missing_role",
			claims: &JWTClaims{
				ClientID: "test-client",
			},
			requiredClaims: []string{"clientId", "role"},
			expectedError:  true,
			errorMessage:   "missing required claim: role",
		},
		{
			name: "empty_client_id",
			claims: &JWTClaims{
				ClientID: "",
				Role:     "user",
			},
			requiredClaims: []string{"clientId", "role"},
			expectedError:  true,
			errorMessage:   "missing required claim: clientId",
		},
		{
			name: "empty_role",
			claims: &JWTClaims{
				ClientID: "test-client",
				Role:     "",
			},
			requiredClaims: []string{"clientId", "role"},
			expectedError:  true,
			errorMessage:   "missing required claim: role",
		},
		{
			name: "only_client_id_required",
			claims: &JWTClaims{
				ClientID: "test-client",
			},
			requiredClaims: []string{"clientId"},
			expectedError:  false,
		},
		{
			name: "only_role_required",
			claims: &JWTClaims{
				Role: "user",
			},
			requiredClaims: []string{"role"},
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredClaims(tt.claims, tt.requiredClaims)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "value_exists_in_slice",
			slice:    []string{"admin", "user", "moderator"},
			value:    "user",
			expected: true,
		},
		{
			name:     "value_does_not_exist_in_slice",
			slice:    []string{"admin", "user", "moderator"},
			value:    "superuser",
			expected: false,
		},
		{
			name:     "empty_slice",
			slice:    []string{},
			value:    "user",
			expected: false,
		},
		{
			name:     "nil_slice",
			slice:    nil,
			value:    "user",
			expected: false,
		},
		{
			name:     "case_sensitive_match",
			slice:    []string{"Admin", "User", "Moderator"},
			value:    "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJWTClaims(t *testing.T) {
	// Test JWTClaims struct marshaling/unmarshaling
	claims := &JWTClaims{
		ClientID: "test-client",
		Role:     "user",
		Roles:    []string{"user", "moderator"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "test-subject",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(claims)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaledClaims JWTClaims
	err = json.Unmarshal(jsonData, &unmarshaledClaims)
	require.NoError(t, err)

	assert.Equal(t, claims.ClientID, unmarshaledClaims.ClientID)
	assert.Equal(t, claims.Role, unmarshaledClaims.Role)
	assert.Equal(t, claims.Roles, unmarshaledClaims.Roles)
}

func TestJWTMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.JWTConfig{
		Secret:          "test-secret-key",
		RequiredClaims:  []string{"clientId", "role"},
		SkipExpiryCheck: false,
	}

	// Create a valid token
	claims := &JWTClaims{
		ClientID: "test-client",
		Role:     "admin",
		Roles:    []string{"admin", "user"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.Secret))

	router := gin.New()
	router.Use(JWTMiddleware(cfg))
	router.Use(RequireRole("admin"))

	router.GET("/admin", func(c *gin.Context) {
		clientID := auth.GetClientID(c)
		role := auth.GetUserRole(c)
		roles := auth.GetUserRoles(c)

		c.JSON(http.StatusOK, gin.H{
			"message":   "admin access granted",
			"client_id": clientID,
			"role":      role,
			"roles":     roles,
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "admin access granted", response["message"])
	assert.Equal(t, "test-client", response["client_id"])
	assert.Equal(t, "admin", response["role"])
	assert.Equal(t, []interface{}{"admin", "user"}, response["roles"])
}
