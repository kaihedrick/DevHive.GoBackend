package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null;size:30"`
	Password  string    `json:"password" gorm:"not null;size:100"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null;size:100"`
	FirstName string    `json:"first_name" gorm:"not null;size:50"`
	LastName  string    `json:"last_name" gorm:"not null;size:50"`
	Active    bool      `json:"active" gorm:"default:true"`
	AvatarURL *string   `json:"avatar_url,omitempty" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// UserCreateRequest represents the request to create a new user
type UserCreateRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=30"`
	Password  string `json:"password" binding:"required,min=6,max=100"`
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required,min=1,max=50"`
	LastName  string `json:"last_name" binding:"required,min=1,max=50"`
}

// UserUpdateRequest represents the request to update a user
type UserUpdateRequest struct {
	Username  *string `json:"username,omitempty" binding:"omitempty,min=3,max=30"`
	Password  *string `json:"password,omitempty" binding:"omitempty,min=6,max=100"`
	Email     *string `json:"email,omitempty" binding:"omitempty,email"`
	FirstName *string `json:"first_name,omitempty" binding:"omitempty,min=1,max=50"`
	LastName  *string `json:"last_name,omitempty" binding:"omitempty,min=1,max=50"`
	Active    *bool   `json:"active,omitempty"`
}

// CreateUser creates a new user in the database using GORM
func CreateUser(db *gorm.DB, req UserCreateRequest) (*User, error) {
	user := &User{
		ID:        uuid.New(),
		Username:  req.Username,
		Password:  req.Password,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Active:    true, // Default to active
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID using GORM
func GetUserByID(db *gorm.DB, userID uuid.UUID) (*User, error) {
	var user User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username using GORM
func GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email using GORM
func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user using GORM
func UpdateUser(db *gorm.DB, userID uuid.UUID, req UserUpdateRequest) (*User, error) {
	updates := make(map[string]interface{})

	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	if err := db.Model(&User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return nil, err
	}

	return GetUserByID(db, userID)
}

// DeleteUser deletes a user by ID using GORM
func DeleteUser(db *gorm.DB, userID uuid.UUID) error {
	return db.Where("id = ?", userID).Delete(&User{}).Error
}

// ListUsers retrieves all users with optional pagination using GORM
func ListUsers(db *gorm.DB, offset, limit int) ([]User, error) {
	var users []User
	if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// SearchUsers searches for users by username or email using GORM
func SearchUsers(db *gorm.DB, query string, offset, limit int) ([]User, error) {
	var users []User
	searchQuery := "%" + query + "%"
	if err := db.Where("username ILIKE ? OR email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
		searchQuery, searchQuery, searchQuery, searchQuery).
		Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
