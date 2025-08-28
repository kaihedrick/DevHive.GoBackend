package controllers

import (
	"net/http"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetCurrentUserID extracts the current user ID from the Gin context
// This function should be called after the auth middleware has set the user_id
func GetCurrentUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}

	if id, ok := userID.(uuid.UUID); ok {
		return id
	}

	if idStr, ok := userID.(string); ok {
		if id, err := uuid.Parse(idStr); err == nil {
			return id
		}
	}

	return uuid.Nil
}

// GetProjects retrieves all projects for the current user
// @Summary Get user projects
// @Description Retrieves all projects where the current user is a member or owner
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Project "List of projects"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects [get]
func GetProjects(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	projects, err := models.GetProjects(db.GetDB(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve projects",
		})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// CreateProject creates a new project
// @Summary Create project
// @Description Creates a new project for the current user
// @Tags projects
// @Accept json
// @Produce json
// @Param project body models.ProjectCreateRequest true "Project to create"
// @Security BearerAuth
// @Success 201 {object} models.Project "Project created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects [post]
func CreateProject(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	var req models.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	project, err := models.CreateProject(db.GetDB(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create project",
		})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProject retrieves a specific project by ID
// @Summary Get project
// @Description Retrieves a specific project by ID if the user has access
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Security BearerAuth
// @Success 200 {object} models.Project "Project details"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Project not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id} [get]
func GetProject(c *gin.Context) {
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

	project, err := models.GetProject(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Project not found",
		})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject updates an existing project
// @Summary Update project
// @Description Updates an existing project if the user has permission
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body models.ProjectUpdateRequest true "Updated project data"
// @Security BearerAuth
// @Success 200 {object} models.Project "Project updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Project not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id} [put]
func UpdateProject(c *gin.Context) {
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

	// Check if user is the project owner
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if userRole != "owner" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Only project owners can update projects",
		})
		return
	}

	var req models.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	project, err := models.UpdateProject(db.GetDB(), projectID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update project",
		})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject deletes a project
// @Summary Delete project
// @Description Deletes a project if the user is the owner
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Security BearerAuth
// @Success 204 "Project deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Project not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id} [delete]
func DeleteProject(c *gin.Context) {
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

	// Check if user is the project owner
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if userRole != "owner" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Only project owners can delete projects",
		})
		return
	}

	err = models.DeleteProject(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete project",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddProjectMember adds a user to a project
// @Summary Add project member
// @Description Adds a user as a member to a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param member body models.AddMemberRequest true "Member to add"
// @Security BearerAuth
// @Success 201 {object} models.ProjectMember "Member added successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/members [post]
func AddProjectMember(c *gin.Context) {
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

	// Check if user has permission to add members
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if userRole != "owner" && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Insufficient permissions to add members",
		})
		return
	}

	var req models.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	err = models.AddProjectMember(db.GetDB(), projectID, req.UserID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to add project member",
		})
		return
	}

	c.Status(http.StatusCreated)
}

// RemoveProjectMember removes a user from a project
// @Summary Remove project member
// @Description Removes a user from a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param userId path string true "User ID to remove"
// @Security BearerAuth
// @Success 204 "Member removed successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/members/{userId} [delete]
func RemoveProjectMember(c *gin.Context) {
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

	userToRemoveStr := c.Param("userId")
	userToRemove, err := uuid.Parse(userToRemoveStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid user ID",
		})
		return
	}

	// Check if user has permission to remove members
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if userRole != "owner" && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Insufficient permissions to remove members",
		})
		return
	}

	// Prevent removing the project owner
	if userToRemove == userID && userRole == "owner" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Cannot remove the project owner",
		})
		return
	}

	err = models.RemoveProjectMember(db.GetDB(), projectID, userToRemove)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to remove project member",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProjectMembers retrieves all members of a project
// @Summary Get project members
// @Description Retrieves all members of a specific project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Security BearerAuth
// @Success 200 {array} models.ProjectMember "List of project members"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/members [get]
func GetProjectMembers(c *gin.Context) {
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

	members, err := models.GetProjectMembers(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve project members",
		})
		return
	}

	c.JSON(http.StatusOK, members)
}

// UpdateProjectMemberRoleRequest represents the request to update a project member's role
type UpdateProjectMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=owner admin member" example:"admin"`
}

// UpdateProjectMemberRole updates a project member's role
// @Summary Update project member role
// @Description Updates the role of a project member
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param userId path string true "User ID"
// @Param role body UpdateProjectMemberRoleRequest true "New role"
// @Security BearerAuth
// @Success 200 {object} models.ProjectMember "Member role updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/members/{userId}/role [put]
func UpdateProjectMemberRole(c *gin.Context) {
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

	memberUserIDStr := c.Param("userId")
	memberUserID, err := uuid.Parse(memberUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid user ID",
		})
		return
	}

	// Check if user has permission to update member roles
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if userRole != "owner" && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Insufficient permissions to update member roles",
		})
		return
	}

	var req UpdateProjectMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Prevent changing the project owner's role
	if memberUserID == userID && userRole == "owner" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Cannot change the project owner's role",
		})
		return
	}

	// Update the member's role
	err = models.UpdateProjectMemberRole(db.GetDB(), projectID, memberUserID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update member role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member role updated successfully",
		"user_id": memberUserID,
		"role":    req.Role,
	})
}
