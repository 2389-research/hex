// Package ui provides Neo-Terminal rendering functions for sophisticated message display.
// ABOUTME: Neo-Terminal renderer - elegant bordered message containers with rich visual hierarchy
// ABOUTME: Implements Swiss typography meets cyberdeck aesthetic for terminal interfaces
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderNeoTerminalMessage creates a compact message with role indicator and timestamp
func (m *Model) renderNeoTerminalMessage(role, content string, timestamp time.Time) string {
	var b strings.Builder

	// Determine colors and label based on role
	var roleColor lipgloss.Color
	var roleLabel string

	switch role {
	case "user":
		roleColor = m.theme.Colors.Orange // Accent Coral
		roleLabel = "YOU"
	case "assistant":
		roleColor = m.theme.Colors.Green // Accent Sage
		roleLabel = "ASSISTANT"
	case "system":
		roleColor = m.theme.Colors.Cyan // Accent Sky
		roleLabel = "SYSTEM"
	default:
		roleColor = m.theme.Colors.Comment
		roleLabel = "MESSAGE"
	}

	// Format timestamp
	timeStr := timestamp.Format("15:04")

	// Create styles
	roleStyle := lipgloss.NewStyle().Foreground(roleColor).Bold(true)
	timeStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)

	// Header line: ROLE [TIME]
	header := roleStyle.Render(roleLabel) + " " + timeStyle.Render("["+timeStr+"]")
	b.WriteString(header + "\n")

	// Content with simple prefix
	prefix := "  " // Simple indent

	// Split on existing newlines (glamour/user input preserves formatting)
	// Trim trailing empty lines to avoid excessive spacing
	content = strings.TrimRight(content, "\n")
	if content == "" {
		// If content is completely empty, just show header
		return b.String()
	}

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if line == "" {
			b.WriteString("\n")
		} else {
			// Don't add extra styling - content is already styled (glamour) or plain (user)
			b.WriteString(prefix + line + "\n")
		}
	}

	return b.String()
}

// renderNeoTerminalStatusBar creates the heavy-bordered top status bar
func (m *Model) renderNeoTerminalStatusBar() string {
	// Format: ┏━━ ⬡ HEX ━━━━━━━━━━━━━━━━ ◆ tokens: 1.2k in · 890 out ━━┓

	// View mode indicator
	viewSymbol := "⬡"
	viewText := "HEX"

	switch m.CurrentView {
	case ViewModeHistory:
		viewText = "HEX › HISTORY"
	case ViewModeTools:
		viewText = "HEX › TOOLS"
	case ViewModeIntro:
		viewText = "HEX › WELCOME"
	}

	if m.SearchMode {
		viewSymbol = "⌕"
		viewText = "SEARCH"
	}

	leftPart := fmt.Sprintf("%s %s", viewSymbol, viewText)

	// Token counter
	var rightPart string
	if m.TokensInput > 0 || m.TokensOutput > 0 {
		rightPart = fmt.Sprintf("◆ tokens: %d in · %d out", m.TokensInput, m.TokensOutput)
	} else {
		rightPart = ""
	}

	// Calculate fill
	leftLen := lipgloss.Width(leftPart)
	rightLen := lipgloss.Width(rightPart)
	fillLen := m.Width - leftLen - rightLen - 8 // Account for borders and padding
	if fillLen < 0 {
		fillLen = 0
	}

	// Create styles
	borderStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Foreground).Bold(true)
	tokenStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)

	bar := "┏━━ " +
		textStyle.Render(leftPart) +
		" " + strings.Repeat("━", fillLen) + " " +
		tokenStyle.Render(rightPart) +
		" ━━┓"

	return borderStyle.Render(bar)
}

// renderNeoTerminalBottomBar creates the heavy-bordered bottom status bar with shortcuts
func (m *Model) renderNeoTerminalBottomBar() string {
	// Status indicator
	var statusText string
	statusStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)

	switch m.Status {
	case StatusIdle:
		statusText = "● idle"
	case StatusStreaming:
		statusText = statusStyle.Foreground(m.theme.Colors.Cyan).Render("● streaming")
	case StatusError:
		statusText = statusStyle.Foreground(m.theme.Colors.Red).Render("● error")
	}

	// Shortcuts
	shortcuts := "/clear /exit /help"

	// Key bindings
	bindings := "⌃C quit · ⇥ views · ⌃L clear"

	// Format: ┗━━ ● status │ shortcuts │ bindings ━━┛
	leftPart := statusText
	middlePart := m.theme.Muted.Render(shortcuts)
	rightPart := m.theme.Muted.Render(bindings)

	leftLen := lipgloss.Width(leftPart)
	middleLen := lipgloss.Width(middlePart)
	rightLen := lipgloss.Width(rightPart)

	// Calculate spacing
	totalContentLen := leftLen + middleLen + rightLen + 8 // " │ " separators and padding
	fillLen := m.Width - totalContentLen
	if fillLen < 0 {
		fillLen = 0
	}

	borderStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment).Bold(true)
	separatorStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)

	bar := "┗━━ " +
		leftPart +
		separatorStyle.Render(" │ ") +
		middlePart +
		separatorStyle.Render(" │ ") +
		rightPart +
		" " + strings.Repeat("━", fillLen) + " ━━┛"

	return borderStyle.Render(bar)
}
