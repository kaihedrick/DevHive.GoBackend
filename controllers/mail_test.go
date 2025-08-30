package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMailService is a mock implementation of the MailService interface
type MockMailService struct {
	mock.Mock
}

func (m *MockMailService) SendEmail(ctx context.Context, emailRequest models.EmailRequest) (bool, error) {
	args := m.Called(ctx, emailRequest)
	return args.Bool(0), args.Error(1)
}

func TestSendEmail_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockMailService := new(MockMailService)
	mailController := NewMailController(mockMailService)

	router.POST("/mail/send", mailController.SendEmail)

	// Test data
	emailRequest := models.EmailRequest{
		To:      "test@example.com",
		Subject: "Test Subject",
		Body:    "Test Body",
	}

	// Mock expectations
	mockMailService.On("SendEmail", mock.Anything, emailRequest).Return(true, nil)

	// Create request
	jsonData, _ := json.Marshal(emailRequest)
	req, _ := http.NewRequest("POST", "/mail/send", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.EmailResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Email sent successfully!", response.Message)

	mockMailService.AssertExpectations(t)
}

func TestSendEmail_InvalidRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockMailService := new(MockMailService)
	mailController := NewMailController(mockMailService)

	router.POST("/mail/send", mailController.SendEmail)

	// Test data - missing required fields
	emailRequest := models.EmailRequest{
		To:      "test@example.com",
		Subject: "", // Missing subject
		Body:    "Test Body",
	}

	// Create request
	jsonData, _ := json.Marshal(emailRequest)
	req, _ := http.NewRequest("POST", "/mail/send", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.EmailError
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid email request", response.Error)
	assert.Equal(t, "INVALID_REQUEST", response.Code)
}

func TestSendEmail_ServiceFailure(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockMailService := new(MockMailService)
	mailController := NewMailController(mockMailService)

	router.POST("/mail/send", mailController.SendEmail)

	// Test data
	emailRequest := models.EmailRequest{
		To:      "test@example.com",
		Subject: "Test Subject",
		Body:    "Test Body",
	}

	// Mock expectations - service failure
	mockMailService.On("SendEmail", mock.Anything, emailRequest).Return(false, assert.AnError)

	// Create request
	jsonData, _ := json.Marshal(emailRequest)
	req, _ := http.NewRequest("POST", "/mail/send", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.EmailError
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Email sending failed", response.Error)
	assert.Equal(t, "SEND_FAILED", response.Code)

	mockMailService.AssertExpectations(t)
}
