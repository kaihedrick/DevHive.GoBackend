package services

import (
	"context"

	"devhive-backend/models"
	"devhive-backend/repositories"

	"github.com/google/uuid"
)

// TaskService defines the interface for task operations
type TaskService interface {
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

// taskService implements TaskService
type taskService struct {
	taskRepo repositories.TaskRepository
}

// NewTaskService creates a new task service instance
func NewTaskService(taskRepo repositories.TaskRepository) TaskService {
	return &taskService{
		taskRepo: taskRepo,
	}
}

// CreateTask creates a new task
func (s *taskService) CreateTask(ctx context.Context, req models.TaskCreateRequest, projectID uuid.UUID) (*models.Task, error) {
	return s.taskRepo.CreateTask(ctx, req, projectID)
}

// GetTaskByID retrieves a task by its ID
func (s *taskService) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	return s.taskRepo.GetTaskByID(ctx, taskID)
}

// GetTasksBySprint retrieves all tasks for a specific sprint
func (s *taskService) GetTasksBySprint(ctx context.Context, sprintID uuid.UUID) ([]*models.Task, error) {
	return s.taskRepo.GetTasksBySprint(ctx, sprintID)
}

// GetTasksByProject retrieves all tasks for a specific project
func (s *taskService) GetTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*models.Task, error) {
	return s.taskRepo.GetTasksByProject(ctx, projectID)
}

// UpdateTask updates an existing task
func (s *taskService) UpdateTask(ctx context.Context, taskID uuid.UUID, req models.TaskUpdateRequest) (*models.Task, error) {
	return s.taskRepo.UpdateTask(ctx, taskID, req)
}

// DeleteTask deletes a task by ID
func (s *taskService) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	return s.taskRepo.DeleteTask(ctx, taskID)
}

// GetTasksByStatus retrieves tasks by status for a project
func (s *taskService) GetTasksByStatus(ctx context.Context, projectID uuid.UUID, status string) ([]*models.Task, error) {
	return s.taskRepo.GetTasksByStatus(ctx, projectID, status)
}

// GetTasksByAssignee retrieves tasks assigned to a specific user
func (s *taskService) GetTasksByAssignee(ctx context.Context, assigneeID uuid.UUID) ([]*models.Task, error) {
	return s.taskRepo.GetTasksByAssignee(ctx, assigneeID)
}

// CountTasksByProject counts the total number of tasks for a project
func (s *taskService) CountTasksByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	return s.taskRepo.CountTasksByProject(ctx, projectID)
}

// CountTasksByStatus counts tasks by status for a project
func (s *taskService) CountTasksByStatus(ctx context.Context, projectID uuid.UUID, status string) (int64, error) {
	return s.taskRepo.CountTasksByStatus(ctx, projectID, status)
}
