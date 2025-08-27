package controllers

import (
	"database/sql"
	"net/http"
	"time"

	"devhive-backend/config"
	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthRequest represents the authentication request
type AuthRequest struct {
	FirebaseIDToken string `json:"firebase_id_token" binding:"required"`
}

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	FirebaseIDToken string `json:"firebase_id_token" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Username        string `json:"username" binding:"required,min=3,max=100"`
	FirstName       string `json:"first_name" binding:"required,min=1,max=100"`
	LastName        string `json:"last_name" binding:"required,min=1,max=100"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Register handles user registration
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Firebase ID token
	firebaseUID, err := config.VerifyFirebaseToken(req.FirebaseIDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Firebase token"})
		return
	}

	// Check if user already exists
	existingUser, err := models.GetUserByFirebaseUID(db.GetDB(), firebaseUID)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Check if username is already taken
	existingUsername, err := models.GetUserByEmail(db.GetDB(), req.Email)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if existingUsername != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Create user
	userReq := models.UserCreateRequest{
		FirebaseUID: firebaseUID,
		Email:       req.Email,
		Username:    req.Username,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
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
func Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Firebase ID token
	firebaseUID, err := config.VerifyFirebaseToken(req.FirebaseIDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Firebase token"})
		return
	}

	// Get user from database
	user, err := models.GetUserByFirebaseUID(db.GetDB(), firebaseUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
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
