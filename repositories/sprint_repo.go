package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"devhive-backend/models"

	"github.com/google/uuid"
)

// SprintRepository defines the interface for sprint data operations
type SprintRepository interface {
	Create(ctx context.Context, sprint *models.Sprint) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Sprint, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error)
	Update(ctx context.Context, sprint *models.Sprint) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetActiveSprint(ctx context.Context, projectID uuid.UUID) (*models.Sprint, error)
	GetUpcomingSprints(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error)
}

// sprintRepository implements SprintRepository
type sprintRepository struct {
	db *sql.DB
}

// NewSprintRepository creates a new sprint repository instance
func NewSprintRepository(db *sql.DB) SprintRepository {
	return &sprintRepository{db: db}
}

// Create creates a new sprint
func (r *sprintRepository) Create(ctx context.Context, sprint *models.Sprint) error {
	query := `
		INSERT INTO sprints (id, name, description, project_id, start_date, end_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		sprint.ID, sprint.Name, sprint.Description, sprint.ProjectID,
		sprint.StartDate, sprint.EndDate, sprint.Status, sprint.CreatedAt, sprint.UpdatedAt)
	
	return err
}

// GetByID retrieves a sprint by ID
func (r *sprintRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Sprint, error) {
	query := `
		SELECT id, name, description, project_id, start_date, end_date, status, created_at, updated_at
		FROM sprints WHERE id = $1
	`
	
	sprint := &models.Sprint{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sprint.ID, &sprint.Name, &sprint.Description, &sprint.ProjectID,
		&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sprint not found: %w", err)
		}
		return nil, fmt.Errorf("error getting sprint: %w", err)
	}
	
	return sprint, nil
}

// GetByProjectID retrieves sprints by project ID
func (r *sprintRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error) {
	query := `
		SELECT id, name, description, project_id, start_date, end_date, status, created_at, updated_at
		FROM sprints WHERE project_id = $1 ORDER BY start_date DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting sprints by project: %w", err)
	}
	defer rows.Close()
	
	var sprints []*models.Sprint
	for rows.Next() {
		sprint := &models.Sprint{}
		err := rows.Scan(
			&sprint.ID, &sprint.Name, &sprint.Description, &sprint.ProjectID,
			&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning sprint: %w", err)
		}
		
		sprints = append(sprints, sprint)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sprints: %w", err)
	}
	
	return sprints, nil
}

// Update updates an existing sprint
func (r *sprintRepository) Update(ctx context.Context, sprint *models.Sprint) error {
	query := `
		UPDATE sprints 
		SET name = $2, description = $3, start_date = $4, end_date = $5, status = $6, updated_at = $7
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		sprint.ID, sprint.Name, sprint.Description, sprint.StartDate,
		sprint.EndDate, sprint.Status, sprint.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("error updating sprint: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("sprint not found")
	}
	
	return nil
}

// Delete deletes a sprint by ID
func (r *sprintRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sprints WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting sprint: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("sprint not found")
	}
	
	return nil
}

// GetActiveSprint retrieves the active sprint for a project
func (r *sprintRepository) GetActiveSprint(ctx context.Context, projectID uuid.UUID) (*models.Sprint, error) {
	query := `
		SELECT id, name, description, project_id, start_date, end_date, status, created_at, updated_at
		FROM sprints WHERE project_id = $1 AND status = 'active' LIMIT 1
	`
	
	sprint := &models.Sprint{}
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&sprint.ID, &sprint.Name, &sprint.Description, &sprint.ProjectID,
		&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active sprint
		}
		return nil, fmt.Errorf("error getting active sprint: %w", err)
	}
	
	return sprint, nil
}

// GetUpcomingSprints retrieves upcoming sprints for a project
func (r *sprintRepository) GetUpcomingSprints(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error) {
	query := `
		SELECT id, name, description, project_id, start_date, end_date, status, created_at, updated_at
		FROM sprints WHERE project_id = $1 AND status = 'planned' ORDER BY start_date ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting upcoming sprints: %w", err)
	}
	defer rows.Close()
	
	var sprints []*models.Sprint
	for rows.Next() {
		sprint := &models.Sprint{}
		err := rows.Scan(
			&sprint.ID, &sprint.Name, &sprint.Description, &sprint.ProjectID,
			&sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning sprint: %w", err)
		}
		
		sprints = append(sprints, sprint)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sprints: %w", err)
	}
	
	return sprints, nil
}
