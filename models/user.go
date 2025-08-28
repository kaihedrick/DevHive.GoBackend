package models

import (
	"database/sql"

	"github.com/google/uuid"
)

// User represents a user in the system
// @Description User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" db:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username  string    `json:"username" db:"username" example:"johndoe"`
	Password  string    `json:"password" db:"password" example:"hashedpassword"`
	Email     string    `json:"email" db:"email" example:"john@example.com"`
	FirstName string    `json:"first_name" db:"first_name" example:"John"`
	LastName  string    `json:"last_name" db:"last_name" example:"Doe"`
	Active    bool      `json:"active" db:"active" example:"true"`
	AvatarURL *string   `json:"avatar_url,omitempty" db:"avatar_url" example:"https://example.com/avatar.jpg"`
}

// UserCreateRequest represents the request to create a new user
// @Description Request to create a new user
type UserCreateRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=30" example:"johndoe"`
	Password  string `json:"password" binding:"required,min=6,max=100" example:"password123"`
	Email     string `json:"email" binding:"required,email" example:"john@example.com"`
	FirstName string `json:"first_name" binding:"required,min=1,max=50" example:"John"`
	LastName  string `json:"last_name" binding:"required,min=1,max=50" example:"Doe"`
}

// UserUpdateRequest represents the request to update a user
// @Description Request to update an existing user
type UserUpdateRequest struct {
	Username  *string `json:"username,omitempty" binding:"omitempty,min=3,max=30" example:"johndoe"`
	Password  *string `json:"password,omitempty" binding:"omitempty,min=6,max=100" example:"newpassword123"`
	Email     *string `json:"email,omitempty" binding:"omitempty,email" example:"john.doe@example.com"`
	FirstName *string `json:"first_name,omitempty" binding:"omitempty,min=1,max=50" example:"John"`
	LastName  *string `json:"last_name,omitempty" binding:"omitempty,min=1,max=50" example:"Doe"`
	Active    *bool   `json:"active,omitempty" example:"true"`
}

// CreateUser creates a new user in the database
func CreateUser(db *sql.DB, req UserCreateRequest) (*User, error) {
	user := &User{
		ID:        uuid.New(),
		Username:  req.Username,
		Password:  req.Password,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Active:    true, // Default to active
	}

	query := `
		INSERT INTO users (id, username, password, email, first_name, last_name, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, username, password, email, first_name, last_name, active
	`

	err := db.QueryRow(
		query,
		user.ID, user.Username, user.Password, user.Email,
		user.FirstName, user.LastName, user.Active,
	).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.FirstName, &user.LastName, &user.Active,
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
		SELECT id, username, password, email, first_name, last_name, active
		FROM users WHERE id = $1
	`

	err := db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.FirstName, &user.LastName, &user.Active,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, username, password, email, first_name, last_name, active
		FROM users WHERE username = $1
	`

	err := db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.FirstName, &user.LastName, &user.Active,
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
		SELECT id, username, password, email, first_name, last_name, active
		FROM users WHERE email = $1
	`

	err := db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.FirstName, &user.LastName, &user.Active,
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
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Password != nil {
		user.Password = *req.Password
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

	query := `
		UPDATE users 
		SET username = $1, password = $2, email = $3, first_name = $4, last_name = $5, active = $6
		WHERE id = $7
		RETURNING id, username, password, email, first_name, last_name, active
	`

	err = db.QueryRow(
		query,
		user.Username, user.Password, user.Email, user.FirstName, user.LastName, user.Active, user.ID,
	).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.FirstName, &user.LastName, &user.Active,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user from the database
func DeleteUser(db *sql.DB, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := db.Exec(query, userID)
	return err
}

// GetUsers retrieves all users
func GetUsers(db *sql.DB) ([]*User, error) {
	query := `
		SELECT id, username, password, email, first_name, last_name, active
		FROM users
		ORDER BY username ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Email,
			&user.FirstName, &user.LastName, &user.Active,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUserAvatar updates the avatar URL for a user
func UpdateUserAvatar(db *sql.DB, userID uuid.UUID, avatarURL string) error {
	query := `UPDATE users SET avatar_url = $1 WHERE id = $2`
	_, err := db.Exec(query, avatarURL, userID)
	return err
}
