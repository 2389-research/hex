// ABOUTME: Bubbletea view function for rendering UI
// ABOUTME: Renders viewport with messages and input textarea
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("235"))

	statusIdleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("35")).
			Bold(true)

	statusStreamingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	viewModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)

	tokenCounterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))

	searchModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	// Task 12: Tool UI styles
	toolApprovalStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("208")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("208")).
				Padding(1, 2)

	toolExecutingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))
)

// View renders the UI
func (m *Model) View() string {
	if !m.Ready {
		return "\n  Initializing..."
	}

	var b strings.Builder

	// Title with status indicator
	title := titleStyle.Render(fmt.Sprintf("Clem • %s", m.Model))
	statusIndicator := m.renderStatusIndicator()
	b.WriteString(title + " " + statusIndicator)

	// Phase 6C: Add streaming indicator if streaming
	if m.Streaming && m.streamingDisplay != nil {
		b.WriteString("  ")
		b.WriteString(m.streamingDisplay.RenderStreamingIndicator())
	}

	// Phase 6C: Add spinner if active
	if m.spinner != nil && m.spinner.IsActive() {
		b.WriteString("  ")
		b.WriteString(m.spinner.View())
	}

	b.WriteString("\n\n")

	// Phase 6C: Show help if toggled
	if m.helpVisible {
		b.WriteString(m.renderHelpPanel() + "\n\n")
	}

	// Render different views based on CurrentView
	switch m.CurrentView {
	case ViewModeChat:
		b.WriteString(m.renderChatView())
	case ViewModeHistory:
		b.WriteString(m.renderHistoryView())
	case ViewModeTools:
		b.WriteString(m.renderToolsView())
	}

	b.WriteString("\n")

	// Phase 6C: Tool approval prompt (takes precedence over everything)
	if m.toolApprovalMode {
		b.WriteString(m.renderToolApprovalPromptEnhanced() + "\n")
	} else if m.executingTool {
		// Task 12: Tool execution indicator
		b.WriteString(m.renderToolStatus() + "\n")
	} else if m.SearchMode {
		// Search mode indicator
		searchPrompt := searchModeStyle.Render(fmt.Sprintf("Search: %s_", m.SearchQuery))
		b.WriteString(searchPrompt + "\n")
	} else {
		// Input (only in chat view)
		if m.CurrentView == ViewModeChat {
			b.WriteString(inputStyle.Render(m.Input.View()) + "\n")
		}
	}

	// Phase 6C: Enhanced status bar
	b.WriteString("\n" + m.renderStatusBarEnhanced())

	return b.String()
}

// renderStatusIndicator renders the current status icon
func (m *Model) renderStatusIndicator() string {
	switch m.Status {
	case StatusStreaming:
		return statusStreamingStyle.Render("●")
	case StatusTyping:
		return statusIdleStyle.Render("●")
	case StatusError:
		return statusErrorStyle.Render("●")
	default:
		return statusIdleStyle.Render("●")
	}
}

// renderChatView renders the chat conversation view
func (m *Model) renderChatView() string {
	return m.Viewport.View()
}

// renderHistoryView renders the conversation history browser
func (m *Model) renderHistoryView() string {
	var b strings.Builder
	b.WriteString(viewModeStyle.Render("📚 History Browser") + "\n\n")
	b.WriteString("(History browser not yet implemented)\n")
	b.WriteString("\nPress Tab to return to chat")
	return b.String()
}

// renderToolsView renders the tool inspector
func (m *Model) renderToolsView() string {
	var b strings.Builder
	b.WriteString(viewModeStyle.Render("🔧 Tool Inspector") + "\n\n")
	b.WriteString("(Tool inspector not yet implemented)\n")
	b.WriteString("\nPress Tab to return to chat")
	return b.String()
}

// renderStatusBar renders the bottom status bar with token counter and help
func (m *Model) renderStatusBar() string {
	// Token counter
	tokenInfo := ""
	if m.TokensInput > 0 || m.TokensOutput > 0 {
		tokenInfo = fmt.Sprintf("Tokens: %d in / %d out", m.TokensInput, m.TokensOutput)
	}

	// View mode indicator
	viewMode := ""
	switch m.CurrentView {
	case ViewModeChat:
		viewMode = "Chat"
	case ViewModeHistory:
		viewMode = "History"
	case ViewModeTools:
		viewMode = "Tools"
	}

	// Help text
	help := "ctrl+c: quit • enter: send • tab: switch view • /: search • j/k: scroll • gg/G: top/bottom"

	// Compose status bar
	leftPart := tokenCounterStyle.Render(tokenInfo)
	middlePart := viewModeStyle.Render(fmt.Sprintf("[%s]", viewMode))
	rightPart := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(help)

	// Calculate spacing
	width := m.Width
	if width < 80 {
		width = 80
	}

	// Calculate spacing with safety check
	spacingWidth := width - lipgloss.Width(leftPart) - lipgloss.Width(middlePart) - lipgloss.Width(rightPart) - 10
	if spacingWidth < 0 {
		spacingWidth = 0
	}

	statusBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPart,
		strings.Repeat(" ", 2),
		middlePart,
		strings.Repeat(" ", spacingWidth),
		rightPart,
	)

	return statusBarStyle.Render(statusBar)
}

// Task 12: Tool UI rendering functions

// renderToolApprovalPrompt renders the tool approval UI
func (m *Model) renderToolApprovalPrompt() string {
	if !m.toolApprovalMode || m.pendingToolUse == nil {
		return ""
	}

	var prompt strings.Builder
	prompt.WriteString("⚠ Tool Approval Required\n\n")
	prompt.WriteString(fmt.Sprintf("Tool: %s\n", m.pendingToolUse.Name))

	// Format parameters nicely
	if len(m.pendingToolUse.Input) > 0 {
		prompt.WriteString("Parameters:\n")
		for key, value := range m.pendingToolUse.Input {
			// Truncate long values
			valueStr := fmt.Sprintf("%v", value)
			if len(valueStr) > 100 {
				valueStr = valueStr[:97] + "..."
			}
			prompt.WriteString(fmt.Sprintf("  %s: %s\n", key, valueStr))
		}
	}

	prompt.WriteString("\nAllow this tool to execute? (y/n): ")

	return toolApprovalStyle.Render(prompt.String())
}

// renderToolStatus renders the tool execution status indicator
func (m *Model) renderToolStatus() string {
	if !m.executingTool {
		return ""
	}

	toolName := "unknown"
	if m.pendingToolUse != nil {
		toolName = m.pendingToolUse.Name
	}

	return toolExecutingStyle.Render(fmt.Sprintf("⏳ Executing tool: %s...", toolName))
}

// Phase 6C: Enhanced rendering methods

// renderHelpPanel renders the help panel when toggled
func (m *Model) renderHelpPanel() string {
	if m.statusBar == nil {
		return ""
	}

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(m.Width - 4)

	return helpStyle.Render(m.statusBar.GetFullHelp())
}

// renderToolApprovalPromptEnhanced renders enhanced tool approval UI
func (m *Model) renderToolApprovalPromptEnhanced() string {
	if !m.toolApprovalMode || m.pendingToolUse == nil {
		return m.renderToolApprovalPrompt() // Fallback to basic version
	}

	// Use the new ApprovalPrompt component if available
	if m.approvalPrompt == nil {
		m.approvalPrompt = NewApprovalPrompt(m.pendingToolUse)
		m.approvalPrompt.SetWidth(m.Width)
	}

	return m.approvalPrompt.View()
}

// renderStatusBarEnhanced renders the enhanced status bar
func (m *Model) renderStatusBarEnhanced() string {
	if m.statusBar == nil {
		return m.renderStatusBar() // Fallback to basic version
	}

	// Update status bar state
	m.statusBar.SetWidth(m.Width)
	m.statusBar.SetTokens(m.TokensInput, m.TokensOutput)

	// Update connection status
	if m.Streaming {
		m.statusBar.SetConnectionStatus(ConnectionStreaming)
	} else if m.apiClient != nil {
		m.statusBar.SetConnectionStatus(ConnectionConnected)
	} else {
		m.statusBar.SetConnectionStatus(ConnectionDisconnected)
	}

	// Update mode
	mode := ""
	switch m.CurrentView {
	case ViewModeChat:
		mode = "chat"
	case ViewModeHistory:
		mode = "history"
	case ViewModeTools:
		mode = "tools"
	}
	m.statusBar.SetMode(mode)

	return m.statusBar.View()
}
