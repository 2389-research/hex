// ABOUTME: OpenRouter provider wraps OpenAI provider with different endpoint
// ABOUTME: OpenRouter is OpenAI-compatible proxy supporting multiple providers
package openrouter

import (
	"context"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/providers/openai"
)

const (
	defaultBaseURL = "https://openrouter.ai/api/v1/chat/completions"
)

// Provider implements the Provider interface for OpenRouter
// OpenRouter uses the same API format as OpenAI, just a different endpoint
type Provider struct {
	openaiProvider *openai.Provider
}

// NewProvider creates a new OpenRouter provider
func NewProvider(config providers.ProviderConfig) *Provider {
	// OpenRouter uses same format as OpenAI, just different endpoint
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{
		openaiProvider: openai.NewProvider(config),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "openrouter"
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig(cfg providers.ProviderConfig) error {
	// Delegate to OpenAI validation
	return p.openaiProvider.ValidateConfig(cfg)
}

// CreateMessage creates a synchronous request via OpenRouter
func (p *Provider) CreateMessage(ctx context.Context, req core.MessageRequest) (*core.MessageResponse, error) {
	// Delegate to OpenAI provider (API is identical)
	return p.openaiProvider.CreateMessage(ctx, req)
}

// CreateMessageStream creates a streaming request via OpenRouter
func (p *Provider) CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error) {
	// Delegate to OpenAI provider (API is identical)
	return p.openaiProvider.CreateMessageStream(ctx, req)
}
