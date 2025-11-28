// ABOUTME: Tests for conversation CRUD operations
// ABOUTME: Validates conversation creation, retrieval, listing, and deletion
package storage_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/harper/clem/internal/storage"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Enable foreign keys (required for CASCADE)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	err = storage.InitializeSchema(db)
	require.NoError(t, err)

	return db
}

func TestCreateConversation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	conv := &storage.Conversation{
		ID:    "conv-123",
		Title: "Test Chat",
		Model: "claude-sonnet-4-5-20250929",
	}

	err := storage.CreateConversation(db, conv)
	require.NoError(t, err)

	// Verify it was created
	retrieved, err := storage.GetConversation(db, "conv-123")
	require.NoError(t, err)
	assert.Equal(t, "Test Chat", retrieved.Title)
	assert.Equal(t, "claude-sonnet-4-5-20250929", retrieved.Model)
}

func TestListConversations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create multiple conversations
	conv1 := &storage.Conversation{ID: "conv-1", Title: "Chat 1", Model: "claude-sonnet-4-5-20250929"}
	conv2 := &storage.Conversation{ID: "conv-2", Title: "Chat 2", Model: "claude-sonnet-4-5-20250929"}

	require.NoError(t, storage.CreateConversation(db, conv1))
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	require.NoError(t, storage.CreateConversation(db, conv2))

	// List conversations (should be ordered by updated_at DESC)
	convs, err := storage.ListConversations(db, 10, 0)
	require.NoError(t, err)
	assert.Len(t, convs, 2)
	assert.Equal(t, "conv-2", convs[0].ID) // Most recent first
	assert.Equal(t, "conv-1", convs[1].ID)
}

func TestGetLatestConversation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create conversations
	conv1 := &storage.Conversation{ID: "conv-1", Title: "Chat 1", Model: "claude-sonnet-4-5-20250929"}
	conv2 := &storage.Conversation{ID: "conv-2", Title: "Chat 2", Model: "claude-sonnet-4-5-20250929"}

	require.NoError(t, storage.CreateConversation(db, conv1))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, storage.CreateConversation(db, conv2))

	// Get latest
	latest, err := storage.GetLatestConversation(db)
	require.NoError(t, err)
	assert.Equal(t, "conv-2", latest.ID)
}

func TestUpdateConversationTitle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	conv := &storage.Conversation{ID: "conv-1", Title: "Original", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Update title
	err := storage.UpdateConversationTitle(db, "conv-1", "Updated Title")
	require.NoError(t, err)

	// Verify
	retrieved, err := storage.GetConversation(db, "conv-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", retrieved.Title)
}

func TestDeleteConversation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	conv := &storage.Conversation{ID: "conv-1", Title: "To Delete", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Delete
	err := storage.DeleteConversation(db, "conv-1")
	require.NoError(t, err)

	// Verify it's gone
	_, err = storage.GetConversation(db, "conv-1")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteConversationCascade(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create conversation
	conv := &storage.Conversation{ID: "conv-cascade", Title: "Cascade Test", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Add two messages
	msg1 := &storage.Message{ConversationID: "conv-cascade", Role: "user", Content: "Hello"}
	msg2 := &storage.Message{ConversationID: "conv-cascade", Role: "assistant", Content: "Hi there"}
	require.NoError(t, storage.CreateMessage(db, msg1))
	require.NoError(t, storage.CreateMessage(db, msg2))

	// Verify messages exist
	messages, err := storage.ListMessages(db, "conv-cascade")
	require.NoError(t, err)
	assert.Len(t, messages, 2)

	// Delete conversation
	err = storage.DeleteConversation(db, "conv-cascade")
	require.NoError(t, err)

	// Verify messages were deleted via CASCADE
	messages, err = storage.ListMessages(db, "conv-cascade")
	require.NoError(t, err)
	assert.Len(t, messages, 0, "messages should be deleted via CASCADE")
}
