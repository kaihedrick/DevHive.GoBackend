package services

import (
	"context"
	"fmt"
	"strings"

	"devhive-backend/models"
	"devhive-backend/repositories"

	"github.com/google/uuid"
)

// MessageService defines the interface for message operations
type MessageService interface {
	CreateMessage(ctx context.Context, req models.MessageCreateRequest, projectID, senderID uuid.UUID) (*models.Message, error)
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*models.Message, error)
	GetMessagesByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*models.Message, error)
	UpdateMessage(ctx context.Context, messageID uuid.UUID, req models.MessageUpdateRequest) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageID uuid.UUID) error
	GetMessagesBySender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]*models.Message, error)
	SearchMessages(ctx context.Context, projectID uuid.UUID, query string, limit, offset int) ([]*models.Message, error)
	GetRecentMessages(ctx context.Context, projectID uuid.UUID, hours int) ([]*models.Message, error)
	CountMessagesByProject(ctx context.Context, projectID uuid.UUID) (int64, error)
	// Mobile-specific methods
	GetMessagesForMobile(projectID uuid.UUID, userID string, page, limit int, search string) ([]models.MobileMessage, int, error)
}

// messageService implements MessageService
type messageService struct {
	messageRepo repositories.MessageRepository
}

// NewMessageService creates a new message service instance
func NewMessageService(messageRepo repositories.MessageRepository) MessageService {
	return &messageService{
		messageRepo: messageRepo,
	}
}

// CreateMessage creates a new message
func (s *messageService) CreateMessage(ctx context.Context, req models.MessageCreateRequest, projectID, senderID uuid.UUID) (*models.Message, error) {
	return s.messageRepo.CreateMessage(ctx, req, projectID, senderID)
}

// GetMessageByID retrieves a message by its ID
func (s *messageService) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*models.Message, error) {
	return s.messageRepo.GetMessageByID(ctx, messageID)
}

// GetMessagesByProject retrieves all messages for a specific project
func (s *messageService) GetMessagesByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	return s.messageRepo.GetMessagesByProject(ctx, projectID, limit, offset)
}

// UpdateMessage updates an existing message
func (s *messageService) UpdateMessage(ctx context.Context, messageID uuid.UUID, req models.MessageUpdateRequest) (*models.Message, error) {
	return s.messageRepo.UpdateMessage(ctx, messageID, req)
}

// DeleteMessage deletes a message by ID
func (s *messageService) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	return s.messageRepo.DeleteMessage(ctx, messageID)
}

// GetMessagesBySender retrieves messages sent by a specific user
func (s *messageService) GetMessagesBySender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	return s.messageRepo.GetMessagesBySender(ctx, senderID, limit, offset)
}

// SearchMessages searches messages by content in a project
func (s *messageService) SearchMessages(ctx context.Context, projectID uuid.UUID, query string, limit, offset int) ([]*models.Message, error) {
	return s.messageRepo.SearchMessages(ctx, projectID, query, limit, offset)
}

// GetRecentMessages retrieves recent messages for a project
func (s *messageService) GetRecentMessages(ctx context.Context, projectID uuid.UUID, hours int) ([]*models.Message, error) {
	return s.messageRepo.GetRecentMessages(ctx, projectID, hours)
}

// CountMessagesByProject counts the total number of messages for a project
func (s *messageService) CountMessagesByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	return s.messageRepo.CountMessagesByProject(ctx, projectID)
}

// GetMessagesForMobile retrieves messages optimized for mobile consumption
func (s *messageService) GetMessagesForMobile(projectID uuid.UUID, userID string, page, limit int, search string) ([]models.MobileMessage, int, error) {
	// Parse user ID
	_, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	// For now, we'll use a simple approach since we need to check project membership
	// TODO: Implement proper project membership check when project service is available

	// Get messages for the project
	offset := (page - 1) * limit
	messages, err := s.messageRepo.GetMessagesByProject(context.Background(), projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting messages: %w", err)
	}

	// Get total count
	total, err := s.messageRepo.CountMessagesByProject(context.Background(), projectID)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting messages: %w", err)
	}

	// Convert to mobile messages
	mobileMessages := make([]models.MobileMessage, 0, len(messages))
	for _, message := range messages {
		// Apply search filter if provided
		if search != "" && !contains(message.Content, search) {
			continue
		}

		mobileMessage := models.MobileMessage{
			ID:         message.ID,
			Content:    message.Content,
			Type:       "message", // Default type since it's not in the model
			CreatedAt:  message.CreatedAt,
			UpdatedAt:  message.UpdatedAt,
			UserID:     message.SenderID,
			UserName:   "", // Will be populated if needed
			UserAvatar: "", // Will be populated if needed
		}

		// TODO: Get user information when user service is available
		mobileMessages = append(mobileMessages, mobileMessage)
	}

	return mobileMessages, int(total), nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}
