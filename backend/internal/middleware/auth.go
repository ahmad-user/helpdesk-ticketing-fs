package middleware

import (
	"errors"
	"net/http"
	"strings"

	"helpdesk-ticketing-system/backend/internal/handler"
	"helpdesk-ticketing-system/backend/internal/repository"
	"helpdesk-ticketing-system/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Gin context keys for authenticated user identity.
const (
	ContextUserIDKey    = "user_id"
	ContextUserEmailKey = "user_email"
	ContextUserRoleKey  = "user_role"
	ContextClaimsKey    = "jwt_claims"
	ContextUserKey      = "current_user"
)

// Context errors
var (
	ErrUserContextNotFound = errors.New("authenticated user context not found")
	ErrInvalidUserContext  = errors.New("invalid user context format")
)

// CurrentUser encapsulates the authenticated user's identity extracted from the JWT token.
type CurrentUser struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

// AuthMiddleware validates JWT Bearer tokens in incoming requests.
// It verifies token signature, expiration, and required claim fields.
// On success, authenticated identity is set in Gin context.
// On failure, it immediately responds with HTTP 401 Unauthorized.
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			handler.ErrorResponse(c, http.StatusUnauthorized, "Authorization header is missing", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			handler.ErrorResponse(c, http.StatusUnauthorized, "Authorization header format must be Bearer <token>", nil)
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			handler.ErrorResponse(c, http.StatusUnauthorized, "Authorization token is empty", nil)
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			handler.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired authorization token", nil)
			c.Abort()
			return
		}

		parsedID, err := uuid.Parse(claims.UserID)
		if err != nil {
			handler.ErrorResponse(c, http.StatusUnauthorized, "Invalid user identifier in token", nil)
			c.Abort()
			return
		}

		userRole := claims.Role
		userEmail := claims.Email

		// Validate latest role from database to ensure role changes take immediate effect even with active JWTs
		if profile, err := authService.GetProfile(c.Request.Context(), parsedID); err == nil {
			if profile != nil {
				userRole = profile.Role
				userEmail = profile.Email
			}
		} else {
			// Fail-closed principle: if user state cannot be verified against the database, deny request
			if errors.Is(err, repository.ErrUserNotFound) {
				handler.ErrorResponse(c, http.StatusUnauthorized, "User account no longer exists", nil)
			} else {
				handler.ErrorResponse(c, http.StatusInternalServerError, "Authentication verification failed", nil)
			}
			c.Abort()
			return
		}

		currentUser := CurrentUser{
			ID:    parsedID,
			Email: userEmail,
			Role:  userRole,
		}

		// Attach authenticated user details to Gin context
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserEmailKey, userEmail)
		c.Set(ContextUserRoleKey, userRole)
		c.Set(ContextClaimsKey, claims)
		c.Set(ContextUserKey, currentUser)

		c.Next()
	}
}

// GetCurrentUser retrieves the full CurrentUser struct from Gin context.
func GetCurrentUser(c *gin.Context) (*CurrentUser, error) {
	val, exists := c.Get(ContextUserKey)
	if !exists {
		return nil, ErrUserContextNotFound
	}
	user, ok := val.(CurrentUser)
	if !ok {
		return nil, ErrInvalidUserContext
	}
	return &user, nil
}

// GetUserID retrieves the authenticated user's UUID from Gin context.
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	if user, err := GetCurrentUser(c); err == nil {
		return user.ID, nil
	}

	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return uuid.Nil, ErrUserContextNotFound
	}
	strID, ok := val.(string)
	if !ok {
		return uuid.Nil, ErrInvalidUserContext
	}
	return uuid.Parse(strID)
}

// GetUserEmail retrieves the authenticated user's email address from Gin context.
func GetUserEmail(c *gin.Context) (string, error) {
	if user, err := GetCurrentUser(c); err == nil {
		return user.Email, nil
	}

	val, exists := c.Get(ContextUserEmailKey)
	if !exists {
		return "", ErrUserContextNotFound
	}
	email, ok := val.(string)
	if !ok {
		return "", ErrInvalidUserContext
	}
	return email, nil
}

// GetUserRole retrieves the authenticated user's role from Gin context.
func GetUserRole(c *gin.Context) (string, error) {
	if user, err := GetCurrentUser(c); err == nil {
		return user.Role, nil
	}

	val, exists := c.Get(ContextUserRoleKey)
	if !exists {
		return "", ErrUserContextNotFound
	}
	role, ok := val.(string)
	if !ok {
		return "", ErrInvalidUserContext
	}
	return role, nil
}
