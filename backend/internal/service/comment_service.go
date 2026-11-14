package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"

	"github.com/google/uuid"
)

// Comment domain errors.
var (
	ErrEmptyCommentBody                  = errors.New("comment body cannot be empty")
	ErrCommentTooLong                    = errors.New("comment body exceeds maximum allowed length of 5000 characters")
	ErrRequesterCannotCreateInternalNote = errors.New("forbidden: requesters are not allowed to create internal notes")
)

// CommentService defines business operations for managing ticket comments and internal notes.
type CommentService interface {
	CreateComment(ctx context.Context, caller model.UserIdentity, ticketID uuid.UUID, req *model.CreateCommentRequest) (*model.CommentResponse, error)
	ListComments(ctx context.Context, caller model.UserIdentity, ticketID uuid.UUID) ([]model.CommentResponse, error)
}

type commentService struct {
	commentRepo repository.CommentRepository
	ticketRepo  repository.TicketRepository
	userRepo    repository.UserRepository
}

// NewCommentService constructs a new CommentService instance.
func NewCommentService(
	commentRepo repository.CommentRepository,
	ticketRepo repository.TicketRepository,
	userRepo repository.UserRepository,
) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		ticketRepo:  ticketRepo,
		userRepo:    userRepo,
	}
}

// CreateComment validates permissions and input, then persists a new comment or internal note.
func (s *commentService) CreateComment(ctx context.Context, caller model.UserIdentity, ticketID uuid.UUID, req *model.CreateCommentRequest) (*model.CommentResponse, error) {
	trimmedBody := strings.TrimSpace(req.Body)
	if trimmedBody == "" {
		return nil, ErrEmptyCommentBody
	}
	if len(trimmedBody) > 5000 {
		return nil, ErrCommentTooLong
	}

	// 1. Verify ticket exists
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// 2. RBAC check: Requesters can only comment on tickets they created
	if caller.Role == model.RoleRequester && ticket.RequesterID != caller.ID {
		return nil, ErrForbiddenTicketAccess
	}

	// 3. Internal note restriction: Requester cannot create internal notes
	isInternal := false
	if req.IsInternal != nil && *req.IsInternal {
		if caller.Role == model.RoleRequester {
			return nil, ErrRequesterCannotCreateInternalNote
		}
		isInternal = true
	}

	// 4. Resolve author details to enrich response
	var authorSummary *model.UserSummary
	if s.userRepo != nil {
		authorUser, err := s.userRepo.FindByID(ctx, caller.ID)
		if err == nil && authorUser != nil {
			authorSummary = &model.UserSummary{
				ID:    authorUser.ID,
				Name:  authorUser.Name,
				Email: authorUser.Email,
				Role:  authorUser.Role,
			}
		}
	}
	if authorSummary == nil {
		authorSummary = &model.UserSummary{
			ID:    caller.ID,
			Email: caller.Email,
			Role:  caller.Role,
		}
	}

	now := time.Now().UTC()
	comment := &model.Comment{
		ID:         uuid.New(),
		TicketID:   ticketID,
		AuthorID:   caller.ID, // Always enforced from authenticated context
		Body:       trimmedBody,
		IsInternal: isInternal,
		CreatedAt:  now,
		Author:     authorSummary,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	resp := comment.ToResponse()
	return &resp, nil
}

// ListComments retrieves all comments for a ticket according to RBAC rules.
// Internal notes are filtered at the repository/database layer for requesters.
func (s *commentService) ListComments(ctx context.Context, caller model.UserIdentity, ticketID uuid.UUID) ([]model.CommentResponse, error) {
	// 1. Verify ticket exists
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// 2. RBAC check: Requesters can only view comments on their own tickets
	if caller.Role == model.RoleRequester && ticket.RequesterID != caller.ID {
		return nil, ErrForbiddenTicketAccess
	}

	// 3. Determine if internal notes should be included
	includeInternal := caller.Role != model.RoleRequester

	comments, err := s.commentRepo.FindByTicketID(ctx, ticketID, includeInternal)
	if err != nil {
		return nil, err
	}

	resps := make([]model.CommentResponse, 0, len(comments))
	for _, c := range comments {
		resps = append(resps, c.ToResponse())
	}

	return resps, nil
}
