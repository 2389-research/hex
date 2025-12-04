// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Status bar component for displaying connection, token, and mode information
// ABOUTME: Provides comprehensive bottom status bar with color-coded indicators
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jeff/internal/ui/themes"
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
	theme         themes.Theme
}

// statusBarStyles holds all dynamically generated styles for the status bar
type statusBarStyles struct {
	background             lipgloss.Style
	modelName              lipgloss.Style
	tokenUsage             lipgloss.Style
	tokenHigh              lipgloss.Style
	contextSize            lipgloss.Style
	connectionConnected    lipgloss.Style
	connectionStreaming    lipgloss.Style
	connectionError        lipgloss.Style
	connectionDisconnected lipgloss.Style
	mode                   lipgloss.Style
	helpKey                lipgloss.Style
	helpDesc               lipgloss.Style
	customMessage          lipgloss.Style
}

// createStatusBarStyles generates styles from the theme
func (s *StatusBar) createStatusBarStyles() statusBarStyles {
	theme := s.theme
	return statusBarStyles{
		background: lipgloss.NewStyle().
			Background(theme.Background()).
			Foreground(theme.Foreground()).
			Padding(0, 1),
		modelName: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Bold(true),
		tokenUsage: lipgloss.NewStyle().
			Foreground(theme.Subtle()),
		tokenHigh: lipgloss.NewStyle().
			Foreground(theme.Warning()).
			Bold(true),
		contextSize: lipgloss.NewStyle().
			Foreground(theme.Secondary()),
		connectionConnected: lipgloss.NewStyle().
			Foreground(theme.Success()).
			Bold(true),
		connectionStreaming: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Bold(true),
		connectionError: lipgloss.NewStyle().
			Foreground(theme.Error()).
			Bold(true),
		connectionDisconnected: lipgloss.NewStyle().
			Foreground(theme.Subtle()),
		mode: lipgloss.NewStyle().
			Foreground(theme.Secondary()).
			Bold(true),
		helpKey: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Bold(true),
		helpDesc: lipgloss.NewStyle().
			Foreground(theme.Subtle()),
		customMessage: lipgloss.NewStyle().
			Foreground(theme.Warning()).
			Italic(true),
	}
}

// NewStatusBar creates a new status bar
func NewStatusBar(model string, width int, theme themes.Theme) *StatusBar {
	return &StatusBar{
		model:       model,
		width:       width,
		connection:  ConnectionDisconnected,
		currentMode: "chat",
		contextSize: 200000, // Default context size
		theme:       theme,
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

// SetConnection is an alias for SetConnectionStatus for API compatibility
func (s *StatusBar) SetConnection(status ConnectionStatus) {
	s.SetConnectionStatus(status)
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
	styles := s.createStatusBarStyles()

	if s.width < 40 {
		// Too narrow, show minimal info
		return styles.background.Render("Pagen")
	}

	var parts []string

	// Left section: Model name
	parts = append(parts, styles.modelName.Render(s.model))

	// Connection indicator
	parts = append(parts, s.renderConnection(styles))

	// Token usage
	if s.tokensTotal > 0 {
		parts = append(parts, s.renderTokenUsage(styles))
	}

	// Context size indicator (if tokens are getting high)
	if s.tokensTotal > s.contextSize/2 {
		parts = append(parts, s.renderContextIndicator(styles))
	}

	// Current mode
	parts = append(parts, styles.mode.Render("["+s.currentMode+"]"))

	// Custom message (if any)
	if s.customMessage != "" {
		parts = append(parts, styles.customMessage.Render(s.customMessage))
	}

	// Combine left parts
	leftSection := strings.Join(parts, " ")

	// Right section: Help text
	rightSection := s.renderHelp(styles)

	// Calculate spacing
	usedWidth := lipgloss.Width(leftSection) + lipgloss.Width(rightSection)
	spacing := s.width - usedWidth - 4 // Account for padding
	if spacing < 1 {
		spacing = 1
	}

	// Combine sections
	bar := leftSection + strings.Repeat(" ", spacing) + rightSection

	return styles.background.Width(s.width).Render(bar)
}

// renderConnection renders the connection status indicator
func (s *StatusBar) renderConnection(styles statusBarStyles) string {
	switch s.connection {
	case ConnectionConnected:
		return styles.connectionConnected.Render("●")
	case ConnectionStreaming:
		return styles.connectionStreaming.Render("◉")
	case ConnectionError:
		return styles.connectionError.Render("◉")
	case ConnectionDisconnected:
		return styles.connectionDisconnected.Render("○")
	default:
		return styles.connectionDisconnected.Render("○")
	}
}

// renderTokenUsage renders token usage information
func (s *StatusBar) renderTokenUsage(styles statusBarStyles) string {
	// Highlight if usage is getting high
	usagePercent := float64(s.tokensTotal) / float64(s.contextSize) * 100

	tokensText := fmt.Sprintf("%dk↓ %dk↑", s.tokensInput/1000, s.tokensOutput/1000)

	if usagePercent > 80 {
		return styles.tokenHigh.Render(tokensText)
	}
	return styles.tokenUsage.Render(tokensText)
}

// renderContextIndicator renders context size indicator
func (s *StatusBar) renderContextIndicator(styles statusBarStyles) string {
	usagePercent := float64(s.tokensTotal) / float64(s.contextSize) * 100
	bars := int(usagePercent / 10)
	if bars > 10 {
		bars = 10
	}

	indicator := "["
	for i := 0; i < 10; i++ {
		if i < bars {
			if usagePercent > 80 {
				indicator += styles.tokenHigh.Render("█")
			} else {
				indicator += styles.contextSize.Render("█")
			}
		} else {
			indicator += "░"
		}
	}
	indicator += "]"

	return indicator
}

// renderHelp renders the help text
func (s *StatusBar) renderHelp(styles statusBarStyles) string {
	if s.helpVisible {
		return s.renderExpandedHelp(styles)
	}
	return s.renderCompactHelp(styles)
}

// renderCompactHelp renders compact help text
func (s *StatusBar) renderCompactHelp(styles statusBarStyles) string {
	shortcuts := []string{
		styles.helpKey.Render("?") + styles.helpDesc.Render(":help"),
		styles.helpKey.Render("^C") + styles.helpDesc.Render(":quit"),
	}
	return strings.Join(shortcuts, " ")
}

// renderExpandedHelp renders expanded help text (when help is toggled)
func (s *StatusBar) renderExpandedHelp(styles statusBarStyles) string {
	// This would be shown in a separate help panel, not in status bar
	return styles.helpDesc.Render("Press ? to toggle help")
}

// GetFullHelp returns full help text for display in a separate panel
func (s *StatusBar) GetFullHelp() string {
	styles := s.createStatusBarStyles()
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(s.theme.Primary()).Render("Keyboard Shortcuts") + "\n\n")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"Ctrl+C", "Quit Pagen"},
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
		b.WriteString(styles.helpKey.Render(fmt.Sprintf("%-12s", shortcut.key)))
		b.WriteString("  ")
		b.WriteString(styles.helpDesc.Render(shortcut.desc))
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
