package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task represents a task in the system
// @Description Task represents a task in the system
type Task struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title       string     `json:"title" gorm:"not null;size:200" example:"Implement user authentication"`
	Description string     `json:"description" gorm:"size:1000" example:"Create login and registration endpoints"`
	ProjectID   uuid.UUID  `json:"project_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000"`
	SprintID    *uuid.UUID `json:"sprint_id,omitempty" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status      string     `json:"status" gorm:"not null;default:'todo';size:20" example:"todo"`
	Priority    string     `json:"priority" gorm:"not null;default:'medium';size:20" example:"medium"`
	StoryPoints *int       `json:"story_points,omitempty" example:"5"`
	DueDate     *time.Time `json:"due_date,omitempty" example:"2024-01-15T00:00:00Z"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	// Additional fields for API responses
	Project  *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Sprint   *Sprint  `json:"sprint,omitempty" gorm:"foreignKey:SprintID"`
	Assignee *User    `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
}

// TableName specifies the table name for the Task model
func (Task) TableName() string {
	return "tasks"
}

// TaskCreateRequest represents the request to create a new task
// @Description Request to create a new task
type TaskCreateRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=200" example:"Implement user authentication"`
	Description string     `json:"description" binding:"max=1000" example:"Create login and registration endpoints"`
	SprintID    *uuid.UUID `json:"sprint_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status      string     `json:"status" binding:"omitempty,oneof=todo in_progress review done" example:"todo"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high urgent" example:"medium"`
	StoryPoints *int       `json:"story_points,omitempty" binding:"omitempty,min=1,max=21" example:"5"`
	DueDate     *time.Time `json:"due_date,omitempty" example:"2024-01-15T00:00:00Z"`
}

// TaskUpdateRequest represents the request to update an existing task
// @Description Request to update an existing task
type TaskUpdateRequest struct {
	Title       *string    `json:"title,omitempty" binding:"omitempty,min=1,max=200" example:"Updated task title"`
	Description *string    `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated task description"`
	SprintID    *uuid.UUID `json:"sprint_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status      *string    `json:"status,omitempty" binding:"omitempty,oneof=todo in_progress review done" example:"in_progress"`
	Priority    *string    `json:"priority,omitempty" binding:"omitempty,oneof=low medium high urgent" example:"high"`
	StoryPoints *int       `json:"story_points,omitempty" binding:"omitempty,min=1,max=21" example:"8"`
	DueDate     *time.Time `json:"due_date,omitempty" example:"2024-01-20T00:00:00Z"`
}

// CreateTask creates a new task in the database using GORM
func CreateTask(db *gorm.DB, req TaskCreateRequest, projectID uuid.UUID) (*Task, error) {
	task := &Task{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		ProjectID:   projectID,
		SprintID:    req.SprintID,
		AssigneeID:  req.AssigneeID,
		Status:      req.Status,
		Priority:    req.Priority,
		StoryPoints: req.StoryPoints,
		DueDate:     req.DueDate,
	}

	// Set defaults if not provided
	if task.Status == "" {
		task.Status = "todo"
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}

	if err := db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask retrieves a task by ID using GORM
func GetTask(db *gorm.DB, taskID uuid.UUID) (*Task, error) {
	var task Task
	if err := db.Preload("Project").Preload("Sprint").Preload("Assignee").
		Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTasks retrieves all tasks for a specific project using GORM
func GetTasks(db *gorm.DB, projectID uuid.UUID) ([]*Task, error) {
	var tasks []*Task
	if err := db.Where("project_id = ?", projectID).
		Preload("Sprint").Preload("Assignee").
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTasksByProject retrieves all tasks for a specific project using GORM
func GetTasksByProject(db *gorm.DB, projectID uuid.UUID) ([]*Task, error) {
	var tasks []*Task
	if err := db.Where("project_id = ?", projectID).
		Preload("Sprint").Preload("Assignee").
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTasksBySprint retrieves all tasks for a specific sprint using GORM
func GetTasksBySprint(db *gorm.DB, sprintID uuid.UUID) ([]*Task, error) {
	var tasks []*Task
	if err := db.Where("sprint_id = ?", sprintID).
		Preload("Assignee").
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateTask updates an existing task using GORM
func UpdateTask(db *gorm.DB, taskID uuid.UUID, req TaskUpdateRequest) (*Task, error) {
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.SprintID != nil {
		updates["sprint_id"] = *req.SprintID
	}
	if req.AssigneeID != nil {
		updates["assignee_id"] = *req.AssigneeID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.StoryPoints != nil {
		updates["story_points"] = *req.StoryPoints
	}
	if req.DueDate != nil {
		updates["due_date"] = *req.DueDate
	}

	if err := db.Model(&Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, err
	}

	return GetTask(db, taskID)
}

// DeleteTask deletes a task by ID using GORM
func DeleteTask(db *gorm.DB, taskID uuid.UUID) error {
	return db.Where("id = ?", taskID).Delete(&Task{}).Error
}

// GetTasksByStatus retrieves tasks by status for a project using GORM
func GetTasksByStatus(db *gorm.DB, projectID uuid.UUID, status string) ([]*Task, error) {
	var tasks []*Task
	if err := db.Where("project_id = ? AND status = ?", projectID, status).
		Preload("Sprint").Preload("Assignee").
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTasksByAssignee retrieves tasks assigned to a specific user using GORM
func GetTasksByAssignee(db *gorm.DB, assigneeID uuid.UUID) ([]*Task, error) {
	var tasks []*Task
	if err := db.Where("assignee_id = ?", assigneeID).
		Preload("Project").Preload("Sprint").
		Order("priority DESC, due_date ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// CountTasks counts the total number of tasks for a project
func CountTasks(db *gorm.DB, projectID uuid.UUID) (int64, error) {
	var count int64
	if err := db.Model(&Task{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountTasksByStatus counts tasks by status for a project
func CountTasksByStatus(db *gorm.DB, projectID uuid.UUID, status string) (int64, error) {
	var count int64
	if err := db.Model(&Task{}).Where("project_id = ? AND status = ?", projectID, status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// BeforeCreate is a GORM hook that runs before creating a task
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a task
func (t *Task) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}

// AssignTask assigns a task to a specific user
func AssignTask(db *gorm.DB, taskID, assigneeID uuid.UUID) (*Task, error) {
	if err := db.Model(&Task{}).Where("id = ?", taskID).Update("assignee_id", assigneeID).Error; err != nil {
		return nil, err
	}
	return GetTask(db, taskID)
}

// UpdateTaskStatus updates the status of a task
func UpdateTaskStatus(db *gorm.DB, taskID uuid.UUID, status string) (*Task, error) {
	if err := db.Model(&Task{}).Where("id = ?", taskID).Update("status", status).Error; err != nil {
		return nil, err
	}
	return GetTask(db, taskID)
}
