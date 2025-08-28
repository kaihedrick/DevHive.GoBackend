package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"devhive-backend/db"

	"github.com/gin-gonic/gin"
)

// DatabaseController handles database-related operations
type DatabaseController struct {
	scriptsDir string
}

// NewDatabaseController creates a new database controller instance
func NewDatabaseController(scriptsDir string) *DatabaseController {
	return &DatabaseController{
		scriptsDir: scriptsDir,
	}
}

// ExecuteScript executes a database script
// @Summary Execute database script
// @Description Executes a database script by name
// @Tags database
// @Accept json
// @Produce json
// @Param request body map[string]string true "Script execution request"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Script executed successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /database/execute-script [post]
func (dc *DatabaseController) ExecuteScript(c *gin.Context) {
	var request map[string]string
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	scriptName, exists := request["script_name"]
	if !exists || scriptName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Script name is required"})
		return
	}

	// Validate script name to prevent directory traversal
	if strings.Contains(scriptName, "..") || strings.Contains(scriptName, "/") || strings.Contains(scriptName, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid script name"})
		return
	}

	scriptPath := filepath.Join(dc.scriptsDir, scriptName)
	if !strings.HasSuffix(scriptName, ".sql") {
		scriptPath += ".sql"
	}

	// Check if script file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Script not found"})
		return
	}

	// Read script content
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read script file"})
		return
	}

	// Get database connection
	dbConn := db.GetDB()
	if dbConn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
		return
	}

	// Execute script using GORM
	// Note: GORM doesn't have a direct Ping method, so we'll use the underlying sql.DB
	sqlDB, err := dbConn.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Split script into individual statements
	statements := strings.Split(string(scriptContent), ";")
	var results []string

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		// Execute statement using GORM
		result := dbConn.Exec(statement)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Script execution failed",
				"details": result.Error.Error(),
				"results": results,
			})
			return
		}

		results = append(results, fmt.Sprintf("Executed: %s (Rows affected: %d)", statement, result.RowsAffected))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Script executed successfully",
		"results": results,
	})
}

// GetDatabaseStatus returns the current database status
// @Summary Get database status
// @Description Returns the current database connection status
// @Tags database
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Database status"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /database/status [get]
func (dc *DatabaseController) GetDatabaseStatus(c *gin.Context) {
	dbConn := db.GetDB()
	if dbConn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
		return
	}

	// Get underlying sql.DB for status checks
	sqlDB, err := dbConn.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Get database stats
	stats := sqlDB.Stats()

	c.JSON(http.StatusOK, gin.H{
		"status": "connected",
		"stats": gin.H{
			"max_open_connections": stats.MaxOpenConnections,
			"open_connections":     stats.OpenConnections,
			"in_use":               stats.InUse,
			"idle":                 stats.Idle,
		},
	})
}

// ListScripts returns a list of available database scripts
// @Summary List database scripts
// @Description Returns a list of available database scripts
// @Tags database
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} string "List of script names"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /database/scripts [get]
func (dc *DatabaseController) ListScripts(c *gin.Context) {
	// Check if scripts directory exists
	if _, err := os.Stat(dc.scriptsDir); os.IsNotExist(err) {
		c.JSON(http.StatusOK, []string{})
		return
	}

	// Read scripts directory
	entries, err := os.ReadDir(dc.scriptsDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read scripts directory"})
		return
	}

	var scripts []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			scripts = append(scripts, entry.Name())
		}
	}

	c.JSON(http.StatusOK, scripts)
}

// GetScriptContent returns the content of a specific script
// @Summary Get script content
// @Description Returns the content of a specific database script
// @Tags database
// @Accept json
// @Produce text/plain
// @Param scriptName path string true "Script name"
// @Security BearerAuth
// @Success 200 {string} string "Script content"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Script not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /database/scripts/{scriptName} [get]
func (dc *DatabaseController) GetScriptContent(c *gin.Context) {
	scriptName := c.Param("scriptName")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Script name is required"})
		return
	}

	// Validate script name to prevent directory traversal
	if strings.Contains(scriptName, "..") || strings.Contains(scriptName, "/") || strings.Contains(scriptName, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid script name"})
		return
	}

	scriptPath := filepath.Join(dc.scriptsDir, scriptName)
	if !strings.HasSuffix(scriptName, ".sql") {
		scriptPath += ".sql"
	}

	// Check if script file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Script not found"})
		return
	}

	// Read and return script content
	scriptFile, err := os.Open(scriptPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open script file"})
		return
	}
	defer scriptFile.Close()

	c.Header("Content-Type", "text/plain")
	io.Copy(c.Writer, scriptFile)
}
