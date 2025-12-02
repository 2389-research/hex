// Package themes provides theming support for the TUI.
// ABOUTME: Gruvbox Dark theme implementation with warm earthy tones
// ABOUTME: Based on official Gruvbox color scheme
package themes

import "github.com/charmbracelet/lipgloss"

// Gruvbox is the Gruvbox Dark theme.
type Gruvbox struct{}

// NewGruvbox creates a new Gruvbox theme instance.
func NewGruvbox() Theme {
	return &Gruvbox{}
}

// Name returns the theme's display name.
func (g *Gruvbox) Name() string { return "Gruvbox Dark" }

// Background returns the background color.
func (g *Gruvbox) Background() lipgloss.Color { return lipgloss.Color("#282828") }

// Foreground returns the foreground text color.
func (g *Gruvbox) Foreground() lipgloss.Color { return lipgloss.Color("#ebdbb2") }

// Primary returns the primary accent color (orange).
func (g *Gruvbox) Primary() lipgloss.Color { return lipgloss.Color("#d79921") }

// Secondary returns the secondary accent color (purple).
func (g *Gruvbox) Secondary() lipgloss.Color { return lipgloss.Color("#b16286") }

// Success returns the success state color (green).
func (g *Gruvbox) Success() lipgloss.Color { return lipgloss.Color("#b8bb26") }

// Warning returns the warning state color (yellow).
func (g *Gruvbox) Warning() lipgloss.Color { return lipgloss.Color("#fabd2f") }

// Error returns the error state color (red).
func (g *Gruvbox) Error() lipgloss.Color { return lipgloss.Color("#fb4934") }

// Border returns the border color.
func (g *Gruvbox) Border() lipgloss.Color { return lipgloss.Color("#504945") }

// BorderFocus returns the focused border color.
func (g *Gruvbox) BorderFocus() lipgloss.Color { return lipgloss.Color("#d79921") }

// Subtle returns the subtle/muted text color.
func (g *Gruvbox) Subtle() lipgloss.Color { return lipgloss.Color("#928374") }

// TitleGradient returns the gradient colors for titles (orange to red).
func (g *Gruvbox) TitleGradient() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color("#d79921"), // Orange
		lipgloss.Color("#fb4934"), // Red
	}
}
