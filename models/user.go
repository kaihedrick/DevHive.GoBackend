package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
// @Description User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null;size:30" example:"johndoe"`
	Password  string    `json:"password" gorm:"not null;size:100" example:"hashedpassword"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null;size:100" example:"john@example.com"`
	FirstName string    `json:"first_name" gorm:"not null;size:50" example:"John"`
	LastName  string    `json:"last_name" gorm:"not null;size:50" example:"Doe"`
	Active    bool      `json:"active" gorm:"default:true" example:"true"`
	AvatarURL *string   `json:"avatar_url,omitempty" gorm:"size:255" example:"https://example.com/avatar.jpg"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
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

// UpdateUserAvatar updates a user's avatar URL using GORM
func UpdateUserAvatar(db *gorm.DB, userID uuid.UUID, avatarURL string) error {
	return db.Model(&User{}).Where("id = ?", userID).Update("avatar_url", avatarURL).Error
}

// DeleteUser deletes a user by ID using GORM
func DeleteUser(db *gorm.DB, userID uuid.UUID) error {
	return db.Where("id = ?", userID).Delete(&User{}).Error
}

// ListUsers retrieves all users with optional filtering using GORM
func ListUsers(db *gorm.DB, limit, offset int, active *bool) ([]*User, error) {
	var users []*User
	query := db

	if active != nil {
		query = query.Where("active = ?", *active)
	}

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// CountUsers counts the total number of users with optional filtering
func CountUsers(db *gorm.DB, active *bool) (int64, error) {
	var count int64
	query := db.Model(&User{})

	if active != nil {
		query = query.Where("active = ?", *active)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// ActivateUser activates a user account
func ActivateUser(db *gorm.DB, userID uuid.UUID) error {
	return db.Model(&User{}).Where("id = ?", userID).Update("active", true).Error
}

// DeactivateUser deactivates a user account
func DeactivateUser(db *gorm.DB, userID uuid.UUID) error {
	return db.Model(&User{}).Where("id = ?", userID).Update("active", false).Error
}

// UpdateUserPassword updates a user's password
func UpdateUserPassword(db *gorm.DB, userID uuid.UUID, hashedPassword string) error {
	return db.Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// SearchUsers searches for users by query string
func SearchUsers(db *gorm.DB, query string) ([]*User, error) {
	var users []*User
	searchQuery := "%" + query + "%"

	if err := db.Where("username ILIKE ? OR email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
		searchQuery, searchQuery, searchQuery, searchQuery).
		Limit(50).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
