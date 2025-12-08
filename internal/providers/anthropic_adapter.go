// ABOUTME: Adapter that wraps core.Client to implement Provider interface
// ABOUTME: Temporary bridge until core.Client is fully refactored into providers package
package providers

import (
	"context"
	"fmt"

	"github.com/2389-research/hex/internal/core"
)

// AnthropicAdapter wraps core.Client to implement the Provider interface
type AnthropicAdapter struct {
	client *core.Client
}

// NewAnthropicAdapter creates a new Anthropic provider from a core.Client
func NewAnthropicAdapter(client *core.Client) *AnthropicAdapter {
	return &AnthropicAdapter{
		client: client,
	}
}

// CreateMessage implements Provider.CreateMessage
func (a *AnthropicAdapter) CreateMessage(ctx context.Context, req core.MessageRequest) (*core.MessageResponse, error) {
	return a.client.CreateMessage(ctx, req)
}

// CreateMessageStream implements Provider.CreateMessageStream
func (a *AnthropicAdapter) CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error) {
	return a.client.CreateMessageStream(ctx, req)
}

// Name implements Provider.Name
func (a *AnthropicAdapter) Name() string {
	return "anthropic"
}

// ValidateConfig implements Provider.ValidateConfig
func (a *AnthropicAdapter) ValidateConfig(cfg ProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("anthropic: API key is required")
	}
	return nil
}
