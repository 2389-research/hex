// ABOUTME: Tests for message summarization using Claude API
// ABOUTME: Verifies summary generation and caching behavior
package context

import (
	"context"
	"testing"

	"github.com/harper/clem/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSummarizer(t *testing.T) {
	client := core.NewClient("test-key")
	s := NewSummarizer(client)
	require.NotNil(t, s)
	assert.NotNil(t, s.client)
	assert.NotNil(t, s.cache)
}

func TestSummarizer_GenerateCacheKey(t *testing.T) {
	messages := []core.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	s := &Summarizer{cache: make(map[string]string)}
	key := s.generateCacheKey(messages)

	// Should generate consistent keys
	key2 := s.generateCacheKey(messages)
	assert.Equal(t, key, key2)

	// Different messages should have different keys
	messages2 := []core.Message{
		{Role: "user", Content: "Different"},
	}
	key3 := s.generateCacheKey(messages2)
	assert.NotEqual(t, key, key3)
}

func TestSummarizer_FormatMessages(t *testing.T) {
	messages := []core.Message{
		{Role: "user", Content: "What's the weather?"},
		{Role: "assistant", Content: "I don't have weather data."},
		{Role: "user", Content: "Tell me a joke"},
	}

	s := &Summarizer{}
	formatted := s.formatMessages(messages)

	assert.Contains(t, formatted, "user:")
	assert.Contains(t, formatted, "assistant:")
	assert.Contains(t, formatted, "What's the weather?")
	assert.Contains(t, formatted, "I don't have weather data.")
	assert.Contains(t, formatted, "Tell me a joke")
}

func TestSummarizer_SummarizeMessages_EmptyMessages(t *testing.T) {
	client := core.NewClient("test-key")
	s := NewSummarizer(client)

	summary, err := s.SummarizeMessages(context.Background(), []core.Message{})
	assert.NoError(t, err)
	assert.Equal(t, "", summary)
}

func TestSummarizer_SummarizeMessages_CacheHit(t *testing.T) {
	client := core.NewClient("test-key")
	s := NewSummarizer(client)

	messages := []core.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	// Pre-populate cache
	key := s.generateCacheKey(messages)
	s.cache[key] = "Cached summary"

	summary, err := s.SummarizeMessages(context.Background(), messages)
	assert.NoError(t, err)
	assert.Equal(t, "Cached summary", summary)
}

func TestSummarizer_CreateSummaryMessage(t *testing.T) {
	summary := "User asked about weather. Assistant said no data available."
	msg := CreateSummaryMessage(summary)

	assert.Equal(t, "system", msg.Role)
	assert.Contains(t, msg.Content, summary)
	assert.Contains(t, msg.Content, "Previous conversation summary")
}

func TestSummarizer_ClearCache(t *testing.T) {
	client := core.NewClient("test-key")
	s := NewSummarizer(client)

	// Add some cache entries
	s.cache["key1"] = "summary1"
	s.cache["key2"] = "summary2"

	assert.Len(t, s.cache, 2)

	s.ClearCache()
	assert.Len(t, s.cache, 0)
}

// Note: We don't test actual API calls here since that requires
// a real API key and would be slow/expensive. Those would be
// integration tests run separately.
