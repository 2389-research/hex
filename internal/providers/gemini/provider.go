// ABOUTME: Google Gemini provider implementation using streaming generateContent API
// ABOUTME: Translates between Gemini's contents/parts format and hex's universal core.StreamChunk format
package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/providers"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// Provider implements the Provider interface for Google Gemini
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
		httpClient: &http.Client{
			// No timeout - rely on request context for cancellation
		},
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

// CreateMessage creates a synchronous (non-streaming) request to Gemini
// Implemented by collecting the streaming response into a complete message
func (p *Provider) CreateMessage(ctx context.Context, req core.MessageRequest) (*core.MessageResponse, error) {
	// Create cancellable context for the stream
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel() // Ensure goroutine is signaled to stop

	// Use streaming internally and collect results
	streamChan, err := p.CreateMessageStream(streamCtx, req)
	if err != nil {
		return nil, err
	}

	// Always drain the channel, even on early exit
	defer func() {
		cancel() // Signal goroutine to stop
		for range streamChan {
			// Drain remaining chunks
		}
	}()

	// Collect all chunks into a complete response
	var textContent strings.Builder
	var stopReason string

	for chunk := range streamChan {
		if chunk.Delta != nil && chunk.Delta.Text != "" {
			textContent.WriteString(chunk.Delta.Text)
		}

		// Capture stop reason
		if chunk.Done {
			stopReason = "end_turn"
			break // Exit early - no need to read more
		}
	}

	// Build response (Gemini doesn't provide token counts in streaming)
	response := &core.MessageResponse{
		ID:   fmt.Sprintf("msg_%d", time.Now().Unix()),
		Type: "message",
		Role: "assistant",
		Content: []core.Content{
			{
				Type: "text",
				Text: textContent.String(),
			},
		},
		Model:      req.Model,
		StopReason: stopReason,
		Usage: core.Usage{
			InputTokens:  0, // Gemini doesn't provide these in streaming
			OutputTokens: 0,
		},
	}

	return response, nil
}

// CreateMessageStream creates a streaming request to Gemini
func (p *Provider) CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error) {
	// Translate to Gemini format
	geminiReq := translateRequest(req)

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Gemini uses model in URL path and API key as query param
	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s",
		p.config.BaseURL, req.Model, p.config.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	//nolint:bodyclose // Body is closed in goroutine defer
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Create channel for chunks
	chunks := make(chan *core.StreamChunk, 10)

	// Start goroutine to parse streaming JSON
	go func() {
		defer func() { _ = resp.Body.Close() }() // Close body first (runs last)
		defer close(chunks)                      // Close channel after body closed

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			// Check context cancellation at top of loop
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			if line == "" {
				continue
			}

			var streamResp geminiStreamResponse
			if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
				continue // Skip malformed lines
			}

			// Translate to core.StreamChunk
			chunk := translateStreamResponse(&streamResp)
			if chunk != nil {
				select {
				case chunks <- chunk:
				case <-ctx.Done():
					return
				}

				if chunk.Done {
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errorChunk := &core.StreamChunk{
				Type: "error",
				Done: true,
				Delta: &core.Delta{
					Type: "error",
					Text: fmt.Sprintf("Stream error: %v", err),
				},
			}
			select {
			case chunks <- errorChunk:
			case <-ctx.Done():
			}
		}
	}()

	return chunks, nil
}

// geminiRequest represents Gemini generateContent request
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

// translateRequest converts core.MessageRequest to Gemini format
func translateRequest(req core.MessageRequest) geminiRequest {
	geminiReq := geminiRequest{
		Contents: make([]geminiContent, len(req.Messages)),
	}

	for i, msg := range req.Messages {
		// Gemini uses "model" instead of "assistant"
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		geminiReq.Contents[i] = geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: msg.Content},
			},
		}
	}

	return geminiReq
}

// geminiStreamResponse represents Gemini streaming response
type geminiStreamResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

// translateStreamResponse converts Gemini stream chunk to core.StreamChunk
func translateStreamResponse(resp *geminiStreamResponse) *core.StreamChunk {
	if len(resp.Candidates) == 0 {
		return nil
	}

	candidate := resp.Candidates[0]

	// Check for completion
	if candidate.FinishReason != "" {
		return &core.StreamChunk{
			Type: "message_stop",
			Done: true,
		}
	}

	// Extract text from parts
	if len(candidate.Content.Parts) > 0 {
		text := candidate.Content.Parts[0].Text
		if text != "" {
			return &core.StreamChunk{
				Type: "content_block_delta",
				Delta: &core.Delta{
					Type: "text_delta",
					Text: text,
				},
			}
		}
	}

	return nil
}
