package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PasswordReset represents a password reset request
type PasswordReset struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// PasswordResetRequest represents a request to reset a password
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// PasswordResetConfirm represents a password reset confirmation
type PasswordResetConfirm struct {
	Token           string `json:"token" binding:"required" example:"reset-token-here"`
	NewPassword     string `json:"newPassword" binding:"required,min=8" example:"newSecurePassword123"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=NewPassword" example:"newSecurePassword123"`
}

// TableName returns the table name for PasswordReset
func (PasswordReset) TableName() string {
	return "password_resets"
}

// CreatePasswordResetToken creates a new password reset token
func CreatePasswordResetToken(db *gorm.DB, userID uuid.UUID, token string, expiresAt time.Time) error {
	passwordReset := &PasswordReset{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return db.Create(passwordReset).Error
}

// GetPasswordResetToken retrieves a password reset token by token string
func GetPasswordResetToken(db *gorm.DB, token string) (*PasswordReset, error) {
	var passwordReset PasswordReset
	if err := db.Where("token = ? AND used = ?", token, false).First(&passwordReset).Error; err != nil {
		return nil, err
	}
	return &passwordReset, nil
}

// DeletePasswordResetToken deletes a password reset token
func DeletePasswordResetToken(db *gorm.DB, token string) error {
	return db.Where("token = ?", token).Delete(&PasswordReset{}).Error
}
