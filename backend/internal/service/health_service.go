package service

import (
	"context"
	"time"

	"helpdesk-ticketing-system/backend/internal/config"
	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"
)

// HealthService defines business logic operations for system health checking.
type HealthService interface {
	CheckHealth(ctx context.Context) (*model.HealthStatus, error)
}

type healthService struct {
	repo      repository.HealthRepository
	appConfig config.AppConfig
}

// NewHealthService creates a new instance of HealthService.
func NewHealthService(repo repository.HealthRepository, appConfig config.AppConfig) HealthService {
	return &healthService{
		repo:      repo,
		appConfig: appConfig,
	}
}

// CheckHealth evaluates the health of the application and its dependencies.
func (s *healthService) CheckHealth(ctx context.Context) (*model.HealthStatus, error) {
	dbStatus := "connected"
	overallStatus := "ok"
	details := make(map[string]string)

	if err := s.repo.Ping(ctx); err != nil {
		dbStatus = "disconnected"
		overallStatus = "degraded"
		details["database_error"] = err.Error()
	}

	status := &model.HealthStatus{
		Status:    overallStatus,
		App:       s.appConfig.Name,
		Version:   "1.0.0",
		Timestamp: time.Now().UTC(),
		Database:  dbStatus,
		Details:   details,
	}

	return status, nil
}
