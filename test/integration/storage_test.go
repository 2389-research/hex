// ABOUTME: Storage integration tests for conversation and message persistence
// ABOUTME: Tests full database lifecycle, foreign keys, transactions, JSON storage

package integration

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/harper/pagent/internal/core"
	"github.com/harper/pagent/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConversationPersistence tests full conversation lifecycle with database
func TestConversationPersistence(t *testing.T) {
	// Use a specific database path that persists
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Phase 1: Create and populate database
	db, err := storage.OpenDatabase(dbPath)
	require.NoError(t, err)

	// Create a conversation
	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	// Add messages to it
	msg1ID := CreateTestMessage(t, db, convID, "user", "Hello, Claude!")
	msg2ID := CreateTestMessage(t, db, convID, "assistant", "Hello! How can I help you today?")

	// Close database
	_ = db.Close()

	// Phase 2: Reopen database (simulating restart)
	db, err = storage.OpenDatabase(dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Retrieve conversation
	conv, err := storage.GetConversation(db, convID)
	require.NoError(t, err)
	assert.Equal(t, "Test Conversation", conv.Title)
	assert.Equal(t, "claude-sonnet-4-5-20250929", conv.Model)

	// Retrieve messages
	msg1, err := storage.GetMessage(db, msg1ID)
	require.NoError(t, err)
	assert.Equal(t, "user", msg1.Role)
	assert.Equal(t, "Hello, Claude!", msg1.Content)

	msg2, err := storage.GetMessage(db, msg2ID)
	require.NoError(t, err)
	assert.Equal(t, "assistant", msg2.Role)
	assert.Equal(t, "Hello! How can I help you today?", msg2.Content)
}

// TestListConversationsOrdering tests that conversations are ordered by updated_at DESC
func TestListConversationsOrdering(t *testing.T) {
	db := SetupTestDB(t)

	// Create three conversations with different timestamps
	conv1 := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")
	time.Sleep(10 * time.Millisecond)

	conv2 := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")
	time.Sleep(10 * time.Millisecond)

	conv3 := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	// List conversations
	conversations, err := storage.ListConversations(db, 10, 0)
	require.NoError(t, err)
	require.Len(t, conversations, 3)

	// Should be in reverse chronological order (most recent first)
	assert.Equal(t, conv3, conversations[0].ID)
	assert.Equal(t, conv2, conversations[1].ID)
	assert.Equal(t, conv1, conversations[2].ID)
}

// TestMessageWithToolResults tests storing and retrieving messages with tool_calls JSON
func TestMessageWithToolResults(t *testing.T) {
	db := SetupTestDB(t)

	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	// Create a message with tool calls
	toolCalls := []core.ToolUse{
		{
			ID:    "tool_123",
			Name:  "read_file",
			Input: map[string]interface{}{"path": "/tmp/test.txt"},
		},
		{
			ID:    "tool_456",
			Name:  "bash",
			Input: map[string]interface{}{"command": "ls -la"},
		},
	}

	toolCallsJSON, err := json.Marshal(toolCalls)
	require.NoError(t, err)

	msg := &storage.Message{
		ID:             "msg_with_tools",
		ConversationID: convID,
		Role:           "assistant",
		Content:        "I'll help you with that.",
		ToolCalls:      string(toolCallsJSON),
		CreatedAt:      time.Now(),
	}

	err = storage.CreateMessage(db, msg)
	require.NoError(t, err)

	// Retrieve and verify JSON roundtrip
	retrieved, err := storage.GetMessage(db, "msg_with_tools")
	require.NoError(t, err)
	assert.Equal(t, "I'll help you with that.", retrieved.Content)

	// Unmarshal tool calls
	var retrievedToolCalls []core.ToolUse
	err = json.Unmarshal([]byte(retrieved.ToolCalls), &retrievedToolCalls)
	require.NoError(t, err)
	assert.Len(t, retrievedToolCalls, 2)
	assert.Equal(t, "read_file", retrievedToolCalls[0].Name)
	assert.Equal(t, "bash", retrievedToolCalls[1].Name)
}

// TestListMessagesByConversation tests retrieving all messages for a conversation
func TestListMessagesByConversation(t *testing.T) {
	db := SetupTestDB(t)

	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	// Create multiple messages
	CreateTestMessage(t, db, convID, "user", "Message 1")
	time.Sleep(5 * time.Millisecond)
	CreateTestMessage(t, db, convID, "assistant", "Message 2")
	time.Sleep(5 * time.Millisecond)
	CreateTestMessage(t, db, convID, "user", "Message 3")

	// List messages
	messages, err := storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 3)

	// Should be in chronological order
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "Message 1", messages[0].Content)
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, "Message 2", messages[1].Content)
	assert.Equal(t, "user", messages[2].Role)
	assert.Equal(t, "Message 3", messages[2].Content)
}

// TestForeignKeyConstraints tests CASCADE DELETE behavior
func TestForeignKeyConstraints(t *testing.T) {
	db := SetupTestDB(t)

	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")
	CreateTestMessage(t, db, convID, "user", "Test message")

	// Verify message exists
	messages, err := storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 1)

	// Delete conversation (should cascade to messages)
	_, err = db.Exec("DELETE FROM conversations WHERE id = ?", convID)
	require.NoError(t, err)

	// Verify messages were also deleted
	messages, err = storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 0, "messages should be deleted when conversation is deleted")
}

// TestConversationTimestampUpdate tests that updated_at is updated when messages are added
func TestConversationTimestampUpdate(t *testing.T) {
	db := SetupTestDB(t)

	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	// Get initial timestamp
	conv, err := storage.GetConversation(db, convID)
	require.NoError(t, err)
	initialTime := conv.UpdatedAt

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Add a message
	CreateTestMessage(t, db, convID, "user", "New message")

	// Get updated timestamp
	conv, err = storage.GetConversation(db, convID)
	require.NoError(t, err)

	// Updated timestamp should be after initial
	assert.True(t, conv.UpdatedAt.After(initialTime), "updated_at should be updated when message is added")
}

// TestEmptyToolCallsAndMetadata tests that NULL JSON fields are handled correctly
func TestEmptyToolCallsAndMetadata(t *testing.T) {
	db := SetupTestDB(t)

	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	// Create message without tool calls or metadata
	msg := &storage.Message{
		ID:             "msg_no_tools",
		ConversationID: convID,
		Role:           "assistant",
		Content:        "Simple message",
		CreatedAt:      time.Now(),
	}

	err := storage.CreateMessage(db, msg)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := storage.GetMessage(db, "msg_no_tools")
	require.NoError(t, err)
	assert.Equal(t, "Simple message", retrieved.Content)
	assert.Empty(t, retrieved.ToolCalls, "tool_calls should be empty string for NULL")
	assert.Empty(t, retrieved.Metadata, "metadata should be empty string for NULL")
}

// TestPaginationWithLargeConversation tests pagination works correctly
func TestPaginationWithLargeConversation(t *testing.T) {
	db := SetupTestDB(t)

	// Create 50 conversations
	for i := 0; i < 50; i++ {
		CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")
		time.Sleep(1 * time.Millisecond)
	}

	// Test pagination
	page1, err := storage.ListConversations(db, 10, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 10)

	page2, err := storage.ListConversations(db, 10, 10)
	require.NoError(t, err)
	assert.Len(t, page2, 10)

	// Pages should be different
	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}
