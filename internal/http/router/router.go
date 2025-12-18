package router

import (
	"database/sql"
	"time"

	"devhive-backend/internal/config"
	"devhive-backend/internal/http/handlers"
	"devhive-backend/internal/http/middleware"
	"devhive-backend/internal/repo"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

// Setup creates and configures the HTTP router
func Setup(cfg *config.Config, queries *repo.Queries, db interface{}) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Logger)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           300,
	}))

	// Health check endpoints
	r.Get("/health", handlers.HealthCheck)
	r.Get("/healthz", handlers.LivenessCheck)
	r.Get("/readyz", handlers.ReadinessCheck(db.(*sql.DB)))

	// API routes
	r.Route("/api", func(api chi.Router) {
		// Mount v1 API
		api.Mount("/v1", setupV1Routes(cfg, queries, db))
	})

	return r
}

// setupV1Routes configures the v1 API routes
func setupV1Routes(cfg *config.Config, queries *repo.Queries, db interface{}) chi.Router {
	r := chi.NewRouter()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg, queries)
	userHandler := handlers.NewUserHandler(queries)
	projectHandler := handlers.NewProjectHandler(queries)
	sprintHandler := handlers.NewSprintHandler(queries)
	taskHandler := handlers.NewTaskHandler(queries)
	messageHandler := handlers.NewMessageHandler(queries)
	mailHandler := handlers.NewMailHandler(cfg)
	migrationHandler := handlers.NewMigrationHandler(queries, db.(*sql.DB))

	// Auth routes (public)
	r.Route("/auth", func(auth chi.Router) {
		auth.Post("/login", authHandler.Login)
		auth.Post("/refresh", authHandler.Refresh)
		auth.Post("/password/reset-request", authHandler.RequestPasswordReset)
		auth.Post("/password/reset", authHandler.ResetPassword)
	})

	// User routes
	r.Route("/users", func(users chi.Router) {
		users.Post("/", userHandler.CreateUser)
		users.Get("/validate-email", userHandler.ValidateEmail)
		users.Post("/validate-email", userHandler.ValidateEmail)
		users.Get("/validate-username", userHandler.ValidateUsername)
		users.Post("/validate-username", userHandler.ValidateUsername)
		users.With(middleware.RequireAuth(cfg.JWT.SigningKey)).Get("/me", userHandler.GetMe)
		users.With(middleware.RequireAuth(cfg.JWT.SigningKey)).Get("/{userId}", userHandler.GetUser)
	})

	// Project routes
	r.Route("/projects", func(projects chi.Router) {
		projects.Use(middleware.RequireAuth(cfg.JWT.SigningKey))
		projects.Get("/", projectHandler.ListProjects)
		projects.Post("/", projectHandler.CreateProject)
		// Join by project code/ID (must be defined before /{projectId} routes)
		projects.Post("/join", projectHandler.JoinProject)
		projects.Get("/{projectId}", projectHandler.GetProject)
		projects.Get("/{projectId}/bundle", projectHandler.GetProjectBundle)
		projects.Patch("/{projectId}", projectHandler.UpdateProject)
		projects.Delete("/{projectId}", projectHandler.DeleteProject)

		// Project members
		projects.Get("/{projectId}/members", projectHandler.ListMembers)
		projects.Put("/{projectId}/members/{userId}", projectHandler.AddMember)
		projects.Delete("/{projectId}/members/{userId}", projectHandler.RemoveMember)

		// Project sprints
		projects.Get("/{projectId}/sprints", sprintHandler.ListSprintsByProject)
		projects.Post("/{projectId}/sprints", sprintHandler.CreateSprint)

		// Project tasks
		projects.Get("/{projectId}/tasks", taskHandler.ListTasksByProject)
		projects.Post("/{projectId}/tasks", taskHandler.CreateTask)

		// Project messages
		projects.Get("/{projectId}/messages", messageHandler.ListMessagesByProject)
		projects.Post("/{projectId}/messages", messageHandler.CreateMessage)
	})

	// Sprint routes
	r.Route("/sprints", func(sprints chi.Router) {
		sprints.Use(middleware.RequireAuth(cfg.JWT.SigningKey))
		sprints.Get("/{sprintId}", sprintHandler.GetSprint)
		sprints.Patch("/{sprintId}", sprintHandler.UpdateSprint)
		sprints.Delete("/{sprintId}", sprintHandler.DeleteSprint)
		sprints.Get("/{sprintId}/tasks", taskHandler.ListTasksBySprint)
	})

	// Task routes
	r.Route("/tasks", func(tasks chi.Router) {
		tasks.Use(middleware.RequireAuth(cfg.JWT.SigningKey))
		tasks.Get("/{taskId}", taskHandler.GetTask)
		tasks.Patch("/{taskId}", taskHandler.UpdateTask)
		tasks.Patch("/{taskId}/status", taskHandler.UpdateTaskStatus)
		tasks.Delete("/{taskId}", taskHandler.DeleteTask)
	})

	// Message routes
	r.Route("/messages", func(messages chi.Router) {
		messages.Use(middleware.RequireAuth(cfg.JWT.SigningKey))
		messages.Post("/", messageHandler.CreateMessage)
		messages.Get("/", messageHandler.ListMessages)
		messages.Get("/ws", messageHandler.WebSocketHandler)
	})

	// Mail routes
	r.Route("/mail", func(mail chi.Router) {
		mail.Use(middleware.RequireAuth(cfg.JWT.SigningKey))
		mail.Post("/send", mailHandler.SendEmail)
	})

	// Migration routes (admin only - no auth for now, but should be protected in production)
	r.Route("/migrations", func(migrations chi.Router) {
		migrations.Post("/run", migrationHandler.RunMigration)
		migrations.Post("/reset", migrationHandler.ResetDatabase)
		migrations.Get("/list", migrationHandler.ListMigrations)
		migrations.Post("/rebuild-deploy", migrationHandler.RebuildAndDeploy)
		migrations.Post("/run-and-deploy", migrationHandler.RunMigrationAndDeploy)
		migrations.Get("/health", migrationHandler.HealthCheck)
	})

	// Legacy route shims (temporary for backward compatibility)
	r.Get("/Scrum/Sprint/{id}", sprintHandler.GetSprint)
	r.Put("/Scrum/Sprint/{id}", sprintHandler.UpdateSprint)
	r.Delete("/Scrum/Sprint/{id}", sprintHandler.DeleteSprint)
	r.Get("/Scrum/Sprint/{id}/Tasks", taskHandler.ListTasksBySprint)
	r.Get("/Scrum/Project/{id}/Tasks", taskHandler.ListTasksByProject)
	r.Get("/Scrum/Task/{id}", taskHandler.GetTask)
	r.Patch("/Scrum/Task/{id}", taskHandler.UpdateTask)
	r.Delete("/Scrum/Task/{id}", taskHandler.DeleteTask)

	return r
}

// setupLegacyRoutes configures legacy API routes for backward compatibility
func setupLegacyRoutes(cfg *config.Config, queries *repo.Queries) chi.Router {
	r := chi.NewRouter()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg, queries)
	userHandler := handlers.NewUserHandler(queries)
	projectHandler := handlers.NewProjectHandler(queries)
	sprintHandler := handlers.NewSprintHandler(queries)
	taskHandler := handlers.NewTaskHandler(queries)
	messageHandler := handlers.NewMessageHandler(queries)
	mailHandler := handlers.NewMailHandler(cfg)

	// Legacy route mappings
	r.Post("/User/ProcessLogin", authHandler.Login)
	r.Post("/User/Register", userHandler.CreateUser)
	r.Get("/User/{id}", userHandler.GetUser)
	r.Post("/User/RequestPasswordReset", authHandler.RequestPasswordReset)
	r.Post("/User/ResetPassword", authHandler.ResetPassword)

	r.Post("/Scrum/Project", projectHandler.CreateProject)
	r.Get("/Scrum/Project/{id}", projectHandler.GetProject)
	r.Put("/Scrum/Project/{id}", projectHandler.UpdateProject)
	r.Delete("/Scrum/Project/{id}", projectHandler.DeleteProject)
	r.Get("/Scrum/Project/{id}/Sprints", sprintHandler.ListSprintsByProject)
	r.Get("/Scrum/Project/{id}/Tasks", taskHandler.ListTasksByProject)

	r.Post("/Scrum/Task", taskHandler.CreateTask)
	r.Get("/Scrum/Task/{id}", taskHandler.GetTask)
	r.Patch("/Scrum/Task/{id}", taskHandler.UpdateTask)
	r.Patch("/Scrum/Task/{id}/status", taskHandler.UpdateTaskStatus)
	r.Delete("/Scrum/Task/{id}", taskHandler.DeleteTask)

	r.Get("/Scrum/Sprint/{id}", sprintHandler.GetSprint)
	r.Put("/Scrum/Sprint/{id}", sprintHandler.UpdateSprint)
	r.Delete("/Scrum/Sprint/{id}", sprintHandler.DeleteSprint)
	r.Get("/Scrum/Sprint/{id}/Tasks", taskHandler.ListTasksBySprint)

	r.Post("/Message/Send", messageHandler.CreateMessage)
	r.Get("/Message/Retrieve/{fromUserId}/{toUserId}/{projectId}", messageHandler.ListMessages)

	r.Post("/Mail/Send", mailHandler.SendEmail)

	return r
}
