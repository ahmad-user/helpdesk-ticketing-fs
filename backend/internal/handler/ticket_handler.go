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

// TicketHandler handles HTTP requests for ticket operations.
type TicketHandler struct {
	ticketService service.TicketService
}

// NewTicketHandler constructs a new TicketHandler.
func NewTicketHandler(ticketService service.TicketService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
	}
}

// getCallerIdentity extracts the authenticated user's identity from Gin context.
func getCallerIdentity(c *gin.Context) (model.UserIdentity, error) {
	idStr := c.GetString("user_id")
	if idStr == "" {
		return model.UserIdentity{}, errors.New("unauthenticated user context")
	}

	uid, err := uuid.Parse(idStr)
	if err != nil {
		return model.UserIdentity{}, errors.New("invalid user identifier in context")
	}

	email := c.GetString("user_email")
	role := c.GetString("user_role")

	return model.UserIdentity{
		ID:    uid,
		Email: email,
		Role:  role,
	}, nil
}

// Create handles POST /api/tickets.
func (h *TicketHandler) Create(c *gin.Context) {
	caller, err := getCallerIdentity(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized request", nil)
		return
	}

	var req model.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrs := validator.FormatValidationErrors(err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", validationErrs)
		return
	}

	ticketResp, err := h.ticketService.CreateTicket(c.Request.Context(), caller, &req)
	if err != nil {
		if errors.Is(err, service.ErrForbiddenTicketCreation) {
			ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		if errors.Is(err, service.ErrInvalidPriority) {
			ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create ticket", nil)
		return
	}

	SuccessResponse(c, http.StatusCreated, "Ticket created successfully", ticketResp)
}

// GetByID handles GET /api/tickets/:id.
func (h *TicketHandler) GetByID(c *gin.Context) {
	caller, err := getCallerIdentity(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized request", nil)
		return
	}

	idStr := c.Param("id")
	ticketID, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID format", nil)
		return
	}

	ticketResp, err := h.ticketService.GetTicketByID(c.Request.Context(), caller, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Ticket not found", nil)
			return
		}
		if errors.Is(err, service.ErrForbiddenTicketAccess) {
			ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve ticket", nil)
		return
	}

	SuccessResponse(c, http.StatusOK, "Ticket details retrieved successfully", ticketResp)
}

// List handles GET /api/tickets with pagination, filtering, searching, and sorting.
func (h *TicketHandler) List(c *gin.Context) {
	caller, err := getCallerIdentity(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized request", nil)
		return
	}

	// Parse pagination query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := strings.TrimSpace(c.Query("status"))
	priority := strings.TrimSpace(c.Query("priority"))
	search := strings.TrimSpace(c.Query("search"))
	sortBy := strings.TrimSpace(c.DefaultQuery("sort_by", "created_at"))
	sortOrder := strings.TrimSpace(c.DefaultQuery("order", "desc"))

	filter := model.TicketFilter{
		Status:    status,
		Priority:  priority,
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Page:      page,
		Limit:     limit,
	}

	// Parse assignee_id filter if provided
	assigneeIDStr := strings.TrimSpace(c.Query("assignee_id"))
	if assigneeIDStr != "" {
		parsedAsgID, err := uuid.Parse(assigneeIDStr)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid assignee_id query format", nil)
			return
		}
		filter.AssigneeID = &parsedAsgID
	}

	// Validate status filter if given
	if filter.Status != "" && !model.IsValidStatus(filter.Status) {
		ErrorResponse(c, http.StatusBadRequest, "Invalid status filter, allowed: open, in_progress, resolved, closed", nil)
		return
	}

	// Validate priority filter if given
	if filter.Priority != "" && !model.IsValidPriority(filter.Priority) {
		ErrorResponse(c, http.StatusBadRequest, "Invalid priority filter, allowed: low, medium, high, urgent", nil)
		return
	}

	tickets, pagination, err := h.ticketService.ListTickets(c.Request.Context(), caller, filter)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tickets", nil)
		return
	}

	PaginatedResponse(c, http.StatusOK, "Tickets retrieved successfully", tickets, pagination)
}

// Update handles PATCH /api/tickets/:id.
func (h *TicketHandler) Update(c *gin.Context) {
	caller, err := getCallerIdentity(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized request", nil)
		return
	}

	idStr := c.Param("id")
	ticketID, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID format", nil)
		return
	}

	var req model.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrs := validator.FormatValidationErrors(err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", validationErrs)
		return
	}

	ticketResp, err := h.ticketService.UpdateTicket(c.Request.Context(), caller, ticketID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Ticket not found", nil)
			return
		}
		if errors.Is(err, service.ErrForbiddenTicketAccess) ||
			errors.Is(err, service.ErrForbiddenStatusTransition) ||
			errors.Is(err, service.ErrRequesterCannotAssign) ||
			errors.Is(err, service.ErrRequesterCannotChangePriority) {
			ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		if errors.Is(err, service.ErrInvalidStatusTransition) ||
			errors.Is(err, service.ErrInvalidPriority) ||
			errors.Is(err, service.ErrInvalidStatus) ||
			errors.Is(err, service.ErrInvalidAssignee) {
			ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update ticket", nil)
		return
	}

	SuccessResponse(c, http.StatusOK, "Ticket updated successfully", ticketResp)
}

// Delete handles DELETE /api/tickets/:id (HTTP 405 Method Not Allowed per security guidelines).
func (h *TicketHandler) Delete(c *gin.Context) {
	c.Header("Allow", "GET, POST, PATCH")
	ErrorResponse(c, http.StatusMethodNotAllowed, "Ticket deletion is not allowed", nil)
}
