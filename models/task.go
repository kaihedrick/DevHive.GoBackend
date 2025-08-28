package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Task represents a task in the system
// @Description Task represents a task in the system
type Task struct {
	ID          uuid.UUID  `json:"id" db:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Description string     `json:"description" db:"description" example:"Implement user authentication"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty" db:"assignee_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	DateCreated time.Time  `json:"date_created" db:"date_created" example:"2025-08-28T14:00:00Z"`
	Status      int        `json:"status" db:"status" example:"1"`
	SprintID    uuid.UUID  `json:"sprint_id" db:"sprint_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Additional fields for API responses
	Assignee *User   `json:"assignee,omitempty"`
	Sprint   *Sprint `json:"sprint,omitempty"`
}

// TaskCreateRequest represents the request to create a new task
// @Description Request to create a new task
type TaskCreateRequest struct {
	Description string     `json:"description" binding:"required,min=1,max=255" example:"Implement user authentication"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status      int        `json:"status" binding:"required,min=0,max=3" example:"1"`
}

// TaskUpdateRequest represents the request to update a task
// @Description Request to update an existing task
type TaskUpdateRequest struct {
	Description *string    `json:"description,omitempty" binding:"omitempty,min=1,max=255" example:"Implement user authentication with OAuth2"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status      *int       `json:"status,omitempty" binding:"omitempty,min=0,max=3" example:"2"`
}

// CreateTask creates a new task in the database
func CreateTask(db *sql.DB, req TaskCreateRequest, sprintID uuid.UUID) (*Task, error) {
	task := &Task{
		ID:          uuid.New(),
		Description: req.Description,
		AssigneeID:  req.AssigneeID,
		DateCreated: time.Now(),
		Status:      req.Status,
		SprintID:    sprintID,
	}

	query := `
		INSERT INTO tasks (id, description, assignee_id, date_created, status, sprint_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, description, assignee_id, date_created, status, sprint_id
	`

	err := db.QueryRow(
		query,
		task.ID, task.Description, task.AssigneeID, task.DateCreated, task.Status, task.SprintID,
	).Scan(
		&task.ID, &task.Description, &task.AssigneeID, &task.DateCreated, &task.Status, &task.SprintID,
	)

	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask retrieves a task by ID
func GetTask(db *sql.DB, taskID uuid.UUID) (*Task, error) {
	task := &Task{}
	query := `
		SELECT id, description, assignee_id, date_created, status, sprint_id
		FROM tasks WHERE id = $1
	`

	err := db.QueryRow(query, taskID).Scan(
		&task.ID, &task.Description, &task.AssigneeID, &task.DateCreated, &task.Status, &task.SprintID,
	)

	if err != nil {
		return nil, err
	}

	// Get assignee information if assigned
	if task.AssigneeID != nil {
		assignee, err := GetUserByID(db, *task.AssigneeID)
		if err != nil {
			return nil, err
		}
		task.Assignee = assignee
	}

	return task, nil
}

// GetTasksBySprint retrieves all tasks for a sprint
func GetTasksBySprint(db *sql.DB, sprintID uuid.UUID) ([]*Task, error) {
	query := `
		SELECT id, description, assignee_id, date_created, status, sprint_id
		FROM tasks WHERE sprint_id = $1
		ORDER BY date_created ASC
	`

	rows, err := db.Query(query, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(
			&task.ID, &task.Description, &task.AssigneeID, &task.DateCreated, &task.Status, &task.SprintID,
		)
		if err != nil {
			return nil, err
		}

		// Get assignee information if assigned
		if task.AssigneeID != nil {
			assignee, err := GetUserByID(db, *task.AssigneeID)
			if err != nil {
				return nil, err
			}
			task.Assignee = assignee
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask updates a task in the database
func UpdateTask(db *sql.DB, taskID uuid.UUID, req TaskUpdateRequest) (*Task, error) {
	task, err := GetTask(db, taskID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	if req.Status != nil {
		task.Status = *req.Status
	}

	query := `
		UPDATE tasks 
		SET description = $1, assignee_id = $2, status = $3
		WHERE id = $4
		RETURNING id, description, assignee_id, date_created, status, sprint_id
	`

	err = db.QueryRow(
		query,
		task.Description, task.AssigneeID, task.Status, task.ID,
	).Scan(
		&task.ID, &task.Description, &task.AssigneeID, &task.DateCreated, &task.Status, &task.SprintID,
	)

	if err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTask deletes a task from the database
func DeleteTask(db *sql.DB, taskID uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := db.Exec(query, taskID)
	return err
}
