// Package components provides reusable UI components for the TUI.
// ABOUTME: Help overlay component showing keyboard shortcuts
// ABOUTME: Displays themed help information with keybindings
package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jeff/internal/ui/layout"
	"github.com/harper/jeff/internal/ui/themes"
)

// HelpOverlay displays keyboard shortcuts
type HelpOverlay struct {
	theme  themes.Theme
	width  int
	height int

	// Phase 1 Task 3: Content caching
	cachedContent string // Cached rendered help text
	contentDirty  bool   // Flag to invalidate cache on size change
}

// NewHelpOverlay creates a new help overlay
func NewHelpOverlay(theme themes.Theme) *HelpOverlay {
	return &HelpOverlay{theme: theme}
}

// Init implements tea.Model
func (h *HelpOverlay) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (h *HelpOverlay) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return h, nil
}

// View implements tea.Model and renders the help overlay
// Phase 1 Task 3: Uses caching to avoid expensive re-renders
func (h *HelpOverlay) View() string {
	// Check cache first (if not dirty and cached content exists)
	if !h.contentDirty && h.cachedContent != "" {
		return h.cachedContent
	}

	// Generate help text (expensive operation)
	content := h.generateHelpText()

	// Cache the result
	h.cachedContent = content
	h.contentDirty = false

	return content
}

// generateHelpText creates the help overlay content
func (h *HelpOverlay) generateHelpText() string {
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
		{":", "Quick actions"},
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

	box := layout.NewBorderStyle(h.theme).
		WithFocus(true).
		WithSpacing(layout.NewPadding(1, 2, 1, 2))

	// Apply size constraints if set
	if h.width > 0 || h.height > 0 {
		box = box.WithSize(h.width, h.height)
	}

	return box.Render(content)
}

// SetSize implements the Sizeable interface
// Phase 1 Task 3: Invalidates cache when size changes
func (h *HelpOverlay) SetSize(width, height int) tea.Cmd {
	// Invalidate cache if size changed
	if width != h.width || height != h.height {
		h.contentDirty = true
	}
	h.width = width
	h.height = height
	return nil
}

// GetSize implements the Sizeable interface
func (h *HelpOverlay) GetSize() (int, int) {
	return h.width, h.height
}
