// ABOUTME: Tests for SessionStorage persistence layer
// ABOUTME: Verifies save/load round-trip, list ordering, delete, and error cases

package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessionStorage(t *testing.T) {
	t.Run("creates directory if not exists", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "sessions")

		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)
		require.NotNil(t, storage)

		// Verify directory was created
		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("works with existing directory", func(t *testing.T) {
		dir := t.TempDir()

		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)
		require.NotNil(t, storage)
	})
}

func TestSessionStorage_SaveLoad(t *testing.T) {
	t.Run("round-trip with all fields", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		// Create a session with all fields populated
		now := time.Now().Truncate(time.Millisecond) // JSON loses sub-millisecond precision
		original := &Session{
			ID:        "test-session-123",
			Title:     "Test Session Title",
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now,
			Favorite:  true,
			Messages: []SessionMessage{
				{
					Role:      "user",
					Content:   "Hello, can you help me?",
					Timestamp: now.Add(-30 * time.Minute),
				},
				{
					Role:      "assistant",
					Content:   "Of course! I'd be happy to help.",
					Timestamp: now.Add(-29 * time.Minute),
					ToolCalls: []SessionToolCall{
						{
							ID:     "tool_001",
							Name:   "read_file",
							Input:  map[string]interface{}{"path": "/test.txt"},
							Output: "file contents here",
							Error:  false,
						},
						{
							ID:     "tool_002",
							Name:   "bash",
							Input:  map[string]interface{}{"command": "ls -la"},
							Output: "error: permission denied",
							Error:  true,
						},
					},
				},
			},
		}

		// Save the session
		err = storage.Save(original)
		require.NoError(t, err)

		// Verify file was created
		filename := filepath.Join(dir, original.ID+".json")
		_, err = os.Stat(filename)
		require.NoError(t, err)

		// Load the session
		loaded, err := storage.Load(original.ID)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// Verify all fields match
		assert.Equal(t, original.ID, loaded.ID)
		assert.Equal(t, original.Title, loaded.Title)
		assert.True(t, original.CreatedAt.Equal(loaded.CreatedAt), "CreatedAt mismatch")
		assert.True(t, original.UpdatedAt.Equal(loaded.UpdatedAt), "UpdatedAt mismatch")
		assert.Equal(t, original.Favorite, loaded.Favorite)

		// Verify messages
		require.Len(t, loaded.Messages, 2)
		assert.Equal(t, original.Messages[0].Role, loaded.Messages[0].Role)
		assert.Equal(t, original.Messages[0].Content, loaded.Messages[0].Content)
		assert.True(t, original.Messages[0].Timestamp.Equal(loaded.Messages[0].Timestamp))

		// Verify tool calls
		require.Len(t, loaded.Messages[1].ToolCalls, 2)
		assert.Equal(t, original.Messages[1].ToolCalls[0].ID, loaded.Messages[1].ToolCalls[0].ID)
		assert.Equal(t, original.Messages[1].ToolCalls[0].Name, loaded.Messages[1].ToolCalls[0].Name)
		assert.Equal(t, original.Messages[1].ToolCalls[0].Input["path"], loaded.Messages[1].ToolCalls[0].Input["path"])
		assert.Equal(t, original.Messages[1].ToolCalls[0].Output, loaded.Messages[1].ToolCalls[0].Output)
		assert.Equal(t, original.Messages[1].ToolCalls[0].Error, loaded.Messages[1].ToolCalls[0].Error)

		// Verify error tool call
		assert.Equal(t, original.Messages[1].ToolCalls[1].Error, loaded.Messages[1].ToolCalls[1].Error)
	})

	t.Run("empty messages", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		original := &Session{
			ID:        "empty-session",
			Title:     "Empty Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []SessionMessage{},
		}

		err = storage.Save(original)
		require.NoError(t, err)

		loaded, err := storage.Load(original.ID)
		require.NoError(t, err)
		assert.Empty(t, loaded.Messages)
	})

	t.Run("overwrite existing session", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		// Save initial version
		session := &Session{
			ID:        "update-session",
			Title:     "Original Title",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []SessionMessage{},
		}
		err = storage.Save(session)
		require.NoError(t, err)

		// Update and save again
		session.Title = "Updated Title"
		session.UpdatedAt = time.Now()
		err = storage.Save(session)
		require.NoError(t, err)

		// Load and verify update
		loaded, err := storage.Load(session.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", loaded.Title)
	})
}

func TestSessionStorage_List(t *testing.T) {
	t.Run("returns sessions in newest-first order", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		now := time.Now()

		// Create sessions with different UpdatedAt times
		sessions := []*Session{
			{
				ID:        "oldest",
				Title:     "Oldest Session",
				CreatedAt: now.Add(-3 * time.Hour),
				UpdatedAt: now.Add(-3 * time.Hour),
			},
			{
				ID:        "newest",
				Title:     "Newest Session",
				CreatedAt: now.Add(-1 * time.Hour),
				UpdatedAt: now, // Most recent update
			},
			{
				ID:        "middle",
				Title:     "Middle Session",
				CreatedAt: now.Add(-2 * time.Hour),
				UpdatedAt: now.Add(-1 * time.Hour),
			},
		}

		// Save in random order
		for _, s := range sessions {
			err := storage.Save(s)
			require.NoError(t, err)
		}

		// List should return newest-first
		listed, err := storage.List()
		require.NoError(t, err)
		require.Len(t, listed, 3)

		assert.Equal(t, "newest", listed[0].ID)
		assert.Equal(t, "middle", listed[1].ID)
		assert.Equal(t, "oldest", listed[2].ID)
	})

	t.Run("returns empty slice when no sessions", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		listed, err := storage.List()
		require.NoError(t, err)
		assert.Empty(t, listed)
	})

	t.Run("skips invalid JSON files", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		// Create a valid session
		validSession := &Session{
			ID:        "valid",
			Title:     "Valid Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = storage.Save(validSession)
		require.NoError(t, err)

		// Create an invalid JSON file
		invalidFile := filepath.Join(dir, "invalid.json")
		err = os.WriteFile(invalidFile, []byte("not valid json{{{"), 0644)
		require.NoError(t, err)

		// List should return only the valid session
		listed, err := storage.List()
		require.NoError(t, err)
		require.Len(t, listed, 1)
		assert.Equal(t, "valid", listed[0].ID)
	})
}

func TestSessionStorage_Delete(t *testing.T) {
	t.Run("deletes existing session", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		// Create and save a session
		session := &Session{
			ID:        "to-delete",
			Title:     "Session to Delete",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = storage.Save(session)
		require.NoError(t, err)

		// Verify file exists
		filename := filepath.Join(dir, session.ID+".json")
		_, err = os.Stat(filename)
		require.NoError(t, err)

		// Delete the session
		err = storage.Delete(session.ID)
		require.NoError(t, err)

		// Verify file no longer exists
		_, err = os.Stat(filename)
		assert.True(t, os.IsNotExist(err))

		// Verify Load returns error
		_, err = storage.Load(session.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("delete non-existent session returns error", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		err = storage.Delete("does-not-exist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})
}

func TestSessionStorage_ErrorCases(t *testing.T) {
	t.Run("save nil session", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		err = storage.Save(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil session")
	})

	t.Run("save session with empty ID", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		session := &Session{
			ID:    "",
			Title: "No ID",
		}
		err = storage.Save(session)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty ID")
	})

	t.Run("load session with empty ID", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		_, err = storage.Load("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty ID")
	})

	t.Run("load non-existent session", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		_, err = storage.Load("non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("delete session with empty ID", func(t *testing.T) {
		dir := t.TempDir()
		storage, err := NewSessionStorage(dir)
		require.NoError(t, err)

		err = storage.Delete("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty ID")
	})
}
