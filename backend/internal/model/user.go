package model

import (
	"time"

	"github.com/google/uuid"
)

// User role constants.
const (
	RoleRequester = "requester"
	RoleAgent     = "agent"
	RoleAdmin     = "admin"
)

// IsValidRole checks if the provided role string is one of the recognized user roles.
func IsValidRole(role string) bool {
	switch role {
	case RoleRequester, RoleAgent, RoleAdmin:
		return true
	default:
		return false
	}
}

// UserIdentity represents minimal verified user context details.
type UserIdentity struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

// User represents the users table domain entity.
type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest holds registration payload sent by clients.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=255"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Role     string `json:"role,omitempty"` // Public registration will always enforce "requester"
}

// LoginRequest holds credentials sent during user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse is the safe public representation of a User without sensitive hashes.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthResponse represents the payload returned after successful login.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// ToResponse converts a User entity into a safe UserResponse DTO.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// UserFilter holds filtering, search, and pagination parameters for user management.
type UserFilter struct {
	Search string // Case-insensitive substring match on name or email
	Page   int    // Default 1
	Limit  int    // Default 10, Max 100
}

// UpdateUserRoleRequest represents the payload for changing a user's role.
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=255"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Role     string `json:"role" binding:"required"`
}

type UpdateUserRequest struct {
	Name  string  `json:"name" binding:"required,min=2,max=255"`
	Email string  `json:"email" binding:"required,email,max=255"`
	Role  *string `json:"role,omitempty"`
}
