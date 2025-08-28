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
	ws "devhive-backend/internal"
	"devhive-backend/middleware"
	"devhive-backend/repositories"
	"devhive-backend/services"

	_ "devhive-backend/docs" // This will be generated

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           DevHive Backend API
// @version         1.0
// @description     A comprehensive project management backend API for DevHive
// @termsOfService  http://swagger.io/terms/

// @contact.name   DevHive Team
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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

	ws.StartWebSocketHub()

	// Initialize database controller
	dbController := controllers.NewDatabaseController("scripts")

	// Initialize mail controller
	mailService := services.NewMailService()
	mailController := controllers.NewMailController(mailService)

	// Initialize services for mobile controller
	rawDB, err := db.GetRawDB()
	if err != nil {
		log.Fatal("Error getting raw database connection:", err)
	}

	projectService := services.NewProjectService(repositories.NewProjectRepository(rawDB), repositories.NewUserRepository(rawDB))
	sprintService := services.NewSprintService(db.GetDB())
	messageService := services.NewMessageService(repositories.NewMessageRepository(db.GetDB()))
	userService := services.NewUserService(repositories.NewUserRepository(rawDB))

	// Initialize mobile controller
	mobileController := controllers.NewMobileController(projectService, sprintService, messageService, userService)

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

	// Redirect root to Swagger documentation
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/")
	})

	router.GET("/ws", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		ws.HandleConnections(ws.GlobalHub, w, r)
	}))

	router.GET("/ws/auth", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		ws.AuthenticatedHandleConnections(ws.GlobalHub, w, r)
	}))

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/refresh", controllers.RefreshToken)
			auth.POST("/forgot-password", controllers.ForgotPassword)
			auth.POST("/reset-password", controllers.ResetPassword)
		}

		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			users := protected.Group("/users")
			{
				users.GET("/profile", controllers.GetUserProfile)
				users.PUT("/profile", controllers.UpdateUserProfile)
				users.POST("/avatar", controllers.UploadAvatar)
				users.PUT("/activate/:id", controllers.ActivateUser)
				users.PUT("/deactivate/:id", controllers.DeactivateUser)
				users.GET("/search", controllers.SearchUsers)
			}

			projects := protected.Group("/projects")
			{
				projects.GET("/", controllers.GetProjects)
				projects.POST("/", controllers.CreateProject)
				projects.GET(":id", controllers.GetProject)
				projects.PUT(":id", controllers.UpdateProject)
				projects.DELETE(":id", controllers.DeleteProject)
				projects.POST(":id/members", controllers.AddProjectMember)
				projects.GET(":id/members", controllers.GetProjectMembers)
				projects.DELETE(":id/members/:userId", controllers.RemoveProjectMember)
				projects.PUT(":id/members/:userId/role", controllers.UpdateProjectMemberRole)
			}

			sprints := protected.Group("/projects/:id/sprints")
			{
				sprints.GET("/", controllers.GetSprints)
				sprints.POST("/", controllers.CreateSprint)
				sprints.GET(":sprintId", controllers.GetSprint)
				sprints.PUT(":sprintId", controllers.UpdateSprint)
				sprints.DELETE(":sprintId", controllers.DeleteSprint)
				sprints.POST(":sprintId/start", controllers.StartSprint)
				sprints.POST(":sprintId/complete", controllers.CompleteSprint)
			}

			// Project-level task management
			projectTasks := protected.Group("/projects/:id/tasks")
			{
				projectTasks.GET("/", controllers.GetTasks)
				projectTasks.POST("/", controllers.CreateTask)
				projectTasks.GET("/:taskId", controllers.GetTask)
				projectTasks.PUT("/:taskId", controllers.UpdateTask)
				projectTasks.DELETE("/:taskId", controllers.DeleteTask)
				projectTasks.POST("/:taskId/assign", controllers.AssignTask)
				projectTasks.PATCH("/:taskId/status", controllers.UpdateTaskStatus)
			}

			// Sprint-level task management
			tasks := protected.Group("/projects/:id/sprints/:sprintId/tasks")
			{
				tasks.GET("/", controllers.GetTasksBySprint)
				tasks.POST("/", controllers.CreateTask)
				tasks.GET("/:taskId", controllers.GetTask)
				tasks.PUT("/:taskId", controllers.UpdateTask)
				tasks.DELETE("/:taskId", controllers.DeleteTask)
				tasks.POST("/:taskId/assign", controllers.AssignTask)
				tasks.PATCH("/:taskId/status", controllers.UpdateTaskStatus)
			}

			messages := protected.Group("/projects/:id/messages")
			{
				messages.GET("/", controllers.GetMessages)
				messages.POST("/", controllers.CreateMessage)
				messages.PUT("/:messageId", controllers.UpdateMessage)
				messages.DELETE("/:messageId", controllers.DeleteMessage)
			}

			database := protected.Group("/database")
			{
				database.POST("/execute-script", dbController.ExecuteScript)
				database.GET("/status", dbController.GetDatabaseStatus)
				database.GET("/scripts", dbController.ListScripts)
			}

			mail := protected.Group("/mail")
			{
				mail.POST("/send", mailController.SendEmail)
			}

			admin := protected.Group("/admin")
			{
				featureFlags := admin.Group("/feature-flags")
				{
					featureFlags.GET("/", controllers.GetFeatureFlags)
					featureFlags.GET("/:key", controllers.GetFeatureFlag)
					featureFlags.POST("/", controllers.CreateFeatureFlag)
					featureFlags.PUT("/:key", controllers.UpdateFeatureFlag)
					featureFlags.DELETE("/:key", controllers.DeleteFeatureFlag)
					featureFlags.POST("/bulk-update", controllers.BulkUpdateFeatureFlags)
				}
			}

			mobile := protected.Group("/mobile/v2")
			{
				mobile.GET("/projects", mobileController.GetMobileProjects)
				mobile.GET("/projects/:id", mobileController.GetMobileProject)
				mobile.GET("/projects/:id/sprints", mobileController.GetMobileSprints)
				mobile.GET("/projects/:id/messages", mobileController.GetMobileMessages)
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
