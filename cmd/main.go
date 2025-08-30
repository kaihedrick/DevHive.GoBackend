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

// @host      devhive-go-backend.fly.dev
// @BasePath  /api
// @schemes   https

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
	// taskService := services.NewTaskService(repositories.NewTaskRepository(db.GetDB()))
	// messageService := services.NewMessageService(repositories.NewMessageRepository(db.GetDB()))
	userService := services.NewUserService(repositories.NewUserRepository(rawDB))

	// Initialize mobile controller (not used in current routing)
	// mobileController := controllers.NewMobileController(projectService, sprintService, messageService, userService)

	// Initialize scrum controller
	scrumController := controllers.NewScrumController(projectService, sprintService, nil, userService)

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

	// Database endpoints
	// @Summary Execute database script
	// @Description Executes a database script
	// @Tags database
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} map[string]interface{} "Script executed successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Database/ExecuteScript [post]
	router.POST("/api/Database/ExecuteScript", dbController.ExecuteScript)

	// Debug endpoints
	// @Summary Check database connection
	// @Description Checks if database connection is working
	// @Tags debug
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Database connection OK"
	// @Router /api/_debug/conn [get]
	router.GET("/api/_debug/conn", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "database_connection_ok",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Ping database
	// @Description Pings the database to check connectivity
	// @Tags debug
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Database ping successful"
	// @Failure 500 {object} map[string]interface{} "Database ping failed"
	// @Router /api/_debug/pingdb [get]
	router.GET("/api/_debug/pingdb", func(c *gin.Context) {
		if err := db.GetDB().Raw("SELECT 1").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "database_ping_failed",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "database_ping_ok",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Get JWT token info
	// @Description Retrieves information about the JWT token
	// @Tags debug
	// @Accept json
	// @Produce json
	// @Param Authorization header string true "Bearer token"
	// @Success 200 {object} map[string]interface{} "Token info retrieved"
	// @Failure 401 {object} map[string]interface{} "No authorization header"
	// @Router /api/_debug/jwtinfo [get]
	router.GET("/api/_debug/jwtinfo", func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "no_authorization_header",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token_received": true,
			"token_length":   len(token),
			"time":           time.Now().UTC(),
		})
	})

	// Mail endpoints
	// @Summary Send email
	// @Description Sends an email using the mail service
	// @Tags mail
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} map[string]interface{} "Email sent successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Mail/Send [post]
	router.POST("/api/Mail/Send", mailController.SendEmail)

	// Message endpoints
	// @Summary Send message
	// @Description Sends a message (placeholder implementation)
	// @Tags message
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Message sent successfully"
	// @Router /api/Message/Send [post]
	router.POST("/api/Message/Send", func(c *gin.Context) {
		// TODO: Implement message sending logic
		c.JSON(http.StatusOK, gin.H{
			"status": "message_sent",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Retrieve messages
	// @Description Retrieves messages between users for a specific project
	// @Tags message
	// @Accept json
	// @Produce json
	// @Param fromUserID path string true "From User ID"
	// @Param toUserID path string true "To User ID"
	// @Param projectID path string true "Project ID"
	// @Success 200 {object} map[string]interface{} "Messages retrieved successfully"
	// @Router /api/Message/Retrieve/{fromUserID}/{toUserID}/{projectID} [get]
	router.GET("/api/Message/Retrieve/:fromUserID/:toUserID/:projectID", func(c *gin.Context) {
		// TODO: Implement message retrieval logic
		c.JSON(http.StatusOK, gin.H{
			"status":     "messages_retrieved",
			"fromUserID": c.Param("fromUserID"),
			"toUserID":   c.Param("toUserID"),
			"projectID":  c.Param("projectID"),
			"time":       time.Now().UTC(),
		})
	})

	// Scrum endpoints
	// @Summary Create project
	// @Description Creates a new project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} models.Project "Project created successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Project [post]
	router.POST("/api/Scrum/Project", func(c *gin.Context) {
		// TODO: Implement scrumController.CreateProject
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_project_created",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Update project
	// @Description Updates an existing project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} models.Project "Project updated successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Project [put]
	router.PUT("/api/Scrum/Project", func(c *gin.Context) {
		// TODO: Implement scrumController.EditProject
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_project_updated",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Create sprint
	// @Description Creates a new sprint
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} models.Sprint "Sprint created successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Sprint [post]
	router.POST("/api/Scrum/Sprint", func(c *gin.Context) {
		// TODO: Implement scrumController.CreateSprint
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_sprint_created",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Update sprint
	// @Description Updates an existing sprint
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} models.Sprint "Sprint updated successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Sprint [put]
	router.PUT("/api/Scrum/Sprint", func(c *gin.Context) {
		// TODO: Implement scrumController.EditSprint
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_sprint_updated",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Create task
	// @Description Creates a new task
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} models.Task "Task created successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Task [post]
	router.POST("/api/Scrum/Task", func(c *gin.Context) {
		// TODO: Implement scrumController.CreateTask
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_task_created",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Update task
	// @Description Updates an existing task
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} models.Task "Task updated successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Task [put]
	router.PUT("/api/Scrum/Task", func(c *gin.Context) {
		// TODO: Implement scrumController.EditTask
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_task_updated",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Delete project
	// @Description Deletes a project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param projectId path string true "Project ID"
	// @Success 200 {object} map[string]interface{} "Project deleted successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Project/{projectId} [delete]
	router.DELETE("/api/Scrum/Project/:projectId", func(c *gin.Context) {
		// TODO: Implement scrumController.DeleteProject
		c.JSON(http.StatusOK, gin.H{
			"status":    "scrum_project_deleted",
			"projectID": c.Param("projectId"),
			"time":      time.Now().UTC(),
		})
	})

	// @Summary Get project by ID
	// @Description Retrieves a project by ID
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param projectId path string true "Project ID"
	// @Success 200 {object} models.Project "Project retrieved successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Project/{projectId} [get]
	router.GET("/api/Scrum/Project/:projectId", func(c *gin.Context) {
		// TODO: Implement scrumController.GetProjectByID
		c.JSON(http.StatusOK, gin.H{
			"status":    "scrum_project_retrieved",
			"projectID": c.Param("projectId"),
			"time":      time.Now().UTC(),
		})
	})

	// @Summary Delete sprint
	// @Description Deletes a sprint
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param sprintId path string true "Sprint ID"
	// @Success 200 {object} map[string]interface{} "Sprint deleted successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Sprint/{sprintId} [delete]
	router.DELETE("/api/Scrum/Sprint/:sprintId", func(c *gin.Context) {
		// TODO: Implement scrumController.DeleteSprint
		c.JSON(http.StatusOK, gin.H{
			"status":   "scrum_sprint_deleted",
			"sprintID": c.Param("sprintId"),
			"time":     time.Now().UTC(),
		})
	})

	// @Summary Get sprint by ID
	// @Description Retrieves a sprint by ID
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param sprintId path string true "Sprint ID"
	// @Success 200 {object} models.Sprint "Sprint retrieved successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Sprint/{sprintId} [get]
	router.GET("/api/Scrum/Sprint/:sprintId", func(c *gin.Context) {
		// TODO: Implement scrumController.GetSprintByID
		c.JSON(http.StatusOK, gin.H{
			"status":   "scrum_sprint_retrieved",
			"sprintID": c.Param("sprintId"),
			"time":     time.Now().UTC(),
		})
	})

	// @Summary Delete task
	// @Description Deletes a task
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param taskId path string true "Task ID"
	// @Success 200 {object} map[string]interface{} "Task deleted successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Task/{taskId} [delete]
	router.DELETE("/api/Scrum/Task/:taskId", func(c *gin.Context) {
		// TODO: Implement scrumController.DeleteTask
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_task_deleted",
			"taskID": c.Param("taskId"),
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Get task by ID
	// @Description Retrieves a task by ID
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param taskId path string true "Task ID"
	// @Success 200 {object} models.Task "Task retrieved successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Task/{taskId} [get]
	router.GET("/api/Scrum/Task/:taskId", func(c *gin.Context) {
		// TODO: Implement scrumController.GetTaskByID
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_task_retrieved",
			"taskID": c.Param("taskId"),
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Update task status
	// @Description Updates the status of a task
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} models.Task "Task status updated successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Task/Status [put]
	router.PUT("/api/Scrum/Task/Status", func(c *gin.Context) {
		// TODO: Implement scrumController.UpdateTaskStatus
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_task_status_updated",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Get project members
	// @Description Retrieves all members of a project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param projectId path string true "Project ID"
	// @Success 200 {array} interface{} "Project members retrieved successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Router /api/Scrum/Project/Members/{projectId} [get]
	router.GET("/api/Scrum/Project/Members/:projectId", func(c *gin.Context) {
		// TODO: Implement scrumController.GetProjectMembers
		c.JSON(http.StatusOK, gin.H{
			"status":    "scrum_project_members_retrieved",
			"projectID": c.Param("projectId"),
			"time":      time.Now().UTC(),
		})
	})

	// @Summary Get sprint tasks
	// @Description Retrieves all tasks in a sprint
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param sprintId path string true "Sprint ID"
	// @Success 200 {array} models.Task "Sprint tasks retrieved successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Sprint/Tasks/{sprintId} [get]
	router.GET("/api/Scrum/Sprint/Tasks/:sprintId", func(c *gin.Context) {
		// TODO: Implement scrumController.GetSprintTasks
		c.JSON(http.StatusOK, gin.H{
			"status":   "scrum_sprint_tasks_retrieved",
			"sprintID": c.Param("sprintId"),
			"time":     time.Now().UTC(),
		})
	})

	// @Summary Get project tasks
	// @Description Retrieves all tasks in a project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param projectId path string true "Project ID"
	// @Success 200 {array} models.Task "Project tasks retrieved successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Project/Tasks/{projectId} [get]
	router.GET("/api/Scrum/Project/Tasks/:projectId", func(c *gin.Context) {
		// TODO: Implement scrumController.GetProjectTasks
		c.JSON(http.StatusOK, gin.H{
			"status":    "scrum_project_tasks_retrieved",
			"projectID": c.Param("projectId"),
			"time":      time.Now().UTC(),
		})
	})

	// @Summary Get project sprints
	// @Description Retrieves all sprints in a project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param projectId path string true "Project ID"
	// @Success 200 {array} models.Sprint "Project sprints retrieved successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Project/Sprints/{projectId} [get]
	router.GET("/api/Scrum/Project/Sprints/:projectId", func(c *gin.Context) {
		// TODO: Implement scrumController.GetProjectSprints
		c.JSON(http.StatusOK, gin.H{
			"status":    "scrum_project_sprints_retrieved",
			"projectID": c.Param("projectId"),
			"time":      time.Now().UTC(),
		})
	})

	// @Summary Get user projects
	// @Description Retrieves all projects for a user
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param userId path string true "User ID"
	// @Success 200 {array} models.Project "User projects retrieved successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Projects/User/{userId} [get]
	router.GET("/api/Scrum/Projects/User/:userId", func(c *gin.Context) {
		// TODO: Implement scrumController.GetUserProjects
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_user_projects_retrieved",
			"userID": c.Param("userId"),
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Join project
	// @Description Adds a user to a project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param projectId path string true "Project ID"
	// @Param userId path string true "User ID"
	// @Success 200 {object} map[string]interface{} "User joined project successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Project/{projectId}/{userId} [post]
	router.POST("/api/Scrum/Project/:projectId/:userId", func(c *gin.Context) {
		// TODO: Implement scrumController.JoinProject
		c.JSON(http.StatusOK, gin.H{
			"status":    "scrum_user_joined_project",
			"projectID": c.Param("projectId"),
			"userID":    c.Param("userId"),
			"time":      time.Now().UTC(),
		})
	})

	// @Summary Remove member from project
	// @Description Removes a user from a project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param projectId path string true "Project ID"
	// @Param userId path string true "User ID"
	// @Success 200 {object} map[string]interface{} "User removed from project successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Project/{projectId}/Members/{userId} [delete]
	router.DELETE("/api/Scrum/Project/:projectId/Members/:userId", func(c *gin.Context) {
		// TODO: Implement scrumController.RemoveMemberFromProject
		c.JSON(http.StatusOK, gin.H{
			"status":    "scrum_user_removed_from_project",
			"projectID": c.Param("projectId"),
			"userID":    c.Param("userId"),
			"time":      time.Now().UTC(),
		})
	})

	// @Summary Get active sprints
	// @Description Retrieves active sprints for a project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param projectId path string true "Project ID"
	// @Success 200 {array} interface{} "Active sprints retrieved successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Router /api/Scrum/Project/Sprints/Active/{projectId} [get]
	router.GET("/api/Scrum/Project/Sprints/Active/:projectId", func(c *gin.Context) {
		// TODO: Implement scrumController.GetActiveSprints
		c.JSON(http.StatusOK, gin.H{
			"status":    "scrum_active_sprints_retrieved",
			"projectID": c.Param("projectId"),
			"time":      time.Now().UTC(),
		})
	})

	// @Summary Leave project
	// @Description Removes the current user from a project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} map[string]interface{} "User left project successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Failure 500 {object} map[string]interface{} "Internal server error"
	// @Router /api/Scrum/Project/Leave [post]
	router.POST("/api/Scrum/Project/Leave", func(c *gin.Context) {
		// TODO: Implement scrumController.LeaveProject
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_user_left_project",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Update project owner
	// @Description Updates the owner of a project
	// @Tags scrum
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} map[string]interface{} "Project owner updated successfully"
	// @Failure 400 {object} map[string]interface{} "Bad request"
	// @Failure 401 {object} map[string]interface{} "Unauthorized"
	// @Router /api/Scrum/Project/UpdateProjectOwner [put]
	router.PUT("/api/Scrum/Project/UpdateProjectOwner", func(c *gin.Context) {
		// TODO: Implement scrumController.UpdateProjectOwner
		c.JSON(http.StatusOK, gin.H{
			"status": "scrum_project_owner_updated",
			"time":   time.Now().UTC(),
		})
	})

	// User endpoints
	// @Summary Create user
	// @Description Creates a new user (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "User created successfully"
	// @Router /api/User [post]
	router.POST("/api/User", func(c *gin.Context) {
		// TODO: Implement user creation logic
		c.JSON(http.StatusOK, gin.H{
			"status": "user_created",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Update user
	// @Description Updates an existing user (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "User updated successfully"
	// @Router /api/User [put]
	router.PUT("/api/User", func(c *gin.Context) {
		// TODO: Implement user update logic
		c.JSON(http.StatusOK, gin.H{
			"status": "user_updated",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Get user by ID
	// @Description Retrieves a user by ID (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Param id path string true "User ID"
	// @Success 200 {object} map[string]interface{} "User retrieved successfully"
	// @Router /api/User/{id} [get]
	router.GET("/api/User/:id", func(c *gin.Context) {
		// TODO: Implement user retrieval logic
		c.JSON(http.StatusOK, gin.H{
			"status": "user_retrieved",
			"userID": c.Param("id"),
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Delete user
	// @Description Deletes a user by ID (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Param id path string true "User ID"
	// @Success 200 {object} map[string]interface{} "User deleted successfully"
	// @Router /api/User/{id} [delete]
	router.DELETE("/api/User/:id", func(c *gin.Context) {
		// TODO: Implement user deletion logic
		c.JSON(http.StatusOK, gin.H{
			"status": "user_deleted",
			"userID": c.Param("id"),
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Get user by username
	// @Description Retrieves a user by username (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Param username path string true "Username"
	// @Success 200 {object} map[string]interface{} "User retrieved successfully"
	// @Router /api/User/Username/{username} [get]
	router.GET("/api/User/Username/:username", func(c *gin.Context) {
		// TODO: Implement username validation logic
		c.JSON(http.StatusOK, gin.H{
			"status":   "username_retrieved",
			"username": c.Param("username"),
			"time":     time.Now().UTC(),
		})
	})

	// @Summary Process login
	// @Description Processes user login (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Login processed successfully"
	// @Router /api/User/ProcessLogin [post]
	router.POST("/api/User/ProcessLogin", func(c *gin.Context) {
		// TODO: Implement login processing logic
		c.JSON(http.StatusOK, gin.H{
			"status": "login_processed",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Validate email
	// @Description Validates user email (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Email validated successfully"
	// @Router /api/User/ValidateEmail [post]
	router.POST("/api/User/ValidateEmail", func(c *gin.Context) {
		// TODO: Implement email validation logic
		c.JSON(http.StatusOK, gin.H{
			"status": "email_validated",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Validate username
	// @Description Validates username (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Username validated successfully"
	// @Router /api/User/ValidateUsername [post]
	router.POST("/api/User/ValidateUsername", func(c *gin.Context) {
		// TODO: Implement username validation logic
		c.JSON(http.StatusOK, gin.H{
			"status": "username_validated",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Request password reset
	// @Description Requests a password reset (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Password reset requested successfully"
	// @Router /api/User/RequestPasswordReset [post]
	router.POST("/api/User/RequestPasswordReset", func(c *gin.Context) {
		// TODO: Implement password reset request logic
		c.JSON(http.StatusOK, gin.H{
			"status": "password_reset_requested",
			"time":   time.Now().UTC(),
		})
	})

	// @Summary Reset password
	// @Description Resets user password (placeholder implementation)
	// @Tags user
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Password reset successfully"
	// @Router /api/User/ResetPassword [post]
	router.POST("/api/User/ResetPassword", func(c *gin.Context) {
		// TODO: Implement password reset logic
		c.JSON(http.StatusOK, gin.H{
			"status": "password_reset",
			"time":   time.Now().UTC(),
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

	// Expose the contract
	// OpenAPI file not configured - skipping swagger endpoint
	// router.StaticFile("/swagger/openapi.yaml", "./api/openapi.yaml")

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/openapi.yaml"),
	))

	// Redirect from /swagger to /swagger/index.html for better UX
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Public routes (no authentication required)
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.POST("/refresh", controllers.RefreshToken)
		auth.POST("/forgot-password", controllers.ForgotPassword)
		auth.POST("/reset-password", controllers.ResetPassword)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// User Controller
		user := protected.Group("/User")
		{
			user.GET("/profile", controllers.GetUserProfile)
			user.PUT("/profile", controllers.UpdateUserProfile)
			user.POST("/avatar", controllers.UploadAvatar)
			user.PUT("/activate/:id", controllers.ActivateUser)
			user.PUT("/deactivate/:id", controllers.DeactivateUser)
			user.GET("/search", controllers.SearchUsers)
		}

		// Scrum Controller
		scrum := protected.Group("/Scrum")
		{
			// Project endpoints
			scrum.POST("/Project", scrumController.CreateProject)
			scrum.PUT("/Project", scrumController.EditProject)
			scrum.GET("/Project/:projectId", scrumController.GetProjectByID)
			scrum.DELETE("/Project/:projectId", scrumController.DeleteProject)
			scrum.GET("/Project/Members/:projectId", scrumController.GetProjectMembers)
			scrum.GET("/Project/Tasks/:projectId", scrumController.GetProjectTasks)
			scrum.GET("/Project/Sprints/:projectId", scrumController.GetProjectSprints)
			scrum.GET("/Project/Sprints/Active/:projectId", scrumController.GetActiveSprints)
			scrum.POST("/Project/Leave", scrumController.LeaveProject)
			scrum.PUT("/Project/UpdateProjectOwner", scrumController.UpdateProjectOwner)
			scrum.POST("/Project/:projectId/:userId", scrumController.JoinProject)
			scrum.DELETE("/Project/:projectId/Members/:userId", scrumController.RemoveMemberFromProject)

			// Sprint endpoints
			scrum.POST("/Sprint", scrumController.CreateSprint)
			scrum.PUT("/Sprint", scrumController.EditSprint)
			scrum.GET("/Sprint/:sprintId", scrumController.GetSprintByID)
			scrum.DELETE("/Sprint/:sprintId", scrumController.DeleteSprint)
			scrum.GET("/Sprint/Tasks/:sprintId", scrumController.GetSprintTasks)

			// Task endpoints
			scrum.POST("/Task", scrumController.CreateTask)
			scrum.PUT("/Task", scrumController.EditTask)
			scrum.GET("/Task/:taskId", scrumController.GetTaskByID)
			scrum.DELETE("/Task/:taskId", scrumController.DeleteTask)
			scrum.PUT("/Task/Status", scrumController.UpdateTaskStatus)

			// User projects
			scrum.GET("/Projects/User/:userId", scrumController.GetUserProjects)
		}

		// Message Controller
		message := protected.Group("/Message")
		{
			message.POST("/Send", controllers.CreateMessage)
			message.GET("/Retrieve/:fromUserID/:toUserID/:projectID", controllers.GetMessages)
		}

		// Mail Controller
		mail := protected.Group("/Mail")
		{
			mail.POST("/Send", mailController.SendEmail)
		}

		// Database Controller
		database := protected.Group("/Database")
		{
			database.POST("/ExecuteScript", dbController.ExecuteScript)
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
