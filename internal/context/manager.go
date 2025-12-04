// ABOUTME: Context pruning and token management for conversations
// ABOUTME: Keeps conversations within token limits while preserving important context

// Package context provides context window management and message pruning for conversations.
package context

import (
	"github.com/harper/jeff/internal/core"
)

const (
	// Token estimation heuristic: ~4 chars per token
	charsPerToken = 4

	// Overhead tokens per message (role, formatting, etc)
	messageOverhead = 4

	// Default threshold for "near limit" warning (90%)
	nearLimitThreshold = 0.9
)

// Manager handles context pruning and token estimation
type Manager struct {
	MaxTokens int
	Strategy  PruneStrategy
}

// PruneStrategy defines how to handle context pruning
type PruneStrategy int

const (
	// StrategyKeepAll preserves all messages without pruning
	StrategyKeepAll PruneStrategy = iota
	// StrategyPrune removes old messages when context limit is reached
	StrategyPrune
	// StrategySummarize summarizes old messages to save tokens
	StrategySummarize
)

// ContextUsage provides information about context token usage
//
//nolint:revive // Explicit name improves clarity and searchability
type ContextUsage struct {
	EstimatedTokens int
	MaxTokens       int
	PercentUsed     float64
	NearLimit       bool
}

// NewManager creates a new context manager
func NewManager(maxTokens int) *Manager {
	return &Manager{
		MaxTokens: maxTokens,
		Strategy:  StrategyPrune,
	}
}

// EstimateTokens estimates tokens in a string using a simple heuristic
// This is a rough estimate: actual token count depends on tokenizer
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return (len(text) + charsPerToken - 1) / charsPerToken
}

// EstimateMessageTokens estimates tokens for a single message
func EstimateMessageTokens(msg core.Message) int {
	tokens := messageOverhead
	tokens += EstimateTokens(msg.Content)

	// Add tokens for tool calls if present
	for _, tool := range msg.ToolCalls {
		tokens += EstimateTokens(tool.Name)
		// Rough estimate for input params
		tokens += 10
	}

	return tokens
}

// EstimateMessagesTokens estimates total tokens for a slice of messages
func EstimateMessagesTokens(messages []core.Message) int {
	total := 0
	for _, msg := range messages {
		total += EstimateMessageTokens(msg)
	}
	return total
}

// PruneContext prunes messages to fit within maxTokens
// Strategy:
// - Always keep first message if it's a system message
// - Keep most recent messages
// - Preserve messages with tool calls
// - Remove middle messages if needed
func PruneContext(messages []core.Message, maxTokens int) []core.Message {
	if len(messages) == 0 {
		return messages
	}

	currentTokens := EstimateMessagesTokens(messages)
	if currentTokens <= maxTokens {
		return messages
	}

	// If only one message, keep it even if over limit
	if len(messages) == 1 {
		return messages
	}

	var result []core.Message
	tokensUsed := 0

	// Step 1: Always keep system message if present
	systemIdx := -1
	if messages[0].Role == "system" {
		result = append(result, messages[0])
		tokensUsed += EstimateMessageTokens(messages[0])
		systemIdx = 0
	}

	// Step 2: Identify important messages (with tool calls)
	importantIndices := make(map[int]bool)
	for i, msg := range messages {
		if i == systemIdx {
			continue
		}
		if len(msg.ToolCalls) > 0 {
			importantIndices[i] = true
		}
	}

	// Step 3: Add recent messages from the end, working backwards
	// We want to keep at least 2-4 recent messages if possible
	recentCount := 0
	maxRecent := 4
	budget := maxTokens - tokensUsed

	// First pass: add recent messages
	for i := len(messages) - 1; i > systemIdx; i-- {
		msg := messages[i]
		msgTokens := EstimateMessageTokens(msg)

		if tokensUsed+msgTokens <= maxTokens && recentCount < maxRecent {
			tokensUsed += msgTokens
			budget -= msgTokens
			recentCount++
		}
	}

	// Build result: system + middle important + recent messages
	middleMessages := []core.Message{}

	// Add important messages from middle if we have budget
	for i := systemIdx + 1; i < len(messages)-recentCount; i++ {
		if importantIndices[i] {
			msg := messages[i]
			msgTokens := EstimateMessageTokens(msg)
			if tokensUsed+msgTokens <= maxTokens {
				middleMessages = append(middleMessages, msg)
				tokensUsed += msgTokens
			}
		}
	}

	// Add middle important messages
	result = append(result, middleMessages...)

	// Add recent messages
	startRecent := len(messages) - recentCount
	if startRecent < 0 {
		startRecent = 0
	}
	if systemIdx >= 0 && startRecent <= systemIdx {
		startRecent = systemIdx + 1
	}

	result = append(result, messages[startRecent:]...)

	// Ensure we have at least something
	if len(result) == 0 && len(messages) > 0 {
		result = messages[len(messages)-1:]
	}

	return result
}

// ShouldPrune returns true if messages should be pruned
func (m *Manager) ShouldPrune(messages []core.Message) bool {
	tokens := EstimateMessagesTokens(messages)
	return tokens > m.MaxTokens
}

// Prune prunes the messages according to the manager's strategy
func (m *Manager) Prune(messages []core.Message) []core.Message {
	return PruneContext(messages, m.MaxTokens)
}

// GetUsage returns current context usage information
func (m *Manager) GetUsage(messages []core.Message) ContextUsage {
	tokens := EstimateMessagesTokens(messages)
	percentUsed := float64(tokens) / float64(m.MaxTokens) * 100.0

	return ContextUsage{
		EstimatedTokens: tokens,
		MaxTokens:       m.MaxTokens,
		PercentUsed:     percentUsed,
		NearLimit:       percentUsed >= (nearLimitThreshold * 100.0),
	}
}
