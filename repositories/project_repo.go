package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"devhive-backend/models"

	"github.com/google/uuid"
)

// ProjectRepository defines the interface for project data operations
type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Project, error)
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*models.Project, error)
	AddMember(ctx context.Context, projectID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
	IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
}

// projectRepository implements ProjectRepository
type projectRepository struct {
	db *sql.DB
}

// NewProjectRepository creates a new project repository instance
func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// Create creates a new project
func (r *projectRepository) Create(ctx context.Context, project *models.Project) error {
	query := `
		INSERT INTO projects (id, name, description, project_owner_id)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query,
		project.ID, project.Name, project.Description, project.ProjectOwnerID)

	return err
}

// GetByID retrieves a project by ID
func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, name, description, project_owner_id
		FROM projects WHERE id = $1
	`

	project := &models.Project{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID, &project.Name, &project.Description, &project.ProjectOwnerID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %w", err)
		}
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	return project, nil
}

// GetByUserID retrieves projects by user ID
func (r *projectRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.project_owner_id
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.project_owner_id = $1 OR pm.user_id = $1
		ORDER BY p.name
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting projects by user: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.ProjectOwnerID)

		if err != nil {
			return nil, fmt.Errorf("error scanning project: %w", err)
		}

		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, nil
}

// Update updates an existing project
func (r *projectRepository) Update(ctx context.Context, project *models.Project) error {
	query := `
		UPDATE projects 
		SET name = $2, description = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		project.ID, project.Name, project.Description)

	if err != nil {
		return fmt.Errorf("error updating project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// Delete deletes a project by ID
func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// GetAll retrieves all projects
func (r *projectRepository) GetAll(ctx context.Context) ([]*models.Project, error) {
	query := `
		SELECT id, name, description, project_owner_id
		FROM projects ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.ProjectOwnerID)

		if err != nil {
			return nil, fmt.Errorf("error scanning project: %w", err)
		}

		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, nil
}

// AddMember adds a user to a project
func (r *projectRepository) AddMember(ctx context.Context, projectID, userID uuid.UUID) error {
	query := `
		INSERT INTO project_members (project_id, user_id, joined_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (project_id, user_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, projectID, userID)
	return err
}

// RemoveMember removes a user from a project
func (r *projectRepository) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, projectID, userID)
	if err != nil {
		return fmt.Errorf("error removing member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

// IsMember checks if a user is a member of a project
func (r *projectRepository) IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM project_members 
			WHERE project_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking membership: %w", err)
	}

	return exists, nil
}
