package controllers

import (
	"net/http"
	"strconv"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetTasks retrieves all tasks for a sprint
// @Summary Get sprint tasks
// @Description Retrieve all tasks for a specific sprint with pagination
// @Tags tasks
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param sprintId path string true "Sprint ID" format(uuid)
// @Param limit query int false "Number of tasks to return (max 100)" default(20)
// @Param offset query int false "Number of tasks to skip" default(0)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Tasks with pagination info"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Access denied to project"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{projectId}/sprints/{sprintId}/tasks [get]
func GetTasks(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get sprint ID from URL parameter
	sprintIDStr := c.Param("sprintId")
	sprintID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	// Get query parameters for pagination
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get tasks from database
	tasks, err := models.GetTasksBySprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	// Apply pagination
	total := len(tasks)
	start := offset
	end := offset + limit
	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedTasks := tasks[start:end]

	c.JSON(http.StatusOK, gin.H{
		"tasks": paginatedTasks,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_more": end < total,
		},
	})
}

// CreateTask creates a new task
// @Summary Create a new task
// @Description Create a new task in a specific sprint
// @Tags tasks
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param sprintId path string true "Sprint ID" format(uuid)
// @Param task body models.TaskCreateRequest true "Task creation request"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Task created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Access denied to project"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{projectId}/sprints/{sprintId}/tasks [post]
func CreateTask(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get sprint ID from URL parameter
	sprintIDStr := c.Param("sprintId")
	sprintID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
		return
	}

	// Check if user is the owner of the project
	isOwner, err := models.IsProjectOwner(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can create tasks"})
		return
	}

	var req models.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create task
	task, err := models.CreateTask(db.GetDB(), req, sprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"task":    task,
		"message": "Task created successfully",
	})
}

// GetTask retrieves a specific task
// @Summary Get a specific task
// @Description Retrieve a specific task by ID from a sprint
// @Tags tasks
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param sprintId path string true "Sprint ID" format(uuid)
// @Param id path string true "Task ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} models.Task "Task details"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Access denied to project"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{projectId}/sprints/{sprintId}/tasks/{id} [get]
func GetTask(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get sprint ID from URL parameter
	sprintIDStr := c.Param("sprintId")
	sprintID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
		return
	}

	// Get task ID from URL parameter
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	// Get task details
	task, err := models.GetTask(db.GetDB(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Verify task belongs to the sprint
	if task.SprintID != sprintID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found in this sprint"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": task,
	})
}

// UpdateTask updates a task
// @Summary Update a task
// @Description Update an existing task in a sprint
// @Tags tasks
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param sprintId path string true "Sprint ID" format(uuid)
// @Param id path string true "Task ID" format(uuid)
// @Param task body models.TaskUpdateRequest true "Task update request"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Task updated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Access denied to project"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{projectId}/sprints/{sprintId}/tasks/{id} [put]
func UpdateTask(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get sprint ID from URL parameter
	sprintIDStr := c.Param("sprintId")
	sprintID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
		return
	}

	// Get task ID from URL parameter
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Check if user is the owner of the project
	isOwner, err := models.IsProjectOwner(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can update tasks"})
		return
	}

	// Verify task belongs to the sprint
	task, err := models.GetTask(db.GetDB(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.SprintID != sprintID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found in this sprint"})
		return
	}

	var req models.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update task
	updatedTask, err := models.UpdateTask(db.GetDB(), taskID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task":    updatedTask,
		"message": "Task updated successfully",
	})
}

// DeleteTask deletes a task
// @Summary Delete a task
// @Description Delete a task from a sprint
// @Tags tasks
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param sprintId path string true "Sprint ID" format(uuid)
// @Param id path string true "Task ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} map[string]string "Task deleted successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Access denied to project"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{projectId}/sprints/{sprintId}/tasks/{id} [delete]
func DeleteTask(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get sprint ID from URL parameter
	sprintIDStr := c.Param("sprintId")
	sprintID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
		return
	}

	// Get task ID from URL parameter
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Check if user is the owner of the project
	isOwner, err := models.IsProjectOwner(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can delete tasks"})
		return
	}

	// Verify task belongs to the sprint
	task, err := models.GetTask(db.GetDB(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.SprintID != sprintID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found in this sprint"})
		return
	}

	// Delete task
	err = models.DeleteTask(db.GetDB(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}
