// ABOUTME: Tests for conversation CRUD operations
// ABOUTME: Validates conversation creation, retrieval, listing, and deletion
package storage_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/harper/clem/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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

func TestSetFavorite(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversation
	conv := &storage.Conversation{ID: "conv-fav", Title: "Favorite Test", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Initially not favorite
	retrieved, err := storage.GetConversation(db, "conv-fav")
	require.NoError(t, err)
	assert.False(t, retrieved.IsFavorite)

	// Set as favorite
	err = storage.SetFavorite(db, "conv-fav", true)
	require.NoError(t, err)

	// Verify it's now favorite
	retrieved, err = storage.GetConversation(db, "conv-fav")
	require.NoError(t, err)
	assert.True(t, retrieved.IsFavorite)

	// Unset favorite
	err = storage.SetFavorite(db, "conv-fav", false)
	require.NoError(t, err)

	// Verify it's no longer favorite
	retrieved, err = storage.GetConversation(db, "conv-fav")
	require.NoError(t, err)
	assert.False(t, retrieved.IsFavorite)
}

func TestListFavorites(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create multiple conversations
	conv1 := &storage.Conversation{ID: "conv-1", Title: "Chat 1", Model: "claude-sonnet-4-5-20250929"}
	conv2 := &storage.Conversation{ID: "conv-2", Title: "Chat 2", Model: "claude-sonnet-4-5-20250929"}
	conv3 := &storage.Conversation{ID: "conv-3", Title: "Chat 3", Model: "claude-sonnet-4-5-20250929"}

	require.NoError(t, storage.CreateConversation(db, conv1))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, storage.CreateConversation(db, conv2))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, storage.CreateConversation(db, conv3))

	// Mark conv-1 and conv-3 as favorites
	require.NoError(t, storage.SetFavorite(db, "conv-1", true))
	time.Sleep(10 * time.Millisecond) // Ensure different updated_at
	require.NoError(t, storage.SetFavorite(db, "conv-3", true))

	// List favorites
	favorites, err := storage.ListFavorites(db)
	require.NoError(t, err)
	assert.Len(t, favorites, 2)
	assert.Equal(t, "conv-3", favorites[0].ID) // Most recently updated first
	assert.Equal(t, "conv-1", favorites[1].ID)
	assert.True(t, favorites[0].IsFavorite)
	assert.True(t, favorites[1].IsFavorite)
}

func TestListFavoritesEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversations but don't favorite any
	conv1 := &storage.Conversation{ID: "conv-1", Title: "Chat 1", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv1))

	// List favorites should be empty
	favorites, err := storage.ListFavorites(db)
	require.NoError(t, err)
	assert.Len(t, favorites, 0)
}

func TestFavoriteBackwardCompatibility(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversation and verify is_favorite defaults to false
	conv := &storage.Conversation{ID: "conv-compat", Title: "Backward Compat", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Verify IsFavorite is false by default
	retrieved, err := storage.GetConversation(db, "conv-compat")
	require.NoError(t, err)
	assert.False(t, retrieved.IsFavorite)
}
