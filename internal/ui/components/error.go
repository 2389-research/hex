// Package components provides reusable UI components for the TUI.
// ABOUTME: Error visualization component with themed styling
// ABOUTME: Displays errors with appropriate color and formatting
package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/pagent/internal/ui/themes"
)

// ErrorDisplay shows formatted errors
type ErrorDisplay struct {
	theme   themes.Theme
	title   string
	message string
	details string
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

// View renders the error display
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

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(e.theme.Error()).
		Padding(1, 2).
		MarginTop(1)

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
	return borderStyle.Render(content)
}
