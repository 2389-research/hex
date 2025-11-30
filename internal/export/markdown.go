// Package export provides conversation export functionality in multiple formats.
// ABOUTME: Markdown exporter for conversations with frontmatter metadata
// ABOUTME: Produces clean, readable Markdown with YAML frontmatter
package export

import (
	"fmt"
	"io"
	"strings"

	"github.com/harper/clem/internal/storage"
)

// MarkdownExporter exports conversations as Markdown with YAML frontmatter
type MarkdownExporter struct{}

// Export implements the Exporter interface for Markdown format
func (e *MarkdownExporter) Export(conv *storage.Conversation, messages []*storage.Message, w io.Writer) error {
	// Write YAML frontmatter
	if _, err := fmt.Fprintf(w, "---\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "title: %s\n", escapeYAML(conv.Title)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "conversation_id: %s\n", conv.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "model: %s\n", conv.Model); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "created_at: %s\n", conv.CreatedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "updated_at: %s\n", conv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return err
	}
	if conv.SystemPrompt != "" {
		if _, err := fmt.Fprintf(w, "system_prompt: |\n"); err != nil {
			return err
		}
		for _, line := range strings.Split(conv.SystemPrompt, "\n") {
			if _, err := fmt.Fprintf(w, "  %s\n", line); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintf(w, "---\n\n"); err != nil {
		return err
	}

	// Write conversation title as H1
	if _, err := fmt.Fprintf(w, "# %s\n\n", conv.Title); err != nil {
		return err
	}

	// Write messages
	for i, msg := range messages {
		// Add separator between messages (except before first message)
		if i > 0 {
			if _, err := fmt.Fprintf(w, "\n---\n\n"); err != nil {
				return err
			}
		}

		// Write message header with role and timestamp
		roleEmoji := roleToEmoji(msg.Role)
		if _, err := fmt.Fprintf(w, "## %s %s\n\n", roleEmoji, strings.Title(msg.Role)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "*%s*\n\n", msg.CreatedAt.Format("2006-01-02 15:04:05")); err != nil {
			return err
		}

		// Write message content
		if _, err := fmt.Fprintf(w, "%s\n", msg.Content); err != nil {
			return err
		}

		// Write tool calls if present
		if msg.ToolCalls != "" && msg.ToolCalls != "null" {
			if _, err := fmt.Fprintf(w, "\n**Tool Calls:**\n```json\n%s\n```\n", msg.ToolCalls); err != nil {
				return err
			}
		}

		// Write metadata if present
		if msg.Metadata != "" && msg.Metadata != "null" {
			if _, err := fmt.Fprintf(w, "\n**Metadata:**\n```json\n%s\n```\n", msg.Metadata); err != nil {
				return err
			}
		}
	}

	return nil
}

// roleToEmoji maps message roles to emoji
func roleToEmoji(role string) string {
	switch role {
	case "user":
		return "👤"
	case "assistant":
		return "🤖"
	case "system":
		return "⚙️"
	default:
		return "💬"
	}
}

// escapeYAML escapes special characters in YAML values
func escapeYAML(s string) string {
	// If the string contains special characters, quote it
	needsQuoting := strings.ContainsAny(s, ":#{}[]|>\"'&*!%@`")
	if needsQuoting {
		// Use double quotes and escape internal quotes
		s = strings.ReplaceAll(s, "\\", "\\\\")
		s = strings.ReplaceAll(s, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", s)
	}
	return s
}
