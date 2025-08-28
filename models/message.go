package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Message represents a message in the system
// @Description Message represents a message in the system
type Message struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	Content   string    `json:"content" gorm:"not null;size:1000" example:"Hello team! How's the project going?"`
	ProjectID uuid.UUID `json:"project_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000"`
	SenderID  uuid.UUID `json:"sender_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// Additional fields for API responses
	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Sender  *User    `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
}

// TableName specifies the table name for the Message model
func (Message) TableName() string {
	return "messages"
}

// MessageCreateRequest represents the request to create a new message
// @Description Request to create a new message
type MessageCreateRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000" example:"Hello team! How's the project going?"`
}

// MessageUpdateRequest represents the request to update a message
// @Description Request to update an existing message
type MessageUpdateRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000" example:"Updated message content"`
}

// CreateMessage creates a new message in the database using GORM
func CreateMessage(db *gorm.DB, req MessageCreateRequest, projectID, senderID uuid.UUID) (*Message, error) {
	message := &Message{
		ID:        uuid.New(),
		Content:   req.Content,
		ProjectID: projectID,
		SenderID:  senderID,
	}

	if err := db.Create(message).Error; err != nil {
		return nil, err
	}

	return message, nil
}

// GetMessage retrieves a message by ID using GORM
func GetMessage(db *gorm.DB, messageID uuid.UUID) (*Message, error) {
	var message Message
	if err := db.Preload("Project").Preload("Sender").
		Where("id = ?", messageID).First(&message).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

// GetMessages retrieves all messages for a specific project using GORM
func GetMessages(db *gorm.DB, projectID uuid.UUID, limit, offset int) ([]*Message, error) {
	var messages []*Message
	if err := db.Where("project_id = ?", projectID).
		Preload("Sender").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// UpdateMessage updates an existing message using GORM
func UpdateMessage(db *gorm.DB, messageID uuid.UUID, req MessageUpdateRequest) (*Message, error) {
	updates := map[string]interface{}{
		"content": req.Content,
	}

	if err := db.Model(&Message{}).Where("id = ?", messageID).Updates(updates).Error; err != nil {
		return nil, err
	}

	return GetMessage(db, messageID)
}

// DeleteMessage deletes a message by ID using GORM
func DeleteMessage(db *gorm.DB, messageID uuid.UUID) error {
	return db.Where("id = ?", messageID).Delete(&Message{}).Error
}

// GetMessagesBySender retrieves messages sent by a specific user using GORM
func GetMessagesBySender(db *gorm.DB, senderID uuid.UUID, limit, offset int) ([]*Message, error) {
	var messages []*Message
	if err := db.Where("sender_id = ?", senderID).
		Preload("Project").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// GetRecentMessages retrieves recent messages for a project using GORM
func GetRecentMessages(db *gorm.DB, projectID uuid.UUID, hours int) ([]*Message, error) {
	var messages []*Message
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	if err := db.Where("project_id = ? AND created_at > ?", projectID, since).
		Preload("Sender").
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// CountMessages counts the total number of messages for a project
func CountMessages(db *gorm.DB, projectID uuid.UUID) (int64, error) {
	var count int64
	if err := db.Model(&Message{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// SearchMessages searches messages by content in a project using GORM
func SearchMessages(db *gorm.DB, projectID uuid.UUID, query string, limit, offset int) ([]*Message, error) {
	var messages []*Message
	searchQuery := "%" + query + "%"

	if err := db.Where("project_id = ? AND content ILIKE ?", projectID, searchQuery).
		Preload("Sender").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// BeforeCreate is a GORM hook that runs before creating a message
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a message
func (m *Message) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
