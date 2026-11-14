package handler

import (
	"errors"
	"net/http"
	"strings"

	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"
	"helpdesk-ticketing-system/backend/internal/service"
	"helpdesk-ticketing-system/backend/internal/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CommentHandler handles HTTP endpoints for ticket comments and internal notes.
type CommentHandler struct {
	commentService service.CommentService
}

// NewCommentHandler constructs a new CommentHandler instance.
func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// Create handles POST /api/tickets/:id/comments
func (h *CommentHandler) Create(c *gin.Context) {
	caller, err := getCallerIdentity(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized request", nil)
		return
	}

	ticketIDStr := c.Param("id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID format", nil)
		return
	}

	var req model.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrs := validator.FormatValidationErrors(err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", validationErrs)
		return
	}

	if strings.TrimSpace(req.Body) == "" {
		ErrorResponse(c, http.StatusBadRequest, "Comment body cannot be empty", nil)
		return
	}

	resp, err := h.commentService.CreateComment(c.Request.Context(), caller, ticketID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Ticket not found", nil)
			return
		}
		if errors.Is(err, service.ErrForbiddenTicketAccess) || errors.Is(err, service.ErrRequesterCannotCreateInternalNote) {
			ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		if errors.Is(err, service.ErrEmptyCommentBody) || errors.Is(err, service.ErrCommentTooLong) {
			ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create comment", nil)
		return
	}

	SuccessResponse(c, http.StatusCreated, "Comment added successfully", resp)
}

// List handles GET /api/tickets/:id/comments
func (h *CommentHandler) List(c *gin.Context) {
	caller, err := getCallerIdentity(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized request", nil)
		return
	}

	ticketIDStr := c.Param("id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID format", nil)
		return
	}

	comments, err := h.commentService.ListComments(c.Request.Context(), caller, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Ticket not found", nil)
			return
		}
		if errors.Is(err, service.ErrForbiddenTicketAccess) {
			ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve comments", nil)
		return
	}

	SuccessResponse(c, http.StatusOK, "Comments retrieved successfully", comments)
}
