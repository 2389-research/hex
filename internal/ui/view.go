// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Bubbletea view function for rendering UI
// ABOUTME: Renders viewport with messages and input textarea
package ui

import (
	"fmt"
	"strings"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/ui/forms"
	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m *Model) View() string {
	if !m.Ready {
		return "\n  Initializing..."
	}

	// Tool log overlay takes precedence over everything
	if m.toolLogOverlay {
		return m.renderToolLogOverlay()
	}

	var b strings.Builder

	// Neo-Terminal top status bar
	b.WriteString(m.renderNeoTerminalStatusBar())
	b.WriteString("\n")

	// Phase 6C: Show help if toggled
	if m.helpVisible {
		b.WriteString(m.renderHelpPanel() + "\n\n")
	}

	// Render different views based on CurrentView
	switch m.CurrentView {
	case ViewModeIntro:
		b.WriteString(m.renderIntroView())
	case ViewModeChat:
		b.WriteString(m.renderChatView())
	case ViewModeHistory:
		b.WriteString(m.renderHistoryView())
	case ViewModeTools:
		b.WriteString(m.renderToolsView())
	}

	b.WriteString("\n")

	// Phase 6C Task 6: Quick actions modal (takes precedence over everything except tool approval)
	if m.quickActionsMode {
		b.WriteString(m.renderQuickActionsModal() + "\n")
		// Skip input and other prompts
		b.WriteString("\n" + m.renderNeoTerminalBottomBar())
		return b.String()
	}

	// Phase 6C: Tool approval prompt (takes precedence over everything)
	if m.toolApprovalMode {
		b.WriteString(m.renderToolApprovalPromptEnhanced() + "\n")
	} else if m.executingTool {
		// Task 12: Tool execution indicator
		b.WriteString(m.renderToolStatus() + "\n")
		// TUI Polish: Show collapsed tool log (last 3 lines)
		if collapsedLog, _ := m.renderCollapsedToolLog(); collapsedLog != "" {
			b.WriteString(collapsedLog)
		}
	} else if m.SearchMode {
		// Search mode indicator
		searchPrompt := m.theme.SearchPrompt.Render(fmt.Sprintf("Search: %s_", m.SearchQuery))
		b.WriteString(searchPrompt + "\n")
	} else {
		// Input (only in chat view)
		if m.CurrentView == ViewModeChat {
			b.WriteString(m.theme.Input.Render(m.Input.View()) + "\n")

			// Phase 6C Task 4: Render autocomplete dropdown
			if m.autocomplete != nil && m.autocomplete.IsActive() {
				b.WriteString(m.renderAutocompleteDropdown() + "\n")
			}
		}
	}

	// Neo-Terminal bottom status bar
	b.WriteString("\n" + m.renderNeoTerminalBottomBar())

	return b.String()
}

// renderIntroView renders the startup welcome screen
func (m *Model) renderIntroView() string {
	var b strings.Builder

	// Center the content
	padding := strings.Repeat(" ", 10)

	// ASCII art logo with Dracula colors
	logoText := `
   ██╗  ██╗███████╗██╗  ██╗
   ██║  ██║██╔════╝╚██╗██╔╝
   ███████║█████╗   ╚███╔╝
   ██╔══██║██╔══╝   ██╔██╗
   ██║  ██║███████╗██╔╝ ██╗
   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
`
	logo := m.theme.Title.Render(logoText)

	// Pad each line of the logo
	logoLines := strings.Split(logo, "\n")
	for _, line := range logoLines {
		if strings.TrimSpace(line) != "" {
			b.WriteString(padding + line + "\n")
		}
	}
	b.WriteString("\n")

	// Welcome message
	welcome := m.theme.Emphasized.Render("Welcome to Hex!")
	tagline := m.theme.Muted.Render("Your intelligent command-line assistant powered by Claude")
	b.WriteString(padding + welcome + "\n")
	b.WriteString(padding + tagline + "\n\n\n")

	// Quick start guide
	quickStart := m.theme.Subtitle.Render("Quick Start:")
	b.WriteString(padding + quickStart + "\n\n")

	features := []struct {
		icon string
		desc string
	}{
		{"💬", "Chat naturally with Claude AI"},
		{"🔧", "Execute tools with approval workflow"},
		{"📝", "Manage conversation history"},
		{"⌨️", "Vi-style navigation (j/k, gg/G)"},
		{":", "Quick actions menu"},
		{"?", "Show help anytime"},
	}

	for _, f := range features {
		icon := m.theme.Success.Render(f.icon)
		desc := m.theme.Body.Render(f.desc)
		b.WriteString(padding + "  " + icon + "  " + desc + "\n")
	}

	b.WriteString("\n\n")

	return b.String()
}

// renderChatView renders the chat conversation view
func (m *Model) renderChatView() string {
	return m.Viewport.View()
}

// renderHistoryView renders the conversation history browser
func (m *Model) renderHistoryView() string {
	var b strings.Builder
	b.WriteString(m.theme.ViewMode.Render("📚 History Browser") + "\n\n")
	b.WriteString("(History browser not yet implemented)\n")
	b.WriteString("\nPress Tab to return to chat")
	return b.String()
}

// renderToolsView renders the tool inspector
func (m *Model) renderToolsView() string {
	var b strings.Builder
	b.WriteString(m.theme.ViewMode.Render("🔧 Tool Inspector") + "\n\n")
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

	// Permission mode indicator (Phase 3)
	permInfo := ""
	if m.toolExecutor != nil && m.toolExecutor.GetPermissionChecker() != nil {
		checker := m.toolExecutor.GetPermissionChecker()
		mode := checker.GetMode()

		var modeStyle lipgloss.Style
		switch mode.String() {
		case "auto":
			modeStyle = m.theme.Success
		case "deny":
			modeStyle = m.theme.Error
		case "ask":
			modeStyle = m.theme.Warning
		default:
			modeStyle = m.theme.Muted
		}

		permInfo = " Perms:" + modeStyle.Render(mode.String())
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
	leftPart := m.theme.TokenCounter.Render(tokenInfo + permInfo)
	middlePart := m.theme.ViewMode.Render(fmt.Sprintf("[%s]", viewMode))
	rightPart := m.theme.Muted.Render(help)

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

	return m.theme.StatusBar.Render(statusBar)
}

// Task 12: Tool UI rendering functions

// renderColoredRiskEmoji returns a risk emoji colored based on risk level
func (m *Model) renderColoredRiskEmoji(risk forms.RiskLevel) string {
	switch risk {
	case forms.RiskSafe:
		return lipgloss.NewStyle().Foreground(m.theme.Colors.Green).Render("✓")
	case forms.RiskCaution:
		return lipgloss.NewStyle().Foreground(m.theme.Colors.Yellow).Render("⚠")
	case forms.RiskDanger:
		return lipgloss.NewStyle().Foreground(m.theme.Colors.Red).Bold(true).Render("⚠⚠")
	default:
		return lipgloss.NewStyle().Foreground(m.theme.Colors.Comment).Render("?")
	}
}

// getToolParamPreview extracts a compact preview of the key parameter for a tool
func (m *Model) getToolParamPreview(tool *core.ToolUse) string {
	switch tool.Name {
	case "bash":
		if cmd, ok := tool.Input["command"].(string); ok {
			return truncateQuoted(cmd, 60)
		}
	case "read_file":
		// read_file uses "path" parameter
		if path, ok := tool.Input["path"].(string); ok {
			return truncateQuoted(path, 60)
		}
	case "write_file", "edit":
		if path, ok := tool.Input["file_path"].(string); ok {
			return truncateQuoted(path, 60)
		}
	case "grep", "glob":
		if pattern, ok := tool.Input["pattern"].(string); ok {
			return truncateQuoted(pattern, 50)
		}
	default:
		// For other tools, show first string parameter
		for _, val := range tool.Input {
			if str, ok := val.(string); ok && str != "" {
				return truncateQuoted(str, 50)
			}
		}
	}
	return ""
}

// truncateQuoted returns a quoted, truncated string
func truncateQuoted(s string, maxLen int) string {
	// Escape newlines for single-line display
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	if len(s) > maxLen {
		return fmt.Sprintf("%q", s[:maxLen-3]+"...")
	}
	return fmt.Sprintf("%q", s)
}

// renderToolStatus renders the tool execution status indicator
func (m *Model) renderToolStatus() string {
	if !m.executingTool {
		return ""
	}

	if len(m.executingToolUses) == 0 {
		return m.theme.ToolExecuting.Render("⏳ Executing tool...")
	}

	if len(m.executingToolUses) > 1 {
		return m.theme.ToolExecuting.Render(fmt.Sprintf("⏳ Executing %d tools...", len(m.executingToolUses)))
	}

	// Single tool - show name and key parameter
	tool := m.executingToolUses[0]
	toolName := tool.Name

	// Extract key parameter based on tool type
	var paramPreview string
	switch toolName {
	case "bash":
		if cmd, ok := tool.Input["command"].(string); ok {
			paramPreview = truncateString(cmd, 60)
		}
	case "read_file", "write_file":
		if path, ok := tool.Input["file_path"].(string); ok {
			paramPreview = truncateString(path, 60)
		}
	case "edit":
		if path, ok := tool.Input["file_path"].(string); ok {
			paramPreview = truncateString(path, 60)
		}
	case "grep":
		if pattern, ok := tool.Input["pattern"].(string); ok {
			paramPreview = truncateString(pattern, 40)
		}
	case "glob":
		if pattern, ok := tool.Input["pattern"].(string); ok {
			paramPreview = truncateString(pattern, 40)
		}
	default:
		// For other tools, try to get a reasonable preview
		for key, val := range tool.Input {
			if str, ok := val.(string); ok && str != "" {
				paramPreview = key + "=" + truncateString(str, 40)
				break
			}
		}
	}

	if paramPreview != "" {
		return m.theme.ToolExecuting.Render(fmt.Sprintf("⏳ %s: %s", toolName, paramPreview))
	}
	return m.theme.ToolExecuting.Render(fmt.Sprintf("⏳ Executing: %s...", toolName))
}

// truncateString truncates a string to maxLen, adding ellipsis if needed
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Phase 6C: Enhanced rendering methods

// renderHelpPanel renders the help panel when toggled
func (m *Model) renderHelpPanel() string {
	if m.statusBar == nil {
		return ""
	}

	helpStyle := m.theme.HelpPanel.Width(m.Width - 4)

	return helpStyle.Render(m.statusBar.GetFullHelp())
}

// renderToolApprovalPromptEnhanced renders enhanced tool approval UI with custom menu
// Uses autocomplete-style highlighting for consistent UX
func (m *Model) renderToolApprovalPromptEnhanced() string {
	if !m.toolApprovalMode || len(m.pendingToolUses) == 0 {
		return ""
	}

	var b strings.Builder

	// Use autocomplete styles for consistent look
	dropdownWidth := m.Width - 4
	if dropdownWidth < 40 {
		dropdownWidth = 40
	}
	boxStyle := m.theme.AutocompleteDropdown.Width(dropdownWidth)
	selectedStyle := m.theme.AutocompleteSelected
	normalStyle := m.theme.AutocompleteItem
	helpStyle := m.theme.AutocompleteHelp

	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Purple).
		Bold(true)

	// Title
	b.WriteString(titleStyle.Render("🛠  Tool Approval Required"))
	b.WriteString("\n")

	// Tool description - compact format
	if len(m.pendingToolUses) == 1 {
		tool := m.pendingToolUses[0]
		riskLevel := forms.AssessRiskLevel(tool)
		coloredRisk := m.renderColoredRiskEmoji(riskLevel)
		paramPreview := m.getToolParamPreview(tool)
		if paramPreview != "" {
			b.WriteString(fmt.Sprintf("%s %s(%s)", coloredRisk, tool.Name, paramPreview))
		} else {
			b.WriteString(fmt.Sprintf("%s %s()", coloredRisk, tool.Name))
		}
	} else {
		b.WriteString(fmt.Sprintf("%s %d tools:", m.renderColoredRiskEmoji(forms.RiskCaution), len(m.pendingToolUses)))
		for _, tool := range m.pendingToolUses {
			riskLevel := forms.AssessRiskLevel(tool)
			coloredRisk := m.renderColoredRiskEmoji(riskLevel)
			paramPreview := m.getToolParamPreview(tool)
			if paramPreview != "" {
				b.WriteString(fmt.Sprintf("\n  %s %s(%s)", coloredRisk, tool.Name, paramPreview))
			} else {
				b.WriteString(fmt.Sprintf("\n  %s %s()", coloredRisk, tool.Name))
			}
		}
	}
	b.WriteString("\n")

	// Options - all visible, highlight selected with autocomplete style
	options := []string{
		"✓ Approve (run this time)",
		"✗ Deny (skip this time)",
		"✓✓ Always Allow (never ask again)",
		"✗✗ Never Allow (block permanently)",
	}

	for i, opt := range options {
		if i == m.selectedApprovalOpt {
			b.WriteString(selectedStyle.Render("▸ " + opt))
		} else {
			b.WriteString(normalStyle.Render("  " + opt))
		}
		b.WriteString("\n")
	}

	// Hint
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓: navigate • Enter: submit"))

	return boxStyle.Render(b.String())
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

	// Check for queued messages and display queue status
	if m.agentSvc != nil && m.ConversationID != "" {
		queuedCount := m.agentSvc.QueuedPrompts(m.ConversationID)
		if queuedCount > 0 {
			var queueMsg string
			switch m.Status {
			case StatusStreaming:
				queueMsg = fmt.Sprintf("Agent working... (%d queued)", queuedCount)
			case StatusQueued:
				if queuedCount == 1 {
					queueMsg = "Queued (processing...)"
				} else {
					queueMsg = fmt.Sprintf("Queued (%d ahead)", queuedCount-1)
				}
			}
			if queueMsg != "" {
				m.statusBar.SetCustomMessage(queueMsg)
			}
		} else if m.Status == StatusQueued {
			// If status is queued but no queue items, reset to idle
			m.Status = StatusIdle
		}
	}

	return m.statusBar.View()
}

// Phase 6C Task 6: Quick Actions Modal Rendering

// renderQuickActionsModal renders the quick actions menu overlay
func (m *Model) renderQuickActionsModal() string {
	// Use theme styles
	modalStyle := m.theme.Modal.Width(60)
	titleStyle := m.theme.ModalTitle
	inputStyle := m.theme.Warning
	actionStyle := m.theme.Muted
	selectedActionStyle := m.theme.Emphasized
	helpStyle := m.theme.HelpDesc

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("Quick Actions"))
	content.WriteString("\n\n")

	// Input with prompt
	content.WriteString(inputStyle.Render(":"))
	content.WriteString(inputStyle.Render(m.quickActionsInput))
	content.WriteString(inputStyle.Render("_"))
	content.WriteString("\n\n")

	// Show filtered actions
	if len(m.quickActionsFiltered) == 0 {
		content.WriteString(actionStyle.Render("No matching actions"))
	} else {
		// Show up to 5 actions
		maxDisplay := 5
		if len(m.quickActionsFiltered) < maxDisplay {
			maxDisplay = len(m.quickActionsFiltered)
		}

		for i := 0; i < maxDisplay; i++ {
			action := m.quickActionsFiltered[i]

			// First action is "selected" (will be executed on Enter)
			var actionLine string
			if i == 0 {
				actionLine = fmt.Sprintf("▸ %s - %s", action.Usage, action.Description)
				content.WriteString(selectedActionStyle.Render(actionLine))
			} else {
				actionLine = fmt.Sprintf("  %s - %s", action.Usage, action.Description)
				content.WriteString(actionStyle.Render(actionLine))
			}
			content.WriteString("\n")
		}

		// Show count if there are more
		if len(m.quickActionsFiltered) > maxDisplay {
			more := fmt.Sprintf("\n  ... and %d more", len(m.quickActionsFiltered)-maxDisplay)
			content.WriteString(actionStyle.Render(more))
		}
	}

	content.WriteString("\n\n")

	// Help text
	helpText := "Enter: execute • Esc: cancel"
	content.WriteString(helpStyle.Render(helpText))

	return modalStyle.Render(content.String())
}

// Phase 6C Task 4: Autocomplete Rendering

// renderAutocompleteDropdown renders the autocomplete suggestions dropdown
func (m *Model) renderAutocompleteDropdown() string {
	if m.autocomplete == nil || !m.autocomplete.IsActive() {
		return ""
	}

	completions := m.autocomplete.GetCompletions()
	if len(completions) == 0 {
		return ""
	}

	// Use dedicated autocomplete styles for high contrast
	// Use full width minus some padding for the dropdown
	dropdownWidth := m.Width - 4
	if dropdownWidth < 40 {
		dropdownWidth = 40
	}
	dropdownStyle := m.theme.AutocompleteDropdown.Width(dropdownWidth)
	selectedStyle := m.theme.AutocompleteSelected
	normalStyle := m.theme.AutocompleteItem
	helpStyle := m.theme.AutocompleteHelp

	// Build dropdown content
	var content strings.Builder

	selectedIndex := m.autocomplete.GetSelectedIndex()

	// Description style - slightly dimmer than main text but still readable
	descStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)

	for i, completion := range completions {
		var line strings.Builder

		// Selection indicator and styling
		if i == selectedIndex {
			line.WriteString("▸ ")
			line.WriteString(completion.Display)

			// Add description if available
			if completion.Description != "" {
				line.WriteString(" (" + completion.Description + ")")
			}

			content.WriteString(selectedStyle.Render(line.String()))
		} else {
			line.WriteString("  ")
			line.WriteString(completion.Display)

			// Add description if available (dimmed but readable)
			if completion.Description != "" {
				line.WriteString(" ")
				line.WriteString(descStyle.Render("(" + completion.Description + ")"))
			}

			content.WriteString(normalStyle.Render(line.String()))
		}

		content.WriteString("\n")
	}

	// Add help text
	helpText := helpStyle.Render("↑↓: navigate • Enter: accept • Esc: cancel")
	content.WriteString("\n")
	content.WriteString(helpText)

	return dropdownStyle.Render(content.String())
}

