// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Status bar component for displaying connection, token, and mode information
// ABOUTME: Provides comprehensive bottom status bar with color-coded indicators
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/hex/internal/ui/theme"
)

// ConnectionStatus represents API connection state
type ConnectionStatus int

const (
	// ConnectionDisconnected indicates no active connection
	ConnectionDisconnected ConnectionStatus = iota
	// ConnectionConnected indicates an active connection
	ConnectionConnected
	// ConnectionStreaming indicates an active streaming connection
	ConnectionStreaming
	// ConnectionError indicates a connection error
	ConnectionError
)

// StatusBar manages the bottom status bar display
type StatusBar struct {
	model         string
	tokensInput   int
	tokensOutput  int
	tokensTotal   int
	contextSize   int
	connection    ConnectionStatus
	currentMode   string
	width         int
	helpVisible   bool
	customMessage string
	theme         *theme.Theme
}

// NewStatusBar creates a new status bar
func NewStatusBar(model string, width int) *StatusBar {
	return &StatusBar{
		model:       model,
		width:       width,
		connection:  ConnectionDisconnected,
		currentMode: "chat",
		contextSize: 200000, // Default context size
		theme:       theme.DraculaTheme(),
	}
}

// SetWidth updates the status bar width
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetModel updates the model name
func (s *StatusBar) SetModel(model string) {
	s.model = model
}

// UpdateTokens updates token counters
func (s *StatusBar) UpdateTokens(input, output int) {
	s.tokensInput += input
	s.tokensOutput += output
	s.tokensTotal = s.tokensInput + s.tokensOutput
}

// SetTokens sets token counters (absolute values)
func (s *StatusBar) SetTokens(input, output int) {
	s.tokensInput = input
	s.tokensOutput = output
	s.tokensTotal = input + output
}

// SetContextSize sets the context window size
func (s *StatusBar) SetContextSize(size int) {
	s.contextSize = size
}

// SetConnectionStatus updates the connection status
func (s *StatusBar) SetConnectionStatus(status ConnectionStatus) {
	s.connection = status
}

// SetMode updates the current mode
func (s *StatusBar) SetMode(mode string) {
	s.currentMode = mode
}

// SetHelpVisible toggles help visibility
func (s *StatusBar) SetHelpVisible(visible bool) {
	s.helpVisible = visible
}

// SetCustomMessage sets a temporary custom message
func (s *StatusBar) SetCustomMessage(msg string) {
	s.customMessage = msg
}

// ClearCustomMessage clears the custom message
func (s *StatusBar) ClearCustomMessage() {
	s.customMessage = ""
}

// View renders the status bar
func (s *StatusBar) View() string {
	if s.width < 40 {
		// Too narrow, show minimal info
		return s.theme.StatusBar.Render("Hex")
	}

	var parts []string

	// Left section: Model name
	parts = append(parts, s.theme.Subtitle.Render(s.model))

	// Connection indicator
	parts = append(parts, s.renderConnection())

	// Token usage
	if s.tokensTotal > 0 {
		parts = append(parts, s.renderTokenUsage())
	}

	// Context size indicator (if tokens are getting high)
	if s.tokensTotal > s.contextSize/2 {
		parts = append(parts, s.renderContextIndicator())
	}

	// Current mode
	parts = append(parts, s.theme.ViewMode.Render("["+s.currentMode+"]"))

	// Custom message (if any)
	if s.customMessage != "" {
		parts = append(parts, s.theme.Warning.Render(s.customMessage))
	}

	// Combine left parts
	leftSection := strings.Join(parts, " ")

	// Right section: Help text
	rightSection := s.renderHelp()

	// Calculate spacing
	usedWidth := lipgloss.Width(leftSection) + lipgloss.Width(rightSection)
	spacing := s.width - usedWidth - 4 // Account for padding
	if spacing < 1 {
		spacing = 1
	}

	// Combine sections
	bar := leftSection + strings.Repeat(" ", spacing) + rightSection

	return s.theme.StatusBar.Width(s.width).Render(bar)
}

// renderConnection renders the connection status indicator
func (s *StatusBar) renderConnection() string {
	switch s.connection {
	case ConnectionConnected:
		return s.theme.Success.Render("●")
	case ConnectionStreaming:
		return s.theme.Info.Render("◉")
	case ConnectionError:
		return s.theme.Error.Render("◉")
	case ConnectionDisconnected:
		return s.theme.Muted.Render("○")
	default:
		return s.theme.Muted.Render("○")
	}
}

// renderTokenUsage renders token usage information
func (s *StatusBar) renderTokenUsage() string {
	// Highlight if usage is getting high
	usagePercent := float64(s.tokensTotal) / float64(s.contextSize) * 100

	tokensText := fmt.Sprintf("%dk↓ %dk↑", s.tokensInput/1000, s.tokensOutput/1000)

	if usagePercent > 80 {
		return s.theme.Warning.Render(tokensText)
	}
	return s.theme.TokenCounter.Render(tokensText)
}

// renderContextIndicator renders context size indicator
func (s *StatusBar) renderContextIndicator() string {
	usagePercent := float64(s.tokensTotal) / float64(s.contextSize) * 100
	bars := int(usagePercent / 10)
	if bars > 10 {
		bars = 10
	}

	indicator := "["
	for i := 0; i < 10; i++ {
		if i < bars {
			if usagePercent > 80 {
				indicator += s.theme.Warning.Render("█")
			} else {
				indicator += s.theme.Info.Render("█")
			}
		} else {
			indicator += "░"
		}
	}
	indicator += "]"

	return indicator
}

// renderHelp renders the help text
func (s *StatusBar) renderHelp() string {
	if s.helpVisible {
		return s.renderExpandedHelp()
	}
	return s.renderCompactHelp()
}

// renderCompactHelp renders compact help text
func (s *StatusBar) renderCompactHelp() string {
	shortcuts := []string{
		s.theme.HelpKey.Render("?") + s.theme.HelpDesc.Render(":help"),
		s.theme.HelpKey.Render("^C") + s.theme.HelpDesc.Render(":quit"),
	}
	return strings.Join(shortcuts, " ")
}

// renderExpandedHelp renders expanded help text (when help is toggled)
func (s *StatusBar) renderExpandedHelp() string {
	// This would be shown in a separate help panel, not in status bar
	return s.theme.HelpDesc.Render("Press ? to toggle help")
}

// GetFullHelp returns full help text for display in a separate panel
func (s *StatusBar) GetFullHelp() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Keyboard Shortcuts") + "\n\n")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"Ctrl+C", "Quit Hex"},
		{"Ctrl+L", "Clear screen"},
		{"Ctrl+K", "Clear conversation"},
		{"Ctrl+S", "Save conversation"},
		{"Ctrl+E", "Export conversation"},
		{"Ctrl+T", "Toggle typewriter mode"},
		{"Tab", "Switch view (Chat/History/Tools)"},
		{"Enter", "Send message"},
		{"j/k", "Scroll down/up"},
		{"gg", "Go to top"},
		{"G", "Go to bottom"},
		{"/", "Search"},
		{"?", "Toggle help"},
		{"Esc", "Exit current mode/quit"},
	}

	for _, shortcut := range shortcuts {
		b.WriteString(s.theme.HelpKey.Render(fmt.Sprintf("%-12s", shortcut.key)))
		b.WriteString("  ")
		b.WriteString(s.theme.HelpDesc.Render(shortcut.desc))
		b.WriteString("\n")
	}

	return b.String()
}

// StatusBarUpdate represents an update to the status bar
type StatusBarUpdate struct {
	Tokens     *TokenUpdate
	Connection *ConnectionStatus
	Mode       *string
	Message    *string
}

// TokenUpdate represents a token count update
type TokenUpdate struct {
	Input  int
	Output int
}

// ApplyUpdate applies a status bar update
func (s *StatusBar) ApplyUpdate(update StatusBarUpdate) {
	if update.Tokens != nil {
		s.UpdateTokens(update.Tokens.Input, update.Tokens.Output)
	}
	if update.Connection != nil {
		s.SetConnectionStatus(*update.Connection)
	}
	if update.Mode != nil {
		s.SetMode(*update.Mode)
	}
	if update.Message != nil {
		s.SetCustomMessage(*update.Message)
	}
}
