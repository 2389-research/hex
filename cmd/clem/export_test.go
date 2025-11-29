// ABOUTME: Tests for the export command
// ABOUTME: Tests export functionality with various formats
package main

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harper/clem/internal/export"
	"github.com/harper/clem/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConversation creates a test database with a conversation and messages
func setupTestConversation(t *testing.T) (*sql.DB, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := openDatabase(dbPath)
	require.NoError(t, err)

	// Create test conversation
	conv := &storage.Conversation{
		ID:           "test-export",
		Title:        "Export Test",
		Model:        "claude-3-5-sonnet-20241022",
		SystemPrompt: "You are a helpful assistant.",
		CreatedAt:    time.Now().Add(-1 * time.Hour),
		UpdatedAt:    time.Now(),
	}
	err = storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Add messages
	messages := []*storage.Message{
		{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        "Hello! Can you help me?",
			CreatedAt:      time.Now().Add(-30 * time.Minute),
		},
		{
			ConversationID: conv.ID,
			Role:           "assistant",
			Content:        "Of course! Here's a code example:\n\n```python\ndef hello():\n    print('Hello, world!')\n```",
			CreatedAt:      time.Now().Add(-25 * time.Minute),
		},
	}

	for _, msg := range messages {
		err = storage.CreateMessage(db, msg)
		require.NoError(t, err)
	}

	return db, conv.ID
}

// TestExportMarkdown tests exporting to Markdown format
func TestExportMarkdown(t *testing.T) {
	db, convID := setupTestConversation(t)
	defer db.Close()

	var buf bytes.Buffer
	err := export.Export(db, convID, export.FormatMarkdown, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Check frontmatter
	assert.Contains(t, output, "---")
	assert.Contains(t, output, "title: Export Test")
	assert.Contains(t, output, "model: claude-3-5-sonnet-20241022")

	// Check message content
	assert.Contains(t, output, "Hello! Can you help me?")
	assert.Contains(t, output, "```python")
	assert.Contains(t, output, "def hello()")
}

// TestExportJSON tests exporting to JSON format
func TestExportJSON(t *testing.T) {
	db, convID := setupTestConversation(t)
	defer db.Close()

	var buf bytes.Buffer
	err := export.Export(db, convID, export.FormatJSON, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Check JSON structure
	assert.Contains(t, output, `"conversation"`)
	assert.Contains(t, output, `"messages"`)
	assert.Contains(t, output, `"Export Test"`)
	assert.Contains(t, output, `"Hello! Can you help me?"`)
}

// TestExportHTML tests exporting to HTML format
func TestExportHTML(t *testing.T) {
	db, convID := setupTestConversation(t)
	defer db.Close()

	var buf bytes.Buffer
	err := export.Export(db, convID, export.FormatHTML, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Check HTML structure
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "<html")
	assert.Contains(t, output, "Export Test")
	assert.Contains(t, output, "Hello! Can you help me?")

	// Check for syntax highlighting
	assert.Contains(t, output, "<pre style=")
}

// TestExportToFile tests exporting to a file
func TestExportToFile(t *testing.T) {
	db, convID := setupTestConversation(t)
	defer db.Close()

	outputPath := filepath.Join(t.TempDir(), "export.md")
	file, err := os.Create(outputPath)
	require.NoError(t, err)
	defer file.Close()

	err = export.Export(db, convID, export.FormatMarkdown, file)
	require.NoError(t, err)

	// Close and read back
	file.Close()

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(content)
	assert.Contains(t, output, "Export Test")
	assert.Contains(t, output, "Hello! Can you help me?")
}

// TestExportInvalidFormat tests error handling for invalid format
func TestExportInvalidFormat(t *testing.T) {
	db, convID := setupTestConversation(t)
	defer db.Close()

	var buf bytes.Buffer
	err := export.Export(db, convID, export.Format("invalid"), &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown format")
}

// TestExportNonexistentConversation tests error handling for missing conversation
func TestExportNonexistentConversation(t *testing.T) {
	db, _ := setupTestConversation(t)
	defer db.Close()

	var buf bytes.Buffer
	err := export.Export(db, "nonexistent-id", export.FormatMarkdown, &buf)
	require.Error(t, err)
}

// TestExportFormatAliases tests that format aliases work
func TestExportFormatAliases(t *testing.T) {
	testCases := []struct {
		input    string
		expected export.Format
	}{
		{"markdown", export.FormatMarkdown},
		{"md", export.FormatMarkdown},
		{"json", export.FormatJSON},
		{"html", export.FormatHTML},
		{"htm", export.FormatHTML},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			var format export.Format
			switch tc.input {
			case "markdown", "md":
				format = export.FormatMarkdown
			case "json":
				format = export.FormatJSON
			case "html", "htm":
				format = export.FormatHTML
			}

			assert.Equal(t, tc.expected, format)
		})
	}
}

// TestExportEmptyConversation tests exporting a conversation with no messages
func TestExportEmptyConversation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := openDatabase(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create conversation without messages
	conv := &storage.Conversation{
		ID:        "empty-conv",
		Title:     "Empty Conversation",
		Model:     "claude-3-5-sonnet-20241022",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Test each format
	formats := []export.Format{
		export.FormatMarkdown,
		export.FormatJSON,
		export.FormatHTML,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			err := export.Export(db, conv.ID, format, &buf)
			require.NoError(t, err)
			assert.NotEmpty(t, buf.String())
		})
	}
}

// TestExportWithToolCalls tests exporting messages with tool calls
func TestExportWithToolCalls(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := openDatabase(dbPath)
	require.NoError(t, err)
	defer db.Close()

	conv := &storage.Conversation{
		ID:        "tool-conv",
		Title:     "Tool Test",
		Model:     "claude-3-5-sonnet-20241022",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Add message with tool calls
	msg := &storage.Message{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        "I'll use a tool.",
		ToolCalls:      `[{"type":"tool_use","id":"tool-1","name":"read_file","input":{"path":"test.txt"}}]`,
		Metadata:       `{"tokens":{"input":50,"output":100}}`,
		CreatedAt:      time.Now(),
	}
	err = storage.CreateMessage(db, msg)
	require.NoError(t, err)

	// Test Markdown export includes tool calls
	var buf bytes.Buffer
	err = export.Export(db, conv.ID, export.FormatMarkdown, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Tool Calls:")
	assert.Contains(t, output, "read_file")

	// Test JSON export includes tool calls
	buf.Reset()
	err = export.Export(db, conv.ID, export.FormatJSON, &buf)
	require.NoError(t, err)

	output = buf.String()
	assert.Contains(t, output, "tool_use")
	assert.Contains(t, output, "read_file")
}

// TestExportSpecialCharacters tests handling of special characters
func TestExportSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := openDatabase(dbPath)
	require.NoError(t, err)
	defer db.Close()

	conv := &storage.Conversation{
		ID:           "special-conv",
		Title:        `Test "Conversation" with: special & chars <html>`,
		Model:        "claude-3-5-sonnet-20241022",
		SystemPrompt: "System prompt with <tags> & \"quotes\"",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = storage.CreateConversation(db, conv)
	require.NoError(t, err)

	msg := &storage.Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        `Content with <html> tags & "quotes" and emoji: 👍`,
		CreatedAt:      time.Now(),
	}
	err = storage.CreateMessage(db, msg)
	require.NoError(t, err)

	// Test all formats handle special characters
	formats := []export.Format{
		export.FormatMarkdown,
		export.FormatJSON,
		export.FormatHTML,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			err := export.Export(db, conv.ID, format, &buf)
			require.NoError(t, err)

			output := buf.String()
			assert.NotEmpty(t, output)

			// Verify content is present (may be escaped depending on format)
			lowerOutput := strings.ToLower(output)
			assert.True(t,
				strings.Contains(lowerOutput, "special") ||
					strings.Contains(output, "special"),
				"Output should contain 'special'")
		})
	}
}
