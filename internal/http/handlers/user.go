package handlers

import (
	"net/http"

	"devhive-backend/internal/http/middleware"
	"devhive-backend/internal/http/response"
	"devhive-backend/internal/repo"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	queries *repo.Queries
}

func NewUserHandler(queries *repo.Queries) *UserHandler {
	return &UserHandler{
		queries: queries,
	}
}

// CreateUserRequest represents the user creation request
type CreateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Active    bool   `json:"active"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CreateUser handles user creation
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalServerError(w, "Failed to hash password")
		return
	}

	// Create user
	user, err := h.queries.CreateUser(r.Context(), repo.CreateUserParams{
		Username:  req.Username,
		Email:     req.Email,
		PasswordH: string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		response.BadRequest(w, "Failed to create user: "+err.Error())
		return
	}

	avatarURL := ""
	if user.AvatarUrl != nil {
		avatarURL = *user.AvatarUrl
	}

	response.JSON(w, http.StatusCreated, UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		AvatarURL: avatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateUserRequest represents a partial user update request (all fields optional)
type UpdateUserRequest struct {
	Username  *string `json:"username,omitempty"`
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}

// GetMe handles getting current user
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
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
	user, err := h.queries.GetUserByID(r.Context(), userUUID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	avatarURL := ""
	if user.AvatarUrl != nil {
		avatarURL = *user.AvatarUrl
	}

	response.JSON(w, http.StatusOK, UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		AvatarURL: avatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateMe handles updating the current user's profile (PATCH /users/me)
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
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

	// Get current user to merge with updates
	currentUser, err := h.queries.GetUserByID(r.Context(), userUUID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	// Decode request body (partial update - all fields optional)
	var req UpdateUserRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// Merge provided fields with existing values
	username := currentUser.Username
	if req.Username != nil {
		username = *req.Username
	}

	email := currentUser.Email
	if req.Email != nil {
		email = *req.Email
	}

	firstName := currentUser.FirstName
	if req.FirstName != nil {
		firstName = *req.FirstName
	}

	lastName := currentUser.LastName
	if req.LastName != nil {
		lastName = *req.LastName
	}

	var avatarURL *string = currentUser.AvatarUrl
	if req.AvatarURL != nil {
		if *req.AvatarURL == "" {
			avatarURL = nil // Allow clearing avatar by sending empty string
		} else {
			avatarURL = req.AvatarURL
		}
	}

	// Update user
	updatedUser, err := h.queries.UpdateUser(r.Context(), repo.UpdateUserParams{
		ID:        userUUID,
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		AvatarUrl: avatarURL,
	})
	if err != nil {
		response.BadRequest(w, "Failed to update user: "+err.Error())
		return
	}

	avatarURLStr := ""
	if updatedUser.AvatarUrl != nil {
		avatarURLStr = *updatedUser.AvatarUrl
	}

	response.JSON(w, http.StatusOK, UserResponse{
		ID:        updatedUser.ID.String(),
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		FirstName: updatedUser.FirstName,
		LastName:  updatedUser.LastName,
		Active:    updatedUser.Active,
		AvatarURL: avatarURLStr,
		CreatedAt: updatedUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetUser handles getting a user by ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), userUUID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	avatarURL := ""
	if user.AvatarUrl != nil {
		avatarURL = *user.AvatarUrl
	}

	response.JSON(w, http.StatusOK, UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		AvatarURL: avatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
