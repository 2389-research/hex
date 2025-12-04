// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Tests for intro screen rendering and behavior
// ABOUTME: Verifies intro shows on startup and hides after first message
package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestIntroRendersOnStartup verifies intro screen shows when model is initialized
func TestIntroRendersOnStartup(t *testing.T) {
	m := NewModel("test-conv", "claude-sonnet-4-5-20250929", "dracula")
	m.Width = 80
	m.Height = 24
	m.Ready = true

	// Should show intro by default
	if !m.showIntro {
		t.Error("Expected showIntro to be true on startup")
	}

	// View should render intro when no messages
	view := m.View()
	// Note: lipgloss adds ANSI escape codes for styling, so we can't do exact string matches
	// We'll check for key content that should be visible even with styling
	if !strings.Contains(view, "|  _ \\") || !strings.Contains(view, "| |_)") {
		t.Errorf("Expected intro to contain Jeff ASCII logo, got: %s", view)
	}
	if !strings.Contains(view, "Productivity AI Agent") {
		t.Error("Expected intro to contain 'Productivity AI Agent' tagline")
	}
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Error("Expected intro to contain 'Keyboard Shortcuts' section")
	}
	if !strings.Contains(view, "Type a message to begin") {
		t.Error("Expected intro to contain 'Type a message to begin' prompt")
	}
}

// TestIntroHidesAfterFirstMessage verifies intro disappears when user sends first message
func TestIntroHidesAfterFirstMessage(t *testing.T) {
	m := NewModel("test-conv", "claude-sonnet-4-5-20250929", "dracula")
	m.Width = 80
	m.Height = 24
	m.Ready = true

	// Set up input
	m.Input.SetValue("Hello, world!")

	// Simulate Enter key to send message
	msg := tea.KeyMsg{
		Type:  tea.KeyEnter,
		Runes: nil,
	}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(*Model)

	// showIntro should now be false
	if m.showIntro {
		t.Error("Expected showIntro to be false after sending first message")
	}

	// View should no longer show intro
	view := m.View()
	if strings.Contains(view, "Type a message to begin") {
		t.Error("Expected intro to be hidden after first message")
	}
}

// TestIntroDoesNotShowWhenResumingConversation verifies intro doesn't show when messages exist
func TestIntroDoesNotShowWhenResumingConversation(t *testing.T) {
	m := NewModel("test-conv", "claude-sonnet-4-5-20250929", "dracula")
	m.Width = 80
	m.Height = 24
	m.Ready = true

	// Simulate resuming a conversation by adding existing messages
	m.AddMessage("user", "Previous message 1")
	m.AddMessage("assistant", "Previous response 1")
	m.AddMessage("user", "Previous message 2")

	// Even though showIntro is true, intro should not render because messages exist
	m.showIntro = true

	view := m.View()
	if strings.Contains(view, "Type a message to begin") {
		t.Error("Expected intro to NOT show when conversation has messages")
	}
	if strings.Contains(view, "Keyboard Shortcuts:") && len(m.Messages) > 0 {
		t.Error("Expected intro keyboard shortcuts to NOT show when resuming conversation")
	}
}

// TestIntroRendersFunctionalContent verifies intro contains accurate keyboard shortcuts
func TestIntroRendersFunctionalContent(t *testing.T) {
	m := NewModel("test-conv", "claude-sonnet-4-5-20250929", "dracula")
	m.Width = 80
	m.Height = 24
	m.Ready = true

	view := m.View()

	// Verify key shortcuts are present and accurate
	expectedShortcuts := []string{
		"ctrl+c",
		"ctrl+s",
		"ctrl+f",
		"ctrl+l",
		"ctrl+k",
		"?",
		"/",
		":",
		"j/k",
		"gg/G",
		"tab",
	}

	for _, shortcut := range expectedShortcuts {
		if !strings.Contains(view, shortcut) {
			t.Errorf("Expected intro to contain shortcut '%s'", shortcut)
		}
	}

	// Verify descriptions are present
	expectedDescriptions := []string{
		"Quit",
		"Save conversation",
		"Toggle favorites",
		"Clear screen",
		"Send message",
		"Toggle help",
	}

	for _, desc := range expectedDescriptions {
		if !strings.Contains(view, desc) {
			t.Errorf("Expected intro to contain description '%s'", desc)
		}
	}
}

// TestIntroRespectsTheme verifies intro uses theme styling
func TestIntroRespectsTheme(t *testing.T) {
	themes := []string{"dracula", "gruvbox", "nord"}

	for _, themeName := range themes {
		m := NewModel("test-conv", "claude-sonnet-4-5-20250929", themeName)
		m.Width = 80
		m.Height = 24
		m.Ready = true

		// Should not panic
		view := m.View()
		if view == "" {
			t.Errorf("Expected non-empty view for theme '%s'", themeName)
		}

		// Should contain intro content (check for ASCII art structure)
		if !strings.Contains(view, "┃") || !strings.Contains(view, "Productivity AI Agent") {
			t.Errorf("Expected intro content for theme '%s'", themeName)
		}
	}
}

// TestIntroASCIIArtRendersCorrectly verifies ASCII art is properly formatted
func TestIntroASCIIArtRendersCorrectly(t *testing.T) {
	m := NewModel("test-conv", "claude-sonnet-4-5-20250929", "dracula")
	m.Width = 80
	m.Height = 24
	m.Ready = true

	view := m.View()

	// Check for ASCII art structural elements
	if !strings.Contains(view, "┏━") {
		t.Error("Expected intro to contain top border of ASCII art")
	}
	if !strings.Contains(view, "┗━") {
		t.Error("Expected intro to contain bottom border of ASCII art")
	}
	if !strings.Contains(view, "┃") {
		t.Error("Expected intro to contain side borders of ASCII art")
	}
}

// TestIntroStateTransitions verifies correct state transitions
func TestIntroStateTransitions(t *testing.T) {
	m := NewModel("test-conv", "claude-sonnet-4-5-20250929", "dracula")
	m.Ready = true

	// Initial state: intro should be shown
	if !m.showIntro {
		t.Error("Expected showIntro=true initially")
	}

	// Add a message without going through Update
	m.showIntro = false
	m.AddMessage("user", "test")

	// Intro should stay hidden
	if m.showIntro {
		t.Error("Expected showIntro to remain false after manually hiding")
	}

	// Even if we try to force it back
	m.showIntro = true
	view := m.View()

	// View should not show intro because messages exist
	if strings.Contains(view, "Type a message to begin") {
		t.Error("Expected intro to NOT show when messages exist, even if showIntro=true")
	}
}

// TestIntroWithWindowSize verifies intro adapts to window size
func TestIntroWithWindowSize(t *testing.T) {
	m := NewModel("test-conv", "claude-sonnet-4-5-20250929", "dracula")
	m.Ready = true

	// Test with different window sizes
	windowSizes := []struct {
		width  int
		height int
	}{
		{80, 24},
		{120, 40},
		{60, 20},
	}

	for _, size := range windowSizes {
		m.Width = size.width
		m.Height = size.height
		m.showIntro = true
		m.Messages = []Message{} // Clear messages

		view := m.View()
		if view == "" {
			t.Errorf("Expected non-empty view for window size %dx%d", size.width, size.height)
		}
		// Check for ASCII art structure instead of plain text
		if !strings.Contains(view, "┃") || !strings.Contains(view, "Productivity AI Agent") {
			t.Errorf("Expected intro content for window size %dx%d", size.width, size.height)
		}
	}
}
