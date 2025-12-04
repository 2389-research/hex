// ABOUTME: Service interface for message management
// ABOUTME: Handles message storage and retrieval with event publishing

package services

import (
	"context"

	"github.com/2389-research/hex/internal/pubsub"
)

// MessageService manages message lifecycle
type MessageService interface {
	pubsub.Subscriber[Message]

	// Add stores a new message
	Add(ctx context.Context, msg *Message) error

	// GetByConversation returns all messages for a conversation
	GetByConversation(ctx context.Context, convID string) ([]*Message, error)

	// GetSummaries returns only summary messages for a conversation
	GetSummaries(ctx context.Context, convID string) ([]*Message, error)
}
