// ABOUTME: Tests for SSE streaming API client
// ABOUTME: Validates chunk parsing, delta accumulation, and stream completion
package core_test

import (
	"context"
	"os"
	"testing"

	"github.com/harper/jeff/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSSEChunk(t *testing.T) {
	data := `data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}`

	chunk, err := core.ParseSSEChunk(data)
	require.NoError(t, err)
	assert.Equal(t, "content_block_delta", chunk.Type)
	assert.Equal(t, "Hello", chunk.Delta.Text)
}

func TestParseSSEChunkDone(t *testing.T) {
	data := `data: [DONE]`

	chunk, err := core.ParseSSEChunk(data)
	require.NoError(t, err)
	assert.Equal(t, "message_stop", chunk.Type)
	assert.True(t, chunk.Done)
}

func TestStreamAccumulator(t *testing.T) {
	acc := core.NewStreamAccumulator()

	acc.Add(&core.StreamChunk{
		Type:  "content_block_delta",
		Delta: &core.Delta{Type: "text_delta", Text: "Hello "},
	})
	acc.Add(&core.StreamChunk{
		Type:  "content_block_delta",
		Delta: &core.Delta{Type: "text_delta", Text: "world"},
	})

	assert.Equal(t, "Hello world", acc.GetText())
}

func TestStreamAccumulatorIgnoresNonTextDeltas(t *testing.T) {
	acc := core.NewStreamAccumulator()

	acc.Add(&core.StreamChunk{
		Type: "message_start",
	})
	acc.Add(&core.StreamChunk{
		Type:  "content_block_delta",
		Delta: &core.Delta{Type: "text_delta", Text: "test"},
	})

	assert.Equal(t, "test", acc.GetText())
}

func TestCreateMessageStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping streaming test in short mode")
	}

	// Use real API key from environment when available
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping streaming test - ANTHROPIC_API_KEY not set")
	}

	client := core.NewClient(apiKey)
	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 100,
		Messages:  []core.Message{{Role: "user", Content: "Say hi"}},
		Stream:    true,
	}

	ctx := context.Background()
	stream, err := client.CreateMessageStream(ctx, req)
	require.NoError(t, err)

	chunks := make([]*core.StreamChunk, 0, 5) // Preallocate for typical stream size
	for chunk := range stream {
		chunks = append(chunks, chunk)
		if chunk.Type == "message_stop" {
			break
		}
	}

	assert.NotEmpty(t, chunks)
}
