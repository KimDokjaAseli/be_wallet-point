package middleware

import (
	"log"
	"net/http"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware checks if user has required role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User role not found in context", nil)
			c.Abort()
			return
		}

		userRole := role.(string)
		log.Printf("DEBUG ROLE: '%s' checking against %v", userRole, allowedRoles)

		// Check if user role is in allowed roles
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", nil)
		c.Abort()
	}
}
