// Package export provides conversation export functionality in multiple formats.
// ABOUTME: Export interface and dispatcher for conversation exports
// ABOUTME: Supports multiple export formats (Markdown, JSON, HTML)
package export

import (
	"database/sql"
	"fmt"
	"io"

	"github.com/harper/clem/internal/storage"
)

// Format represents an export format
type Format string

const (
	// FormatMarkdown exports conversations as Markdown with YAML frontmatter
	FormatMarkdown Format = "markdown"
	// FormatJSON exports conversations as structured JSON
	FormatJSON Format = "json"
	// FormatHTML exports conversations as styled HTML
	FormatHTML Format = "html"
)

// Exporter defines the interface for conversation exporters
type Exporter interface {
	Export(conv *storage.Conversation, messages []*storage.Message, w io.Writer) error
}

// Export exports a conversation in the specified format
func Export(db *sql.DB, conversationID string, format Format, w io.Writer) error {
	// Retrieve conversation
	conv, err := storage.GetConversation(db, conversationID)
	if err != nil {
		return fmt.Errorf("get conversation: %w", err)
	}

	// Retrieve messages
	messages, err := storage.ListMessages(db, conversationID)
	if err != nil {
		return fmt.Errorf("list messages: %w", err)
	}

	// Select exporter based on format
	var exporter Exporter
	switch format {
	case FormatMarkdown:
		exporter = &MarkdownExporter{}
	case FormatJSON:
		exporter = &JSONExporter{}
	case FormatHTML:
		exporter = &HTMLExporter{}
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	// Export
	if err := exporter.Export(conv, messages, w); err != nil {
		return fmt.Errorf("export as %s: %w", format, err)
	}

	return nil
}
