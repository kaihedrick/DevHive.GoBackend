package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Sprint represents a sprint in the system
type Sprint struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ProjectID   uuid.UUID `json:"project_id" db:"project_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	StartDate   time.Time `json:"start_date" db:"start_date"`
	EndDate     time.Time `json:"end_date" db:"end_date"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// SprintCreateRequest represents the request to create a new sprint
type SprintCreateRequest struct {
	Name        string    `json:"name" binding:"required,min=1,max=255"`
	Description *string   `json:"description,omitempty" binding:"omitempty,max=1000"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
}

// SprintUpdateRequest represents the request to update a sprint
type SprintUpdateRequest struct {
	Name        *string    `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string    `json:"description,omitempty" binding:"omitempty,max=1000"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Status      *string    `json:"status,omitempty" binding:"omitempty,oneof=planned active completed cancelled"`
}

// CreateSprint creates a new sprint in the database
func CreateSprint(db *sql.DB, req SprintCreateRequest, projectID uuid.UUID) (*Sprint, error) {
	sprint := &Sprint{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      "planned",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO sprints (id, project_id, name, description, start_date, end_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, project_id, name, description, start_date, end_date, status, created_at, updated_at
	`

	err := db.QueryRow(
		query,
		sprint.ID, sprint.ProjectID, sprint.Name, sprint.Description,
		sprint.StartDate, sprint.EndDate, sprint.Status, sprint.CreatedAt, sprint.UpdatedAt,
	).Scan(
		&sprint.ID, &sprint.ProjectID, &sprint.Name, &sprint.Description,
		&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return sprint, nil
}

// GetSprint retrieves a sprint by ID
func GetSprint(db *sql.DB, sprintID uuid.UUID) (*Sprint, error) {
	sprint := &Sprint{}
	query := `
		SELECT id, project_id, name, description, start_date, end_date, status, created_at, updated_at
		FROM sprints WHERE id = $1
	`

	err := db.QueryRow(query, sprintID).Scan(
		&sprint.ID, &sprint.ProjectID, &sprint.Name, &sprint.Description,
		&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return sprint, nil
}

// GetSprints retrieves all sprints for a project
func GetSprints(db *sql.DB, projectID uuid.UUID) ([]*Sprint, error) {
	query := `
		SELECT id, project_id, name, description, start_date, end_date, status, created_at, updated_at
		FROM sprints 
		WHERE project_id = $1
		ORDER BY start_date ASC
	`

	rows, err := db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		sprint := &Sprint{}
		err := rows.Scan(
			&sprint.ID, &sprint.ProjectID, &sprint.Name, &sprint.Description,
			&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		sprints = append(sprints, sprint)
	}

	return sprints, nil
}

// UpdateSprint updates a sprint in the database
func UpdateSprint(db *sql.DB, sprintID uuid.UUID, req SprintUpdateRequest) (*Sprint, error) {
	sprint, err := GetSprint(db, sprintID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		sprint.Name = *req.Name
	}
	if req.Description != nil {
		sprint.Description = req.Description
	}
	if req.StartDate != nil {
		sprint.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		sprint.EndDate = *req.EndDate
	}
	if req.Status != nil {
		sprint.Status = *req.Status
	}

	sprint.UpdatedAt = time.Now()

	query := `
		UPDATE sprints 
		SET name = $1, description = $2, start_date = $3, end_date = $4, status = $5, updated_at = $6
		WHERE id = $7
		RETURNING id, project_id, name, description, start_date, end_date, status, created_at, updated_at
	`

	err = db.QueryRow(
		query,
		sprint.Name, sprint.Description, sprint.StartDate, sprint.EndDate,
		sprint.Status, sprint.UpdatedAt, sprint.ID,
	).Scan(
		&sprint.ID, &sprint.ProjectID, &sprint.Name, &sprint.Description,
		&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return sprint, nil
}

// DeleteSprint deletes a sprint from the database
func DeleteSprint(db *sql.DB, sprintID uuid.UUID) error {
	query := `DELETE FROM sprints WHERE id = $1`
	_, err := db.Exec(query, sprintID)
	return err
}

// GetActiveSprint retrieves the currently active sprint for a project
func GetActiveSprint(db *sql.DB, projectID uuid.UUID) (*Sprint, error) {
	sprint := &Sprint{}
	query := `
		SELECT id, project_id, name, description, start_date, end_date, status, created_at, updated_at
		FROM sprints 
		WHERE project_id = $1 AND status = 'active'
		LIMIT 1
	`

	err := db.QueryRow(query, projectID).Scan(
		&sprint.ID, &sprint.ProjectID, &sprint.Name, &sprint.Description,
		&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active sprint
		}
		return nil, err
	}

	return sprint, nil
}

// GetUpcomingSprints retrieves upcoming sprints for a project
func GetUpcomingSprints(db *sql.DB, projectID uuid.UUID) ([]*Sprint, error) {
	query := `
		SELECT id, project_id, name, description, start_date, end_date, status, created_at, updated_at
		FROM sprints 
		WHERE project_id = $1 AND status = 'planned' AND start_date > $2
		ORDER BY start_date ASC
	`

	rows, err := db.Query(query, projectID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		sprint := &Sprint{}
		err := rows.Scan(
			&sprint.ID, &sprint.ProjectID, &sprint.Name, &sprint.Description,
			&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		sprints = append(sprints, sprint)
	}

	return sprints, nil
}
