package services

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"devhive-backend/models"
)

// MailService defines the interface for mail operations
type MailService interface {
	SendEmail(ctx context.Context, emailRequest models.EmailRequest) (bool, error)
}

// mailService implements MailService
type mailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
}

// NewMailService creates a new mail service instance
func NewMailService() MailService {
	return &mailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("FROM_EMAIL"),
	}
}

// SendEmail sends an email using the configured SMTP service
func (ms *mailService) SendEmail(ctx context.Context, emailRequest models.EmailRequest) (bool, error) {
	// Validate email fields
	if emailRequest.To == "" || emailRequest.Subject == "" || emailRequest.Body == "" {
		return false, fmt.Errorf("invalid email request: missing required fields")
	}

	// Check if SMTP configuration is available
	if ms.smtpHost == "" || ms.smtpPort == "" || ms.smtpUsername == "" || ms.smtpPassword == "" {
		log.Println("SMTP configuration not found, using mock email service")
		return ms.sendMockEmail(ctx, emailRequest)
	}

	// Prepare email content
	to := []string{emailRequest.To}
	subject := emailRequest.Subject
	body := emailRequest.Body

	// Create email message
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s",
		strings.Join(to, ","), subject, body)

	// SMTP authentication
	auth := smtp.PlainAuth("", ms.smtpUsername, ms.smtpPassword, ms.smtpHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", ms.smtpHost, ms.smtpPort)
	err := smtp.SendMail(addr, auth, ms.fromEmail, to, []byte(message))
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return false, fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully to %s", emailRequest.To)
	return true, nil
}

// sendMockEmail sends a mock email for development/testing purposes
func (ms *mailService) sendMockEmail(ctx context.Context, emailRequest models.EmailRequest) (bool, error) {
	log.Printf("MOCK EMAIL SENT:")
	log.Printf("  To: %s", emailRequest.To)
	log.Printf("  Subject: %s", emailRequest.Subject)
	log.Printf("  Body: %s", emailRequest.Body)

	// In a real implementation, you might want to store this in a database
	// or send it to a development email service like MailHog

	return true, nil
}
