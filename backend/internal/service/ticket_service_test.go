package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"

	"github.com/google/uuid"
)

// mockTicketRepository implements repository.TicketRepository in-memory for testing.
type mockTicketRepository struct {
	mu      sync.RWMutex
	tickets map[uuid.UUID]*model.Ticket
}

func newMockTicketRepository() *mockTicketRepository {
	return &mockTicketRepository{
		tickets: make(map[uuid.UUID]*model.Ticket),
	}
}

func (m *mockTicketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ticket.ID == uuid.Nil {
		ticket.ID = uuid.New()
	}
	cloned := *ticket
	m.tickets[ticket.ID] = &cloned
	return nil
}

func (m *mockTicketRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, exists := m.tickets[id]
	if !exists {
		return nil, repository.ErrTicketNotFound
	}
	cloned := *t
	return &cloned, nil
}

func (m *mockTicketRepository) FindAll(ctx context.Context, filter model.TicketFilter) ([]*model.Ticket, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.Ticket
	for _, t := range m.tickets {
		// Filter by requester
		if filter.RequesterID != nil && t.RequesterID != *filter.RequesterID {
			continue
		}
		// Filter by assignee
		if filter.AssigneeID != nil {
			if t.AssigneeID == nil || *t.AssigneeID != *filter.AssigneeID {
				continue
			}
		}
		// Filter by status
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		// Filter by priority
		if filter.Priority != "" && t.Priority != filter.Priority {
			continue
		}
		// Search on title
		if filter.Search != "" && !strings.Contains(strings.ToLower(t.Title), strings.ToLower(filter.Search)) {
			continue
		}

		cloned := *t
		result = append(result, &cloned)
	}

	total := int64(len(result))

	// Simple pagination slice
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	}

	start := (page - 1) * limit
	if start >= len(result) {
		return []*model.Ticket{}, total, nil
	}
	end := start + limit
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], total, nil
}

func (m *mockTicketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tickets[ticket.ID]; !exists {
		return repository.ErrTicketNotFound
	}
	cloned := *ticket
	m.tickets[ticket.ID] = &cloned
	return nil
}

// mockUserRepo implements repository.UserRepository for testing.
type mockUserRepoForTicket struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*model.User
}

func newMockUserRepoForTicket() *mockUserRepoForTicket {
	return &mockUserRepoForTicket{
		users: make(map[uuid.UUID]*model.User),
	}
}

func (m *mockUserRepoForTicket) Create(ctx context.Context, user *model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepoForTicket) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepoForTicket) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, exists := m.users[id]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepoForTicket) FindAll(ctx context.Context, filter model.UserFilter) ([]*model.User, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*model.User
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, int64(len(result)), nil
}

func (m *mockUserRepoForTicket) UpdateUserRoleTx(ctx context.Context, id uuid.UUID, newRole string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, exists := m.users[id]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	if u.Role == model.RoleAdmin && newRole != model.RoleAdmin {
		var count int64
		for _, user := range m.users {
			if user.Role == model.RoleAdmin {
				count++
			}
		}
		if count <= 1 {
			return nil, repository.ErrCannotDemoteLastAdmin
		}
	}
	u.Role = newRole
	u.UpdatedAt = time.Now().UTC()
	cloned := *u
	return &cloned, nil
}

func (m *mockUserRepoForTicket) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
	_, err := m.UpdateUserRoleTx(ctx, id, role)
	return err
}

func (m *mockUserRepoForTicket) CountByRole(ctx context.Context, role string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, u := range m.users {
		if u.Role == role {
			count++
		}
	}
	return count, nil
}

func (m *mockUserRepoForTicket) UpdateUserTx(ctx context.Context, user *model.User) (*model.User, error) {
	return user, nil
}
func (m *mockUserRepoForTicket) DeleteUserTx(ctx context.Context, actorID, id uuid.UUID) error {
	return nil
}

// Unit test: ValidateStatusTransition matrix
func TestValidateStatusTransition_Matrix(t *testing.T) {
	tests := []struct {
		name          string
		from          string
		to            string
		role          string
		expectAllowed bool
		expectErr     error
	}{
		// 1. Open -> In Progress
		{"Open -> In Progress (Agent)", model.StatusOpen, model.StatusInProgress, model.RoleAgent, true, nil},
		{"Open -> In Progress (Admin)", model.StatusOpen, model.StatusInProgress, model.RoleAdmin, true, nil},
		{"Open -> In Progress (Requester forbidden)", model.StatusOpen, model.StatusInProgress, model.RoleRequester, false, ErrForbiddenStatusTransition},

		// 2. In Progress -> Resolved
		{"In Progress -> Resolved (Agent)", model.StatusInProgress, model.StatusResolved, model.RoleAgent, true, nil},
		{"In Progress -> Resolved (Admin)", model.StatusInProgress, model.StatusResolved, model.RoleAdmin, true, nil},
		{"In Progress -> Resolved (Requester forbidden)", model.StatusInProgress, model.StatusResolved, model.RoleRequester, false, ErrForbiddenStatusTransition},

		// 3. Resolved -> Closed
		{"Resolved -> Closed (Requester)", model.StatusResolved, model.StatusClosed, model.RoleRequester, true, nil},
		{"Resolved -> Closed (Agent)", model.StatusResolved, model.StatusClosed, model.RoleAgent, true, nil},
		{"Resolved -> Closed (Admin)", model.StatusResolved, model.StatusClosed, model.RoleAdmin, true, nil},

		// 4. Resolved -> Open
		{"Resolved -> Open (Requester)", model.StatusResolved, model.StatusOpen, model.RoleRequester, true, nil},
		{"Resolved -> Open (Agent forbidden)", model.StatusResolved, model.StatusOpen, model.RoleAgent, false, ErrForbiddenStatusTransition},
		{"Resolved -> Open (Admin forbidden)", model.StatusResolved, model.StatusOpen, model.RoleAdmin, false, ErrForbiddenStatusTransition},

		// 5. Open -> Closed
		{"Open -> Closed (Requester)", model.StatusOpen, model.StatusClosed, model.RoleRequester, true, nil},
		{"Open -> Closed (Admin)", model.StatusOpen, model.StatusClosed, model.RoleAdmin, true, nil},
		{"Open -> Closed (Agent forbidden)", model.StatusOpen, model.StatusClosed, model.RoleAgent, false, ErrForbiddenStatusTransition},

		// 6. Invalid transitions
		{"In Progress -> Open (Disallowed transition)", model.StatusInProgress, model.StatusOpen, model.RoleAgent, false, ErrInvalidStatusTransition},
		{"Closed -> Open (Disallowed transition)", model.StatusClosed, model.StatusOpen, model.RoleAdmin, false, ErrInvalidStatusTransition},
		{"Closed -> In Progress (Disallowed transition)", model.StatusClosed, model.StatusInProgress, model.RoleAgent, false, ErrInvalidStatusTransition},
		{"Closed -> Resolved (Disallowed transition)", model.StatusClosed, model.StatusResolved, model.RoleAgent, false, ErrInvalidStatusTransition},
		{"Resolved -> In Progress (Disallowed transition)", model.StatusResolved, model.StatusInProgress, model.RoleAgent, false, ErrInvalidStatusTransition},
		{"Open -> Resolved (Disallowed transition)", model.StatusOpen, model.StatusResolved, model.RoleAgent, false, ErrInvalidStatusTransition},
		{"Same status Open->Open (No-op idempotent)", model.StatusOpen, model.StatusOpen, model.RoleRequester, true, nil},
		{"Same status Closed->Closed (No-op idempotent)", model.StatusClosed, model.StatusClosed, model.RoleAdmin, true, nil},
		{"Invalid from status", "unknown", model.StatusOpen, model.RoleAdmin, false, ErrInvalidStatusTransition},
		{"Invalid to status", model.StatusOpen, "unknown", model.RoleAdmin, false, ErrInvalidStatusTransition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatusTransition(tt.from, tt.to, tt.role)
			if tt.expectAllowed {
				if err != nil {
					t.Errorf("Expected allowed transition, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected transition error, got nil")
				}
				if tt.expectErr != nil && !errors.Is(err, tt.expectErr) {
					t.Errorf("Expected error wrapping %v, got %v", tt.expectErr, err)
				}
			}
		})
	}
}

// Unit test: TicketService Create, Update, Get, List
func TestTicketService_Operations(t *testing.T) {
	ticketRepo := newMockTicketRepository()
	userRepo := newMockUserRepoForTicket()
	svc := NewTicketService(ticketRepo, userRepo)

	ctx := context.Background()

	requesterID := uuid.New()
	requesterCaller := model.UserIdentity{
		ID:    requesterID,
		Email: "requester@example.com",
		Role:  model.RoleRequester,
	}

	agentID := uuid.New()
	agentUser := &model.User{
		ID:    agentID,
		Name:  "Agent Alice",
		Email: "agent@example.com",
		Role:  model.RoleAgent,
	}
	userRepo.Create(ctx, agentUser)
	agentCaller := model.UserIdentity{
		ID:    agentID,
		Email: "agent@example.com",
		Role:  model.RoleAgent,
	}

	// 1. Create Ticket
	createReq := &model.CreateTicketRequest{
		Title:       "Network issue",
		Description: "Cannot connect to VPN",
		Category:    "Network",
		Priority:    model.PriorityHigh,
	}
	ticketResp, err := svc.CreateTicket(ctx, requesterCaller, createReq)
	if err != nil {
		t.Fatalf("CreateTicket failed: %v", err)
	}
	if ticketResp.RequesterID != requesterID {
		t.Errorf("Expected requester ID %s, got %s", requesterID, ticketResp.RequesterID)
	}
	if ticketResp.Status != model.StatusOpen {
		t.Errorf("Expected default status 'open', got '%s'", ticketResp.Status)
	}
	if ticketResp.AssigneeID != nil {
		t.Errorf("Expected nil assignee_id on creation")
	}

	// 2. Requester cannot change priority
	newPriority := model.PriorityUrgent
	_, err = svc.UpdateTicket(ctx, requesterCaller, ticketResp.ID, &model.UpdateTicketRequest{
		Priority: &newPriority,
	})
	if !errors.Is(err, ErrRequesterCannotChangePriority) {
		t.Errorf("Expected ErrRequesterCannotChangePriority, got %v", err)
	}

	// 3. Requester cannot assign
	_, err = svc.UpdateTicket(ctx, requesterCaller, ticketResp.ID, &model.UpdateTicketRequest{
		AssigneeID: &agentID,
	})
	if !errors.Is(err, ErrRequesterCannotAssign) {
		t.Errorf("Expected ErrRequesterCannotAssign, got %v", err)
	}

	// 4. Agent can change priority and assign
	updatedTicket, err := svc.UpdateTicket(ctx, agentCaller, ticketResp.ID, &model.UpdateTicketRequest{
		Priority:   &newPriority,
		AssigneeID: &agentID,
	})
	if err != nil {
		t.Fatalf("Agent update failed: %v", err)
	}
	if updatedTicket.Priority != model.PriorityUrgent {
		t.Errorf("Expected updated priority 'urgent', got '%s'", updatedTicket.Priority)
	}
	if updatedTicket.AssigneeID == nil || *updatedTicket.AssigneeID != agentID {
		t.Errorf("Expected assignee %s, got %v", agentID, updatedTicket.AssigneeID)
	}

	// 5. Another requester cannot view this ticket
	otherRequester := model.UserIdentity{
		ID:    uuid.New(),
		Email: "other@example.com",
		Role:  model.RoleRequester,
	}
	_, err = svc.GetTicketByID(ctx, otherRequester, ticketResp.ID)
	if !errors.Is(err, ErrForbiddenTicketAccess) {
		t.Errorf("Expected ErrForbiddenTicketAccess for other requester, got %v", err)
	}

	// 6. Agent can view this ticket
	retrievedTicket, err := svc.GetTicketByID(ctx, agentCaller, ticketResp.ID)
	if err != nil {
		t.Errorf("Agent should be able to view any ticket, got err: %v", err)
	}
	if retrievedTicket.ID != ticketResp.ID {
		t.Errorf("Expected ticket ID %s, got %s", ticketResp.ID, retrievedTicket.ID)
	}
}
