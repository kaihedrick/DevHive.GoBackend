package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Message represents a message in the system
type Message struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ProjectID       uuid.UUID  `json:"project_id" db:"project_id"`
	SenderID        uuid.UUID  `json:"sender_id" db:"sender_id"`
	Content         string     `json:"content" db:"content"`
	MessageType     string     `json:"message_type" db:"message_type"`
	ParentMessageID *uuid.UUID `json:"parent_message_id,omitempty" db:"parent_message_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	// Additional fields for API responses
	Sender  *User      `json:"sender,omitempty"`
	Replies []*Message `json:"replies,omitempty"`
}

// MessageCreateRequest represents the request to create a new message
type MessageCreateRequest struct {
	Content         string     `json:"content" binding:"required,min=1,max=2000"`
	MessageType     string     `json:"message_type" binding:"required,oneof=text file image system"`
	ParentMessageID *uuid.UUID `json:"parent_message_id,omitempty"`
}

// MessageUpdateRequest represents the request to update a message
type MessageUpdateRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// CreateMessage creates a new message in the database
func CreateMessage(db *sql.DB, req MessageCreateRequest, projectID, senderID uuid.UUID) (*Message, error) {
	message := &Message{
		ID:              uuid.New(),
		ProjectID:       projectID,
		SenderID:        senderID,
		Content:         req.Content,
		MessageType:     req.MessageType,
		ParentMessageID: req.ParentMessageID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	query := `
		INSERT INTO messages (id, project_id, sender_id, content, message_type, parent_message_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, project_id, sender_id, content, message_type, parent_message_id, created_at, updated_at
	`

	err := db.QueryRow(
		query,
		message.ID, message.ProjectID, message.SenderID, message.Content,
		message.MessageType, message.ParentMessageID, message.CreatedAt, message.UpdatedAt,
	).Scan(
		&message.ID, &message.ProjectID, &message.SenderID, &message.Content,
		&message.MessageType, &message.ParentMessageID, &message.CreatedAt, &message.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return message, nil
}

// GetMessage retrieves a message by ID with sender information
func GetMessage(db *sql.DB, messageID uuid.UUID) (*Message, error) {
	message := &Message{}
	query := `
		SELECT m.id, m.project_id, m.sender_id, m.content, m.message_type, m.parent_message_id, m.created_at, m.updated_at
		FROM messages m WHERE m.id = $1
	`

	err := db.QueryRow(query, messageID).Scan(
		&message.ID, &message.ProjectID, &message.SenderID, &message.Content,
		&message.MessageType, &message.ParentMessageID, &message.CreatedAt, &message.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Get sender information
	sender, err := GetUserByID(db, message.SenderID)
	if err != nil {
		return nil, err
	}
	message.Sender = sender

	return message, nil
}

// GetMessages retrieves messages for a project with pagination
func GetMessages(db *sql.DB, projectID uuid.UUID, limit, offset int) ([]*Message, error) {
	query := `
		SELECT m.id, m.project_id, m.sender_id, m.content, m.message_type, m.parent_message_id, m.created_at, m.updated_at
		FROM messages m 
		WHERE m.project_id = $1 AND m.parent_message_id IS NULL
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		message := &Message{}
		err := rows.Scan(
			&message.ID, &message.ProjectID, &message.SenderID, &message.Content,
			&message.MessageType, &message.ParentMessageID, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get sender information
		sender, err := GetUserByID(db, message.SenderID)
		if err != nil {
			return nil, err
		}
		message.Sender = sender

		// Get replies
		replies, err := GetMessageReplies(db, message.ID)
		if err != nil {
			return nil, err
		}
		message.Replies = replies

		messages = append(messages, message)
	}

	return messages, nil
}

// GetMessageReplies retrieves replies to a specific message
func GetMessageReplies(db *sql.DB, parentMessageID uuid.UUID) ([]*Message, error) {
	query := `
		SELECT m.id, m.project_id, m.sender_id, m.content, m.message_type, m.parent_message_id, m.created_at, m.updated_at
		FROM messages m 
		WHERE m.parent_message_id = $1
		ORDER BY m.created_at ASC
	`

	rows, err := db.Query(query, parentMessageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []*Message
	for rows.Next() {
		message := &Message{}
		err := rows.Scan(
			&message.ID, &message.ProjectID, &message.SenderID, &message.Content,
			&message.MessageType, &message.ParentMessageID, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get sender information
		sender, err := GetUserByID(db, message.SenderID)
		if err != nil {
			return nil, err
		}
		message.Sender = sender

		replies = append(replies, message)
	}

	return replies, nil
}

// UpdateMessage updates a message in the database
func UpdateMessage(db *sql.DB, messageID uuid.UUID, req MessageUpdateRequest) (*Message, error) {
	message, err := GetMessage(db, messageID)
	if err != nil {
		return nil, err
	}

	message.Content = req.Content
	message.UpdatedAt = time.Now()

	query := `
		UPDATE messages 
		SET content = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, project_id, sender_id, content, message_type, parent_message_id, created_at, updated_at
	`

	err = db.QueryRow(
		query,
		message.Content, message.UpdatedAt, message.ID,
	).Scan(
		&message.ID, &message.ProjectID, &message.SenderID, &message.Content,
		&message.MessageType, &message.ParentMessageID, &message.CreatedAt, &message.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return message, nil
}

// DeleteMessage deletes a message from the database
func DeleteMessage(db *sql.DB, messageID uuid.UUID) error {
	// First delete all replies
	_, err := db.Exec("DELETE FROM messages WHERE parent_message_id = $1", messageID)
	if err != nil {
		return err
	}

	// Then delete the main message
	query := `DELETE FROM messages WHERE id = $1`
	_, err = db.Exec(query, messageID)
	return err
}

// GetMessageCount retrieves the total count of messages for a project
func GetMessageCount(db *sql.DB, projectID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM messages WHERE project_id = $1 AND parent_message_id IS NULL`
	err := db.QueryRow(query, projectID).Scan(&count)
	return count, err
}

// SearchMessages searches for messages in a project by content
func SearchMessages(db *sql.DB, projectID uuid.UUID, searchTerm string, limit, offset int) ([]*Message, error) {
	query := `
		SELECT m.id, m.project_id, m.sender_id, m.content, m.message_type, m.parent_message_id, m.created_at, m.updated_at
		FROM messages m 
		WHERE m.project_id = $1 AND m.content ILIKE $2
		ORDER BY m.created_at DESC
		LIMIT $3 OFFSET $4
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := db.Query(query, projectID, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		message := &Message{}
		err := rows.Scan(
			&message.ID, &message.ProjectID, &message.SenderID, &message.Content,
			&message.MessageType, &message.ParentMessageID, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get sender information
		sender, err := GetUserByID(db, message.SenderID)
		if err != nil {
			return nil, err
		}
		message.Sender = sender

		messages = append(messages, message)
	}

	return messages, nil
}
