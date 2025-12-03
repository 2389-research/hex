// ABOUTME: Comprehensive tests for conversation export functionality
// ABOUTME: Tests all export formats with various message types and edge cases
package export

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestData creates test conversation and messages
func setupTestData() (*storage.Conversation, []*storage.Message) {
	now := time.Now()

	conv := &storage.Conversation{
		ID:           "test-conv-123",
		Title:        "Test Conversation",
		Model:        "claude-3-5-sonnet-20241022",
		SystemPrompt: "You are a helpful assistant.",
		CreatedAt:    now.Add(-1 * time.Hour),
		UpdatedAt:    now,
	}

	messages := []*storage.Message{
		{
			ID:             "msg-1",
			ConversationID: "test-conv-123",
			Role:           "user",
			Content:        "Hello! Can you help me write a Python function?",
			CreatedAt:      now.Add(-50 * time.Minute),
		},
		{
			ID:             "msg-2",
			ConversationID: "test-conv-123",
			Role:           "assistant",
			Content:        "Of course! Here's a simple example:\n\n```python\ndef greet(name):\n    return f\"Hello, {name}!\"\n```\n\nThis function takes a name and returns a greeting.",
			CreatedAt:      now.Add(-45 * time.Minute),
		},
		{
			ID:             "msg-3",
			ConversationID: "test-conv-123",
			Role:           "user",
			Content:        "Thanks! Can you add error handling?",
			CreatedAt:      now.Add(-40 * time.Minute),
		},
		{
			ID:             "msg-4",
			ConversationID: "test-conv-123",
			Role:           "assistant",
			Content:        "Sure! Here's the improved version:\n\n```python\ndef greet(name):\n    if not isinstance(name, str):\n        raise TypeError(\"Name must be a string\")\n    if not name.strip():\n        raise ValueError(\"Name cannot be empty\")\n    return f\"Hello, {name}!\"\n```",
			ToolCalls:      `[{"type":"tool_use","id":"tool-123","name":"python_repl","input":{"code":"print('test')"}}]`,
			Metadata:       `{"tokens":{"input":100,"output":150}}`,
			CreatedAt:      now.Add(-35 * time.Minute),
		},
	}

	return conv, messages
}

// setupTestDataWithSpecialChars creates test data with special characters
func setupTestDataWithSpecialChars() (*storage.Conversation, []*storage.Message) {
	now := time.Now()

	conv := &storage.Conversation{
		ID:           "test-conv-special",
		Title:        `Test "Conversation" with: special & chars <html>`,
		Model:        "claude-3-5-sonnet-20241022",
		SystemPrompt: "System prompt with special chars: <>\"'&",
		CreatedAt:    now.Add(-1 * time.Hour),
		UpdatedAt:    now,
	}

	messages := []*storage.Message{
		{
			ID:             "msg-1",
			ConversationID: "test-conv-special",
			Role:           "user",
			Content:        `Content with <html> tags & "quotes" and 'apostrophes'`,
			CreatedAt:      now.Add(-30 * time.Minute),
		},
		{
			ID:             "msg-2",
			ConversationID: "test-conv-special",
			Role:           "assistant",
			Content:        "Response with emoji: 👍 and unicode: café",
			CreatedAt:      now.Add(-25 * time.Minute),
		},
	}

	return conv, messages
}

// TestMarkdownExporter tests the Markdown export format
func TestMarkdownExporter(t *testing.T) {
	conv, messages := setupTestData()
	exporter := &MarkdownExporter{}

	var buf bytes.Buffer
	err := exporter.Export(conv, messages, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Check frontmatter
	assert.Contains(t, output, "---")
	assert.Contains(t, output, "title: Test Conversation")
	assert.Contains(t, output, "conversation_id: test-conv-123")
	assert.Contains(t, output, "model: claude-3-5-sonnet-20241022")
	assert.Contains(t, output, "system_prompt: |")
	assert.Contains(t, output, "  You are a helpful assistant.")

	// Check message content
	assert.Contains(t, output, "## 👤 User")
	assert.Contains(t, output, "## 🤖 Assistant")
	assert.Contains(t, output, "Hello! Can you help me write a Python function?")
	assert.Contains(t, output, "```python")
	assert.Contains(t, output, "def greet(name):")

	// Check tool calls
	assert.Contains(t, output, "**Tool Calls:**")
	assert.Contains(t, output, "python_repl")

	// Check metadata
	assert.Contains(t, output, "**Metadata:**")
	assert.Contains(t, output, "tokens")
}

// TestMarkdownExporterSpecialChars tests special character handling
func TestMarkdownExporterSpecialChars(t *testing.T) {
	conv, messages := setupTestDataWithSpecialChars()
	exporter := &MarkdownExporter{}

	var buf bytes.Buffer
	err := exporter.Export(conv, messages, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Check that special characters are properly escaped in YAML
	assert.Contains(t, output, "title:")
	assert.Contains(t, output, "special")

	// Check message content preserves special characters
	assert.Contains(t, output, `<html>`)
	assert.Contains(t, output, `"quotes"`)
	assert.Contains(t, output, "emoji: 👍")
	assert.Contains(t, output, "café")
}

// TestJSONExporter tests the JSON export format
func TestJSONExporter(t *testing.T) {
	conv, messages := setupTestData()
	exporter := &JSONExporter{}

	var buf bytes.Buffer
	err := exporter.Export(conv, messages, &buf)
	require.NoError(t, err)

	// Parse the JSON to verify structure
	var export ConversationExport
	err = json.Unmarshal(buf.Bytes(), &export)
	require.NoError(t, err)

	// Verify conversation data
	assert.Equal(t, "test-conv-123", export.Conversation.ID)
	assert.Equal(t, "Test Conversation", export.Conversation.Title)
	assert.Equal(t, "claude-3-5-sonnet-20241022", export.Conversation.Model)
	assert.Equal(t, "You are a helpful assistant.", export.Conversation.SystemPrompt)

	// Verify messages
	assert.Len(t, export.Messages, 4)
	assert.Equal(t, "user", export.Messages[0].Role)
	assert.Equal(t, "assistant", export.Messages[1].Role)
	assert.Contains(t, export.Messages[1].Content, "```python")

	// Verify tool calls
	assert.NotEmpty(t, export.Messages[3].ToolCalls)
	var toolCalls []map[string]interface{}
	err = json.Unmarshal(export.Messages[3].ToolCalls, &toolCalls)
	require.NoError(t, err)
	assert.Len(t, toolCalls, 1)
	assert.Equal(t, "python_repl", toolCalls[0]["name"])

	// Verify metadata
	assert.NotEmpty(t, export.Messages[3].Metadata)
	var metadata map[string]interface{}
	err = json.Unmarshal(export.Messages[3].Metadata, &metadata)
	require.NoError(t, err)
	assert.Contains(t, metadata, "tokens")
}

// TestJSONExporterRoundTrip tests that exported JSON can be parsed back
func TestJSONExporterRoundTrip(t *testing.T) {
	conv, messages := setupTestData()
	exporter := &JSONExporter{}

	var buf bytes.Buffer
	err := exporter.Export(conv, messages, &buf)
	require.NoError(t, err)

	// Parse back
	var export ConversationExport
	err = json.Unmarshal(buf.Bytes(), &export)
	require.NoError(t, err)

	// Verify round-trip
	assert.Equal(t, conv.ID, export.Conversation.ID)
	assert.Equal(t, conv.Title, export.Conversation.Title)
	assert.Len(t, export.Messages, len(messages))

	// Export again and compare
	exporter2 := &JSONExporter{}
	var buf2 bytes.Buffer
	err = exporter2.Export(conv, messages, &buf2)
	require.NoError(t, err)

	// The JSON should be identical
	assert.JSONEq(t, buf.String(), buf2.String())
}

// TestJSONExporterSpecialChars tests special character handling
func TestJSONExporterSpecialChars(t *testing.T) {
	conv, messages := setupTestDataWithSpecialChars()
	exporter := &JSONExporter{}

	var buf bytes.Buffer
	err := exporter.Export(conv, messages, &buf)
	require.NoError(t, err)

	// Parse the JSON
	var export ConversationExport
	err = json.Unmarshal(buf.Bytes(), &export)
	require.NoError(t, err)

	// Verify special characters are preserved
	assert.Contains(t, export.Conversation.Title, `"Conversation"`)
	assert.Contains(t, export.Conversation.Title, "<html>")
	assert.Contains(t, export.Messages[0].Content, `<html>`)
	assert.Contains(t, export.Messages[0].Content, `"quotes"`)
	assert.Contains(t, export.Messages[1].Content, "👍")
	assert.Contains(t, export.Messages[1].Content, "café")
}

// TestHTMLExporter tests the HTML export format
func TestHTMLExporter(t *testing.T) {
	conv, messages := setupTestData()
	exporter := &HTMLExporter{}

	var buf bytes.Buffer
	err := exporter.Export(conv, messages, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Check HTML structure
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "<html lang=\"en\">")
	assert.Contains(t, output, "</html>")

	// Check metadata
	assert.Contains(t, output, "<h1>Test Conversation</h1>")
	assert.Contains(t, output, "test-conv-123")
	assert.Contains(t, output, "claude-3-5-sonnet-20241022")
	assert.Contains(t, output, "You are a helpful assistant.")

	// Check messages
	assert.Contains(t, output, "message-user")
	assert.Contains(t, output, "message-assistant")
	assert.Contains(t, output, "👤")
	assert.Contains(t, output, "🤖")

	// Check content
	assert.Contains(t, output, "Hello! Can you help me write a Python function?")

	// Check that code blocks are present
	// Chroma should have highlighted the code, so we look for its output
	// The code will be wrapped in spans with inline styles
	assert.Contains(t, output, "greet")       // Function name should be present
	assert.Contains(t, output, "<pre style=") // Should have syntax highlighting

	// Check tool calls
	assert.Contains(t, output, "Tool Calls:")

	// Verify it's valid HTML (basic check)
	assert.Equal(t, strings.Count(output, "<body>"), strings.Count(output, "</body>"))
	assert.Equal(t, strings.Count(output, "<div"), strings.Count(output, "</div>"))
}

// TestHTMLExporterSpecialChars tests special character escaping
func TestHTMLExporterSpecialChars(t *testing.T) {
	conv, messages := setupTestDataWithSpecialChars()
	exporter := &HTMLExporter{}

	var buf bytes.Buffer
	err := exporter.Export(conv, messages, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Check that HTML special characters are escaped
	// The title should be escaped in the HTML
	assert.NotContains(t, output, `Test "Conversation" with: special & chars <html>`)
	// But it should contain the escaped versions or be in a safe context
	assert.Contains(t, output, "Test")
	assert.Contains(t, output, "special")

	// Check that emoji is preserved
	assert.Contains(t, output, "👍")
	assert.Contains(t, output, "café")
}

// TestHTMLExporterToFile tests writing HTML to a file
func TestHTMLExporterToFile(t *testing.T) {
	conv, messages := setupTestData()
	exporter := &HTMLExporter{}

	// Create temp file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "export.html")

	file, err := os.Create(outputPath) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	err = exporter.Export(conv, messages, file)
	require.NoError(t, err)

	// Close and read back
	_ = file.Close()
	//nolint:gosec // G304: Test file reads/writes are safe

	content, err := os.ReadFile(outputPath) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err)                 //nolint:gosec // G304: Path validated by caller

	output := string(content)
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "Test Conversation")
}

// TestExportFunction tests the main Export dispatcher
func TestExportFunction(t *testing.T) {
	// Setup test database
	db, err := storage.OpenDatabase(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create conversation and messages
	conv, messages := setupTestData()
	err = storage.CreateConversation(db, conv)
	require.NoError(t, err)

	for _, msg := range messages {
		err = storage.CreateMessage(db, msg)
		require.NoError(t, err)
	}

	// Test each format
	formats := []Format{FormatMarkdown, FormatJSON, FormatHTML}
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			err := Export(db, conv.ID, format, &buf)
			require.NoError(t, err)
			assert.NotEmpty(t, buf.String())
		})
	}
}

// TestExportUnknownFormat tests error handling for unknown format
func TestExportUnknownFormat(t *testing.T) {
	db, err := storage.OpenDatabase(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	conv, messages := setupTestData()
	err = storage.CreateConversation(db, conv)
	require.NoError(t, err)

	for _, msg := range messages {
		err = storage.CreateMessage(db, msg)
		require.NoError(t, err)
	}

	var buf bytes.Buffer
	err = Export(db, conv.ID, Format("invalid"), &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown format")
}

// TestExportNonexistentConversation tests error handling for missing conversation
func TestExportNonexistentConversation(t *testing.T) {
	db, err := storage.OpenDatabase(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	var buf bytes.Buffer
	err = Export(db, "nonexistent-id", FormatMarkdown, &buf)
	assert.Error(t, err)
}

// TestEmptyConversation tests exporting a conversation with no messages
func TestEmptyConversation(t *testing.T) {
	db, err := storage.OpenDatabase(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	conv := &storage.Conversation{
		ID:        "empty-conv",
		Title:     "Empty Conversation",
		Model:     "claude-3-5-sonnet-20241022",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Test each format with empty conversation
	formats := []Format{FormatMarkdown, FormatJSON, FormatHTML}
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			err := Export(db, conv.ID, format, &buf)
			require.NoError(t, err)
			assert.NotEmpty(t, buf.String())
		})
	}
}
