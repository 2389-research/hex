// ABOUTME: Tests for context pruning and token management
// ABOUTME: Ensures context stays within token limits while preserving important messages
package context

import (
	"testing"

	"github.com/harper/pagent/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "short text",
			text:     "Hello world",
			expected: 2, // ~11 chars / 4 = 2.75 -> 3, but our heuristic may vary
		},
		{
			name:     "longer text",
			text:     "This is a longer piece of text that should have more tokens",
			expected: 15, // ~61 chars / 4 = 15.25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := EstimateTokens(tt.text)
			// Allow some margin since it's an estimate
			assert.InDelta(t, tt.expected, tokens, 5)
		})
	}
}

func TestEstimateMessageTokens(t *testing.T) {
	msg := core.Message{
		Role:    "user",
		Content: "Hello, how are you today?",
	}

	tokens := EstimateMessageTokens(msg)
	// Should include role overhead + content
	assert.Greater(t, tokens, 0)
	assert.Less(t, tokens, 100) // Should be a small message
}

func TestEstimateMessagesTokens(t *testing.T) {
	messages := []core.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "What's the weather?"},
		{Role: "assistant", Content: "I don't have access to weather data."},
	}

	tokens := EstimateMessagesTokens(messages)
	assert.Greater(t, tokens, 0)
	// Should be roughly 3 messages * ~10 tokens each
	assert.InDelta(t, 30, tokens, 20)
}

func TestPruneContext_NoNeedToPrune(t *testing.T) {
	messages := []core.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	pruned := PruneContext(messages, 1000)
	assert.Equal(t, len(messages), len(pruned))
	assert.Equal(t, messages, pruned)
}

func TestPruneContext_EmptyMessages(t *testing.T) {
	messages := []core.Message{}
	pruned := PruneContext(messages, 1000)
	assert.Empty(t, pruned)
}

func TestPruneContext_SingleMessage(t *testing.T) {
	messages := []core.Message{
		{Role: "user", Content: "Hello"},
	}
	pruned := PruneContext(messages, 1000)
	assert.Equal(t, messages, pruned)
}

func TestPruneContext_KeepsSystemMessage(t *testing.T) {
	messages := []core.Message{
		{Role: "system", Content: "You are a helpful assistant with special instructions"},
		{Role: "user", Content: "This is the first message with some content"},
		{Role: "assistant", Content: "This is the first response with some content"},
		{Role: "user", Content: "This is the second message with some content"},
		{Role: "assistant", Content: "This is the second response with some content"},
		{Role: "user", Content: "This is the third message with some content"},
		{Role: "assistant", Content: "This is the third response with some content"},
	}

	// Set very low limit to force pruning
	pruned := PruneContext(messages, 80)

	// Should keep system message
	assert.Equal(t, "system", pruned[0].Role)
	assert.Contains(t, pruned[0].Content, "helpful assistant")

	// Should keep recent messages
	assert.Greater(t, len(pruned), 1)
	assert.Less(t, len(pruned), len(messages))
}

func TestPruneContext_KeepsRecentMessages(t *testing.T) {
	messages := []core.Message{
		{Role: "user", Content: "Old message 1"},
		{Role: "assistant", Content: "Old response 1"},
		{Role: "user", Content: "Recent message"},
		{Role: "assistant", Content: "Recent response"},
	}

	// Set limit to keep only recent messages
	pruned := PruneContext(messages, 30)

	// Should keep the most recent exchange
	assert.GreaterOrEqual(t, len(pruned), 2)
	lastMsg := pruned[len(pruned)-1]
	assert.Equal(t, "Recent response", lastMsg.Content)
}

func TestPruneContext_PreservesToolCalls(t *testing.T) {
	messages := []core.Message{
		{Role: "user", Content: "Message 1"},
		{Role: "assistant", Content: "Response 1"},
		{Role: "user", Content: "Use a tool"},
		{
			Role:    "assistant",
			Content: "Using tool",
			ToolCalls: []core.ToolUse{
				{Name: "read_file", ID: "tool-1"},
			},
		},
		{Role: "user", Content: "Latest message"},
	}

	// Prune to force decision
	pruned := PruneContext(messages, 50)

	// Should preserve message with tool calls
	hasToolCall := false
	for _, msg := range pruned {
		if len(msg.ToolCalls) > 0 {
			hasToolCall = true
			break
		}
	}
	assert.True(t, hasToolCall, "Should preserve messages with tool calls")
}

func TestPruneContext_VeryLowLimit(t *testing.T) {
	messages := []core.Message{
		{Role: "system", Content: "System"},
		{Role: "user", Content: "Hello"},
	}

	// Even with very low limit, should keep at least system + one message
	pruned := PruneContext(messages, 1)
	assert.GreaterOrEqual(t, len(pruned), 1)
}

func TestNewManager(t *testing.T) {
	m := NewManager(100000)
	require.NotNil(t, m)
	assert.Equal(t, 100000, m.MaxTokens)
}

func TestManager_ShouldPrune(t *testing.T) {
	m := NewManager(100)

	messages := []core.Message{
		{Role: "user", Content: "Short"},
	}

	assert.False(t, m.ShouldPrune(messages))

	// Add many messages to exceed limit
	longMessages := make([]core.Message, 50)
	for i := range longMessages {
		longMessages[i] = core.Message{
			Role:    "user",
			Content: "This is a reasonably long message that will add up",
		}
	}

	assert.True(t, m.ShouldPrune(longMessages))
}

func TestManager_Prune(t *testing.T) {
	m := NewManager(100)

	messages := []core.Message{
		{Role: "system", Content: "System"},
		{Role: "user", Content: "Message 1"},
		{Role: "assistant", Content: "Response 1"},
		{Role: "user", Content: "Message 2"},
		{Role: "assistant", Content: "Response 2"},
	}

	pruned := m.Prune(messages)
	assert.LessOrEqual(t, EstimateMessagesTokens(pruned), m.MaxTokens)
}

func TestManager_GetUsage(t *testing.T) {
	m := NewManager(1000)

	messages := []core.Message{
		{Role: "user", Content: "Hello world"},
	}

	usage := m.GetUsage(messages)
	assert.Greater(t, usage.EstimatedTokens, 0)
	assert.Greater(t, usage.PercentUsed, 0.0)
	assert.Less(t, usage.PercentUsed, 100.0)
	assert.False(t, usage.NearLimit)
}

func TestManager_GetUsage_NearLimit(t *testing.T) {
	m := NewManager(100)

	// Create messages that use ~95% of limit
	longMessages := make([]core.Message, 30)
	for i := range longMessages {
		longMessages[i] = core.Message{
			Role:    "user",
			Content: "This is a long message",
		}
	}

	usage := m.GetUsage(longMessages)
	assert.True(t, usage.NearLimit, "Should be near limit at 90%+")
	assert.Greater(t, usage.PercentUsed, 90.0)
}
