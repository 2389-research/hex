// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Intro screen rendering for first-run experience
// ABOUTME: Displays ASCII logo and keyboard shortcuts on startup
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderIntro returns the intro screen content
func (m *Model) RenderIntro() string {
	// ASCII art logo for "Pagen"
	logo := `
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  ____                          ┃
┃ |  _ \ __ _  __ _  ___ _ __   ┃
┃ | |_) / _' |/ _' |/ _ \ '_ \  ┃
┃ |  __/ (_| | (_| |  __/ | | | ┃
┃ |_|   \__,_|\__, |\___|_| |_| ┃
┃             |___/              ┃
┃                                ┃
┃  Productivity AI Agent         ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
`

	// Build keyboard shortcuts list with accurate shortcuts from the codebase
	shortcuts := []struct {
		key  string
		desc string
	}{
		{"ctrl+c", "Quit"},
		{"enter", "Send message"},
		{"ctrl+s", "Save conversation"},
		{"ctrl+f", "Toggle favorites"},
		{"ctrl+l", "Clear screen"},
		{"ctrl+k", "Clear conversation"},
		{"?", "Toggle help"},
		{"/", "Search messages"},
		{":", "Quick actions"},
		{"j/k", "Scroll up/down"},
		{"gg/G", "Jump to top/bottom"},
		{"tab", "Switch view"},
	}

	var shortcutsBuilder strings.Builder
	shortcutsBuilder.WriteString("\n\nKeyboard Shortcuts:\n\n")

	// Create styles for shortcuts
	keyStyle := lipgloss.NewStyle().
		Foreground(m.theme.Primary()).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.theme.Subtle())

	for _, s := range shortcuts {
		shortcutsBuilder.WriteString("  ")
		shortcutsBuilder.WriteString(keyStyle.Render(s.key))
		shortcutsBuilder.WriteString(" - ")
		shortcutsBuilder.WriteString(descStyle.Render(s.desc))
		shortcutsBuilder.WriteString("\n")
	}

	shortcutsBuilder.WriteString("\n")
	promptStyle := lipgloss.NewStyle().
		Foreground(m.theme.Secondary()).
		Italic(true)
	shortcutsBuilder.WriteString(promptStyle.Render("Type a message to begin..."))

	// Center the logo and render with theme title style
	logoStyle := lipgloss.NewStyle().
		Foreground(m.theme.Primary()).
		Bold(true).
		Align(lipgloss.Center)

	// Build final output
	var output strings.Builder
	output.WriteString(logoStyle.Render(logo))
	output.WriteString(shortcutsBuilder.String())

	// Center the entire intro on screen if we have width
	if m.Width > 0 {
		containerStyle := lipgloss.NewStyle().
			Width(m.Width).
			Align(lipgloss.Center).
			PaddingTop(2)
		return containerStyle.Render(output.String())
	}

	return output.String()
}
