package handlers

import (
	"net/http"
	"strconv"

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

	err = h.queries.RemoveProjectMember(r.Context(), repo.RemoveProjectMemberParams{
		ProjectID: projectUUID,
		UserID:    memberUUID,
	})
	if err != nil {
		response.BadRequest(w, "Failed to remove member: "+err.Error())
		return
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
