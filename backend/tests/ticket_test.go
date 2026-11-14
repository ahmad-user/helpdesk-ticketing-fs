package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"helpdesk-ticketing-system/backend/internal/config"
	"helpdesk-ticketing-system/backend/internal/handler"
	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"
	"helpdesk-ticketing-system/backend/internal/routes"
	"helpdesk-ticketing-system/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockTicketRepository implements repository.TicketRepository in-memory for integration tests.
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
		if filter.RequesterID != nil && t.RequesterID != *filter.RequesterID {
			continue
		}
		if filter.AssigneeID != nil {
			if t.AssigneeID == nil || *t.AssigneeID != *filter.AssigneeID {
				continue
			}
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.Priority != "" && t.Priority != filter.Priority {
			continue
		}
		if filter.Search != "" && !strings.Contains(strings.ToLower(t.Title), strings.ToLower(filter.Search)) {
			continue
		}

		cloned := *t
		result = append(result, &cloned)
	}

	total := int64(len(result))

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

// createUniqueTestUserAndToken generates a unique user to avoid email collisions.
func createUniqueTestUserAndToken(t *testing.T, authSvc service.AuthService, mockRepo *mockUserRepository, role string) (*model.User, string) {
	uniqueID := uuid.New()
	user := &model.User{
		ID:           uniqueID,
		Name:         fmt.Sprintf("User_%s_%s", role, uniqueID.String()[:8]),
		Email:        fmt.Sprintf("%s_%s@example.com", role, uniqueID.String()[:8]),
		PasswordHash: "hashed_dummy_password",
		Role:         role,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := mockRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create unique mock user: %v", err)
	}

	token, err := authSvc.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token for %s: %v", role, err)
	}

	return user, token
}

// setupTicketIntegrationRouter prepares a router with in-memory repos for ticket integration tests.
func setupTicketIntegrationRouter() (*gin.Engine, service.AuthService, service.TicketService, *mockUserRepository, *mockTicketRepository) {
	jwtSecret := "ticket_integration_secret_key_12345"
	userRepo := newMockUserRepository()
	ticketRepo := newMockTicketRepository()

	appCfg := config.AppConfig{
		Name: "helpdesk-ticket-test",
		Env:  "test",
		Port: "8080",
		Host: "0.0.0.0",
	}

	jwtCfg := config.JWTConfig{
		Secret:          jwtSecret,
		ExpirationHours: 24,
	}

	authSvc := service.NewAuthService(userRepo, jwtCfg)
	ticketSvc := service.NewTicketService(ticketRepo, userRepo)

	authHdlr := handler.NewAuthHandler(authSvc)
	ticketHdlr := handler.NewTicketHandler(ticketSvc)

	router := routes.SetupRouter(routes.RouterParams{
		AppConfig:     appCfg,
		AuthHandler:   authHdlr,
		TicketHandler: ticketHdlr,
		AuthService:   authSvc,
	})

	return router, authSvc, ticketSvc, userRepo, ticketRepo
}

// 1. Test Requester Berhasil Membuat Tiket
func TestTicket_1_RequesterCreatesTicket_Success(t *testing.T) {
	router, authSvc, _, userRepo, _ := setupTicketIntegrationRouter()

	requester, token := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)

	payload := map[string]interface{}{
		"title":       "Printer Error",
		"description": "Paper jam on second floor printer",
		"category":    "Hardware",
		"priority":    model.PriorityMedium,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/tickets", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                 `json:"success"`
		Message string               `json:"message"`
		Data    model.TicketResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Data.Title != "Printer Error" {
		t.Errorf("Expected title 'Printer Error', got '%s'", resp.Data.Title)
	}
	if resp.Data.Status != model.StatusOpen {
		t.Errorf("Expected initial status 'open', got '%s'", resp.Data.Status)
	}
	if resp.Data.RequesterID != requester.ID {
		t.Errorf("Expected RequesterID to match authenticated user %s, got %s", requester.ID, resp.Data.RequesterID)
	}
	if resp.Data.AssigneeID != nil {
		t.Errorf("Expected assignee_id to be nil, got %v", resp.Data.AssigneeID)
	}
}

// Agents are denied ticket creation by the API; admins retain requester-style creation rights.
func TestTicket_AgentCannotCreateTicket_AdminCanCreateTicket(t *testing.T) {
	router, authSvc, _, userRepo, _ := setupTicketIntegrationRouter()
	payload := map[string]interface{}{
		"title":       "Access request",
		"description": "Please review access",
		"category":    "Access",
		"priority":    model.PriorityMedium,
	}
	body, _ := json.Marshal(payload)

	for _, tc := range []struct {
		role       string
		wantStatus int
	}{
		{model.RoleAgent, http.StatusForbidden},
		{model.RoleAdmin, http.StatusCreated},
	} {
		t.Run(tc.role, func(t *testing.T) {
			_, token := createUniqueTestUserAndToken(t, authSvc, userRepo, tc.role)
			req, _ := http.NewRequest(http.MethodPost, "/api/tickets", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Fatalf("Expected HTTP %d, got %d. Body: %s", tc.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

// 2. Test Requester Tidak Bisa Membuat Tiket Atas Nama User Lain
func TestTicket_2_RequesterCannotSpoofRequesterID(t *testing.T) {
	router, authSvc, _, userRepo, _ := setupTicketIntegrationRouter()

	requester, token := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	victimID := uuid.New()

	// Malicious payload trying to inject victim's requester_id
	payload := map[string]interface{}{
		"title":        "Spoofed Ticket",
		"description":  "Attempting to impersonate someone else",
		"category":     "Security",
		"priority":     model.PriorityHigh,
		"requester_id": victimID.String(),
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/tickets", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                 `json:"success"`
		Data    model.TicketResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify requester_id was FORCED to authenticated user, ignoring spoofed ID in body
	if resp.Data.RequesterID != requester.ID {
		t.Errorf("Security violation! Spoofed requester_id was accepted: expected %s, got %s", requester.ID, resp.Data.RequesterID)
	}
}

// 3. Test Requester Tidak Bisa Melihat Tiket Orang Lain
func TestTicket_3_RequesterCannotViewOtherUserTickets(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	userA, _ := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	userB, tokenB := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)

	// Create ticket owned by User A
	ticketA := &model.Ticket{
		ID:          uuid.New(),
		Title:       "User A's Secret Ticket",
		Description: "Private information",
		Category:    "Finance",
		Priority:    model.PriorityLow,
		Status:      model.StatusOpen,
		RequesterID: userA.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticketA)

	// User B attempts to access User A's ticket detail -> 403 Forbidden
	reqDetail, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/tickets/%s", ticketA.ID), nil)
	reqDetail.Header.Set("Authorization", "Bearer "+tokenB)

	wDetail := httptest.NewRecorder()
	router.ServeHTTP(wDetail, reqDetail)

	if wDetail.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 Forbidden for User B accessing User A's ticket, got %d. Body: %s", wDetail.Code, wDetail.Body.String())
	}

	// User B lists tickets -> Should NOT contain User A's ticket
	reqList, _ := http.NewRequest(http.MethodGet, "/api/tickets", nil)
	reqList.Header.Set("Authorization", "Bearer "+tokenB)

	wList := httptest.NewRecorder()
	router.ServeHTTP(wList, reqList)

	var listResp struct {
		Success bool                   `json:"success"`
		Data    []model.TicketResponse `json:"data"`
	}
	json.Unmarshal(wList.Body.Bytes(), &listResp)

	for _, item := range listResp.Data {
		if item.RequesterID != userB.ID {
			t.Errorf("Requester isolation failed! User B saw ticket of User %s", item.RequesterID)
		}
	}
}

// 4. Test Agent Dapat Melihat Seluruh Tiket
func TestTicket_4_AgentCanViewAllTickets(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	user1, _ := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	user2, _ := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	_, agentToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAgent)

	ticket1 := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Ticket from User 1",
		Description: "Desc 1",
		Category:    "Software",
		Priority:    model.PriorityLow,
		Status:      model.StatusOpen,
		RequesterID: user1.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticket2 := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Ticket from User 2",
		Description: "Desc 2",
		Category:    "Hardware",
		Priority:    model.PriorityHigh,
		Status:      model.StatusOpen,
		RequesterID: user2.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket1)
	ticketRepo.Create(context.Background(), ticket2)

	// Agent GET /api/tickets
	req, _ := http.NewRequest(http.MethodGet, "/api/tickets", nil)
	req.Header.Set("Authorization", "Bearer "+agentToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK for Agent, got %d. Body: %s", w.Code, w.Body.String())
	}

	var listResp struct {
		Success bool                   `json:"success"`
		Data    []model.TicketResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)

	if len(listResp.Data) < 2 {
		t.Errorf("Agent should see all tickets, expected >= 2, got %d", len(listResp.Data))
	}
}

// 5. Test Admin Dapat Melihat Seluruh Tiket
func TestTicket_5_AdminCanViewAllTickets(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	user, _ := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	_, adminToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAdmin)

	t1 := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Admin View Ticket",
		Description: "Any description",
		Category:    "Access",
		Priority:    model.PriorityUrgent,
		Status:      model.StatusOpen,
		RequesterID: user.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), t1)

	// Admin GET /api/tickets/:id
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/tickets/%s", t1.ID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK for Admin viewing ticket, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// 6. Test Agent Berhasil Melakukan Assignment
func TestTicket_6_AgentAssignsTicket_Success(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, _ := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	agentUser, agentToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAgent)

	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Unassigned ticket",
		Description: "Needs triage",
		Category:    "Support",
		Priority:    model.PriorityMedium,
		Status:      model.StatusOpen,
		RequesterID: requester.ID,
		AssigneeID:  nil,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket)

	// Agent assigns ticket to themselves
	body, _ := json.Marshal(map[string]interface{}{
		"assignee_id": agentUser.ID.String(),
	})

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+agentToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK for Agent assignment, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                 `json:"success"`
		Data    model.TicketResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.AssigneeID == nil || *resp.Data.AssigneeID != agentUser.ID {
		t.Errorf("Expected assignee to be %s, got %v", agentUser.ID, resp.Data.AssigneeID)
	}
}

// 7. Test Requester Ditolak Saat Melakukan Assignment (HTTP 403 Forbidden)
func TestTicket_7_RequesterAssignment_Forbidden(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, requesterToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	agentUser, _ := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAgent)

	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "My ticket",
		Description: "I need help",
		Category:    "Support",
		Priority:    model.PriorityMedium,
		Status:      model.StatusOpen,
		RequesterID: requester.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket)

	body, _ := json.Marshal(map[string]interface{}{
		"assignee_id": agentUser.ID.String(),
	})

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+requesterToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected HTTP 403 Forbidden when requester tries assignment, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// 8. Test Validasi Priority
func TestTicket_8_PriorityValidation(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, requesterToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	_, agentToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAgent)

	// 8a. Invalid priority on creation rejected with 400 Bad Request
	badCreate, _ := json.Marshal(map[string]interface{}{
		"title":       "Bad Priority",
		"description": "testing",
		"category":    "General",
		"priority":    "super-urgent", // invalid
	})
	req1, _ := http.NewRequest(http.MethodPost, "/api/tickets", bytes.NewBuffer(badCreate))
	req1.Header.Set("Authorization", "Bearer "+requesterToken)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusBadRequest {
		t.Errorf("Expected HTTP 400 for invalid priority on create, got %d", w1.Code)
	}

	// 8b. Requester cannot change priority on update (HTTP 403)
	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Priority test ticket",
		Description: "testing",
		Category:    "General",
		Priority:    model.PriorityLow,
		Status:      model.StatusOpen,
		RequesterID: requester.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket)

	reqPriorityUpdate, _ := json.Marshal(map[string]interface{}{
		"priority": model.PriorityUrgent,
	})
	req2, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(reqPriorityUpdate))
	req2.Header.Set("Authorization", "Bearer "+requesterToken)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 when requester attempts to update priority, got %d", w2.Code)
	}

	// 8c. Agent can change priority to valid value
	req3, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(reqPriorityUpdate))
	req3.Header.Set("Authorization", "Bearer "+agentToken)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("Expected HTTP 200 when agent updates priority, got %d. Body: %s", w3.Code, w3.Body.String())
	}
}

// 9. Test Transisi Status Open -> In Progress
func TestTicket_9_Transition_Open_To_InProgress(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, requesterToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	_, agentToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAgent)

	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Transition Test",
		Description: "Status test",
		Category:    "Support",
		Priority:    model.PriorityMedium,
		Status:      model.StatusOpen,
		RequesterID: requester.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket)

	body, _ := json.Marshal(map[string]interface{}{"status": model.StatusInProgress})

	// Requester is forbidden (403)
	reqReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(body))
	reqReq.Header.Set("Authorization", "Bearer "+requesterToken)
	reqReq.Header.Set("Content-Type", "application/json")
	wReq := httptest.NewRecorder()
	router.ServeHTTP(wReq, reqReq)
	if wReq.Code != http.StatusForbidden {
		t.Errorf("Requester should be forbidden from Open -> In Progress, got %d", wReq.Code)
	}

	// Agent succeeds (200)
	reqAgent, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(body))
	reqAgent.Header.Set("Authorization", "Bearer "+agentToken)
	reqAgent.Header.Set("Content-Type", "application/json")
	wAgent := httptest.NewRecorder()
	router.ServeHTTP(wAgent, reqAgent)
	if wAgent.Code != http.StatusOK {
		t.Errorf("Agent should succeed Open -> In Progress, got %d. Body: %s", wAgent.Code, wAgent.Body.String())
	}
}

// 10. Test Transisi Status In Progress -> Resolved
func TestTicket_10_Transition_InProgress_To_Resolved(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, requesterToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	_, agentToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAgent)

	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Transition InProgress to Resolved",
		Description: "Status test",
		Category:    "Support",
		Priority:    model.PriorityMedium,
		Status:      model.StatusInProgress,
		RequesterID: requester.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket)

	body, _ := json.Marshal(map[string]interface{}{"status": model.StatusResolved})

	// Requester is forbidden (403)
	reqReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(body))
	reqReq.Header.Set("Authorization", "Bearer "+requesterToken)
	reqReq.Header.Set("Content-Type", "application/json")
	wReq := httptest.NewRecorder()
	router.ServeHTTP(wReq, reqReq)
	if wReq.Code != http.StatusForbidden {
		t.Errorf("Requester should be forbidden from In Progress -> Resolved, got %d", wReq.Code)
	}

	// Agent succeeds (200)
	reqAgent, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(body))
	reqAgent.Header.Set("Authorization", "Bearer "+agentToken)
	reqAgent.Header.Set("Content-Type", "application/json")
	wAgent := httptest.NewRecorder()
	router.ServeHTTP(wAgent, reqAgent)
	if wAgent.Code != http.StatusOK {
		t.Errorf("Agent should succeed In Progress -> Resolved, got %d. Body: %s", wAgent.Code, wAgent.Body.String())
	}
}

// 11. Test Transisi Status Resolved -> Closed (Requester, Agent, Admin boleh)
func TestTicket_11_Transition_Resolved_To_Closed(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, requesterToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)

	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Resolved Ticket",
		Description: "Issue resolved",
		Category:    "Support",
		Priority:    model.PriorityMedium,
		Status:      model.StatusResolved,
		RequesterID: requester.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket)

	body, _ := json.Marshal(map[string]interface{}{"status": model.StatusClosed})

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+requesterToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Requester should be allowed to close resolved ticket, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// 12. Test Transisi Status Resolved -> Open (Hanya Requester)
func TestTicket_12_Transition_Resolved_To_Open(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, requesterToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	_, agentToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAgent)

	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Reopen Ticket",
		Description: "Issue recurred",
		Category:    "Support",
		Priority:    model.PriorityMedium,
		Status:      model.StatusResolved,
		RequesterID: requester.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket)

	body, _ := json.Marshal(map[string]interface{}{"status": model.StatusOpen})

	// Agent is forbidden (403) from Resolved -> Open
	reqAgent, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(body))
	reqAgent.Header.Set("Authorization", "Bearer "+agentToken)
	reqAgent.Header.Set("Content-Type", "application/json")
	wAgent := httptest.NewRecorder()
	router.ServeHTTP(wAgent, reqAgent)
	if wAgent.Code != http.StatusForbidden {
		t.Errorf("Agent should be forbidden from Resolved -> Open, got %d", wAgent.Code)
	}

	// Requester succeeds (200)
	reqReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket.ID), bytes.NewBuffer(body))
	reqReq.Header.Set("Authorization", "Bearer "+requesterToken)
	reqReq.Header.Set("Content-Type", "application/json")
	wReq := httptest.NewRecorder()
	router.ServeHTTP(wReq, reqReq)
	if wReq.Code != http.StatusOK {
		t.Errorf("Requester should succeed Resolved -> Open, got %d. Body: %s", wReq.Code, wReq.Body.String())
	}
}

// 13. Test Transisi Status Open -> Closed (Requester, Admin boleh; Agent tidak boleh)
func TestTicket_13_Transition_Open_To_Closed(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, requesterToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	_, agentToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAgent)
	_, adminToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAdmin)

	ticket1 := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Direct Close 1",
		Description: "No longer needed",
		Category:    "General",
		Priority:    model.PriorityLow,
		Status:      model.StatusOpen,
		RequesterID: requester.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket1)

	body, _ := json.Marshal(map[string]interface{}{"status": model.StatusClosed})

	// 13a. Agent is forbidden (403) from Open -> Closed
	reqAgent, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket1.ID), bytes.NewBuffer(body))
	reqAgent.Header.Set("Authorization", "Bearer "+agentToken)
	reqAgent.Header.Set("Content-Type", "application/json")
	wAgent := httptest.NewRecorder()
	router.ServeHTTP(wAgent, reqAgent)
	if wAgent.Code != http.StatusForbidden {
		t.Errorf("Agent should be forbidden from Open -> Closed, got %d", wAgent.Code)
	}

	// 13b. Requester succeeds (200)
	reqReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket1.ID), bytes.NewBuffer(body))
	reqReq.Header.Set("Authorization", "Bearer "+requesterToken)
	reqReq.Header.Set("Content-Type", "application/json")
	wReq := httptest.NewRecorder()
	router.ServeHTTP(wReq, reqReq)
	if wReq.Code != http.StatusOK {
		t.Errorf("Requester should succeed Open -> Closed, got %d. Body: %s", wReq.Code, wReq.Body.String())
	}

	// 13c. Admin succeeds (200) on fresh open ticket
	ticket2 := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Direct Close 2",
		Description: "Duplicate ticket",
		Category:    "General",
		Priority:    model.PriorityLow,
		Status:      model.StatusOpen,
		RequesterID: requester.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), ticket2)

	reqAdmin, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", ticket2.ID), bytes.NewBuffer(body))
	reqAdmin.Header.Set("Authorization", "Bearer "+adminToken)
	reqAdmin.Header.Set("Content-Type", "application/json")
	wAdmin := httptest.NewRecorder()
	router.ServeHTTP(wAdmin, reqAdmin)
	if wAdmin.Code != http.StatusOK {
		t.Errorf("Admin should succeed Open -> Closed, got %d. Body: %s", wAdmin.Code, wAdmin.Body.String())
	}
}

// 14. Test Transisi Status yang Tidak Diizinkan Harus Ditolak (HTTP 400 Bad Request)
func TestTicket_14_DisallowedStatusTransitions_Rejected(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, _ := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	_, adminToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAdmin)

	// Create a closed ticket
	closedTicket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Closed Ticket",
		Description: "Already done",
		Category:    "Support",
		Priority:    model.PriorityMedium,
		Status:      model.StatusClosed,
		RequesterID: requester.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), closedTicket)

	// Attempting to reopen a closed ticket -> 400 Bad Request
	body, _ := json.Marshal(map[string]interface{}{"status": model.StatusOpen})
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/tickets/%s", closedTicket.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected HTTP 400 Bad Request for disallowed transition from closed, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// 15. Test Pagination, Filter, Search, dan Sorting Berjalan Benar
func TestTicket_15_PaginationFilterSearchSorting(t *testing.T) {
	router, authSvc, _, userRepo, ticketRepo := setupTicketIntegrationRouter()

	requester, _ := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleRequester)
	agentUser, agentToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAgent)

	// Seed multiple tickets
	t1 := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Alpha Database Connection Failure",
		Description: "DB down",
		Category:    "Database",
		Priority:    model.PriorityUrgent,
		Status:      model.StatusOpen,
		RequesterID: requester.ID,
		AssigneeID:  &agentUser.ID,
		CreatedAt:   time.Now().Add(-5 * time.Hour).UTC(),
		UpdatedAt:   time.Now().Add(-5 * time.Hour).UTC(),
	}
	t2 := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Beta Email Delivery Issue",
		Description: "Emails delayed",
		Category:    "Email",
		Priority:    model.PriorityLow,
		Status:      model.StatusResolved,
		RequesterID: requester.ID,
		AssigneeID:  nil,
		CreatedAt:   time.Now().Add(-2 * time.Hour).UTC(),
		UpdatedAt:   time.Now().Add(-2 * time.Hour).UTC(),
	}
	t3 := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Gamma Database Slow Query",
		Description: "Slow queries on report",
		Category:    "Database",
		Priority:    model.PriorityHigh,
		Status:      model.StatusOpen,
		RequesterID: requester.ID,
		AssigneeID:  &agentUser.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	ticketRepo.Create(context.Background(), t1)
	ticketRepo.Create(context.Background(), t2)
	ticketRepo.Create(context.Background(), t3)

	// 15a. Search by title keyword "Database"
	reqSearch, _ := http.NewRequest(http.MethodGet, "/api/tickets?search=database", nil)
	reqSearch.Header.Set("Authorization", "Bearer "+agentToken)
	wSearch := httptest.NewRecorder()
	router.ServeHTTP(wSearch, reqSearch)
	if wSearch.Code != http.StatusOK {
		t.Fatalf("Search failed: %d", wSearch.Code)
	}
	var searchResp struct {
		Data       []model.TicketResponse   `json:"data"`
		Pagination model.PaginationMetadata `json:"pagination"`
	}
	json.Unmarshal(wSearch.Body.Bytes(), &searchResp)
	if searchResp.Pagination.Total != 2 {
		t.Errorf("Expected 2 search results for 'database', got %d", searchResp.Pagination.Total)
	}

	// 15b. Filter by status=resolved
	reqFilter, _ := http.NewRequest(http.MethodGet, "/api/tickets?status=resolved", nil)
	reqFilter.Header.Set("Authorization", "Bearer "+agentToken)
	wFilter := httptest.NewRecorder()
	router.ServeHTTP(wFilter, reqFilter)
	var filterResp struct {
		Data       []model.TicketResponse   `json:"data"`
		Pagination model.PaginationMetadata `json:"pagination"`
	}
	json.Unmarshal(wFilter.Body.Bytes(), &filterResp)
	if filterResp.Pagination.Total != 1 {
		t.Errorf("Expected 1 resolved ticket, got %d", filterResp.Pagination.Total)
	}

	// 15c. Pagination limit=1, page=1
	reqPage, _ := http.NewRequest(http.MethodGet, "/api/tickets?limit=1&page=1", nil)
	reqPage.Header.Set("Authorization", "Bearer "+agentToken)
	wPage := httptest.NewRecorder()
	router.ServeHTTP(wPage, reqPage)
	var pageResp struct {
		Data       []model.TicketResponse   `json:"data"`
		Pagination model.PaginationMetadata `json:"pagination"`
	}
	json.Unmarshal(wPage.Body.Bytes(), &pageResp)
	if len(pageResp.Data) != 1 {
		t.Errorf("Expected page size 1, got %d", len(pageResp.Data))
	}
	if pageResp.Pagination.Total != 3 {
		t.Errorf("Expected total 3, got %d", pageResp.Pagination.Total)
	}
	if pageResp.Pagination.TotalPages != 3 {
		t.Errorf("Expected total_pages 3, got %d", pageResp.Pagination.TotalPages)
	}
}

// 16. Test DELETE /api/tickets/:id Mengembalikan HTTP 405 Method Not Allowed
func TestTicket_16_DeleteMethodNotAllowed(t *testing.T) {
	router, authSvc, _, userRepo, _ := setupTicketIntegrationRouter()

	_, adminToken := createUniqueTestUserAndToken(t, authSvc, userRepo, model.RoleAdmin)
	dummyID := uuid.New().String()

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tickets/%s", dummyID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected HTTP 405 Method Not Allowed for DELETE /api/tickets/:id, got %d. Body: %s", w.Code, w.Body.String())
	}
}
