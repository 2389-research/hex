// ABOUTME: Implementation of MessageService with event publishing
// ABOUTME: Handles message storage, retrieval, filtering, and PubSub events

package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/2389-research/hex/internal/pubsub"
	"github.com/google/uuid"
)

// messageService implements MessageService with PubSub support
type messageService struct {
	db     *sql.DB
	broker *pubsub.Broker[Message]
}

// NewMessageService creates a new message service with event publishing
func NewMessageService(db *sql.DB) MessageService {
	return &messageService{
		db:     db,
		broker: pubsub.NewBroker[Message](),
	}
}

// Subscribe allows subscribing to message events
func (s *messageService) Subscribe(ctx context.Context) <-chan pubsub.Event[Message] {
	return s.broker.Subscribe(ctx)
}

// Add stores a new message and publishes a Created event
func (s *messageService) Add(_ context.Context, msg *Message) error {
	// Generate ID if not provided
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	// Set timestamp
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	// Convert empty strings to NULL for nullable fields
	var provider, model sql.NullString
	if msg.Provider != "" {
		provider = sql.NullString{String: msg.Provider, Valid: true}
	}
	if msg.Model != "" {
		model = sql.NullString{String: msg.Model, Valid: true}
	}

	// Insert into database
	query := `
		INSERT INTO messages (
			id,
			conversation_id,
			role,
			content,
			provider,
			model,
			is_summary,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		msg.ID,
		msg.ConversationID,
		msg.Role,
		msg.Content,
		provider,
		model,
		msg.IsSummary,
		msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	// Publish Created event
	s.broker.Publish(pubsub.Created, *msg)

	return nil
}

// GetByConversation returns all messages for a conversation ordered by created_at ASC
func (s *messageService) GetByConversation(_ context.Context, convID string) ([]*Message, error) {
	query := `
		SELECT
			id,
			conversation_id,
			role,
			content,
			provider,
			model,
			is_summary,
			created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, convID)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = fmt.Errorf("close rows: %w", closeErr)
		}
	}()

	return s.scanMessages(rows)
}

// GetSummaries returns only summary messages for a conversation
func (s *messageService) GetSummaries(_ context.Context, convID string) ([]*Message, error) {
	query := `
		SELECT
			id,
			conversation_id,
			role,
			content,
			provider,
			model,
			is_summary,
			created_at
		FROM messages
		WHERE conversation_id = ?
		  AND is_summary = 1
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, convID)
	if err != nil {
		return nil, fmt.Errorf("query summaries: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = fmt.Errorf("close rows: %w", closeErr)
		}
	}()

	return s.scanMessages(rows)
}

// scanMessages is a helper to scan message rows and handle NULL fields
func (s *messageService) scanMessages(rows *sql.Rows) ([]*Message, error) {
	messages := make([]*Message, 0)

	for rows.Next() {
		var msg Message
		var provider, model sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.Role,
			&msg.Content,
			&provider,
			&model,
			&msg.IsSummary,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		// Convert NULL to empty string
		if provider.Valid {
			msg.Provider = provider.String
		}
		if model.Valid {
			msg.Model = model.String
		}

		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}
