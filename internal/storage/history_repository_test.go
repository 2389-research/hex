// ABOUTME: Tests for history repository with FTS5 search
// ABOUTME: Validates CRUD operations and full-text search functionality
package storage_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/hex/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddHistoryEntry(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create a test conversation first
	conv := &storage.Conversation{
		ID:    "test-conv-" + uuid.New().String(),
		Title: "Test",
		Model: "test-model",
	}
	err := storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Add history entry
	entry := &storage.HistoryEntry{
		ID:                uuid.New().String(),
		ConversationID:    conv.ID,
		UserMessage:       "How do I use Docker?",
		AssistantResponse: "Docker is a container platform...",
	}

	err = storage.AddHistoryEntry(db, entry)
	assert.NoError(t, err)

	// Verify it was saved
	rows, err := db.Query("SELECT id, user_message, assistant_response FROM history WHERE id = ?", entry.ID)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	assert.True(t, rows.Next())
	var id, userMsg, assistantResp string
	err = rows.Scan(&id, &userMsg, &assistantResp)
	require.NoError(t, err)

	assert.Equal(t, entry.ID, id)
	assert.Equal(t, entry.UserMessage, userMsg)
	assert.Equal(t, entry.AssistantResponse, assistantResp)
}

func TestSearchHistory(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversation
	conv := &storage.Conversation{
		ID:    "test-conv-" + uuid.New().String(),
		Title: "Test",
		Model: "test-model",
	}
	err := storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Add several history entries
	entries := []*storage.HistoryEntry{
		{
			ID:                uuid.New().String(),
			ConversationID:    conv.ID,
			UserMessage:       "How do I use Docker containers?",
			AssistantResponse: "Docker containers are...",
		},
		{
			ID:                uuid.New().String(),
			ConversationID:    conv.ID,
			UserMessage:       "What is Kubernetes?",
			AssistantResponse: "Kubernetes is an orchestration platform...",
		},
		{
			ID:                uuid.New().String(),
			ConversationID:    conv.ID,
			UserMessage:       "How to debug Python code?",
			AssistantResponse: "Use pdb for debugging...",
		},
	}

	for _, entry := range entries {
		err := storage.AddHistoryEntry(db, entry)
		require.NoError(t, err)
	}

	// Search for "docker"
	results, err := storage.SearchHistory(db, "docker", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].UserMessage, "Docker")

	// Search for "kubernetes"
	results, err = storage.SearchHistory(db, "kubernetes", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].UserMessage, "Kubernetes")

	// Search for "debug"
	results, err = storage.SearchHistory(db, "debug", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].UserMessage, "debug")
}

func TestSearchHistoryMultipleMatches(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversation
	conv := &storage.Conversation{
		ID:    "test-conv-" + uuid.New().String(),
		Title: "Test",
		Model: "test-model",
	}
	err := storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Add entries with "python" in different contexts
	entries := []*storage.HistoryEntry{
		{
			ID:                uuid.New().String(),
			ConversationID:    conv.ID,
			UserMessage:       "How to install Python?",
			AssistantResponse: "Use brew install python...",
		},
		{
			ID:                uuid.New().String(),
			ConversationID:    conv.ID,
			UserMessage:       "Python decorators explained",
			AssistantResponse: "Decorators are a way to modify functions...",
		},
		{
			ID:                uuid.New().String(),
			ConversationID:    conv.ID,
			UserMessage:       "Best Python libraries",
			AssistantResponse: "Here are some great libraries...",
		},
	}

	for _, entry := range entries {
		err := storage.AddHistoryEntry(db, entry)
		require.NoError(t, err)
	}

	// Search should return all three
	results, err := storage.SearchHistory(db, "python", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestGetRecentHistory(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversation
	conv := &storage.Conversation{
		ID:    "test-conv-" + uuid.New().String(),
		Title: "Test",
		Model: "test-model",
	}
	err := storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Add several entries with slight delays to ensure ordering
	for i := 0; i < 5; i++ {
		entry := &storage.HistoryEntry{
			ID:                uuid.New().String(),
			ConversationID:    conv.ID,
			UserMessage:       "Message " + string(rune('A'+i)),
			AssistantResponse: "Response " + string(rune('A'+i)),
		}
		err := storage.AddHistoryEntry(db, entry)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get recent history
	results, err := storage.GetRecentHistory(db, 3)
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// Should be in reverse chronological order (most recent first)
	assert.Contains(t, results[0].UserMessage, "Message E")
	assert.Contains(t, results[1].UserMessage, "Message D")
	assert.Contains(t, results[2].UserMessage, "Message C")
}

func TestSearchHistoryLimit(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversation
	conv := &storage.Conversation{
		ID:    "test-conv-" + uuid.New().String(),
		Title: "Test",
		Model: "test-model",
	}
	err := storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Add many entries with same keyword
	for i := 0; i < 10; i++ {
		entry := &storage.HistoryEntry{
			ID:                uuid.New().String(),
			ConversationID:    conv.ID,
			UserMessage:       "How to use Docker feature " + string(rune('A'+i)),
			AssistantResponse: "Docker feature explanation...",
		}
		err := storage.AddHistoryEntry(db, entry)
		require.NoError(t, err)
	}

	// Search with limit
	results, err := storage.SearchHistory(db, "docker", 5)
	assert.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestSearchHistoryNoResults(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Search empty database
	results, err := storage.SearchHistory(db, "nonexistent", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestSearchHistorySpecialCharacters(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversation
	conv := &storage.Conversation{
		ID:    "test-conv-" + uuid.New().String(),
		Title: "Test",
		Model: "test-model",
	}
	err := storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Add entry with special characters
	entry := &storage.HistoryEntry{
		ID:                uuid.New().String(),
		ConversationID:    conv.ID,
		UserMessage:       "How to use && and || operators?",
		AssistantResponse: "These are logical operators...",
	}
	err = storage.AddHistoryEntry(db, entry)
	require.NoError(t, err)

	// Search should handle special characters
	results, err := storage.SearchHistory(db, "operators", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestHistoryCascadeDelete(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversation
	conv := &storage.Conversation{
		ID:    "test-conv-" + uuid.New().String(),
		Title: "Test",
		Model: "test-model",
	}
	err := storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Add history entry
	entry := &storage.HistoryEntry{
		ID:                uuid.New().String(),
		ConversationID:    conv.ID,
		UserMessage:       "Test message",
		AssistantResponse: "Test response",
	}
	err = storage.AddHistoryEntry(db, entry)
	require.NoError(t, err)

	// Delete conversation
	_, err = db.Exec("DELETE FROM conversations WHERE id = ?", conv.ID)
	require.NoError(t, err)

	// History should be deleted too (cascade)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM history WHERE conversation_id = ?", conv.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
