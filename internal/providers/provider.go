// ABOUTME: Provider interface defines the contract for all LLM provider implementations
// ABOUTME: Each provider translates between hex's universal format and provider-specific APIs
package providers

import (
	"context"

	"github.com/2389-research/hex/internal/core"
)

// Provider defines the interface all LLM providers must implement
// This matches the orchestrator.APIClient interface for easy integration
type Provider interface {
	// CreateMessageStream sends a message request and returns a streaming response
	CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error)

	// Name returns the provider's identifier (e.g., "anthropic", "openai")
	Name() string

	// ValidateConfig checks if the provider configuration is valid
	ValidateConfig(cfg ProviderConfig) error
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	APIKey  string
	BaseURL string
}
