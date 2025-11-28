// ABOUTME: Message summarization using Claude API
// ABOUTME: Reduces token usage by summarizing old conversation context
package context

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/harper/clem/internal/core"
)

// Summarizer creates summaries of conversation history
type Summarizer struct {
	client *core.Client
	cache  map[string]string // Cache summaries to avoid re-summarizing
}

// NewSummarizer creates a new summarizer
func NewSummarizer(client *core.Client) *Summarizer {
	return &Summarizer{
		client: client,
		cache:  make(map[string]string),
	}
}

// SummarizeMessages creates a concise summary of messages using Claude
func (s *Summarizer) SummarizeMessages(ctx context.Context, messages []core.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	// Check cache first
	cacheKey := s.generateCacheKey(messages)
	if summary, ok := s.cache[cacheKey]; ok {
		return summary, nil
	}

	// Format messages for summarization
	formatted := s.formatMessages(messages)

	// Create summarization request
	req := core.MessageRequest{
		Model:     "claude-3-5-haiku-20241022", // Use fast, cheap model for summaries
		MaxTokens: 500,
		Messages: []core.Message{
			{
				Role: "user",
				Content: fmt.Sprintf(`Summarize the following conversation concisely. Focus on:
- Key topics discussed
- Important decisions or conclusions
- Any context needed for future messages

Keep the summary under 200 words.

Conversation:
%s`, formatted),
			},
		},
	}

	// Call API
	resp, err := s.client.CreateMessage(ctx, req)
	if err != nil {
		return "", fmt.Errorf("summarize messages: %w", err)
	}

	summary := resp.GetTextContent()

	// Cache the summary
	s.cache[cacheKey] = summary

	return summary, nil
}

// generateCacheKey creates a cache key from messages
func (s *Summarizer) generateCacheKey(messages []core.Message) string {
	// Create a simple hash of message content
	hash := sha256.New()
	for _, msg := range messages {
		hash.Write([]byte(msg.Role))
		hash.Write([]byte(msg.Content))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// formatMessages formats messages for summarization
func (s *Summarizer) formatMessages(messages []core.Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}
	return sb.String()
}

// CreateSummaryMessage creates a system message containing a summary
func CreateSummaryMessage(summary string) core.Message {
	return core.Message{
		Role: "system",
		Content: fmt.Sprintf(`Previous conversation summary:

%s

---
The above is a summary of earlier messages that were removed to save context space.`, summary),
	}
}

// ClearCache clears the summary cache
func (s *Summarizer) ClearCache() {
	s.cache = make(map[string]string)
}
