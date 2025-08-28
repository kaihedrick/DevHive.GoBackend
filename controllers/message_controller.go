package controllers

import (
	"net/http"
	"strconv"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetMessages retrieves all messages for a project
// @Summary Get project messages
// @Description Retrieves all messages for a specific project with pagination
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param limit query int false "Number of messages to return (default: 50)"
// @Param offset query int false "Number of messages to skip (default: 0)"
// @Security BearerAuth
// @Success 200 {array} models.Message "List of messages"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/messages [get]
func GetMessages(c *gin.Context) {
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

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	messages, err := models.GetMessages(db.GetDB(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve messages",
		})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// CreateMessage creates a new message
// @Summary Create message
// @Description Creates a new message in a project
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param message body models.MessageCreateRequest true "Message to create"
// @Security BearerAuth
// @Success 201 {object} models.Message "Message created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/messages [post]
func CreateMessage(c *gin.Context) {
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

	var req models.MessageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	message, err := models.CreateMessage(db.GetDB(), req, projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create message",
		})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// UpdateMessage updates an existing message
// @Summary Update message
// @Description Updates an existing message if the user has permission
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param messageId path string true "Message ID"
// @Param message body models.MessageUpdateRequest true "Updated message data"
// @Security BearerAuth
// @Success 200 {object} models.Message "Message updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Message not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/messages/{messageId} [put]
func UpdateMessage(c *gin.Context) {
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

	messageIDStr := c.Param("messageId")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid message ID",
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

	// Get the message to check ownership
	message, err := models.GetMessage(db.GetDB(), messageID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Message not found",
		})
		return
	}

	// Verify message belongs to the project
	if message.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Message not found in this project",
		})
		return
	}

	// Check if user is the message sender (only sender can edit)
	if message.SenderID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Only message sender can edit messages",
		})
		return
	}

	var req models.MessageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	updatedMessage, err := models.UpdateMessage(db.GetDB(), messageID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update message",
		})
		return
	}

	c.JSON(http.StatusOK, updatedMessage)
}

// DeleteMessage deletes a message
// @Summary Delete message
// @Description Deletes a message if the user has permission
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param messageId path string true "Message ID"
// @Security BearerAuth
// @Success 204 "Message deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Message not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id}/messages/{messageId} [delete]
func DeleteMessage(c *gin.Context) {
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

	messageIDStr := c.Param("messageId")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid message ID",
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

	// Get the message to check ownership
	message, err := models.GetMessage(db.GetDB(), messageID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Message not found",
		})
		return
	}

	// Verify message belongs to the project
	if message.ProjectID != projectID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Message not found in this project",
		})
		return
	}

	// Check if user is the message sender (only sender can delete)
	if message.SenderID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Only message sender can delete messages",
		})
		return
	}

	err = models.DeleteMessage(db.GetDB(), messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete message",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
