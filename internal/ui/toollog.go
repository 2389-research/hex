// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Tool output log functionality for displaying tool execution output
// ABOUTME: Collapsed 3-line view with Ctrl+O overlay for full log
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// appendToolLogLine adds a single line to the tool log
func (m *Model) appendToolLogLine(line string) {
	m.toolLogLines = append(m.toolLogLines, line)
}

// appendToolLogOutput adds multi-line output to the tool log
func (m *Model) appendToolLogOutput(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue // Skip empty lines
		}
		// Strip STDOUT:/STDERR: prefixes from tool output
		line = strings.TrimPrefix(line, "STDOUT:")
		line = strings.TrimPrefix(line, "STDERR:")
		line = strings.TrimSpace(line)
		if line != "" {
			m.toolLogLines = append(m.toolLogLines, line)
		}
	}
}

// getToolLogLastN returns the last n lines of tool output
func (m *Model) getToolLogLastN(n int) []string {
	if len(m.toolLogLines) <= n {
		return m.toolLogLines
	}
	return m.toolLogLines[len(m.toolLogLines)-n:]
}

// clearToolLogChunk clears the current tool log chunk
func (m *Model) clearToolLogChunk() {
	m.toolLogLines = nil
	m.currentToolLogName = ""
	m.currentToolLogParam = ""
}

// startToolLogEntry starts a new tool entry in the log
func (m *Model) startToolLogEntry(toolName, paramPreview string) {
	m.currentToolLogName = toolName
	m.currentToolLogParam = paramPreview
	// Add header line for this tool
	header := fmt.Sprintf("─── %s(%s) ───", toolName, paramPreview)
	m.toolLogLines = append(m.toolLogLines, header)
}

// updateMostRecentToolID updates the cached most recent tool ID
// This should be called whenever m.toolResults changes
func (m *Model) updateMostRecentToolID() {
	// Iterate through messages in reverse order to find most recent
	for i := len(m.Messages) - 1; i >= 0; i-- {
		msg := m.Messages[i]
		// Check content blocks in reverse order within the message
		for j := len(msg.ContentBlock) - 1; j >= 0; j-- {
			block := msg.ContentBlock[j]
			if block.Type == "tool_use" {
				// Check if this tool has a result in history
				for _, tr := range m.toolResultHistory {
					if tr.ToolUseID == block.ID {
						// Found a tool with a result - cache it
						m.mostRecentToolID = block.ID
						return
					}
				}
			}
		}
	}
	// No tool with result found
	m.mostRecentToolID = ""
}

// getMostRecentToolWithResult finds the most recent tool_use ID that has a result (not pending)
// Returns empty string if no tool with result is found
// DEPRECATED: Use m.mostRecentToolID instead (updated via updateMostRecentToolID)
func (m *Model) getMostRecentToolWithResult() string {
	// Iterate through messages in reverse order to find most recent
	for i := len(m.Messages) - 1; i >= 0; i-- {
		msg := m.Messages[i]
		// Check content blocks in reverse order within the message
		for j := len(msg.ContentBlock) - 1; j >= 0; j-- {
			block := msg.ContentBlock[j]
			if block.Type == "tool_use" {
				// Check if this tool has a result in history
				for _, tr := range m.toolResultHistory {
					if tr.ToolUseID == block.ID {
						// Found a tool with a result
						return block.ID
					}
				}
			}
		}
	}
	return ""
}

// renderCollapsedToolLog renders the last 3 lines of tool output with dimmed style
// Returns the rendered string and the number of hidden lines (for combining with hint)
func (m *Model) renderCollapsedToolLog() (string, int) {
	if len(m.toolLogLines) == 0 {
		return "", 0
	}

	// Style: dimmed with │ prefix
	dimStyle := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Comment)

	var b strings.Builder

	// Calculate hidden lines
	totalLines := len(m.toolLogLines)
	hiddenLines := 0
	if totalLines > 3 {
		hiddenLines = totalLines - 3
	}

	// Show last 3 lines
	last3 := m.getToolLogLastN(3)
	for _, line := range last3 {
		b.WriteString(dimStyle.Render("│ " + line))
		b.WriteString("\n")
	}

	return b.String(), hiddenLines
}
