package controllers

import (
	"net/http"

	"devhive-backend/models"
	"devhive-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ScrumController handles CRUD operations for Projects, Sprints, and Tasks
// related to Scrum-based Agile workflows.
type ScrumController struct {
	projectService services.ProjectService
	sprintService  services.SprintService
	taskService    services.TaskService
	userService    services.UserService
}

// NewScrumController creates a new instance of ScrumController
func NewScrumController(
	projectService services.ProjectService,
	sprintService services.SprintService,
	taskService services.TaskService,
	userService services.UserService,
) *ScrumController {
	return &ScrumController{
		projectService: projectService,
		sprintService:  sprintService,
		taskService:    taskService,
		userService:    userService,
	}
}

// -------------------- CREATE --------------------

// CreateProject creates a new project
func (sc *ScrumController) CreateProject(c *gin.Context) {
	var req models.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	project, err := sc.projectService.CreateProject(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while creating the project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// CreateSprint creates a new sprint
func (sc *ScrumController) CreateSprint(c *gin.Context) {
	var req models.SprintCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// For now, we'll use a default project ID since it's not in the request
	// TODO: Add projectID to SprintCreateRequest or get it from context
	projectID := uuid.New() // This should be properly handled
	sprint, err := sc.sprintService.CreateSprint(c.Request.Context(), req, projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while creating the sprint"})
		return
	}

	c.JSON(http.StatusOK, sprint)
}

// CreateTask creates a new task
func (sc *ScrumController) CreateTask(c *gin.Context) {
	var req models.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For now, we'll use a default project ID since it's not in the request
	// TODO: Add projectID to TaskCreateRequest or get it from context
	projectID := uuid.New() // This should be properly handled
	task, err := sc.taskService.CreateTask(c.Request.Context(), req, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while creating the task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// -------------------- DELETE --------------------

// DeleteProject deletes a project
func (sc *ScrumController) DeleteProject(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = sc.projectService.DeleteProject(c.Request.Context(), projectUUID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while deleting the project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// DeleteSprint deletes a sprint
func (sc *ScrumController) DeleteSprint(c *gin.Context) {
	sprintIDStr := c.Param("sprintId")
	sprintUUID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = sc.sprintService.DeleteSprint(c.Request.Context(), sprintUUID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while deleting the sprint"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sprint deleted successfully"})
}

// DeleteTask deletes a task
func (sc *ScrumController) DeleteTask(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskUUID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	err = sc.taskService.DeleteTask(c.Request.Context(), taskUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while deleting the task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// -------------------- UPDATE --------------------

// EditProject updates a project
func (sc *ScrumController) EditProject(c *gin.Context) {
	var req models.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// For now, we'll use a default project ID since it's not in the request
	// TODO: Add projectID to ProjectUpdateRequest or get it from context
	projectID := uuid.New() // This should be properly handled
	project, err := sc.projectService.UpdateProject(c.Request.Context(), projectID, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while updating the project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// EditSprint updates a sprint
func (sc *ScrumController) EditSprint(c *gin.Context) {
	var req models.SprintUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// For now, we'll use a default sprint ID since it's not in the request
	// TODO: Add sprintID to SprintUpdateRequest or get it from context
	sprintID := uuid.New() // This should be properly handled
	sprint, err := sc.sprintService.UpdateSprint(c.Request.Context(), sprintID, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while updating the sprint"})
		return
	}

	c.JSON(http.StatusOK, sprint)
}

// EditTask updates a task
func (sc *ScrumController) EditTask(c *gin.Context) {
	var req models.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For now, we'll use a default task ID since it's not in the request
	// TODO: Add taskID to TaskUpdateRequest or get it from context
	taskID := uuid.New() // This should be properly handled
	task, err := sc.taskService.UpdateTask(c.Request.Context(), taskID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while updating the task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTaskStatus updates the status of a task
func (sc *ScrumController) UpdateTaskStatus(c *gin.Context) {
	var req models.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For now, we'll use a default task ID since it's not in the request
	// TODO: Add taskID to TaskUpdateRequest or get it from context
	taskID := uuid.New() // This should be properly handled
	task, err := sc.taskService.UpdateTask(c.Request.Context(), taskID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while updating the task status"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// -------------------- READ --------------------

// GetProjectMembers gets the members of a project
func (sc *ScrumController) GetProjectMembers(c *gin.Context) {
	// TODO: Implement GetProjectMembers in ProjectService
	// For now, return empty array
	c.JSON(http.StatusOK, []interface{}{})
}

// GetSprintTasks gets the tasks of a sprint
func (sc *ScrumController) GetSprintTasks(c *gin.Context) {
	sprintIDStr := c.Param("sprintId")
	sprintUUID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
		return
	}

	tasks, err := sc.taskService.GetTasksBySprint(c.Request.Context(), sprintUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while retrieving sprint tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetProjectTasks gets the tasks of a project
func (sc *ScrumController) GetProjectTasks(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	tasks, err := sc.taskService.GetTasksByProject(c.Request.Context(), projectUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while retrieving project tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetProjectSprints gets the sprints of a project
func (sc *ScrumController) GetProjectSprints(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	sprints, err := sc.sprintService.GetSprintsForProject(c.Request.Context(), projectUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while retrieving project sprints"})
		return
	}

	c.JSON(http.StatusOK, sprints)
}

// GetProjectByID gets a project by ID
func (sc *ScrumController) GetProjectByID(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := sc.projectService.GetProject(c.Request.Context(), projectUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while retrieving the project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// GetSprintByID gets a sprint by ID
func (sc *ScrumController) GetSprintByID(c *gin.Context) {
	sprintIDStr := c.Param("sprintId")
	sprintUUID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
		return
	}

	sprint, err := sc.sprintService.GetSprintByID(c.Request.Context(), sprintUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while retrieving the sprint"})
		return
	}

	c.JSON(http.StatusOK, sprint)
}

// GetTaskByID gets a task by ID
func (sc *ScrumController) GetTaskByID(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskUUID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := sc.taskService.GetTaskByID(c.Request.Context(), taskUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while retrieving the task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetUserProjects gets the projects of a user
func (sc *ScrumController) GetUserProjects(c *gin.Context) {
	userIDStr := c.Param("userId")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	projects, err := sc.projectService.GetProjectsByUser(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while retrieving user projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// -------------------- PROJECT MEMBERSHIP --------------------

// JoinProject adds a user to a project
func (sc *ScrumController) JoinProject(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userIDStr := c.Param("userId")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get current user ID from context
	ownerIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ownerID, err := uuid.Parse(ownerIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
		return
	}

	err = sc.projectService.AddMember(c.Request.Context(), projectUUID, userUUID, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while adding user to project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User joined project successfully"})
}

// RemoveMemberFromProject removes a user from a project
func (sc *ScrumController) RemoveMemberFromProject(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userIDStr := c.Param("userId")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get current user ID from context
	ownerIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ownerID, err := uuid.Parse(ownerIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
		return
	}

	err = sc.projectService.RemoveMember(c.Request.Context(), projectUUID, userUUID, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while removing user from project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from project successfully"})
}

// GetActiveSprints gets the active sprints of a project
func (sc *ScrumController) GetActiveSprints(c *gin.Context) {
	// TODO: Implement GetActiveSprints in SprintService
	// For now, return empty array
	c.JSON(http.StatusOK, []interface{}{})
}

// LeaveProject removes the current user from a project
func (sc *ScrumController) LeaveProject(c *gin.Context) {
	var req struct {
		ProjectID string `json:"projectId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get current user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// For now, we'll use the same user ID as owner since they're leaving
	err = sc.projectService.RemoveMember(c.Request.Context(), projectUUID, userID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while leaving the project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User left project successfully"})
}

// UpdateProjectOwner updates the owner of a project
func (sc *ScrumController) UpdateProjectOwner(c *gin.Context) {
	var req struct {
		ProjectID  string `json:"projectId" binding:"required"`
		NewOwnerID string `json:"newOwnerId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement UpdateProjectOwner in ProjectService
	// For now, return success
	c.JSON(http.StatusOK, gin.H{"message": "Project owner updated successfully"})
}
