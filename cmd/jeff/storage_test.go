// ABOUTME: Tests for database initialization and management in CLI
// ABOUTME: Validates database opening, schema setup, and error handling
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harper/pagent/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDatabase(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := openDatabase(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func() { _ = db.Close() }()

	// Verify database file was created
	_, err = os.Stat(dbPath)
	require.NoError(t, err, "database file should exist")

	// Verify schema was initialized
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='conversations'").Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "conversations", tableName)
}

func TestOpenDatabaseCreatesDirectory(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nested", "dir", "test.db")

	// Open database (should create nested directories)
	db, err := openDatabase(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func() { _ = db.Close() }()

	// Verify nested directories were created
	_, err = os.Stat(filepath.Dir(dbPath))
	require.NoError(t, err)
}

func TestDefaultDBPath(t *testing.T) {
	path := defaultDBPath()

	// Should end with .pagen/pagent.db
	assert.Contains(t, path, ".pagen")
	assert.Contains(t, path, "pagent.db")

	// Should be absolute path
	assert.True(t, filepath.IsAbs(path))
}

func TestLoadConversationHistory(t *testing.T) {
	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := openDatabase(dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create test conversation
	testConv := &storage.Conversation{
		ID:    "test-conv-123",
		Title: "Test Conversation",
		Model: "claude-sonnet-4-5-20250929",
	}
	err = storage.CreateConversation(db, testConv)
	require.NoError(t, err)

	// Create test messages
	msg1 := &storage.Message{
		ID:             "msg-1",
		ConversationID: "test-conv-123",
		Role:           "user",
		Content:        "Hello",
	}
	msg2 := &storage.Message{
		ID:             "msg-2",
		ConversationID: "test-conv-123",
		Role:           "assistant",
		Content:        "Hi there!",
	}
	require.NoError(t, storage.CreateMessage(db, msg1))
	require.NoError(t, storage.CreateMessage(db, msg2))

	// Load conversation history
	conv, messages, err := loadConversationHistory(db, "test-conv-123")
	require.NoError(t, err)
	require.NotNil(t, conv)
	assert.Equal(t, "Test Conversation", conv.Title)
	assert.Len(t, messages, 2)
	assert.Equal(t, "Hello", messages[0].Content)
	assert.Equal(t, "Hi there!", messages[1].Content)
}
