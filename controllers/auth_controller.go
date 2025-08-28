package controllers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"devhive-backend/config"
	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthRequest represents the authentication request
// @Description Authentication request for login
type AuthRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// RegisterRequest represents the user registration request
// @Description User registration request
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=30" example:"johndoe"`
	Password  string `json:"password" binding:"required,min=6,max=100" example:"password123"`
	Email     string `json:"email" binding:"required,email" example:"john@example.com"`
	FirstName string `json:"first_name" binding:"required,min=1,max=50" example:"John"`
	LastName  string `json:"last_name" binding:"required,min=1,max=50" example:"Doe"`
}

// AuthResponse represents the authentication response
// @Description Authentication response with tokens
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int64        `json:"expires_in" example:"3600"`
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ForgotPasswordRequest represents the forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

// ResetPasswordRequest represents the reset password request
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required" example:"reset_token_here"`
	Password string `json:"password" binding:"required,min=6,max=100" example:"newpassword123"`
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration request"
// @Success 201 {object} models.User "User created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 409 {object} map[string]string "Username or email already taken"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	existingUser, err := models.GetUserByUsername(db.GetDB(), req.Username)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	}

	// Check if email is already registered
	existingEmail, err := models.GetUserByEmail(db.GetDB(), req.Email)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if existingEmail != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	userReq := models.UserCreateRequest{
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	user, err := models.CreateUser(db.GetDB(), userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := generateTokens(user.ID.String(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	response := AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
// @Summary Authenticate user
// @Description Authenticate user with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body AuthRequest true "User credentials"
// @Success 200 {object} AuthResponse "Authentication successful"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from database
	user, err := models.GetUserByUsername(db.GetDB(), req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user is active
	if !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is deactivated"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := generateTokens(user.ID.String(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	response := AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body object true "Refresh token request" schema(object{refresh_token=string})
// @Success 200 {object} AuthResponse "Token refreshed successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Invalid refresh token"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse and validate refresh token
	token, err := jwt.ParseWithClaims(req.RefreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Get user from database
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := models.GetUserByID(db.GetDB(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := generateTokens(user.ID.String(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	response := AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
	}

	c.JSON(http.StatusOK, response)
}

// ForgotPassword handles password reset request
// @Summary Request password reset
// @Description Sends a password reset email to the user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Password reset request"
// @Success 200 {object} map[string]string "Password reset email sent"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/forgot-password [post]
func ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from database
	user, err := models.GetUserByEmail(db.GetDB(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Generate password reset token
	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour

	// Save password reset token to database
	err = models.CreatePasswordResetToken(db.GetDB(), user.ID, token, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset token"})
		return
	}

	// Send password reset email
	resetLink := config.AppConfig.FrontendURL + "/reset-password?token=" + token
	err = sendPasswordResetEmail(user.Email, user.FirstName, resetLink)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent successfully",
	})
}

// ResetPassword handles password reset with token
// @Summary Reset password with token
// @Description Resets user password using the reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Password reset with token"
// @Success 200 {object} map[string]string "Password reset successful"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Invalid or expired token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/reset-password [post]
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and get password reset token
	resetToken, err := models.GetPasswordResetToken(db.GetDB(), req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Reset token has expired"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update user password
	err = models.UpdateUserPassword(db.GetDB(), resetToken.UserID, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Delete used reset token
	err = models.DeletePasswordResetToken(db.GetDB(), req.Token)
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to delete reset token: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// generateTokens generates JWT access and refresh tokens
func generateTokens(userID, email string) (string, string, error) {
	// Generate access token
	accessClaims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshClaims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// sendPasswordResetEmail sends a password reset email
func sendPasswordResetEmail(email, firstName, resetLink string) error {
	// This would integrate with your mail service
	// For now, just log the reset link
	log.Printf("Password reset email would be sent to %s with link: %s", email, resetLink)
	return nil
}
