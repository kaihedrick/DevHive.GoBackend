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

type TaskHandler struct {
	queries *repo.Queries
}

func NewTaskHandler(queries *repo.Queries) *TaskHandler {
	return &TaskHandler{
		queries: queries,
	}
}

// CreateTaskRequest represents the task creation request
type CreateTaskRequest struct {
	Description string `json:"description"`
	SprintID    string `json:"sprintId,omitempty"`
	AssigneeID  string `json:"assigneeId,omitempty"`
	Status      int32  `json:"status"`
}

// UpdateTaskRequest represents the task update request
type UpdateTaskRequest struct {
	Description *string `json:"description,omitempty"`
	AssigneeID  *string `json:"assigneeId,omitempty"`
}

// UpdateTaskStatusRequest represents the task status update request
type UpdateTaskStatusRequest struct {
	Status int32 `json:"status"`
}

// TaskResponse represents a task response
type TaskResponse struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	SprintID    string `json:"sprintId,omitempty"`
	AssigneeID  string `json:"assigneeId,omitempty"`
	Description string `json:"description"`
	Status      int32  `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Assignee    *struct {
		Username  string `json:"username"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"assignee,omitempty"`
	Owner OwnerInfo `json:"owner"`
}

// ListTasksByProject handles listing tasks for a project
func (h *TaskHandler) ListTasksByProject(w http.ResponseWriter, r *http.Request) {
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

	tasks, err := h.queries.ListTasksByProject(r.Context(), repo.ListTasksByProjectParams{
		ProjectID: projectUUID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		response.InternalServerError(w, "Failed to list tasks")
		return
	}

	// Convert to response format
	var taskResponses []TaskResponse
	for _, task := range tasks {
		description := ""
		if task.Description != nil {
			description = *task.Description
		}

		taskResp := TaskResponse{
			ID:          task.ID.String(),
			ProjectID:   task.ProjectID.String(),
			Description: description,
			Status:      task.Status,
			CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Owner: OwnerInfo{
				ID:        task.OwnerID.String(),
				Username:  task.OwnerUsername,
				Email:     task.OwnerEmail,
				FirstName: task.OwnerFirstName,
				LastName:  task.OwnerLastName,
			},
		}

		if task.SprintID.Valid {
			sprintUUID := uuid.UUID(task.SprintID.Bytes)
			taskResp.SprintID = sprintUUID.String()
		}
		if task.AssigneeID.Valid {
			assigneeUUID := uuid.UUID(task.AssigneeID.Bytes)
			taskResp.AssigneeID = assigneeUUID.String()
			assigneeUsername := ""
			if task.AssigneeUsername != nil {
				assigneeUsername = *task.AssigneeUsername
			}
			assigneeFirstName := ""
			if task.AssigneeFirstName != nil {
				assigneeFirstName = *task.AssigneeFirstName
			}
			assigneeLastName := ""
			if task.AssigneeLastName != nil {
				assigneeLastName = *task.AssigneeLastName
			}
			taskResp.Assignee = &struct {
				Username  string `json:"username"`
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			}{
				Username:  assigneeUsername,
				FirstName: assigneeFirstName,
				LastName:  assigneeLastName,
			}
		}

		taskResponses = append(taskResponses, taskResp)
	}

	// Ensure tasks is always an array, never null
	if taskResponses == nil {
		taskResponses = []TaskResponse{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"tasks":  taskResponses,
		"limit":  limit,
		"offset": offset,
	})
}

// ListTasksBySprint handles listing tasks for a sprint
func (h *TaskHandler) ListTasksBySprint(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	sprintID := chi.URLParam(r, "sprintId")
	if sprintID == "" {
		response.BadRequest(w, "Sprint ID is required")
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

	sprintUUID, err := uuid.Parse(sprintID)
	if err != nil {
		response.BadRequest(w, "Invalid sprint ID")
		return
	}

	tasks, err := h.queries.ListTasksBySprint(r.Context(), repo.ListTasksBySprintParams{
		SprintID: pgtype.UUID{Bytes: sprintUUID, Valid: true},
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		response.InternalServerError(w, "Failed to list tasks")
		return
	}

	// Convert to response format
	var taskResponses []TaskResponse
	for _, task := range tasks {
		description := ""
		if task.Description != nil {
			description = *task.Description
		}

		taskResp := TaskResponse{
			ID:          task.ID.String(),
			ProjectID:   task.ProjectID.String(),
			Description: description,
			Status:      task.Status,
			CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Owner: OwnerInfo{
				ID:        task.OwnerID.String(),
				Username:  task.OwnerUsername,
				Email:     task.OwnerEmail,
				FirstName: task.OwnerFirstName,
				LastName:  task.OwnerLastName,
			},
		}

		if task.SprintID.Valid {
			sprintUUID := uuid.UUID(task.SprintID.Bytes)
			taskResp.SprintID = sprintUUID.String()
		}
		if task.AssigneeID.Valid {
			assigneeUUID := uuid.UUID(task.AssigneeID.Bytes)
			taskResp.AssigneeID = assigneeUUID.String()
			assigneeUsername := ""
			if task.AssigneeUsername != nil {
				assigneeUsername = *task.AssigneeUsername
			}
			assigneeFirstName := ""
			if task.AssigneeFirstName != nil {
				assigneeFirstName = *task.AssigneeFirstName
			}
			assigneeLastName := ""
			if task.AssigneeLastName != nil {
				assigneeLastName = *task.AssigneeLastName
			}
			taskResp.Assignee = &struct {
				Username  string `json:"username"`
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			}{
				Username:  assigneeUsername,
				FirstName: assigneeFirstName,
				LastName:  assigneeLastName,
			}
		}

		taskResponses = append(taskResponses, taskResp)
	}

	// Ensure tasks is always an array, never null
	if taskResponses == nil {
		taskResponses = []TaskResponse{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"tasks":  taskResponses,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateTask handles task creation
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
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

	var req CreateTaskRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Set default status if not provided
	if req.Status == 0 {
		req.Status = 0 // Default status
	}

	// Parse optional fields
	var sprintUUID pgtype.UUID
	var assigneeUUID pgtype.UUID

	if req.SprintID != "" {
		sprintID, err := uuid.Parse(req.SprintID)
		if err != nil {
			response.BadRequest(w, "Invalid sprint ID format")
			return
		}
		sprintUUID = pgtype.UUID{Bytes: sprintID, Valid: true}
	}

	if req.AssigneeID != "" {
		assigneeID, err := uuid.Parse(req.AssigneeID)
		if err != nil {
			response.BadRequest(w, "Invalid assignee ID format")
			return
		}
		assigneeUUID = pgtype.UUID{Bytes: assigneeID, Valid: true}
	}

	task, err := h.queries.CreateTask(r.Context(), repo.CreateTaskParams{
		ProjectID:   projectUUID,
		SprintID:    sprintUUID,
		AssigneeID:  assigneeUUID,
		Description: &req.Description,
		Status:      req.Status,
	})
	if err != nil {
		response.BadRequest(w, "Failed to create task: "+err.Error())
		return
	}

	// Get project owner information
	project, err := h.queries.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		response.InternalServerError(w, "Failed to get project owner")
		return
	}

	// Get owner details
	owner, err := h.queries.GetUserByID(r.Context(), project.OwnerID)
	if err != nil {
		response.InternalServerError(w, "Failed to get owner details")
		return
	}

	description := ""
	if task.Description != nil {
		description = *task.Description
	}

	taskResp := TaskResponse{
		ID:          task.ID.String(),
		ProjectID:   task.ProjectID.String(),
		Description: description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Owner: OwnerInfo{
			ID:        owner.ID.String(),
			Username:  owner.Username,
			Email:     owner.Email,
			FirstName: owner.FirstName,
			LastName:  owner.LastName,
		},
	}

	// Add sprint ID if assigned
	if task.SprintID.Valid {
		sprintUUID := uuid.UUID(task.SprintID.Bytes)
		taskResp.SprintID = sprintUUID.String()
	}

	// Add assignee ID if assigned
	if task.AssigneeID.Valid {
		assigneeUUID := uuid.UUID(task.AssigneeID.Bytes)
		taskResp.AssigneeID = assigneeUUID.String()
	}

	response.JSON(w, http.StatusCreated, taskResp)
}

// GetTask handles getting a task by ID
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	taskID := chi.URLParam(r, "taskId")
	if taskID == "" {
		response.BadRequest(w, "Task ID is required")
		return
	}

	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		response.BadRequest(w, "Invalid task ID")
		return
	}
	task, err := h.queries.GetTaskByID(r.Context(), taskUUID)
	if err != nil {
		response.NotFound(w, "Task not found")
		return
	}

	// Check if user has access to project
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      task.ProjectID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to task")
		return
	}

	description := ""
	if task.Description != nil {
		description = *task.Description
	}

	taskResp := TaskResponse{
		ID:          task.ID.String(),
		ProjectID:   task.ProjectID.String(),
		Description: description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Owner: OwnerInfo{
			ID:        task.OwnerID.String(),
			Username:  task.OwnerUsername,
			Email:     task.OwnerEmail,
			FirstName: task.OwnerFirstName,
			LastName:  task.OwnerLastName,
		},
	}

	if task.SprintID.Valid {
		sprintUUID := uuid.UUID(task.SprintID.Bytes)
		taskResp.SprintID = sprintUUID.String()
	}
	if task.AssigneeID.Valid {
		assigneeUUID := uuid.UUID(task.AssigneeID.Bytes)
		taskResp.AssigneeID = assigneeUUID.String()
		assigneeUsername := ""
		if task.AssigneeUsername != nil {
			assigneeUsername = *task.AssigneeUsername
		}
		assigneeFirstName := ""
		if task.AssigneeFirstName != nil {
			assigneeFirstName = *task.AssigneeFirstName
		}
		assigneeLastName := ""
		if task.AssigneeLastName != nil {
			assigneeLastName = *task.AssigneeLastName
		}
		taskResp.Assignee = &struct {
			Username  string `json:"username"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}{
			Username:  assigneeUsername,
			FirstName: assigneeFirstName,
			LastName:  assigneeLastName,
		}
	}

	response.JSON(w, http.StatusOK, taskResp)
}

// UpdateTask handles task updates
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	taskID := chi.URLParam(r, "taskId")
	if taskID == "" {
		response.BadRequest(w, "Task ID is required")
		return
	}

	// Get current task to check access
	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		response.BadRequest(w, "Invalid task ID")
		return
	}
	currentTask, err := h.queries.GetTaskByID(r.Context(), taskUUID)
	if err != nil {
		response.NotFound(w, "Task not found")
		return
	}

	// Check if user has access to project
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      currentTask.ProjectID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to task")
		return
	}

	var req UpdateTaskRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Merge updates
	description := *currentTask.Description
	assigneeID := currentTask.AssigneeID

	if req.Description != nil {
		description = *req.Description
	}
	if req.AssigneeID != nil {
		// TODO: Validate assignee is member of project
		// For now, we'll skip assignee assignment
		_ = req.AssigneeID // Suppress unused variable warning
	}

	task, err := h.queries.UpdateTask(r.Context(), repo.UpdateTaskParams{
		ID:          taskUUID,
		Description: &description,
		AssigneeID:  assigneeID,
	})
	if err != nil {
		response.BadRequest(w, "Failed to update task: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, TaskResponse{
		ID:          task.ID.String(),
		ProjectID:   task.ProjectID.String(),
		Description: *task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateTaskStatus handles task status updates
func (h *TaskHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	taskID := chi.URLParam(r, "taskId")
	if taskID == "" {
		response.BadRequest(w, "Task ID is required")
		return
	}

	// Get current task to check access
	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		response.BadRequest(w, "Invalid task ID")
		return
	}
	currentTask, err := h.queries.GetTaskByID(r.Context(), taskUUID)
	if err != nil {
		response.NotFound(w, "Task not found")
		return
	}

	// Check if user has access to project
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      currentTask.ProjectID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to task")
		return
	}

	var req UpdateTaskStatusRequest
	if !response.Decode(w, r, &req) {
		return
	}

	task, err := h.queries.UpdateTaskStatus(r.Context(), repo.UpdateTaskStatusParams{
		ID:     taskUUID,
		Status: req.Status,
	})
	if err != nil {
		response.BadRequest(w, "Failed to update task status: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, TaskResponse{
		ID:          task.ID.String(),
		ProjectID:   task.ProjectID.String(),
		Description: *task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteTask handles task deletion
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	taskID := chi.URLParam(r, "taskId")
	if taskID == "" {
		response.BadRequest(w, "Task ID is required")
		return
	}

	// Get current task to check access
	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		response.BadRequest(w, "Invalid task ID")
		return
	}
	currentTask, err := h.queries.GetTaskByID(r.Context(), taskUUID)
	if err != nil {
		response.NotFound(w, "Task not found")
		return
	}

	// Check if user has access to project
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	hasAccess, err := h.queries.CheckProjectAccess(r.Context(), repo.CheckProjectAccessParams{
		ID:      currentTask.ProjectID,
		OwnerID: userUUID,
	})
	if err != nil || !hasAccess {
		response.Forbidden(w, "Access denied to task")
		return
	}

	err = h.queries.DeleteTask(r.Context(), taskUUID)
	if err != nil {
		response.InternalServerError(w, "Failed to delete task")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
}
