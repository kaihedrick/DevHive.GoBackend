package controllers

import (
	"net/http"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetTasks retrieves all tasks for a specific project
// @Summary Get project tasks
// @Description Retrieves all tasks for a specific project
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Security BearerAuth
// @Success 200 {array} models.Task "List of tasks"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/tasks [get]
func GetTasks(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID",
		})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied to project",
		})
		return
	}

	tasks, err := models.GetTasks(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve tasks",
		})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// CreateTask creates a new task
// @Summary Create task
// @Description Creates a new task for a specific project
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param task body models.TaskCreateRequest true "Task to create"
// @Security BearerAuth
// @Success 201 {object} models.Task "Task created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/tasks [post]
func CreateTask(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID",
		})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied to project",
		})
		return
	}

	var req models.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	task, err := models.CreateTask(db.GetDB(), req, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create task",
		})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask retrieves a specific task by ID
// @Summary Get task
// @Description Retrieves a specific task by ID if the user has access
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param taskId path string true "Task ID"
// @Security BearerAuth
// @Success 200 {object} models.Task "Task details"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Task not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/tasks/{taskId} [get]
func GetTask(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID",
		})
		return
	}

	taskIDStr := c.Param("taskId")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid task ID",
		})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied to project",
		})
		return
	}

	task, err := models.GetTask(db.GetDB(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Task not found",
		})
		return
	}

	// Verify task belongs to the project
	if task.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Task not found in this project",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask updates an existing task
// @Summary Update task
// @Description Updates an existing task if the user has permission
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param taskId path string true "Task ID"
// @Param task body models.TaskUpdateRequest true "Updated task data"
// @Security BearerAuth
// @Success 200 {object} models.Task "Task updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Task not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/tasks/{taskId} [put]
func UpdateTask(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID",
		})
		return
	}

	taskIDStr := c.Param("taskId")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid task ID",
		})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied to project",
		})
		return
	}

	var req models.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	task, err := models.UpdateTask(db.GetDB(), taskID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update task",
		})
		return
	}

	// Verify task belongs to the project
	if task.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Task not found in this project",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask deletes a task
// @Summary Delete task
// @Description Deletes a task if the user has permission
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param taskId path string true "Task ID"
// @Security BearerAuth
// @Success 204 "Task deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Task not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/tasks/{taskId} [delete]
func DeleteTask(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID",
		})
		return
	}

	taskIDStr := c.Param("taskId")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid task ID",
		})
		return
	}

	// Check if user has permission to delete tasks
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if userRole != "owner" && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Insufficient permissions to delete tasks",
		})
		return
	}

	err = models.DeleteTask(db.GetDB(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete task",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTasksBySprint retrieves all tasks for a specific sprint
// @Summary Get sprint tasks
// @Description Retrieves all tasks for a specific sprint
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param sprintId path string true "Sprint ID"
// @Security BearerAuth
// @Success 200 {array} models.Task "List of tasks"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/sprints/{sprintId}/tasks [get]
func GetTasksBySprint(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID",
		})
		return
	}

	sprintIDStr := c.Param("sprintId")
	sprintID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid sprint ID",
		})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied to project",
		})
		return
	}

	tasks, err := models.GetTasksBySprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve tasks",
		})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// AssignTaskRequest represents the request to assign a task
type AssignTaskRequest struct {
	AssigneeID uuid.UUID `json:"assignee_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// UpdateTaskStatusRequest represents the request to update task status
type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=todo in_progress review done cancelled" example:"in_progress"`
}

// AssignTask assigns a task to a user
// @Summary Assign task
// @Description Assigns a task to a specific user
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param taskId path string true "Task ID"
// @Param assignment body AssignTaskRequest true "Task assignment"
// @Security BearerAuth
// @Success 200 {object} models.Task "Task assigned successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Task not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/tasks/{taskId}/assign [post]
func AssignTask(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID",
		})
		return
	}

	taskIDStr := c.Param("taskId")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid task ID",
		})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied to project",
		})
		return
	}

	var req AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Verify the assignee is a member of the project
	isAssigneeMember, err := models.IsProjectMember(db.GetDB(), projectID, req.AssigneeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if !isAssigneeMember {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Assignee must be a member of the project",
		})
		return
	}

	// Assign the task
	task, err := models.AssignTask(db.GetDB(), taskID, req.AssigneeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to assign task",
		})
		return
	}

	// Verify task belongs to the project
	if task.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Task not found in this project",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTaskStatus updates the status of a task
// @Summary Update task status
// @Description Updates the status of a specific task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param taskId path string true "Task ID"
// @Param status body UpdateTaskStatusRequest true "Task status update"
// @Security BearerAuth
// @Success 200 {object} models.Task "Task status updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Task not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/tasks/{taskId}/status [patch]
func UpdateTaskStatus(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID",
		})
		return
	}

	taskIDStr := c.Param("taskId")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid task ID",
		})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied to project",
		})
		return
	}

	var req UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Update the task status
	task, err := models.UpdateTaskStatus(db.GetDB(), taskID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update task status",
		})
		return
	}

	// Verify task belongs to the project
	if task.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Task not found in this project",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}
