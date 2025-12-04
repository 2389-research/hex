// Package components provides reusable UI components for the TUI.
// ABOUTME: Error visualization component with themed styling
// ABOUTME: Displays errors with appropriate color and formatting
package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jeff/internal/ui/layout"
	"github.com/harper/jeff/internal/ui/themes"
)

// ErrorDisplay shows formatted errors
type ErrorDisplay struct {
	theme   themes.Theme
	title   string
	message string
	details string
	width   int
	height  int
}

// NewErrorDisplay creates a new error display
func NewErrorDisplay(theme themes.Theme, title, message, details string) *ErrorDisplay {
	return &ErrorDisplay{
		theme:   theme,
		title:   title,
		message: message,
		details: details,
	}
}

// Init implements tea.Model
func (e *ErrorDisplay) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (e *ErrorDisplay) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return e, nil
}

// View implements tea.Model and renders the error display
func (e *ErrorDisplay) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(e.theme.Error()).
		Bold(true).
		MarginBottom(1)

	messageStyle := lipgloss.NewStyle().
		Foreground(e.theme.Foreground())

	detailsStyle := lipgloss.NewStyle().
		Foreground(e.theme.Subtle()).
		Italic(true)

	var lines []string
	lines = append(lines, titleStyle.Render("❌ "+e.title))

	if e.message != "" {
		lines = append(lines, "")
		lines = append(lines, messageStyle.Render(e.message))
	}

	if e.details != "" {
		lines = append(lines, "")
		lines = append(lines, detailsStyle.Render("Details: "+e.details))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Create spacing with padding and margin
	spacing := layout.NewPadding(1, 2, 1, 2)
	spacing.MarginTop = 1

	box := layout.NewBorderStyle(e.theme).
		WithColor(e.theme.Error()).
		WithSpacing(spacing)

	// Apply size constraints if set
	if e.width > 0 || e.height > 0 {
		box = box.WithSize(e.width, e.height)
	}

	return box.Render(content)
}

// SetSize implements the Sizeable interface
func (e *ErrorDisplay) SetSize(width, height int) tea.Cmd {
	e.width = width
	e.height = height
	return nil
}

// GetSize implements the Sizeable interface
func (e *ErrorDisplay) GetSize() (int, int) {
	return e.width, e.height
}
