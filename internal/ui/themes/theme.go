// Package themes provides theming support for the TUI.
// ABOUTME: Theme system with interface and color definitions
// ABOUTME: Provides base types for Dracula, Gruvbox, and Nord themes
package themes

import "github.com/charmbracelet/lipgloss"

// Theme defines the color scheme and styling for the TUI.
type Theme interface {
	// Name returns the theme's display name
	Name() string

	// Base colors
	Background() lipgloss.Color
	Foreground() lipgloss.Color

	// Semantic colors
	Primary() lipgloss.Color   // Main accent color
	Secondary() lipgloss.Color // Secondary accent color
	Success() lipgloss.Color   // Green for completed/success states
	Warning() lipgloss.Color   // Yellow for pending/warning states
	Error() lipgloss.Color     // Red for errors

	// UI element colors
	Border() lipgloss.Color
	BorderFocus() lipgloss.Color
	Subtle() lipgloss.Color // Muted text

	// Gradients (returns slice of colors for interpolation)
	TitleGradient() []lipgloss.Color
}

// GetTheme returns a theme by name.
// Returns Dracula theme if name is unknown.
func GetTheme(name string) Theme {
	switch name {
	case "dracula":
		return NewDracula()
	case "gruvbox":
		return NewGruvbox()
	case "nord":
		return NewNord()
	default:
		return NewDracula() // Default to Dracula
	}
}

// AvailableThemes returns the list of available theme names.
func AvailableThemes() []string {
	return []string{"dracula", "gruvbox", "nord"}
}
