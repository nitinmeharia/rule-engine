package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"github.com/rule-engine/internal/auth"
	"github.com/rule-engine/internal/config"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	ClientID string   `json:"clientId"`
	Role     string   `json:"role"`
	Roles    []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// JWTMiddleware validates JWT tokens and extracts user information
func JWTMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip JWT validation for health check endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Error().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Msg("Missing Authorization header")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Missing Authorization header",
				"code":    "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Error().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Msg("Invalid Authorization header format")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Invalid Authorization header format",
				"code":    "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil {
			log.Error().
				Str("error", err.Error()).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Msg("JWT token validation failed")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Invalid JWT token",
				"code":    "INVALID_JWT_TOKEN",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			log.Error().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Msg("Invalid JWT claims")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Invalid JWT claims",
				"code":    "INVALID_JWT_CLAIMS",
			})
			c.Abort()
			return
		}

		// Validate required claims
		if err := validateRequiredClaims(claims, cfg.RequiredClaims); err != nil {
			log.Error().
				Str("error", err.Error()).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Msg("Missing required JWT claims")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Missing required JWT claims",
				"code":    "MISSING_REQUIRED_CLAIMS",
			})
			c.Abort()
			return
		}

		// Check token expiration (unless skipped for development)
		if !cfg.SkipExpiryCheck {
			if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
				log.Error().
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("client_ip", c.ClientIP()).
					Str("client_id", claims.ClientID).
					Time("expires_at", claims.ExpiresAt.Time).
					Msg("JWT token expired")

				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "UNAUTHORIZED",
					"message": "JWT token expired",
					"code":    "JWT_TOKEN_EXPIRED",
				})
				c.Abort()
				return
			}
		}

		// Store user information in context
		c.Set("client_id", claims.ClientID)
		c.Set("role", claims.Role)
		c.Set("roles", claims.Roles)
		c.Set("user_id", claims.ClientID) // For backward compatibility

		// Log successful authentication
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Str("client_id", claims.ClientID).
			Str("role", claims.Role).
			Msg("JWT authentication successful")

		c.Next()
	}
}

// validateRequiredClaims checks if all required claims are present
func validateRequiredClaims(claims *JWTClaims, requiredClaims []string) error {
	for _, requiredClaim := range requiredClaims {
		switch requiredClaim {
		case "clientId":
			if claims.ClientID == "" {
				return fmt.Errorf("missing required claim: clientId")
			}
		case "role":
			if claims.Role == "" {
				return fmt.Errorf("missing required claim: role")
			}
		}
	}
	return nil
}

// RequireRole middleware checks if user has required role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := auth.GetUserRole(c)
		userRoles := auth.GetUserRoles(c)

		// Check single role
		if userRole == requiredRole {
			c.Next()
			return
		}

		// Check multiple roles
		for _, role := range userRoles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		// Check admin role (admin has access to everything)
		if userRole == "admin" || contains(userRoles, "admin") {
			c.Next()
			return
		}

		log.Error().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Str("client_id", auth.GetClientID(c)).
			Str("user_role", userRole).
			Str("required_role", requiredRole).
			Msg("Insufficient permissions")

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "FORBIDDEN",
			"message": "Insufficient permissions",
			"code":    "INSUFFICIENT_PERMISSIONS",
		})
		c.Abort()
	}
}

// RequireAnyRole middleware checks if user has any of the required roles
func RequireAnyRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := auth.GetUserRole(c)
		userRoles := auth.GetUserRoles(c)

		// Check if user has any of the required roles
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				c.Next()
				return
			}

			for _, role := range userRoles {
				if role == requiredRole {
					c.Next()
					return
				}
			}
		}

		// Check admin role (admin has access to everything)
		if userRole == "admin" || contains(userRoles, "admin") {
			c.Next()
			return
		}

		log.Error().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Str("client_id", auth.GetClientID(c)).
			Str("user_role", userRole).
			Interface("required_roles", requiredRoles).
			Msg("Insufficient permissions")

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "FORBIDDEN",
			"message": "Insufficient permissions",
			"code":    "INSUFFICIENT_PERMISSIONS",
		})
		c.Abort()
	}
}

// Helper function to check if slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
