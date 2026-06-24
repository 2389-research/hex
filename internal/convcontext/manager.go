// ABOUTME: Context pruning and token management for conversations
// ABOUTME: Keeps conversations within token limits while preserving important context

// Package convcontext provides context window management and message pruning for conversations.
package convcontext

import (
	"fmt"
	"strings"

	"github.com/2389-research/hex/internal/core"
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
	return NewManagerWithStrategy(maxTokens, StrategyPrune)
}

// NewManagerWithStrategy creates a new context manager with an explicit pruning strategy.
func NewManagerWithStrategy(maxTokens int, strategy PruneStrategy) *Manager {
	return &Manager{
		MaxTokens: maxTokens,
		Strategy:  strategy,
	}
}

// ParsePruneStrategy converts a CLI/config string into a pruning strategy.
func ParsePruneStrategy(strategy string) (PruneStrategy, error) {
	switch strings.TrimSpace(strings.ToLower(strategy)) {
	case "keep-all":
		return StrategyKeepAll, nil
	case "prune":
		return StrategyPrune, nil
	case "summarize":
		return StrategySummarize, nil
	default:
		return StrategyPrune, fmt.Errorf("invalid context strategy %q: must be keep-all, prune, or summarize", strategy)
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

	// Add tokens for content blocks (tool calls, tool results, etc.)
	for _, block := range msg.ContentBlock {
		tokens += EstimateTokens(block.Text)
		tokens += EstimateTokens(block.Content)
		if block.Name != "" {
			tokens += EstimateTokens(block.Name)
		}
		if block.Type == "tool_use" && block.Input != nil {
			tokens += 20
		}
	}

	// Add tokens for legacy tool calls if present
	for _, tool := range msg.ToolCalls {
		tokens += EstimateTokens(tool.Name)
		// Rough estimate for input params
		tokens += 10
	}

	return tokens
}

// SummarizeToolResult creates a brief summary of a tool result for context pruning
func SummarizeToolResult(toolName, content string) string {
	if content == "" {
		return "[Previously: " + toolName + " returned empty]"
	}

	lines := 1
	for _, c := range content {
		if c == '\n' {
			lines++
		}
	}

	if len(content) > 200 {
		return fmt.Sprintf("[Previously: %s returned %d lines]", toolName, lines)
	}
	return "[Previously: " + toolName + " executed]"
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

	// Step 2: Identify important messages (with tool calls or errors)
	importantIndices := make(map[int]bool)
	for i, msg := range messages {
		if i == systemIdx {
			continue
		}
		if len(msg.ToolCalls) > 0 {
			importantIndices[i] = true
		}
		// Prioritize messages containing errors
		if strings.Contains(strings.ToLower(msg.Content), "error") {
			importantIndices[i] = true
		}
		// Check content blocks for errors too
		for _, block := range msg.ContentBlock {
			if strings.Contains(strings.ToLower(block.Content), "error") {
				importantIndices[i] = true
			}
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
	if m.Strategy == StrategyKeepAll {
		return false
	}
	tokens := EstimateMessagesTokens(messages)
	return tokens > m.MaxTokens
}

// Prune prunes the messages according to the manager's strategy
func (m *Manager) Prune(messages []core.Message) []core.Message {
	if m.Strategy == StrategyKeepAll {
		return messages
	}
	if m.Strategy == StrategySummarize {
		return SummarizeContext(messages, m.MaxTokens)
	}
	return PruneContext(messages, m.MaxTokens)
}

// SummarizeContext replaces pruned messages with a compact summary message.
func SummarizeContext(messages []core.Message, maxTokens int) []core.Message {
	pruned := PruneContext(messages, maxTokens)
	if len(pruned) == len(messages) {
		return pruned
	}

	removed := removedMessages(messages, pruned)
	if len(removed) == 0 {
		return pruned
	}

	summary := core.Message{
		Role:    "system",
		Content: compactPrunedSummary(removed),
	}

	insertAt := 0
	if len(pruned) > 0 && pruned[0].Role == "system" {
		insertAt = 1
	}

	result := make([]core.Message, 0, len(pruned)+1)
	result = append(result, pruned[:insertAt]...)
	result = append(result, summary)
	result = append(result, pruned[insertAt:]...)

	for EstimateMessagesTokens(result) > maxTokens && len(result) > insertAt+2 {
		removeAt := insertAt + 1
		if removeAt >= len(result)-1 {
			break
		}
		result = append(result[:removeAt], result[removeAt+1:]...)
	}

	return result
}

func removedMessages(messages []core.Message, kept []core.Message) []core.Message {
	keptCounts := make(map[string]int, len(kept))
	for _, msg := range kept {
		keptCounts[messageSignature(msg)]++
	}

	removed := make([]core.Message, 0, len(messages)-len(kept))
	for _, msg := range messages {
		signature := messageSignature(msg)
		if keptCounts[signature] > 0 {
			keptCounts[signature]--
			continue
		}
		removed = append(removed, msg)
	}
	return removed
}

func messageSignature(msg core.Message) string {
	return fmt.Sprintf("%s\x00%s\x00%d\x00%d", msg.Role, msg.Content, len(msg.ContentBlock), len(msg.ToolCalls))
}

func compactPrunedSummary(messages []core.Message) string {
	var builder strings.Builder
	builder.WriteString("Previous conversation summary:\n")
	builder.WriteString(fmt.Sprintf("Pruned %d earlier message(s).", len(messages)))

	for _, msg := range messages {
		if msg.Role == "system" {
			continue
		}
		line := strings.TrimSpace(strings.ReplaceAll(msg.Content, "\n", " "))
		if line == "" {
			continue
		}
		if len(line) > 120 {
			line = line[:120] + "..."
		}
		builder.WriteString("\n- ")
		builder.WriteString(msg.Role)
		builder.WriteString(": ")
		builder.WriteString(line)
	}

	return builder.String()
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
