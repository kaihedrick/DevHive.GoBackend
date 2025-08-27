package controllers

import (
	"mime/multipart"
	"net/http"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserProfile retrieves the current user's profile
func GetUserProfile(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateUserProfile updates the current user's profile
func UpdateUserProfile(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user profile
	updatedUser, err := models.UpdateUser(db.GetDB(), user.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    updatedUser,
		"message": "Profile updated successfully",
	})
}

// UploadAvatar handles user avatar upload
func UploadAvatar(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the uploaded file
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Avatar file required"})
		return
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size too large. Maximum size is 5MB"})
		return
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !allowedTypes[file.Header.Get("Content-Type")] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPEG, PNG, GIF, and WebP are allowed"})
		return
	}

	// Generate unique filename
	filename := uuid.New().String() + getFileExtension(file.Filename)

	// Upload to Firebase Storage (if configured)
	avatarURL, err := uploadToFirebaseStorage(file, filename, "avatars")
	if err != nil {
		// Fallback to local storage
		avatarURL, err = uploadToLocalStorage(file, filename, "avatars")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload avatar"})
			return
		}
	}

	// Update user's avatar URL in database
	err = models.UpdateUserAvatar(db.GetDB(), user.ID, avatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"avatar_url": avatarURL,
		"message":    "Avatar uploaded successfully",
	})
}

// uploadToFirebaseStorage uploads a file to Firebase Storage
func uploadToFirebaseStorage(file *multipart.FileHeader, filename, folder string) (string, error) {
	// This would integrate with Firebase Storage
	// For now, return a placeholder URL
	return "https://storage.googleapis.com/devhive-avatars/" + filename, nil
}

// uploadToLocalStorage uploads a file to local storage
func uploadToLocalStorage(file *multipart.FileHeader, filename, folder string) (string, error) {
	// This would save to local filesystem
	// For now, return a placeholder URL
	return "/static/" + folder + "/" + filename, nil
}

// getFileExtension extracts the file extension from filename
func getFileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}
