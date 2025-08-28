package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Sprint represents a sprint in the system
// @Description Sprint represents a sprint in the system
type Sprint struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name" gorm:"not null;size:100" example:"Sprint 1"`
	Description string    `json:"description" gorm:"size:500" example:"First sprint of the project"`
	ProjectID   uuid.UUID `json:"project_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000"`
	StartDate   time.Time `json:"start_date" gorm:"not null" example:"2024-01-01T00:00:00Z"`
	EndDate     time.Time `json:"end_date" gorm:"not null" example:"2024-01-15T00:00:00Z"`
	Status      string    `json:"status" gorm:"not null;default:'planned';size:20" example:"planned"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// Additional fields for API responses
	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Tasks   []*Task  `json:"tasks,omitempty" gorm:"foreignKey:SprintID"`
}

// TableName specifies the table name for the Sprint model
func (Sprint) TableName() string {
	return "sprints"
}

// SprintCreateRequest represents the request to create a new sprint
// @Description Request to create a new sprint
type SprintCreateRequest struct {
	Name        string    `json:"name" binding:"required,min=1,max=100" example:"Sprint 1"`
	Description string    `json:"description" binding:"max=500" example:"First sprint of the project"`
	StartDate   time.Time `json:"start_date" binding:"required" example:"2024-01-01T00:00:00Z"`
	EndDate     time.Time `json:"end_date" binding:"required" example:"2024-01-15T00:00:00Z"`
}

// SprintUpdateRequest represents the request to update a sprint
// @Description Request to update an existing sprint
type SprintUpdateRequest struct {
	Name        *string    `json:"name,omitempty" binding:"omitempty,min=1,max=100" example:"Updated Sprint Name"`
	Description *string    `json:"description,omitempty" binding:"omitempty,max=500" example:"Updated sprint description"`
	StartDate   *time.Time `json:"start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	EndDate     *time.Time `json:"end_date,omitempty" example:"2024-01-15T00:00:00Z"`
	Status      *string    `json:"status,omitempty" binding:"omitempty,oneof=planned active completed cancelled" example:"active"`
}

// CreateSprint creates a new sprint in the database using GORM
func CreateSprint(db *gorm.DB, req SprintCreateRequest, projectID uuid.UUID) (*Sprint, error) {
	sprint := &Sprint{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   projectID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      "planned", // Default status
	}

	if err := db.Create(sprint).Error; err != nil {
		return nil, err
	}

	return sprint, nil
}

// GetSprint retrieves a sprint by ID using GORM
func GetSprint(db *gorm.DB, sprintID uuid.UUID) (*Sprint, error) {
	var sprint Sprint
	if err := db.Preload("Project").Preload("Tasks").Where("id = ?", sprintID).First(&sprint).Error; err != nil {
		return nil, err
	}
	return &sprint, nil
}

// GetSprints retrieves all sprints for a specific project using GORM
func GetSprints(db *gorm.DB, projectID uuid.UUID) ([]*Sprint, error) {
	var sprints []*Sprint
	if err := db.Where("project_id = ?", projectID).Order("start_date ASC").Find(&sprints).Error; err != nil {
		return nil, err
	}
	return sprints, nil
}

// UpdateSprint updates an existing sprint using GORM
func UpdateSprint(db *gorm.DB, sprintID uuid.UUID, req SprintUpdateRequest) (*Sprint, error) {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := db.Model(&Sprint{}).Where("id = ?", sprintID).Updates(updates).Error; err != nil {
		return nil, err
	}

	return GetSprint(db, sprintID)
}

// DeleteSprint deletes a sprint by ID using GORM
func DeleteSprint(db *gorm.DB, sprintID uuid.UUID) error {
	return db.Where("id = ?", sprintID).Delete(&Sprint{}).Error
}

// GetActiveSprint retrieves the currently active sprint for a project using GORM
func GetActiveSprint(db *gorm.DB, projectID uuid.UUID) (*Sprint, error) {
	var sprint Sprint
	if err := db.Where("project_id = ? AND status = ?", projectID, "active").First(&sprint).Error; err != nil {
		return nil, err
	}
	return &sprint, nil
}

// GetUpcomingSprints retrieves upcoming sprints for a project using GORM
func GetUpcomingSprints(db *gorm.DB, projectID uuid.UUID) ([]*Sprint, error) {
	var sprints []*Sprint
	now := time.Now()

	if err := db.Where("project_id = ? AND start_date > ? AND status = ?", projectID, now, "planned").
		Order("start_date ASC").Find(&sprints).Error; err != nil {
		return nil, err
	}

	return sprints, nil
}

// GetSprintByStatus retrieves sprints by status for a project using GORM
func GetSprintByStatus(db *gorm.DB, projectID uuid.UUID, status string) ([]*Sprint, error) {
	var sprints []*Sprint
	if err := db.Where("project_id = ? AND status = ?", projectID, status).
		Order("start_date ASC").Find(&sprints).Error; err != nil {
		return nil, err
	}
	return sprints, nil
}

// CountSprints counts the total number of sprints for a project
func CountSprints(db *gorm.DB, projectID uuid.UUID) (int64, error) {
	var count int64
	if err := db.Model(&Sprint{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ValidateSprintDates validates that sprint dates are logical
func (s *Sprint) ValidateSprintDates() error {
	if s.StartDate.After(s.EndDate) {
		return ErrInvalidSprintDates
	}
	return nil
}

// IsSprintActive checks if a sprint is currently active
func (s *Sprint) IsSprintActive() bool {
	now := time.Now()
	return s.Status == "active" && now.After(s.StartDate) && now.Before(s.EndDate)
}

// IsSprintCompleted checks if a sprint is completed
func (s *Sprint) IsSprintCompleted() bool {
	return s.Status == "completed"
}

// BeforeCreate is a GORM hook that runs before creating a sprint
func (s *Sprint) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return s.ValidateSprintDates()
}

// BeforeUpdate is a GORM hook that runs before updating a sprint
func (s *Sprint) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return s.ValidateSprintDates()
}

// StartSprint starts a sprint by changing its status to active
func StartSprint(db *gorm.DB, sprintID uuid.UUID) (*Sprint, error) {
	// First check if there's already an active sprint in the same project
	sprint, err := GetSprint(db, sprintID)
	if err != nil {
		return nil, err
	}

	// Check if there's already an active sprint in this project
	var activeSprint Sprint
	if err := db.Where("project_id = ? AND status = ?", sprint.ProjectID, "active").First(&activeSprint).Error; err == nil {
		// There's already an active sprint, return error
		return nil, ErrSprintAlreadyActive
	}

	// Start the sprint
	if err := db.Model(&Sprint{}).Where("id = ?", sprintID).Update("status", "active").Error; err != nil {
		return nil, err
	}

	return GetSprint(db, sprintID)
}

// CompleteSprint completes a sprint by changing its status to completed
func CompleteSprint(db *gorm.DB, sprintID uuid.UUID) (*Sprint, error) {
	// Complete the sprint
	if err := db.Model(&Sprint{}).Where("id = ?", sprintID).Update("status", "completed").Error; err != nil {
		return nil, err
	}

	return GetSprint(db, sprintID)
}
