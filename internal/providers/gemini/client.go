// ABOUTME: Google Gemini API client implementation of Provider interface
// ABOUTME: Handles streamGenerateContent API with model ID in URL path
package gemini

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
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1"
)

// Provider implements the Provider interface for Gemini
type Provider struct {
	config     providers.ProviderConfig
	httpClient *http.Client
}

// NewProvider creates a new Gemini provider
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
	return "gemini"
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig(cfg providers.ProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("gemini: API key is required")
	}
	return nil
}

// CreateStream creates a streaming request to Gemini
func (p *Provider) CreateStream(ctx context.Context, req *providers.MessageRequest) (providers.Stream, error) {
	// Translate to Gemini format
	geminiReq := TranslateRequest(req)

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Model ID goes in URL path
	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s",
		p.config.BaseURL, req.Model, p.config.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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
