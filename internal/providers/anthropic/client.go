// ABOUTME: Anthropic API client implementation of Provider interface
// ABOUTME: Handles Anthropic-specific message format and streaming
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/ratelimit"
)

const (
	defaultBaseURL = "https://api.anthropic.com/v1/messages"
	apiVersion     = "2023-06-01"
)

// Global rate limiter shared across all clients to prevent 429 errors
// Anthropic API allows ~50 requests per minute (Tier 1).
var globalLimiter = ratelimit.NewSharedLimiter(50, time.Minute)

// Provider implements the Provider interface for Anthropic
type Provider struct {
	config     providers.ProviderConfig
	httpClient *http.Client
}

// NewProvider creates a new Anthropic provider
func NewProvider(config providers.ProviderConfig) *Provider {
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "anthropic"
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig(cfg providers.ProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("anthropic: API key is required")
	}
	return nil
}

// CreateStream creates a streaming request to Anthropic
func (p *Provider) CreateStream(ctx context.Context, req *providers.MessageRequest) (providers.Stream, error) {
	// Acquire rate limit token before making API call
	if err := globalLimiter.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	// Translate to Anthropic format
	anthropicReq := translateRequest(req)

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return newStream(resp.Body), nil
}
