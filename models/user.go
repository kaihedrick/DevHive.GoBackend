package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FirebaseUID string    `json:"firebase_uid" db:"firebase_uid"`
	Email       string    `json:"email" db:"email"`
	Username    string    `json:"username" db:"username"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	Bio         *string   `json:"bio,omitempty" db:"bio"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserCreateRequest represents the request to create a new user
type UserCreateRequest struct {
	FirebaseUID string `json:"firebase_uid" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=100"`
	FirstName   string `json:"first_name" binding:"required,min=1,max=100"`
	LastName    string `json:"last_name" binding:"required,min=1,max=100"`
}

// UserUpdateRequest represents the request to update a user
type UserUpdateRequest struct {
	FirstName *string `json:"first_name,omitempty" binding:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" binding:"omitempty,min=1,max=100"`
	Bio       *string `json:"bio,omitempty" binding:"omitempty,max=500"`
}

// CreateUser creates a new user in the database
func CreateUser(db *sql.DB, req UserCreateRequest) (*User, error) {
	user := &User{
		ID:          uuid.New(),
		FirebaseUID: req.FirebaseUID,
		Email:       req.Email,
		Username:    req.Username,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO users (id, firebase_uid, email, username, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, firebase_uid, email, username, first_name, last_name, avatar_url, bio, created_at, updated_at
	`

	err := db.QueryRow(
		query,
		user.ID, user.FirebaseUID, user.Email, user.Username,
		user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt,
	).Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.Username,
		&user.FirstName, &user.LastName, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(db *sql.DB, userID uuid.UUID) (*User, error) {
	user := &User{}
	query := `
		SELECT id, firebase_uid, email, username, first_name, last_name, avatar_url, bio, created_at, updated_at
		FROM users WHERE id = $1
	`

	err := db.QueryRow(query, userID).Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.Username,
		&user.FirstName, &user.LastName, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByFirebaseUID retrieves a user by Firebase UID
func GetUserByFirebaseUID(db *sql.DB, firebaseUID string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, firebase_uid, email, username, first_name, last_name, avatar_url, bio, created_at, updated_at
		FROM users WHERE firebase_uid = $1
	`

	err := db.QueryRow(query, firebaseUID).Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.Username,
		&user.FirstName, &user.LastName, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, firebase_uid, email, username, first_name, last_name, avatar_url, bio, created_at, updated_at
		FROM users WHERE email = $1
	`

	err := db.QueryRow(query, email).Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.Username,
		&user.FirstName, &user.LastName, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates a user in the database
func UpdateUser(db *sql.DB, userID uuid.UUID, req UserUpdateRequest) (*User, error) {
	user, err := GetUserByID(db, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users 
		SET first_name = $1, last_name = $2, bio = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, firebase_uid, email, username, first_name, last_name, avatar_url, bio, created_at, updated_at
	`

	err = db.QueryRow(
		query,
		user.FirstName, user.LastName, user.Bio, user.UpdatedAt, user.ID,
	).Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.Username,
		&user.FirstName, &user.LastName, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserAvatar updates a user's avatar URL
func UpdateUserAvatar(db *sql.DB, userID uuid.UUID, avatarURL string) error {
	query := `UPDATE users SET avatar_url = $1, updated_at = $2 WHERE id = $3`
	_, err := db.Exec(query, avatarURL, time.Now(), userID)
	return err
}

// DeleteUser deletes a user from the database
func DeleteUser(db *sql.DB, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := db.Exec(query, userID)
	return err
}
