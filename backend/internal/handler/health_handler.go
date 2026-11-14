package handler

import (
	"net/http"

	"helpdesk-ticketing-system/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles HTTP requests related to system health checks.
type HealthHandler struct {
	service service.HealthService
}

// NewHealthHandler creates a new instance of HealthHandler.
func NewHealthHandler(service service.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

// GetHealth handles GET /health requests.
func (h *HealthHandler) GetHealth(c *gin.Context) {
	status, err := h.service.CheckHealth(c.Request.Context())
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to check system health", err.Error())
		return
	}

	SuccessResponse(c, http.StatusOK, "System is running", status)
}
