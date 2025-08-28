package services

import (
	"context"
	"fmt"

	"devhive-backend/models"
	"devhive-backend/repositories"

	"github.com/google/uuid"
)

// UserService defines the interface for user management operations
type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req models.UserUpdateRequest) (*models.User, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error
	GetUsersByProject(ctx context.Context, projectID uuid.UUID) ([]*models.User, error)
}

// userService implements UserService
type userService struct {
	userRepo repositories.UserRepository
}

// NewUserService creates a new user service instance
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// GetProfile retrieves a user's profile
func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user profile: %w", err)
	}

	// Clear sensitive information
	user.Password = ""

	return user, nil
}

// UpdateProfile updates a user's profile
func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req models.UserUpdateRequest) (*models.User, error) {
	// Get current user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Update fields if provided
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Active != nil {
		user.Active = *req.Active
	}

	// Update password if provided
	if req.Password != nil {
		// TODO: Hash password before updating
		user.Password = *req.Password
	}

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("error updating user profile: %w", err)
	}

	// Clear sensitive information
	user.Password = ""

	return user, nil
}

// UploadAvatar updates a user's avatar URL
func (s *userService) UploadAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	if err := s.userRepo.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		return fmt.Errorf("error updating avatar: %w", err)
	}

	return nil
}

// GetUsersByProject retrieves users for a specific project
func (s *userService) GetUsersByProject(ctx context.Context, projectID uuid.UUID) ([]*models.User, error) {
	users, err := s.userRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting project users: %w", err)
	}

	// Clear sensitive information from all users
	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}
