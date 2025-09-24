package response

import (
	"encoding/json"
	"net/http"
)

// JSON sends a JSON response
func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Log error but don't send response as header already sent
		// TODO: Add proper logging
	}
}

// Problem represents an RFC 7807 problem details response
type Problem struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// Problemf sends an RFC 7807 problem details response
func Problemf(w http.ResponseWriter, status int, typ, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	problem := Problem{
		Type:   typ,
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	}

	if err := json.NewEncoder(w).Encode(problem); err != nil {
		// Log error but don't send response as header already sent
		// TODO: Add proper logging
	}
}

// Decode decodes JSON request body into destination
func Decode(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		Problemf(w, http.StatusBadRequest, "invalid_json", err.Error())
		return false
	}

	return true
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// BadRequest sends a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message string) {
	Problemf(w, http.StatusBadRequest, "bad_request", message)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	Problemf(w, http.StatusUnauthorized, "unauthorized", message)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(w http.ResponseWriter, message string) {
	Problemf(w, http.StatusForbidden, "forbidden", message)
}

// NotFound sends a 404 Not Found response
func NotFound(w http.ResponseWriter, message string) {
	Problemf(w, http.StatusNotFound, "not_found", message)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(w http.ResponseWriter, message string) {
	Problemf(w, http.StatusInternalServerError, "internal_server_error", message)
}
