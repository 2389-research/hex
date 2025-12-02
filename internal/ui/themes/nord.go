// Package themes provides theming support for the TUI.
// ABOUTME: Nord theme implementation with cool nordic aesthetic
// ABOUTME: Based on official Nord color scheme
package themes

import "github.com/charmbracelet/lipgloss"

// Nord is the Nord theme.
type Nord struct{}

// NewNord creates a new Nord theme instance.
func NewNord() Theme {
	return &Nord{}
}

// Name returns the theme's display name.
func (n *Nord) Name() string { return "Nord" }

// Background returns the background color.
func (n *Nord) Background() lipgloss.Color { return lipgloss.Color("#2e3440") }

// Foreground returns the foreground text color.
func (n *Nord) Foreground() lipgloss.Color { return lipgloss.Color("#eceff4") }

// Primary returns the primary accent color (cyan).
func (n *Nord) Primary() lipgloss.Color { return lipgloss.Color("#88c0d0") }

// Secondary returns the secondary accent color (blue).
func (n *Nord) Secondary() lipgloss.Color { return lipgloss.Color("#81a1c1") }

// Success returns the success state color (green).
func (n *Nord) Success() lipgloss.Color { return lipgloss.Color("#a3be8c") }

// Warning returns the warning state color (yellow).
func (n *Nord) Warning() lipgloss.Color { return lipgloss.Color("#ebcb8b") }

// Error returns the error state color (red).
func (n *Nord) Error() lipgloss.Color { return lipgloss.Color("#bf616a") }

// Border returns the border color.
func (n *Nord) Border() lipgloss.Color { return lipgloss.Color("#4c566a") }

// BorderFocus returns the focused border color.
func (n *Nord) BorderFocus() lipgloss.Color { return lipgloss.Color("#88c0d0") }

// Subtle returns the subtle/muted text color.
func (n *Nord) Subtle() lipgloss.Color { return lipgloss.Color("#4c566a") }

// TitleGradient returns the gradient colors for titles (cyan to blue).
func (n *Nord) TitleGradient() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color("#88c0d0"), // Cyan
		lipgloss.Color("#81a1c1"), // Blue
	}
}
