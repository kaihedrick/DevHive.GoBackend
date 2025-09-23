package handlers

import (
	"net/http"
	"strconv"
	"time"

	"devhive-backend/internal/http/middleware"
	"devhive-backend/internal/http/response"
	"devhive-backend/internal/repo"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SprintHandler struct {
	queries *repo.Queries
}

func NewSprintHandler(queries *repo.Queries) *SprintHandler {
	return &SprintHandler{
		queries: queries,
	}
}

// CreateSprintRequest represents the sprint creation request
type CreateSprintRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
}

// UpdateSprintRequest represents the sprint update request
type UpdateSprintRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	StartDate   *string `json:"startDate,omitempty"`
	EndDate     *string `json:"endDate,omitempty"`
}

// SprintResponse represents a sprint response
type SprintResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   string    `json:"startDate"`
	EndDate     string    `json:"endDate"`
	IsCompleted bool      `json:"isCompleted"`
	IsStarted   bool      `json:"isStarted"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	Owner       OwnerInfo `json:"owner"`
}

// OwnerInfo represents project owner information
type OwnerInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// ListSprintsByProject handles listing sprints for a project
func (h *SprintHandler) ListSprintsByProject(w http.ResponseWriter, r *http.Request) {
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

	sprints, err := h.queries.ListSprintsByProject(r.Context(), repo.ListSprintsByProjectParams{
		ProjectID: projectUUID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		response.InternalServerError(w, "Failed to list sprints")
		return
	}

	// Convert to response format
	var sprintResponses []SprintResponse
	for _, sprint := range sprints {
		description := ""
		if sprint.Description != nil {
			description = *sprint.Description
		}

		sprintResponses = append(sprintResponses, SprintResponse{
			ID:          sprint.ID.String(),
			ProjectID:   sprint.ProjectID.String(),
			Name:        sprint.Name,
			Description: description,
			StartDate:   sprint.StartDate.Format("2006-01-02T15:04:05Z07:00"),
			EndDate:     sprint.EndDate.Format("2006-01-02T15:04:05Z07:00"),
			IsCompleted: sprint.IsCompleted,
			IsStarted:   sprint.IsStarted,
			CreatedAt:   sprint.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   sprint.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Owner: OwnerInfo{
				ID:        sprint.OwnerID.String(),
				Username:  sprint.OwnerUsername,
				Email:     sprint.OwnerEmail,
				FirstName: sprint.OwnerFirstName,
				LastName:  sprint.OwnerLastName,
			},
		})
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"sprints": sprintResponses,
		"limit":   limit,
		"offset":  offset,
	})
}

// CreateSprint handles sprint creation
func (h *SprintHandler) CreateSprint(w http.ResponseWriter, r *http.Request) {
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

	var req CreateSprintRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Parse dates - try multiple formats
	var startDate, endDate time.Time

	// Try RFC3339 format first
	startDate, err = time.Parse("2006-01-02T15:04:05Z07:00", req.StartDate)
	if err != nil {
		// Try simple date format
		startDate, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			response.BadRequest(w, "Invalid start date format. Expected YYYY-MM-DD or RFC3339")
			return
		}
	}

	endDate, err = time.Parse("2006-01-02T15:04:05Z07:00", req.EndDate)
	if err != nil {
		// Try simple date format
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			response.BadRequest(w, "Invalid end date format. Expected YYYY-MM-DD or RFC3339")
			return
		}
	}

	sprint, err := h.queries.CreateSprint(r.Context(), repo.CreateSprintParams{
		ProjectID:   projectUUID,
		Name:        req.Name,
		Description: &req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		response.BadRequest(w, "Failed to create sprint: "+err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, SprintResponse{
		ID:          sprint.ID.String(),
		ProjectID:   sprint.ProjectID.String(),
		Name:        sprint.Name,
		Description: *sprint.Description,
		StartDate:   sprint.StartDate.Format("2006-01-02T15:04:05Z07:00"),
		EndDate:     sprint.EndDate.Format("2006-01-02T15:04:05Z07:00"),
		IsCompleted: sprint.IsCompleted,
		IsStarted:   sprint.IsStarted,
		CreatedAt:   sprint.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   sprint.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetSprint handles getting a sprint by ID
func (h *SprintHandler) GetSprint(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	sprintID := chi.URLParam(r, "sprintId")
	if sprintID == "" {
		response.BadRequest(w, "Sprint ID is required")
		return
	}

	sprintUUID, err := uuid.Parse(sprintID)
	if err != nil {
		response.BadRequest(w, "Invalid sprint ID")
		return
	}
	sprint, err := h.queries.GetSprintByID(r.Context(), sprintUUID)
	if err != nil {
		response.NotFound(w, "Sprint not found")
		return
	}

	// Check if user has access to project
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      sprint.ProjectID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to sprint")
		return
	}

	description := ""
	if sprint.Description != nil {
		description = *sprint.Description
	}

	response.JSON(w, http.StatusOK, SprintResponse{
		ID:          sprint.ID.String(),
		ProjectID:   sprint.ProjectID.String(),
		Name:        sprint.Name,
		Description: description,
		StartDate:   sprint.StartDate.Format("2006-01-02T15:04:05Z07:00"),
		EndDate:     sprint.EndDate.Format("2006-01-02T15:04:05Z07:00"),
		IsCompleted: sprint.IsCompleted,
		IsStarted:   sprint.IsStarted,
		CreatedAt:   sprint.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   sprint.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Owner: OwnerInfo{
			ID:        sprint.OwnerID.String(),
			Username:  sprint.OwnerUsername,
			Email:     sprint.OwnerEmail,
			FirstName: sprint.OwnerFirstName,
			LastName:  sprint.OwnerLastName,
		},
	})
}

// UpdateSprint handles sprint updates
func (h *SprintHandler) UpdateSprint(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	sprintID := chi.URLParam(r, "sprintId")
	if sprintID == "" {
		response.BadRequest(w, "Sprint ID is required")
		return
	}

	// Get current sprint to check access
	sprintUUID, err := uuid.Parse(sprintID)
	if err != nil {
		response.BadRequest(w, "Invalid sprint ID")
		return
	}
	currentSprint, err := h.queries.GetSprintByID(r.Context(), sprintUUID)
	if err != nil {
		response.NotFound(w, "Sprint not found")
		return
	}

	// Check if user has access to project
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      currentSprint.ProjectID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to sprint")
		return
	}

	var req UpdateSprintRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Merge updates
	name := currentSprint.Name
	description := *currentSprint.Description
	startDate := currentSprint.StartDate
	endDate := currentSprint.EndDate

	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.StartDate != nil {
		parsedStartDate, err := time.Parse("2006-01-02T15:04:05Z07:00", *req.StartDate)
		if err != nil {
			response.BadRequest(w, "Invalid start date format")
			return
		}
		startDate = parsedStartDate
	}
	if req.EndDate != nil {
		parsedEndDate, err := time.Parse("2006-01-02T15:04:05Z07:00", *req.EndDate)
		if err != nil {
			response.BadRequest(w, "Invalid end date format")
			return
		}
		endDate = parsedEndDate
	}

	sprint, err := h.queries.UpdateSprint(r.Context(), repo.UpdateSprintParams{
		ID:          sprintUUID,
		Name:        name,
		Description: &description,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		response.BadRequest(w, "Failed to update sprint: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, SprintResponse{
		ID:          sprint.ID.String(),
		ProjectID:   sprint.ProjectID.String(),
		Name:        sprint.Name,
		Description: *sprint.Description,
		StartDate:   sprint.StartDate.Format("2006-01-02T15:04:05Z07:00"),
		EndDate:     sprint.EndDate.Format("2006-01-02T15:04:05Z07:00"),
		IsCompleted: sprint.IsCompleted,
		IsStarted:   sprint.IsStarted,
		CreatedAt:   sprint.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   sprint.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteSprint handles sprint deletion
func (h *SprintHandler) DeleteSprint(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	sprintID := chi.URLParam(r, "sprintId")
	if sprintID == "" {
		response.BadRequest(w, "Sprint ID is required")
		return
	}

	// Get current sprint to check access
	sprintUUID, err := uuid.Parse(sprintID)
	if err != nil {
		response.BadRequest(w, "Invalid sprint ID")
		return
	}
	currentSprint, err := h.queries.GetSprintByID(r.Context(), sprintUUID)
	if err != nil {
		response.NotFound(w, "Sprint not found")
		return
	}

	// Check if user has access to project
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      currentSprint.ProjectID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to sprint")
		return
	}

	err = h.queries.DeleteSprint(r.Context(), sprintUUID)
	if err != nil {
		response.InternalServerError(w, "Failed to delete sprint")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Sprint deleted successfully"})
}
