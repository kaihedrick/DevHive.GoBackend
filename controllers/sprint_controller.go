package controllers

import (
	"net/http"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetSprints retrieves all sprints for a specific project
// @Summary Get project sprints
// @Description Retrieves all sprints for a specific project
// @Tags sprints
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Security BearerAuth
// @Success 200 {array} models.Sprint "List of sprints"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/sprints [get]
func GetSprints(c *gin.Context) {
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

	sprints, err := models.GetSprints(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve sprints",
		})
		return
	}

	c.JSON(http.StatusOK, sprints)
}

// CreateSprint creates a new sprint
// @Summary Create sprint
// @Description Creates a new sprint for a specific project
// @Tags sprints
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param sprint body models.SprintCreateRequest true "Sprint to create"
// @Security BearerAuth
// @Success 201 {object} models.Sprint "Sprint created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/sprints [post]
func CreateSprint(c *gin.Context) {
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

	var req models.SprintCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	sprint, err := models.CreateSprint(db.GetDB(), req, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create sprint",
		})
		return
	}

	c.JSON(http.StatusCreated, sprint)
}

// GetSprint retrieves a specific sprint by ID
// @Summary Get sprint
// @Description Retrieves a specific sprint by ID if the user has access
// @Tags sprints
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param sprintId path string true "Sprint ID"
// @Security BearerAuth
// @Success 200 {object} models.Sprint "Sprint details"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Sprint not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/sprints/{sprintId} [get]
func GetSprint(c *gin.Context) {
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

	sprint, err := models.GetSprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Sprint not found",
		})
		return
	}

	// Verify sprint belongs to the project
	if sprint.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Sprint not found in this project",
		})
		return
	}

	c.JSON(http.StatusOK, sprint)
}

// UpdateSprint updates an existing sprint
// @Summary Update sprint
// @Description Updates an existing sprint if the user has permission
// @Tags sprints
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param sprintId path string true "Sprint ID"
// @Param sprint body models.SprintUpdateRequest true "Updated sprint data"
// @Security BearerAuth
// @Success 200 {object} models.Sprint "Sprint updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Sprint not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/sprints/{sprintId} [put]
func UpdateSprint(c *gin.Context) {
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

	var req models.SprintUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	sprint, err := models.UpdateSprint(db.GetDB(), sprintID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update sprint",
		})
		return
	}

	// Verify sprint belongs to the project
	if sprint.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Sprint not found in this project",
		})
		return
	}

	c.JSON(http.StatusOK, sprint)
}

// DeleteSprint deletes a sprint
// @Summary Delete sprint
// @Description Deletes a sprint if the user has permission
// @Tags sprints
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param sprintId path string true "Sprint ID"
// @Security BearerAuth
// @Success 204 "Sprint deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Sprint not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/sprints/{sprintId} [delete]
func DeleteSprint(c *gin.Context) {
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

	// Check if user has permission to delete sprints
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if userRole != "owner" && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Insufficient permissions to delete sprints",
		})
		return
	}

	err = models.DeleteSprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete sprint",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// StartSprint starts a sprint
// @Summary Start sprint
// @Description Starts a sprint and changes its status to active
// @Tags sprints
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param sprintId path string true "Sprint ID"
// @Security BearerAuth
// @Success 200 {object} models.Sprint "Sprint started successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Sprint not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/sprints/{sprintId}/start [post]
func StartSprint(c *gin.Context) {
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

	// Check if user has permission to start sprints
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if userRole != "owner" && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Insufficient permissions to start sprints",
		})
		return
	}

	// Start the sprint
	sprint, err := models.StartSprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to start sprint",
		})
		return
	}

	// Verify sprint belongs to the project
	if sprint.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Sprint not found in this project",
		})
		return
	}

	c.JSON(http.StatusOK, sprint)
}

// CompleteSprint completes a sprint
// @Summary Complete sprint
// @Description Completes a sprint and changes its status to completed
// @Tags sprints
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param sprintId path string true "Sprint ID"
// @Security BearerAuth
// @Success 200 {object} models.Sprint "Sprint completed successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Sprint not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/sprints/{sprintId}/complete [post]
func CompleteSprint(c *gin.Context) {
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

	// Check if user has permission to complete sprints
	userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if userRole != "owner" && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Insufficient permissions to complete sprints",
		})
		return
	}

	// Complete the sprint
	sprint, err := models.CompleteSprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to complete sprint",
		})
		return
	}

	// Verify sprint belongs to the project
	if sprint.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Sprint not found in this project",
		})
		return
	}

	c.JSON(http.StatusOK, sprint)
}
