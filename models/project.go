package models

import (
	"database/sql"

	"github.com/google/uuid"
)

// Project represents a project in the system
// @Description Project represents a project in the system
type Project struct {
	ID             uuid.UUID `json:"id" db:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name           string    `json:"name" db:"name" example:"My Project"`
	Description    string    `json:"description" db:"description" example:"A description of my project"`
	ProjectOwnerID uuid.UUID `json:"project_owner_id" db:"project_owner_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Additional fields for API responses
	Owner       *User   `json:"owner,omitempty"`
	Members     []*User `json:"members,omitempty"`
	MemberCount int     `json:"member_count,omitempty" example:"5"`
}

// ProjectCreateRequest represents the request to create a new project
// @Description Request to create a new project
type ProjectCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50" example:"My New Project"`
	Description string `json:"description" binding:"required,min=1,max=255" example:"Description of my new project"`
}

// ProjectUpdateRequest represents the request to update a project
// @Description Request to update an existing project
type ProjectUpdateRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=50" example:"Updated Project Name"`
	Description *string `json:"description,omitempty" binding:"omitempty,min=1,max=255" example:"Updated project description"`
}

// ProjectMember represents a project member (from project_has_users table)
// @Description Project member information
type ProjectMember struct {
	ProjectID uuid.UUID `json:"project_id" db:"project_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID    uuid.UUID `json:"user_id" db:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Additional fields for API responses
	User *User `json:"user,omitempty"`
}

// AddMemberRequest represents the request to add a member to a project
// @Description Request to add a user as a project member
type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// CreateProject creates a new project in the database
func CreateProject(db *sql.DB, req ProjectCreateRequest, ownerID uuid.UUID) (*Project, error) {
	project := &Project{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		ProjectOwnerID: ownerID,
	}

	query := `
		INSERT INTO projects (id, name, description, project_owner_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, project_owner_id
	`

	err := db.QueryRow(
		query,
		project.ID, project.Name, project.Description, project.ProjectOwnerID,
	).Scan(
		&project.ID, &project.Name, &project.Description, &project.ProjectOwnerID,
	)

	if err != nil {
		return nil, err
	}

	// Add owner as project member
	if err := AddProjectMember(db, project.ID, ownerID); err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject retrieves a project by ID with owner and member information
func GetProject(db *sql.DB, projectID uuid.UUID) (*Project, error) {
	project := &Project{}
	query := `
		SELECT p.id, p.name, p.description, p.project_owner_id
		FROM projects p WHERE p.id = $1
	`

	err := db.QueryRow(query, projectID).Scan(
		&project.ID, &project.Name, &project.Description, &project.ProjectOwnerID,
	)

	if err != nil {
		return nil, err
	}

	// Get owner information
	owner, err := GetUserByID(db, project.ProjectOwnerID)
	if err != nil {
		return nil, err
	}
	project.Owner = owner

	// Get member count
	var memberCount int
	countQuery := `SELECT COUNT(*) FROM project_has_users WHERE project_id = $1`
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
		SELECT DISTINCT p.id, p.name, p.description, p.project_owner_id
		FROM projects p
		INNER JOIN project_has_users phu ON p.id = phu.project_id
		WHERE phu.user_id = $1
		ORDER BY p.id DESC
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
			&project.ID, &project.Name, &project.Description, &project.ProjectOwnerID,
		)
		if err != nil {
			return nil, err
		}

		// Get owner information
		owner, err := GetUserByID(db, project.ProjectOwnerID)
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
		project.Description = *req.Description
	}

	query := `
		UPDATE projects 
		SET name = $1, description = $2
		WHERE id = $3
		RETURNING id, name, description, project_owner_id
	`

	err = db.QueryRow(
		query,
		project.Name, project.Description, project.ID,
	).Scan(
		&project.ID, &project.Name, &project.Description, &project.ProjectOwnerID,
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
func AddProjectMember(db *sql.DB, projectID, userID uuid.UUID) error {
	query := `
		INSERT INTO project_has_users (project_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (project_id, user_id) DO NOTHING
	`
	_, err := db.Exec(query, projectID, userID)
	return err
}

// RemoveProjectMember removes a member from a project
func RemoveProjectMember(db *sql.DB, projectID, userID uuid.UUID) error {
	query := `DELETE FROM project_has_users WHERE project_id = $1 AND user_id = $2`
	_, err := db.Exec(query, projectID, userID)
	return err
}

// GetProjectMembers retrieves all members of a project
func GetProjectMembers(db *sql.DB, projectID uuid.UUID) ([]*ProjectMember, error) {
	query := `
		SELECT phu.project_id, phu.user_id,
		       u.id, u.username, u.email, u.first_name, u.last_name, u.active
		FROM project_has_users phu
		INNER JOIN users u ON phu.user_id = u.id
		WHERE phu.project_id = $1
		ORDER BY u.first_name ASC
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
			&member.ProjectID, &member.UserID,
			&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Active,
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
	query := `SELECT EXISTS(SELECT 1 FROM project_has_users WHERE project_id = $1 AND user_id = $2)`
	err := db.QueryRow(query, projectID, userID).Scan(&exists)
	return exists, err
}

// IsProjectOwner checks if a user is the owner of a project
func IsProjectOwner(db *sql.DB, projectID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND project_owner_id = $2)`
	err := db.QueryRow(query, projectID, userID).Scan(&exists)
	return exists, err
}

// GetProjectMemberRole returns the role of a user in a project
// Currently supports "owner" and "member" roles
func GetProjectMemberRole(db *sql.DB, projectID, userID uuid.UUID) (string, error) {
	// Check if user is the project owner
	isOwner, err := IsProjectOwner(db, projectID, userID)
	if err != nil {
		return "", err
	}
	if isOwner {
		return "owner", nil
	}

	// Check if user is a project member
	isMember, err := IsProjectMember(db, projectID, userID)
	if err != nil {
		return "", err
	}
	if isMember {
		return "member", nil
	}

	return "", nil
}
