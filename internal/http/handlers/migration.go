package handlers

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"devhive-backend/internal/http/response"
	"devhive-backend/internal/repo"
)

type MigrationHandler struct {
	queries *repo.Queries
	db      *sql.DB
}

func NewMigrationHandler(queries *repo.Queries, db *sql.DB) *MigrationHandler {
	return &MigrationHandler{
		queries: queries,
		db:      db,
	}
}

// MigrationRequest represents a migration request
type MigrationRequest struct {
	ScriptName string `json:"scriptName"`
	Action     string `json:"action"` // "run", "reset", "list"
}

// MigrationResponse represents a migration response
type MigrationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
}

// RunMigration handles running a specific migration script
func (h *MigrationHandler) RunMigration(w http.ResponseWriter, r *http.Request) {
	var req MigrationRequest
	if !response.Decode(w, r, &req) {
		return
	}

	if req.ScriptName == "" {
		response.BadRequest(w, "script name is required")
		return
	}

	// Read the migration script
	// Try multiple possible paths for migrations directory
	migrationsDirs := []string{
		"cmd/devhive-api/migrations",
		"./cmd/devhive-api/migrations",
		"migrations",
		"./migrations",
	}

	var scriptContent []byte
	var err error
	for _, dir := range migrationsDirs {
		scriptPath := filepath.Join(dir, req.ScriptName)
		scriptContent, err = ioutil.ReadFile(scriptPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		response.BadRequest(w, fmt.Sprintf("failed to read script %s (tried paths: %v): %v", req.ScriptName, migrationsDirs, err))
		return
	}

	// Execute the SQL script
	_, err = h.db.Exec(string(scriptContent))
	if err != nil {
		response.InternalServerError(w, fmt.Sprintf("failed to execute script: %v", err))
		return
	}

	response.JSON(w, http.StatusOK, MigrationResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully executed migration script: %s", req.ScriptName),
	})
}

// ResetDatabase handles resetting the entire database
func (h *MigrationHandler) ResetDatabase(w http.ResponseWriter, r *http.Request) {
	// Read the reset script
	scriptContent, err := ioutil.ReadFile("reset_database.sql")
	if err != nil {
		response.BadRequest(w, fmt.Sprintf("failed to read reset script: %v", err))
		return
	}

	// Execute the reset script
	_, err = h.db.Exec(string(scriptContent))
	if err != nil {
		response.InternalServerError(w, fmt.Sprintf("failed to reset database: %v", err))
		return
	}

	response.JSON(w, http.StatusOK, MigrationResponse{
		Success: true,
		Message: "Database reset completed successfully",
	})
}

// ListMigrations handles listing available migration scripts
func (h *MigrationHandler) ListMigrations(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("migrations")
	if err != nil {
		response.InternalServerError(w, fmt.Sprintf("failed to read migrations directory: %v", err))
		return
	}

	var migrations []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"migrations": migrations,
		"count":      len(migrations),
	})
}

// RebuildAndDeploy handles rebuilding and deploying the application
func (h *MigrationHandler) RebuildAndDeploy(w http.ResponseWriter, r *http.Request) {
	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "devhive-api", "./cmd/devhive-api")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		response.InternalServerError(w, fmt.Sprintf("build failed: %v\nOutput: %s", err, string(buildOutput)))
		return
	}

	// Deploy to Fly.io
	deployCmd := exec.Command("fly", "deploy")
	deployOutput, err := deployCmd.CombinedOutput()
	if err != nil {
		response.InternalServerError(w, fmt.Sprintf("deploy failed: %v\nOutput: %s", err, string(deployOutput)))
		return
	}

	response.JSON(w, http.StatusOK, MigrationResponse{
		Success: true,
		Message: "Application rebuilt and deployed successfully",
		Output:  string(deployOutput),
	})
}

// RunMigrationAndDeploy handles running a migration and then rebuilding/deploying
func (h *MigrationHandler) RunMigrationAndDeploy(w http.ResponseWriter, r *http.Request) {
	var req MigrationRequest
	if !response.Decode(w, r, &req) {
		return
	}

	if req.ScriptName == "" {
		response.BadRequest(w, "script name is required")
		return
	}

	// Read and execute the migration script
	scriptPath := filepath.Join("migrations", req.ScriptName)
	scriptContent, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		response.BadRequest(w, fmt.Sprintf("failed to read script %s: %v", req.ScriptName, err))
		return
	}

	_, err = h.db.Exec(string(scriptContent))
	if err != nil {
		response.InternalServerError(w, fmt.Sprintf("failed to execute script: %v", err))
		return
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "devhive-api", "./cmd/devhive-api")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		response.InternalServerError(w, fmt.Sprintf("migration succeeded but build failed: %v\nOutput: %s", err, string(buildOutput)))
		return
	}

	// Deploy to Fly.io
	deployCmd := exec.Command("fly", "deploy")
	deployOutput, err := deployCmd.CombinedOutput()
	if err != nil {
		response.InternalServerError(w, fmt.Sprintf("migration and build succeeded but deploy failed: %v\nOutput: %s", err, string(deployOutput)))
		return
	}

	response.JSON(w, http.StatusOK, MigrationResponse{
		Success: true,
		Message: fmt.Sprintf("Migration %s executed and application deployed successfully", req.ScriptName),
		Output:  string(deployOutput),
	})
}

// HealthCheck handles checking database connectivity
func (h *MigrationHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping()
	if err != nil {
		response.InternalServerError(w, fmt.Sprintf("database connection failed: %v", err))
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"database":  "connected",
		"timestamp": "2025-09-29T21:15:00Z",
	})
}
