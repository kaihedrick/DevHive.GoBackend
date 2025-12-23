package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"devhive-backend/internal/config"
	"devhive-backend/internal/http/middleware"
	"devhive-backend/internal/http/response"
	"devhive-backend/internal/repo"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
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

	// Check password (user.PasswordH is *string for nullable field)
	if user.PasswordH == nil {
		response.Unauthorized(w, "Invalid credentials - OAuth user")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordH), []byte(req.Password)); err != nil {
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Check if user is active
	if !user.Active {
		response.Unauthorized(w, "Account is deactivated")
		return
	}

	// Generate access token (short-lived)
	accessToken, err := h.generateJWT(user.ID.String())
	if err != nil {
		response.InternalServerError(w, "Failed to generate token")
		return
	}

	// Generate refresh token with appropriate expiry based on rememberMe
	refreshToken := generateRandomToken(64)
	var refreshExpiresAt time.Time
	var cookieMaxAge int

	if req.RememberMe {
		// Persistent login: 30 days
		refreshExpiresAt = time.Now().Add(h.cfg.JWT.RefreshTokenPersistentExpiration)
		cookieMaxAge = int(h.cfg.JWT.RefreshTokenPersistentExpiration.Seconds())
	} else {
		// Session-only: cookie expires when browser closes (MaxAge = 0)
		// But store in DB with 7-day expiry as backup
		refreshExpiresAt = time.Now().Add(h.cfg.JWT.RefreshTokenExpiration)
		cookieMaxAge = 0 // Session cookie
	}

	// Store refresh token in database
	_, err = h.queries.CreateRefreshToken(r.Context(), repo.CreateRefreshTokenParams{
		UserID:       user.ID,
		Token:        refreshToken,
		ExpiresAt:    refreshExpiresAt,
		IsPersistent: req.RememberMe,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to create refresh token")
		return
	}

	// Set refresh token as HttpOnly cookie with appropriate MaxAge
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   true, // Always true in production (HTTPS required for SameSite=None)
		SameSite: http.SameSiteNoneMode, // NoneMode required for cross-origin requests
	})

	response.JSON(w, http.StatusOK, LoginResponse{
		Token:  accessToken,
		UserID: user.ID.String(),
	})
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		response.Unauthorized(w, "Refresh token not found")
		return
	}

	refreshToken := cookie.Value

	// Get refresh token from database
	tokenRecord, err := h.queries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		response.Unauthorized(w, "Invalid refresh token")
		return
	}

	// Check if token is expired
	if time.Now().After(tokenRecord.ExpiresAt) {
		// Delete expired token
		_ = h.queries.DeleteRefreshToken(r.Context(), refreshToken)
		response.Unauthorized(w, "Refresh token has expired")
		return
	}

	// Verify user still exists and is active
	user, err := h.queries.GetUserByID(r.Context(), tokenRecord.UserID)
	if err != nil {
		response.Unauthorized(w, "User not found")
		return
	}

	if !user.Active {
		// Delete all refresh tokens for inactive user
		_ = h.queries.DeleteUserRefreshTokens(r.Context(), user.ID)
		response.Unauthorized(w, "Account is deactivated")
		return
	}

	// Generate new access token
	accessToken, err := h.generateJWT(user.ID.String())
	if err != nil {
		response.InternalServerError(w, "Failed to generate token")
		return
	}

	// Rotate refresh token (security best practice)
	// Preserve the is_persistent flag from the old token
	isPersistent := tokenRecord.IsPersistent
	var newRefreshExpiresAt time.Time
	var cookieMaxAge int

	if isPersistent {
		// Persistent login: extend by 30 days from now
		newRefreshExpiresAt = time.Now().Add(h.cfg.JWT.RefreshTokenPersistentExpiration)
		cookieMaxAge = int(h.cfg.JWT.RefreshTokenPersistentExpiration.Seconds())
	} else {
		// Session-only: keep session cookie behavior
		newRefreshExpiresAt = time.Now().Add(h.cfg.JWT.RefreshTokenExpiration)
		cookieMaxAge = 0 // Session cookie
	}

	// Delete old refresh token and create new one
	_ = h.queries.DeleteRefreshToken(r.Context(), refreshToken)
	newRefreshToken := generateRandomToken(64)
	_, err = h.queries.CreateRefreshToken(r.Context(), repo.CreateRefreshTokenParams{
		UserID:       user.ID,
		Token:        newRefreshToken,
		ExpiresAt:    newRefreshExpiresAt,
		IsPersistent: isPersistent,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to create refresh token")
		return
	}

	// Set new refresh token cookie with appropriate MaxAge
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   true, // Always true in production (HTTPS required for SameSite=None)
		SameSite: http.SameSiteNoneMode, // NoneMode required for cross-origin requests
	})

	response.JSON(w, http.StatusOK, LoginResponse{
		Token:  accessToken,
		UserID: user.ID.String(),
	})
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
	passwordStr := string(hashedPassword)
	err = h.queries.UpdateUserPassword(r.Context(), repo.UpdateUserPasswordParams{
		ID:        reset.UserID,
		PasswordH: &passwordStr,
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

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		// Delete refresh token from database
		_ = h.queries.DeleteRefreshToken(r.Context(), cookie.Value)
	}

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   true, // Always true in production (HTTPS required for SameSite=None)
		SameSite: http.SameSiteNoneMode, // NoneMode required for cross-origin requests
	})

	response.JSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// VerifyPassword handles admin password verification
func (h *AuthHandler) VerifyPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if !response.Decode(w, r, &req) {
		return
	}

	// Get admin password from config (fallback to environment variable)
	adminPassword := h.cfg.AdminPassword
	if adminPassword == "" {
		adminPassword = "jtAppmine2021" // Default fallback
	}

	if req.Password == adminPassword {
		// Set cookie that expires in 30 days
		maxAge := 30 * 24 * 60 * 60 // 30 days in seconds
		http.SetCookie(w, &http.Cookie{
			Name:     "admin_certificates_verified",
			Value:    "true",
			Path:     "/",
			MaxAge:   maxAge,
			HttpOnly: true,
			Secure:   r.TLS != nil, // Secure in production (HTTPS)
			SameSite: http.SameSiteNoneMode, // NoneMode for cross-origin
		})
		response.JSON(w, http.StatusOK, map[string]bool{"success": true})
		return
	}

	response.JSON(w, http.StatusUnauthorized, map[string]interface{}{
		"success": false,
		"message": "Invalid password",
	})
}

// CheckAuth checks if admin is authenticated
func (h *AuthHandler) CheckAuth(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("admin_certificates_verified")
	if err != nil || cookie.Value != "true" {
		response.JSON(w, http.StatusUnauthorized, map[string]bool{"authenticated": false})
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"authenticated": true})
}

// ChangePasswordRequest represents a password change request (authenticated user)
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangePassword handles authenticated user password changes
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by RequireAuth middleware)
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	// Decode request
	var req ChangePasswordRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Validate request
	if req.CurrentPassword == "" {
		response.BadRequest(w, "Current password is required")
		return
	}
	if req.NewPassword == "" {
		response.BadRequest(w, "New password is required")
		return
	}
	if len(req.NewPassword) < 8 {
		response.BadRequest(w, "New password must be at least 8 characters")
		return
	}

	// Get user to verify current password (with password hash)
	user, err := h.queries.GetUserByIDWithPassword(r.Context(), userUUID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	// Verify current password (user.PasswordH is *string)
	if user.PasswordH == nil {
		response.Unauthorized(w, "Cannot change password for OAuth user")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordH), []byte(req.CurrentPassword)); err != nil {
		response.Unauthorized(w, "Current password is incorrect")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.InternalServerError(w, "Failed to hash password")
		return
	}

	// Update password
	passwordStr := string(hashedPassword)
	err = h.queries.UpdateUserPassword(r.Context(), repo.UpdateUserPasswordParams{
		ID:        userUUID,
		PasswordH: &passwordStr,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to update password")
		return
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

// ValidateToken handles token validation requests
// Returns token validity and expiration info without requiring refresh
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header or query param
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}
	if tokenString == "" {
		tokenString = r.URL.Query().Get("token")
	}

	if tokenString == "" {
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"valid": false,
			"error": "No token provided",
		})
		return
	}

	// Parse token without validation to check expiration
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.cfg.JWT.SigningKey), nil
	}, jwt.WithoutClaimsValidation())

	var isValid bool
	var expiresAt *time.Time
	var errorMsg string

	if err != nil {
		isValid = false
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "expired") || strings.Contains(errStr, "exp") {
			errorMsg = "Token has expired"
		} else {
			errorMsg = "Invalid token"
		}
	} else if token != nil {
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			// Check expiration
			if exp, ok := claims["exp"].(float64); ok {
				expTime := time.Unix(int64(exp), 0)
				expiresAt = &expTime
				isValid = time.Now().Before(expTime)
			} else {
				isValid = false
				errorMsg = "Token missing expiration claim"
			}
		} else {
			isValid = false
			errorMsg = "Invalid token claims"
		}
	} else {
		isValid = false
		errorMsg = "Token parsing failed"
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"valid":     isValid,
		"expiresAt": expiresAt,
		"error":     errorMsg,
	})
}

// Google OAuth Handlers

// GoogleLoginRequest represents the query parameters for initiating Google OAuth
type GoogleLoginRequest struct {
	RememberMe  bool   `json:"rememberMe"`
	RedirectURL string `json:"redirectUrl"`
}

// GoogleUserInfo represents the user information returned by Google
type GoogleUserInfo struct {
	Sub           string `json:"sub"`            // Google user ID
	Email         string `json:"email"`          // User email
	EmailVerified bool   `json:"email_verified"` // Email verification status
	Name          string `json:"name"`           // Full name
	GivenName     string `json:"given_name"`     // First name
	FamilyName    string `json:"family_name"`    // Last name
	Picture       string `json:"picture"`        // Profile picture URL
}

// GoogleLogin initiates the Google OAuth flow
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	rememberMeStr := r.URL.Query().Get("remember_me")
	redirectURL := r.URL.Query().Get("redirect")

	rememberMe := rememberMeStr == "true"

	// Generate random state token for CSRF protection
	stateToken := generateRandomToken(32)

	// Store state token in database with remember_me preference
	expiresAt := time.Now().Add(10 * time.Minute)
	_, err := h.queries.CreateOAuthState(r.Context(), repo.CreateOAuthStateParams{
		StateToken:  stateToken,
		RememberMe:  rememberMe,
		RedirectUrl: &redirectURL,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to create OAuth state")
		return
	}

	// Build Google OAuth config
	oauthConfig := &oauth2.Config{
		ClientID:     h.cfg.GoogleOAuth.ClientID,
		ClientSecret: h.cfg.GoogleOAuth.ClientSecret,
		RedirectURL:  h.cfg.GoogleOAuth.RedirectURL,
		Scopes:       h.cfg.GoogleOAuth.Scopes,
		Endpoint:     google.Endpoint,
	}

	// Generate authorization URL
	authURL := oauthConfig.AuthCodeURL(stateToken, oauth2.AccessTypeOffline)

	response.JSON(w, http.StatusOK, map[string]string{
		"authUrl": authURL,
		"state":   stateToken,
	})
}

// GoogleCallback handles the OAuth callback from Google
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Get authorization code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		response.BadRequest(w, "Missing code or state parameter")
		return
	}

	// Validate state token (CSRF protection)
	stateRecord, err := h.queries.GetOAuthState(r.Context(), state)
	if err != nil {
		response.BadRequest(w, "Invalid or expired state token")
		return
	}

	// Get remember_me preference from state
	rememberMe := stateRecord.RememberMe

	// Build OAuth config
	oauthConfig := &oauth2.Config{
		ClientID:     h.cfg.GoogleOAuth.ClientID,
		ClientSecret: h.cfg.GoogleOAuth.ClientSecret,
		RedirectURL:  h.cfg.GoogleOAuth.RedirectURL,
		Scopes:       h.cfg.GoogleOAuth.Scopes,
		Endpoint:     google.Endpoint,
	}

	// Exchange authorization code for tokens
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		response.InternalServerError(w, "Failed to exchange code for token")
		return
	}

	// Fetch user info from Google
	userInfo, err := h.fetchGoogleUserInfo(token.AccessToken)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch user info from Google")
		return
	}

	// Check if user exists by Google ID
	user, err := h.queries.GetUserByGoogleID(r.Context(), &userInfo.Sub)
	var userID uuid.UUID
	isNewUser := false

	if err != nil {
		// User doesn't exist, create new user
		isNewUser = true

		// Generate unique username from email
		username := h.generateUsernameFromEmail(userInfo.Email)

		// Create OAuth user
		authProvider := "google"
		newUser, err := h.queries.CreateOAuthUser(r.Context(), repo.CreateOAuthUserParams{
			Username:          username,
			Email:             userInfo.Email,
			FirstName:         userInfo.GivenName,
			LastName:          userInfo.FamilyName,
			AuthProvider:      &authProvider,
			GoogleID:          &userInfo.Sub,
			ProfilePictureUrl: &userInfo.Picture,
		})
		if err != nil {
			response.InternalServerError(w, "Failed to create user")
			return
		}

		userID = newUser.ID
	} else {
		// User exists, update profile picture if changed
		if user.ProfilePictureUrl == nil || *user.ProfilePictureUrl != userInfo.Picture {
			err = h.queries.UpdateUserProfilePicture(r.Context(), repo.UpdateUserProfilePictureParams{
				ID:                user.ID,
				ProfilePictureUrl: &userInfo.Picture,
			})
			if err != nil {
				// Log error but don't fail the request
			}
		}

		userID = user.ID
	}

	// Generate DevHive access token
	accessToken, err := h.generateJWT(userID.String())
	if err != nil {
		response.InternalServerError(w, "Failed to generate access token")
		return
	}

	// Generate DevHive refresh token with appropriate expiry
	refreshToken := generateRandomToken(64)
	var refreshExpiresAt time.Time
	var cookieMaxAge int

	if rememberMe {
		// Persistent login: 30 days
		refreshExpiresAt = time.Now().Add(h.cfg.JWT.RefreshTokenPersistentExpiration)
		cookieMaxAge = int(h.cfg.JWT.RefreshTokenPersistentExpiration.Seconds())
	} else {
		// Session-only: cookie expires when browser closes (MaxAge = 0)
		// But store in DB with 7-day expiry as backup
		refreshExpiresAt = time.Now().Add(7 * 24 * time.Hour)
		cookieMaxAge = 0 // Session cookie
	}

	// Store refresh token with Google tokens
	var googleRefreshToken *string
	if token.RefreshToken != "" {
		googleRefreshToken = &token.RefreshToken
	}

	googleAccessToken := token.AccessToken

	// Convert time.Time to pgtype.Timestamptz
	var googleTokenExpiry pgtype.Timestamptz
	if err := googleTokenExpiry.Scan(token.Expiry); err != nil {
		response.InternalServerError(w, "Failed to process token expiry")
		return
	}

	_, err = h.queries.CreateRefreshTokenWithGoogle(r.Context(), repo.CreateRefreshTokenWithGoogleParams{
		UserID:             userID,
		Token:              refreshToken,
		ExpiresAt:          refreshExpiresAt,
		IsPersistent:       rememberMe,
		GoogleRefreshToken: googleRefreshToken,
		GoogleAccessToken:  &googleAccessToken,
		GoogleTokenExpiry:  googleTokenExpiry,
	})
	if err != nil {
		response.InternalServerError(w, "Failed to create refresh token")
		return
	}

	// Set HttpOnly cookie with appropriate MaxAge
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	// Delete state token from database
	_ = h.queries.DeleteOAuthState(r.Context(), state)

	// Return access token and user info
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"token":     accessToken,
		"userId":    userID.String(),
		"isNewUser": isNewUser,
		"user": map[string]interface{}{
			"id":             userID.String(),
			"email":          userInfo.Email,
			"firstName":      userInfo.GivenName,
			"lastName":       userInfo.FamilyName,
			"profilePicture": userInfo.Picture,
		},
	})
}

// fetchGoogleUserInfo fetches user information from Google using the access token
func (h *AuthHandler) fetchGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	// Call Google UserInfo API
	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user info: status %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// generateUsernameFromEmail generates a username from an email address
// Handles conflicts by appending random numbers if needed
func (h *AuthHandler) generateUsernameFromEmail(email string) string {
	// Extract username part from email (before @)
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return "user_" + generateRandomToken(8)
	}

	baseUsername := parts[0]
	// Replace non-alphanumeric characters with underscores
	baseUsername = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, baseUsername)

	// Try base username first
	username := baseUsername

	// If username exists, append random numbers until we find a unique one
	// In practice, the DB will reject duplicates and we'll need to retry
	// For now, append a random suffix
	username = username + "_" + generateRandomToken(4)

	return username
}

