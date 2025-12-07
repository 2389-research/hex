// Package core provides the Anthropic API client and core conversation functionality.
// ABOUTME: SSE streaming support for Anthropic API
// ABOUTME: Parses server-sent events, accumulates deltas, yields chunks
package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ParseSSEChunk parses a single SSE data line into a StreamChunk
func ParseSSEChunk(data string) (*StreamChunk, error) {
	// SSE format: "data: {...json...}"
	if !strings.HasPrefix(data, "data: ") {
		return nil, nil // Ignore non-data lines
	}

	jsonData := strings.TrimPrefix(data, "data: ")
	if jsonData == "[DONE]" {
		return &StreamChunk{Type: "message_stop", Done: true}, nil
	}

	var chunk StreamChunk
	if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
		return nil, fmt.Errorf("parse chunk: %w", err)
	}

	return &chunk, nil
}

// StreamAccumulator accumulates text deltas from streaming chunks
type StreamAccumulator struct {
	text string
}

// NewStreamAccumulator creates a new accumulator
func NewStreamAccumulator() *StreamAccumulator {
	return &StreamAccumulator{}
}

// Add accumulates a chunk's text delta
func (a *StreamAccumulator) Add(chunk *StreamChunk) {
	if chunk.Delta != nil && chunk.Delta.Text != "" {
		a.text += chunk.Delta.Text
	}
}

// GetText returns the accumulated text
func (a *StreamAccumulator) GetText() string {
	return a.text
}

// Reset clears the accumulated text
func (a *StreamAccumulator) Reset() {
	a.text = ""
}

// CreateMessageStream sends a streaming request and returns a channel of chunks
func (c *Client) CreateMessageStream(ctx context.Context, req MessageRequest) (<-chan *StreamChunk, error) {
	req.Stream = true

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/messages",
		strings.NewReader(string(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq) //nolint:bodyclose // Body is closed in goroutine below
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		_ = httpResp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", httpResp.StatusCode, string(body))
	}

	// Create channel for chunks with configurable buffer size
	chunks := make(chan *StreamChunk, c.streamBufferSize)

	// Start goroutine to read SSE stream
	go func() {
		defer close(chunks)
		defer func() { _ = httpResp.Body.Close() }()

		scanner := bufio.NewScanner(httpResp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue // Skip empty lines
			}

			chunk, err := ParseSSEChunk(line)
			if err != nil {
				// Send error chunk so consumer can handle parsing errors
				errorChunk := &StreamChunk{
					Type: "error",
					Done: true,
					Delta: &Delta{
						Type: "error",
						Text: fmt.Sprintf("SSE parse error: %v", err),
					},
				}
				select {
				case chunks <- errorChunk:
				case <-ctx.Done():
					return
				}
				return
			}
			if chunk == nil {
				continue // Ignore non-data lines
			}

			select {
			case chunks <- chunk:
			case <-ctx.Done():
				return
			}

			if chunk.Done {
				return
			}
		}

		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			errorChunk := &StreamChunk{
				Type: "error",
				Done: true,
				Delta: &Delta{
					Type: "error",
					Text: fmt.Sprintf("Stream scanner error: %v", err),
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
