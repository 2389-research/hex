// ABOUTME: Integration tests for storage functionality
// ABOUTME: Tests end-to-end conversation and message persistence
package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/harper/clem/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageIntegration(t *testing.T) {
	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integration_test.db")

	// Test 1: Open database and create conversation
	db, err := openDatabase(dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create conversation
	conv := &storage.Conversation{
		ID:    "test-conv-1",
		Title: "New Conversation",
		Model: "claude-sonnet-4-5-20250929",
	}
	err = storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Test 2: Save user message
	userMsg := &storage.Message{
		ConversationID: "test-conv-1",
		Role:           "user",
		Content:        "Hello, how are you?",
	}
	err = storage.CreateMessage(db, userMsg)
	require.NoError(t, err)

	// Test 3: Update conversation title
	title := generateConversationTitle("Hello, how are you?")
	err = storage.UpdateConversationTitle(db, "test-conv-1", title)
	require.NoError(t, err)

	// Test 4: Save assistant message
	assistantMsg := &storage.Message{
		ConversationID: "test-conv-1",
		Role:           "assistant",
		Content:        "I'm doing great, thanks for asking!",
	}
	err = storage.CreateMessage(db, assistantMsg)
	require.NoError(t, err)

	// Test 5: Load conversation history
	loadedConv, messages, err := loadConversationHistory(db, "test-conv-1")
	require.NoError(t, err)
	assert.Equal(t, "Hello, how are you?", loadedConv.Title)
	assert.Len(t, messages, 2)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "Hello, how are you?", messages[0].Content)
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, "I'm doing great, thanks for asking!", messages[1].Content)

	// Test 6: Get latest conversation
	latestConv, err := storage.GetLatestConversation(db)
	require.NoError(t, err)
	assert.Equal(t, "test-conv-1", latestConv.ID)
}

func TestConversationTitleGeneration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short message",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "long message truncated",
			input:    "This is a very long message that should be truncated because it exceeds the maximum length allowed for titles",
			expected: "This is a very long message that should be trun...",
		},
		{
			name:     "message with newlines",
			input:    "Hello\nworld\nfrom\nGo",
			expected: "Hello world from Go",
		},
		{
			name:     "message with extra whitespace",
			input:    "  Hello   world  ",
			expected: "Hello   world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateConversationTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDatabasePersistence(t *testing.T) {
	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persistence_test.db")

	// Create database and add data
	db1, err := openDatabase(dbPath)
	require.NoError(t, err)

	conv := &storage.Conversation{
		ID:    "persist-conv",
		Title: "Test Persistence",
		Model: "claude-sonnet-4-5-20250929",
	}
	err = storage.CreateConversation(db1, conv)
	require.NoError(t, err)

	msg := &storage.Message{
		ConversationID: "persist-conv",
		Role:           "user",
		Content:        "This should persist",
	}
	err = storage.CreateMessage(db1, msg)
	require.NoError(t, err)

	_ = db1.Close()

	// Reopen database and verify data persists
	db2, err := openDatabase(dbPath)
	require.NoError(t, err)
	defer func() { _ = db2.Close() }()

	loadedConv, messages, err := loadConversationHistory(db2, "persist-conv")
	require.NoError(t, err)
	assert.Equal(t, "Test Persistence", loadedConv.Title)
	assert.Len(t, messages, 1)
	assert.Equal(t, "This should persist", messages[0].Content)

	// Verify database file exists
	_, err = os.Stat(dbPath)
	require.NoError(t, err)
}

func TestMultipleConversations(t *testing.T) {
	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "multi_conv_test.db")

	db, err := openDatabase(dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create multiple conversations
	for i := 1; i <= 3; i++ {
		conv := &storage.Conversation{
			ID:    filepath.Join("conv", string(rune('0'+i))),
			Title: filepath.Join("Conversation", string(rune('0'+i))),
			Model: "claude-sonnet-4-5-20250929",
		}
		err = storage.CreateConversation(db, conv)
		require.NoError(t, err)
	}

	// List conversations
	convs, err := storage.ListConversations(db, 10, 0)
	require.NoError(t, err)
	assert.Len(t, convs, 3)

	// Get latest should return the most recently updated
	latest, err := storage.GetLatestConversation(db)
	require.NoError(t, err)
	assert.NotNil(t, latest)
}

// generateConversationTitle generates a title from the first user message (test helper)
func generateConversationTitle(content string) string {
	title := content
	if len(title) > 50 {
		title = title[:47] + "..."
	}
	// Replace newlines with spaces
	newTitle := ""
	for _, ch := range title {
		if ch == '\n' {
			newTitle += " "
		} else {
			newTitle += string(ch)
		}
	}
	title = newTitle

	// Trim spaces
	for len(title) > 0 && (title[0] == ' ' || title[0] == '\t') {
		title = title[1:]
	}
	for len(title) > 0 && (title[len(title)-1] == ' ' || title[len(title)-1] == '\t') {
		title = title[:len(title)-1]
	}

	return title
}

func TestContinueFlagWithEmptyDatabase(t *testing.T) {
	// Test that --continue flag with empty database returns sql.ErrNoRows (not a real error)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "empty.db")

	db, err := openDatabase(dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Try to get latest conversation from empty database
	conv, err := storage.GetLatestConversation(db)

	// Should return sql.ErrNoRows (not a panic or corrupt DB error)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, conv)
}

func TestContinueFlagErrorDifferentiation(t *testing.T) {
	// Test that we differentiate between sql.ErrNoRows and real errors
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := openDatabase(dbPath)
	require.NoError(t, err)

	// Add a conversation
	conv := &storage.Conversation{
		ID:    "test-conv",
		Title: "Test",
		Model: "claude-sonnet-4-5-20250929",
	}
	err = storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Close database to simulate connection error
	_ = db.Close()

	// Now try to get latest conversation - should get a real error (not ErrNoRows)
	_, err = storage.GetLatestConversation(db)
	assert.Error(t, err)
	assert.NotEqual(t, sql.ErrNoRows, err, "Should be a real database error, not ErrNoRows")
}
