package handler

import (
	"errors"
	"net/http"

	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"
	"helpdesk-ticketing-system/backend/internal/service"
	"helpdesk-ticketing-system/backend/internal/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication and user profile endpoints.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrs := validator.FormatValidationErrors(err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", validationErrs)
		return
	}

	userResponse, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			ErrorResponse(c, http.StatusConflict, "Email address is already registered", nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to register user", nil)
		return
	}

	SuccessResponse(c, http.StatusCreated, "User registered successfully", userResponse)
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrs := validator.FormatValidationErrors(err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", validationErrs)
		return
	}

	authResponse, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password", nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to authenticate user", nil)
		return
	}

	SuccessResponse(c, http.StatusOK, "Login successful", authResponse)
}

// GetMe handles GET /api/me.
func (h *AuthHandler) GetMe(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized request", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format in token", nil)
		return
	}

	profile, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			ErrorResponse(c, http.StatusNotFound, "User not found", nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user profile", nil)
		return
	}

	SuccessResponse(c, http.StatusOK, "User profile retrieved successfully", profile)
}
