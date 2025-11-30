// Package storage provides database operations for conversations, messages, and metadata.
// ABOUTME: Message CRUD operations for SQLite storage
// ABOUTME: Create, read, and list messages within conversations
package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message
type Message struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	ToolCalls      string // JSON string
	Metadata       string // JSON string
	CreatedAt      time.Time
}

// CreateMessage inserts a new message and updates conversation timestamp atomically
func CreateMessage(db *sql.DB, msg *Message) error {
	// Generate UUID if ID is empty
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	// Use transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Insert message
	query := `
		INSERT INTO messages (id, conversation_id, role, content, tool_calls, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(query, msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.ToolCalls, msg.Metadata, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	// Update conversation timestamp
	if err := updateConversationTimestampTx(tx, msg.ConversationID); err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// updateConversationTimestampTx updates conversation timestamp within a transaction
func updateConversationTimestampTx(tx *sql.Tx, id string) error {
	query := `UPDATE conversations SET updated_at = ? WHERE id = ?`
	_, err := tx.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update timestamp: %w", err)
	}
	return nil
}

// GetMessage retrieves a message by ID
func GetMessage(db *sql.DB, id string) (*Message, error) {
	query := `
		SELECT id, conversation_id, role, content, tool_calls, metadata, created_at
		FROM messages
		WHERE id = ?
	`

	msg := &Message{}
	var toolCalls, metadata sql.NullString
	err := db.QueryRow(query, id).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.Role,
		&msg.Content,
		&toolCalls,
		&metadata,
		&msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get message: %w", err)
	}

	if toolCalls.Valid {
		msg.ToolCalls = toolCalls.String
	}
	if metadata.Valid {
		msg.Metadata = metadata.String
	}

	return msg, nil
}

// ListMessages returns all messages for a conversation ordered by created_at
func ListMessages(db *sql.DB, conversationID string) ([]*Message, error) {
	query := `
		SELECT id, conversation_id, role, content, tool_calls, metadata, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`

	rows, err := db.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var msgs []*Message
	for rows.Next() {
		msg := &Message{}
		var toolCalls, metadata sql.NullString
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &toolCalls, &metadata, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		if toolCalls.Valid {
			msg.ToolCalls = toolCalls.String
		}
		if metadata.Valid {
			msg.Metadata = metadata.String
		}

		msgs = append(msgs, msg)
	}
	return msgs, nil
}
