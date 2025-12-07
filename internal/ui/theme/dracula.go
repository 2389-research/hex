// Package theme provides Dracula color palette and lipgloss styles for the hex TUI.
// ABOUTME: Dracula theme color palette and lipgloss styles for the hex TUI
// ABOUTME: Provides consistent color scheme throughout the application
package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Dracula color palette constants
// Official Dracula theme specification from https://draculatheme.com/contribute
const (
	// Background colors
	Background  = "#282a36"
	CurrentLine = "#44475a"
	Selection   = "#44475a"

	// Foreground colors
	Foreground = "#f8f8f2"
	Comment    = "#6272a4"

	// Accent colors
	Cyan   = "#8be9fd"
	Green  = "#50fa7b"
	Orange = "#ffb86c"
	Pink   = "#ff79c6"
	Purple = "#bd93f9"
	Red    = "#ff5555"
	Yellow = "#f1fa8c"
)

// Theme provides a complete set of lipgloss styles for the UI
type Theme struct {
	// Color references (for direct use when needed)
	Colors struct {
		Background  lipgloss.Color
		CurrentLine lipgloss.Color
		Foreground  lipgloss.Color
		Comment     lipgloss.Color
		Cyan        lipgloss.Color
		Green       lipgloss.Color
		Orange      lipgloss.Color
		Pink        lipgloss.Color
		Purple      lipgloss.Color
		Red         lipgloss.Color
		Yellow      lipgloss.Color
	}

	// Text styles
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Body       lipgloss.Style
	Muted      lipgloss.Style
	Emphasized lipgloss.Style

	// Status styles
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style

	// Interactive element styles
	Border        lipgloss.Style
	BorderFocused lipgloss.Style
	Input         lipgloss.Style
	InputFocused  lipgloss.Style
	Button        lipgloss.Style
	ButtonActive  lipgloss.Style

	// Specialized component styles
	StatusBar    lipgloss.Style
	ViewMode     lipgloss.Style
	SearchPrompt lipgloss.Style
	TokenCounter lipgloss.Style

	// Tool-related styles
	ToolApproval  lipgloss.Style
	ToolExecuting lipgloss.Style
	ToolSuccess   lipgloss.Style
	ToolError     lipgloss.Style
	ToolResult    lipgloss.Style
	ToolCall      lipgloss.Style

	// Suggestion styles
	SuggestionBox    lipgloss.Style
	SuggestionTitle  lipgloss.Style
	SuggestionReason lipgloss.Style
	SuggestionHint   lipgloss.Style

	// List and selection styles
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemActive   lipgloss.Style

	// Help and modal styles
	HelpPanel  lipgloss.Style
	HelpKey    lipgloss.Style
	HelpDesc   lipgloss.Style
	Modal      lipgloss.Style
	ModalTitle lipgloss.Style

	// Code and syntax styles
	Code      lipgloss.Style
	CodeBlock lipgloss.Style
	Keyword   lipgloss.Style
	String    lipgloss.Style
	Number    lipgloss.Style

	// Link and reference styles
	Link      lipgloss.Style
	LinkHover lipgloss.Style

	// Message styles
	UserMessage lipgloss.Style
}

// NewDraculaTheme creates and returns a new Dracula theme with all styles initialized
func NewDraculaTheme() *Theme {
	t := &Theme{}

	// Initialize color references
	t.Colors.Background = lipgloss.Color(Background)
	t.Colors.CurrentLine = lipgloss.Color(CurrentLine)
	t.Colors.Foreground = lipgloss.Color(Foreground)
	t.Colors.Comment = lipgloss.Color(Comment)
	t.Colors.Cyan = lipgloss.Color(Cyan)
	t.Colors.Green = lipgloss.Color(Green)
	t.Colors.Orange = lipgloss.Color(Orange)
	t.Colors.Pink = lipgloss.Color(Pink)
	t.Colors.Purple = lipgloss.Color(Purple)
	t.Colors.Red = lipgloss.Color(Red)
	t.Colors.Yellow = lipgloss.Color(Yellow)

	// Text styles
	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Colors.Purple)

	t.Subtitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Colors.Pink)

	t.Body = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground)

	t.Muted = lipgloss.NewStyle().
		Foreground(t.Colors.Comment)

	t.Emphasized = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Colors.Cyan)

	// Status styles
	t.Success = lipgloss.NewStyle().
		Foreground(t.Colors.Green).
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

	// Interactive element styles
	t.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Comment)

	t.BorderFocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Purple)

	t.Input = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Comment).
		Padding(0, 1)

	t.InputFocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Purple).
		Padding(0, 1)

	t.Button = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground).
		Background(t.Colors.CurrentLine).
		Padding(0, 2).
		Bold(true)

	t.ButtonActive = lipgloss.NewStyle().
		Foreground(t.Colors.Background).
		Background(t.Colors.Purple).
		Padding(0, 2).
		Bold(true)

	// Specialized component styles
	t.StatusBar = lipgloss.NewStyle().
		Foreground(t.Colors.Comment).
		Background(t.Colors.Background)

	t.ViewMode = lipgloss.NewStyle().
		Foreground(t.Colors.Purple).
		Bold(true)

	t.SearchPrompt = lipgloss.NewStyle().
		Foreground(t.Colors.Yellow).
		Background(t.Colors.CurrentLine).
		Padding(0, 1)

	t.TokenCounter = lipgloss.NewStyle().
		Foreground(t.Colors.Comment)

	// Tool-related styles
	t.ToolApproval = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Colors.Orange).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Orange).
		Padding(1, 2)

	t.ToolExecuting = lipgloss.NewStyle().
		Foreground(t.Colors.Yellow).
		Bold(true)

	t.ToolSuccess = lipgloss.NewStyle().
		Foreground(t.Colors.Green).
		Bold(true)

	t.ToolError = lipgloss.NewStyle().
		Foreground(t.Colors.Red).
		Bold(true)

	t.ToolResult = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Bold(true)

	t.ToolCall = lipgloss.NewStyle().
		Foreground(t.Colors.Purple).
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
		Foreground(t.Colors.Purple).
		Bold(true).
		Background(t.Colors.CurrentLine)

	t.ListItemActive = lipgloss.NewStyle().
		Foreground(t.Colors.Background).
		Background(t.Colors.Purple).
		Bold(true)

	// Help and modal styles
	t.HelpPanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Purple).
		Padding(1, 2)

	t.HelpKey = lipgloss.NewStyle().
		Foreground(t.Colors.Pink).
		Bold(true)

	t.HelpDesc = lipgloss.NewStyle().
		Foreground(t.Colors.Comment)

	t.Modal = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Purple).
		Padding(1, 2)

	t.ModalTitle = lipgloss.NewStyle().
		Foreground(t.Colors.Purple).
		Bold(true)

	// Code and syntax styles
	t.Code = lipgloss.NewStyle().
		Foreground(t.Colors.Green)

	t.CodeBlock = lipgloss.NewStyle().
		Foreground(t.Colors.Foreground).
		Background(t.Colors.CurrentLine).
		Padding(0, 1)

	t.Keyword = lipgloss.NewStyle().
		Foreground(t.Colors.Pink).
		Bold(true)

	t.String = lipgloss.NewStyle().
		Foreground(t.Colors.Yellow)

	t.Number = lipgloss.NewStyle().
		Foreground(t.Colors.Purple)

	// Link and reference styles
	t.Link = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan).
		Underline(true)

	t.LinkHover = lipgloss.NewStyle().
		Foreground(t.Colors.Pink).
		Underline(true).
		Bold(true)

	// Message styles
	t.UserMessage = lipgloss.NewStyle().
		Foreground(t.Colors.Cyan)

	return t
}

// DraculaTheme returns a pre-configured Dracula theme instance
// This is the recommended way to get a theme for use in the application
func DraculaTheme() *Theme {
	return NewDraculaTheme()
}
