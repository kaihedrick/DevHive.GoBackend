package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devhive-backend/config"
	"devhive-backend/controllers"
	"devhive-backend/db"
	"devhive-backend/internal/flags"
	"devhive-backend/internal/middleware"
	"devhive-backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/gzip"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.CloseDB()

	// Initialize Firebase
	if err := config.InitFirebase(); err != nil {
		log.Fatal("Error initializing Firebase:", err)
	}

	// Initialize feature flags
	flags.InitGlobalManager(db.GetDB())

	// Start WebSocket hub
	ws.StartWebSocketHub()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add gzip compression if enabled
	if flags.IsEnabledGlobal("gzip_compression") {
		router.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	// Add rate limiting middleware
	router.Use(middleware.RateLimiter(middleware.DefaultRateLimit))

	// CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Apply CORS middleware
	router.Use(func(c *gin.Context) {
		corsMiddleware.HandlerFunc(c.Writer, c.Request)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "DevHive Backend",
			"time":    time.Now().UTC(),
		})
	})

	// WebSocket endpoint for real-time updates
	if flags.IsEnabledGlobal("enable_websockets") {
		router.GET("/ws", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
			ws.HandleConnections(ws.GlobalHub, w, r)
		}))
	}

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/refresh", controllers.RefreshToken)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(controllers.AuthMiddleware())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", controllers.GetUserProfile)
				users.PUT("/profile", controllers.UpdateUserProfile)
				users.POST("/avatar", controllers.UploadAvatar)
			}

			// Project routes
			projects := protected.Group("/projects")
			{
				projects.GET("/", controllers.GetProjects)
				projects.POST("/", controllers.CreateProject)
				projects.GET("/:id", controllers.GetProject)
				projects.PUT("/:id", controllers.UpdateProject)
				projects.DELETE("/:id", controllers.DeleteProject)
				projects.POST("/:id/members", controllers.AddProjectMember)
				projects.DELETE("/:id/members/:userId", controllers.RemoveProjectMember)
			}

			// Sprint routes
			sprints := protected.Group("/projects/:projectId/sprints")
			{
				sprints.GET("/", controllers.GetSprints)
				sprints.POST("/", controllers.CreateSprint)
				sprints.GET("/:id", controllers.GetSprint)
				sprints.PUT("/:id", controllers.UpdateSprint)
				sprints.DELETE("/:id", controllers.DeleteSprint)
			}

			// Message routes
			messages := protected.Group("/projects/:projectId/messages")
			{
				messages.GET("/", controllers.GetMessages)
				messages.POST("/", controllers.CreateMessage)
				messages.PUT("/:id", controllers.UpdateMessage)
				messages.DELETE("/:id", controllers.DeleteMessage)
			}

			// Mobile API routes (v2)
			if flags.IsEnabledGlobal("mobile_v2_api") {
				mobile := protected.Group("/mobile/v2")
				{
					mobile.Use(middleware.MobileRateLimit())
					mobile.GET("/projects", controllers.GetMobileProjects)
					mobile.GET("/projects/:id", controllers.GetMobileProject)
					mobile.GET("/projects/:projectId/sprints", controllers.GetMobileSprints)
					mobile.GET("/projects/:projectId/messages", controllers.GetMobileMessages)
				}
			}
		}
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting DevHive Backend on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
