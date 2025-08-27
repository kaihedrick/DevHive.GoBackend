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

	// Check if user has admin/owner access to the project
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if userRole != "admin" && userRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update project"})
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
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if userRole != "owner" {
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

	// Check if user has admin/owner access to the project
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if userRole != "admin" && userRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to add members"})
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
	err = models.AddProjectMember(db.GetDB(), projectID, req.UserID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member added successfully",
		"user":    userToAdd,
		"role":    req.Role,
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

	// Check if user has admin/owner access to the project
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if userRole != "admin" && userRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to remove members"})
		return
	}

	// Check if trying to remove the owner
	if userRole != "owner" {
		memberToRemoveRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userToRemoveID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if memberToRemoveRole == "owner" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove project owner"})
			return
		}
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
