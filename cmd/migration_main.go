//go:build migration
// +build migration

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
	"devhive-backend/db"
	ws "devhive-backend/internal"
	"devhive-backend/internal/migration"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Migration-compatible main that maintains exact .NET API contract
// This implements the strangler pattern for zero-downtime migration
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

	// Initialize migration router with exact .NET contract compatibility
	r := chi.NewRouter()

	// Middleware stack (order matters - matches .NET pipeline)
	r.Use(middleware.RequestID) // Maps to x-trace-id equivalent
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	// CORS - mirror current .NET CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health endpoint - exact same response as current
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"DevHive Backend","time":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
	})

	// Initialize migration handlers
	migrationHandlers := migration.NewHandlers()

	// API routes - EXACT contract matching .NET backend
	// All routes under /api/... must remain identical for frontend compatibility
	r.Route("/api", func(api chi.Router) {
		// Debug endpoints (keep or gate behind auth/role)
		api.Route("/_debug", func(debug chi.Router) {
			debug.Get("/conn", migrationHandlers.DebugConn)
			debug.Get("/pingdb", migrationHandlers.DebugPingDB)
			debug.Get("/jwtinfo", migrationHandlers.DebugJWTInfo)
		})

		// Database endpoints
		api.Route("/Database", func(db chi.Router) {
			db.Post("/ExecuteScript", migrationHandlers.ExecuteScript)
		})

		// Mail endpoints
		api.Route("/Mail", func(mail chi.Router) {
			mail.Post("/Send", migrationHandlers.SendEmail)
		})

		// Message endpoints - maintain Firestore initially
		api.Route("/Message", func(msg chi.Router) {
			msg.Post("/Send", migrationHandlers.SendMessage)
			msg.Get("/Retrieve/{fromUserID}/{toUserID}/{projectID}", migrationHandlers.RetrieveMessages)
		})

		// User endpoints - exact .NET contract
		api.Route("/User", func(user chi.Router) {
			user.Post("/", migrationHandlers.CreateUser)
			user.Put("/", migrationHandlers.UpdateUser)
			user.Get("/{id}", migrationHandlers.GetUserByID)
			user.Delete("/{id}", migrationHandlers.DeleteUser)
			user.Get("/Username/{username}", migrationHandlers.GetUserByUsername)
			user.Post("/ProcessLogin", migrationHandlers.ProcessLogin)
			user.Post("/ValidateEmail", migrationHandlers.ValidateEmail)
			user.Post("/ValidateUsername", migrationHandlers.ValidateUsername)
			user.Post("/RequestPasswordReset", migrationHandlers.RequestPasswordReset)
			user.Post("/ResetPassword", migrationHandlers.ResetPassword)
		})

		// Scrum endpoints - exact .NET contract
		api.Route("/Scrum", func(scrum chi.Router) {
			// Project endpoints
			scrum.Post("/Project", migrationHandlers.CreateProject)
			scrum.Put("/Project", migrationHandlers.UpdateProject)
			scrum.Get("/Project/{projectId}", migrationHandlers.GetProjectByID)
			scrum.Delete("/Project/{projectId}", migrationHandlers.DeleteProject)
			scrum.Get("/Project/Members/{projectId}", migrationHandlers.GetProjectMembers)
			scrum.Get("/Project/Tasks/{projectId}", migrationHandlers.GetProjectTasks)
			scrum.Get("/Project/Sprints/{projectId}", migrationHandlers.GetProjectSprints)
			scrum.Get("/Project/Sprints/Active/{projectId}", migrationHandlers.GetActiveSprints)
			scrum.Post("/Project/Leave", migrationHandlers.LeaveProject)
			scrum.Put("/Project/UpdateProjectOwner", migrationHandlers.UpdateProjectOwner)
			scrum.Post("/Project/{projectId}/{userId}", migrationHandlers.JoinProject)
			scrum.Delete("/Project/{projectId}/Members/{userId}", migrationHandlers.RemoveMemberFromProject)

			// Sprint endpoints
			scrum.Post("/Sprint", migrationHandlers.CreateSprint)
			scrum.Put("/Sprint", migrationHandlers.UpdateSprint)
			scrum.Get("/Sprint/{sprintId}", migrationHandlers.GetSprintByID)
			scrum.Delete("/Sprint/{sprintId}", migrationHandlers.DeleteSprint)
			scrum.Get("/Sprint/Tasks/{sprintId}", migrationHandlers.GetSprintTasks)

			// Task endpoints
			scrum.Post("/Task", migrationHandlers.CreateTask)
			scrum.Put("/Task", migrationHandlers.UpdateTask)
			scrum.Get("/Task/{taskId}", migrationHandlers.GetTaskByID)
			scrum.Delete("/Task/{taskId}", migrationHandlers.DeleteTask)
			scrum.Put("/Task/Status", migrationHandlers.UpdateTaskStatus)

			// User projects
			scrum.Get("/Projects/User/{userId}", migrationHandlers.GetUserProjects)
		})
	})

	// WebSocket endpoints - exact same paths
	r.Get("/ws", migrationHandlers.HandleWS)
	r.Get("/ws/auth", migrationHandlers.HandleWSAuth)

	// Root redirect to swagger (maintain existing behavior)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	// Swagger documentation (keep for contract verification)
	r.Get("/swagger/*", migrationHandlers.SwaggerHandler)
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Starting DevHive Migration Backend on port %s", port)
		log.Printf("API contract: EXACT .NET compatibility maintained")
		log.Printf("Frontend base URL: https://api.devhive.it.com/api")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down migration server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Migration server exited")
}
