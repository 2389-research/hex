// ABOUTME: Conversation CRUD operations for SQLite storage
// ABOUTME: Create, read, update, delete, and list conversations
package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Conversation represents a chat conversation
type Conversation struct {
	ID           string
	Title        string
	Model        string
	SystemPrompt string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsFavorite   bool
}

// CreateConversation inserts a new conversation
func CreateConversation(db *sql.DB, conv *Conversation) error {
	now := time.Now()

	// Generate UUID if ID is empty
	if conv.ID == "" {
		conv.ID = uuid.New().String()
	}

	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = now
	}
	if conv.UpdatedAt.IsZero() {
		conv.UpdatedAt = now
	}

	query := `
		INSERT INTO conversations (id, title, model, system_prompt, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, conv.ID, conv.Title, conv.Model, conv.SystemPrompt, conv.CreatedAt, conv.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert conversation: %w", err)
	}
	return nil
}

// GetConversation retrieves a conversation by ID
func GetConversation(db *sql.DB, id string) (*Conversation, error) {
	query := `
		SELECT id, title, model, system_prompt, created_at, updated_at, COALESCE(is_favorite, 0) as is_favorite
		FROM conversations
		WHERE id = ?
	`

	conv := &Conversation{}
	var systemPrompt sql.NullString
	err := db.QueryRow(query, id).Scan(
		&conv.ID,
		&conv.Title,
		&conv.Model,
		&systemPrompt,
		&conv.CreatedAt,
		&conv.UpdatedAt,
		&conv.IsFavorite,
	)
	if err != nil {
		return nil, err
	}

	if systemPrompt.Valid {
		conv.SystemPrompt = systemPrompt.String
	}

	return conv, nil
}

// GetLatestConversation retrieves the most recently updated conversation
func GetLatestConversation(db *sql.DB) (*Conversation, error) {
	query := `
		SELECT id, title, model, system_prompt, created_at, updated_at, COALESCE(is_favorite, 0) as is_favorite
		FROM conversations
		ORDER BY updated_at DESC
		LIMIT 1
	`

	conv := &Conversation{}
	var systemPrompt sql.NullString
	err := db.QueryRow(query).Scan(
		&conv.ID,
		&conv.Title,
		&conv.Model,
		&systemPrompt,
		&conv.CreatedAt,
		&conv.UpdatedAt,
		&conv.IsFavorite,
	)
	if err != nil {
		return nil, err
	}

	if systemPrompt.Valid {
		conv.SystemPrompt = systemPrompt.String
	}

	return conv, nil
}

// ListConversations returns conversations ordered by updated_at DESC
func ListConversations(db *sql.DB, limit, offset int) ([]*Conversation, error) {
	query := `
		SELECT id, title, model, system_prompt, created_at, updated_at, COALESCE(is_favorite, 0) as is_favorite
		FROM conversations
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var convs []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		var systemPrompt sql.NullString
		err := rows.Scan(&conv.ID, &conv.Title, &conv.Model, &systemPrompt, &conv.CreatedAt, &conv.UpdatedAt, &conv.IsFavorite)
		if err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}

		if systemPrompt.Valid {
			conv.SystemPrompt = systemPrompt.String
		}

		convs = append(convs, conv)
	}
	return convs, nil
}

// UpdateConversationTitle updates the title of a conversation
func UpdateConversationTitle(db *sql.DB, id, title string) error {
	query := `UPDATE conversations SET title = ?, updated_at = ? WHERE id = ?`
	_, err := db.Exec(query, title, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update title: %w", err)
	}
	return nil
}

// UpdateConversationTimestamp updates the updated_at field
func UpdateConversationTimestamp(db *sql.DB, id string) error {
	query := `UPDATE conversations SET updated_at = ? WHERE id = ?`
	_, err := db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update timestamp: %w", err)
	}
	return nil
}

// DeleteConversation deletes a conversation and its messages (via CASCADE)
func DeleteConversation(db *sql.DB, id string) error {
	query := `DELETE FROM conversations WHERE id = ?`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}
	return nil
}

// SetFavorite sets or unsets a conversation as favorite
func SetFavorite(db *sql.DB, id string, isFavorite bool) error {
	query := `UPDATE conversations SET is_favorite = ?, updated_at = ? WHERE id = ?`
	_, err := db.Exec(query, isFavorite, time.Now(), id)
	if err != nil {
		return fmt.Errorf("set favorite: %w", err)
	}
	return nil
}

// ListFavorites returns all favorite conversations ordered by updated_at DESC
func ListFavorites(db *sql.DB) ([]*Conversation, error) {
	query := `
		SELECT id, title, model, system_prompt, created_at, updated_at, is_favorite
		FROM conversations
		WHERE is_favorite = 1
		ORDER BY updated_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var convs []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		var systemPrompt sql.NullString
		err := rows.Scan(&conv.ID, &conv.Title, &conv.Model, &systemPrompt, &conv.CreatedAt, &conv.UpdatedAt, &conv.IsFavorite)
		if err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}

		if systemPrompt.Valid {
			conv.SystemPrompt = systemPrompt.String
		}

		convs = append(convs, conv)
	}
	return convs, nil
}
