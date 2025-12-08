// ABOUTME: OpenRouter provider wraps OpenAI format with different endpoint
// ABOUTME: OpenRouter is OpenAI-compatible proxy supporting multiple providers
package openrouter

import (
	"context"
	"fmt"

	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/providers/openai"
)

const (
	defaultBaseURL = "https://openrouter.ai/api/v1/chat/completions"
)

// Provider implements the Provider interface for OpenRouter
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
	if cfg.APIKey == "" {
		return fmt.Errorf("openrouter: API key is required")
	}
	return nil
}

// CreateStream delegates to OpenAI provider (OpenRouter is compatible)
func (p *Provider) CreateStream(ctx context.Context, req *providers.MessageRequest) (providers.Stream, error) {
	// Model IDs in OpenRouter use provider/model format (e.g., "anthropic/claude-sonnet-4-5")
	// No translation needed, OpenRouter handles routing
	return p.openaiProvider.CreateStream(ctx, req)
}
