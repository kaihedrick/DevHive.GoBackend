package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"devhive-backend/models"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*models.User, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.User, error)
	UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error
}

// userRepository implements UserRepository
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, username, password, first_name, last_name, active, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.Password,
		user.FirstName, user.LastName, user.Active, user.AvatarURL)
	
	return err
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name, active, avatar_url
		FROM users WHERE id = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password,
		&user.FirstName, &user.LastName, &user.Active, &user.AvatarURL)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name, active, avatar_url
		FROM users WHERE email = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password,
		&user.FirstName, &user.LastName, &user.Active, &user.AvatarURL)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}
	
	return user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = $2, username = $3, password = $4, first_name = $5, last_name = $6, active = $7, avatar_url = $8
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.Password,
		user.FirstName, user.LastName, user.Active, user.AvatarURL)
	
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// GetAll retrieves all users
func (r *userRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name, active, avatar_url
		FROM users ORDER BY username
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.Password,
			&user.FirstName, &user.LastName, &user.Active, &user.AvatarURL)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}
	
	return users, nil
}

// GetByProjectID retrieves users by project ID
func (r *userRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.password, u.first_name, u.last_name, u.active, u.avatar_url
		FROM users u
		JOIN project_members pm ON u.id = pm.user_id
		WHERE pm.project_id = $1
		ORDER BY u.username
	`
	
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting users by project: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.Password,
			&user.FirstName, &user.LastName, &user.Active, &user.AvatarURL)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}
	
	return users, nil
}

// UpdateAvatar updates a user's avatar URL
func (r *userRepository) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	query := `UPDATE users SET avatar_url = $2 WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, userID, avatarURL)
	if err != nil {
		return fmt.Errorf("error updating avatar: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}
