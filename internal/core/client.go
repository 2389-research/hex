// Package core provides the Anthropic API client and core conversation functionality.
// ABOUTME: HTTP client for Anthropic Messages API
// ABOUTME: Handles authentication, request formatting, and response parsing
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/2389-research/hex/internal/cost"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/ratelimit"
)

const (
	defaultBaseURL = "https://api.anthropic.com/v1"
	apiVersion     = "2023-06-01"
)

// Global rate limiter shared across all clients to prevent 429 errors
// Anthropic API allows ~50 requests per minute (Tier 1).
//
// NOTE: This limiter is never stopped and runs a background goroutine
// for the lifetime of the process. This is acceptable for a singleton
// that lives for the entire application lifecycle.
var globalLimiter = ratelimit.NewSharedLimiter(50, time.Minute)

// Client is the Anthropic API client
type Client struct {
	apiKey           string
	baseURL          string
	httpClient       *http.Client
	streamBufferSize int // Configurable buffer size for streaming channels
}

// ClientOption configures a Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithBaseURL sets a custom base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithStreamBufferSize sets the streaming channel buffer size
func WithStreamBufferSize(size int) ClientOption {
	return func(c *Client) {
		if size > 0 {
			c.streamBufferSize = size
		}
	}
}

// NewClient creates a new API client
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:           apiKey,
		baseURL:          defaultBaseURL,
		streamBufferSize: 10, // Default buffer size
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// CreateMessage sends a message to the API
func (c *Client) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
	// Acquire rate limit token before making API call
	if err := globalLimiter.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Debug logging: Log the full request
	if logging.IsDebugEnabled() {
		// Pretty-print the request for readability
		var prettyReq bytes.Buffer
		if err := json.Indent(&prettyReq, body, "", "  "); err == nil {
			logging.Debug("API Request to /messages", "body", prettyReq.String())
		} else {
			logging.Debug("API Request to /messages", "body", string(body))
		}
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/messages",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Debug logging: Log the full response
	if logging.IsDebugEnabled() {
		// Pretty-print the response for readability
		var prettyResp bytes.Buffer
		if err := json.Indent(&prettyResp, respBody, "", "  "); err == nil {
			logging.Debug("API Response from /messages", "status", httpResp.StatusCode, "body", prettyResp.String())
		} else {
			logging.Debug("API Response from /messages", "status", httpResp.StatusCode, "body", string(respBody))
		}
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Unmarshal response
	var resp MessageResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Debug logging: Log token usage
	logging.DebugIf("API Usage",
		"input_tokens", resp.Usage.InputTokens,
		"output_tokens", resp.Usage.OutputTokens,
		"stop_reason", resp.StopReason,
	)

	// Record cost tracking
	agentID := os.Getenv("HEX_AGENT_ID")
	parentID := os.Getenv("HEX_PARENT_AGENT_ID")

	if agentID != "" {
		// Convert core.Usage to cost.Usage
		costUsage := cost.Usage{
			InputTokens:      resp.Usage.InputTokens,
			OutputTokens:     resp.Usage.OutputTokens,
			CacheReadTokens:  resp.Usage.CacheReadTokens,
			CacheWriteTokens: resp.Usage.CacheWriteTokens,
		}
		if err := cost.Global().RecordUsage(agentID, parentID, req.Model, costUsage); err != nil {
			// Log error but don't fail the request
			logging.DebugIf("Cost tracking failed", "error", err)
		}
	}

	return &resp, nil
}

// GetTextContent extracts text content from response
func (r *MessageResponse) GetTextContent() string {
	for _, content := range r.Content {
		if content.Type == "text" {
			return content.Text
		}
	}
	return ""
}
