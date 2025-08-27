package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Project represents a project in the system
type Project struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	// Additional fields for API responses
	Owner       *User   `json:"owner,omitempty"`
	Members     []*User `json:"members,omitempty"`
	MemberCount int     `json:"member_count,omitempty"`
}

// ProjectCreateRequest represents the request to create a new project
type ProjectCreateRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// ProjectUpdateRequest represents the request to update a project
type ProjectUpdateRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=active archived completed"`
}

// ProjectMember represents a project member
type ProjectMember struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
	// Additional fields for API responses
	User *User `json:"user,omitempty"`
}

// AddMemberRequest represents the request to add a member to a project
type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Role   string    `json:"role" binding:"required,oneof=admin member viewer"`
}

// CreateProject creates a new project in the database
func CreateProject(db *sql.DB, req ProjectCreateRequest, ownerID uuid.UUID) (*Project, error) {
	project := &Project{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO projects (id, name, description, owner_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, description, owner_id, status, created_at, updated_at
	`

	err := db.QueryRow(
		query,
		project.ID, project.Name, project.Description, project.OwnerID,
		project.Status, project.CreatedAt, project.UpdatedAt,
	).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID,
		&project.Status, &project.CreatedAt, &project.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Add owner as project member
	if err := AddProjectMember(db, project.ID, ownerID, "owner"); err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject retrieves a project by ID with owner and member information
func GetProject(db *sql.DB, projectID uuid.UUID) (*Project, error) {
	project := &Project{}
	query := `
		SELECT p.id, p.name, p.description, p.owner_id, p.status, p.created_at, p.updated_at
		FROM projects p WHERE p.id = $1
	`

	err := db.QueryRow(query, projectID).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID,
		&project.Status, &project.CreatedAt, &project.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Get owner information
	owner, err := GetUserByID(db, project.OwnerID)
	if err != nil {
		return nil, err
	}
	project.Owner = owner

	// Get member count
	var memberCount int
	countQuery := `SELECT COUNT(*) FROM project_members WHERE project_id = $1`
	err = db.QueryRow(countQuery, projectID).Scan(&memberCount)
	if err != nil {
		return nil, err
	}
	project.MemberCount = memberCount

	return project, nil
}

// GetProjects retrieves projects for a user (either owned or member of)
func GetProjects(db *sql.DB, userID uuid.UUID) ([]*Project, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.status, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1 AND p.status != 'archived'
		ORDER BY p.updated_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.OwnerID,
			&project.Status, &project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get owner information
		owner, err := GetUserByID(db, project.OwnerID)
		if err != nil {
			return nil, err
		}
		project.Owner = owner

		projects = append(projects, project)
	}

	return projects, nil
}

// UpdateProject updates a project in the database
func UpdateProject(db *sql.DB, projectID uuid.UUID, req ProjectUpdateRequest) (*Project, error) {
	project, err := GetProject(db, projectID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = req.Description
	}
	if req.Status != nil {
		project.Status = *req.Status
	}

	project.UpdatedAt = time.Now()

	query := `
		UPDATE projects 
		SET name = $1, description = $2, status = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, name, description, owner_id, status, created_at, updated_at
	`

	err = db.QueryRow(
		query,
		project.Name, project.Description, project.Status, project.UpdatedAt, project.ID,
	).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID,
		&project.Status, &project.CreatedAt, &project.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return project, nil
}

// DeleteProject deletes a project from the database
func DeleteProject(db *sql.DB, projectID uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := db.Exec(query, projectID)
	return err
}

// AddProjectMember adds a member to a project
func AddProjectMember(db *sql.DB, projectID, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, user_id) 
		DO UPDATE SET role = $3
	`
	_, err := db.Exec(query, projectID, userID, role)
	return err
}

// RemoveProjectMember removes a member from a project
func RemoveProjectMember(db *sql.DB, projectID, userID uuid.UUID) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`
	_, err := db.Exec(query, projectID, userID)
	return err
}

// GetProjectMembers retrieves all members of a project
func GetProjectMembers(db *sql.DB, projectID uuid.UUID) ([]*ProjectMember, error) {
	query := `
		SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.joined_at,
		       u.id, u.firebase_uid, u.email, u.username, u.first_name, u.last_name, u.avatar_url, u.bio, u.created_at, u.updated_at
		FROM project_members pm
		INNER JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at ASC
	`

	rows, err := db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*ProjectMember
	for rows.Next() {
		member := &ProjectMember{}
		user := &User{}

		err := rows.Scan(
			&member.ID, &member.ProjectID, &member.UserID, &member.Role, &member.JoinedAt,
			&user.ID, &user.FirebaseUID, &user.Email, &user.Username,
			&user.FirstName, &user.LastName, &user.AvatarURL, &user.Bio,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		member.User = user
		members = append(members, member)
	}

	return members, nil
}

// IsProjectMember checks if a user is a member of a project
func IsProjectMember(db *sql.DB, projectID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)`
	err := db.QueryRow(query, projectID, userID).Scan(&exists)
	return exists, err
}

// GetProjectMemberRole gets the role of a user in a project
func GetProjectMemberRole(db *sql.DB, projectID, userID uuid.UUID) (string, error) {
	var role string
	query := `SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2`
	err := db.QueryRow(query, projectID, userID).Scan(&role)
	return role, err
}
