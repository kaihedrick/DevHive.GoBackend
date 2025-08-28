package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"devhive-backend/models"
	"devhive-backend/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req models.UserCreateRequest) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (*jwt.Token, error)
	ResetPassword(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
}

// authService implements AuthService
type authService struct {
	userRepo   repositories.UserRepository
	jwtSecret  string
}

// NewAuthService creates a new auth service instance
func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: "your-jwt-secret-here", // TODO: Get from config
	}
}

// Register creates a new user account
func (s *authService) Register(ctx context.Context, req models.UserCreateRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Active:    true,
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Clear password from response
	user.Password = ""

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.Active {
		return "", fmt.Errorf("account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateJWT(user.ID, user.Email)
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	return token, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Parse and validate refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID in token")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID format: %w", err)
	}

	// Verify user still exists
	_, err = s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Generate new access token
	newToken, err := s.generateJWT(userUUID, claims["email"].(string))
	if err != nil {
		return "", fmt.Errorf("error generating new token: %w", err)
	}

	return newToken, nil
}

// ValidateToken validates a JWT token and returns the parsed token
func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

// ResetPassword initiates a password reset process
func (s *authService) ResetPassword(ctx context.Context, email string) error {
	// Get user by email
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// TODO: Store reset token in database with expiration
	// TODO: Send email with reset link

	return nil
}

// ConfirmPasswordReset confirms a password reset with the provided token
func (s *authService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	// TODO: Validate reset token from database
	// TODO: Update user password
	// TODO: Mark reset token as used

	return fmt.Errorf("password reset confirmation not implemented")
}

// generateJWT generates a JWT token for a user
func (s *authService) generateJWT(userID uuid.UUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		"iat":     time.Now().Unix(),
		"iss":     "devhive-backend",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateResetToken generates a secure random token for password reset
func (s *authService) generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
