# Mail Functionality

This document describes the mail functionality implemented in the DevHive Go Backend, which was translated from the original .NET API controller.

## Overview

The mail functionality allows the application to send emails through a configurable SMTP service. It includes:

- **Mail Controller**: Handles HTTP requests for email operations
- **Mail Service**: Business logic for email processing and SMTP communication
- **Email Models**: Data structures for email requests and responses
- **Mock Email Service**: Development/testing fallback when SMTP is not configured

## Architecture

### Components

1. **`models/mail.go`**: Contains email-related data structures
   - `EmailRequest`: Input model for email requests
   - `EmailResponse`: Success response model
   - `EmailError`: Error response model

2. **`internal/service/mail_service.go`**: Service layer implementation
   - `MailService` interface
   - SMTP-based email sending
   - Mock email service for development

3. **`controllers/mail.go`**: HTTP controller
   - RESTful endpoint for sending emails
   - Input validation and error handling
   - gRPC documentation

### API Endpoint

```
POST /api/v1/mail/send
```

**Request Body:**
```json
{
  "to": "recipient@example.com",
  "subject": "Email Subject",
  "body": "Email content"
}
```

**Success Response (200):**
```json
{
  "message": "Email sent successfully!",
  "sentAt": "2024-01-15T10:30:00Z"
}
```

**Error Response (400/500):**
```json
{
  "error": "Error description",
  "code": "ERROR_CODE",
  "details": "Additional error details"
}
```

## Configuration

### Environment Variables

Add the following to your `.env` file:

```bash
# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@devhive.com
```

### SMTP Providers

The service supports any SMTP provider. Common configurations:

**Gmail:**
- Host: `smtp.gmail.com`
- Port: `587` (TLS) or `465` (SSL)
- Username: Your Gmail address
- Password: App-specific password (not your regular password)

**Outlook/Hotmail:**
- Host: `smtp-mail.outlook.com`
- Port: `587`
- Username: Your Outlook email
- Password: Your account password

**Custom SMTP Server:**
- Host: Your SMTP server address
- Port: Your SMTP server port
- Username: SMTP username
- Password: SMTP password

## Development Mode

When SMTP configuration is not provided, the service automatically falls back to a mock email service that logs email details to the console. This is useful for development and testing.

## Security Considerations

1. **Authentication Required**: The mail endpoint is protected by JWT authentication
2. **Input Validation**: All email fields are validated before processing
3. **Environment Variables**: Sensitive SMTP credentials are stored in environment variables
4. **Rate Limiting**: Subject to the same rate limiting as other protected endpoints

## Testing

Run the mail controller tests:

```bash
go test ./controllers -v -run TestSendEmail
```

## Usage Examples

### Send Email via API

```bash
curl -X POST http://localhost:8080/api/v1/mail/send \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "to": "user@example.com",
    "subject": "Welcome to DevHive",
    "body": "Thank you for joining our platform!"
  }'
```

### Programmatic Usage

```go
mailService := service.NewMailService()
emailRequest := models.EmailRequest{
    To:      "user@example.com",
    Subject: "Test Email",
    Body:    "This is a test email",
}

success, err := mailService.SendEmail(context.Background(), emailRequest)
if err != nil {
    log.Printf("Failed to send email: %v", err)
    return
}

if success {
    log.Println("Email sent successfully")
}
```

## Error Handling

The service handles various error scenarios:

- **Missing SMTP Configuration**: Falls back to mock service
- **Invalid Email Format**: Returns 400 Bad Request
- **SMTP Connection Failure**: Returns 500 Internal Server Error
- **Authentication Failure**: Returns 500 Internal Server Error

## Future Enhancements

Potential improvements for the mail functionality:

1. **Email Templates**: Support for HTML templates and dynamic content
2. **Attachment Support**: Handle file attachments
3. **Email Queue**: Asynchronous email processing
4. **Multiple Recipients**: Support for CC, BCC, and bulk emails
5. **Email Tracking**: Delivery and read receipts
6. **Rate Limiting**: Email-specific rate limiting
7. **Email Logging**: Database storage of sent emails

## Migration from .NET

This implementation maintains the same API contract as the original .NET controller:

| .NET Feature | Go Equivalent |
|--------------|----------------|
| `[HttpPost("Send")]` | `POST /api/v1/mail/send` |
| `EmailRequest` | `models.EmailRequest` |
| `IActionResult` | `gin.Context` responses |
| `IMailService` | `service.MailService` interface |
| `BadRequest()` | `c.JSON(http.StatusBadRequest, ...)` |
| `Ok()` | `c.JSON(http.StatusOK, ...)` |
| `StatusCode(500, ...)` | `c.JSON(http.StatusInternalServerError, ...)` |

The Go implementation provides the same functionality with improved error handling, comprehensive testing, and better separation of concerns.
