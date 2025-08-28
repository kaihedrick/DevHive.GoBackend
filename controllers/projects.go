package controllers

import (
	"net/http"
	"strconv"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetProjects retrieves all projects for the current user
// @Summary Get user projects
// @Description Retrieve all projects for the authenticated user with pagination
// @Tags projects
// @Accept json
// @Produce json
// @Param limit query int false "Number of projects to return (max 100)" default(20)
// @Param offset query int false "Number of projects to skip" default(0)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Projects with pagination info"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects [get]
func GetProjects(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
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

	// Get projects from database
	projects, err := models.GetProjects(db.GetDB(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
		return
	}

	// Apply pagination
	total := len(projects)
	start := offset
	end := offset + limit
	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedProjects := projects[start:end]

	c.JSON(http.StatusOK, gin.H{
		"projects": paginatedProjects,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_more": end < total,
		},
	})
}

// CreateProject creates a new project
// @Summary Create a new project
// @Description Create a new project for the authenticated user
// @Tags projects
// @Accept json
// @Produce json
// @Param project body models.ProjectCreateRequest true "Project creation request"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Project created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects [post]
func CreateProject(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create project
	project, err := models.CreateProject(db.GetDB(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"project": project,
		"message": "Project created successfully",
	})
}

// GetProject retrieves a specific project
// @Summary Get a specific project
// @Description Retrieve a specific project by ID for the authenticated user
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} models.Project "Project details"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - User not a member"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /projects/{id} [get]
func GetProject(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
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

	// Get project details
	project, err := models.GetProject(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Get project members
	members, err := models.GetProjectMembers(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve project members"})
		return
	}

	project.Members = make([]*models.User, len(members))
	for i, member := range members {
		project.Members[i] = member.User
	}

	c.JSON(http.StatusOK, gin.H{
		"project": project,
	})
}

// UpdateProject updates a project
func UpdateProject(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check if user is the owner of the project
	isOwner, err := models.IsProjectOwner(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can update the project"})
		return
	}

	var req models.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update project
	project, err := models.UpdateProject(db.GetDB(), projectID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": project,
		"message": "Project updated successfully",
	})
}

// DeleteProject deletes a project
func DeleteProject(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check if user is the owner of the project
	isOwner, err := models.IsProjectOwner(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can delete the project"})
		return
	}

	// Delete project
	err = models.DeleteProject(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}

// AddProjectMember adds a member to a project
func AddProjectMember(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check if user is the owner of the project
	isOwner, err := models.IsProjectOwner(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can add members"})
		return
	}

	var req models.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the user to be added exists
	userToAdd, err := models.GetUserByID(db.GetDB(), req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Add member to project
	err = models.AddProjectMember(db.GetDB(), projectID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member added successfully",
		"user":    userToAdd,
	})
}

// RemoveProjectMember removes a member from a project
func RemoveProjectMember(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get user ID to remove from URL parameter
	userToRemoveStr := c.Param("userId")
	userToRemoveID, err := uuid.Parse(userToRemoveStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if user is the owner of the project
	isOwner, err := models.IsProjectOwner(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can remove members"})
		return
	}

	// Check if trying to remove the owner
	if userToRemoveID == userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove project owner"})
		return
	}

	// Remove member from project
	err = models.RemoveProjectMember(db.GetDB(), projectID, userToRemoveID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member from project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed successfully",
	})
}
