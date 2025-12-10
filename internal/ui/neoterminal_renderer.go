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

	// Model name (extract short name from full model ID)
	modelName := m.getShortModelName()

	// CWD (shortened)
	cwd := m.getShortCwd()

	// Format: ┗━━ ● status │ model │ cwd │ ⌃C quit ━━┛
	borderStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment).Bold(true)
	separatorStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Comment)
	modelStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Cyan)
	cwdStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Green)
	bindingsStyle := m.theme.Muted
	warningStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Yellow).Bold(true)

	leftPart := statusText
	modelPart := modelStyle.Render(modelName)
	cwdPart := cwdStyle.Render(cwd)

	// Show quit confirmation warning if pending
	var rightPart string
	if m.pendingQuit && time.Since(m.pendingQuitTime) < 2*time.Second {
		rightPart = warningStyle.Render("⌃C again to quit")
	} else {
		rightPart = bindingsStyle.Render("⌃C quit · ⇥ views")
	}

	leftLen := lipgloss.Width(leftPart)
	modelLen := lipgloss.Width(modelPart)
	cwdLen := lipgloss.Width(cwdPart)
	rightLen := lipgloss.Width(rightPart)

	// Calculate spacing - 12 for separators and borders
	totalContentLen := leftLen + modelLen + cwdLen + rightLen + 16
	fillLen := m.Width - totalContentLen
	if fillLen < 0 {
		fillLen = 0
	}

	bar := "┗━━ " +
		leftPart +
		separatorStyle.Render(" │ ") +
		modelPart +
		separatorStyle.Render(" │ ") +
		cwdPart +
		separatorStyle.Render(" │ ") +
		rightPart +
		" " + strings.Repeat("━", fillLen) + " ━━┛"

	return borderStyle.Render(bar)
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
