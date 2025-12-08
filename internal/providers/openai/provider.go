// ABOUTME: OpenAI provider implementation using streaming chat completions API
// ABOUTME: Translates between OpenAI's format and hex's universal core.StreamChunk format
package openai

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
		httpClient: &http.Client{
			// No timeout - rely on request context for cancellation
		},
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

// CreateMessage creates a synchronous (non-streaming) request to OpenAI
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
	var inputTokens, outputTokens int

	for chunk := range streamChan {
		if chunk.Delta != nil && chunk.Delta.Text != "" {
			textContent.WriteString(chunk.Delta.Text)
		}

		// Extract usage if present (typically in final chunk)
		if chunk.Usage != nil {
			inputTokens = chunk.Usage.InputTokens
			outputTokens = chunk.Usage.OutputTokens
		}

		// Capture stop reason
		if chunk.Done {
			stopReason = "end_turn"
			break // Exit early - no need to read more
		}
	}

	// Build response
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
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
		},
	}

	return response, nil
}

// CreateMessageStream creates a streaming request to OpenAI
func (p *Provider) CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error) {
	// Translate to OpenAI format
	openaiReq := translateRequest(req)

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

	// Start goroutine to parse SSE stream
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

			// OpenAI SSE format: "data: {...}"
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				// Send completion event
				select {
				case chunks <- &core.StreamChunk{Type: "message_stop", Done: true}:
				case <-ctx.Done():
				}
				return
			}

			var streamResp openaiStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
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

// openaiRequest represents OpenAI chat completion request
type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Tools    []openaiTool    `json:"tools,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiTool struct {
	Type     string             `json:"type"`
	Function openaiToolFunction `json:"function"`
}

type openaiToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// translateRequest converts core.MessageRequest to OpenAI format
func translateRequest(req core.MessageRequest) openaiRequest {
	openaiReq := openaiRequest{
		Model:    req.Model,
		Stream:   true,
		Messages: make([]openaiMessage, len(req.Messages)),
	}

	// Translate messages
	for i, msg := range req.Messages {
		openaiReq.Messages[i] = openaiMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Translate tools if present
	if len(req.Tools) > 0 {
		openaiReq.Tools = make([]openaiTool, len(req.Tools))
		for i, tool := range req.Tools {
			openaiReq.Tools[i] = openaiTool{
				Type: "function",
				Function: openaiToolFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			}
		}
	}

	return openaiReq
}

// openaiStreamResponse represents OpenAI streaming response
type openaiStreamResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []openaiStreamChoice `json:"choices"`
}

type openaiStreamChoice struct {
	Index        int               `json:"index"`
	Delta        openaiStreamDelta `json:"delta"`
	FinishReason *string           `json:"finish_reason"`
}

type openaiStreamDelta struct {
	Role      string                 `json:"role,omitempty"`
	Content   string                 `json:"content,omitempty"`
	ToolCalls []openaiStreamToolCall `json:"tool_calls,omitempty"`
}

type openaiStreamToolCall struct {
	Index    int                      `json:"index"`
	ID       string                   `json:"id,omitempty"`
	Type     string                   `json:"type,omitempty"`
	Function openaiStreamToolFunction `json:"function,omitempty"`
}

type openaiStreamToolFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// translateStreamResponse converts OpenAI stream chunk to core.StreamChunk
func translateStreamResponse(resp *openaiStreamResponse) *core.StreamChunk {
	if len(resp.Choices) == 0 {
		return nil
	}

	choice := resp.Choices[0]

	// Check for completion
	if choice.FinishReason != nil && *choice.FinishReason != "" {
		return &core.StreamChunk{
			Type: "message_stop",
			Done: true,
		}
	}

	// Handle text content
	if choice.Delta.Content != "" {
		return &core.StreamChunk{
			Type: "content_block_delta",
			Delta: &core.Delta{
				Type: "text_delta",
				Text: choice.Delta.Content,
			},
		}
	}

	// Handle tool calls (simplified - OpenAI tool calling is more complex)
	// For now, we'll just handle text content
	// Full tool support would require accumulating tool call deltas

	return nil
}
