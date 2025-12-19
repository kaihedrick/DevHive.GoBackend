package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"devhive-backend/internal/http/middleware"
	"devhive-backend/internal/http/response"
	"devhive-backend/internal/repo"
	"devhive-backend/internal/ws"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	queries *repo.Queries
	hub     *ws.Hub
}

func NewProjectHandler(queries *repo.Queries, hub *ws.Hub) *ProjectHandler {
	return &ProjectHandler{
		queries: queries,
		hub:     hub,
	}
}

// CreateProjectRequest represents the project creation request
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// JoinProjectRequest represents the join-project by code request
type JoinProjectRequest struct {
	ProjectID string `json:"projectId"`
}

// UpdateProjectRequest represents the project update request
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateInviteRequest represents the invite creation request
type CreateInviteRequest struct {
	ExpiresInMinutes *int `json:"expiresInMinutes,omitempty"` // Default: 30
	MaxUses          *int `json:"maxUses,omitempty"`          // Default: unlimited
}

// InviteResponse represents an invite response
type InviteResponse struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	InviteToken string `json:"inviteToken"`
	InviteURL   string `json:"inviteUrl"` // Full URL for frontend
	ExpiresAt   string `json:"expiresAt"`
	MaxUses     *int   `json:"maxUses,omitempty"`
	UsedCount   int    `json:"usedCount"`
	IsActive    bool   `json:"isActive"`
	CreatedAt   string `json:"createdAt"`
}

// InviteDetailsResponse represents invite details (for accepting)
type InviteDetailsResponse struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`
	ExpiresAt   string `json:"expiresAt"`
	IsExpired   bool   `json:"isExpired"`
	IsValid     bool   `json:"isValid"`
}

// ProjectResponse represents a project response
type ProjectResponse struct {
	ID          string `json:"id"`
	OwnerID     string `json:"ownerId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Owner       struct {
		ID        string `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"owner"`
}

// ListProjects handles listing projects for a user
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
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

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	projects, err := h.queries.ListProjectsByUser(r.Context(), repo.ListProjectsByUserParams{
		OwnerID: userUUID,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		response.InternalServerError(w, "Failed to list projects")
		return
	}

	// Set cache headers for list endpoint (60 seconds)
	w.Header().Set("Cache-Control", "private, max-age=60")

	// Convert to response format
	var projectResponses []ProjectResponse
	for _, project := range projects {
		projectResponses = append(projectResponses, ProjectResponse{
			ID:          project.ID.String(),
			OwnerID:     project.OwnerID.String(),
			Name:        project.Name,
			Description: *project.Description,
			CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Owner: struct {
				ID        string `json:"id"`
				Username  string `json:"username"`
				Email     string `json:"email"`
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			}{
				ID:        project.OwnerID.String(),
				Username:  project.OwnerUsername,
				Email:     project.OwnerEmail,
				FirstName: project.OwnerFirstName,
				LastName:  project.OwnerLastName,
			},
		})
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"projects": projectResponses,
		"limit":    limit,
		"offset":   offset,
	})
}

// CreateProject handles project creation
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	var req CreateProjectRequest
	if !response.Decode(w, r, &req) {
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	project, err := h.queries.CreateProject(r.Context(), repo.CreateProjectParams{
		OwnerID:     userUUID,
		Name:        req.Name,
		Description: &req.Description,
	})
	if err != nil {
		response.BadRequest(w, "Failed to create project: "+err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, ProjectResponse{
		ID:          project.ID.String(),
		OwnerID:     project.OwnerID.String(),
		Name:        project.Name,
		Description: *project.Description,
		CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// JoinProject handles joining a project by project code/ID
// This endpoint requires authentication but does not require prior membership.
// It adds the authenticated user as a member of the target project (idempotently).
func (h *ProjectHandler) JoinProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	var req JoinProjectRequest
	if !response.Decode(w, r, &req) {
		return
	}

	if strings.TrimSpace(req.ProjectID) == "" {
		response.BadRequest(w, "Project ID is required")
		return
	}

	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		response.BadRequest(w, "Invalid project ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	// Ensure project exists
	project, err := h.queries.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	// CRITICAL: Always set role to "member" - never allow "owner"
	// Project ownership is determined by projects.owner_id, not project_members.role
	// Users joining via invite code should never become project owners
	if err := h.queries.AddProjectMember(r.Context(), repo.AddProjectMemberParams{
		ProjectID: projectUUID,
		UserID:    userUUID,
		Role:      "member", // Always "member" - owner role is not allowed
	}); err != nil {
		response.InternalServerError(w, "Failed to join project")
		return
	}

	// Explicitly broadcast cache invalidation for project_members change
	// This ensures all users viewing the project get notified immediately when someone joins
	if h.hub != nil {
		messageData := map[string]interface{}{
			"resource":   "project_members",
			"action":     "INSERT",
			"project_id": projectUUID.String(),
			"timestamp":  time.Now().Format(time.RFC3339),
		}
		h.hub.BroadcastToProject(projectUUID.String(), "cache_invalidate", messageData)
	}

	// Return project details so the frontend can navigate to it
	response.JSON(w, http.StatusOK, ProjectResponse{
		ID:          project.ID.String(),
		OwnerID:     project.OwnerID.String(),
		Name:        project.Name,
		Description: *project.Description,
		CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Owner: struct {
			ID        string `json:"id"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}{
			ID:        project.OwnerID.String(),
			Username:  project.OwnerUsername,
			Email:     project.OwnerEmail,
			FirstName: project.OwnerFirstName,
			LastName:  project.OwnerLastName,
		},
	})
}

// GetProject handles getting a project by ID
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	// Set cache headers for single project endpoint (5 minutes)
	w.Header().Set("Cache-Control", "private, max-age=300")

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

	project, err := h.queries.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	response.JSON(w, http.StatusOK, ProjectResponse{
		ID:          project.ID.String(),
		OwnerID:     project.OwnerID.String(),
		Name:        project.Name,
		Description: *project.Description,
		CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Owner: struct {
			ID        string `json:"id"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}{
			ID:        project.OwnerID.String(),
			Username:  project.OwnerUsername,
			Email:     project.OwnerEmail,
			FirstName: project.OwnerFirstName,
			LastName:  project.OwnerLastName,
		},
	})
}

// GetProjectBundle handles getting a project with optional includes (members, owner, etc.)
func (h *ProjectHandler) GetProjectBundle(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		response.BadRequest(w, "Project ID is required")
		return
	}

	// Parse include parameter
	include := r.URL.Query().Get("include")
	includeMembers := strings.Contains(include, "members")
	includeOwner := strings.Contains(include, "owner")
	includeSprints := strings.Contains(include, "sprints")
	includeTasks := strings.Contains(include, "tasks")

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		response.BadRequest(w, "Invalid project ID")
		return
	}

	// Get project
	project, err := h.queries.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	// Build response bundle
	bundle := map[string]interface{}{
		"project": ProjectResponse{
			ID:          project.ID.String(),
			OwnerID:     project.OwnerID.String(),
			Name:        project.Name,
			Description: *project.Description,
			CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	// Add owner if requested
	if includeOwner {
		bundle["owner"] = map[string]interface{}{
			"id":        project.OwnerID.String(),
			"username":  project.OwnerUsername,
			"email":     project.OwnerEmail,
			"firstName": project.OwnerFirstName,
			"lastName":  project.OwnerLastName,
		}
	}

	// Add members if requested
	if includeMembers {
		members, err := h.queries.GetProjectMembers(r.Context(), projectUUID)
		if err == nil {
			bundle["members"] = members
		}
	}

	// Add sprints if requested
	if includeSprints {
		sprints, err := h.queries.ListSprintsByProject(r.Context(), repo.ListSprintsByProjectParams{
			ProjectID: projectUUID,
			Limit:     50,
			Offset:    0,
		})
		if err == nil {
			bundle["sprints"] = sprints
		}
	}

	// Add tasks if requested
	if includeTasks {
		tasks, err := h.queries.ListTasksByProject(r.Context(), repo.ListTasksByProjectParams{
			ProjectID: projectUUID,
			Limit:     50,
			Offset:    0,
		})
		if err == nil {
			bundle["tasks"] = tasks
		}
	}

	// Set caching headers
	w.Header().Set("Cache-Control", "private, max-age=60")
	w.Header().Set("ETag", `"`+project.ID.String()+`-`+strconv.FormatInt(project.UpdatedAt.Unix(), 10)+`"`)

	response.JSON(w, http.StatusOK, bundle)
}

// UpdateProject handles project updates
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateProjectRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Get current project to merge updates
	currentProject, err := h.queries.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	// Merge updates
	name := currentProject.Name
	description := *currentProject.Description

	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		description = *req.Description
	}

	project, err := h.queries.UpdateProject(r.Context(), repo.UpdateProjectParams{
		ID:          projectUUID,
		Name:        name,
		Description: &description,
	})
	if err != nil {
		response.BadRequest(w, "Failed to update project: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, ProjectResponse{
		ID:          project.ID.String(),
		OwnerID:     project.OwnerID.String(),
		Name:        project.Name,
		Description: *project.Description,
		CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteProject handles project deletion
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is the project owner (only owners can delete projects)
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

	// Verify user is the project owner
	isOwner, err := h.queries.CheckProjectOwner(r.Context(), repo.CheckProjectOwnerParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to verify project ownership")
		return
	}
	if !isOwner {
		response.Forbidden(w, "Only project owners can delete projects")
		return
	}

	// Verify project exists before deletion
	_, err = h.queries.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	err = h.queries.DeleteProject(r.Context(), projectUUID)
	if err != nil {
		response.InternalServerError(w, "Failed to delete project")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Project deleted successfully"})
}

// AddMember handles adding a member to a project
func (h *ProjectHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	projectID := chi.URLParam(r, "projectId")
	memberID := chi.URLParam(r, "userId")
	role := r.URL.Query().Get("role")
	if role == "" {
		role = "member"
	}

	// CRITICAL: Prevent setting role to "owner"
	// Owner is determined by projects.owner_id, not project_members.role
	if strings.EqualFold(role, "owner") {
		response.BadRequest(w, "Cannot set role to 'owner'. Owner is determined by project ownership.")
		return
	}

	// Validate role is one of allowed values
	allowedRoles := []string{"member", "admin", "viewer"}
	roleValid := false
	for _, allowedRole := range allowedRoles {
		if strings.EqualFold(role, allowedRole) {
			role = allowedRole // Normalize case
			roleValid = true
			break
		}
	}
	if !roleValid {
		response.BadRequest(w, "Invalid role. Allowed roles: member, admin, viewer")
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

	memberUUID, err := uuid.Parse(memberID)
	if err != nil {
		response.BadRequest(w, "Invalid member ID")
		return
	}

	err = h.queries.AddProjectMember(r.Context(), repo.AddProjectMemberParams{
		ProjectID: projectUUID,
		UserID:    memberUUID,
		Role:      role,
	})
	if err != nil {
		response.BadRequest(w, "Failed to add member: "+err.Error())
		return
	}

	// Explicitly broadcast cache invalidation for project_members change
	// This ensures all users viewing the project get notified immediately
	if h.hub != nil {
		messageData := map[string]interface{}{
			"resource":   "project_members",
			"action":     "INSERT",
			"project_id": projectUUID.String(),
			"timestamp":  time.Now().Format(time.RFC3339),
		}
		h.hub.BroadcastToProject(projectUUID.String(), "cache_invalidate", messageData)
		log.Printf("Explicit broadcast: project_members INSERT for project %s", projectUUID.String())
	} else {
		log.Printf("ERROR: Hub is nil, cannot broadcast cache invalidation for project %s", projectUUID.String())
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Member added successfully"})
}

// RemoveMember handles removing a member from a project
func (h *ProjectHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	projectID := chi.URLParam(r, "projectId")
	memberID := chi.URLParam(r, "userId")

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

	memberUUID, err := uuid.Parse(memberID)
	if err != nil {
		response.BadRequest(w, "Invalid member ID")
		return
	}

	err = h.queries.RemoveProjectMember(r.Context(), repo.RemoveProjectMemberParams{
		ProjectID: projectUUID,
		UserID:    memberUUID,
	})
	if err != nil {
		response.BadRequest(w, "Failed to remove member: "+err.Error())
		return
	}

	// Explicitly broadcast cache invalidation for project_members change
	// This ensures all users viewing the project get notified immediately
	if h.hub != nil {
		messageData := map[string]interface{}{
			"resource":   "project_members",
			"action":     "DELETE",
			"project_id": projectUUID.String(),
			"timestamp":  time.Now().Format(time.RFC3339),
		}
		h.hub.BroadcastToProject(projectUUID.String(), "cache_invalidate", messageData)
		log.Printf("Explicit broadcast: project_members DELETE for project %s", projectUUID.String())
	} else {
		log.Printf("ERROR: Hub is nil, cannot broadcast cache invalidation for project %s", projectUUID.String())
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Member removed successfully"})
}

// ListMembers handles listing project members
func (h *ProjectHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
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

	// Get project members
	members, err := h.queries.ListProjectMembers(r.Context(), projectUUID)
	if err != nil {
		response.BadRequest(w, "Failed to get project members: "+err.Error())
		return
	}

	// Convert to response format
	var memberResponses []map[string]interface{}
	for _, member := range members {
		memberResponses = append(memberResponses, map[string]interface{}{
			"id":        member.ID.String(),
			"username":  member.Username,
			"email":     member.Email,
			"firstName": member.FirstName,
			"lastName":  member.LastName,
			"joinedAt":  member.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
			"role":      member.Role,
		})
	}

	// Ensure members is always an array, never null
	if memberResponses == nil {
		memberResponses = []map[string]interface{}{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"members": memberResponses,
		"count":   len(memberResponses),
	})
}

// CreateInvite handles creating a project invite link
func (h *ProjectHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	projectID := chi.URLParam(r, "projectId")
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

	// Check if user has access to project (must be owner or member)
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to project")
		return
	}

	var req CreateInviteRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Default expiration: 30 minutes
	expiresInMinutes := 30
	if req.ExpiresInMinutes != nil && *req.ExpiresInMinutes > 0 {
		expiresInMinutes = *req.ExpiresInMinutes
	}

	expiresAt := time.Now().Add(time.Duration(expiresInMinutes) * time.Minute)

	// Generate secure invite token (UUID-based)
	inviteToken := uuid.New().String()

	// Create invite
	var maxUsesPtr *int32
	if req.MaxUses != nil && *req.MaxUses > 0 {
		maxUsesVal := int32(*req.MaxUses)
		maxUsesPtr = &maxUsesVal
	}

	invite, err := h.queries.CreateProjectInvite(r.Context(), repo.CreateProjectInviteParams{
		ProjectID:   projectUUID,
		CreatedBy:   userUUID,
		InviteToken: inviteToken,
		ExpiresAt:   expiresAt,
		MaxUses:     maxUsesPtr,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to create invite: "+err.Error())
		return
	}

	// Build invite URL (frontend will handle routing)
	// Get frontend URL from request or use default
	frontendURL := r.Header.Get("Origin")
	if frontendURL == "" {
		frontendURL = "https://devhive.it.com" // Default frontend URL
	}
	inviteURL := fmt.Sprintf("%s/invite/%s", frontendURL, inviteToken)

	var maxUsesResponse *int
	if invite.MaxUses != nil {
		maxUsesVal := int(*invite.MaxUses)
		maxUsesResponse = &maxUsesVal
	}

	response.JSON(w, http.StatusCreated, InviteResponse{
		ID:          invite.ID.String(),
		ProjectID:   invite.ProjectID.String(),
		InviteToken: invite.InviteToken,
		InviteURL:   inviteURL,
		ExpiresAt:   invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		MaxUses:     maxUsesResponse,
		UsedCount:   int(invite.UsedCount),
		IsActive:    invite.IsActive,
		CreatedAt:   invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetInviteDetails handles getting invite details by token (public endpoint, no auth required)
func (h *ProjectHandler) GetInviteDetails(w http.ResponseWriter, r *http.Request) {
	inviteToken := chi.URLParam(r, "inviteToken")
	if inviteToken == "" {
		response.BadRequest(w, "Invite token is required")
		return
	}

	invite, err := h.queries.GetProjectInviteByToken(r.Context(), inviteToken)
	if err != nil {
		response.NotFound(w, "Invite not found or expired")
		return
	}

	// Check if expired
	isExpired := time.Now().After(invite.ExpiresAt)
	isValid := invite.IsActive && !isExpired &&
		(invite.MaxUses == nil || invite.UsedCount < *invite.MaxUses)

	// Get project details
	project, err := h.queries.GetProjectByID(r.Context(), invite.ProjectID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	response.JSON(w, http.StatusOK, InviteDetailsResponse{
		ID:          invite.ID.String(),
		ProjectID:   invite.ProjectID.String(),
		ProjectName: project.Name,
		ExpiresAt:   invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		IsExpired:   isExpired,
		IsValid:     isValid,
	})
}

// AcceptInvite handles accepting an invite and joining the project
func (h *ProjectHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	inviteToken := chi.URLParam(r, "inviteToken")
	if inviteToken == "" {
		response.BadRequest(w, "Invite token is required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	// Get invite
	invite, err := h.queries.GetProjectInviteByToken(r.Context(), inviteToken)
	if err != nil {
		response.NotFound(w, "Invite not found or expired")
		return
	}

	// Validate invite
	isExpired := time.Now().After(invite.ExpiresAt)
	if !invite.IsActive {
		response.BadRequest(w, "Invite has been deactivated")
		return
	}
	if isExpired {
		response.BadRequest(w, "Invite has expired")
		return
	}
	if invite.MaxUses != nil && invite.UsedCount >= *invite.MaxUses {
		response.BadRequest(w, "Invite has reached maximum uses")
		return
	}

	// Check if user is already a member
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      invite.ProjectID,
		OwnerID: userUUID,
	})
	if err == nil && hasAccess {
		// User is already a member, return project details
		project, err := h.queries.GetProjectByID(r.Context(), invite.ProjectID)
		if err != nil {
			response.InternalServerError(w, "Failed to get project details")
			return
		}

		response.JSON(w, http.StatusOK, ProjectResponse{
			ID:          project.ID.String(),
			OwnerID:     project.OwnerID.String(),
			Name:        project.Name,
			Description: *project.Description,
			CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Owner: struct {
				ID        string `json:"id"`
				Username  string `json:"username"`
				Email     string `json:"email"`
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			}{
				ID:        project.OwnerID.String(),
				Username:  project.OwnerUsername,
				Email:     project.OwnerEmail,
				FirstName: project.OwnerFirstName,
				LastName:  project.OwnerLastName,
			},
		})
		return
	}

	// Increment use count
	err = h.queries.IncrementInviteUseCount(r.Context(), invite.ID)
	if err != nil {
		response.InternalServerError(w, "Failed to update invite")
		return
	}

	// Add user as member
	err = h.queries.AddProjectMember(r.Context(), repo.AddProjectMemberParams{
		ProjectID: invite.ProjectID,
		UserID:    userUUID,
		Role:      "member", // Always "member" for invites
	})
	if err != nil {
		response.InternalServerError(w, "Failed to join project")
		return
	}

	// Explicitly broadcast cache invalidation for project_members change
	// This ensures all users viewing the project get notified immediately when someone joins
	if h.hub != nil {
		messageData := map[string]interface{}{
			"resource":   "project_members",
			"action":     "INSERT",
			"project_id": invite.ProjectID.String(),
			"timestamp":  time.Now().Format(time.RFC3339),
		}
		h.hub.BroadcastToProject(invite.ProjectID.String(), "cache_invalidate", messageData)
		log.Printf("Explicit broadcast: project_members INSERT (via invite) for project %s", invite.ProjectID.String())
	} else {
		log.Printf("ERROR: Hub is nil, cannot broadcast cache invalidation for project %s", invite.ProjectID.String())
	}

	// Get project details
	project, err := h.queries.GetProjectByID(r.Context(), invite.ProjectID)
	if err != nil {
		response.InternalServerError(w, "Failed to get project details")
		return
	}

	// Return project details
	response.JSON(w, http.StatusOK, ProjectResponse{
		ID:          project.ID.String(),
		OwnerID:     project.OwnerID.String(),
		Name:        project.Name,
		Description: *project.Description,
		CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Owner: struct {
			ID        string `json:"id"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}{
			ID:        project.OwnerID.String(),
			Username:  project.OwnerUsername,
			Email:     project.OwnerEmail,
			FirstName: project.OwnerFirstName,
			LastName:  project.OwnerLastName,
		},
	})
}

// ListInvites handles listing all active invites for a project
func (h *ProjectHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	projectID := chi.URLParam(r, "projectId")
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

	// Check if user has access to project
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to project")
		return
	}

	// Get all active invites for the project
	invites, err := h.queries.ListProjectInvites(r.Context(), projectUUID)
	if err != nil {
		response.InternalServerError(w, "Failed to list invites")
		return
	}

	// Get frontend URL from request or use default
	frontendURL := r.Header.Get("Origin")
	if frontendURL == "" {
		frontendURL = "https://devhive.it.com" // Default frontend URL
	}

	// Convert to response format
	var inviteResponses []InviteResponse
	for _, invite := range invites {
		var maxUsesPtr *int
		if invite.MaxUses != nil {
			maxUsesVal := int(*invite.MaxUses)
			maxUsesPtr = &maxUsesVal
		}

		inviteURL := fmt.Sprintf("%s/invite/%s", frontendURL, invite.InviteToken)

		inviteResponses = append(inviteResponses, InviteResponse{
			ID:          invite.ID.String(),
			ProjectID:   invite.ProjectID.String(),
			InviteToken: invite.InviteToken,
			InviteURL:   inviteURL,
			ExpiresAt:   invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			MaxUses:     maxUsesPtr,
			UsedCount:   int(invite.UsedCount),
			IsActive:    invite.IsActive,
			CreatedAt:   invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	if inviteResponses == nil {
		inviteResponses = []InviteResponse{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"invites": inviteResponses,
		"count":   len(inviteResponses),
	})
}

// RevokeInvite handles deactivating an invite
func (h *ProjectHandler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	projectID := chi.URLParam(r, "projectId")
	inviteID := chi.URLParam(r, "inviteId")

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		response.BadRequest(w, "Invalid project ID")
		return
	}

	inviteUUID, err := uuid.Parse(inviteID)
	if err != nil {
		response.BadRequest(w, "Invalid invite ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	// Check if user has access to project
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to project")
		return
	}

	// Get invite to verify it belongs to the project
	invite, err := h.queries.GetProjectInviteByID(r.Context(), inviteUUID)
	if err != nil {
		response.NotFound(w, "Invite not found")
		return
	}

	// Verify invite belongs to the project
	if invite.ProjectID != projectUUID {
		response.Forbidden(w, "Invite does not belong to this project")
		return
	}

	// Deactivate invite
	err = h.queries.DeactivateInvite(r.Context(), inviteUUID)
	if err != nil {
		response.InternalServerError(w, "Failed to revoke invite")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Invite revoked successfully"})
}
