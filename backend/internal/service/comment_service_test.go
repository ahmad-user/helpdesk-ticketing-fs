package service

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"

	"github.com/google/uuid"
)

// mockCommentRepo is an in-memory mock implementation of repository.CommentRepository
type mockCommentRepo struct {
	mu       sync.RWMutex
	comments map[uuid.UUID][]*model.Comment
}

func newMockCommentRepo() *mockCommentRepo {
	return &mockCommentRepo{
		comments: make(map[uuid.UUID][]*model.Comment),
	}
}

func (m *mockCommentRepo) Create(ctx context.Context, comment *model.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}
	cloned := *comment
	m.comments[comment.TicketID] = append(m.comments[comment.TicketID], &cloned)
	return nil
}

func (m *mockCommentRepo) FindByTicketID(ctx context.Context, ticketID uuid.UUID, includeInternal bool) ([]*model.Comment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list, exists := m.comments[ticketID]
	if !exists {
		return []*model.Comment{}, nil
	}
	var res []*model.Comment
	for _, c := range list {
		if !includeInternal && c.IsInternal {
			continue
		}
		cloned := *c
		res = append(res, &cloned)
	}
	return res, nil
}

func TestCommentService_CreateComment(t *testing.T) {
	commentRepo := newMockCommentRepo()
	ticketRepo := newMockTicketRepository()
	userRepo := newMockUserRepoForTicket()

	svc := NewCommentService(commentRepo, ticketRepo, userRepo)

	requesterID := uuid.New()
	otherRequesterID := uuid.New()
	agentID := uuid.New()
	adminID := uuid.New()

	// Seed ticket owned by requesterID
	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Test Ticket",
		Description: "Test Description",
		RequesterID: requesterID,
		Status:      model.StatusOpen,
		Priority:    model.PriorityMedium,
	}
	_ = ticketRepo.Create(context.Background(), ticket)

	t.Run("Empty body fails validation", func(t *testing.T) {
		caller := model.UserIdentity{ID: requesterID, Role: model.RoleRequester, Email: "req@example.com"}
		_, err := svc.CreateComment(context.Background(), caller, ticket.ID, &model.CreateCommentRequest{
			Body: "   ",
		})
		if err != ErrEmptyCommentBody {
			t.Errorf("Expected ErrEmptyCommentBody, got %v", err)
		}
	})

	t.Run("Body too long fails validation", func(t *testing.T) {
		caller := model.UserIdentity{ID: requesterID, Role: model.RoleRequester, Email: "req@example.com"}
		longBody := strings.Repeat("A", 5001)
		_, err := svc.CreateComment(context.Background(), caller, ticket.ID, &model.CreateCommentRequest{
			Body: longBody,
		})
		if err != ErrCommentTooLong {
			t.Errorf("Expected ErrCommentTooLong, got %v", err)
		}
	})

	t.Run("Ticket not found fails", func(t *testing.T) {
		caller := model.UserIdentity{ID: requesterID, Role: model.RoleRequester, Email: "req@example.com"}
		_, err := svc.CreateComment(context.Background(), caller, uuid.New(), &model.CreateCommentRequest{
			Body: "Valid comment",
		})
		if err != repository.ErrTicketNotFound {
			t.Errorf("Expected ErrTicketNotFound, got %v", err)
		}
	})

	t.Run("Requester forbidden on ticket owned by another user", func(t *testing.T) {
		caller := model.UserIdentity{ID: otherRequesterID, Role: model.RoleRequester, Email: "other@example.com"}
		_, err := svc.CreateComment(context.Background(), caller, ticket.ID, &model.CreateCommentRequest{
			Body: "Trying to comment on another ticket",
		})
		if err != ErrForbiddenTicketAccess {
			t.Errorf("Expected ErrForbiddenTicketAccess, got %v", err)
		}
	})

	t.Run("Requester forbidden from creating internal note", func(t *testing.T) {
		caller := model.UserIdentity{ID: requesterID, Role: model.RoleRequester, Email: "req@example.com"}
		isInternal := true
		_, err := svc.CreateComment(context.Background(), caller, ticket.ID, &model.CreateCommentRequest{
			Body:       "My secret note",
			IsInternal: &isInternal,
		})
		if err != ErrRequesterCannotCreateInternalNote {
			t.Errorf("Expected ErrRequesterCannotCreateInternalNote, got %v", err)
		}
	})

	t.Run("Requester successfully creates regular comment", func(t *testing.T) {
		caller := model.UserIdentity{ID: requesterID, Role: model.RoleRequester, Email: "req@example.com"}
		resp, err := svc.CreateComment(context.Background(), caller, ticket.ID, &model.CreateCommentRequest{
			Body: "Requester comment",
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if resp.Body != "Requester comment" {
			t.Errorf("Expected body 'Requester comment', got '%s'", resp.Body)
		}
		if resp.IsInternal {
			t.Errorf("Expected is_internal=false")
		}
		if resp.AuthorID != requesterID {
			t.Errorf("Expected author_id=%s, got %s", requesterID, resp.AuthorID)
		}
	})

	t.Run("Agent successfully creates regular comment and internal note", func(t *testing.T) {
		caller := model.UserIdentity{ID: agentID, Role: model.RoleAgent, Email: "agent@example.com"}

		// Regular comment
		resp1, err := svc.CreateComment(context.Background(), caller, ticket.ID, &model.CreateCommentRequest{
			Body: "Agent public reply",
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if resp1.IsInternal {
			t.Errorf("Expected is_internal=false")
		}

		// Internal note
		isInternal := true
		resp2, err := svc.CreateComment(context.Background(), caller, ticket.ID, &model.CreateCommentRequest{
			Body:       "Agent internal note",
			IsInternal: &isInternal,
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !resp2.IsInternal {
			t.Errorf("Expected is_internal=true")
		}
	})

	t.Run("Admin successfully creates regular comment and internal note", func(t *testing.T) {
		caller := model.UserIdentity{ID: adminID, Role: model.RoleAdmin, Email: "admin@example.com"}

		isInternal := true
		resp, err := svc.CreateComment(context.Background(), caller, ticket.ID, &model.CreateCommentRequest{
			Body:       "Admin internal investigation notes",
			IsInternal: &isInternal,
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !resp.IsInternal {
			t.Errorf("Expected is_internal=true")
		}
	})
}

func TestCommentService_ListComments(t *testing.T) {
	commentRepo := newMockCommentRepo()
	ticketRepo := newMockTicketRepository()
	userRepo := newMockUserRepoForTicket()

	svc := NewCommentService(commentRepo, ticketRepo, userRepo)

	requesterID := uuid.New()
	otherRequesterID := uuid.New()
	agentID := uuid.New()
	adminID := uuid.New()

	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Test Ticket",
		Description: "Test Description",
		RequesterID: requesterID,
		Status:      model.StatusOpen,
		Priority:    model.PriorityMedium,
	}
	_ = ticketRepo.Create(context.Background(), ticket)

	// Add 1 public comment and 1 internal note
	_ = commentRepo.Create(context.Background(), &model.Comment{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		AuthorID:   requesterID,
		Body:       "Public comment",
		IsInternal: false,
		CreatedAt:  time.Now().UTC(),
	})
	_ = commentRepo.Create(context.Background(), &model.Comment{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		AuthorID:   agentID,
		Body:       "Confidential internal note",
		IsInternal: true,
		CreatedAt:  time.Now().UTC(),
	})

	t.Run("Other requester cannot list comments (403)", func(t *testing.T) {
		caller := model.UserIdentity{ID: otherRequesterID, Role: model.RoleRequester, Email: "other@example.com"}
		_, err := svc.ListComments(context.Background(), caller, ticket.ID)
		if err != ErrForbiddenTicketAccess {
			t.Errorf("Expected ErrForbiddenTicketAccess, got %v", err)
		}
	})

	t.Run("Requester owner lists comments without seeing internal note", func(t *testing.T) {
		caller := model.UserIdentity{ID: requesterID, Role: model.RoleRequester, Email: "req@example.com"}
		comments, err := svc.ListComments(context.Background(), caller, ticket.ID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(comments) != 1 {
			t.Fatalf("Expected 1 public comment, got %d", len(comments))
		}
		if comments[0].IsInternal {
			t.Errorf("Requester must never see internal notes")
		}
		if comments[0].Body != "Public comment" {
			t.Errorf("Expected 'Public comment', got '%s'", comments[0].Body)
		}
	})

	t.Run("Agent sees both public comment and internal note", func(t *testing.T) {
		caller := model.UserIdentity{ID: agentID, Role: model.RoleAgent, Email: "agent@example.com"}
		comments, err := svc.ListComments(context.Background(), caller, ticket.ID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(comments) != 2 {
			t.Fatalf("Expected 2 comments (public + internal), got %d", len(comments))
		}
	})

	t.Run("Admin sees both public comment and internal note", func(t *testing.T) {
		caller := model.UserIdentity{ID: adminID, Role: model.RoleAdmin, Email: "admin@example.com"}
		comments, err := svc.ListComments(context.Background(), caller, ticket.ID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(comments) != 2 {
			t.Fatalf("Expected 2 comments (public + internal), got %d", len(comments))
		}
	})
}
