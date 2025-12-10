// Package theme provides Neo-Terminal color palette and lipgloss styles for the hex TUI.
// ABOUTME: Neo-Terminal theme - sophisticated, information-rich terminal design
// ABOUTME: Celebrates modern terminal capabilities with elegant typography and rich visual hierarchy
package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Neo-Terminal color palette - Information-rich sophistication
// Inspired by Swiss typography meets cyberdeck aesthetics
const (
	// Primary Palette
	DeepInk     = "#1a1b26" // Background - rich black-blue
	SoftPaper   = "#c0caf5" // Primary text - cool white
	AccentCoral = "#ff9e64" // User messages - warm
	AccentSage  = "#9ece6a" // Assistant - natural green
	AccentSky   = "#7aa2f7" // Tools/system - cool blue

	// Secondary Palette
	DimInk       = "#565f89" // Borders, secondary text
	Ghost        = "#414868" // Subtle elements
	WarningAmber = "#e0af68" // Warnings
	ErrorRuby    = "#f7768e" // Errors
	SuccessJade  = "#73daca" // Success states
)

// NewNeoTerminalTheme creates and returns a new Neo-Terminal theme with all styles initialized
func NewNeoTerminalTheme() *Theme {
	t := &Theme{}

	// Initialize color references
	t.Colors.Background = lipgloss.Color(DeepInk)
	t.Colors.CurrentLine = lipgloss.Color(Ghost)
	t.Colors.Foreground = lipgloss.Color(SoftPaper)
	t.Colors.Comment = lipgloss.Color(DimInk)
	t.Colors.Cyan = lipgloss.Color(AccentSky)
	t.Colors.Green = lipgloss.Color(AccentSage)
	t.Colors.Orange = lipgloss.Color(AccentCoral)
	t.Colors.Pink = lipgloss.Color(AccentCoral) // Reuse coral for consistency
	t.Colors.Purple = lipgloss.Color(AccentSky) // Reuse sky for consistency
	t.Colors.Red = lipgloss.Color(ErrorRuby)
	t.Colors.Yellow = lipgloss.Color(WarningAmber)

	// Text styles - refined hierarchy
	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Colors.Foreground)

	t.Subtitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Colors.Comment)

	t.Body = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground)

	t.Muted = lipgloss.NewStyle().
		Foreground(t.Colors.Comment)

	t.Emphasized = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Colors.Foreground)

	// Status styles
	t.Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color(SuccessJade)).
		Bold(true)

	t.Error = lipgloss.NewStyle().
		Foreground(t.Colors.Red).
		Bold(true)

	t.Warning = lipgloss.NewStyle().
		Foreground(t.Colors.Yellow).
		Bold(true)

	t.Info = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Bold(true)

	// Interactive element styles - rounded sophistication
	t.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Comment).
		Padding(0, 1)

	t.BorderFocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Cyan).
		Padding(0, 1)

	t.Input = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Comment).
		Padding(0, 1)

	t.InputFocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Cyan).
		Padding(0, 1)

	t.Button = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground).
		Background(t.Colors.CurrentLine).
		Padding(0, 2).
		Bold(true)

	t.ButtonActive = lipgloss.NewStyle().
		Foreground(t.Colors.Background).
		Background(t.Colors.Cyan).
		Padding(0, 2).
		Bold(true)

	// Specialized component styles
	t.StatusBar = lipgloss.NewStyle().
		Foreground(t.Colors.Comment).
		Background(t.Colors.Background).
		Bold(false)

	t.ViewMode = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground).
		Bold(true)

	t.SearchPrompt = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Background(t.Colors.CurrentLine).
		Padding(0, 1).
		Bold(true)

	t.TokenCounter = lipgloss.NewStyle().
		Foreground(t.Colors.Comment)

	// Tool-related styles - nested box hierarchy
	t.ToolApproval = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Colors.Orange).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Orange).
		Padding(1, 2)

	t.ToolExecuting = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Bold(true)

	t.ToolSuccess = lipgloss.NewStyle().
		Foreground(lipgloss.Color(SuccessJade)).
		Bold(true)

	t.ToolError = lipgloss.NewStyle().
		Foreground(t.Colors.Red).
		Bold(true)

	// Suggestion styles
	t.SuggestionBox = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Background(t.Colors.CurrentLine).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Cyan).
		Padding(0, 1)

	t.SuggestionTitle = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Bold(true)

	t.SuggestionReason = lipgloss.NewStyle().
		Foreground(t.Colors.Comment).
		Italic(true)

	t.SuggestionHint = lipgloss.NewStyle().
		Foreground(t.Colors.Comment)

	// List and selection styles
	t.ListItem = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground)

	t.ListItemSelected = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Bold(true).
		Background(t.Colors.CurrentLine)

	t.ListItemActive = lipgloss.NewStyle().
		Foreground(t.Colors.Background).
		Background(t.Colors.Cyan).
		Bold(true)

	// Autocomplete dropdown styles - high contrast for readability
	// Use dark background with bright foreground for maximum legibility
	t.AutocompleteDropdown = lipgloss.NewStyle().
		Background(t.Colors.Background). // Dark background (DeepInk)
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Cyan).
		Padding(0, 1)

	t.AutocompleteItem = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground). // Bright text (SoftPaper)
		Background(t.Colors.Background)  // Dark background (DeepInk)

	t.AutocompleteSelected = lipgloss.NewStyle().
		Foreground(t.Colors.Background). // Dark text (DeepInk)
		Background(t.Colors.Cyan).       // Bright background (AccentSky)
		Bold(true)

	t.AutocompleteHelp = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground). // Brighter than Comment for visibility
		Background(t.Colors.Background).
		Italic(true)

	// Help and modal styles
	t.HelpPanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Cyan).
		Padding(1, 2)

	t.HelpKey = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Bold(true)

	t.HelpDesc = lipgloss.NewStyle().
		Foreground(t.Colors.Comment)

	t.Modal = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Cyan).
		Padding(1, 2)

	t.ModalTitle = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Bold(true)

	// Code and syntax styles
	t.Code = lipgloss.NewStyle().
		Foreground(t.Colors.Green)

	t.CodeBlock = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground).
		Background(t.Colors.CurrentLine).
		Padding(0, 1)

	t.Keyword = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Bold(true)

	t.String = lipgloss.NewStyle().
		Foreground(t.Colors.Green)

	t.Number = lipgloss.NewStyle().
		Foreground(t.Colors.Orange)

	// Link and reference styles
	t.Link = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Underline(true)

	t.LinkHover = lipgloss.NewStyle().
		Foreground(t.Colors.Orange).
		Underline(true).
		Bold(true)

	// Message styles - role-specific
	t.UserMessage = lipgloss.NewStyle().
		Foreground(t.Colors.Orange)

	return t
}

// NeoTerminalTheme returns a pre-configured Neo-Terminal theme instance
// This is the recommended way to get a theme for use in the application
func NeoTerminalTheme() *Theme {
	return NewNeoTerminalTheme()
}
