package services

import (
	"context"
	"fmt"
	"strings"

	"devhive-backend/models"
	"devhive-backend/repositories"

	"github.com/google/uuid"
)

// ProjectService defines the interface for project management operations
type ProjectService interface {
	CreateProject(ctx context.Context, req models.ProjectCreateRequest, ownerID uuid.UUID) (*models.Project, error)
	GetProject(ctx context.Context, projectID uuid.UUID) (*models.Project, error)
	GetProjectsByUser(ctx context.Context, userID uuid.UUID) ([]*models.Project, error)
	UpdateProject(ctx context.Context, projectID uuid.UUID, req models.ProjectUpdateRequest, userID uuid.UUID) (*models.Project, error)
	DeleteProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error
	AddMember(ctx context.Context, projectID, userID, ownerID uuid.UUID) error
	RemoveMember(ctx context.Context, projectID, userID, ownerID uuid.UUID) error
	// Mobile-specific methods
	GetProjectsForMobile(userID string, page, limit int, search string) ([]models.MobileProject, int, error)
	GetProjectForMobile(projectID uuid.UUID, userID string) (*models.MobileProject, error)
}

// projectService implements ProjectService
type projectService struct {
	projectRepo repositories.ProjectRepository
	userRepo    repositories.UserRepository
}

// NewProjectService creates a new project service instance
func NewProjectService(projectRepo repositories.ProjectRepository, userRepo repositories.UserRepository) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

// CreateProject creates a new project
func (s *projectService) CreateProject(ctx context.Context, req models.ProjectCreateRequest, ownerID uuid.UUID) (*models.Project, error) {
	// Create project
	project := &models.Project{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		ProjectOwnerID: ownerID,
	}

	// Save to database
	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("error creating project: %w", err)
	}

	// Add owner as project member
	if err := s.projectRepo.AddMember(ctx, project.ID, ownerID); err != nil {
		return nil, fmt.Errorf("error adding owner as member: %w", err)
	}

	return project, nil
}

// GetProject retrieves a project by ID
func (s *projectService) GetProject(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	return project, nil
}

// GetProjectsByUser retrieves projects for a specific user
func (s *projectService) GetProjectsByUser(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	projects, err := s.projectRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user projects: %w", err)
	}

	return projects, nil
}

// UpdateProject updates an existing project
func (s *projectService) UpdateProject(ctx context.Context, projectID uuid.UUID, req models.ProjectUpdateRequest, userID uuid.UUID) (*models.Project, error) {
	// Get current project
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	// Check if user is the owner
	if project.ProjectOwnerID != userID {
		return nil, fmt.Errorf("only project owner can update project")
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}

	// Save changes
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("error updating project: %w", err)
	}

	return project, nil
}

// DeleteProject deletes a project
func (s *projectService) DeleteProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error {
	// Get current project
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("error getting project: %w", err)
	}

	// Check if user is the owner
	if project.ProjectOwnerID != userID {
		return fmt.Errorf("only project owner can delete project")
	}

	// Delete project
	if err := s.projectRepo.Delete(ctx, projectID); err != nil {
		return fmt.Errorf("error deleting project: %w", err)
	}

	return nil
}

// AddMember adds a user to a project
func (s *projectService) AddMember(ctx context.Context, projectID, userID, ownerID uuid.UUID) error {
	// Check if user is the owner
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("error getting project: %w", err)
	}

	if project.ProjectOwnerID != ownerID {
		return fmt.Errorf("only project owner can add members")
	}

	// Add member
	if err := s.projectRepo.AddMember(ctx, projectID, userID); err != nil {
		return fmt.Errorf("error adding member: %w", err)
	}

	return nil
}

// RemoveMember removes a user from a project
func (s *projectService) RemoveMember(ctx context.Context, projectID, userID, ownerID uuid.UUID) error {
	// Check if user is the owner
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("error getting project: %w", err)
	}

	if project.ProjectOwnerID != ownerID {
		return fmt.Errorf("only project owner can remove members")
	}

	// Remove member
	if err := s.projectRepo.RemoveMember(ctx, projectID, userID); err != nil {
		return fmt.Errorf("error removing member: %w", err)
	}

	return nil
}

// GetProjectsForMobile retrieves projects optimized for mobile consumption
func (s *projectService) GetProjectsForMobile(userID string, page, limit int, search string) ([]models.MobileProject, int, error) {
	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get projects for user
	projects, err := s.projectRepo.GetByUserID(context.Background(), userUUID)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting user projects: %w", err)
	}

	// Convert to mobile projects
	mobileProjects := make([]models.MobileProject, 0, len(projects))
	for _, project := range projects {
		// Apply search filter if provided
		if search != "" && !containsString(project.Name, search) && !containsString(project.Description, search) {
			continue
		}

		// For now, we'll use default values for counts since the repository doesn't have these methods yet
		memberCount := 1 // At least the owner
		sprintCount := 0 // Will be implemented later
		taskCount := 0   // Will be implemented later

		mobileProject := models.MobileProject{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			Status:      "active", // Default status since it's not in the model
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
			OwnerID:     project.ProjectOwnerID,
			OwnerName:   "", // Will be populated if needed
			MemberCount: memberCount,
			SprintCount: sprintCount,
			TaskCount:   taskCount,
			IsOwner:     project.ProjectOwnerID == userUUID,
			IsMember:    true, // User is a member if they can see the project
		}

		// Get owner name
		if owner, err := s.userRepo.GetByID(context.Background(), project.ProjectOwnerID); err == nil {
			mobileProject.OwnerName = owner.Username
		}

		mobileProjects = append(mobileProjects, mobileProject)
	}

	// Apply pagination
	total := len(mobileProjects)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return []models.MobileProject{}, total, nil
	}

	if end > total {
		end = total
	}

	return mobileProjects[start:end], total, nil
}

// GetProjectForMobile retrieves a single project optimized for mobile consumption
func (s *projectService) GetProjectForMobile(projectID uuid.UUID, userID string) (*models.MobileProject, error) {
	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get project
	project, err := s.projectRepo.GetByID(context.Background(), projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}

	// Check if user is a member
	isMember, err := s.projectRepo.IsMember(context.Background(), projectID, userUUID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("access denied")
	}

	// For now, we'll use default values for counts since the repository doesn't have these methods yet
	memberCount := 1 // At least the owner
	sprintCount := 0 // Will be implemented later
	taskCount := 0   // Will be implemented later

	mobileProject := &models.MobileProject{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Status:      "active", // Default status since it's not in the model
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		OwnerID:     project.ProjectOwnerID,
		OwnerName:   "", // Will be populated if needed
		MemberCount: memberCount,
		SprintCount: sprintCount,
		TaskCount:   taskCount,
		IsOwner:     project.ProjectOwnerID == userUUID,
		IsMember:    true,
	}

	// Get owner name
	if owner, err := s.userRepo.GetByID(context.Background(), project.ProjectOwnerID); err == nil {
		mobileProject.OwnerName = owner.Username
	}

	return mobileProject, nil
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}
