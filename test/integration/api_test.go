// ABOUTME: API integration tests for Anthropic client interaction
// ABOUTME: Tests message creation, streaming, tool use flows (with VCR or mocks)

// Package integration provides end-to-end integration tests for Clem.
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/harper/jeff/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIClientInitialization tests basic client setup
func TestAPIClientInitialization(t *testing.T) {
	client := core.NewClient("test-api-key")
	assert.NotNil(t, client)
}

// TestAPIMessageRequest tests creating a basic message request
func TestAPIMessageRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping real API test")
	}

	client := core.NewClient(apiKey)

	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 100,
		Messages: []core.Message{
			{Role: "user", Content: "Say 'test successful' and nothing else"},
		},
	}

	ctx := context.Background()
	resp, err := client.CreateMessage(ctx, req)
	require.NoError(t, err, "API request should succeed")

	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "message", resp.Type)
	assert.Len(t, resp.Content, 1)
	assert.Contains(t, resp.Content[0].Text, "test successful")
}

// TestAPIStreamingFlow tests streaming message response
func TestAPIStreamingFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping streaming test in short mode")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping real API test")
	}

	client := core.NewClient(apiKey)

	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 50,
		Messages: []core.Message{
			{Role: "user", Content: "Count to 3"},
		},
		Stream: true,
	}

	ctx := context.Background()
	stream, err := client.CreateMessageStream(ctx, req)
	require.NoError(t, err, "stream creation should succeed")

	// Accumulate chunks
	acc := core.NewStreamAccumulator()
	chunkCount := 0

	for chunk := range stream {
		if chunk.Type == "content_block_delta" && chunk.Delta != nil {
			acc.Add(chunk)
			chunkCount++
		}
		if chunk.Type == "message_stop" || chunk.Done {
			break
		}
	}

	// Should have received multiple chunks
	assert.Greater(t, chunkCount, 0, "should receive at least one chunk")

	// Accumulated text should not be empty
	text := acc.GetText()
	assert.NotEmpty(t, text, "accumulated text should not be empty")

	t.Logf("Received %d chunks, accumulated text: %s", chunkCount, text)
}

// TestAPITokenCounting tests that token usage is reported
func TestAPITokenCounting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping real API test")
	}

	client := core.NewClient(apiKey)

	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 50,
		Messages: []core.Message{
			{Role: "user", Content: "Hi"},
		},
	}

	ctx := context.Background()
	resp, err := client.CreateMessage(ctx, req)
	require.NoError(t, err)

	// Verify usage is reported
	assert.Greater(t, resp.Usage.InputTokens, 0, "should report input tokens")
	assert.Greater(t, resp.Usage.OutputTokens, 0, "should report output tokens")

	t.Logf("Token usage - Input: %d, Output: %d", resp.Usage.InputTokens, resp.Usage.OutputTokens)
}

// TestAPIErrorHandling tests API error responses
func TestAPIErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	// Use invalid API key
	client := core.NewClient("invalid-api-key-12345")

	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 10,
		Messages: []core.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := client.CreateMessage(ctx, req)

	// Should get authentication error
	assert.Error(t, err, "should fail with invalid API key")
	assert.Contains(t, err.Error(), "401", "should be 401 Unauthorized")
}

// TestAPIContextCancellation tests that context cancellation stops requests
func TestAPIContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping real API test")
	}

	client := core.NewClient(apiKey)

	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 1000, // Large response
		Messages: []core.Message{
			{Role: "user", Content: "Write a long story"},
		},
	}

	// Create context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.CreateMessage(ctx, req)

	// Should fail due to context cancellation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestAPIWithSystemPrompt tests using system prompts
func TestAPIWithSystemPrompt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping real API test")
	}

	client := core.NewClient(apiKey)

	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 50,
		System:    "You are a pirate. Always respond like a pirate.",
		Messages: []core.Message{
			{Role: "user", Content: "Say hello"},
		},
	}

	ctx := context.Background()
	resp, err := client.CreateMessage(ctx, req)
	require.NoError(t, err)

	// Response should have pirate-like language
	text := resp.Content[0].Text
	assert.NotEmpty(t, text)

	t.Logf("Pirate response: %s", text)
	// Note: We can't reliably assert pirate language, but we verify the request works
}

// TestAPIMultiTurnConversation tests a multi-turn conversation
func TestAPIMultiTurnConversation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping real API test")
	}

	client := core.NewClient(apiKey)

	// First turn
	req1 := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 50,
		Messages: []core.Message{
			{Role: "user", Content: "My name is Alice"},
		},
	}

	ctx := context.Background()
	resp1, err := client.CreateMessage(ctx, req1)
	require.NoError(t, err)

	// Second turn - ask about the name
	req2 := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 50,
		Messages: []core.Message{
			{Role: "user", Content: "My name is Alice"},
			{Role: "assistant", Content: resp1.Content[0].Text},
			{Role: "user", Content: "What's my name?"},
		},
	}

	resp2, err := client.CreateMessage(ctx, req2)
	require.NoError(t, err)

	// Should remember the name
	text := resp2.Content[0].Text
	assert.Contains(t, text, "Alice", "should remember the name from conversation history")
}

// TestStreamAccumulator tests the stream accumulator utility
func TestStreamAccumulator(t *testing.T) {
	acc := core.NewStreamAccumulator()

	// Add chunks
	acc.Add(&core.StreamChunk{
		Type:  "content_block_delta",
		Delta: &core.Delta{Type: "text_delta", Text: "Hello "},
	})
	acc.Add(&core.StreamChunk{
		Type:  "content_block_delta",
		Delta: &core.Delta{Type: "text_delta", Text: "world"},
	})
	acc.Add(&core.StreamChunk{
		Type:  "content_block_delta",
		Delta: &core.Delta{Type: "text_delta", Text: "!"},
	})

	// Should accumulate correctly
	assert.Equal(t, "Hello world!", acc.GetText())
}

// TestStreamChunkParsing tests parsing SSE chunks
func TestStreamChunkParsing(t *testing.T) {
	// Test valid data chunk
	data := `data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}`
	chunk, err := core.ParseSSEChunk(data)
	require.NoError(t, err)
	assert.Equal(t, "content_block_delta", chunk.Type)
	assert.Equal(t, "Hello", chunk.Delta.Text)

	// Test [DONE] chunk
	doneData := `data: [DONE]`
	doneChunk, err := core.ParseSSEChunk(doneData)
	require.NoError(t, err)
	assert.True(t, doneChunk.Done)

	// Test non-data line (should be ignored)
	nonData := `: comment line`
	nilChunk, err := core.ParseSSEChunk(nonData)
	require.NoError(t, err)
	assert.Nil(t, nilChunk)
}
