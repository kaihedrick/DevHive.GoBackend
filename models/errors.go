package models

import "errors"

// Common application errors
var (
	ErrAccessDenied        = errors.New("access denied")
	ErrInvalidSprintStatus = errors.New("invalid sprint status")
	ErrNotFound            = errors.New("resource not found")
	ErrInvalidInput        = errors.New("invalid input")
	ErrDatabaseError       = errors.New("database error")
	ErrUnauthorized        = errors.New("unauthorized")
)
