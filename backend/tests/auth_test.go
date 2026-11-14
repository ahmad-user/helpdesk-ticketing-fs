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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// mockUserRepository implements an in-memory thread-safe repository.UserRepository.
type mockUserRepository struct {
	mu    sync.RWMutex
	users map[string]*model.User // keyed by email
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[strings.ToLower(user.Email)]; exists {
		return repository.ErrEmailAlreadyExists
	}

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	cloned := *user
	m.users[strings.ToLower(user.Email)] = &cloned
	return nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[strings.ToLower(email)]
	if !exists {
		return nil, repository.ErrUserNotFound
	}

	cloned := *user
	return &cloned, nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.ID == id {
			cloned := *user
			return &cloned, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepository) FindAll(ctx context.Context, filter model.UserFilter) ([]*model.User, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*model.User
	for _, user := range m.users {
		if filter.Search != "" {
			s := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(user.Name), s) && !strings.Contains(strings.ToLower(user.Email), s) {
				continue
			}
		}
		cloned := *user
		matched = append(matched, &cloned)
	}

	total := int64(len(matched))

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	start := (page - 1) * limit
	if start >= len(matched) {
		return []*model.User{}, total, nil
	}
	end := start + limit
	if end > len(matched) {
		end = len(matched)
	}

	return matched[start:end], total, nil
}

func (m *mockUserRepository) UpdateUserRoleTx(ctx context.Context, id uuid.UUID, newRole string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var targetUser *model.User
	var targetEmail string
	for email, user := range m.users {
		if user.ID == id {
			targetUser = user
			targetEmail = email
			break
		}
	}
	if targetUser == nil {
		return nil, repository.ErrUserNotFound
	}

	if targetUser.Role == model.RoleAdmin && newRole != model.RoleAdmin {
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

	targetUser.Role = newRole
	targetUser.UpdatedAt = time.Now().UTC()
	m.users[targetEmail] = targetUser
	cloned := *targetUser
	return &cloned, nil
}

func (m *mockUserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
	_, err := m.UpdateUserRoleTx(ctx, id, role)
	return err
}

func (m *mockUserRepository) CountByRole(ctx context.Context, role string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, user := range m.users {
		if user.Role == role {
			count++
		}
	}
	return count, nil
}

func (m *mockUserRepository) UpdateUserTx(ctx context.Context, user *model.User) (*model.User, error) {
	return user, nil
}
func (m *mockUserRepository) DeleteUserTx(ctx context.Context, actorID, id uuid.UUID) error {
	return nil
}

// setupAuthTestRouter initializes a test router with in-memory repo and auth services.
func setupAuthTestRouter(mockRepo *mockUserRepository, jwtExpiryHours int) (*gin.Engine, service.AuthService) {
	appCfg := config.AppConfig{
		Name: "helpdesk-ticketing-system",
		Env:  "test",
		Port: "8080",
		Host: "0.0.0.0",
	}

	jwtCfg := config.JWTConfig{
		Secret:          "test_secret_key_1234567890_abcdef",
		ExpirationHours: jwtExpiryHours,
	}

	authSvc := service.NewAuthService(mockRepo, jwtCfg)
	authHdlr := handler.NewAuthHandler(authSvc)

	r := routes.SetupRouter(routes.RouterParams{
		AppConfig:   appCfg,
		AuthHandler: authHdlr,
		UserHandler: handler.NewUserHandler(service.NewUserService(mockRepo)),
		AuthService: authSvc,
	})

	return r, authSvc
}

// 1. Test Register Berhasil & Password Tersimpan dalam Bentuk Hash
func TestRegister_Success_And_PasswordHashed(t *testing.T) {
	mockRepo := newMockUserRepository()
	router, _ := setupAuthTestRouter(mockRepo, 24)

	payload := map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "Password123!",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool               `json:"success"`
		Message string             `json:"message"`
		Data    model.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected Success=true")
	}
	if resp.Data.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", resp.Data.Email)
	}
	if resp.Data.Role != model.RoleRequester {
		t.Errorf("Expected role 'requester', got '%s'", resp.Data.Role)
	}

	// Verify that password in database is hashed with bcrypt and not plain text
	userInDb, err := mockRepo.FindByEmail(context.Background(), "john@example.com")
	if err != nil {
		t.Fatalf("User should exist in database: %v", err)
	}
	if userInDb.PasswordHash == "Password123!" {
		t.Fatalf("Password must NOT be stored in plain text!")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userInDb.PasswordHash), []byte("Password123!")); err != nil {
		t.Errorf("Password hash does not match original password: %v", err)
	}
}

// 2. Test Register Email Duplikat (HTTP 409 Conflict)
func TestRegister_DuplicateEmail_ReturnsConflict(t *testing.T) {
	mockRepo := newMockUserRepository()
	router, _ := setupAuthTestRouter(mockRepo, 24)

	payload := map[string]string{
		"name":     "Jane Doe",
		"email":    "jane@example.com",
		"password": "Password123!",
	}
	body, _ := json.Marshal(payload)

	// First registration
	req1, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("Initial register failed: code %d", w1.Code)
	}

	// Duplicate registration with same email
	req2, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("Expected HTTP 409 Conflict for duplicate email, got %d. Body: %s", w2.Code, w2.Body.String())
	}
}

// 3. Test Public Registration Tidak Dapat Membuat Akun Admin / Agent
func TestRegister_PublicRegistrationForcesRequesterRole(t *testing.T) {
	mockRepo := newMockUserRepository()
	router, _ := setupAuthTestRouter(mockRepo, 24)

	// Attempt to register with role "admin"
	payload := map[string]string{
		"name":     "Malicious Admin Attempt",
		"email":    "hacker@example.com",
		"password": "Password123!",
		"role":     "admin",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	userInDb, _ := mockRepo.FindByEmail(context.Background(), "hacker@example.com")
	if userInDb.Role != model.RoleRequester {
		t.Errorf("Security violation: public registration must force role to 'requester', but got '%s'", userInDb.Role)
	}
}

// 4. Test Login Berhasil (HTTP 200 & JWT Valid)
func TestLogin_Success(t *testing.T) {
	mockRepo := newMockUserRepository()
	router, _ := setupAuthTestRouter(mockRepo, 24)

	// Pre-create user with hashed password
	hashed, _ := bcrypt.GenerateFromPassword([]byte("SecretPass123!"), bcrypt.DefaultCost)
	mockRepo.Create(context.Background(), &model.User{
		ID:           uuid.New(),
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: string(hashed),
		Role:         model.RoleRequester,
	})

	loginPayload := map[string]string{
		"email":    "alice@example.com",
		"password": "SecretPass123!",
	}
	body, _ := json.Marshal(loginPayload)

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool               `json:"success"`
		Message string             `json:"message"`
		Data    model.AuthResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Data.Token == "" {
		t.Errorf("Expected non-empty JWT token")
	}
	if resp.Data.User.Email != "alice@example.com" {
		t.Errorf("Expected user email 'alice@example.com', got '%s'", resp.Data.User.Email)
	}
}

// 5. Test Login dengan Password Salah (HTTP 401 Unauthorized)
func TestLogin_WrongPassword_ReturnsUnauthorized(t *testing.T) {
	mockRepo := newMockUserRepository()
	router, _ := setupAuthTestRouter(mockRepo, 24)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("CorrectPassword!"), bcrypt.DefaultCost)
	mockRepo.Create(context.Background(), &model.User{
		ID:           uuid.New(),
		Name:         "Bob",
		Email:        "bob@example.com",
		PasswordHash: string(hashed),
		Role:         model.RoleRequester,
	})

	loginPayload := map[string]string{
		"email":    "bob@example.com",
		"password": "WrongPassword!",
	}
	body, _ := json.Marshal(loginPayload)

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected HTTP 401 Unauthorized for wrong password, got %d", w.Code)
	}
}

// 6. Test GET /api/me dengan JWT Valid
func TestGetMe_ValidToken(t *testing.T) {
	mockRepo := newMockUserRepository()
	router, authSvc := setupAuthTestRouter(mockRepo, 24)

	userID := uuid.New()
	user := &model.User{
		ID:           userID,
		Name:         "Charlie",
		Email:        "charlie@example.com",
		PasswordHash: "hashed",
		Role:         model.RoleAgent,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	mockRepo.Create(context.Background(), user)

	token, err := authSvc.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool               `json:"success"`
		Data    model.UserResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.ID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, resp.Data.ID)
	}
	if resp.Data.Email != "charlie@example.com" {
		t.Errorf("Expected email 'charlie@example.com', got '%s'", resp.Data.Email)
	}
	if resp.Data.Role != model.RoleAgent {
		t.Errorf("Expected role 'agent', got '%s'", resp.Data.Role)
	}
}

// 7. Test JWT Expired Ditolak (HTTP 401 Unauthorized)
func TestGetMe_ExpiredToken_ReturnsUnauthorized(t *testing.T) {
	mockRepo := newMockUserRepository()
	secret := "test_secret_key_1234567890_abcdef"
	router, _ := setupAuthTestRouter(mockRepo, 24)

	// Create an expired token manually
	claims := &service.JWTClaims{
		UserID: uuid.New().String(),
		Email:  "expired@example.com",
		Role:   "requester",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // expired 1 hour ago
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	req, _ := http.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected HTTP 401 Unauthorized for expired token, got %d", w.Code)
	}
}

// 8. Test Request ke /api/me tanpa Authorization Header (HTTP 401)
func TestGetMe_NoToken_ReturnsUnauthorized(t *testing.T) {
	mockRepo := newMockUserRepository()
	router, _ := setupAuthTestRouter(mockRepo, 24)

	req, _ := http.NewRequest(http.MethodGet, "/api/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected HTTP 401 Unauthorized when missing header, got %d", w.Code)
	}
}
