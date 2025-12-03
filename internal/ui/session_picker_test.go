// Package ui provides TUI components including the session picker for resuming conversations.
// ABOUTME: Tests for session picker component
package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/pagent/internal/storage"
)

func TestSessionItem_Title(t *testing.T) {
	tests := []struct {
		name     string
		item     sessionItem
		expected string
	}{
		{
			name:     "new session item",
			item:     sessionItem{isNewItem: true},
			expected: "✨ New Session",
		},
		{
			name: "regular conversation",
			item: sessionItem{
				conv: &storage.Conversation{
					ID:    "conv-123",
					Title: "My Conversation",
				},
			},
			expected: "My Conversation",
		},
		{
			name: "favorite conversation",
			item: sessionItem{
				conv: &storage.Conversation{
					ID:         "conv-456",
					Title:      "Favorite Chat",
					IsFavorite: true,
				},
			},
			expected: "★ Favorite Chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.Title()
			if got != tt.expected {
				t.Errorf("Title() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSessionItem_Description(t *testing.T) {
	now := time.Now()
	twoHoursAgo := now.Add(-2 * time.Hour)

	tests := []struct {
		name     string
		item     sessionItem
		contains []string
	}{
		{
			name:     "new session item",
			item:     sessionItem{isNewItem: true},
			contains: []string{"Start a fresh conversation"},
		},
		{
			name: "conversation with details",
			item: sessionItem{
				conv: &storage.Conversation{
					ID:        "conv-123456789",
					Title:     "Test Chat",
					Model:     "claude-sonnet-4-5-20250929",
					UpdatedAt: twoHoursAgo,
				},
			},
			contains: []string{"Updated:", "Model:", "ID:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.Description()
			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("Description() = %q, should contain %q", got, substr)
				}
			}
		})
	}
}

func TestSessionItem_FilterValue(t *testing.T) {
	tests := []struct {
		name     string
		item     sessionItem
		expected string
	}{
		{
			name:     "new session item",
			item:     sessionItem{isNewItem: true},
			expected: "new session",
		},
		{
			name: "regular conversation",
			item: sessionItem{
				conv: &storage.Conversation{
					Title: "My Chat Title",
				},
			},
			expected: "My Chat Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.FilterValue()
			if got != tt.expected {
				t.Errorf("FilterValue() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewSessionPicker(t *testing.T) {
	tests := []struct {
		name              string
		conversations     []*storage.Conversation
		expectedItemCount int
	}{
		{
			name:              "empty conversation list",
			conversations:     []*storage.Conversation{},
			expectedItemCount: 1, // Just the "New Session" item
		},
		{
			name: "single conversation",
			conversations: []*storage.Conversation{
				{
					ID:    "conv-1",
					Title: "Test Chat",
					Model: "claude-sonnet-4",
				},
			},
			expectedItemCount: 2, // "New Session" + 1 conversation
		},
		{
			name: "multiple conversations",
			conversations: []*storage.Conversation{
				{
					ID:    "conv-1",
					Title: "Chat 1",
					Model: "claude-sonnet-4",
				},
				{
					ID:         "conv-2",
					Title:      "Chat 2",
					Model:      "claude-opus-4",
					IsFavorite: true,
				},
			},
			expectedItemCount: 3, // "New Session" + 2 conversations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picker := NewSessionPicker(tt.conversations)

			// Verify picker was created
			if picker.list.Items() == nil {
				t.Fatal("NewSessionPicker() returned picker with nil items")
			}

			// Verify item count
			itemCount := len(picker.list.Items())
			if itemCount != tt.expectedItemCount {
				t.Errorf("NewSessionPicker() created %d items, want %d", itemCount, tt.expectedItemCount)
			}

			// Verify first item is always "New Session"
			firstItem, ok := picker.list.Items()[0].(sessionItem)
			if !ok {
				t.Fatal("First item is not a sessionItem")
			}
			if !firstItem.isNewItem {
				t.Error("First item should be the 'New Session' item")
			}
		})
	}
}

func TestSessionPicker_Init(t *testing.T) {
	picker := NewSessionPicker([]*storage.Conversation{})
	cmd := picker.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestSessionPicker_Update_WindowSize(t *testing.T) {
	picker := NewSessionPicker([]*storage.Conversation{})

	msg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	updatedModel, _ := picker.Update(msg)
	updatedPicker, ok := updatedModel.(SessionPicker)
	if !ok {
		t.Fatal("Update() didn't return SessionPicker")
	}

	// Verify the model was updated (no panic)
	if updatedPicker.list.Width() != 80 {
		t.Errorf("List width = %d, want 80", updatedPicker.list.Width())
	}
}

func TestSessionPicker_Update_EnterKey(t *testing.T) {
	conversations := []*storage.Conversation{
		{
			ID:    "conv-123",
			Title: "Test Chat",
			Model: "claude-sonnet-4",
		},
	}

	picker := NewSessionPicker(conversations)

	// Move to second item (first conversation)
	picker.list.Select(1)

	msg := tea.KeyMsg{
		Type:  tea.KeyEnter,
		Runes: []rune{'\r'},
	}

	updatedModel, cmd := picker.Update(msg)
	updatedPicker, ok := updatedModel.(SessionPicker)
	if !ok {
		t.Fatal("Update() didn't return SessionPicker")
	}

	// Verify conversation was selected
	if updatedPicker.SelectedID() != "conv-123" {
		t.Errorf("SelectedID() = %q, want %q", updatedPicker.SelectedID(), "conv-123")
	}

	// Verify quitting
	if !updatedPicker.quitting {
		t.Error("Should be quitting after selection")
	}

	// Verify Quit command was returned
	if cmd == nil {
		t.Error("Update() should return tea.Quit command")
	}
}

func TestSessionPicker_Update_EnterOnNewSession(t *testing.T) {
	picker := NewSessionPicker([]*storage.Conversation{
		{ID: "conv-1", Title: "Test"},
	})

	// Select first item (New Session)
	picker.list.Select(0)

	msg := tea.KeyMsg{
		Type:  tea.KeyEnter,
		Runes: []rune{'\r'},
	}

	updatedModel, _ := picker.Update(msg)
	updatedPicker, ok := updatedModel.(SessionPicker)
	if !ok {
		t.Fatal("Update() didn't return SessionPicker")
	}

	// Verify new session flag is set
	if !updatedPicker.IsNewSession() {
		t.Error("IsNewSession() should be true when selecting 'New Session' item")
	}

	// Verify no conversation was selected
	if updatedPicker.SelectedID() != "" {
		t.Errorf("SelectedID() should be empty for new session, got %q", updatedPicker.SelectedID())
	}
}

func TestSessionPicker_Update_EscKey(t *testing.T) {
	picker := NewSessionPicker([]*storage.Conversation{})

	msg := tea.KeyMsg{
		Type: tea.KeyEsc,
	}

	updatedModel, cmd := picker.Update(msg)
	updatedPicker, ok := updatedModel.(SessionPicker)
	if !ok {
		t.Fatal("Update() didn't return SessionPicker")
	}

	// Verify new session mode
	if !updatedPicker.IsNewSession() {
		t.Error("Esc should trigger new session mode")
	}

	// Verify quitting
	if !updatedPicker.quitting {
		t.Error("Should be quitting after Esc")
	}

	// Verify Quit command
	if cmd == nil {
		t.Error("Update() should return tea.Quit command")
	}
}

func TestSessionPicker_View(t *testing.T) {
	picker := NewSessionPicker([]*storage.Conversation{})

	// Before quitting
	view := picker.View()
	if view == "" {
		t.Error("View() should return non-empty string before quitting")
	}

	// After quitting
	picker.quitting = true
	view = picker.View()
	if view != "" {
		t.Error("View() should return empty string after quitting")
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "just now",
			time:     now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			time:     now.Add(-1 * time.Minute),
			expected: "1 min ago",
		},
		{
			name:     "5 minutes ago",
			time:     now.Add(-5 * time.Minute),
			expected: "5 mins ago",
		},
		{
			name:     "1 hour ago",
			time:     now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			time:     now.Add(-3 * time.Hour),
			expected: "3 hours ago",
		},
		{
			name:     "1 day ago",
			time:     now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "3 days ago",
			time:     now.Add(-72 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "long ago",
			time:     now.Add(-30 * 24 * time.Hour),
			expected: "", // Will be in "Jan 2" format, we'll check it differently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimeAgo(tt.time)
			if tt.expected == "" {
				// For long ago dates, just verify format
				if len(got) < 3 {
					t.Errorf("formatTimeAgo() = %q, expected date format", got)
				}
			} else if got != tt.expected {
				t.Errorf("formatTimeAgo() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTruncateModel(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected string
	}{
		{
			name:     "short model name",
			model:    "gpt-4",
			expected: "gpt-4",
		},
		{
			name:     "long model name with date",
			model:    "claude-sonnet-4-5-20250929",
			expected: "claude-sonnet-4-5",
		},
		{
			name:     "model name with exactly 4 parts",
			model:    "claude-opus-4-turbo",
			expected: "claude-opus-4-turbo",
		},
		{
			name:     "model name with 3 parts",
			model:    "claude-haiku-3",
			expected: "claude-haiku-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateModel(tt.model)
			if got != tt.expected {
				t.Errorf("truncateModel(%q) = %q, want %q", tt.model, got, tt.expected)
			}
		})
	}
}

func TestTruncateID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "short ID",
			id:       "conv-123",
			expected: "conv-123",
		},
		{
			name:     "long UUID",
			id:       "550e8400-e29b-41d4-a716-446655440000",
			expected: "550e8400...",
		},
		{
			name:     "long conv ID",
			id:       "conv-1234567890",
			expected: "conv-123...",
		},
		{
			name:     "exactly 12 chars",
			id:       "conv-1234567",
			expected: "conv-1234567",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateID(tt.id)
			if got != tt.expected {
				t.Errorf("truncateID(%q) = %q, want %q", tt.id, got, tt.expected)
			}
		})
	}
}

func TestSessionPicker_EmptyList(t *testing.T) {
	// Edge case: empty conversation list should still work
	picker := NewSessionPicker([]*storage.Conversation{})

	// Should have one item (New Session)
	items := picker.list.Items()
	if len(items) != 1 {
		t.Errorf("Empty conversation list should have 1 item (New Session), got %d", len(items))
	}

	// Verify it's the new session item
	item, ok := items[0].(sessionItem)
	if !ok || !item.isNewItem {
		t.Error("Only item should be the 'New Session' item")
	}
}

func TestSessionPicker_FavoriteConversations(t *testing.T) {
	conversations := []*storage.Conversation{
		{
			ID:         "conv-1",
			Title:      "Favorite Chat",
			Model:      "claude-sonnet-4",
			IsFavorite: true,
		},
		{
			ID:         "conv-2",
			Title:      "Regular Chat",
			Model:      "claude-opus-4",
			IsFavorite: false,
		},
	}

	picker := NewSessionPicker(conversations)
	items := picker.list.Items()

	// Check favorite conversation (second item, first is "New Session")
	favoriteItem, ok := items[1].(sessionItem)
	if !ok {
		t.Fatal("Item should be sessionItem")
	}

	title := favoriteItem.Title()
	if !strings.HasPrefix(title, "★") {
		t.Errorf("Favorite conversation title should start with ★, got %q", title)
	}

	// Check regular conversation
	regularItem, ok := items[2].(sessionItem)
	if !ok {
		t.Fatal("Item should be sessionItem")
	}

	title = regularItem.Title()
	if strings.HasPrefix(title, "★") {
		t.Errorf("Regular conversation title should not start with ★, got %q", title)
	}
}
