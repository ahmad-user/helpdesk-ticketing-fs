package routes

import (
	"net/http"
	"os"

	"helpdesk-ticketing-system/backend/internal/config"
	"helpdesk-ticketing-system/backend/internal/handler"
	"helpdesk-ticketing-system/backend/internal/middleware"
	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// RouterParams encapsulates dependencies required to configure HTTP routes.
type RouterParams struct {
	AppConfig      config.AppConfig
	HealthHandler  *handler.HealthHandler
	AuthHandler    *handler.AuthHandler
	TicketHandler  *handler.TicketHandler
	CommentHandler *handler.CommentHandler
	UserHandler    *handler.UserHandler
	AuthService    service.AuthService
}

// SetupRouter initializes Gin engine, attaches middlewares, and registers routes.
func SetupRouter(params RouterParams) *gin.Engine {
	if params.AppConfig.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global Middlewares
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint (Prompt 1)
	if params.HealthHandler != nil {
		router.GET("/health", params.HealthHandler.GetHealth)
	}

	// OpenAPI 3.0 specification & Swagger UI documentation
	router.GET("/docs/openapi.yaml", func(c *gin.Context) {
		candidates := []string{"docs/openapi.yaml", "../docs/openapi.yaml", "../../docs/openapi.yaml"}
		for _, p := range candidates {
			if data, err := os.ReadFile(p); err == nil {
				c.Data(http.StatusOK, "application/yaml", data)
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "OpenAPI specification file not found"})
	})

	router.GET("/docs", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Helpdesk API Documentation - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5/favicon-32x32.png" sizes="32x32" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin:0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/docs/openapi.yaml",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
      window.ui = ui;
    };
  </script>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// API Group
	api := router.Group("/api")
	{
		// 1. Public Authentication routes
		if params.AuthHandler != nil {
			auth := api.Group("/auth")
			{
				auth.POST("/register", params.AuthHandler.Register)
				auth.POST("/login", params.AuthHandler.Login)
			}
		}

		// 2. Protected routes (requires valid JWT authentication)
		if params.AuthService != nil {
			authMiddleware := middleware.AuthMiddleware(params.AuthService)

			// Authenticated user profile
			if params.AuthHandler != nil {
				api.GET("/me", authMiddleware, params.AuthHandler.GetMe)
			}

			// Tickets routes: Protected by JWT Authentication
			tickets := api.Group("/tickets")
			tickets.Use(authMiddleware)
			{
				if params.TicketHandler != nil {
					tickets.POST("", params.TicketHandler.Create)
					tickets.GET("", params.TicketHandler.List)
					tickets.GET("/:id", params.TicketHandler.GetByID)
					tickets.PATCH("/:id", params.TicketHandler.Update)
					tickets.DELETE("/:id", params.TicketHandler.Delete)
				} else {
					tickets.GET("", handler.Placeholder("Tickets list retrieved successfully"))
					tickets.POST("", handler.Placeholder("Ticket created successfully"))
				}

				if params.CommentHandler != nil {
					tickets.POST("/:id/comments", params.CommentHandler.Create)
					tickets.GET("/:id/comments", params.CommentHandler.List)
				}
			}

			// Admin routes: Protected by JWT Authentication and RBAC (Admin role required)
			users := api.Group("/users")
			users.Use(authMiddleware, middleware.RequireRole(model.RoleAdmin))
			{
				if params.UserHandler != nil {
					users.GET("", params.UserHandler.List)
					users.POST("", params.UserHandler.Create)
					users.PATCH("/:id", params.UserHandler.Update)
					users.DELETE("/:id", params.UserHandler.Delete)
					users.PATCH("/:id/role", params.UserHandler.UpdateRole)
				} else {
					users.GET("", handler.Placeholder("User list retrieved successfully"))
					users.POST("", handler.Placeholder("User created successfully"))
					users.PATCH("/:id", handler.Placeholder("User updated successfully"))
					users.DELETE("/:id", handler.Placeholder("User deleted successfully"))
					users.PATCH("/:id/role", handler.Placeholder("User role updated successfully"))
				}
			}
		}
	}

	return router
}
