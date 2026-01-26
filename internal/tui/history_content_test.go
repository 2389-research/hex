// ABOUTME: Tests for HistoryContent component
// Tests session list rendering, keyboard navigation, selection, favorites, and deletion.

package tui

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/2389-research/tux/theme"
)

// createTestStorage creates a SessionStorage in a temp directory for testing.
func createTestStorage(t *testing.T) *SessionStorage {
	t.Helper()
	tmpDir := t.TempDir()
	storage, err := NewSessionStorage(tmpDir)
	require.NoError(t, err)
	return storage
}

// createTestSessions creates and saves test sessions to storage.
func createTestSessions(t *testing.T, storage *SessionStorage, count int) []*Session {
	t.Helper()
	sessions := make([]*Session, count)
	for i := 0; i < count; i++ {
		sessions[i] = &Session{
			ID:        filepath.Base(t.TempDir()) + "-" + string(rune('a'+i)),
			Title:     "Test Session " + string(rune('A'+i)),
			CreatedAt: time.Now().Add(-time.Duration(count-i) * time.Hour),
			UpdatedAt: time.Now().Add(-time.Duration(count-i) * time.Hour),
			Favorite:  false,
		}
		require.NoError(t, storage.Save(sessions[i]))
	}
	return sessions
}

// TestHistoryContent_View tests that the view renders session titles and dates.
func TestHistoryContent_View(t *testing.T) {
	storage := createTestStorage(t)
	th := theme.NewDraculaTheme()

	// Test empty state
	t.Run("empty state", func(t *testing.T) {
		content := NewHistoryContent(storage, th, nil)
		// Initialize by sending sessions loaded message with empty list
		content.Update(sessionsLoadedMsg{sessions: nil, err: nil})

		view := content.View()
		assert.Contains(t, view, "No sessions yet")
		assert.Contains(t, view, "n")
	})

	// Test with sessions
	t.Run("with sessions", func(t *testing.T) {
		sessions := createTestSessions(t, storage, 3)
		content := NewHistoryContent(storage, th, nil)
		content.SetSize(80, 20)

		// Load sessions into content
		content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

		view := content.View()

		// Should contain session titles
		for _, s := range sessions {
			assert.Contains(t, view, s.Title, "view should contain session title")
		}

		// Should show relative time (varies, but should have content)
		assert.NotEmpty(t, view)
	})

	// Test cursor indicator
	t.Run("cursor indicator", func(t *testing.T) {
		sessions := createTestSessions(t, storage, 2)
		content := NewHistoryContent(storage, th, nil)
		content.SetSize(80, 20)
		content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

		view := content.View()
		// First item should have cursor
		assert.Contains(t, view, ">", "should show cursor indicator")
	})
}

// TestHistoryContent_Navigation tests keyboard navigation (j, k, g, G).
func TestHistoryContent_Navigation(t *testing.T) {
	storage := createTestStorage(t)
	sessions := createTestSessions(t, storage, 5)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)
	content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

	// Initial cursor should be at 0
	assert.Equal(t, 0, content.cursor, "initial cursor should be at 0")

	// Test 'j' - move down
	t.Run("j moves down", func(t *testing.T) {
		content.cursor = 0
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		assert.Equal(t, 1, content.cursor, "j should move cursor down")

		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		assert.Equal(t, 2, content.cursor, "j should move cursor down again")
	})

	// Test 'k' - move up
	t.Run("k moves up", func(t *testing.T) {
		content.cursor = 2
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		assert.Equal(t, 1, content.cursor, "k should move cursor up")

		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		assert.Equal(t, 0, content.cursor, "k should move cursor up again")
	})

	// Test 'k' at top - should not go negative
	t.Run("k at top stays at 0", func(t *testing.T) {
		content.cursor = 0
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		assert.Equal(t, 0, content.cursor, "k at top should stay at 0")
	})

	// Test 'j' at bottom - should not exceed length
	t.Run("j at bottom stays at end", func(t *testing.T) {
		content.cursor = len(sessions) - 1
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		assert.Equal(t, len(sessions)-1, content.cursor, "j at bottom should stay at end")
	})

	// Test 'g' - go to top
	t.Run("g goes to top", func(t *testing.T) {
		content.cursor = 3
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
		assert.Equal(t, 0, content.cursor, "g should go to top")
	})

	// Test 'G' - go to bottom
	t.Run("G goes to bottom", func(t *testing.T) {
		content.cursor = 1
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
		assert.Equal(t, len(sessions)-1, content.cursor, "G should go to bottom")
	})

	// Test arrow keys
	t.Run("arrow keys work", func(t *testing.T) {
		content.cursor = 2
		content.Update(tea.KeyMsg{Type: tea.KeyUp})
		assert.Equal(t, 1, content.cursor, "up arrow should move cursor up")

		content.Update(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 2, content.cursor, "down arrow should move cursor down")
	})
}

// TestHistoryContent_Select tests session selection with Enter key.
func TestHistoryContent_Select(t *testing.T) {
	storage := createTestStorage(t)
	sessions := createTestSessions(t, storage, 3)
	th := theme.NewDraculaTheme()

	var selectedSession *Session
	callback := func(s *Session) {
		selectedSession = s
	}

	content := NewHistoryContent(storage, th, callback)
	content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

	// Select first session
	t.Run("enter selects current session", func(t *testing.T) {
		content.cursor = 0
		selectedSession = nil

		_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// Callback should have been called
		assert.NotNil(t, selectedSession, "callback should be called")
		assert.Equal(t, sessions[0].ID, selectedSession.ID, "should select first session")

		// Command should return SessionSelectedMsg
		if cmd != nil {
			msg := cmd()
			if selMsg, ok := msg.(SessionSelectedMsg); ok {
				assert.Equal(t, sessions[0].ID, selMsg.Session.ID)
			}
		}
	})

	// Select different session
	t.Run("select different session", func(t *testing.T) {
		content.cursor = 2
		selectedSession = nil

		content.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.NotNil(t, selectedSession, "callback should be called")
		assert.Equal(t, sessions[2].ID, selectedSession.ID, "should select third session")
	})

	// Test without callback
	t.Run("works without callback", func(t *testing.T) {
		contentNoCallback := NewHistoryContent(storage, th, nil)
		contentNoCallback.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

		// Should not panic
		_, cmd := contentNoCallback.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// Should still return message
		if cmd != nil {
			msg := cmd()
			_, ok := msg.(SessionSelectedMsg)
			assert.True(t, ok, "should return SessionSelectedMsg")
		}
	})

	// Test empty list
	t.Run("select on empty list does nothing", func(t *testing.T) {
		emptyContent := NewHistoryContent(storage, th, callback)
		emptyContent.Update(sessionsLoadedMsg{sessions: nil, err: nil})
		selectedSession = nil

		_, cmd := emptyContent.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Nil(t, selectedSession, "callback should not be called for empty list")
		assert.Nil(t, cmd, "command should be nil for empty list")
	})
}

// TestHistoryContent_Favorite tests toggling favorite status with 'f' key.
func TestHistoryContent_Favorite(t *testing.T) {
	storage := createTestStorage(t)
	sessions := createTestSessions(t, storage, 3)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)
	content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

	// Toggle favorite on
	t.Run("f toggles favorite on", func(t *testing.T) {
		content.cursor = 1
		assert.False(t, sessions[1].Favorite, "session should not be favorite initially")

		_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

		assert.True(t, sessions[1].Favorite, "session should be favorite after toggle")

		// Verify command returns correct message
		if cmd != nil {
			msg := cmd()
			if favMsg, ok := msg.(SessionFavoriteToggledMsg); ok {
				assert.Equal(t, sessions[1].ID, favMsg.Session.ID)
				assert.True(t, favMsg.Session.Favorite)
			}
		}

		// Verify saved to storage
		loaded, err := storage.Load(sessions[1].ID)
		require.NoError(t, err)
		assert.True(t, loaded.Favorite, "favorite status should be saved to storage")
	})

	// Toggle favorite off
	t.Run("f toggles favorite off", func(t *testing.T) {
		content.cursor = 1
		assert.True(t, sessions[1].Favorite, "session should be favorite")

		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

		assert.False(t, sessions[1].Favorite, "session should not be favorite after toggle")

		// Verify saved to storage
		loaded, err := storage.Load(sessions[1].ID)
		require.NoError(t, err)
		assert.False(t, loaded.Favorite, "favorite status should be saved to storage")
	})

	// Test empty list
	t.Run("f on empty list does nothing", func(t *testing.T) {
		emptyContent := NewHistoryContent(storage, th, nil)
		emptyContent.Update(sessionsLoadedMsg{sessions: nil, err: nil})

		_, cmd := emptyContent.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

		assert.Nil(t, cmd, "command should be nil for empty list")
	})
}

// TestHistoryContent_Delete tests delete confirmation flow with 'd' key.
func TestHistoryContent_Delete(t *testing.T) {
	th := theme.NewDraculaTheme()

	// Test single 'd' enters confirm state
	t.Run("first d enters confirm state", func(t *testing.T) {
		storage := createTestStorage(t)
		sessions := createTestSessions(t, storage, 3)
		content := NewHistoryContent(storage, th, nil)
		content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

		content.cursor = 1
		assert.False(t, content.deleteConfirm, "should not be in confirm state initially")

		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

		assert.True(t, content.deleteConfirm, "should be in confirm state after first d")
		assert.Equal(t, 1, content.deleteTarget, "delete target should match cursor")
	})

	// Test view shows confirmation prompt
	t.Run("view shows delete confirmation", func(t *testing.T) {
		storage := createTestStorage(t)
		sessions := createTestSessions(t, storage, 3)
		content := NewHistoryContent(storage, th, nil)
		content.SetSize(80, 20)
		content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

		content.cursor = 1
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

		view := content.View()
		assert.Contains(t, view, "press 'd' again to delete", "should show delete confirmation")
	})

	// Test second 'd' confirms deletion
	t.Run("second d confirms deletion", func(t *testing.T) {
		storage := createTestStorage(t)
		sessions := createTestSessions(t, storage, 3)
		content := NewHistoryContent(storage, th, nil)
		content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

		content.cursor = 1
		deletedID := sessions[1].ID

		// First d
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		assert.True(t, content.deleteConfirm)

		// Second d
		_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

		assert.False(t, content.deleteConfirm, "should exit confirm state after deletion")
		assert.Len(t, content.sessions, 2, "should have one less session")

		// Verify session was removed from list
		for _, s := range content.sessions {
			assert.NotEqual(t, deletedID, s.ID, "deleted session should not be in list")
		}

		// Verify command returns correct message
		if cmd != nil {
			msg := cmd()
			if delMsg, ok := msg.(SessionDeletedMsg); ok {
				assert.Equal(t, deletedID, delMsg.SessionID)
			}
		}

		// Verify deleted from storage
		_, err := storage.Load(deletedID)
		assert.Error(t, err, "session should be deleted from storage")
	})

	// Test other key cancels deletion
	t.Run("other key cancels deletion", func(t *testing.T) {
		storage := createTestStorage(t)
		sessions := createTestSessions(t, storage, 3)
		content := NewHistoryContent(storage, th, nil)
		content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

		content.cursor = 1

		// First d
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		assert.True(t, content.deleteConfirm)

		// Press any other key (like 'j')
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

		assert.False(t, content.deleteConfirm, "should cancel confirm state")
		assert.Len(t, content.sessions, 3, "should still have all sessions")
	})

	// Test escape cancels deletion
	t.Run("escape cancels deletion", func(t *testing.T) {
		storage := createTestStorage(t)
		sessions := createTestSessions(t, storage, 3)
		content := NewHistoryContent(storage, th, nil)
		content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

		content.cursor = 1

		// First d
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		assert.True(t, content.deleteConfirm)

		// Press escape
		content.Update(tea.KeyMsg{Type: tea.KeyEsc})

		assert.False(t, content.deleteConfirm, "should cancel confirm state")
		assert.Len(t, content.sessions, 3, "should still have all sessions")
	})

	// Test cursor adjustment after delete
	t.Run("cursor adjusts after deleting last item", func(t *testing.T) {
		storage := createTestStorage(t)
		sessions := createTestSessions(t, storage, 3)
		content := NewHistoryContent(storage, th, nil)
		content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

		// Delete last item
		content.cursor = 2
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

		assert.Equal(t, 1, content.cursor, "cursor should adjust to stay in bounds")
	})

	// Test delete on empty list
	t.Run("delete on empty list does nothing", func(t *testing.T) {
		storage := createTestStorage(t)
		content := NewHistoryContent(storage, th, nil)
		content.Update(sessionsLoadedMsg{sessions: nil, err: nil})

		_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

		assert.Nil(t, cmd, "command should be nil for empty list")
		assert.False(t, content.deleteConfirm, "should not enter confirm state")
	})
}

// TestHistoryContent_NewSession tests 'n' key for new session request.
func TestHistoryContent_NewSession(t *testing.T) {
	storage := createTestStorage(t)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)
	content.Update(sessionsLoadedMsg{sessions: nil, err: nil})

	_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	require.NotNil(t, cmd, "should return command")
	msg := cmd()
	_, ok := msg.(NewSessionRequestedMsg)
	assert.True(t, ok, "should return NewSessionRequestedMsg")
}

// TestHistoryContent_Refresh tests 'r' key for refresh.
func TestHistoryContent_Refresh(t *testing.T) {
	storage := createTestStorage(t)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)

	// Create initial sessions
	sessions := createTestSessions(t, storage, 2)
	content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})
	assert.Len(t, content.sessions, 2)

	// Add another session to storage directly
	newSession := &Session{
		ID:        "new-session",
		Title:     "New Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, storage.Save(newSession))

	// Press 'r' to refresh
	_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	require.NotNil(t, cmd, "should return refresh command")

	// Execute the command to get the message
	msg := cmd()
	loadedMsg, ok := msg.(sessionsLoadedMsg)
	require.True(t, ok, "should return sessionsLoadedMsg")
	require.NoError(t, loadedMsg.err)

	// Update with loaded sessions
	content.Update(loadedMsg)
	assert.Len(t, content.sessions, 3, "should have 3 sessions after refresh")
}

// TestHistoryContent_SelectedSession tests SelectedSession accessor.
func TestHistoryContent_SelectedSession(t *testing.T) {
	storage := createTestStorage(t)
	sessions := createTestSessions(t, storage, 3)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)
	content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

	// Test getting selected session
	t.Run("returns selected session", func(t *testing.T) {
		content.cursor = 1
		selected := content.SelectedSession()
		require.NotNil(t, selected)
		assert.Equal(t, sessions[1].ID, selected.ID)
	})

	// Test empty list
	t.Run("returns nil for empty list", func(t *testing.T) {
		emptyContent := NewHistoryContent(storage, th, nil)
		emptyContent.Update(sessionsLoadedMsg{sessions: nil, err: nil})

		selected := emptyContent.SelectedSession()
		assert.Nil(t, selected)
	})
}

// TestHistoryContent_Sessions tests Sessions accessor.
func TestHistoryContent_Sessions(t *testing.T) {
	storage := createTestStorage(t)
	sessions := createTestSessions(t, storage, 3)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)
	content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

	result := content.Sessions()
	assert.Len(t, result, 3)
}

// TestHistoryContent_Value tests Value accessor (implements content.Content).
func TestHistoryContent_Value(t *testing.T) {
	storage := createTestStorage(t)
	sessions := createTestSessions(t, storage, 3)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)
	content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

	value := content.Value()
	result, ok := value.([]*Session)
	require.True(t, ok, "Value should return []*Session")
	assert.Len(t, result, 3)
}

// TestHistoryContent_SetSize tests SetSize method.
func TestHistoryContent_SetSize(t *testing.T) {
	storage := createTestStorage(t)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)

	content.SetSize(100, 50)
	assert.Equal(t, 100, content.width)
	assert.Equal(t, 50, content.height)
}

// TestHistoryContent_OnActivate tests OnActivate returns refresh command.
func TestHistoryContent_OnActivate(t *testing.T) {
	storage := createTestStorage(t)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)

	cmd := content.OnActivate()
	require.NotNil(t, cmd, "OnActivate should return a command")

	// Execute command and check result
	msg := cmd()
	_, ok := msg.(sessionsLoadedMsg)
	assert.True(t, ok, "should return sessionsLoadedMsg")
}

// TestHistoryContent_OnDeactivate tests OnDeactivate clears delete state.
func TestHistoryContent_OnDeactivate(t *testing.T) {
	storage := createTestStorage(t)
	sessions := createTestSessions(t, storage, 3)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)
	content.Update(sessionsLoadedMsg{sessions: sessions, err: nil})

	// Enter delete confirm state
	content.cursor = 1
	content.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.True(t, content.deleteConfirm)
	assert.Equal(t, 1, content.deleteTarget)

	// Deactivate
	content.OnDeactivate()

	assert.False(t, content.deleteConfirm, "delete confirm should be cleared")
	assert.Equal(t, -1, content.deleteTarget, "delete target should be reset")
}

// TestHistoryContent_Init tests Init returns load command.
func TestHistoryContent_Init(t *testing.T) {
	storage := createTestStorage(t)
	th := theme.NewDraculaTheme()
	content := NewHistoryContent(storage, th, nil)

	cmd := content.Init()
	require.NotNil(t, cmd, "Init should return a command")

	msg := cmd()
	_, ok := msg.(sessionsLoadedMsg)
	assert.True(t, ok, "should return sessionsLoadedMsg")
}

// TestFormatRelativeTime tests the relative time formatting function.
func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"just now", now, "just now"},
		{"30 seconds ago", now.Add(-30 * time.Second), "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute), "1 minute ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1 hour ago"},
		{"3 hours ago", now.Add(-3 * time.Hour), "3 hours ago"},
		{"yesterday", now.Add(-36 * time.Hour), "yesterday"},
		{"3 days ago", now.Add(-3 * 24 * time.Hour), "3 days ago"},
		{"1 week ago", now.Add(-10 * 24 * time.Hour), "1 week ago"},
		{"2 weeks ago", now.Add(-14 * 24 * time.Hour), "2 weeks ago"},
		{"1 month ago", now.Add(-35 * 24 * time.Hour), "1 month ago"},
		{"3 months ago", now.Add(-100 * 24 * time.Hour), "3 months ago"},
		{"1 year ago", now.Add(-400 * 24 * time.Hour), "1 year ago"},
		{"2 years ago", now.Add(-800 * 24 * time.Hour), "2 years ago"},
		{"future (edge case)", now.Add(1 * time.Hour), "just now"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRelativeTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}
