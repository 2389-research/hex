// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Tool approval UI component for displaying and handling tool approval prompts
// ABOUTME: Shows formatted tool information with risk levels and interactive approval controls
package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/pagent/internal/core"
)

// RiskLevel represents the risk level of a tool operation
type RiskLevel int

const (
	// RiskSafe indicates a tool operation is safe and requires no approval
	RiskSafe RiskLevel = iota
	// RiskCaution indicates a tool operation should be reviewed before execution
	RiskCaution
	// RiskDanger indicates a tool operation is potentially dangerous
	RiskDanger
)

var (
	approvalBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("208")).
				Padding(1, 2).
				Width(70)

	approvalTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("208"))

	approvalToolNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))

	approvalParamKeyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))

	approvalParamValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	approvalHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)

	riskSafeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("35"))

	riskCautionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226"))

	riskDangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	syntaxKeywordStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)

	syntaxStringStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("114"))

	syntaxNumberStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("141"))
)

// ApprovalPrompt manages the tool approval UI
type ApprovalPrompt struct {
	toolUse     *core.ToolUse
	riskLevel   RiskLevel
	showDetails bool
	width       int
}

// NewApprovalPrompt creates a new approval prompt for a tool use
func NewApprovalPrompt(toolUse *core.ToolUse) *ApprovalPrompt {
	return &ApprovalPrompt{
		toolUse:     toolUse,
		riskLevel:   assessRiskLevel(toolUse),
		showDetails: false,
		width:       70,
	}
}

// ToggleDetails toggles the detailed view
func (a *ApprovalPrompt) ToggleDetails() {
	a.showDetails = !a.showDetails
}

// SetWidth sets the width of the approval prompt
func (a *ApprovalPrompt) SetWidth(width int) {
	a.width = width
	if a.width > 80 {
		a.width = 80
	}
	if a.width < 50 {
		a.width = 50
	}
}

// View renders the approval prompt
func (a *ApprovalPrompt) View() string {
	var b strings.Builder

	// Title
	b.WriteString(approvalTitleStyle.Render("┌─ Tool Approval Required "))
	b.WriteString(strings.Repeat("─", a.width-28))
	b.WriteString("┐\n")

	// Tool name
	b.WriteString("│ Tool: ")
	b.WriteString(approvalToolNameStyle.Render(a.toolUse.Name))
	b.WriteString(strings.Repeat(" ", a.width-10-len(a.toolUse.Name)))
	b.WriteString("│\n")

	// Risk level
	riskText, riskStyle := a.formatRiskLevel()
	b.WriteString("│ Risk: ")
	b.WriteString(riskStyle.Render(riskText))
	padding := a.width - 10 - lipgloss.Width(riskText)
	if padding > 0 {
		b.WriteString(strings.Repeat(" ", padding))
	}
	b.WriteString("│\n")

	// Separator
	b.WriteString("│" + strings.Repeat(" ", a.width-2) + "│\n")

	// Parameters
	if len(a.toolUse.Input) > 0 {
		b.WriteString("│ Parameters:")
		b.WriteString(strings.Repeat(" ", a.width-14))
		b.WriteString("│\n")

		for key, value := range a.toolUse.Input {
			formattedValue := a.formatParameterValue(value)
			lines := a.wrapText(formattedValue, a.width-10)

			// First line with key
			b.WriteString("│   ")
			b.WriteString(approvalParamKeyStyle.Render(key + ":"))
			b.WriteString(" ")
			b.WriteString(lines[0])
			padding := a.width - 6 - len(key) - lipgloss.Width(lines[0])
			if padding > 0 {
				b.WriteString(strings.Repeat(" ", padding))
			}
			b.WriteString("│\n")

			// Additional lines for wrapped values
			for i := 1; i < len(lines); i++ {
				b.WriteString("│     ")
				b.WriteString(lines[i])
				padding := a.width - 7 - lipgloss.Width(lines[i])
				if padding > 0 {
					b.WriteString(strings.Repeat(" ", padding))
				}
				b.WriteString("│\n")
			}
		}
		b.WriteString("│" + strings.Repeat(" ", a.width-2) + "│\n")
	}

	// Details section (if toggled)
	if a.showDetails {
		b.WriteString("│ Working Dir: /current/directory")
		b.WriteString(strings.Repeat(" ", a.width-35))
		b.WriteString("│\n")
		b.WriteString("│ ID: " + a.toolUse.ID)
		b.WriteString(strings.Repeat(" ", a.width-7-len(a.toolUse.ID)))
		b.WriteString("│\n")
		b.WriteString("│" + strings.Repeat(" ", a.width-2) + "│\n")
	}

	// Help text
	helpText := "[A]pprove  [D]eny"
	if !a.showDetails {
		helpText += "  [V]iew Details"
	}
	b.WriteString("│ ")
	b.WriteString(approvalHelpStyle.Render(helpText))
	padding = a.width - 4 - len(helpText)
	if padding > 0 {
		b.WriteString(strings.Repeat(" ", padding))
	}
	b.WriteString("│\n")

	// Bottom border
	b.WriteString("└")
	b.WriteString(strings.Repeat("─", a.width-2))
	b.WriteString("┘")

	return approvalBoxStyle.Render(b.String())
}

// formatRiskLevel returns formatted risk level text and style
func (a *ApprovalPrompt) formatRiskLevel() (string, lipgloss.Style) {
	switch a.riskLevel {
	case RiskSafe:
		return "Safe ✓", riskSafeStyle
	case RiskCaution:
		return "Caution ⚠", riskCautionStyle
	case RiskDanger:
		return "Danger ⚠⚠", riskDangerStyle
	default:
		return "Unknown", riskCautionStyle
	}
}

// formatParameterValue formats a parameter value with syntax highlighting
func (a *ApprovalPrompt) formatParameterValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Apply syntax highlighting for code-like strings
		if a.looksLikeCode(v) {
			return a.highlightCode(v)
		}
		// Truncate long strings
		if len(v) > 200 {
			return syntaxStringStyle.Render("\"" + v[:197] + "...\"")
		}
		return syntaxStringStyle.Render("\"" + v + "\"")
	case int, int64, float64:
		return syntaxNumberStyle.Render(fmt.Sprintf("%v", v))
	case bool:
		return syntaxKeywordStyle.Render(fmt.Sprintf("%v", v))
	case map[string]interface{}:
		// Format as JSON
		jsonBytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		jsonStr := string(jsonBytes)
		if len(jsonStr) > 200 {
			jsonStr = jsonStr[:197] + "..."
		}
		return approvalParamValueStyle.Render(jsonStr)
	default:
		str := fmt.Sprintf("%v", v)
		if len(str) > 200 {
			str = str[:197] + "..."
		}
		return approvalParamValueStyle.Render(str)
	}
}

// looksLikeCode checks if a string looks like code (contains keywords, operators, etc.)
func (a *ApprovalPrompt) looksLikeCode(s string) bool {
	codeIndicators := []string{"git ", "rm ", "mkdir ", "cd ", "bash ", "sh ", "python ", "node "}
	for _, indicator := range codeIndicators {
		if strings.Contains(s, indicator) {
			return true
		}
	}
	return false
}

// highlightCode applies basic syntax highlighting to code strings
func (a *ApprovalPrompt) highlightCode(code string) string {
	// Very basic highlighting - just highlight dangerous commands
	dangerous := []string{"rm", "delete", "drop", "truncate", "format"}
	result := code

	for _, cmd := range dangerous {
		if strings.Contains(strings.ToLower(result), cmd) {
			result = strings.ReplaceAll(result, cmd, riskDangerStyle.Render(cmd))
		}
	}

	return result
}

// wrapText wraps text to fit within a given width
func (a *ApprovalPrompt) wrapText(text string, width int) []string {
	if lipgloss.Width(text) <= width {
		return []string{text}
	}

	// Simple word wrapping
	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if lipgloss.Width(testLine) <= width {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// assessRiskLevel determines the risk level of a tool operation
func assessRiskLevel(toolUse *core.ToolUse) RiskLevel {
	if toolUse == nil {
		return RiskCaution
	}

	// Check tool name for dangerous operations
	dangerousTools := []string{"bash", "exec", "shell"}
	for _, dangerous := range dangerousTools {
		if strings.Contains(strings.ToLower(toolUse.Name), dangerous) {
			// Check parameters for dangerous commands
			if input, ok := toolUse.Input["command"].(string); ok {
				if containsDangerousCommand(input) {
					return RiskDanger
				}
				return RiskCaution
			}
			return RiskCaution
		}
	}

	// Read-only operations are generally safe
	safeTools := []string{"read", "list", "get", "search", "find"}
	for _, safe := range safeTools {
		if strings.Contains(strings.ToLower(toolUse.Name), safe) {
			return RiskSafe
		}
	}

	return RiskCaution
}

// containsDangerousCommand checks if a command string contains dangerous operations
func containsDangerousCommand(cmd string) bool {
	dangerousCommands := []string{
		"rm ", "rm -rf", "delete", "drop", "truncate",
		"format", "mkfs", "dd if=", "> /dev/",
		"curl | sh", "wget | sh", "chmod +x", "sudo",
	}

	cmdLower := strings.ToLower(cmd)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(cmdLower, dangerous) {
			return true
		}
	}
	return false
}
