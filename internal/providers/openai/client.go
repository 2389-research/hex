// ABOUTME: OpenAI API client implementation of Provider interface
// ABOUTME: Handles Chat Completions API with streaming support
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/2389-research/hex/internal/providers"
)

const (
	defaultBaseURL = "https://api.openai.com/v1/chat/completions"
)

// Provider implements the Provider interface for OpenAI
type Provider struct {
	config     providers.ProviderConfig
	httpClient *http.Client
}

// NewProvider creates a new OpenAI provider
func NewProvider(config providers.ProviderConfig) *Provider {
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{
		config:     config,
		httpClient: &http.Client{},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "openai"
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig(cfg providers.ProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("openai: API key is required")
	}
	return nil
}

// CreateStream creates a streaming request to OpenAI
func (p *Provider) CreateStream(ctx context.Context, req *providers.MessageRequest) (providers.Stream, error) {
	// Translate to OpenAI format
	openaiReq := TranslateRequest(req)

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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
