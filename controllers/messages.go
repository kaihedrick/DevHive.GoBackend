package controllers

import (
	"net/http"
	"strconv"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetMessages retrieves messages for a project
func GetMessages(c *gin.Context) {
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

	// Get messages from database
	messages, err := models.GetMessages(db.GetDB(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	// Get total message count
	total, err := models.GetMessageCount(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get message count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_more": offset+limit < total,
		},
	})
}

// CreateMessage creates a new message
func CreateMessage(c *gin.Context) {
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

	var req models.MessageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If this is a reply, verify the parent message exists and belongs to the project
	if req.ParentMessageID != nil {
		parentMessage, err := models.GetMessage(db.GetDB(), *req.ParentMessageID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent message not found"})
			return
		}

		if parentMessage.ProjectID != projectID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent message does not belong to this project"})
			return
		}
	}

	// Create message
	message, err := models.CreateMessage(db.GetDB(), req, projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	// Get the complete message with sender information
	completeMessage, err := models.GetMessage(db.GetDB(), message.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve created message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      completeMessage,
		"message_text": "Message created successfully",
	})
}

// UpdateMessage updates a message
func UpdateMessage(c *gin.Context) {
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

	// Get message ID from URL parameter
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
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

	// Get the message to verify ownership and project
	message, err := models.GetMessage(db.GetDB(), messageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Verify message belongs to the project
	if message.ProjectID != projectID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found in this project"})
		return
	}

	// Check if user is the sender of the message
	if message.SenderID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Can only edit your own messages"})
		return
	}

	var req models.MessageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update message
	updatedMessage, err := models.UpdateMessage(db.GetDB(), messageID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      updatedMessage,
		"message_text": "Message updated successfully",
	})
}

// DeleteMessage deletes a message
func DeleteMessage(c *gin.Context) {
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

	// Get message ID from URL parameter
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
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

	// Get the message to verify ownership and project
	message, err := models.GetMessage(db.GetDB(), messageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Verify message belongs to the project
	if message.ProjectID != projectID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found in this project"})
		return
	}

	// Check if user is the sender of the message or has admin/owner role
	if message.SenderID != userID {
		userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if userRole != "admin" && userRole != "owner" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Can only delete your own messages or need admin/owner role"})
			return
		}
	}

	// Delete message
	err = models.DeleteMessage(db.GetDB(), messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message_text": "Message deleted successfully",
	})
}
