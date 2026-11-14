package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"

	"github.com/google/uuid"
)

// Ticket service domain errors.
var (
	ErrInvalidPriority               = errors.New("invalid priority, allowed values: low, medium, high, urgent")
	ErrForbiddenTicketCreation       = errors.New("forbidden: agents cannot create tickets")
	ErrInvalidStatus                 = errors.New("invalid status, allowed values: open, in_progress, resolved, closed")
	ErrForbiddenTicketAccess         = errors.New("forbidden: you can only access your own tickets")
	ErrForbiddenStatusTransition     = errors.New("forbidden: your role is not authorized to perform this status transition")
	ErrInvalidStatusTransition       = errors.New("invalid status transition")
	ErrRequesterCannotAssign         = errors.New("forbidden: requesters are not allowed to assign tickets")
	ErrInvalidAssignee               = errors.New("assignee must be an active user with agent or admin role")
	ErrRequesterCannotChangePriority = errors.New("forbidden: requesters are not allowed to change ticket priority")
)

// TicketService defines business operations for managing tickets.
type TicketService interface {
	CreateTicket(ctx context.Context, caller model.UserIdentity, req *model.CreateTicketRequest) (*model.TicketResponse, error)
	GetTicketByID(ctx context.Context, caller model.UserIdentity, id uuid.UUID) (*model.TicketResponse, error)
	ListTickets(ctx context.Context, caller model.UserIdentity, filter model.TicketFilter) ([]model.TicketResponse, *model.PaginationMetadata, error)
	UpdateTicket(ctx context.Context, caller model.UserIdentity, id uuid.UUID, req *model.UpdateTicketRequest) (*model.TicketResponse, error)
}

type ticketService struct {
	ticketRepo repository.TicketRepository
	userRepo   repository.UserRepository
}

// NewTicketService instantiates a new TicketService.
func NewTicketService(ticketRepo repository.TicketRepository, userRepo repository.UserRepository) TicketService {
	return &ticketService{
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
	}
}

// CreateTicket validates input and creates a new ticket entity.
// Default status is 'open', requester_id is always taken from authenticated context, and assignee_id is null.
func (s *ticketService) CreateTicket(ctx context.Context, caller model.UserIdentity, req *model.CreateTicketRequest) (*model.TicketResponse, error) {
	if caller.Role != model.RoleRequester && caller.Role != model.RoleAdmin {
		return nil, ErrForbiddenTicketCreation
	}
	if !model.IsValidPriority(req.Priority) {
		return nil, ErrInvalidPriority
	}

	now := time.Now().UTC()
	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
		Status:      model.StatusOpen, // Default status open
		RequesterID: caller.ID,        // Enforced from authenticated user context
		AssigneeID:  nil,              // Default null
		CreatedAt:   now,
		UpdatedAt:   now,
		Requester: &model.UserSummary{
			ID:    caller.ID,
			Email: caller.Email,
			Role:  caller.Role,
		},
	}

	if err := s.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to save ticket: %w", err)
	}

	resp := ticket.ToResponse()
	return &resp, nil
}

// GetTicketByID retrieves a ticket by UUID, applying RBAC access restrictions.
// Requesters can only view their own tickets; Agents and Admins can view any ticket.
func (s *ticketService) GetTicketByID(ctx context.Context, caller model.UserIdentity, id uuid.UUID) (*model.TicketResponse, error) {
	ticket, err := s.ticketRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// RBAC Check: Requesters can only view tickets they created
	if caller.Role == model.RoleRequester && ticket.RequesterID != caller.ID {
		return nil, ErrForbiddenTicketAccess
	}

	resp := ticket.ToResponse()
	return &resp, nil
}

// ListTickets lists tickets with pagination, filters, search, and sorting.
// Enforces that requesters only receive their own tickets.
func (s *ticketService) ListTickets(ctx context.Context, caller model.UserIdentity, filter model.TicketFilter) ([]model.TicketResponse, *model.PaginationMetadata, error) {
	// Requesters are strictly confined to their own tickets
	if caller.Role == model.RoleRequester {
		filter.RequesterID = &caller.ID
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}

	tickets, total, err := s.ticketRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tickets: %w", err)
	}

	responses := make([]model.TicketResponse, len(tickets))
	for i, t := range tickets {
		responses[i] = t.ToResponse()
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(filter.Limit)))
	}

	pagination := &model.PaginationMetadata{
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	return responses, pagination, nil
}

// UpdateTicket validates status transition, assignment, and priority rules before applying updates.
func (s *ticketService) UpdateTicket(ctx context.Context, caller model.UserIdentity, id uuid.UUID, req *model.UpdateTicketRequest) (*model.TicketResponse, error) {
	ticket, err := s.ticketRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// RBAC Check: Requesters can only modify tickets they created
	if caller.Role == model.RoleRequester && ticket.RequesterID != caller.ID {
		return nil, ErrForbiddenTicketAccess
	}

	// 1. Priority Update Validation
	if req.Priority != nil {
		if caller.Role == model.RoleRequester {
			return nil, ErrRequesterCannotChangePriority
		}
		if !model.IsValidPriority(*req.Priority) {
			return nil, ErrInvalidPriority
		}
		ticket.Priority = *req.Priority
	}

	// 2. Assignment Validation
	if req.AssigneeID != nil {
		if caller.Role == model.RoleRequester {
			return nil, ErrRequesterCannotAssign
		}

		// Verify assignee user exists and has role agent or admin
		assigneeUser, err := s.userRepo.FindByID(ctx, *req.AssigneeID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return nil, fmt.Errorf("%w: user not found", ErrInvalidAssignee)
			}
			return nil, fmt.Errorf("failed to verify assignee: %w", err)
		}

		if assigneeUser.Role != model.RoleAgent && assigneeUser.Role != model.RoleAdmin {
			return nil, fmt.Errorf("%w: user has role '%s'", ErrInvalidAssignee, assigneeUser.Role)
		}

		ticket.AssigneeID = req.AssigneeID
		ticket.Assignee = &model.UserSummary{
			ID:    assigneeUser.ID,
			Name:  assigneeUser.Name,
			Email: assigneeUser.Email,
			Role:  assigneeUser.Role,
		}
	}

	// 3. Status Transition Validation
	if req.Status != nil {
		if !model.IsValidStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}

		if err := ValidateStatusTransition(ticket.Status, *req.Status, caller.Role); err != nil {
			return nil, err
		}

		ticket.Status = *req.Status
	}

	ticket.UpdatedAt = time.Now().UTC()

	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	resp := ticket.ToResponse()
	return &resp, nil
}

// ValidateStatusTransition enforces the strict state machine matrix from Technical Specifications.
//
// Matrix:
// - Open -> In Progress    : Agent, Admin
// - In Progress -> Resolved: Agent, Admin
// - Resolved -> Closed     : Requester, Agent, Admin
// - Resolved -> Open       : Requester
// - Open -> Closed         : Requester, Admin
//
// All other transitions are rejected as invalid transitions (400 Bad Request).
// Transitions blocked by insufficient role return ErrForbiddenStatusTransition (403 Forbidden).
func ValidateStatusTransition(currentStatus, newStatus, userRole string) error {
	if currentStatus == newStatus {
		return nil // No-op idempotent transition
	}

	switch currentStatus {
	case model.StatusOpen:
		switch newStatus {
		case model.StatusInProgress:
			if userRole == model.RoleAgent || userRole == model.RoleAdmin {
				return nil
			}
			return ErrForbiddenStatusTransition
		case model.StatusClosed:
			if userRole == model.RoleRequester || userRole == model.RoleAdmin {
				return nil
			}
			return ErrForbiddenStatusTransition
		default:
			return fmt.Errorf("%w: cannot transition from '%s' to '%s'", ErrInvalidStatusTransition, currentStatus, newStatus)
		}

	case model.StatusInProgress:
		switch newStatus {
		case model.StatusResolved:
			if userRole == model.RoleAgent || userRole == model.RoleAdmin {
				return nil
			}
			return ErrForbiddenStatusTransition
		default:
			return fmt.Errorf("%w: cannot transition from '%s' to '%s'", ErrInvalidStatusTransition, currentStatus, newStatus)
		}

	case model.StatusResolved:
		switch newStatus {
		case model.StatusClosed:
			if userRole == model.RoleRequester || userRole == model.RoleAgent || userRole == model.RoleAdmin {
				return nil
			}
			return ErrForbiddenStatusTransition
		case model.StatusOpen:
			if userRole == model.RoleRequester {
				return nil
			}
			return ErrForbiddenStatusTransition
		default:
			return fmt.Errorf("%w: cannot transition from '%s' to '%s'", ErrInvalidStatusTransition, currentStatus, newStatus)
		}

	case model.StatusClosed:
		return fmt.Errorf("%w: closed tickets cannot transition to '%s'", ErrInvalidStatusTransition, newStatus)

	default:
		return fmt.Errorf("%w: unknown current status '%s'", ErrInvalidStatusTransition, currentStatus)
	}
}
