package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"devhive-backend/internal/broadcast"
	"devhive-backend/internal/config"
	"devhive-backend/internal/http/middleware"
	"devhive-backend/internal/http/response"
	"devhive-backend/internal/repo"
	"devhive-backend/internal/ws"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
)

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

type MessageHandler struct {
	queries *repo.Queries
	cfg     *config.Config
	hub     *ws.Hub
}

func NewMessageHandler(queries *repo.Queries, cfg *config.Config, hub *ws.Hub) *MessageHandler {
	return &MessageHandler{
		queries: queries,
		cfg:     cfg,
		hub:     hub,
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
		ProjectID: projectUUID,
		UserID:    userUUID,
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
		ProjectID: projectUUID,
		UserID:    userUUID,
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

	// Get sender details for full response
	sender, _ := h.queries.GetUserByID(r.Context(), userUUID2)

	messageResp := MessageResponse{
		ID:          message.ID.String(),
		ProjectID:   message.ProjectID.String(),
		SenderID:    message.SenderID.String(),
		Content:     message.Content,
		MessageType: message.MessageType,
		CreatedAt:   message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   message.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if sender.ID != uuid.Nil {
		avatarURL := ""
		if sender.AvatarUrl != nil {
			avatarURL = *sender.AvatarUrl
		}
		messageResp.Sender = struct {
			Username  string `json:"username"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			AvatarURL string `json:"avatarUrl,omitempty"`
		}{
			Username:  sender.Username,
			FirstName: sender.FirstName,
			LastName:  sender.LastName,
			AvatarURL: avatarURL,
		}
	}

	// Broadcast message created event
	broadcast.Send(r.Context(), projectID, broadcast.EventMessageCreated, messageResp)

	response.JSON(w, http.StatusCreated, messageResp)
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
		ProjectID: projectUUID,
		UserID:    userUUID,
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate against CORS allowed origins
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range []string{"http://localhost:3000", "http://localhost:5173", "https://devhive.it.com"} {
			if origin == allowedOrigin {
				return true
			}
		}
		return true // Allow for now, but should be restricted in production
	},
}

// WebSocketHandler handles WebSocket connections for real-time messaging and cache invalidation
func (h *MessageHandler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Check if hub is available (nil when running in Lambda/serverless)
	if h.hub == nil {
		http.Error(w, "WebSocket connections are not supported in this deployment. Use API Gateway WebSocket API for real-time features.", http.StatusServiceUnavailable)
		return
	}

	// Extract JWT token from multiple sources (in order of preference):
	// 1. Authorization header (most secure)
	// 2. Cookie (secure, httpOnly)
	// 3. Query param (fallback, less secure - for debugging/backward compatibility)
	var token string

	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Fallback to cookie if header not present
	if token == "" {
		cookie, err := r.Cookie("access_token")
		if err == nil && cookie.Value != "" {
			token = cookie.Value
		}
	}

	// Fallback to query param (for backward compatibility, but should be deprecated)
	if token == "" {
		token = r.URL.Query().Get("token")
		if token != "" {
			log.Printf("WARNING: WebSocket token provided via query parameter (less secure). Consider using Authorization header or cookie.")
		}
	}

	if token == "" {
		http.Error(w, "Missing authentication token. Provide token via Authorization header, cookie, or query parameter.", http.StatusUnauthorized)
		return
	}

	// Validate JWT token and extract user_id
	userID, err := h.validateJWTToken(token)
	if err != nil {
		// Provide more specific error messages for expired tokens
		errStr := err.Error()
		if contains(errStr, "expired") || contains(errStr, "exp") {
			log.Printf("JWT validation error: token has invalid claims: token is expired")
			http.Error(w, "Authentication token has expired. Please refresh your token and reconnect.", http.StatusUnauthorized)
			return
		}
		if contains(errStr, "malformed") || contains(errStr, "invalid") {
			log.Printf("JWT validation error: malformed token")
			http.Error(w, "Invalid authentication token format", http.StatusUnauthorized)
			return
		}
		log.Printf("JWT validation error: %v", err)
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}

	// Extract project_id from query params
	projectID := r.URL.Query().Get("projectId")
	if projectID == "" {
		// Fallback for backward compatibility
		projectID = r.URL.Query().Get("project_id")
	}
	if projectID == "" {
		http.Error(w, "Missing projectId", http.StatusBadRequest)
		return
	}

	// Validate project access (DO NOT trust query parameters)
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ProjectID: projectUUID,
		UserID:    userUUID,
	})
	if err != nil {
		log.Printf("ERROR: CheckProjectAccess failed for project %s, user %s: %v",
			projectUUID.String(), userUUID.String(), err)
		http.Error(w, "Failed to verify project access", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		// Check if project exists (simple check without JOIN)
		projectExists, existsErr := h.queries.ProjectExists(r.Context(), projectUUID)
		if existsErr != nil {
			log.Printf("ERROR: ProjectExists query failed for project %s: %v",
				projectUUID.String(), existsErr)
		} else if !projectExists {
			log.Printf("WARN: Project %s does not exist (user %s)",
				projectUUID.String(), userUUID.String())
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		} else {
			// Project exists, check if user is owner
			isOwner, ownerErr := h.queries.CheckProjectOwner(r.Context(), repo.CheckProjectOwnerParams{
				ID:      projectUUID,
				OwnerID: userUUID,
			})
			if ownerErr != nil {
				log.Printf("ERROR: CheckProjectOwner failed for project %s, user %s: %v",
					projectUUID.String(), userUUID.String(), ownerErr)
			} else if isOwner {
				log.Printf("ERROR: Project owner %s denied access to project %s - BUG! CheckProjectAccess returned false but CheckProjectOwner returned true",
					userUUID.String(), projectUUID.String())
			} else {
				log.Printf("WARN: User %s is not a member of project %s",
					userUUID.String(), projectUUID.String())
			}
		}
		// Return 403 with clear message - user is authenticated but not authorized for this project
		// Frontend should handle this by clearing selectedProjectId, not logging out
		http.Error(w, "You are not a member of this project. Please select a project you have access to.", http.StatusForbidden)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client using the helper function
	client := ws.NewClient(conn, userID, projectID, h.hub)

	// Register client with hub (using the Register channel)
	select {
	case h.hub.Register <- client:
		// Client registered successfully
		log.Printf("WebSocket client registered: user=%s, project=%s", userID, projectID)
	default:
		log.Printf("Failed to register WebSocket client: hub Register channel full")
		// Send proper close frame before closing
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Server overloaded"))
		conn.Close()
		return
	}

	// Start goroutines for reading and writing
	// Handler returns here - connection is managed by ReadPump/WritePump
	go client.ReadPump()
	go client.WritePump()

	// Handler returns immediately - connection lifecycle is managed by ReadPump/WritePump goroutines
	// This is correct: we don't want to block the HTTP handler
}

// GetWebSocketStatus returns connection status for debugging
func (h *MessageHandler) GetWebSocketStatus(w http.ResponseWriter, r *http.Request) {
	// Check if hub is available (nil when running in Lambda/serverless)
	if h.hub == nil {
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"status":  "unavailable",
			"message": "WebSocket is not available in serverless deployment",
		})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		response.BadRequest(w, "Project ID required")
		return
	}

	totalClients, matchingClients, userIDs := h.hub.GetProjectConnections(projectID)

	clientDetails := make([]map[string]string, len(userIDs))
	for i, userID := range userIDs {
		clientDetails[i] = map[string]string{
			"userId":    userID,
			"projectId": projectID,
		}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"projectId":       projectID,
		"totalClients":    totalClients,
		"matchingClients": matchingClients,
		"clients":         clientDetails,
	})
}

// validateJWTToken validates a JWT token and returns the user ID
// Returns a more specific error if the token is expired
func (h *MessageHandler) validateJWTToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.cfg.JWT.SigningKey), nil
	})

	if err != nil {
		// Return the error as-is so we can check for ValidationErrorExpired
		return "", err
	}

	if !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrSignatureInvalid
	}

	// Extract user ID from claims (sub claim)
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", jwt.ErrSignatureInvalid
	}

	return userID, nil
}
