// Package themes provides theming support for the TUI.
// ABOUTME: Dracula theme implementation with purple/pink accent colors
// ABOUTME: Based on official Dracula color scheme (draculatheme.com)
package themes

import "github.com/charmbracelet/lipgloss"

// Dracula is the Dracula theme.
type Dracula struct{}

// NewDracula creates a new Dracula theme instance.
func NewDracula() Theme {
	return &Dracula{}
}

// Name returns the theme's display name.
func (d *Dracula) Name() string { return "Dracula" }

// Background returns the background color.
func (d *Dracula) Background() lipgloss.Color { return lipgloss.Color("#282a36") }

// Foreground returns the foreground text color.
func (d *Dracula) Foreground() lipgloss.Color { return lipgloss.Color("#f8f8f2") }

// Primary returns the primary accent color (purple).
func (d *Dracula) Primary() lipgloss.Color { return lipgloss.Color("#bd93f9") }

// Secondary returns the secondary accent color (pink).
func (d *Dracula) Secondary() lipgloss.Color { return lipgloss.Color("#ff79c6") }

// Success returns the success state color (green).
func (d *Dracula) Success() lipgloss.Color { return lipgloss.Color("#50fa7b") }

// Warning returns the warning state color (yellow).
func (d *Dracula) Warning() lipgloss.Color { return lipgloss.Color("#f1fa8c") }

// Error returns the error state color (red).
func (d *Dracula) Error() lipgloss.Color { return lipgloss.Color("#ff5555") }

// Border returns the border color.
func (d *Dracula) Border() lipgloss.Color { return lipgloss.Color("#6272a4") }

// BorderFocus returns the focused border color.
func (d *Dracula) BorderFocus() lipgloss.Color { return lipgloss.Color("#bd93f9") }

// Subtle returns the subtle/muted text color.
func (d *Dracula) Subtle() lipgloss.Color { return lipgloss.Color("#6272a4") }

// TitleGradient returns the gradient colors for titles (purple to pink).
func (d *Dracula) TitleGradient() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color("#bd93f9"), // Purple
		lipgloss.Color("#ff79c6"), // Pink
	}
}
