package model

import (
	"time"

	"github.com/google/uuid"
)

// Ticket priority constants.
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// Ticket status constants.
const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"
)

// IsValidPriority validates whether the given priority is permitted.
func IsValidPriority(priority string) bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

// IsValidStatus validates whether the given status is permitted.
func IsValidStatus(status string) bool {
	switch status {
	case StatusOpen, StatusInProgress, StatusResolved, StatusClosed:
		return true
	default:
		return false
	}
}

// Ticket represents the tickets table domain entity.
type Ticket struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	RequesterID uuid.UUID  `json:"requester_id"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relational fields populated via JOINs to prevent N+1 queries
	Requester *UserSummary `json:"requester,omitempty"`
	Assignee  *UserSummary `json:"assignee,omitempty"`
}

// UserSummary is a lightweight representation of a user linked to a ticket.
type UserSummary struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

// CreateTicketRequest represents the incoming payload for creating a new ticket.
type CreateTicketRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"required,min=5"`
	Category    string `json:"category" binding:"required,min=2,max=100"`
	Priority    string `json:"priority" binding:"required"`
}

// UpdateTicketRequest represents partial updates allowed for a ticket.
type UpdateTicketRequest struct {
	Status     *string    `json:"status,omitempty"`
	Priority   *string    `json:"priority,omitempty"`
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty"`
}

// TicketFilter encapsulates filtering, searching, and pagination parameters.
type TicketFilter struct {
	RequesterID *uuid.UUID // Enforced for requester role; nil for agent/admin viewing all
	AssigneeID  *uuid.UUID // Filter by specific assignee
	Status      string     // Filter by ticket status
	Priority    string     // Filter by ticket priority
	Search      string     // Case-insensitive substring search on title
	SortBy      string     // created_at, updated_at, priority, status
	SortOrder   string     // asc, desc
	Page        int        // Default 1
	Limit       int        // Default 10, Max 100
}

// PaginationMetadata holds pagination state information.
type PaginationMetadata struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// TicketResponse represents the external API presentation of a ticket.
type TicketResponse struct {
	ID          uuid.UUID    `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Category    string       `json:"category"`
	Priority    string       `json:"priority"`
	Status      string       `json:"status"`
	RequesterID uuid.UUID    `json:"requester_id"`
	Requester   *UserSummary `json:"requester,omitempty"`
	AssigneeID  *uuid.UUID   `json:"assignee_id,omitempty"`
	Assignee    *UserSummary `json:"assignee,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ToResponse maps a domain Ticket entity to TicketResponse DTO.
func (t *Ticket) ToResponse() TicketResponse {
	return TicketResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Category:    t.Category,
		Priority:    t.Priority,
		Status:      t.Status,
		RequesterID: t.RequesterID,
		Requester:   t.Requester,
		AssigneeID:  t.AssigneeID,
		Assignee:    t.Assignee,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
