// ABOUTME: Provider interface defines the contract for all LLM provider implementations
// ABOUTME: Each provider translates between hex's universal format and provider-specific APIs
package providers

import (
	"context"
)

// Provider defines the interface all LLM providers must implement
type Provider interface {
	// CreateStream sends a message request and returns a streaming response
	CreateStream(ctx context.Context, req *MessageRequest) (Stream, error)

	// Name returns the provider's identifier (e.g., "anthropic", "openai")
	Name() string

	// ValidateConfig checks if the provider configuration is valid
	ValidateConfig(cfg ProviderConfig) error
}

// Stream represents a streaming response from a provider
type Stream interface {
	// Next returns the next chunk in the stream
	Next() (*StreamChunk, error)

	// Close closes the stream and releases resources
	Close() error
}

// MessageRequest represents a universal message request format
type MessageRequest struct {
	Model       string
	Messages    []Message
	Tools       []Tool
	MaxTokens   int
	Stream      bool
	Temperature float64
}

// Message represents a single message in the conversation
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// Tool represents a tool that can be called by the model
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Type    string // "message_start", "content_block_delta", "message_stop"
	Content string
	Done    bool
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	APIKey  string
	BaseURL string
}
