// ABOUTME: Service interface for LLM agent interactions
// ABOUTME: Handles message queuing, execution, and conversation state

package services

import "context"

// AgentService coordinates LLM interactions with queuing support
type AgentService interface {
	// Run executes a prompt (queues if conversation is busy)
	Run(ctx context.Context, call AgentCall) (*AgentResult, error)

	// Stream executes a prompt with streaming response
	Stream(ctx context.Context, call AgentCall) (<-chan StreamEvent, error)

	// IsConversationBusy returns true if conversation has active request
	IsConversationBusy(convID string) bool

	// QueuedPrompts returns number of queued messages for conversation
	QueuedPrompts(convID string) int

	// CancelConversation cancels active request and clears queue
	CancelConversation(convID string)
}
