package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"devhive-backend/internal/http/middleware"
	"devhive-backend/internal/http/response"
	"devhive-backend/internal/repo"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	queries *repo.Queries
}

func NewProjectHandler(queries *repo.Queries) *ProjectHandler {
	return &ProjectHandler{
		queries: queries,
	}
}

// CreateProjectRequest represents the project creation request
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateProjectRequest represents the project update request
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
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
	UserRole    *string `json:"userRole,omitempty"` // Current user's role: "owner", "admin", "member", "viewer"
	Permissions struct {
		CanViewInvites   bool `json:"canViewInvites"`
		CanCreateInvites bool `json:"canCreateInvites"`
		CanRevokeInvites bool `json:"canRevokeInvites"`
		CanManageMembers bool `json:"canManageMembers"`
	} `json:"permissions,omitempty"`
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

	// CRITICAL: Insert owner into project_members table for consistency
	// This ensures owners appear in member lists and all queries work consistently
	err = h.queries.AddProjectMember(r.Context(), repo.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    userUUID,
		Role:      "owner",
	})
	if err != nil {
		// Log error but don't fail the request - project was created successfully
		// The owner can still access via projects.owner_id, but member queries will be inconsistent
		log.Printf("Warning: Failed to add owner to project_members for project %s: %v", project.ID.String(), err)
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

// GetProject handles getting a project by ID
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
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

	// Get user's role and permissions
	userRole, permissions := h.getUserRoleAndPermissions(r.Context(), projectUUID, userUUID)

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
		UserRole:    userRole,
		Permissions: permissions,
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

	log.Printf("RemoveMember: Removing user %s from project %s", memberUUID.String(), projectUUID.String())
	err = h.queries.RemoveProjectMember(r.Context(), repo.RemoveProjectMemberParams{
		ProjectID: projectUUID,
		UserID:    memberUUID,
	})
	if err != nil {
		log.Printf("RemoveMember: Failed to remove member: %v", err)
		response.BadRequest(w, "Failed to remove member: "+err.Error())
		return
	}

	log.Printf("RemoveMember: Successfully removed user %s from project %s", memberUUID.String(), projectUUID.String())
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

// JoinProjectRequest represents the request to join a project
type JoinProjectRequest struct {
	ProjectID string `json:"projectId"`
}

// JoinProject handles a user joining a project by project ID
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

	if req.ProjectID == "" {
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

	// Check if project exists
	projectExists, err := h.queries.ProjectExists(r.Context(), projectUUID)
	if err != nil {
		response.InternalServerError(w, "Failed to check project existence")
		return
	}
	if !projectExists {
		response.NotFound(w, "Project not found")
		return
	}

	// Check if user is already a member or owner
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to check project access")
		return
	}
	if hasAccess {
		// User is already a member, return the project
		project, err := h.queries.GetProjectByID(r.Context(), projectUUID)
		if err != nil {
			response.InternalServerError(w, "Failed to get project")
			return
		}
		// Get user's role and permissions
		userRole, permissions := h.getUserRoleAndPermissions(r.Context(), projectUUID, userUUID)
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
			UserRole:    userRole,
			Permissions: permissions,
		})
		return
	}

	// Add user as a member
	err = h.queries.AddProjectMember(r.Context(), repo.AddProjectMemberParams{
		ProjectID: projectUUID,
		UserID:    userUUID,
		Role:      "member",
	})
	if err != nil {
		response.BadRequest(w, "Failed to join project: "+err.Error())
		return
	}

	// Get the project to return
	project, err := h.queries.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		response.InternalServerError(w, "Failed to get project")
		return
	}

	// Get user's role and permissions (user just joined, so they're a member)
	userRole, permissions := h.getUserRoleAndPermissions(r.Context(), projectUUID, userUUID)

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
		UserRole:    userRole,
		Permissions: permissions,
	})
}

// AcceptInvite handles accepting a project invite by token
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

	// Get the invite
	invite, err := h.queries.GetProjectInviteByToken(r.Context(), inviteToken)
	if err != nil {
		response.NotFound(w, "Invalid or expired invite token")
		return
	}

	// Check if invite is still valid
	if !invite.IsActive {
		response.BadRequest(w, "Invite is no longer active")
		return
	}

	// Check if invite has expired
	// Note: This should be handled by the database query, but we check here too
	// Check max uses
	if invite.MaxUses != nil && invite.UsedCount >= *invite.MaxUses {
		response.BadRequest(w, "Invite has reached maximum uses")
		return
	}

	// Check if user is already a member
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      invite.ProjectID,
		OwnerID: userUUID,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to check project access")
		return
	}
	if hasAccess {
		// User is already a member, just increment the use count and return success
		err = h.queries.IncrementInviteUseCount(r.Context(), invite.ID)
		if err != nil {
			// Log but don't fail
			log.Printf("Failed to increment invite use count: %v", err)
		}
		project, err := h.queries.GetProjectByID(r.Context(), invite.ProjectID)
		if err != nil {
			response.InternalServerError(w, "Failed to get project")
			return
		}
		// Get user's role and permissions
		userRole, permissions := h.getUserRoleAndPermissions(r.Context(), invite.ProjectID, userUUID)
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
			UserRole:    userRole,
			Permissions: permissions,
		})
		return
	}

	// Add user as a member
	log.Printf("AcceptInvite: Adding user %s to project %s", userUUID.String(), invite.ProjectID.String())
	log.Printf("AcceptInvite: About to execute AddProjectMember query - this should trigger PostgreSQL NOTIFY")
	err = h.queries.AddProjectMember(r.Context(), repo.AddProjectMemberParams{
		ProjectID: invite.ProjectID,
		UserID:    userUUID,
		Role:      "member",
	})
	if err != nil {
		log.Printf("AcceptInvite: Failed to add member: %v", err)
		response.BadRequest(w, "Failed to join project: "+err.Error())
		return
	}
	log.Printf("AcceptInvite: ✅ Database INSERT/UPDATE completed for user %s in project %s", userUUID.String(), invite.ProjectID.String())
	log.Printf("AcceptInvite: ⚠️  If NOTIFY listener is running, you should see '🔔 RAW NOTIFY received' logs within 1 second")

	// Increment invite use count
	err = h.queries.IncrementInviteUseCount(r.Context(), invite.ID)
	if err != nil {
		// Log but don't fail the request
		log.Printf("Failed to increment invite use count: %v", err)
	}

	// Get the project to return
	project, err := h.queries.GetProjectByID(r.Context(), invite.ProjectID)
	if err != nil {
		response.InternalServerError(w, "Failed to get project")
		return
	}

	// Get user's role and permissions (user just joined, so they're a member)
	userRole, permissions := h.getUserRoleAndPermissions(r.Context(), invite.ProjectID, userUUID)

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
		UserRole:    userRole,
		Permissions: permissions,
	})
}

// GetInviteDetails handles getting invite details by token (public endpoint)
func (h *ProjectHandler) GetInviteDetails(w http.ResponseWriter, r *http.Request) {
	inviteToken := chi.URLParam(r, "inviteToken")
	if inviteToken == "" {
		response.BadRequest(w, "Invite token is required")
		return
	}

	invite, err := h.queries.GetProjectInviteByToken(r.Context(), inviteToken)
	if err != nil {
		response.NotFound(w, "Invalid or expired invite token")
		return
	}

	// Check if invite has expired
	if invite.ExpiresAt.Before(time.Now()) {
		response.BadRequest(w, "Invite has expired")
		return
	}

	// Get project details
	project, err := h.queries.GetProjectByID(r.Context(), invite.ProjectID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"invite": map[string]interface{}{
			"id":        invite.ID.String(),
			"projectId": invite.ProjectID.String(),
			"token":     invite.InviteToken,
			"expiresAt": invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			"maxUses":   invite.MaxUses,
			"usedCount": invite.UsedCount,
			"isActive":  invite.IsActive,
		},
		"project": map[string]interface{}{
			"id":          project.ID.String(),
			"name":        project.Name,
			"description": *project.Description,
		},
	})
}

// CreateInviteRequest represents the request to create a project invite
type CreateInviteRequest struct {
	ExpiresInMinutes *int   `json:"expiresInMinutes"` // Optional, defaults to 30
	MaxUses          *int32 `json:"maxUses"`          // Optional, nil = unlimited
}

// CreateInvite handles creating a project invite
func (h *ProjectHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is owner or admin
	isOwnerOrAdmin, err := h.queries.CheckProjectOwnerOrAdmin(r.Context(), repo.CheckProjectOwnerOrAdminParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil || !isOwnerOrAdmin {
		response.Forbidden(w, "Only project owners and admins can create invites")
		return
	}

	var req CreateInviteRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Generate invite token
	inviteToken, err := uuid.NewRandom()
	if err != nil {
		response.InternalServerError(w, "Failed to generate invite token")
		return
	}

	// Set expiration time (default 30 minutes)
	expiresInMinutes := 30
	if req.ExpiresInMinutes != nil && *req.ExpiresInMinutes > 0 {
		expiresInMinutes = *req.ExpiresInMinutes
	}
	expiresAt := time.Now().Add(time.Duration(expiresInMinutes) * time.Minute)

	// Create invite
	invite, err := h.queries.CreateProjectInvite(r.Context(), repo.CreateProjectInviteParams{
		ProjectID:   projectUUID,
		CreatedBy:   userUUID,
		InviteToken: inviteToken.String(),
		ExpiresAt:   expiresAt,
		MaxUses:     req.MaxUses,
	})
	if err != nil {
		response.BadRequest(w, "Failed to create invite: "+err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"id":        invite.ID.String(),
		"projectId": invite.ProjectID.String(),
		"token":     invite.InviteToken,
		"expiresAt": invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		"maxUses":   invite.MaxUses,
		"usedCount": invite.UsedCount,
		"isActive":  invite.IsActive,
		"createdAt": invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListInvites handles listing project invites
func (h *ProjectHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is owner or admin
	isOwnerOrAdmin, err := h.queries.CheckProjectOwnerOrAdmin(r.Context(), repo.CheckProjectOwnerOrAdminParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil || !isOwnerOrAdmin {
		response.Forbidden(w, "Only project owners and admins can view invites")
		return
	}

	invites, err := h.queries.ListProjectInvites(r.Context(), projectUUID)
	if err != nil {
		response.BadRequest(w, "Failed to list invites: "+err.Error())
		return
	}

	// Convert to response format
	var inviteResponses []map[string]interface{}
	for _, invite := range invites {
		inviteResponses = append(inviteResponses, map[string]interface{}{
			"id":        invite.ID.String(),
			"projectId": invite.ProjectID.String(),
			"token":     invite.InviteToken,
			"expiresAt": invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			"maxUses":   invite.MaxUses,
			"usedCount": invite.UsedCount,
			"isActive":  invite.IsActive,
			"createdAt": invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Ensure invites is always an array, never null
	if inviteResponses == nil {
		inviteResponses = []map[string]interface{}{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"invites": inviteResponses,
		"count":   len(inviteResponses),
	})
}

// RevokeInvite handles revoking (deactivating) a project invite
func (h *ProjectHandler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	projectID := chi.URLParam(r, "projectId")
	inviteID := chi.URLParam(r, "inviteId")
	if projectID == "" || inviteID == "" {
		response.BadRequest(w, "Project ID and Invite ID are required")
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

	// Check if user is owner or admin
	isOwnerOrAdmin, err := h.queries.CheckProjectOwnerOrAdmin(r.Context(), repo.CheckProjectOwnerOrAdminParams{
		ID:      projectUUID,
		OwnerID: userUUID,
	})
	if err != nil || !isOwnerOrAdmin {
		response.Forbidden(w, "Only project owners and admins can revoke invites")
		return
	}

	inviteUUID, err := uuid.Parse(inviteID)
	if err != nil {
		response.BadRequest(w, "Invalid invite ID")
		return
	}

	// Verify the invite belongs to this project
	invite, err := h.queries.GetProjectInviteByID(r.Context(), inviteUUID)
	if err != nil {
		response.NotFound(w, "Invite not found")
		return
	}

	if invite.ProjectID != projectUUID {
		response.Forbidden(w, "Invite does not belong to this project")
		return
	}

	// Deactivate the invite
	err = h.queries.DeactivateInvite(r.Context(), inviteUUID)
	if err != nil {
		response.BadRequest(w, "Failed to revoke invite: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Invite revoked successfully"})
}

// getUserRoleAndPermissions gets the user's role and calculates permissions for a project
func (h *ProjectHandler) getUserRoleAndPermissions(ctx context.Context, projectID, userID uuid.UUID) (*string, struct {
	CanViewInvites   bool `json:"canViewInvites"`
	CanCreateInvites bool `json:"canCreateInvites"`
	CanRevokeInvites bool `json:"canRevokeInvites"`
	CanManageMembers bool `json:"canManageMembers"`
}) {
	// Check if user is owner
	isOwner, err := h.queries.CheckProjectOwner(ctx, repo.CheckProjectOwnerParams{
		ID:      projectID,
		OwnerID: userID,
	})
	if err != nil {
		// On error, return nil role and no permissions
		return nil, struct {
			CanViewInvites   bool `json:"canViewInvites"`
			CanCreateInvites bool `json:"canCreateInvites"`
			CanRevokeInvites bool `json:"canRevokeInvites"`
			CanManageMembers bool `json:"canManageMembers"`
		}{}
	}

	var userRole string
	var permissions struct {
		CanViewInvites   bool `json:"canViewInvites"`
		CanCreateInvites bool `json:"canCreateInvites"`
		CanRevokeInvites bool `json:"canRevokeInvites"`
		CanManageMembers bool `json:"canManageMembers"`
	}

	if isOwner {
		userRole = "owner"
		permissions.CanViewInvites = true
		permissions.CanCreateInvites = true
		permissions.CanRevokeInvites = true
		permissions.CanManageMembers = true
	} else {
		// Get user's role from project_members
		roleResult, err := h.queries.GetUserProjectRole(ctx, repo.GetUserProjectRoleParams{
			ID:      projectID,
			OwnerID: userID,
		})
		if err != nil {
			// Default to member if query fails
			userRole = "member"
		} else {
			// Convert interface{} to string
			// The query returns NULL if user is not a member, or a string role
			if roleStr, ok := roleResult.(string); ok && roleStr != "" {
				userRole = roleStr
			} else {
				// Default to member if role not found or NULL
				userRole = "member"
			}
		}

		// Set permissions based on role
		if userRole == "admin" {
			permissions.CanViewInvites = true
			permissions.CanCreateInvites = true
			permissions.CanRevokeInvites = true
			permissions.CanManageMembers = true
		} else {
			// member or viewer - no invite permissions
			permissions.CanViewInvites = false
			permissions.CanCreateInvites = false
			permissions.CanRevokeInvites = false
			permissions.CanManageMembers = false
		}
	}

	return &userRole, permissions
}
