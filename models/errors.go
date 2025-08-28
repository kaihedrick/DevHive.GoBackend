package models

import "errors"

// Common application errors
var (
	ErrAccessDenied        = errors.New("access denied")
	ErrInvalidSprintStatus = errors.New("invalid sprint status")
	ErrInvalidSprintDates  = errors.New("invalid sprint dates: start date must be before end date")
	ErrSprintAlreadyActive = errors.New("there is already an active sprint in this project")
	ErrNotFound            = errors.New("resource not found")
	ErrInvalidInput        = errors.New("invalid input")
	ErrDatabaseError       = errors.New("database error")
	ErrUnauthorized        = errors.New("unauthorized")
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error message"`
}
