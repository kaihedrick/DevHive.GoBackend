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
	"devhive-backend/internal/middleware"
	"devhive-backend/internal/ws"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	if err := db.InitDB(); err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.CloseDB()

	if err := config.InitFirebase(); err != nil {
		log.Fatal("Error initializing Firebase:", err)
	}

	log.Println("Feature flags initialization skipped during startup")

	ws.StartWebSocketHub()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.RateLimiter(middleware.DefaultRateLimit))

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	router.Use(func(c *gin.Context) {
		corsMiddleware.HandlerFunc(c.Writer, c.Request)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "DevHive Backend",
			"time":    time.Now().UTC(),
		})
	})

	router.GET("/ws", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		ws.HandleConnections(ws.GlobalHub, w, r)
	}))

	router.GET("/ws/auth", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		ws.AuthenticatedHandleConnections(ws.GlobalHub, w, r)
	}))

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/refresh", controllers.RefreshToken)
		}

		protected := api.Group("/")
		protected.Use(controllers.AuthMiddleware())
		{
			users := protected.Group("/users")
			{
				users.GET("/profile", controllers.GetUserProfile)
				users.PUT("/profile", controllers.UpdateUserProfile)
				users.POST("/avatar", controllers.UploadAvatar)
			}

			projects := protected.Group("/projects")
			{
				projects.GET("/", controllers.GetProjects)
				projects.POST("/", controllers.CreateProject)
				projects.GET(":id", controllers.GetProject)
				projects.PUT(":id", controllers.UpdateProject)
				projects.DELETE(":id", controllers.DeleteProject)
				projects.POST(":id/members", controllers.AddProjectMember)
				projects.DELETE(":id/members/:userId", controllers.RemoveProjectMember)
			}

			sprints := protected.Group("/projects/:id/sprints")
			{
				sprints.GET("/", controllers.GetSprints)
				sprints.POST("/", controllers.CreateSprint)
				sprints.GET(":sprintId", controllers.GetSprint)
				sprints.PUT(":sprintId", controllers.UpdateSprint)
				sprints.DELETE(":sprintId", controllers.DeleteSprint)
			}

			messages := protected.Group("/projects/:id/messages")
			{
				messages.GET("/", controllers.GetMessages)
				messages.POST("/", controllers.CreateMessage)
				messages.PUT(":messageId", controllers.UpdateMessage)
				messages.DELETE(":messageId", controllers.DeleteMessage)
			}

			featureFlags := protected.Group("/admin/feature-flags")
			{
				featureFlags.GET("/", controllers.GetFeatureFlags)
				featureFlags.GET(":key", controllers.GetFeatureFlag)
				featureFlags.POST("/", controllers.CreateFeatureFlag)
				featureFlags.PUT(":key", controllers.UpdateFeatureFlag)
				featureFlags.DELETE(":key", controllers.DeleteFeatureFlag)
				featureFlags.POST("/bulk-update", controllers.BulkUpdateFeatureFlags)
			}

			mobile := protected.Group("/mobile/v2")
			{
				mobile.Use(middleware.MobileRateLimit())
				mobile.GET("/projects", controllers.GetMobileProjects)
				mobile.GET("/projects/:id", controllers.GetMobileProject)
				mobile.GET("/projects/:id/sprints", controllers.GetMobileSprints)
				mobile.GET("/projects/:id/messages", controllers.GetMobileMessages)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting DevHive Backend on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
