package handlers

import (
	"net/http"
	"strings"

	"devhive-backend/internal/http/response"

	"github.com/jackc/pgx/v5"
)

// ValidateEmailRequest represents the email validation request
type ValidateEmailRequest struct {
	Email string `json:"email"`
}

// ValidateUsernameRequest represents the username validation request
type ValidateUsernameRequest struct {
	Username string `json:"username"`
}

// ValidateResponse represents the validation response
type ValidateResponse struct {
	Available bool `json:"available"`
}

// ValidateEmail handles email availability validation
func (h *UserHandler) ValidateEmail(w http.ResponseWriter, r *http.Request) {
	var email string

	// Support both GET (query param) and POST (JSON body) requests
	if r.Method == "GET" {
		email = strings.TrimSpace(strings.ToLower(r.URL.Query().Get("email")))
	} else {
		var req ValidateEmailRequest
		if !response.Decode(w, r, &req) {
			return
		}
		email = strings.TrimSpace(strings.ToLower(req.Email))
	}

	if email == "" {
		response.BadRequest(w, "email is required")
		return
	}

	// Validate email format (basic validation for live typing)
	if !isValidEmailForLiveValidation(email) {
		response.BadRequest(w, "invalid email format")
		return
	}

	_, err := h.queries.GetUserByEmail(r.Context(), email)
	available := false
	if err != nil {
		if err == pgx.ErrNoRows {
			available = true
		} else {
			response.InternalServerError(w, "database error")
			return
		}
	}

	response.JSON(w, http.StatusOK, ValidateResponse{
		Available: available,
	})
}

// ValidateUsername handles username availability validation
func (h *UserHandler) ValidateUsername(w http.ResponseWriter, r *http.Request) {
	var username string

	// Support both GET (query param) and POST (JSON body) requests
	if r.Method == "GET" {
		username = strings.TrimSpace(strings.ToLower(r.URL.Query().Get("username")))
	} else {
		var req ValidateUsernameRequest
		if !response.Decode(w, r, &req) {
			return
		}
		username = strings.TrimSpace(strings.ToLower(req.Username))
	}

	if username == "" {
		response.BadRequest(w, "username is required")
		return
	}

	// Validate username format (basic validation for live typing)
	if !isValidUsernameForLiveValidation(username) {
		response.BadRequest(w, "invalid username format")
		return
	}

	_, err := h.queries.GetUserByUsername(r.Context(), username)
	available := false
	if err != nil {
		if err == pgx.ErrNoRows {
			available = true
		} else {
			response.InternalServerError(w, "database error")
			return
		}
	}

	response.JSON(w, http.StatusOK, ValidateResponse{
		Available: available,
	})
}

// isValidEmailForLiveValidation performs lenient email validation for live typing
func isValidEmailForLiveValidation(email string) bool {
	// Allow empty or very short strings during typing
	if len(email) == 0 {
		return true
	}

	// Reject obviously invalid characters
	if strings.ContainsAny(email, " \t\n\r") {
		return false
	}

	// If it has an @, do basic validation
	if strings.Contains(email, "@") {
		atCount := strings.Count(email, "@")
		if atCount > 1 {
			return false
		}

		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return false
		}

		local, domain := parts[0], parts[1]
		if len(local) == 0 {
			return false
		}

		// Allow incomplete domains during typing
		// Only reject if domain has invalid characters
		if len(domain) > 0 && strings.ContainsAny(domain, " \t\n\r") {
			return false
		}
	}

	return true
}

// isValidEmail performs strict email validation for final submission
func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 254 {
		return false
	}

	atCount := strings.Count(email, "@")
	if atCount != 1 {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]
	if len(local) == 0 || len(domain) == 0 {
		return false
	}

	// Check for basic domain format
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

// isValidUsernameForLiveValidation performs lenient username validation for live typing
func isValidUsernameForLiveValidation(username string) bool {
	// Allow empty strings during typing
	if len(username) == 0 {
		return true
	}

	// Reject obviously invalid characters
	if strings.ContainsAny(username, " \t\n\r") {
		return false
	}

	// Allow alphanumeric characters, underscores, and hyphens
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// isValidUsername performs strict username validation for final submission
func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// Allow alphanumeric characters, underscores, and hyphens
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	// Must start and end with alphanumeric character
	if (username[0] < 'a' || username[0] > 'z') && (username[0] < '0' || username[0] > '9') {
		return false
	}
	if (username[len(username)-1] < 'a' || username[len(username)-1] > 'z') && (username[len(username)-1] < '0' || username[len(username)-1] > '9') {
		return false
	}

	return true
}
