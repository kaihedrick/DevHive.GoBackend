package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"devhive-backend/internal/config"
	"devhive-backend/internal/http/response"
	"devhive-backend/internal/repo"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	cfg     *config.Config
	queries *repo.Queries
}

func NewAuthHandler(cfg *config.Config, queries *repo.Queries) *AuthHandler {
	return &AuthHandler{
		cfg:     cfg,
		queries: queries,
	}
}

// LoginRequest represents the login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"userId"`
}

// PasswordResetRequest represents the password reset request
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordReset represents the password reset
type PasswordReset struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Get user by username
	user, err := h.queries.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordH), []byte(req.Password)); err != nil {
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Check if user is active
	if !user.Active {
		response.Unauthorized(w, "Account is deactivated")
		return
	}

	// Generate JWT token
	token, err := h.generateJWT(user.ID.String())
	if err != nil {
		response.InternalServerError(w, "Failed to generate token")
		return
	}

	response.JSON(w, http.StatusOK, LoginResponse{
		Token:  token,
		UserID: user.ID.String(),
	})
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// For now, just return an error as refresh is optional
	response.BadRequest(w, "Token refresh not implemented")
}

// RequestPasswordReset handles password reset requests
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Get user by email
	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		// Don't reveal if email exists or not
		response.JSON(w, http.StatusOK, map[string]string{"message": "If the email exists, a reset link has been sent"})
		return
	}

	// Generate reset token
	token := generateRandomToken(32)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Store reset token
	_, err = h.queries.CreatePasswordReset(r.Context(), repo.CreatePasswordResetParams{
		UserID:     user.ID,
		ResetToken: token,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to create reset token")
		return
	}

	// TODO: Send email with reset link
	// For now, just return the token (remove this in production)
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Reset token created",
		"token":   token, // Remove this in production
	})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req PasswordReset
	if !response.Decode(w, r, &req) {
		return
	}

	// Get reset token
	reset, err := h.queries.GetPasswordResetByToken(r.Context(), req.Token)
	if err != nil {
		response.BadRequest(w, "Invalid or expired reset token")
		return
	}

	// Check if token is expired
	if time.Now().After(reset.ExpiresAt) {
		response.BadRequest(w, "Reset token has expired")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalServerError(w, "Failed to hash password")
		return
	}

	// Update user password
	err = h.queries.UpdateUserPassword(r.Context(), repo.UpdateUserPasswordParams{
		ID:        reset.UserID,
		PasswordH: string(hashedPassword),
	})
	if err != nil {
		response.InternalServerError(w, "Failed to update password")
		return
	}

	// Delete reset token
	if err := h.queries.DeletePasswordReset(r.Context(), req.Token); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

// generateJWT generates a JWT token for the user
func (h *AuthHandler) generateJWT(userID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(h.cfg.JWT.Expiration).Unix(),
		"iss": h.cfg.JWT.Issuer,
		"aud": h.cfg.JWT.Audience,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWT.SigningKey))
}

// generateRandomToken generates a random token
func generateRandomToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
