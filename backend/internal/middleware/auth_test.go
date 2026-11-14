package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"helpdesk-ticketing-system/backend/internal/config"
	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// mockAuthService implements service.AuthService for testing middleware in isolation.
type mockAuthService struct {
	validateFunc func(tokenString string) (*service.JWTClaims, error)
}

func (m *mockAuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.UserResponse, error) {
	return nil, nil
}
func (m *mockAuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	return nil, nil
}
func (m *mockAuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error) {
	return nil, nil
}
func (m *mockAuthService) GenerateToken(user *model.User) (string, error) {
	return "", nil
}
func (m *mockAuthService) ValidateToken(tokenString string) (*service.JWTClaims, error) {
	if m.validateFunc != nil {
		return m.validateFunc(tokenString)
	}
	return nil, service.ErrInvalidToken
}

func init() {
	gin.SetMode(gin.TestMode)
}

// Test AuthMiddleware with various header & token conditions
func TestAuthMiddleware(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name           string
		authHeader     string
		validateFunc   func(tokenString string) (*service.JWTClaims, error)
		expectedStatus int
		expectContext  bool
	}{
		{
			name:           "Missing Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "Malformed Authorization header (not Bearer)",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "Bearer prefix with empty token",
			authHeader:     "Bearer   ",
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:       "Invalid or expired token",
			authHeader: "Bearer invalid.jwt.token",
			validateFunc: func(tokenString string) (*service.JWTClaims, error) {
				return nil, service.ErrInvalidToken
			},
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:       "Token with non-UUID user ID in claims",
			authHeader: "Bearer token.with.invalid-uuid",
			validateFunc: func(tokenString string) (*service.JWTClaims, error) {
				return &service.JWTClaims{
					UserID: "not-a-valid-uuid",
					Email:  "user@example.com",
					Role:   model.RoleRequester,
				}, nil
			},
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:       "Valid token successfully attaches context",
			authHeader: "Bearer valid.jwt.token",
			validateFunc: func(tokenString string) (*service.JWTClaims, error) {
				return &service.JWTClaims{
					UserID: testUserID.String(),
					Email:  "valid@example.com",
					Role:   model.RoleRequester,
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectContext:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAuthService{validateFunc: tt.validateFunc}
			r := gin.New()

			var capturedUser *CurrentUser
			var capturedRole string

			r.Use(AuthMiddleware(mockSvc))
			r.GET("/test-protected", func(c *gin.Context) {
				capturedUser, _ = GetCurrentUser(c)
				capturedRole, _ = GetUserRole(c)
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req, _ := http.NewRequest(http.MethodGet, "/test-protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.expectContext {
				if capturedUser == nil {
					t.Fatalf("Expected CurrentUser in context, got nil")
				}
				if capturedUser.ID != testUserID {
					t.Errorf("Expected user ID %s, got %s", testUserID, capturedUser.ID)
				}
				if capturedUser.Email != "valid@example.com" {
					t.Errorf("Expected user email 'valid@example.com', got '%s'", capturedUser.Email)
				}
				if capturedRole != model.RoleRequester {
					t.Errorf("Expected user role 'requester', got '%s'", capturedRole)
				}
			}
		})
	}
}

// Test Context Helpers
func TestContextHelpers(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test on empty context
	if _, err := GetCurrentUser(c); err == nil {
		t.Errorf("Expected error from GetCurrentUser on empty context")
	}
	if _, err := GetUserID(c); err == nil {
		t.Errorf("Expected error from GetUserID on empty context")
	}
	if _, err := GetUserEmail(c); err == nil {
		t.Errorf("Expected error from GetUserEmail on empty context")
	}
	if _, err := GetUserRole(c); err == nil {
		t.Errorf("Expected error from GetUserRole on empty context")
	}

	// Test with populated context
	testID := uuid.New()
	user := CurrentUser{
		ID:    testID,
		Email: "test@example.com",
		Role:  model.RoleAdmin,
	}

	c.Set(ContextUserIDKey, testID.String())
	c.Set(ContextUserEmailKey, "test@example.com")
	c.Set(ContextUserRoleKey, model.RoleAdmin)
	c.Set(ContextUserKey, user)

	retrievedUser, err := GetCurrentUser(c)
	if err != nil || retrievedUser.ID != testID {
		t.Errorf("GetCurrentUser failed: %v", err)
	}

	retrievedID, err := GetUserID(c)
	if err != nil || retrievedID != testID {
		t.Errorf("GetUserID failed: %v", err)
	}

	retrievedEmail, err := GetUserEmail(c)
	if err != nil || retrievedEmail != "test@example.com" {
		t.Errorf("GetUserEmail failed: %v", err)
	}

	retrievedRole, err := GetUserRole(c)
	if err != nil || retrievedRole != model.RoleAdmin {
		t.Errorf("GetUserRole failed: %v", err)
	}
}

// Test HasRole logic and hierarchy
func TestHasRole(t *testing.T) {
	tests := []struct {
		name         string
		userRole     string
		allowedRoles []string
		expected     bool
	}{
		{"Requester accessing requester", model.RoleRequester, []string{model.RoleRequester}, true},
		{"Requester accessing agent", model.RoleRequester, []string{model.RoleAgent}, false},
		{"Requester accessing admin", model.RoleRequester, []string{model.RoleAdmin}, false},
		{"Agent accessing agent", model.RoleAgent, []string{model.RoleAgent}, true},
		{"Agent accessing admin", model.RoleAgent, []string{model.RoleAdmin}, false},
		{"Admin accessing admin", model.RoleAdmin, []string{model.RoleAdmin}, true},
		{"Admin accessing agent (inherited)", model.RoleAdmin, []string{model.RoleAgent}, true},
		{"Unknown role rejected", "guest", []string{model.RoleRequester, model.RoleAgent, model.RoleAdmin}, false},
		{"Empty role rejected", "", []string{model.RoleAdmin}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasRole(tt.userRole, tt.allowedRoles...)
			if got != tt.expected {
				t.Errorf("HasRole(%q, %v) = %v; want %v", tt.userRole, tt.allowedRoles, got, tt.expected)
			}
		})
	}
}

// Test RequireRole middleware
func TestRequireRoleMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		middleware     gin.HandlerFunc
		expectedStatus int
	}{
		{
			name:           "No user context returns 401",
			setupContext:   func(c *gin.Context) {},
			middleware:     RequireRole(model.RoleAdmin),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Requester accessing admin endpoint returns 403",
			setupContext: func(c *gin.Context) {
				c.Set(ContextUserRoleKey, model.RoleRequester)
			},
			middleware:     RequireRole(model.RoleAdmin),
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Agent accessing admin endpoint returns 403",
			setupContext: func(c *gin.Context) {
				c.Set(ContextUserRoleKey, model.RoleAgent)
			},
			middleware:     RequireRole(model.RoleAdmin),
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Admin accessing admin endpoint returns 200",
			setupContext: func(c *gin.Context) {
				c.Set(ContextUserRoleKey, model.RoleAdmin)
			},
			middleware:     RequireRole(model.RoleAdmin),
			expectedStatus: http.StatusOK,
		},
		{
			name: "Admin accessing agent endpoint (hierarchy) returns 200",
			setupContext: func(c *gin.Context) {
				c.Set(ContextUserRoleKey, model.RoleAdmin)
			},
			middleware:     RequireRole(model.RoleAgent),
			expectedStatus: http.StatusOK,
		},
		{
			name: "RequireAdmin convenience helper allows admin",
			setupContext: func(c *gin.Context) {
				c.Set(ContextUserRoleKey, model.RoleAdmin)
			},
			middleware:     RequireAdmin(),
			expectedStatus: http.StatusOK,
		},
		{
			name: "RequireAgent convenience helper allows agent",
			setupContext: func(c *gin.Context) {
				c.Set(ContextUserRoleKey, model.RoleAgent)
			},
			middleware:     RequireAgent(),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/protected", func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			}, tt.middleware, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"data": "authorized"})
			})

			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

// Test ValidateToken logic in service directly for strict claim checks
func TestAuthService_ValidateToken_StrictClaims(t *testing.T) {
	jwtCfg := config.JWTConfig{
		Secret:          "unit_test_secret_key_12345",
		ExpirationHours: 1,
	}
	svc := service.NewAuthService(nil, jwtCfg)

	// Helper to generate custom signed token
	makeToken := func(claims *service.JWTClaims, secret string) string {
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		str, _ := tok.SignedString([]byte(secret))
		return str
	}

	validID := uuid.New().String()

	tests := []struct {
		name        string
		tokenString string
		expectErr   bool
	}{
		{
			name: "Valid claims and signature",
			tokenString: makeToken(&service.JWTClaims{
				UserID: validID,
				Email:  "user@example.com",
				Role:   model.RoleRequester,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}, jwtCfg.Secret),
			expectErr: false,
		},
		{
			name: "Wrong secret signature",
			tokenString: makeToken(&service.JWTClaims{
				UserID: validID,
				Email:  "user@example.com",
				Role:   model.RoleRequester,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}, "wrong_secret_key"),
			expectErr: true,
		},
		{
			name: "Expired token",
			tokenString: makeToken(&service.JWTClaims{
				UserID: validID,
				Email:  "user@example.com",
				Role:   model.RoleRequester,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(-10 * time.Minute)),
				},
			}, jwtCfg.Secret),
			expectErr: true,
		},
		{
			name: "Missing user ID",
			tokenString: makeToken(&service.JWTClaims{
				UserID: "",
				Email:  "user@example.com",
				Role:   model.RoleRequester,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}, jwtCfg.Secret),
			expectErr: true,
		},
		{
			name: "Invalid non-UUID user ID",
			tokenString: makeToken(&service.JWTClaims{
				UserID: "not-a-valid-uuid",
				Email:  "user@example.com",
				Role:   model.RoleRequester,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}, jwtCfg.Secret),
			expectErr: true,
		},
		{
			name: "Missing email",
			tokenString: makeToken(&service.JWTClaims{
				UserID: validID,
				Email:  "",
				Role:   model.RoleRequester,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}, jwtCfg.Secret),
			expectErr: true,
		},
		{
			name: "Invalid role",
			tokenString: makeToken(&service.JWTClaims{
				UserID: validID,
				Email:  "user@example.com",
				Role:   "superadmin", // unknown role
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}, jwtCfg.Secret),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := svc.ValidateToken(tt.tokenString)
			if (err != nil) != tt.expectErr {
				t.Fatalf("ValidateToken() error = %v, expectErr %v", err, tt.expectErr)
			}
			if !tt.expectErr && claims == nil {
				t.Fatalf("Expected non-nil claims for valid token")
			}
		})
	}
}
