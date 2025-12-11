// Package ui provides Neo-Terminal rendering functions for sophisticated message display.
// ABOUTME: Neo-Terminal renderer - elegant bordered message containers with rich visual hierarchy
// ABOUTME: Implements Swiss typography meets cyberdeck aesthetic for terminal interfaces
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderNeoTerminalMessage creates a message with colored left border
func (m *Model) renderNeoTerminalMessage(role, content string, timestamp time.Time) string {
	var b strings.Builder

	// Determine border color based on role
	var borderColor lipgloss.Color

	switch role {
	case "user":
		borderColor = m.theme.Colors.Orange // Orange for user
	case "assistant":
		borderColor = m.theme.Colors.Green // Green for Hex/assistant
	case "system":
		borderColor = m.theme.Colors.Cyan // Cyan for system
	default:
		borderColor = m.theme.Colors.Comment
	}

	// Create border style
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	border := borderStyle.Render("┃")

	// Trim trailing empty lines to avoid excessive spacing
	content = strings.TrimRight(content, "\n")
	if content == "" {
		// Empty message - just return empty string
		return ""
	}

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// Every line gets the colored border prefix
		b.WriteString(border + " " + line + "\n")
	}

	return b.String()
}

// renderNeoTerminalStatusBar creates the heavy-bordered top status bar
func (m *Model) renderNeoTerminalStatusBar() string {
	// Format: ┏━ ⬡ HEX ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ ◆ tokens: 1.2k in · 890 out ━┓
	// Total width = m.Width exactly

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

	// Token counter
	var rightPart string
	if m.TokensInput > 0 || m.TokensOutput > 0 {
		rightPart = fmt.Sprintf("◆ tokens: %d in · %d out", m.TokensInput, m.TokensOutput)
	} else {
		rightPart = ""
	}

	// Create styles
	borderStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment).Bold(true)
	symbolStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Foreground).Bold(true)
	hexStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Green).Bold(true) // Hex green!
	textStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Foreground).Bold(true)
	tokenStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)

	// Build styled left part (symbol + HEX in green + any suffix)
	var leftPart string
	if viewText == "HEX" {
		leftPart = symbolStyle.Render(viewSymbol) + " " + hexStyle.Render("HEX")
	} else if strings.HasPrefix(viewText, "HEX › ") {
		// For "HEX › HISTORY", "HEX › TOOLS", etc.
		suffix := strings.TrimPrefix(viewText, "HEX › ")
		leftPart = symbolStyle.Render(viewSymbol) + " " + hexStyle.Render("HEX") + " " + textStyle.Render("› "+suffix)
	} else {
		// For "SEARCH" and other non-HEX modes
		leftPart = symbolStyle.Render(viewSymbol) + " " + textStyle.Render(viewText)
	}

	// Build bar content first to measure
	leftLen := lipgloss.Width(leftPart)
	rightLen := lipgloss.Width(rightPart)

	// Structure: ┏━ (2) + space + leftPart + space + fill + [space + rightPart + space] + ━┓ (2)
	// With rightPart: ┏━ (2) + sp (1) + left + sp (1) + fill + sp (1) + right + sp (1) + ━┓ (2) = 8 fixed
	// Without rightPart: ┏━ (2) + sp (1) + left + sp (1) + fill + ━┓ (2) = 6 fixed (no trailing space before ━┓)
	var fixedChars int
	if rightPart != "" {
		fixedChars = 8
	} else {
		fixedChars = 6
	}
	fillLen := m.Width - leftLen - rightLen - fixedChars
	if fillLen < 0 {
		fillLen = 0
	}

	var bar string
	if rightPart != "" {
		bar = borderStyle.Render("┏━") + " " +
			leftPart +
			" " + borderStyle.Render(strings.Repeat("━", fillLen)) + " " +
			tokenStyle.Render(rightPart) +
			" " + borderStyle.Render("━┓")
	} else {
		// No trailing space when no rightPart - fill goes directly to corner
		bar = borderStyle.Render("┏━") + " " +
			leftPart +
			" " + borderStyle.Render(strings.Repeat("━", fillLen)+"━┓")
	}

	return bar
}

// renderNeoTerminalBottomBar creates the heavy-bordered bottom status bar with shortcuts
func (m *Model) renderNeoTerminalBottomBar() string {
	// Format: ┗━ ● status │ model │ cwd │ ⌃C quit ━━━━━━━━━━━━━━━━━━━━━━━━ ━┛
	// Total width = m.Width exactly (matches top bar structure)

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

	// Model name (extract short name from full model ID)
	modelName := m.getShortModelName()

	// CWD (shortened)
	cwd := m.getShortCwd()

	borderStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment).Bold(true)
	separatorStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)
	modelStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Cyan)
	cwdStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Green)
	bindingsStyle := m.theme.Muted
	warningStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Yellow).Bold(true)

	leftPart := statusText
	modelPart := modelStyle.Render(modelName)
	cwdPart := cwdStyle.Render(cwd)

	// Show quit confirmation warning if pending, or timestamp if hovering over a message
	var rightPart string
	var timestampSuffix string
	if m.pendingQuit && time.Since(m.pendingQuitTime) < 2*time.Second {
		rightPart = warningStyle.Render("⌃C again to quit")
	} else if m.hoveredMessageIndex >= 0 && !m.hoveredMessageTime.IsZero() {
		// Show timestamp when hovering over a message
		timeStr := m.hoveredMessageTime.Format("15:04")
		timestampStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)
		timestampSuffix = timestampStyle.Render("sent at " + timeStr)
		rightPart = bindingsStyle.Render("⌃C quit · ⇥ views")
	} else {
		rightPart = bindingsStyle.Render("⌃C quit · ⇥ views")
	}

	// Build the content part
	contentPart := leftPart +
		separatorStyle.Render(" │ ") +
		modelPart +
		separatorStyle.Render(" │ ") +
		cwdPart +
		separatorStyle.Render(" │ ") +
		rightPart

	contentLen := lipgloss.Width(contentPart)
	timestampLen := lipgloss.Width(timestampSuffix)

	// Structure: ┗━ (2) + space + content + space + fill + [timestamp] + ━┛ (2)
	// With timestamp: ┗━ (2) + sp (1) + content + sp (1) + fill + sp (1) + timestamp + sp (1) + ━┛ (2) = 8 fixed
	// Without timestamp: ┗━ (2) + sp (1) + content + sp (1) + fill + ━┛ (2) = 6 fixed
	var fillLen int
	if timestampSuffix != "" {
		fixedChars := 8
		fillLen = m.Width - contentLen - timestampLen - fixedChars
	} else {
		fixedChars := 6
		fillLen = m.Width - contentLen - fixedChars
	}
	if fillLen < 0 {
		fillLen = 0
	}

	var bar string
	if timestampSuffix != "" {
		bar = borderStyle.Render("┗━") + " " +
			contentPart +
			" " + borderStyle.Render(strings.Repeat("━", fillLen)) + " " +
			timestampSuffix +
			" " + borderStyle.Render("━┛")
	} else {
		bar = borderStyle.Render("┗━") + " " +
			contentPart +
			" " + borderStyle.Render(strings.Repeat("━", fillLen)+"━┛")
	}

	return bar
}

// getShortModelName extracts a short, readable model name from the full model ID
func (m *Model) getShortModelName() string {
	if m.Model == "" {
		return "unknown"
	}

	// Common model ID patterns to short names
	modelName := m.Model
	switch {
	case strings.Contains(modelName, "opus"):
		return "opus"
	case strings.Contains(modelName, "sonnet"):
		return "sonnet"
	case strings.Contains(modelName, "haiku"):
		return "haiku"
	case strings.Contains(modelName, "gpt-4"):
		return "gpt-4"
	case strings.Contains(modelName, "gpt-3"):
		return "gpt-3.5"
	case strings.Contains(modelName, "gemini"):
		if strings.Contains(modelName, "pro") {
			return "gemini-pro"
		}
		return "gemini"
	default:
		// Truncate long model names
		if len(modelName) > 15 {
			return modelName[:12] + "..."
		}
		return modelName
	}
}

// getShortCwd returns a shortened version of the current working directory
func (m *Model) getShortCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "~"
	}

	// Try to make it relative to home directory
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(cwd, home) {
		cwd = "~" + cwd[len(home):]
	}

	// Shorten long paths by showing only last 2 components
	parts := strings.Split(cwd, string(filepath.Separator))
	if len(parts) > 3 {
		return "…/" + strings.Join(parts[len(parts)-2:], "/")
	}

	return cwd
}
