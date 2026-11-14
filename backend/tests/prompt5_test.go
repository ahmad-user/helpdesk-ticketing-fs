package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"helpdesk-ticketing-system/backend/internal/config"
	"helpdesk-ticketing-system/backend/internal/handler"
	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/routes"
	"helpdesk-ticketing-system/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockPrompt5CommentRepository implements repository.CommentRepository in-memory for integration tests.
type mockPrompt5CommentRepository struct {
	mu       sync.RWMutex
	comments map[uuid.UUID][]*model.Comment
}

func newMockPrompt5CommentRepository() *mockPrompt5CommentRepository {
	return &mockPrompt5CommentRepository{
		comments: make(map[uuid.UUID][]*model.Comment),
	}
}

func (m *mockPrompt5CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}
	cloned := *comment
	m.comments[comment.TicketID] = append(m.comments[comment.TicketID], &cloned)
	return nil
}

func (m *mockPrompt5CommentRepository) FindByTicketID(ctx context.Context, ticketID uuid.UUID, includeInternal bool) ([]*model.Comment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, exists := m.comments[ticketID]
	if !exists {
		return []*model.Comment{}, nil
	}

	var result []*model.Comment
	for _, c := range list {
		if !includeInternal && c.IsInternal {
			continue
		}
		cloned := *c
		result = append(result, &cloned)
	}
	return result, nil
}

// setupPrompt5TestRouter creates a Gin router with all handlers and mock repositories wired for Prompt 5.
func setupPrompt5TestRouter() (*gin.Engine, service.AuthService, *mockUserRepository, *mockTicketRepository, *mockPrompt5CommentRepository) {
	jwtSecret := "test_prompt5_secret_key_1234567890_xyz"
	mockUserRepo := newMockUserRepository()
	mockTicketRepo := newMockTicketRepository()
	mockCommentRepo := newMockPrompt5CommentRepository()

	appCfg := config.AppConfig{
		Name: "helpdesk-ticketing-system-test",
		Env:  "test",
		Port: "8080",
		Host: "0.0.0.0",
	}

	jwtCfg := config.JWTConfig{
		Secret:          jwtSecret,
		ExpirationHours: 24,
	}

	authSvc := service.NewAuthService(mockUserRepo, jwtCfg)
	authHdlr := handler.NewAuthHandler(authSvc)

	ticketSvc := service.NewTicketService(mockTicketRepo, mockUserRepo)
	ticketHdlr := handler.NewTicketHandler(ticketSvc)

	commentSvc := service.NewCommentService(mockCommentRepo, mockTicketRepo, mockUserRepo)
	commentHdlr := handler.NewCommentHandler(commentSvc)

	userSvc := service.NewUserService(mockUserRepo)
	userHdlr := handler.NewUserHandler(userSvc)

	router := routes.SetupRouter(routes.RouterParams{
		AppConfig:      appCfg,
		AuthHandler:    authHdlr,
		TicketHandler:  ticketHdlr,
		CommentHandler: commentHdlr,
		UserHandler:    userHdlr,
		AuthService:    authSvc,
	})

	return router, authSvc, mockUserRepo, mockTicketRepo, mockCommentRepo
}

// Helper to seed a ticket
func seedTestTicket(mockTicketRepo *mockTicketRepository, requesterID uuid.UUID) *model.Ticket {
	ticket := &model.Ticket{
		ID:          uuid.New(),
		Title:       "Test System Issue",
		Description: "Detailed issue report description",
		Category:    "Software",
		Priority:    model.PriorityMedium,
		Status:      model.StatusOpen,
		RequesterID: requesterID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	_ = mockTicketRepo.Create(context.Background(), ticket)
	return ticket
}

// createPrompt5UserAndToken creates a test user with a guaranteed unique email and returns token.
func createPrompt5UserAndToken(t *testing.T, authSvc service.AuthService, mockRepo *mockUserRepository, role string) (*model.User, string) {
	uniqueID := uuid.New().String()
	user := &model.User{
		ID:           uuid.MustParse(uniqueID),
		Name:         fmt.Sprintf("User_%s_%s", role, uniqueID[:6]),
		Email:        fmt.Sprintf("%s_%s@example.com", role, uniqueID[:8]),
		PasswordHash: "hashed_dummy_password",
		Role:         role,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := mockRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create mock user: %v", err)
	}

	token, err := authSvc.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token for %s: %v", role, err)
	}

	return user, token
}

// 1. Requester berhasil menambahkan komentar
func TestPrompt5_1_RequesterSuccessfullyAddsComment(t *testing.T) {
	router, authSvc, mockUserRepo, mockTicketRepo, _ := setupPrompt5TestRouter()
	requester, token := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	ticket := seedTestTicket(mockTicketRepo, requester.ID)

	payload := `{"body": "This is a comment from the ticket owner"}`
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/tickets/%s/comments", ticket.ID), bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected Success=true")
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be a JSON object, got %T", resp.Data)
	}

	if dataMap["body"] != "This is a comment from the ticket owner" {
		t.Errorf("Expected body 'This is a comment from the ticket owner', got '%v'", dataMap["body"])
	}
	if dataMap["author_id"] != requester.ID.String() {
		t.Errorf("Expected author_id '%s', got '%v'", requester.ID.String(), dataMap["author_id"])
	}
	if dataMap["is_internal"] != false {
		t.Errorf("Expected is_internal=false, got %v", dataMap["is_internal"])
	}
}

// 2. Requester tidak bisa berkomentar pada tiket orang lain
func TestPrompt5_2_RequesterCannotCommentOnOtherUsersTicket(t *testing.T) {
	router, authSvc, mockUserRepo, mockTicketRepo, _ := setupPrompt5TestRouter()
	ownerRequester, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	_, otherToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	ticket := seedTestTicket(mockTicketRepo, ownerRequester.ID)

	payload := `{"body": "Attempting unauthorized comment on another user's ticket"}`
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/tickets/%s/comments", ticket.ID), bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Authorization", "Bearer "+otherToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 Forbidden, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// 3. Agent berhasil menambahkan komentar
func TestPrompt5_3_AgentSuccessfullyAddsComment(t *testing.T) {
	router, authSvc, mockUserRepo, mockTicketRepo, _ := setupPrompt5TestRouter()
	ownerRequester, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	agent, agentToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAgent)
	ticket := seedTestTicket(mockTicketRepo, ownerRequester.ID)

	payload := `{"body": "Agent public response regarding the ticket"}`
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/tickets/%s/comments", ticket.ID), bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Authorization", "Bearer "+agentToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	dataMap := resp.Data.(map[string]interface{})
	if dataMap["author_id"] != agent.ID.String() {
		t.Errorf("Expected author_id '%s', got '%v'", agent.ID.String(), dataMap["author_id"])
	}
}

// 4. Admin berhasil menambahkan komentar
func TestPrompt5_4_AdminSuccessfullyAddsComment(t *testing.T) {
	router, authSvc, mockUserRepo, mockTicketRepo, _ := setupPrompt5TestRouter()
	ownerRequester, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	admin, adminToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAdmin)
	ticket := seedTestTicket(mockTicketRepo, ownerRequester.ID)

	payload := `{"body": "Admin executive update on ticket"}`
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/tickets/%s/comments", ticket.ID), bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	dataMap := resp.Data.(map[string]interface{})
	if dataMap["author_id"] != admin.ID.String() {
		t.Errorf("Expected author_id '%s', got '%v'", admin.ID.String(), dataMap["author_id"])
	}
}

// 5. Requester tidak bisa membuat internal notes
func TestPrompt5_5_RequesterCannotCreateInternalNotes(t *testing.T) {
	router, authSvc, mockUserRepo, mockTicketRepo, mockCommentRepo := setupPrompt5TestRouter()
	requester, token := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	ticket := seedTestTicket(mockTicketRepo, requester.ID)

	payload := `{"body": "Requester attempting to create secret note", "is_internal": true}`
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/tickets/%s/comments", ticket.ID), bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 Forbidden, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify nothing was saved in repository
	comments, _ := mockCommentRepo.FindByTicketID(context.Background(), ticket.ID, true)
	if len(comments) != 0 {
		t.Errorf("Expected 0 comments stored, got %d", len(comments))
	}
}

// 6. Requester tidak bisa membaca internal notes
func TestPrompt5_6_RequesterCannotReadInternalNotes(t *testing.T) {
	router, authSvc, mockUserRepo, mockTicketRepo, mockCommentRepo := setupPrompt5TestRouter()
	requester, requesterToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	agent, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAgent)
	ticket := seedTestTicket(mockTicketRepo, requester.ID)

	// Seed 1 public comment and 1 internal note
	_ = mockCommentRepo.Create(context.Background(), &model.Comment{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		AuthorID:   requester.ID,
		Body:       "Public conversation between requester and agent",
		IsInternal: false,
		CreatedAt:  time.Now().UTC(),
	})
	_ = mockCommentRepo.Create(context.Background(), &model.Comment{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		AuthorID:   agent.ID,
		Body:       "Confidential internal investigation note",
		IsInternal: true,
		CreatedAt:  time.Now().UTC(),
	})

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/tickets/%s/comments", ticket.ID), nil)
	req.Header.Set("Authorization", "Bearer "+requesterToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	dataList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("Expected data to be a list, got %T", resp.Data)
	}

	if len(dataList) != 1 {
		t.Fatalf("Expected requester to receive exactly 1 comment, got %d", len(dataList))
	}

	commentObj := dataList[0].(map[string]interface{})
	if commentObj["is_internal"] != false {
		t.Errorf("Internal note leaked to requester!")
	}
	if commentObj["body"] != "Public conversation between requester and agent" {
		t.Errorf("Expected public comment body, got '%v'", commentObj["body"])
	}
}

// 7. Agent dapat membaca internal notes
func TestPrompt5_7_AgentCanReadInternalNotes(t *testing.T) {
	router, authSvc, mockUserRepo, mockTicketRepo, mockCommentRepo := setupPrompt5TestRouter()
	requester, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	agent, agentToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAgent)
	ticket := seedTestTicket(mockTicketRepo, requester.ID)

	_ = mockCommentRepo.Create(context.Background(), &model.Comment{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		AuthorID:   requester.ID,
		Body:       "Public comment",
		IsInternal: false,
		CreatedAt:  time.Now().UTC(),
	})
	_ = mockCommentRepo.Create(context.Background(), &model.Comment{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		AuthorID:   agent.ID,
		Body:       "Internal note for support staff",
		IsInternal: true,
		CreatedAt:  time.Now().UTC(),
	})

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/tickets/%s/comments", ticket.ID), nil)
	req.Header.Set("Authorization", "Bearer "+agentToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	dataList := resp.Data.([]interface{})
	if len(dataList) != 2 {
		t.Fatalf("Expected agent to see both public and internal comments (2 total), got %d", len(dataList))
	}

	foundInternal := false
	for _, item := range dataList {
		c := item.(map[string]interface{})
		if c["is_internal"] == true {
			foundInternal = true
			break
		}
	}
	if !foundInternal {
		t.Errorf("Agent did not receive internal note in comments list")
	}
}

// 8. Admin dapat membaca internal notes
func TestPrompt5_8_AdminCanReadInternalNotes(t *testing.T) {
	router, authSvc, mockUserRepo, mockTicketRepo, mockCommentRepo := setupPrompt5TestRouter()
	requester, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	agent, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAgent)
	_, adminToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAdmin)
	ticket := seedTestTicket(mockTicketRepo, requester.ID)

	_ = mockCommentRepo.Create(context.Background(), &model.Comment{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		AuthorID:   requester.ID,
		Body:       "Public comment",
		IsInternal: false,
		CreatedAt:  time.Now().UTC(),
	})
	_ = mockCommentRepo.Create(context.Background(), &model.Comment{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		AuthorID:   agent.ID,
		Body:       "Staff internal note",
		IsInternal: true,
		CreatedAt:  time.Now().UTC(),
	})

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/tickets/%s/comments", ticket.ID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	dataList := resp.Data.([]interface{})
	if len(dataList) != 2 {
		t.Fatalf("Expected admin to see 2 comments, got %d", len(dataList))
	}
}

// 9. Admin berhasil melihat daftar pengguna
func TestPrompt5_9_AdminSuccessfullyListsUsers(t *testing.T) {
	router, authSvc, mockUserRepo, _, _ := setupPrompt5TestRouter()
	_, adminToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAdmin)
	_, _ = createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	_, _ = createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAgent)

	req, _ := http.NewRequest(http.MethodGet, "/api/users?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if resp.Pagination == nil {
		t.Fatalf("Expected pagination metadata in response")
	}
	if resp.Pagination.Total < 3 {
		t.Errorf("Expected total >= 3, got %d", resp.Pagination.Total)
	}

	dataList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("Expected data to be a slice of users, got %T", resp.Data)
	}

	// Verify password hash is NEVER leaked in any user object
	for _, item := range dataList {
		u := item.(map[string]interface{})
		if _, exists := u["password_hash"]; exists {
			t.Errorf("Security violation: password_hash field is present in user response: %v", u)
		}
		if _, exists := u["password"]; exists {
			t.Errorf("Security violation: password field is present in user response: %v", u)
		}
	}

	// Test search filter
	reqSearch, _ := http.NewRequest(http.MethodGet, "/api/users?search=admin", nil)
	reqSearch.Header.Set("Authorization", "Bearer "+adminToken)
	wSearch := httptest.NewRecorder()
	router.ServeHTTP(wSearch, reqSearch)

	if wSearch.Code != http.StatusOK {
		t.Errorf("Expected HTTP 200 OK on user search, got %d", wSearch.Code)
	}
}

// 10. Requester ditolak saat melihat daftar pengguna
func TestPrompt5_10_RequesterRejectedFromListingUsers(t *testing.T) {
	router, authSvc, mockUserRepo, _, _ := setupPrompt5TestRouter()
	_, requesterToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)

	req, _ := http.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+requesterToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 Forbidden for requester on GET /api/users, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// 11. Agent ditolak saat melihat daftar pengguna
func TestPrompt5_11_AgentRejectedFromListingUsers(t *testing.T) {
	router, authSvc, mockUserRepo, _, _ := setupPrompt5TestRouter()
	_, agentToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAgent)

	req, _ := http.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+agentToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 Forbidden for agent on GET /api/users, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// 12. Admin berhasil mengubah role
func TestPrompt5_12_AdminSuccessfullyUpdatesRole(t *testing.T) {
	router, authSvc, mockUserRepo, _, _ := setupPrompt5TestRouter()
	_, adminToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAdmin)
	targetUser, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)

	payload := `{"role": "agent"}`
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/role", targetUser.ID), bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	dataMap := resp.Data.(map[string]interface{})
	if dataMap["role"] != model.RoleAgent {
		t.Errorf("Expected updated role 'agent', got '%v'", dataMap["role"])
	}

	// Verify in database / repo
	freshUser, err := mockUserRepo.FindByID(context.Background(), targetUser.ID)
	if err != nil {
		t.Fatalf("Failed to fetch fresh user: %v", err)
	}
	if freshUser.Role != model.RoleAgent {
		t.Errorf("Expected repo role 'agent', got '%s'", freshUser.Role)
	}
}

// 13. Requester dan Agent ditolak saat mengubah role
func TestPrompt5_13_RequesterAndAgentRejectedFromUpdatingRole(t *testing.T) {
	router, authSvc, mockUserRepo, _, _ := setupPrompt5TestRouter()
	_, requesterToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)
	_, agentToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAgent)
	targetUser, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)

	payload := `{"role": "admin"}`

	// 1. Requester attempt
	reqReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/role", targetUser.ID), bytes.NewBuffer([]byte(payload)))
	reqReq.Header.Set("Authorization", "Bearer "+requesterToken)
	reqReq.Header.Set("Content-Type", "application/json")
	wReq := httptest.NewRecorder()
	router.ServeHTTP(wReq, reqReq)

	if wReq.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 Forbidden for requester role update attempt, got %d", wReq.Code)
	}

	// 2. Agent attempt
	reqAgent, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/role", targetUser.ID), bytes.NewBuffer([]byte(payload)))
	reqAgent.Header.Set("Authorization", "Bearer "+agentToken)
	reqAgent.Header.Set("Content-Type", "application/json")
	wAgent := httptest.NewRecorder()
	router.ServeHTTP(wAgent, reqAgent)

	if wAgent.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 Forbidden for agent role update attempt, got %d", wAgent.Code)
	}
}

// 14. Role yang tidak valid ditolak
func TestPrompt5_14_InvalidRoleRejected(t *testing.T) {
	router, authSvc, mockUserRepo, _, _ := setupPrompt5TestRouter()
	_, adminToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAdmin)
	targetUser, _ := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleRequester)

	invalidRoles := []string{"superadmin", "owner", "guest", "moderator", ""}

	for _, invalidRole := range invalidRoles {
		t.Run("Role_"+invalidRole, func(t *testing.T) {
			payload := fmt.Sprintf(`{"role": "%s"}`, invalidRole)
			req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/role", targetUser.ID), bytes.NewBuffer([]byte(payload)))
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected HTTP 400 Bad Request for invalid role '%s', got %d. Body: %s", invalidRole, w.Code, w.Body.String())
			}
		})
	}
}

// 15. User ID yang tidak ditemukan ditangani dengan benar
func TestPrompt5_15_NonExistentUserIDHandledCorrectly(t *testing.T) {
	router, authSvc, mockUserRepo, _, _ := setupPrompt5TestRouter()
	_, adminToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAdmin)

	// Case A: Non-existent UUID (404 Not Found)
	nonExistentID := uuid.New().String()
	payload := `{"role": "agent"}`
	req404, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/role", nonExistentID), bytes.NewBuffer([]byte(payload)))
	req404.Header.Set("Authorization", "Bearer "+adminToken)
	req404.Header.Set("Content-Type", "application/json")

	w404 := httptest.NewRecorder()
	router.ServeHTTP(w404, req404)

	if w404.Code != http.StatusNotFound {
		t.Errorf("Expected HTTP 404 Not Found for non-existent user ID, got %d. Body: %s", w404.Code, w404.Body.String())
	}

	// Case B: Malformed non-UUID format (400 Bad Request)
	req400, _ := http.NewRequest(http.MethodPatch, "/api/users/not-a-valid-uuid/role", bytes.NewBuffer([]byte(payload)))
	req400.Header.Set("Authorization", "Bearer "+adminToken)
	req400.Header.Set("Content-Type", "application/json")

	w400 := httptest.NewRecorder()
	router.ServeHTTP(w400, req400)

	if w400.Code != http.StatusBadRequest {
		t.Errorf("Expected HTTP 400 Bad Request for malformed UUID, got %d. Body: %s", w400.Code, w400.Body.String())
	}
}

// 16. Token lama tidak mempertahankan privilege setelah role berubah
func TestPrompt5_16_OldTokenDoesNotRetainPrivilegeAfterRoleChange(t *testing.T) {
	router, authSvc, mockUserRepo, _, _ := setupPrompt5TestRouter()

	// 1. Create User 1 who is an Admin and User 2 who is also an Admin (so demoting User 1 doesn't leave 0 admins)
	demotedAdmin, demotedAdminOldToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAdmin)
	_, activeAdminToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAdmin)

	// Verify demotedAdmin can initially access admin endpoint with the token
	reqInitial, _ := http.NewRequest(http.MethodGet, "/api/users", nil)
	reqInitial.Header.Set("Authorization", "Bearer "+demotedAdminOldToken)
	wInitial := httptest.NewRecorder()
	router.ServeHTTP(wInitial, reqInitial)
	if wInitial.Code != http.StatusOK {
		t.Fatalf("Expected initial admin request to succeed, got %d. Body: %s", wInitial.Code, wInitial.Body.String())
	}

	// 2. Active admin changes demotedAdmin's role to "requester"
	patchPayload := `{"role": "requester"}`
	reqPatch, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/role", demotedAdmin.ID), bytes.NewBuffer([]byte(patchPayload)))
	reqPatch.Header.Set("Authorization", "Bearer "+activeAdminToken)
	reqPatch.Header.Set("Content-Type", "application/json")
	wPatch := httptest.NewRecorder()
	router.ServeHTTP(wPatch, reqPatch)

	if wPatch.Code != http.StatusOK {
		t.Fatalf("Failed to demote admin: status %d. Body: %s", wPatch.Code, wPatch.Body.String())
	}

	// 3. User attempts to access admin endpoint using their OLD token (whose JWT claim role is "admin")
	// The middleware should check the database, find the updated role ("requester"), and immediately return 403 Forbidden!
	reqOldToken, _ := http.NewRequest(http.MethodGet, "/api/users", nil)
	reqOldToken.Header.Set("Authorization", "Bearer "+demotedAdminOldToken)
	wOldToken := httptest.NewRecorder()
	router.ServeHTTP(wOldToken, reqOldToken)

	if wOldToken.Code != http.StatusForbidden {
		t.Errorf("CRITICAL SECURITY FLAW: Old token retained admin privileges! Expected HTTP 403 Forbidden, got %d. Body: %s", wOldToken.Code, wOldToken.Body.String())
	}
}

// Additional test: Last Admin demotion protection
func TestPrompt5_DemoteLastAdminProtected(t *testing.T) {
	router, authSvc, mockUserRepo, _, _ := setupPrompt5TestRouter()
	onlyAdmin, adminToken := createPrompt5UserAndToken(t, authSvc, mockUserRepo, model.RoleAdmin)

	// There is only 1 admin in the system. Demoting them should fail.
	payload := `{"role": "requester"}`
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/role", onlyAdmin.ID), bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected HTTP 400 Bad Request when attempting to demote last administrator, got %d. Body: %s", w.Code, w.Body.String())
	}
}
