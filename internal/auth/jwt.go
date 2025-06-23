package auth

import (
	"github.com/gin-gonic/gin"
)

// GetClientID retrieves client ID from context
func GetClientID(c *gin.Context) string {
	if clientID, exists := c.Get("client_id"); exists {
		if id, ok := clientID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUserRole retrieves user role from context
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return ""
}

// GetUserRoles retrieves user roles from context
func GetUserRoles(c *gin.Context) []string {
	if roles, exists := c.Get("roles"); exists {
		if r, ok := roles.([]string); ok {
			return r
		}
	}
	return []string{}
}
