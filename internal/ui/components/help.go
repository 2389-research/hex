// Package components provides reusable UI components for the TUI.
// ABOUTME: Help overlay component showing keyboard shortcuts
// ABOUTME: Displays themed help information with keybindings
package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/pagent/internal/ui/layout"
	"github.com/harper/pagent/internal/ui/themes"
)

// HelpOverlay displays keyboard shortcuts
type HelpOverlay struct {
	theme themes.Theme
}

// NewHelpOverlay creates a new help overlay
func NewHelpOverlay(theme themes.Theme) *HelpOverlay {
	return &HelpOverlay{theme: theme}
}

// View renders the help overlay
func (h *HelpOverlay) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(h.theme.Primary()).
		Bold(true).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(h.theme.Secondary()).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(h.theme.Foreground())

	title := titleStyle.Render("Keyboard Shortcuts")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"ctrl+c", "Quit"},
		{"enter", "Send message"},
		{"ctrl+p", "Quick actions"},
		{"ctrl+l", "Clear screen"},
		{"esc", "Cancel/close"},
		{"↑↓", "Navigate tables"},
		{"?", "Toggle help"},
	}

	// Pre-allocate slice with capacity for title, blank line, and shortcuts
	lines := make([]string, 0, len(shortcuts)+2)
	lines = append(lines, title)
	lines = append(lines, "")

	for _, s := range shortcuts {
		key := keyStyle.Render(s.key)
		desc := descStyle.Render(s.desc)
		line := lipgloss.JoinHorizontal(lipgloss.Left, key, "  ", desc)
		lines = append(lines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return layout.NewBorderStyle(h.theme).
		WithFocus(true).
		WithSpacing(layout.NewPadding(1, 2, 1, 2)).
		Render(content)
}
