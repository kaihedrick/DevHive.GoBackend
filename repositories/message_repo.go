package repositories

import (
	"context"

	"devhive-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageRepository defines the interface for message operations
type MessageRepository interface {
	CreateMessage(ctx context.Context, req models.MessageCreateRequest, projectID, senderID uuid.UUID) (*models.Message, error)
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*models.Message, error)
	GetMessagesByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*models.Message, error)
	UpdateMessage(ctx context.Context, messageID uuid.UUID, req models.MessageUpdateRequest) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageID uuid.UUID) error
	GetMessagesBySender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]*models.Message, error)
	SearchMessages(ctx context.Context, projectID uuid.UUID, query string, limit, offset int) ([]*models.Message, error)
	GetRecentMessages(ctx context.Context, projectID uuid.UUID, hours int) ([]*models.Message, error)
	CountMessagesByProject(ctx context.Context, projectID uuid.UUID) (int64, error)
}

// messageRepository implements MessageRepository
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository instance
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{
		db: db,
	}
}

// CreateMessage creates a new message in the database
func (r *messageRepository) CreateMessage(ctx context.Context, req models.MessageCreateRequest, projectID, senderID uuid.UUID) (*models.Message, error) {
	return models.CreateMessage(r.db, req, projectID, senderID)
}

// GetMessageByID retrieves a message by its ID
func (r *messageRepository) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*models.Message, error) {
	return models.GetMessage(r.db, messageID)
}

// GetMessagesByProject retrieves all messages for a specific project
func (r *messageRepository) GetMessagesByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	return models.GetMessages(r.db, projectID, limit, offset)
}

// UpdateMessage updates an existing message
func (r *messageRepository) UpdateMessage(ctx context.Context, messageID uuid.UUID, req models.MessageUpdateRequest) (*models.Message, error) {
	return models.UpdateMessage(r.db, messageID, req)
}

// DeleteMessage deletes a message by ID
func (r *messageRepository) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	return models.DeleteMessage(r.db, messageID)
}

// GetMessagesBySender retrieves messages sent by a specific user
func (r *messageRepository) GetMessagesBySender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	return models.GetMessagesBySender(r.db, senderID, limit, offset)
}

// SearchMessages searches messages by content in a project
func (r *messageRepository) SearchMessages(ctx context.Context, projectID uuid.UUID, query string, limit, offset int) ([]*models.Message, error) {
	return models.SearchMessages(r.db, projectID, query, limit, offset)
}

// GetRecentMessages retrieves recent messages for a project
func (r *messageRepository) GetRecentMessages(ctx context.Context, projectID uuid.UUID, hours int) ([]*models.Message, error) {
	return models.GetRecentMessages(r.db, projectID, hours)
}

// CountMessagesByProject counts the total number of messages for a project
func (r *messageRepository) CountMessagesByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	return models.CountMessages(r.db, projectID)
}
