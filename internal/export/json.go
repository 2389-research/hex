// Package export provides conversation export functionality in multiple formats.
// ABOUTME: JSON exporter for conversations with full structure
// ABOUTME: Produces complete JSON export suitable for parsing and round-trip
package export

import (
	"encoding/json"
	"io"

	"github.com/harper/hex/internal/storage"
)

// JSONExporter exports conversations as JSON
type JSONExporter struct{}

// ConversationExport represents the full conversation export structure
type ConversationExport struct {
	Conversation ConversationData `json:"conversation"`
	Messages     []MessageData    `json:"messages"`
}

// ConversationData represents conversation metadata
type ConversationData struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// MessageData represents a message with metadata
type MessageData struct {
	ID             string          `json:"id"`
	ConversationID string          `json:"conversation_id"`
	Role           string          `json:"role"`
	Content        string          `json:"content"`
	ToolCalls      json.RawMessage `json:"tool_calls,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      string          `json:"created_at"`
}

// Export implements the Exporter interface for JSON format
func (e *JSONExporter) Export(conv *storage.Conversation, messages []*storage.Message, w io.Writer) error {
	// Build export structure
	export := ConversationExport{
		Conversation: ConversationData{
			ID:           conv.ID,
			Title:        conv.Title,
			Model:        conv.Model,
			SystemPrompt: conv.SystemPrompt,
			CreatedAt:    conv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    conv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Messages: make([]MessageData, 0, len(messages)),
	}

	// Convert messages
	for _, msg := range messages {
		msgData := MessageData{
			ID:             msg.ID,
			ConversationID: msg.ConversationID,
			Role:           msg.Role,
			Content:        msg.Content,
			CreatedAt:      msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Include tool calls if present
		if msg.ToolCalls != "" && msg.ToolCalls != "null" {
			msgData.ToolCalls = json.RawMessage(msg.ToolCalls)
		}

		// Include metadata if present
		if msg.Metadata != "" && msg.Metadata != "null" {
			msgData.Metadata = json.RawMessage(msg.Metadata)
		}

		export.Messages = append(export.Messages, msgData)
	}

	// Encode as pretty-printed JSON
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(export); err != nil {
		return err
	}

	return nil
}
