package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Placeholder returns a Gin handler function responding with HTTP 200 OK.
// It is used to wire protected routes before full domain implementations (e.g. Prompt 4).
func Placeholder(message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		SuccessResponse(c, http.StatusOK, message, nil)
	}
}
