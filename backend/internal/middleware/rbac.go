package middleware

import (
	"net/http"

	"helpdesk-ticketing-system/backend/internal/handler"
	"helpdesk-ticketing-system/backend/internal/model"

	"github.com/gin-gonic/gin"
)

// HasRole checks whether userRole satisfies any of the required roles.
//
// Role hierarchy rule:
// - "requester" can only access features permitted for requester.
// - "agent" can access agent features.
// - "admin" possesses all agent privileges plus administrative features.
//
// Role verification is strictly decoupled from request body or query parameters
// and only checks the authenticated identity extracted by AuthMiddleware.
func HasRole(userRole string, allowedRoles ...string) bool {
	for _, allowed := range allowedRoles {
		if userRole == allowed {
			return true
		}
		// Admin inherits all agent privileges
		if allowed == model.RoleAgent && userRole == model.RoleAdmin {
			return true
		}
	}
	return false
}

// RequireRole creates a Gin middleware that enforces Role-Based Access Control.
// It verifies that the authenticated user possesses one of the allowed roles.
// - If user identity context is missing, HTTP 401 Unauthorized is returned.
// - If user's role does not satisfy the requirements, HTTP 403 Forbidden is returned.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := GetUserRole(c)
		if err != nil || role == "" {
			handler.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized: authentication required", nil)
			c.Abort()
			return
		}

		if !HasRole(role, allowedRoles...) {
			handler.ErrorResponse(c, http.StatusForbidden, "Forbidden: insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a convenience middleware enforcing admin-only access.
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleAdmin)
}

// RequireAgent is a convenience middleware enforcing agent (or admin) access.
func RequireAgent() gin.HandlerFunc {
	return RequireRole(model.RoleAgent)
}
