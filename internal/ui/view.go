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

	// Check for fullscreen overlay first (tool log, help, etc.)
	if m.overlayManager != nil && m.overlayManager.IsFullscreen() {
		return m.overlayManager.Render(m.Width, m.Height)
	}

	var b strings.Builder

	// Neo-Terminal top status bar
	b.WriteString(m.renderNeoTerminalStatusBar())
	b.WriteString("\n")

	// Phase 6C: Show help if toggled
	if m.helpVisible {
		b.WriteString(m.renderHelpPanel() + "\n\n")
	}

	// Calculate bottom overlay height if present (for viewport adjustment)
	var bottomOverlayContent string
	var bottomOverlayHeight int
	if m.overlayManager != nil && m.overlayManager.HasActive() && !m.overlayManager.IsFullscreen() {
		active := m.overlayManager.GetActive()

		// Calculate desired height with 50% cap (consistent with adjustViewportForOverlay)
		desiredHeight := active.GetDesiredHeight()
		maxAllowed := m.Height / 2
		if desiredHeight > maxAllowed {
			bottomOverlayHeight = maxAllowed
		} else {
			bottomOverlayHeight = desiredHeight
		}
		if bottomOverlayHeight < 1 {
			bottomOverlayHeight = 1
		}

		// Render overlay
		bottomOverlayContent = active.Render(m.Width, bottomOverlayHeight)
		// Note: Viewport height is adjusted in Update via adjustViewportForOverlay()
		// when overlays are pushed/popped - no mutation needed here
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

	// Render bottom overlay between viewport and input (if not fullscreen or quick actions)
	if !m.quickActionsMode && bottomOverlayContent != "" {
		b.WriteString(bottomOverlayContent + "\n")
	}

	// Phase 6C Task 6: Quick actions modal (takes precedence over everything except tool approval)
	if m.quickActionsMode {
		b.WriteString(m.renderQuickActionsModal() + "\n")
		// Skip input and other prompts
		b.WriteString("\n" + m.renderNeoTerminalBottomBar())
		return b.String()
	}

	// Input (only in chat view) - always render in chat mode
	if m.CurrentView == ViewModeChat {
		// Show queued messages above input if any exist
		if m.QueueCount() > 0 {
			// Show first message preview and count using theme styles
			// Use rune-based truncation to handle UTF-8 properly
			firstMsg := m.PeekQueue()
			runes := []rune(firstMsg)
			if len(runes) > 40 {
				firstMsg = string(runes[:37]) + "..."
			}
			queueIndicator := m.theme.QueueIndicator.Render(fmt.Sprintf("[Q:%d] ", m.QueueCount()))
			b.WriteString(queueIndicator + m.theme.QueuedMessage.Render("◷ "+firstMsg+" (↑ to edit)") + "\n")
		}

		// Add top border for input
		inputWidth := m.Width - 2 // Account for side borders
		if inputWidth < 10 {
			inputWidth = 10
		}
		borderStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)
		b.WriteString(borderStyle.Render(strings.Repeat("─", inputWidth)) + "\n")

		b.WriteString(m.theme.Input.Render(m.Input.View()) + "\n")

		// Add bottom border for input
		b.WriteString(borderStyle.Render(strings.Repeat("─", inputWidth)) + "\n")
	}

	// Note: Bottom overlays are rendered above, between viewport and input
	// Fullscreen overlays are handled at the top of View() function
	if m.SearchMode {
		// Search mode indicator
		searchPrompt := m.theme.SearchPrompt.Render(fmt.Sprintf("Search: %s_", m.SearchQuery))
		b.WriteString(searchPrompt + "\n")
	}

	// Display error message if present
	if m.ErrorMessage != "" {
		b.WriteString(m.renderErrorMessage() + "\n")
	}

	// Neo-Terminal bottom status bar (no extra newline before it)
	b.WriteString(m.renderNeoTerminalBottomBar())

	result := m.wrapContentWithBorders(b.String())

	// Ensure the total output doesn't exceed terminal height
	// This prevents the top frame from being pushed out of view
	lines := strings.Split(result, "\n")
	if len(lines) > m.Height && m.Height > 10 {
		// Keep top frame (line 0) and bottom frame (last line)
		// Truncate viewport content in the middle
		topFrame := lines[0]
		bottomFrame := lines[len(lines)-1]

		// Calculate how many middle lines we can keep
		middleAllowed := m.Height - 2 // Reserve space for top and bottom frames

		// Find where to split: keep some viewport and all of input/overlay/footer
		// For now, keep the last (middleAllowed) lines before the bottom frame
		// This ensures input/overlays are always visible
		startIdx := len(lines) - 1 - middleAllowed
		if startIdx < 1 {
			startIdx = 1
		}

		middleLines := lines[startIdx : len(lines)-1]

		// Reconstruct with top frame, truncated middle, and bottom frame
		newLines := []string{topFrame}
		newLines = append(newLines, middleLines...)
		newLines = append(newLines, bottomFrame)
		result = strings.Join(newLines, "\n")
	}

	return result
}

// wrapContentWithBorders adds side borders (┃) to each line of content
// This completes the Neo-Terminal frame started by the top/bottom status bars
func (m *Model) wrapContentWithBorders(content string) string {
	borderStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment).Bold(true)
	leftBorder := borderStyle.Render("┃")
	rightBorder := borderStyle.Render("┃")

	lines := strings.Split(content, "\n")
	var result strings.Builder

	// Width available for content (total width minus 2 for borders)
	contentWidth := m.Width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	for i, line := range lines {
		// Skip adding borders to the top and bottom status bars (they already have corners)
		if i == 0 || (i == len(lines)-1 && strings.Contains(line, "┗━")) {
			result.WriteString(line)
		} else {
			// Pad or truncate line to exact content width
			lineWidth := lipgloss.Width(line)
			if lineWidth < contentWidth {
				line = line + strings.Repeat(" ", contentWidth-lineWidth)
			} else if lineWidth > contentWidth {
				// Truncate if too long
				line = line[:contentWidth]
			}
			result.WriteString(leftBorder + line + rightBorder)
		}
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// renderIntroView renders the startup welcome screen with bold hexagonal branding
func (m *Model) renderIntroView() string {
	var b strings.Builder

	// Hexagonal ASCII art - distinctive and memorable
	hexLogo := []string{
		"           ╱╲",
		"          ╱  ╲",
		"         ╱ ⬡⬡ ╲",
		"        ╱  ██  ╲",
		"       ╱   ██   ╲",
		"      ╱────────────╲",
		"      ╲────────────╱",
		"       ╲   ██   ╱",
		"        ╲  ██  ╱",
		"         ╲ ⬡⬡ ╱",
		"          ╲  ╱",
		"           ╲╱",
	}

	// Title text using block characters
	titleLines := []string{
		"██╗  ██╗ ███████╗ ██╗  ██╗",
		"██║  ██║ ██╔════╝ ╚██╗██╔╝",
		"███████║ █████╗    ╚███╔╝ ",
		"██╔══██║ ██╔══╝    ██╔██╗ ",
		"██║  ██║ ███████╗ ██╔╝ ██╗",
		"╚═╝  ╚═╝ ╚══════╝ ╚═╝  ╚═╝",
	}

	// Render hexagon with accent color
	hexStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Cyan)
	hexInnerStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Green).Bold(true)

	// Center padding based on terminal width
	contentWidth := 60
	leftPad := (m.Width - contentWidth) / 2
	if leftPad < 4 {
		leftPad = 4
	}
	padding := strings.Repeat(" ", leftPad)

	b.WriteString("\n")

	// Render hex logo with inner detail highlighting
	for _, line := range hexLogo {
		styled := hexStyle.Render(line)
		// Highlight inner elements
		styled = strings.ReplaceAll(styled, "██", hexInnerStyle.Render("██"))
		styled = strings.ReplaceAll(styled, "⬡⬡", hexInnerStyle.Render("⬡⬡"))
		b.WriteString(padding + "       " + styled + "\n")
	}

	b.WriteString("\n")

	// Render title with gradient-like effect (left to right color shift)
	titleStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Green).Bold(true)
	for _, line := range titleLines {
		b.WriteString(padding + titleStyle.Render(line) + "\n")
	}

	b.WriteString("\n")

	// Tagline with subtle styling
	taglineStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment).Italic(true)
	tagline := "Claude-powered intelligence for your terminal"
	b.WriteString(padding + taglineStyle.Render(tagline) + "\n")

	b.WriteString("\n")

	// Decorative separator
	sepStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)
	separator := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	b.WriteString(padding + sepStyle.Render(separator) + "\n\n")

	// Quick commands in a more compact, scannable format
	cmdStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Cyan).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Foreground)
	mutedStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)

	commands := []struct {
		key  string
		desc string
	}{
		{"Enter", "Send message"},
		{"Tab", "Cycle views"},
		{"Ctrl+H", "Toggle help"},
		{":", "Quick actions"},
		{"Ctrl+C", "Quit"},
	}

	// Render commands in a horizontal layout
	var cmdParts []string
	for _, cmd := range commands {
		part := cmdStyle.Render(cmd.key) + mutedStyle.Render(" → ") + descStyle.Render(cmd.desc)
		cmdParts = append(cmdParts, part)
	}

	// Two rows of commands
	row1 := strings.Join(cmdParts[:3], "   ")
	row2 := strings.Join(cmdParts[3:], "   ")
	b.WriteString(padding + row1 + "\n")
	b.WriteString(padding + row2 + "\n")

	b.WriteString("\n")

	// Ready prompt with subtle animation hint
	readyStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Green)
	b.WriteString(padding + readyStyle.Render("▸ ") + descStyle.Render("Type a message to begin...") + "\n")

	b.WriteString("\n")

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

// Phase 6C: Enhanced rendering methods

// renderHelpPanel renders the help panel when toggled
func (m *Model) renderHelpPanel() string {
	if m.statusBar == nil {
		return ""
	}

	helpStyle := m.theme.HelpPanel.Width(m.Width - 4)

	return helpStyle.Render(m.statusBar.GetFullHelp())
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

// renderErrorMessage renders error messages prominently
func (m *Model) renderErrorMessage() string {
	if m.ErrorMessage == "" {
		return ""
	}

	// Use a prominent error style with full width
	errorBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Colors.Red).
		Foreground(m.theme.Colors.Red).
		Padding(0, 1).
		Width(m.Width - 4)

	// Format the error with icon
	errorText := fmt.Sprintf("⚠ Error: %s", m.ErrorMessage)

	return errorBoxStyle.Render(errorText)
}
