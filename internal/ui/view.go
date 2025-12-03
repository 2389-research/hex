// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Bubbletea view function for rendering UI
// ABOUTME: Renders viewport with messages and input textarea
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/pagent/internal/ui/themes"
)

// Consistent spacing and layout constants
const (
	paddingHorizontal = 2
	paddingVertical   = 1
	marginBottom      = 1
	borderRadius      = 1
)

// viewStyles holds all the lipgloss styles for the view, created from the theme
type viewStyles struct {
	title            lipgloss.Style
	input            lipgloss.Style
	statusBar        lipgloss.Style
	statusIdle       lipgloss.Style
	statusStreaming  lipgloss.Style
	statusError      lipgloss.Style
	viewMode         lipgloss.Style
	tokenCounter     lipgloss.Style
	searchMode       lipgloss.Style
	toolApproval     lipgloss.Style
	toolExecuting    lipgloss.Style
	suggestionBox    lipgloss.Style
	suggestionTool   lipgloss.Style
	suggestionReason lipgloss.Style
	suggestionHint   lipgloss.Style
}

// createViewStyles creates all view styles from the current theme with consistent spacing
func (m *Model) createViewStyles() viewStyles {
	theme := m.theme

	return viewStyles{
		title: lipgloss.NewStyle().
			Padding(paddingVertical, paddingHorizontal).
			Bold(true).
			Foreground(theme.Primary()).
			MarginBottom(marginBottom),

		input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderFocus()).
			Padding(0, paddingHorizontal).
			MarginTop(marginBottom),

		statusBar: lipgloss.NewStyle().
			Foreground(theme.Subtle()).
			Background(theme.Background()),

		statusIdle: lipgloss.NewStyle().
			Foreground(theme.Success()).
			Bold(true),

		statusStreaming: lipgloss.NewStyle().
			Foreground(theme.Warning()).
			Bold(true),

		statusError: lipgloss.NewStyle().
			Foreground(theme.Error()).
			Bold(true),

		viewMode: lipgloss.NewStyle().
			Foreground(theme.Secondary()).
			Bold(true),

		tokenCounter: lipgloss.NewStyle().
			Foreground(theme.Subtle()),

		searchMode: lipgloss.NewStyle().
			Foreground(theme.Warning()).
			Background(theme.Background()).
			Padding(0, 1),

		toolApproval: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Warning()).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Warning()).
			Padding(1, 2),

		toolExecuting: lipgloss.NewStyle().
			Foreground(theme.Primary()),

		suggestionBox: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Background(theme.Background()).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Primary()).
			Padding(0, 1),

		suggestionTool: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Bold(true),

		suggestionReason: lipgloss.NewStyle().
			Foreground(theme.Subtle()).
			Italic(true),

		suggestionHint: lipgloss.NewStyle().
			Foreground(theme.Subtle()),
	}
}

// View renders the UI
func (m *Model) View() string {
	if !m.Ready {
		return "\n  Initializing..."
	}

	// Create styles from theme
	styles := m.createViewStyles()

	var b strings.Builder

	// Title with status indicator and favorite star - use gradient with decorative border!
	titleText := fmt.Sprintf("Pagen • %s", m.Model)
	if m.IsFavorite {
		titleText = "⭐ " + titleText
	}
	// Apply gradient to title using theme's title gradient
	titleGradient := themes.RenderGradient(titleText, m.theme.TitleGradient())

	// Add decorative border with theme colors
	titleStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.BorderFocus()).
		Padding(0, 2).
		MarginBottom(1)

	titleWithBorder := titleStyle.Render(titleGradient)
	statusIndicator := m.renderStatusIndicator(styles)
	b.WriteString(titleWithBorder + " " + statusIndicator)

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
		b.WriteString(m.renderHistoryView(styles))
	case ViewModeTools:
		b.WriteString(m.renderToolsView(styles))
	}

	b.WriteString("\n")

	// Phase 6C Task 6: Quick actions modal (takes precedence over everything except tool approval)
	if m.quickActionsMode {
		b.WriteString(m.renderQuickActionsModal() + "\n")
		// Skip input and other prompts
		b.WriteString("\n" + m.renderStatusBarEnhanced(styles))
		return b.String()
	}

	// Phase 6C: Tool approval prompt (takes precedence over everything)
	// Phase 2: Use Huh approval if available
	if m.toolApprovalMode && m.huhApproval != nil {
		approvalView := m.huhApproval.View()
		b.WriteString(approvalView + "\n")
	} else if m.toolApprovalMode {
		// Fallback to old approval prompt
		b.WriteString(m.renderToolApprovalPromptEnhanced(styles) + "\n")
	} else if m.executingTool {
		// Task 12: Tool execution indicator
		b.WriteString(m.renderToolStatus(styles) + "\n")
	} else if m.SearchMode {
		// Search mode indicator
		searchPrompt := styles.searchMode.Render(fmt.Sprintf("Search: %s_", m.SearchQuery))
		b.WriteString(searchPrompt + "\n")
	} else {
		// Input (only in chat view)
		if m.CurrentView == ViewModeChat {
			b.WriteString(styles.input.Render(m.Input.View()) + "\n")

			// Phase 6C Task 4: Render autocomplete dropdown
			if m.autocomplete != nil && m.autocomplete.IsActive() {
				b.WriteString(m.renderAutocompleteDropdown() + "\n")
			}

			// Phase 6C Task 8: Render smart suggestions
			if m.showSuggestions && len(m.suggestions) > 0 {
				b.WriteString(m.renderSuggestions(styles) + "\n")
			}
		}
	}

	if m.tokenViz != nil && m.TokensInput+m.TokensOutput > 0 {
		m.tokenViz.SetWidth(m.Width - 4)
		b.WriteString(m.tokenViz.RenderCompact() + "\n")
	}

	// Phase 6C: Enhanced status bar
	b.WriteString("\n" + m.renderStatusBarEnhanced(styles))

	return b.String()
}

// renderStatusIndicator renders the current status icon
func (m *Model) renderStatusIndicator(styles viewStyles) string {
	switch m.Status {
	case StatusStreaming:
		return styles.statusStreaming.Render("●")
	case StatusTyping:
		return styles.statusIdle.Render("●")
	case StatusError:
		return styles.statusError.Render("●")
	default:
		return styles.statusIdle.Render("●")
	}
}

// renderChatView renders the chat conversation view
func (m *Model) renderChatView() string {
	return m.Viewport.View()
}

// renderHistoryView renders the conversation history browser
func (m *Model) renderHistoryView(styles viewStyles) string {
	var b strings.Builder
	b.WriteString(styles.viewMode.Render("📚 History Browser") + "\n\n")
	b.WriteString("(History browser not yet implemented)\n")
	b.WriteString("\nPress Tab to return to chat")
	return b.String()
}

// renderToolsView renders the tool inspector
func (m *Model) renderToolsView(styles viewStyles) string {
	var b strings.Builder
	b.WriteString(styles.viewMode.Render("🔧 Tool Inspector") + "\n\n")
	b.WriteString("(Tool inspector not yet implemented)\n")
	b.WriteString("\nPress Tab to return to chat")
	return b.String()
}

// renderStatusBar renders the bottom status bar with token counter and help
func (m *Model) renderStatusBar(styles viewStyles) string {
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
	leftPart := styles.tokenCounter.Render(tokenInfo)
	middlePart := styles.viewMode.Render(fmt.Sprintf("[%s]", viewMode))
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

	return styles.statusBar.Render(statusBar)
}

// Task 12: Tool UI rendering functions

// renderToolApprovalPrompt renders the tool approval UI
func (m *Model) renderToolApprovalPrompt(styles viewStyles) string {
	if !m.toolApprovalMode || len(m.pendingToolUses) == 0 {
		return ""
	}

	var prompt strings.Builder
	prompt.WriteString("⚠ Tool Approval Required\n\n")

	// Handle single vs multiple tools
	if len(m.pendingToolUses) == 1 {
		tool := m.pendingToolUses[0]
		prompt.WriteString(fmt.Sprintf("Tool: %s\n", tool.Name))

		// Format parameters nicely
		if len(tool.Input) > 0 {
			prompt.WriteString("Parameters:\n")
			for key, value := range tool.Input {
				// Truncate long values
				valueStr := fmt.Sprintf("%v", value)
				if len(valueStr) > 100 {
					valueStr = valueStr[:97] + "..."
				}
				prompt.WriteString(fmt.Sprintf("  %s: %s\n", key, valueStr))
			}
		}
	} else {
		// Multiple tools - show summary
		prompt.WriteString(fmt.Sprintf("The assistant wants to execute %d tools:\n\n", len(m.pendingToolUses)))
		for i, tool := range m.pendingToolUses {
			prompt.WriteString(fmt.Sprintf("%d. %s", i+1, tool.Name))
			if len(tool.Input) > 0 {
				// Show brief parameter summary
				keys := make([]string, 0, len(tool.Input))
				for k := range tool.Input {
					keys = append(keys, k)
				}
				prompt.WriteString(fmt.Sprintf(" (%s)", strings.Join(keys, ", ")))
			}
			prompt.WriteString("\n")
		}
	}

	prompt.WriteString("\nAllow these tool(s) to execute? (y/n): ")

	return styles.toolApproval.Render(prompt.String())
}

// renderToolStatus renders the tool execution status indicator
func (m *Model) renderToolStatus(styles viewStyles) string {
	if !m.executingTool {
		return ""
	}

	toolName := "unknown"
	if len(m.executingToolUses) > 0 {
		if len(m.executingToolUses) == 1 {
			toolName = m.executingToolUses[0].Name
		} else {
			toolName = fmt.Sprintf("%d tools", len(m.executingToolUses))
		}
	}

	return styles.toolExecuting.Render(fmt.Sprintf("⏳ Executing: %s...", toolName))
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
func (m *Model) renderToolApprovalPromptEnhanced(styles viewStyles) string {
	if !m.toolApprovalMode || len(m.pendingToolUses) == 0 {
		return m.renderToolApprovalPrompt(styles) // Fallback to basic version
	}

	// Use the new ApprovalPrompt component if available
	// For now, only use enhanced prompt for single tools
	if len(m.pendingToolUses) == 1 {
		if m.approvalPrompt == nil {
			m.approvalPrompt = NewApprovalPrompt(m.pendingToolUses[0])
			m.approvalPrompt.SetWidth(m.Width)
		}
		return m.approvalPrompt.View()
	}

	// For multiple tools, fall back to basic version which handles the list
	return m.renderToolApprovalPrompt(styles)
}

// renderStatusBarEnhanced renders the enhanced status bar
func (m *Model) renderStatusBarEnhanced(styles viewStyles) string {
	if m.statusBar == nil {
		return m.renderStatusBar(styles) // Fallback to basic version
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

// Phase 6C Task 6: Quick Actions Modal Rendering

// renderQuickActionsModal renders the quick actions menu overlay
func (m *Model) renderQuickActionsModal() string {
	// Styles for the modal using theme colors
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border()).
		Padding(1, 2).
		Width(60)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary())

	inputStyle := lipgloss.NewStyle().
		Foreground(m.theme.Warning())

	actionStyle := lipgloss.NewStyle().
		Foreground(m.theme.Subtle())

	selectedActionStyle := lipgloss.NewStyle().
		Foreground(m.theme.Success()).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Subtle()).
		Italic(true)

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

	// Styles
	dropdownStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		MaxWidth(60)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Background(lipgloss.Color("237"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	typeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	// Build dropdown content
	var content strings.Builder

	selectedIndex := m.autocomplete.GetSelectedIndex()

	for i, completion := range completions {
		var line strings.Builder

		// Selection indicator and styling
		if i == selectedIndex {
			line.WriteString("▸ ")
			// Highlight the completion
			line.WriteString(completion.Display)

			// Add description if available
			if completion.Description != "" {
				line.WriteString(" ")
				line.WriteString(typeStyle.Render("(" + completion.Description + ")"))
			}

			content.WriteString(selectedStyle.Render(line.String()))
		} else {
			line.WriteString("  ")
			line.WriteString(completion.Display)

			// Add description if available
			if completion.Description != "" {
				line.WriteString(" ")
				line.WriteString(typeStyle.Render("(" + completion.Description + ")"))
			}

			content.WriteString(normalStyle.Render(line.String()))
		}

		content.WriteString("\n")
	}

	// Add help text
	helpText := typeStyle.Render("↑↓: navigate • Enter: accept • Esc: cancel")
	content.WriteString("\n")
	content.WriteString(helpText)

	return dropdownStyle.Render(content.String())
}

// Phase 6C Task 8: renderSuggestions renders smart tool suggestions
func (m *Model) renderSuggestions(styles viewStyles) string {
	if !m.showSuggestions || len(m.suggestions) == 0 {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(styles.suggestionTool.Render("💡 Suggestions") + "\n\n")

	// Show top suggestion prominently
	topSuggestion := m.suggestions[0]
	content.WriteString(styles.suggestionTool.Render(fmt.Sprintf("→ %s", topSuggestion.ToolName)) + "\n")
	content.WriteString("  " + styles.suggestionReason.Render(topSuggestion.Reason) + "\n")
	content.WriteString("  " + styles.suggestionHint.Render(fmt.Sprintf("Action: %s", topSuggestion.Action)) + "\n")

	// Show additional suggestions if any
	if len(m.suggestions) > 1 {
		content.WriteString("\n" + styles.suggestionReason.Render("Other suggestions:") + "\n")
		for i := 1; i < len(m.suggestions) && i < 3; i++ {
			s := m.suggestions[i]
			content.WriteString(fmt.Sprintf("  • %s ", s.ToolName))
			content.WriteString(styles.suggestionReason.Render(fmt.Sprintf("(%.0f%% confident)", s.Confidence*100)) + "\n")
		}
	}

	// Help text
	content.WriteString("\n" + styles.suggestionHint.Render("Tab: accept • Esc: dismiss"))

	return styles.suggestionBox.Render(content.String())
}
