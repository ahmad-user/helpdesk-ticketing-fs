package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"helpdesk-ticketing-system/backend/internal/config"
	"helpdesk-ticketing-system/backend/internal/handler"
	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/routes"
	"helpdesk-ticketing-system/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// mockHealthRepository implements repository.HealthRepository for isolated testing.
type mockHealthRepository struct {
	pingErr error
}

func (m *mockHealthRepository) Ping(ctx context.Context) error {
	return m.pingErr
}

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter(pingErr error) *gin.Engine {
	appCfg := config.AppConfig{
		Name: "helpdesk-ticketing-system",
		Env:  "test",
		Port: "8080",
		Host: "0.0.0.0",
	}

	mockRepo := &mockHealthRepository{pingErr: pingErr}
	healthService := service.NewHealthService(mockRepo, appCfg)
	healthHandler := handler.NewHealthHandler(healthService)

	return routes.SetupRouter(routes.RouterParams{
		AppConfig:     appCfg,
		HealthHandler: healthHandler,
	})
}

func TestHealthCheckEndpoint_DatabaseConnected(t *testing.T) {
	router := setupTestRouter(nil)

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response struct {
		Success bool               `json:"success"`
		Message string             `json:"message"`
		Data    model.HealthStatus `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v. Body: %s", err, w.Body.String())
	}

	if !response.Success {
		t.Errorf("Expected response.Success to be true, got false")
	}

	if response.Data.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Data.Status)
	}

	if response.Data.Database != "connected" {
		t.Errorf("Expected database status 'connected', got '%s'", response.Data.Database)
	}

	if response.Data.App != "helpdesk-ticketing-system" {
		t.Errorf("Expected app name 'helpdesk-ticketing-system', got '%s'", response.Data.App)
	}
}

func TestHealthCheckEndpoint_DatabaseDisconnected(t *testing.T) {
	router := setupTestRouter(errors.New("connection refused"))

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// In degraded mode, server responds with 200 OK and degraded status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response struct {
		Success bool               `json:"success"`
		Message string             `json:"message"`
		Data    model.HealthStatus `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v. Body: %s", err, w.Body.String())
	}

	if !response.Success {
		t.Errorf("Expected response.Success to be true, got false")
	}

	if response.Data.Status != "degraded" {
		t.Errorf("Expected status 'degraded', got '%s'", response.Data.Status)
	}

	if response.Data.Database != "disconnected" {
		t.Errorf("Expected database status 'disconnected', got '%s'", response.Data.Database)
	}

	if _, exists := response.Data.Details["database_error"]; !exists {
		t.Errorf("Expected details to contain 'database_error', got %v", response.Data.Details)
	}
}
