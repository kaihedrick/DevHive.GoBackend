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
