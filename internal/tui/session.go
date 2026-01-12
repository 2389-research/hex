// ABOUTME: Session data model for conversation history persistence
// Defines structs for storing and managing chat sessions including messages,
// tool calls, and session metadata like favorites and timestamps.

package tui

import (
	"strings"
	"time"
	"unicode"
)

// Session represents a single conversation session with all its messages.
type Session struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`      // Auto-generated from first message
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Messages  []SessionMessage `json:"messages"`
	Favorite  bool             `json:"favorite"`
}

// SessionMessage represents a single message in a session.
type SessionMessage struct {
	Role      string            `json:"role"`      // "user" or "assistant"
	Content   string            `json:"content"`
	Timestamp time.Time         `json:"timestamp"`
	ToolCalls []SessionToolCall `json:"tool_calls,omitempty"`
}

// SessionToolCall represents a tool invocation and its result within a message.
type SessionToolCall struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Input  map[string]interface{} `json:"input"`
	Output string                 `json:"output"`
	Error  bool                   `json:"error"`
}

// GenerateTitle creates a session title from the first user message.
// It truncates to approximately 50 characters at a word boundary and
// cleans up whitespace.
func GenerateTitle(firstMessage string) string {
	const maxLength = 50

	// Clean up the message: trim whitespace and collapse multiple spaces
	title := strings.TrimSpace(firstMessage)
	title = collapseWhitespace(title)

	// If empty, return a default title
	if title == "" {
		return "Untitled Session"
	}

	// If short enough, return as-is
	if len(title) <= maxLength {
		return title
	}

	// Truncate at word boundary
	truncated := title[:maxLength]

	// Find the last space to avoid cutting words
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLength/2 {
		// Only use word boundary if it doesn't cut off too much
		truncated = truncated[:lastSpace]
	}

	return strings.TrimSpace(truncated) + "..."
}

// collapseWhitespace replaces multiple consecutive whitespace characters
// (including newlines and tabs) with a single space.
func collapseWhitespace(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	inWhitespace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !inWhitespace {
				result.WriteRune(' ')
				inWhitespace = true
			}
		} else {
			result.WriteRune(r)
			inWhitespace = false
		}
	}

	return result.String()
}
