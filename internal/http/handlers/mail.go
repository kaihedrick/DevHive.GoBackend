package handlers

import (
	"net/http"

	"devhive-backend/internal/config"
	"devhive-backend/internal/http/response"
)

type MailHandler struct {
	cfg *config.Config
}

func NewMailHandler(cfg *config.Config) *MailHandler {
	return &MailHandler{
		cfg: cfg,
	}
}

// SendEmailRequest represents the email sending request
type SendEmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	HTML    string `json:"html,omitempty"`
}

// SendEmail handles sending emails
func (h *MailHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req SendEmailRequest
	if !response.Decode(w, r, &req) {
		return
	}

	// TODO: Implement actual email sending using Mailgun or another service
	// For now, just return a success response
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Email sent successfully",
		"to":      req.To,
		"subject": req.Subject,
	})
}

