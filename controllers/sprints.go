package controllers

import (
	"net/http"
	"strconv"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetSprints retrieves all sprints for a project
func GetSprints(c *gin.Context) {
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

	// Get sprints from database
	sprints, err := models.GetSprints(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sprints"})
		return
	}

	// Apply pagination
	total := len(sprints)
	start := offset
	end := offset + limit
	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedSprints := sprints[start:end]

	c.JSON(http.StatusOK, gin.H{
		"sprints": paginatedSprints,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_more": end < total,
		},
	})
}

// CreateSprint creates a new sprint
func CreateSprint(c *gin.Context) {
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

	// Check if user is the owner of the project
	isOwner, err := models.IsProjectOwner(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can create sprints"})
		return
	}

	var req models.SprintCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate dates
	if req.StartDate.After(req.EndDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start date cannot be after end date"})
		return
	}

	// Create sprint
	sprint, err := models.CreateSprint(db.GetDB(), req, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sprint"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"sprint":  sprint,
		"message": "Sprint created successfully",
	})
}

// GetSprint retrieves a specific sprint
func GetSprint(c *gin.Context) {
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
	sprintIDStr := c.Param("id")
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

	// Get sprint details
	sprint, err := models.GetSprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}

	// Verify sprint belongs to the project
	if sprint.ProjectID != projectID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found in this project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sprint": sprint,
	})
}

// UpdateSprint updates a sprint
func UpdateSprint(c *gin.Context) {
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
	sprintIDStr := c.Param("id")
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can update sprints"})
		return
	}

	// Verify sprint belongs to the project
	sprint, err := models.GetSprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}

	if sprint.ProjectID != projectID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found in this project"})
		return
	}

	var req models.SprintUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate dates if provided
	if req.StartDate != nil && req.EndDate != nil {
		if req.StartDate.After(*req.EndDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start date cannot be after end date"})
			return
		}
	}

	// Update sprint
	updatedSprint, err := models.UpdateSprint(db.GetDB(), sprintID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sprint"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sprint":  updatedSprint,
		"message": "Sprint updated successfully",
	})
}

// DeleteSprint deletes a sprint
func DeleteSprint(c *gin.Context) {
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
	sprintIDStr := c.Param("id")
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can delete sprints"})
		return
	}

	// Verify sprint belongs to the project
	sprint, err := models.GetSprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}

	if sprint.ProjectID != projectID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found in this project"})
		return
	}

	// Delete sprint
	err = models.DeleteSprint(db.GetDB(), sprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sprint"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sprint deleted successfully",
	})
}
