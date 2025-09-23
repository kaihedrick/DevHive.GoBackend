package handlers

import (
	"net/http"
	"strconv"

	"devhive-backend/internal/http/middleware"
	"devhive-backend/internal/http/response"
	"devhive-backend/internal/repo"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type MessageHandler struct {
	queries *repo.Queries
}

func NewMessageHandler(queries *repo.Queries) *MessageHandler {
	return &MessageHandler{
		queries: queries,
	}
}

// CreateMessageRequest represents the message creation request
type CreateMessageRequest struct {
	Content         string `json:"content"`
	MessageType     string `json:"messageType"`
	ParentMessageID string `json:"parentMessageId,omitempty"`
}

// MessageResponse represents a message response
type MessageResponse struct {
	ID              string `json:"id"`
	ProjectID       string `json:"projectId"`
	SenderID        string `json:"senderId"`
	Content         string `json:"content"`
	MessageType     string `json:"messageType"`
	ParentMessageID string `json:"parentMessageId,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
	Sender          struct {
		Username  string `json:"username"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		AvatarURL string `json:"avatarUrl,omitempty"`
	} `json:"sender"`
}

// ListMessagesByProject handles listing messages for a project
func (h *MessageHandler) ListMessagesByProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		response.BadRequest(w, "Project ID is required")
		return
	}

	// Check if user has access to project
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		response.BadRequest(w, "Invalid project ID")
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to project")
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	projectUUID2, err := uuid.Parse(projectID)
	if err != nil {
		response.BadRequest(w, "Invalid project ID")
		return
	}
	messages, err := h.queries.ListMessagesByProject(r.Context(), repo.ListMessagesByProjectParams{
		ProjectID: projectUUID2,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		response.InternalServerError(w, "Failed to list messages")
		return
	}

	// Convert to response format
	var messageResponses []MessageResponse
	for _, message := range messages {
		messageResp := MessageResponse{
			ID:          message.ID.String(),
			ProjectID:   message.ProjectID.String(),
			SenderID:    message.SenderID.String(),
			Content:     message.Content,
			MessageType: message.MessageType,
			CreatedAt:   message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   message.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if message.ParentMessageID.Valid {
			parentUUID := uuid.UUID(message.ParentMessageID.Bytes)
			messageResp.ParentMessageID = parentUUID.String()
		}

		avatarURL := ""
		if message.SenderAvatarUrl != nil {
			avatarURL = *message.SenderAvatarUrl
		}
		messageResp.Sender = struct {
			Username  string `json:"username"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			AvatarURL string `json:"avatarUrl,omitempty"`
		}{
			Username:  message.SenderUsername,
			FirstName: message.SenderFirstName,
			LastName:  message.SenderLastName,
			AvatarURL: avatarURL,
		}

		messageResponses = append(messageResponses, messageResp)
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"messages": messageResponses,
		"limit":    limit,
		"offset":   offset,
	})
}

// CreateMessage handles message creation
func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		response.BadRequest(w, "Project ID is required")
		return
	}

	// Check if user has access to project
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		response.BadRequest(w, "Invalid project ID")
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to project")
		return
	}

	var req CreateMessageRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Set default message type if not provided
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	projectUUID3, err := uuid.Parse(projectID)
	if err != nil {
		response.BadRequest(w, "Invalid project ID")
		return
	}
	userUUID2, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	var parentMessageID pgtype.UUID
	if req.ParentMessageID != "" {
		parentUUID, err := uuid.Parse(req.ParentMessageID)
		if err != nil {
			response.BadRequest(w, "Invalid parent message ID")
			return
		}
		parentMessageID = pgtype.UUID{Bytes: parentUUID, Valid: true}
	}

	message, err := h.queries.CreateMessage(r.Context(), repo.CreateMessageParams{
		ProjectID:       projectUUID3,
		SenderID:        userUUID2,
		Content:         req.Content,
		MessageType:     req.MessageType,
		ParentMessageID: parentMessageID,
	})
	if err != nil {
		response.BadRequest(w, "Failed to create message: "+err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, MessageResponse{
		ID:          message.ID.String(),
		ProjectID:   message.ProjectID.String(),
		SenderID:    message.SenderID.String(),
		Content:     message.Content,
		MessageType: message.MessageType,
		CreatedAt:   message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   message.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListMessages handles listing messages with filters
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	// Parse query parameters
	projectID := r.URL.Query().Get("projectId")
	afterID := r.URL.Query().Get("afterId")
	limitStr := r.URL.Query().Get("limit")

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// For now, we'll implement a simple project-based message listing
	// In a full implementation, you might want to add more sophisticated filtering
	if projectID == "" {
		response.BadRequest(w, "Project ID is required")
		return
	}

	// Check if user has access to project
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		response.BadRequest(w, "Invalid project ID")
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to project")
		return
	}

	// If afterId is provided, get messages after that ID
	if afterID != "" {
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			response.BadRequest(w, "Invalid project ID")
			return
		}
		afterUUID, err := uuid.Parse(afterID)
		if err != nil {
			response.BadRequest(w, "Invalid after ID")
			return
		}
		messages, err := h.queries.ListMessagesByProjectAfter(r.Context(), repo.ListMessagesByProjectAfterParams{
			ProjectID: projectUUID,
			ID:        afterUUID,
			Limit:     int32(limit),
		})
		if err != nil {
			response.InternalServerError(w, "Failed to list messages")
			return
		}

		// Convert to response format
		var messageResponses []MessageResponse
		for _, message := range messages {
			messageResp := MessageResponse{
				ID:          message.ID.String(),
				ProjectID:   message.ProjectID.String(),
				SenderID:    message.SenderID.String(),
				Content:     message.Content,
				MessageType: message.MessageType,
				CreatedAt:   message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:   message.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}

			if message.ParentMessageID.Valid {
				parentUUID := uuid.UUID(message.ParentMessageID.Bytes)
				messageResp.ParentMessageID = parentUUID.String()
			}

			avatarURL := ""
			if message.SenderAvatarUrl != nil {
				avatarURL = *message.SenderAvatarUrl
			}
			messageResp.Sender = struct {
				Username  string `json:"username"`
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
				AvatarURL string `json:"avatarUrl,omitempty"`
			}{
				Username:  message.SenderUsername,
				FirstName: message.SenderFirstName,
				LastName:  message.SenderLastName,
				AvatarURL: avatarURL,
			}

			messageResponses = append(messageResponses, messageResp)
		}

		response.JSON(w, http.StatusOK, map[string]interface{}{
			"messages": messageResponses,
			"limit":    limit,
		})
		return
	}

	// Fallback to regular project message listing
	// This is a simplified implementation - in production you'd want more sophisticated filtering
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Use project-specific message endpoint for better performance",
		"endpoint": "/api/v1/projects/" + projectID + "/messages",
	})
}

// WebSocketHandler handles WebSocket connections for real-time messaging
func (h *MessageHandler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket handler for real-time messaging
	// This would typically involve upgrading the HTTP connection to WebSocket
	// and managing real-time message broadcasting
	response.JSON(w, http.StatusNotImplemented, map[string]string{
		"message": "WebSocket messaging not yet implemented",
	})
}
