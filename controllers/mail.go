package controllers

import (
	"net/http"
	"time"

	"devhive-backend/models"
	"devhive-backend/services"

	"github.com/gin-gonic/gin"
)

// MailController handles email dispatch operations through the configured mail service
type MailController struct {
	mailService services.MailService
}

// NewMailController creates a new mail controller instance
func NewMailController(mailService services.MailService) *MailController {
	return &MailController{
		mailService: mailService,
	}
}

// SendEmail sends an email to the specified recipient using the mail service
// @Summary Send email
// @Description Sends an email to the specified recipient using the mail service
// @Tags mail
// @Accept json
// @Produce json
// @Param request body models.EmailRequest true "Email request containing To, Subject, and Body"
// @Security BearerAuth
// @Success 200 {object} models.EmailResponse "Email sent successfully"
// @Failure 400 {object} models.EmailError "Bad request - invalid email request"
// @Failure 500 {object} models.EmailError "Internal server error - email sending failed"
// @Router /api/Mail/Send [post]
func (mc *MailController) SendEmail(c *gin.Context) {
	var emailRequest models.EmailRequest

	// Bind and validate the request body
	if err := c.ShouldBindJSON(&emailRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.EmailError{
			Error:   "Invalid email request",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	// Validate email fields
	if emailRequest.To == "" || emailRequest.Subject == "" || emailRequest.Body == "" {
		c.JSON(http.StatusBadRequest, models.EmailError{
			Error:   "Invalid email request",
			Code:    "MISSING_FIELDS",
			Details: "To, Subject, and Body are required fields",
		})
		return
	}

	// Send email via service
	isSent, err := mc.mailService.SendEmail(c.Request.Context(), emailRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.EmailError{
			Error:   "Email sending failed",
			Code:    "SEND_FAILED",
			Details: err.Error(),
		})
		return
	}

	if !isSent {
		c.JSON(http.StatusInternalServerError, models.EmailError{
			Error: "Email sending failed",
			Code:  "SEND_FAILED",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, models.EmailResponse{
		Message: "Email sent successfully!",
		SentAt:  time.Now(),
	})
}
