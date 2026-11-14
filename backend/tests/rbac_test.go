package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"helpdesk-ticketing-system/backend/internal/config"
	"helpdesk-ticketing-system/backend/internal/handler"
	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/routes"
	"helpdesk-ticketing-system/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// setupRBACTestRouter builds a full Gin engine with route protections and in-memory mock repository.
func setupRBACTestRouter() (*gin.Engine, service.AuthService, *mockUserRepository, string) {
	jwtSecret := "test_rbac_secret_key_9876543210_xyz"
	mockRepo := newMockUserRepository()

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

	authSvc := service.NewAuthService(mockRepo, jwtCfg)
	authHdlr := handler.NewAuthHandler(authSvc)

	router := routes.SetupRouter(routes.RouterParams{
		AppConfig:   appCfg,
		AuthHandler: authHdlr,
		AuthService: authSvc,
	})

	return router, authSvc, mockRepo, jwtSecret
}

// createTestUserAndToken helper creates a user with the specified role and generates a signed JWT.
func createTestUserAndToken(t *testing.T, authSvc service.AuthService, mockRepo *mockUserRepository, role string) (*model.User, string) {
	user := &model.User{
		ID:           uuid.New(),
		Name:         fmt.Sprintf("User_%s", role),
		Email:        fmt.Sprintf("%s@example.com", role),
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

// 1. Test Request Tanpa Token Mendapatkan HTTP 401
func TestRBAC_1_MissingToken_ReturnsUnauthorized(t *testing.T) {
	router, _, _, _ := setupRBACTestRouter()

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/me"},
		{http.MethodGet, "/api/tickets"},
		{http.MethodPost, "/api/tickets"},
		{http.MethodGet, "/api/users"},
		{http.MethodPatch, "/api/users/" + uuid.New().String() + "/role"},
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s %s", ep.method, ep.path), func(t *testing.T) {
			req, _ := http.NewRequest(ep.method, ep.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected HTTP 401 Unauthorized, got %d. Body: %s", w.Code, w.Body.String())
			}

			var resp handler.APIResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Response is not valid JSON: %v", err)
			}
			if resp.Success {
				t.Errorf("Expected Success=false for unauthorized request")
			}
		})
	}
}

// 2. Test Token Invalid Mendapatkan HTTP 401
func TestRBAC_2_InvalidToken_ReturnsUnauthorized(t *testing.T) {
	router, _, _, _ := setupRBACTestRouter()

	invalidTokens := []struct {
		name        string
		tokenHeader string
	}{
		{"Malformed JWT string", "Bearer invalid-malformed-jwt-token"},
		{"Non-Bearer format scheme", "Token some-token-string"},
		{"Bearer with empty token", "Bearer "},
		{"Signed with wrong secret", func() string {
			claims := &service.JWTClaims{
				UserID: uuid.New().String(),
				Email:  "user@example.com",
				Role:   model.RoleRequester,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signed, _ := tok.SignedString([]byte("wrong_tampered_secret"))
			return "Bearer " + signed
		}()},
		{"Valid signature but invalid claims (non-UUID user ID)", func() string {
			claims := &service.JWTClaims{
				UserID: "not-a-valid-uuid",
				Email:  "user@example.com",
				Role:   model.RoleRequester,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signed, _ := tok.SignedString([]byte("test_rbac_secret_key_9876543210_xyz"))
			return "Bearer " + signed
		}()},
		{"Valid signature but invalid role claim", func() string {
			claims := &service.JWTClaims{
				UserID: uuid.New().String(),
				Email:  "user@example.com",
				Role:   "superuser",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signed, _ := tok.SignedString([]byte("test_rbac_secret_key_9876543210_xyz"))
			return "Bearer " + signed
		}()},
	}

	for _, tt := range invalidTokens {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/api/tickets", nil)
			req.Header.Set("Authorization", tt.tokenHeader)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected HTTP 401 Unauthorized for %s, got %d. Body: %s", tt.name, w.Code, w.Body.String())
			}
		})
	}
}

// 3. Test Token Expired Mendapatkan HTTP 401
func TestRBAC_3_ExpiredToken_ReturnsUnauthorized(t *testing.T) {
	router, _, _, jwtSecret := setupRBACTestRouter()

	// Token expired 2 hours ago
	claims := &service.JWTClaims{
		UserID: uuid.New().String(),
		Email:  "expired_user@example.com",
		Role:   model.RoleRequester,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-3 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := tok.SignedString([]byte(jwtSecret))

	endpoints := []string{"/api/me", "/api/tickets", "/api/users"}

	for _, path := range endpoints {
		t.Run(path, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected HTTP 401 Unauthorized for expired token on %s, got %d. Body: %s", path, w.Code, w.Body.String())
			}
		})
	}
}

// 4. Test Token Valid Dapat Mengakses Endpoint Authenticated
func TestRBAC_4_ValidToken_CanAccessAuthenticatedEndpoints(t *testing.T) {
	router, authSvc, mockRepo, _ := setupRBACTestRouter()

	roles := []string{model.RoleRequester, model.RoleAgent, model.RoleAdmin}

	for _, role := range roles {
		t.Run(fmt.Sprintf("Role %s", role), func(t *testing.T) {
			_, token := createTestUserAndToken(t, authSvc, mockRepo, role)

			// 1. GET /api/me
			reqMe, _ := http.NewRequest(http.MethodGet, "/api/me", nil)
			reqMe.Header.Set("Authorization", "Bearer "+token)
			wMe := httptest.NewRecorder()
			router.ServeHTTP(wMe, reqMe)
			if wMe.Code != http.StatusOK {
				t.Errorf("Role %s: Expected HTTP 200 on GET /api/me, got %d. Body: %s", role, wMe.Code, wMe.Body.String())
			}

			// 2. GET /api/tickets
			reqTickets, _ := http.NewRequest(http.MethodGet, "/api/tickets", nil)
			reqTickets.Header.Set("Authorization", "Bearer "+token)
			wTickets := httptest.NewRecorder()
			router.ServeHTTP(wTickets, reqTickets)
			if wTickets.Code != http.StatusOK {
				t.Errorf("Role %s: Expected HTTP 200 on GET /api/tickets, got %d. Body: %s", role, wTickets.Code, wTickets.Body.String())
			}

			// 3. POST /api/tickets
			reqPostTicket, _ := http.NewRequest(http.MethodPost, "/api/tickets", bytes.NewBuffer([]byte(`{}`)))
			reqPostTicket.Header.Set("Authorization", "Bearer "+token)
			reqPostTicket.Header.Set("Content-Type", "application/json")
			wPostTicket := httptest.NewRecorder()
			router.ServeHTTP(wPostTicket, reqPostTicket)
			if wPostTicket.Code != http.StatusOK {
				t.Errorf("Role %s: Expected HTTP 200 on POST /api/tickets, got %d. Body: %s", role, wPostTicket.Code, wPostTicket.Body.String())
			}
		})
	}
}

// 5. Test Requester Ditolak Ketika Mengakses Endpoint Admin (HTTP 403 Forbidden)
func TestRBAC_5_Requester_ForbiddenOnAdminEndpoints(t *testing.T) {
	router, authSvc, mockRepo, _ := setupRBACTestRouter()

	_, requesterToken := createTestUserAndToken(t, authSvc, mockRepo, model.RoleRequester)

	adminEndpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/users"},
		{http.MethodPatch, "/api/users/" + uuid.New().String() + "/role"},
	}

	for _, ep := range adminEndpoints {
		t.Run(fmt.Sprintf("%s %s", ep.method, ep.path), func(t *testing.T) {
			req, _ := http.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Authorization", "Bearer "+requesterToken)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Errorf("Expected HTTP 403 Forbidden for requester on %s %s, got %d. Body: %s", ep.method, ep.path, w.Code, w.Body.String())
			}

			var resp handler.APIResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Response is not JSON: %v", err)
			}
			if resp.Success {
				t.Errorf("Expected Success=false for forbidden access")
			}
		})
	}
}

// 6. Test Agent Ditolak Ketika Mengakses Endpoint Admin (HTTP 403 Forbidden)
func TestRBAC_6_Agent_ForbiddenOnAdminEndpoints(t *testing.T) {
	router, authSvc, mockRepo, _ := setupRBACTestRouter()

	_, agentToken := createTestUserAndToken(t, authSvc, mockRepo, model.RoleAgent)

	adminEndpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/users"},
		{http.MethodPatch, "/api/users/" + uuid.New().String() + "/role"},
	}

	for _, ep := range adminEndpoints {
		t.Run(fmt.Sprintf("%s %s", ep.method, ep.path), func(t *testing.T) {
			req, _ := http.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Authorization", "Bearer "+agentToken)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Errorf("Expected HTTP 403 Forbidden for agent on %s %s, got %d. Body: %s", ep.method, ep.path, w.Code, w.Body.String())
			}
		})
	}
}

// 7. Test Admin Berhasil Melewati Middleware Admin (HTTP 200 OK)
func TestRBAC_7_Admin_AuthorizedOnAdminEndpoints(t *testing.T) {
	router, authSvc, mockRepo, _ := setupRBACTestRouter()

	_, adminToken := createTestUserAndToken(t, authSvc, mockRepo, model.RoleAdmin)

	targetUserID := uuid.New().String()
	adminEndpoints := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/users", ""},
		{http.MethodPatch, "/api/users/" + targetUserID + "/role", `{"role":"agent"}`},
	}

	for _, ep := range adminEndpoints {
		t.Run(fmt.Sprintf("%s %s", ep.method, ep.path), func(t *testing.T) {
			var bodyReader *bytes.Buffer
			if ep.body != "" {
				bodyReader = bytes.NewBuffer([]byte(ep.body))
			} else {
				bodyReader = bytes.NewBuffer(nil)
			}

			req, _ := http.NewRequest(ep.method, ep.path, bodyReader)
			req.Header.Set("Authorization", "Bearer "+adminToken)
			if ep.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected HTTP 200 OK for admin on %s %s, got %d. Body: %s", ep.method, ep.path, w.Code, w.Body.String())
			}

			var resp handler.APIResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Response is not JSON: %v", err)
			}
			if !resp.Success {
				t.Errorf("Expected Success=true for admin access")
			}
		})
	}
}

// 8. Test Manipulasi Role Melalui Request Tidak Mengubah Role Pengguna
func TestRBAC_8_RoleManipulationAttempt_DoesNotGrantPrivileges(t *testing.T) {
	router, authSvc, mockRepo, _ := setupRBACTestRouter()

	// Authenticate as a legitimate requester
	_, requesterToken := createTestUserAndToken(t, authSvc, mockRepo, model.RoleRequester)

	manipulationAttempts := []struct {
		name   string
		method string
		path   string
		body   string
		header map[string]string
	}{
		{
			name:   "Query parameter spoofing (?role=admin)",
			method: http.MethodGet,
			path:   "/api/users?role=admin",
			body:   "",
		},
		{
			name:   "Query parameter spoofing (?user_role=admin)",
			method: http.MethodGet,
			path:   "/api/users?user_role=admin",
			body:   "",
		},
		{
			name:   "JSON Body manipulation ({\"role\":\"admin\"}) on PATCH",
			method: http.MethodPatch,
			path:   "/api/users/" + uuid.New().String() + "/role",
			body:   `{"role":"admin"}`,
		},
		{
			name:   "Header injection (X-Role: admin)",
			method: http.MethodGet,
			path:   "/api/users",
			body:   "",
			header: map[string]string{"X-Role": "admin"},
		},
		{
			name:   "Header injection (X-User-Role: admin)",
			method: http.MethodGet,
			path:   "/api/users",
			body:   "",
			header: map[string]string{"X-User-Role": "admin"},
		},
	}

	for _, tt := range manipulationAttempts {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Buffer
			if tt.body != "" {
				bodyReader = bytes.NewBuffer([]byte(tt.body))
			} else {
				bodyReader = bytes.NewBuffer(nil)
			}

			req, _ := http.NewRequest(tt.method, tt.path, bodyReader)
			req.Header.Set("Authorization", "Bearer "+requesterToken)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			for k, v := range tt.header {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// MUST still be rejected with HTTP 403 Forbidden!
			if w.Code != http.StatusForbidden {
				t.Fatalf("Security failure: %s bypassed RBAC middleware! Got status %d, expected 403 Forbidden. Body: %s",
					tt.name, w.Code, w.Body.String())
			}

			var resp handler.APIResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Response parsing failed: %v", err)
			}
			if resp.Success {
				t.Errorf("Security failure: Success was true despite privilege manipulation attempt")
			}
		})
	}
}
