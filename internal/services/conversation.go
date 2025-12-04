// ABOUTME: Service interface for conversation management
// ABOUTME: Handles CRUD operations and token tracking for conversations

package services

import (
	"context"

	"github.com/2389-research/hex/internal/pubsub"
)

// ConversationService manages conversation lifecycle and state
type ConversationService interface {
	pubsub.Subscriber[Conversation]

	// Create creates a new conversation
	Create(ctx context.Context, title string) (*Conversation, error)

	// Get retrieves a conversation by ID
	Get(ctx context.Context, id int64) (*Conversation, error)

	// List returns all conversations ordered by updated_at DESC
	List(ctx context.Context) ([]*Conversation, error)

	// Update saves conversation changes
	Update(ctx context.Context, conv *Conversation) error

	// Delete removes a conversation and its messages
	Delete(ctx context.Context, id int64) error

	// UpdateTokenUsage updates token counts and calculates cost
	UpdateTokenUsage(ctx context.Context, id int64, promptTokens, completionTokens int64) error
}
