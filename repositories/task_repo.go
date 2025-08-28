package repositories

import (
	"context"

	"devhive-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskRepository defines the interface for task operations
type TaskRepository interface {
	CreateTask(ctx context.Context, req models.TaskCreateRequest, projectID uuid.UUID) (*models.Task, error)
	GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.Task, error)
	GetTasksBySprint(ctx context.Context, sprintID uuid.UUID) ([]*models.Task, error)
	GetTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*models.Task, error)
	UpdateTask(ctx context.Context, taskID uuid.UUID, req models.TaskUpdateRequest) (*models.Task, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	GetTasksByStatus(ctx context.Context, projectID uuid.UUID, status string) ([]*models.Task, error)
	GetTasksByAssignee(ctx context.Context, assigneeID uuid.UUID) ([]*models.Task, error)
	CountTasksByProject(ctx context.Context, projectID uuid.UUID) (int64, error)
	CountTasksByStatus(ctx context.Context, projectID uuid.UUID, status string) (int64, error)
}

// taskRepository implements TaskRepository
type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new task repository instance
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{
		db: db,
	}
}

// CreateTask creates a new task in the database
func (r *taskRepository) CreateTask(ctx context.Context, req models.TaskCreateRequest, projectID uuid.UUID) (*models.Task, error) {
	return models.CreateTask(r.db, req, projectID)
}

// GetTaskByID retrieves a task by its ID
func (r *taskRepository) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	return models.GetTask(r.db, taskID)
}

// GetTasksBySprint retrieves all tasks for a specific sprint
func (r *taskRepository) GetTasksBySprint(ctx context.Context, sprintID uuid.UUID) ([]*models.Task, error) {
	return models.GetTasks(r.db, sprintID)
}

// GetTasksByProject retrieves all tasks for a specific project
func (r *taskRepository) GetTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*models.Task, error) {
	return models.GetTasksByProject(r.db, projectID)
}

// UpdateTask updates an existing task
func (r *taskRepository) UpdateTask(ctx context.Context, taskID uuid.UUID, req models.TaskUpdateRequest) (*models.Task, error) {
	return models.UpdateTask(r.db, taskID, req)
}

// DeleteTask deletes a task by ID
func (r *taskRepository) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	return models.DeleteTask(r.db, taskID)
}

// GetTasksByStatus retrieves tasks by status for a project
func (r *taskRepository) GetTasksByStatus(ctx context.Context, projectID uuid.UUID, status string) ([]*models.Task, error) {
	return models.GetTasksByStatus(r.db, projectID, status)
}

// GetTasksByAssignee retrieves tasks assigned to a specific user
func (r *taskRepository) GetTasksByAssignee(ctx context.Context, assigneeID uuid.UUID) ([]*models.Task, error) {
	return models.GetTasksByAssignee(r.db, assigneeID)
}

// CountTasksByProject counts the total number of tasks for a project
func (r *taskRepository) CountTasksByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	return models.CountTasks(r.db, projectID)
}

// CountTasksByStatus counts tasks by status for a project
func (r *taskRepository) CountTasksByStatus(ctx context.Context, projectID uuid.UUID, status string) (int64, error) {
	return models.CountTasksByStatus(r.db, projectID, status)
}
