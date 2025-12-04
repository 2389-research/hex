// ABOUTME: Implementation of ConversationService with event publishing
// ABOUTME: Handles conversation CRUD, token tracking, and cost calculation

package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/2389-research/hex/internal/pubsub"
	"github.com/2389-research/hex/internal/storage"
	"github.com/google/uuid"
)

// conversationService implements ConversationService with PubSub support
type conversationService struct {
	db     *sql.DB
	broker *pubsub.Broker[Conversation]
}

// NewConversationService creates a new conversation service with event publishing
func NewConversationService(db *sql.DB) ConversationService {
	return &conversationService{
		db:     db,
		broker: pubsub.NewBroker[Conversation](),
	}
}

// Subscribe allows subscribing to conversation events
func (s *conversationService) Subscribe(ctx context.Context) <-chan pubsub.Event[Conversation] {
	return s.broker.Subscribe(ctx)
}

// Create creates a new conversation and publishes a Created event
func (s *conversationService) Create(_ context.Context, title string) (*Conversation, error) {
	now := time.Now()
	id := uuid.New().String()

	// Create storage conversation
	storageConv := &storage.Conversation{
		ID:        id,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := storage.CreateConversation(s.db, storageConv); err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}

	// Convert to service model
	conv := &Conversation{
		ID:               id,
		Title:            title,
		CreatedAt:        now,
		UpdatedAt:        now,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalCost:        0.0,
		SummaryMessageID: nil,
	}

	// Publish Created event
	s.broker.Publish(pubsub.Created, *conv)

	return conv, nil
}

// Get retrieves a conversation by ID
func (s *conversationService) Get(_ context.Context, id string) (*Conversation, error) {
	storageConv, err := storage.GetConversation(s.db, id)
	if err != nil {
		return nil, err
	}

	// Query token tracking fields
	var promptTokens, completionTokens int64
	var totalCost float64
	var summaryMessageID sql.NullString

	query := `
		SELECT
			COALESCE(prompt_tokens, 0),
			COALESCE(completion_tokens, 0),
			COALESCE(total_cost, 0.0),
			summary_message_id
		FROM conversations
		WHERE id = ?
	`
	err = s.db.QueryRow(query, id).Scan(&promptTokens, &completionTokens, &totalCost, &summaryMessageID)
	if err != nil {
		return nil, fmt.Errorf("query token fields: %w", err)
	}

	conv := &Conversation{
		ID:               storageConv.ID,
		Title:            storageConv.Title,
		CreatedAt:        storageConv.CreatedAt,
		UpdatedAt:        storageConv.UpdatedAt,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalCost:        totalCost,
		SummaryMessageID: nil,
	}

	if summaryMessageID.Valid {
		conv.SummaryMessageID = &summaryMessageID.String
	}

	return conv, nil
}

// List returns all conversations ordered by updated_at DESC
func (s *conversationService) List(_ context.Context) ([]*Conversation, error) {
	// Use a large limit to get all conversations
	storageConvs, err := storage.ListConversations(s.db, 1000, 0)
	if err != nil {
		return nil, err
	}

	convs := make([]*Conversation, 0, len(storageConvs))
	for _, sc := range storageConvs {
		// Query token tracking fields for each conversation
		var promptTokens, completionTokens int64
		var totalCost float64
		var summaryMessageID sql.NullString

		query := `
			SELECT
				COALESCE(prompt_tokens, 0),
				COALESCE(completion_tokens, 0),
				COALESCE(total_cost, 0.0),
				summary_message_id
			FROM conversations
			WHERE id = ?
		`
		err := s.db.QueryRow(query, sc.ID).Scan(&promptTokens, &completionTokens, &totalCost, &summaryMessageID)
		if err != nil {
			return nil, fmt.Errorf("query token fields for %s: %w", sc.ID, err)
		}

		conv := &Conversation{
			ID:               sc.ID,
			Title:            sc.Title,
			CreatedAt:        sc.CreatedAt,
			UpdatedAt:        sc.UpdatedAt,
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalCost:        totalCost,
			SummaryMessageID: nil,
		}

		if summaryMessageID.Valid {
			conv.SummaryMessageID = &summaryMessageID.String
		}

		convs = append(convs, conv)
	}

	return convs, nil
}

// Update saves conversation changes and publishes an Updated event
func (s *conversationService) Update(ctx context.Context, conv *Conversation) error {
	// Update basic fields
	if err := storage.UpdateConversationTitle(s.db, conv.ID, conv.Title); err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}

	// Get updated conversation to publish
	updated, err := s.Get(ctx, conv.ID)
	if err != nil {
		return fmt.Errorf("get updated conversation: %w", err)
	}

	// Publish Updated event
	s.broker.Publish(pubsub.Updated, *updated)

	return nil
}

// Delete removes a conversation and publishes a Deleted event
func (s *conversationService) Delete(ctx context.Context, id string) error {
	// Get conversation before deleting (for event payload)
	conv, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get conversation before delete: %w", err)
	}

	// Delete from storage
	if err := storage.DeleteConversation(s.db, id); err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}

	// Publish Deleted event
	s.broker.Publish(pubsub.Deleted, *conv)

	return nil
}

// UpdateTokenUsage updates token counts and calculates cost
// Cost calculation: (promptTokens/1M * $3) + (completionTokens/1M * $15)
func (s *conversationService) UpdateTokenUsage(ctx context.Context, id string, promptTokens, completionTokens int64) error {
	// Get current token values
	current, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get current conversation: %w", err)
	}

	// Accumulate tokens
	newPromptTokens := current.PromptTokens + promptTokens
	newCompletionTokens := current.CompletionTokens + completionTokens

	// Calculate cost: (promptTokens/1M * $3) + (completionTokens/1M * $15)
	promptCost := (float64(newPromptTokens) / 1_000_000.0) * 3.0
	completionCost := (float64(newCompletionTokens) / 1_000_000.0) * 15.0
	totalCost := promptCost + completionCost

	// Update in database
	query := `
		UPDATE conversations
		SET
			prompt_tokens = ?,
			completion_tokens = ?,
			total_cost = ?,
			updated_at = ?
		WHERE id = ?
	`
	_, err = s.db.Exec(query, newPromptTokens, newCompletionTokens, totalCost, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update token usage: %w", err)
	}

	// Get updated conversation
	updated, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get updated conversation: %w", err)
	}

	// Publish Updated event
	s.broker.Publish(pubsub.Updated, *updated)

	return nil
}
