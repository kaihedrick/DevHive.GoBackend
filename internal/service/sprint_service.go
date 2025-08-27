package service

import (
	"context"
	"database/sql"

	"devhive-backend/internal/ws"
	"devhive-backend/models"

	"github.com/google/uuid"
)

// SprintService defines the interface for sprint operations
type SprintService interface {
	GetSprintsForProject(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error)
	GetSprintByID(ctx context.Context, sprintID uuid.UUID) (*models.Sprint, error)
	CreateSprint(ctx context.Context, req models.SprintCreateRequest, projectID uuid.UUID, userID uuid.UUID) (*models.Sprint, error)
	UpdateSprint(ctx context.Context, sprintID uuid.UUID, req models.SprintUpdateRequest, userID uuid.UUID) (*models.Sprint, error)
	DeleteSprint(ctx context.Context, sprintID uuid.UUID, userID uuid.UUID) error
	StartSprint(ctx context.Context, sprintID uuid.UUID, userID uuid.UUID) error
	CompleteSprint(ctx context.Context, sprintID uuid.UUID, userID uuid.UUID) error
	GetActiveSprint(ctx context.Context, projectID uuid.UUID) (*models.Sprint, error)
	GetUpcomingSprints(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error)
}

// sprintService implements SprintService
type sprintService struct {
	db *sql.DB
}

// NewSprintService creates a new sprint service instance
func NewSprintService(db *sql.DB) SprintService {
	return &sprintService{
		db: db,
	}
}

// GetSprintsForProject retrieves all sprints for a specific project
func (s *sprintService) GetSprintsForProject(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error) {
	return models.GetSprints(s.db, projectID)
}

// GetSprintByID retrieves a sprint by its ID
func (s *sprintService) GetSprintByID(ctx context.Context, sprintID uuid.UUID) (*models.Sprint, error) {
	return models.GetSprint(s.db, sprintID)
}

// CreateSprint creates a new sprint and broadcasts the update
func (s *sprintService) CreateSprint(ctx context.Context, req models.SprintCreateRequest, projectID uuid.UUID, userID uuid.UUID) (*models.Sprint, error) {
	// Check if user has permission to create sprints in this project
	isMember, err := models.IsProjectMember(s.db, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, models.ErrAccessDenied
	}

	sprint, err := models.CreateSprint(s.db, req, projectID)
	if err != nil {
		return nil, err
	}

	// Broadcast the new sprint to all project members
	ws.BroadcastSprintUpdate(projectID.String(), map[string]interface{}{
		"action": "created",
		"sprint": sprint,
	})

	return sprint, nil
}

// UpdateSprint updates an existing sprint and broadcasts the update
func (s *sprintService) UpdateSprint(ctx context.Context, sprintID uuid.UUID, req models.SprintUpdateRequest, userID uuid.UUID) (*models.Sprint, error) {
	// Get the sprint to check permissions
	sprint, err := models.GetSprint(s.db, sprintID)
	if err != nil {
		return nil, err
	}

	// Check if user has permission to update this sprint
	isMember, err := models.IsProjectMember(s.db, sprint.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, models.ErrAccessDenied
	}

	updatedSprint, err := models.UpdateSprint(s.db, sprintID, req)
	if err != nil {
		return nil, err
	}

	// Broadcast the sprint update to all project members
	ws.BroadcastSprintUpdate(sprint.ProjectID.String(), map[string]interface{}{
		"action": "updated",
		"sprint": updatedSprint,
	})

	return updatedSprint, nil
}

// DeleteSprint deletes a sprint and broadcasts the update
func (s *sprintService) DeleteSprint(ctx context.Context, sprintID uuid.UUID, userID uuid.UUID) error {
	// Get the sprint to check permissions and get project ID
	sprint, err := models.GetSprint(s.db, sprintID)
	if err != nil {
		return err
	}

	// Check if user has permission to delete this sprint
	userRole, err := models.GetProjectMemberRole(s.db, sprint.ProjectID, userID)
	if err != nil {
		return err
	}
	if userRole != "admin" && userRole != "owner" {
		return models.ErrAccessDenied
	}

	err = models.DeleteSprint(s.db, sprintID)
	if err != nil {
		return err
	}

	// Broadcast the sprint deletion to all project members
	ws.BroadcastSprintUpdate(sprint.ProjectID.String(), map[string]interface{}{
		"action": "deleted",
		"sprint_id": sprintID.String(),
	})

	return nil
}

// StartSprint starts a sprint and broadcasts the update
func (s *sprintService) StartSprint(ctx context.Context, sprintID uuid.UUID, userID uuid.UUID) error {
	// Get the sprint to check permissions
	sprint, err := models.GetSprint(s.db, sprintID)
	if err != nil {
		return err
	}

	// Check if user has permission to start this sprint
	isMember, err := models.IsProjectMember(s.db, sprint.ProjectID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return models.ErrAccessDenied
	}

	// Check if sprint is in planned status
	if sprint.Status != "planned" {
		return models.ErrInvalidSprintStatus
	}

	// Update sprint status to active
	req := models.SprintUpdateRequest{
		Status: &[]string{"active"}[0],
	}

	updatedSprint, err := models.UpdateSprint(s.db, sprintID, req)
	if err != nil {
		return err
	}

	// Broadcast the sprint start to all project members
	ws.BroadcastSprintUpdate(sprint.ProjectID.String(), map[string]interface{}{
		"action": "started",
		"sprint": updatedSprint,
	})

	return nil
}

// CompleteSprint completes a sprint and broadcasts the update
func (s *sprintService) CompleteSprint(ctx context.Context, sprintID uuid.UUID, userID uuid.UUID) error {
	// Get the sprint to check permissions
	sprint, err := models.GetSprint(s.db, sprintID)
	if err != nil {
		return err
	}

	// Check if user has permission to complete this sprint
	isMember, err := models.IsProjectMember(s.db, sprint.ProjectID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return models.ErrAccessDenied
	}

	// Check if sprint is in active status
	if sprint.Status != "active" {
		return models.ErrInvalidSprintStatus
	}

	// Update sprint status to completed
	req := models.SprintUpdateRequest{
		Status: &[]string{"completed"}[0],
	}

	updatedSprint, err := models.UpdateSprint(s.db, sprintID, req)
	if err != nil {
		return err
	}

	// Broadcast the sprint completion to all project members
	ws.BroadcastSprintUpdate(sprint.ProjectID.String(), map[string]interface{}{
		"action": "completed",
		"sprint": updatedSprint,
	})

	return nil
}

// GetActiveSprint retrieves the currently active sprint for a project
func (s *sprintService) GetActiveSprint(ctx context.Context, projectID uuid.UUID) (*models.Sprint, error) {
	return models.GetActiveSprint(s.db, projectID)
}

// GetUpcomingSprints retrieves upcoming sprints for a project
func (s *sprintService) GetUpcomingSprints(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error) {
	return models.GetUpcomingSprints(s.db, projectID)
}
