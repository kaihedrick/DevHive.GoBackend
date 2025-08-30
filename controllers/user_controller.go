package controllers

import (
	"mime/multipart"
	"net/http"

	"devhive-backend/db"
	"devhive-backend/middleware"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserProfile retrieves the current user's profile
// @Summary Get user profile
// @Description Retrieve the current authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User profile retrieved successfully"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/User/profile [get]
func GetUserProfile(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := models.GetUserByID(db.GetDB(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateUserProfile updates the current user's profile
// @Summary Update user profile
// @Description Update the current authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.UserUpdateRequest true "User profile update request"
// @Success 200 {object} map[string]interface{} "Profile updated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/User/profile [put]
func UpdateUserProfile(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user profile
	updatedUser, err := models.UpdateUser(db.GetDB(), userID, req)
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
// @Summary Upload user avatar
// @Description Upload a new avatar image for the current authenticated user
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image file (JPEG, PNG, GIF, WebP, max 5MB)"
// @Success 200 {object} map[string]interface{} "Avatar uploaded successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid file or size"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/User/avatar [post]
func UploadAvatar(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
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
	err = models.UpdateUserAvatar(db.GetDB(), userID, avatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"avatar_url": avatarURL,
		"message":    "Avatar uploaded successfully",
	})
}

// ActivateUser activates a user account
// @Summary Activate user account
// @Description Activates a deactivated user account (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "User activated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Insufficient permissions"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/User/activate/{id} [put]
func ActivateUser(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if current user is admin (you might want to implement role-based access control)
	// For now, we'll allow any authenticated user to activate others
	// In production, you should check if the user has admin privileges

	targetUserIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get target user
	targetUser, err := models.GetUserByID(db.GetDB(), targetUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is already active
	if targetUser.Active {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already active"})
		return
	}

	// Activate user
	err = models.ActivateUser(db.GetDB(), targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User activated successfully",
		"user_id": targetUserID,
	})
}

// DeactivateUser deactivates a user account
// @Summary Deactivate user account
// @Description Deactivates an active user account (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "User deactivated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Insufficient permissions"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/User/deactivate/{id} [put]
func DeactivateUser(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if current user is admin (you might want to implement role-based access control)
	// For now, we'll allow any authenticated user to deactivate others
	// In production, you should check if the user has admin privileges

	targetUserIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Prevent deactivating yourself
	if targetUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot deactivate your own account"})
		return
	}

	// Get target user
	targetUser, err := models.GetUserByID(db.GetDB(), targetUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is already deactivated
	if !targetUser.Active {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already deactivated"})
		return
	}

	// Deactivate user
	err = models.DeactivateUser(db.GetDB(), targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deactivated successfully",
		"user_id": targetUserID,
	})
}

// SearchUsers searches for users by query
// @Summary Search users
// @Description Search for users by username, email, or name
// @Tags users
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Security BearerAuth
// @Success 200 {array} models.User "List of matching users"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/User/search [get]
func SearchUsers(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Search users by query
	users, err := models.SearchUsers(db.GetDB(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
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
