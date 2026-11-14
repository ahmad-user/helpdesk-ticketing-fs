package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"helpdesk-ticketing-system/backend/internal/handler"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics and responds with a 500 JSON error.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVERED] %v\nStack trace:\n%s\n", r, string(debug.Stack()))
				handler.ErrorResponse(c, http.StatusInternalServerError, "Internal server error occurred", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}
