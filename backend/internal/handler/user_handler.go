package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"
	"helpdesk-ticketing-system/backend/internal/service"
	"helpdesk-ticketing-system/backend/internal/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles HTTP endpoints for administrative user management.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler constructs a new UserHandler instance.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// List handles GET /api/users
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := strings.TrimSpace(c.Query("search"))

	filter := model.UserFilter{
		Search: search,
		Page:   page,
		Limit:  limit,
	}

	users, pagination, err := h.userService.ListUsers(c.Request.Context(), filter)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users", nil)
		return
	}

	PaginatedResponse(c, http.StatusOK, "Users retrieved successfully", users, pagination)
}

// UpdateRole handles PATCH /api/users/:id/role
func (h *UserHandler) UpdateRole(c *gin.Context) {
	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	var req model.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrs := validator.FormatValidationErrors(err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", validationErrs)
		return
	}

	req.Role = strings.TrimSpace(strings.ToLower(req.Role))
	if !model.IsValidRole(req.Role) {
		ErrorResponse(c, http.StatusBadRequest, "Invalid role, allowed: requester, agent, admin", nil)
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))

	updatedUser, err := h.userService.UpdateUserRole(c.Request.Context(), adminID, targetID, req.Role)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			ErrorResponse(c, http.StatusNotFound, "User not found", nil)
			return
		}
		if errors.Is(err, service.ErrInvalidRole) || errors.Is(err, service.ErrCannotDemoteLastAdmin) {
			ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update user role", nil)
		return
	}

	SuccessResponse(c, http.StatusOK, "User role updated successfully", updatedUser)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", validator.FormatValidationErrors(err))
		return
	}
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	u, err := h.userService.CreateUser(c.Request.Context(), req)
	if err != nil {
		h.respondMutationError(c, err, "Failed to create user")
		return
	}
	SuccessResponse(c, http.StatusCreated, "User created successfully", u)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}
	var req model.UpdateUserRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", validator.FormatValidationErrors(err))
		return
	}
	u, err := h.userService.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		h.respondMutationError(c, err, "Failed to update user")
		return
	}
	SuccessResponse(c, http.StatusOK, "User updated successfully", u)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}
	actor, _ := uuid.Parse(c.GetString("user_id"))
	if err = h.userService.DeleteUser(c.Request.Context(), actor, id); err != nil {
		h.respondMutationError(c, err, "Failed to delete user")
		return
	}
	SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}

func (h *UserHandler) respondMutationError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		ErrorResponse(c, http.StatusNotFound, "User not found", nil)
	case errors.Is(err, repository.ErrEmailAlreadyExists):
		ErrorResponse(c, http.StatusConflict, "Email address is already registered", nil)
	case errors.Is(err, service.ErrInvalidRole):
		ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, service.ErrCannotDemoteLastAdmin):
		ErrorResponse(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, repository.ErrCannotDeleteSelf):
		ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, repository.ErrUserHasHistory):
		ErrorResponse(c, http.StatusConflict, err.Error(), nil)
	default:
		ErrorResponse(c, http.StatusInternalServerError, fallback, nil)
	}
}
