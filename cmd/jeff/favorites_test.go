// ABOUTME: Tests for favorite commands
// ABOUTME: Validates favorite toggling and listing functionality
package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harper/pagent/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFavoritesTestDB(t *testing.T) (*sql.DB, string) {
	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "pagent-favorites-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := storage.OpenDatabase(dbPath)
	require.NoError(t, err)

	return db, tmpDir
}

func TestFavoriteCommand(t *testing.T) {
	db, tmpDir := setupFavoritesTestDB(t)
	defer func() { _ = db.Close() }()
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a conversation
	conv := &storage.Conversation{
		ID:    "test-conv-1",
		Title: "Test Conversation",
		Model: "claude-sonnet-4-5-20250929",
	}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Verify initially not favorite
	retrieved, err := storage.GetConversation(db, "test-conv-1")
	require.NoError(t, err)
	assert.False(t, retrieved.IsFavorite)

	// Toggle to favorite
	err = storage.SetFavorite(db, "test-conv-1", true)
	require.NoError(t, err)

	retrieved, err = storage.GetConversation(db, "test-conv-1")
	require.NoError(t, err)
	assert.True(t, retrieved.IsFavorite)

	// Toggle back to not favorite
	err = storage.SetFavorite(db, "test-conv-1", false)
	require.NoError(t, err)

	retrieved, err = storage.GetConversation(db, "test-conv-1")
	require.NoError(t, err)
	assert.False(t, retrieved.IsFavorite)
}

func TestFavoritesListCommand(t *testing.T) {
	db, tmpDir := setupFavoritesTestDB(t)
	defer func() { _ = db.Close() }()
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create multiple conversations
	convs := []*storage.Conversation{
		{ID: "conv-1", Title: "First Chat", Model: "claude-sonnet-4-5-20250929"},
		{ID: "conv-2", Title: "Second Chat", Model: "claude-sonnet-4-5-20250929"},
		{ID: "conv-3", Title: "Third Chat", Model: "claude-sonnet-4-5-20250929"},
	}

	for _, conv := range convs {
		require.NoError(t, storage.CreateConversation(db, conv))
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Mark conv-1 and conv-3 as favorites
	require.NoError(t, storage.SetFavorite(db, "conv-1", true))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, storage.SetFavorite(db, "conv-3", true))

	// List favorites
	favorites, err := storage.ListFavorites(db)
	require.NoError(t, err)
	assert.Len(t, favorites, 2)

	// Verify order (most recently updated first)
	assert.Equal(t, "conv-3", favorites[0].ID)
	assert.Equal(t, "conv-1", favorites[1].ID)
}

func TestFavoritesListEmpty(t *testing.T) {
	db, tmpDir := setupFavoritesTestDB(t)
	defer func() { _ = db.Close() }()
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create conversation but don't favorite it
	conv := &storage.Conversation{
		ID:    "conv-1",
		Title: "Regular Chat",
		Model: "claude-sonnet-4-5-20250929",
	}
	require.NoError(t, storage.CreateConversation(db, conv))

	// List favorites should be empty
	favorites, err := storage.ListFavorites(db)
	require.NoError(t, err)
	assert.Len(t, favorites, 0)
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute), "1 minute ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1 hour ago"},
		{"3 hours ago", now.Add(-3 * time.Hour), "3 hours ago"},
		{"1 day ago", now.Add(-24 * time.Hour), "1 day ago"},
		{"3 days ago", now.Add(-3 * 24 * time.Hour), "3 days ago"},
		{"1 week ago", now.Add(-7 * 24 * time.Hour), "1 week ago"},
		{"2 weeks ago", now.Add(-14 * 24 * time.Hour), "2 weeks ago"},
		{"1 month ago", now.Add(-30 * 24 * time.Hour), "1 month ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRelativeTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}
