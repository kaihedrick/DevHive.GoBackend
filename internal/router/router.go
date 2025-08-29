package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"devhive-backend/internal/auth"
	"devhive-backend/internal/middleware"
	"devhive-backend/internal/ws"
)

// Register registers all routes with the Gin engine
func Register(r *gin.Engine) {
	// Apply global middleware
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter(middleware.DefaultRateLimit))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "DevHive Backend",
		})
	})

	// WebSocket endpoints
	r.GET("/ws", gin.WrapF(ws.HandleWS))
	r.GET("/ws/auth", gin.WrapF(ws.HandleWSAuth))

	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		authController := auth.NewController()
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/refresh", authController.RefreshToken)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// TODO: Register feature controllers
			// users := protected.Group("/users")
			// projects := protected.Group("/projects")
			// sprints := protected.Group("/projects/:id/sprints")
			// tasks := protected.Group("/projects/:id/tasks")
			// messages := protected.Group("/projects/:id/messages")
			// database := protected.Group("/database")
			// mail := protected.Group("/mail")
		}
	}

	// Swagger documentation
	r.StaticFile("/swagger/openapi.yaml", "./api/openapi.yaml")
	r.StaticFile("/swagger/doc.json", "./api/openapi.yaml")

	// Redirect root to API documentation
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/openapi.yaml")
	})
}
