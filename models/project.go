package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project represents a project in the system
// @Description Project represents a project in the system
type Project struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name           string    `json:"name" gorm:"not null;size:50" example:"My Project"`
	Description    string    `json:"description" gorm:"not null;size:255" example:"A description of my project"`
	ProjectOwnerID uuid.UUID `json:"project_owner_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// Additional fields for API responses
	Owner       *User   `json:"owner,omitempty" gorm:"foreignKey:ProjectOwnerID"`
	Members     []*User `json:"members,omitempty" gorm:"many2many:project_members;"`
	MemberCount int     `json:"member_count,omitempty" gorm:"-" example:"5"`
}

// TableName specifies the table name for the Project model
func (Project) TableName() string {
	return "projects"
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
	ProjectID uuid.UUID `json:"project_id" gorm:"type:uuid;primaryKey" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey" example:"123e4567-e89b-12d3-a456-426614174000"`
	Role      string    `json:"role" gorm:"not null;default:'member';size:20" example:"member"`
	JoinedAt  time.Time `json:"joined_at" gorm:"autoCreateTime"`
	// Additional fields for API responses
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for the ProjectMember model
func (ProjectMember) TableName() string {
	return "project_members"
}

// AddMemberRequest represents the request to add a member to a project
// @Description Request to add a user as a project member
type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Role   string    `json:"role" binding:"required,oneof=viewer member admin owner" example:"member"`
}

// CreateProject creates a new project in the database using GORM
func CreateProject(db *gorm.DB, req ProjectCreateRequest, ownerID uuid.UUID) (*Project, error) {
	project := &Project{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		ProjectOwnerID: ownerID,
	}

	if err := db.Create(project).Error; err != nil {
		return nil, err
	}

	// Add owner as project member with owner role
	projectMember := &ProjectMember{
		ProjectID: project.ID,
		UserID:    ownerID,
		Role:      "owner",
	}

	if err := db.Create(projectMember).Error; err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject retrieves a project by ID with owner and member information using GORM
func GetProject(db *gorm.DB, projectID uuid.UUID) (*Project, error) {
	var project Project
	if err := db.Preload("Owner").Preload("Members").Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, err
	}

	// Calculate member count
	var memberCount int64
	db.Model(&ProjectMember{}).Where("project_id = ?", projectID).Count(&memberCount)
	project.MemberCount = int(memberCount)

	return &project, nil
}

// GetProjects retrieves all projects for a user using GORM
func GetProjects(db *gorm.DB, userID uuid.UUID) ([]*Project, error) {
	var projects []*Project

	// Get projects where user is owner or member
	if err := db.Joins("JOIN project_members ON projects.id = project_members.project_id").
		Where("project_members.user_id = ? OR projects.project_owner_id = ?", userID, userID).
		Preload("Owner").
		Find(&projects).Error; err != nil {
		return nil, err
	}

	return projects, nil
}

// UpdateProject updates an existing project using GORM
func UpdateProject(db *gorm.DB, projectID uuid.UUID, req ProjectUpdateRequest) (*Project, error) {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if err := db.Model(&Project{}).Where("id = ?", projectID).Updates(updates).Error; err != nil {
		return nil, err
	}

	return GetProject(db, projectID)
}

// DeleteProject deletes a project and all its members using GORM
func DeleteProject(db *gorm.DB, projectID uuid.UUID) error {
	// Delete project members first
	if err := db.Where("project_id = ?", projectID).Delete(&ProjectMember{}).Error; err != nil {
		return err
	}

	// Delete the project
	return db.Where("id = ?", projectID).Delete(&Project{}).Error
}

// AddProjectMember adds a user as a member of a project using GORM
func AddProjectMember(db *gorm.DB, projectID, userID uuid.UUID, role string) error {
	projectMember := &ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}

	return db.Create(projectMember).Error
}

// RemoveProjectMember removes a user from a project using GORM
func RemoveProjectMember(db *gorm.DB, projectID, userID uuid.UUID) error {
	return db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&ProjectMember{}).Error
}

// IsProjectMember checks if a user is a member of a project using GORM
func IsProjectMember(db *gorm.DB, projectID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := db.Model(&ProjectMember{}).Where("project_id = ? AND user_id = ?", projectID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetProjectMemberRole gets the role of a user in a project using GORM
func GetProjectMemberRole(db *gorm.DB, projectID, userID uuid.UUID) (string, error) {
	var projectMember ProjectMember
	if err := db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&projectMember).Error; err != nil {
		return "", err
	}
	return projectMember.Role, nil
}

// GetProjectMembers gets all members of a project using GORM
func GetProjectMembers(db *gorm.DB, projectID uuid.UUID) ([]*ProjectMember, error) {
	var members []*ProjectMember
	if err := db.Where("project_id = ?", projectID).Preload("User").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// UpdateProjectMemberRole updates the role of a project member
func UpdateProjectMemberRole(db *gorm.DB, projectID, userID uuid.UUID, newRole string) error {
	return db.Model(&ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Update("role", newRole).Error
}

// BeforeCreate is a GORM hook that runs before creating a project
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a project
func (p *Project) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}
