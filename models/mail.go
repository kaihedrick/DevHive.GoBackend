package models

import (
	"time"
)

// EmailRequest represents the request to send an email
type EmailRequest struct {
	To      string `json:"to" binding:"required,email" example:"recipient@example.com"`
	Subject string `json:"subject" binding:"required" example:"Important Notification"`
	Body    string `json:"body" binding:"required" example:"This is the email content"`
}

// EmailResponse represents the response from email operations
type EmailResponse struct {
	Message string    `json:"message" example:"Email sent successfully!"`
	SentAt  time.Time `json:"sentAt"`
}

// EmailError represents email-related errors
type EmailError struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}
