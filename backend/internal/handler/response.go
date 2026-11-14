package handler

import (
	"helpdesk-ticketing-system/backend/internal/model"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard JSON response envelope for all API endpoints.
type APIResponse struct {
	Success    bool                      `json:"success"`
	Message    string                    `json:"message"`
	Data       interface{}               `json:"data,omitempty"`
	Pagination *model.PaginationMetadata `json:"pagination,omitempty"`
	Errors     interface{}               `json:"errors,omitempty"`
}

// SuccessResponse sends a standard formatted JSON success response.
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// PaginatedResponse sends a standard formatted JSON paginated success response.
func PaginatedResponse(c *gin.Context, statusCode int, message string, data interface{}, pagination *model.PaginationMetadata) {
	c.JSON(statusCode, APIResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}

// ErrorResponse sends a standard formatted JSON error response.
func ErrorResponse(c *gin.Context, statusCode int, message string, errors interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}
