// Package forms provides beautiful huh-based forms for the clem TUI.
// ABOUTME: Huh-based tool approval form for interactive tool permission requests
// ABOUTME: Provides beautiful select-based approval UI with multiple options
package forms

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/ui/theme"
)

// ApprovalDecision represents the user's decision on tool approval
type ApprovalDecision string

const (
	// DecisionApprove approves this single tool execution
	DecisionApprove ApprovalDecision = "approve"
	// DecisionDeny denies this single tool execution
	DecisionDeny ApprovalDecision = "deny"
	// DecisionAlwaysAllow always allows this tool without future prompts
	DecisionAlwaysAllow ApprovalDecision = "always_allow"
	// DecisionNeverAllow blocks this tool permanently
	DecisionNeverAllow ApprovalDecision = "never_allow"
)

// ApprovalFormResult contains the result of the approval form
type ApprovalFormResult struct {
	Decision ApprovalDecision
	ToolUse  *core.ToolUse
}

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

// ToolApprovalForm creates a beautiful huh form for tool approval
type ToolApprovalForm struct {
	toolUse   *core.ToolUse
	riskLevel RiskLevel
	decision  ApprovalDecision
	theme     *theme.Theme
}

// NewToolApprovalForm creates a new tool approval form
func NewToolApprovalForm(toolUse *core.ToolUse) *ToolApprovalForm {
	return &ToolApprovalForm{
		toolUse:   toolUse,
		riskLevel: assessRiskLevel(toolUse),
		theme:     theme.DraculaTheme(),
	}
}

// Run displays the form and returns the user's decision
func (f *ToolApprovalForm) Run() (ApprovalFormResult, error) {
	// Build the approval options
	options := []huh.Option[ApprovalDecision]{
		huh.NewOption("✓ Approve (run this time)", DecisionApprove),
		huh.NewOption("✗ Deny (skip this time)", DecisionDeny),
		huh.NewOption("✓✓ Always Allow (never ask again)", DecisionAlwaysAllow),
		huh.NewOption("✗✗ Never Allow (block permanently)", DecisionNeverAllow),
	}

	// Format tool information for display
	toolInfo := f.formatToolInfo()
	riskInfo := f.formatRiskInfo()
	paramInfo := f.formatParameterInfo()

	// Build the form description
	description := fmt.Sprintf("%s\n\n%s\n\n%s",
		toolInfo,
		riskInfo,
		paramInfo,
	)

	// Create the form with Dracula theme colors
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[ApprovalDecision]().
				Title("🛠  Tool Approval Required").
				Description(description).
				Options(options...).
				Value(&f.decision).
				Height(12),
		),
	).WithTheme(f.getDraculaTheme())

	// Run the form
	err := form.Run()
	if err != nil {
		return ApprovalFormResult{}, err
	}

	return ApprovalFormResult{
		Decision: f.decision,
		ToolUse:  f.toolUse,
	}, nil
}

// formatToolInfo formats basic tool information
func (f *ToolApprovalForm) formatToolInfo() string {
	return fmt.Sprintf("Tool: %s\nID: %s", f.toolUse.Name, f.toolUse.ID)
}

// formatRiskInfo formats risk level information with color
func (f *ToolApprovalForm) formatRiskInfo() string {
	switch f.riskLevel {
	case RiskSafe:
		return "Risk Level: Safe ✓"
	case RiskCaution:
		return "Risk Level: Caution ⚠"
	case RiskDanger:
		return "Risk Level: DANGER ⚠⚠"
	default:
		return "Risk Level: Unknown"
	}
}

// formatParameterInfo formats tool parameters for display
func (f *ToolApprovalForm) formatParameterInfo() string {
	if len(f.toolUse.Input) == 0 {
		return "Parameters: (none)"
	}

	var b strings.Builder
	b.WriteString("Parameters:\n")

	for key, value := range f.toolUse.Input {
		valueStr := formatValue(value)
		// Truncate long values
		if len(valueStr) > 100 {
			valueStr = valueStr[:97] + "..."
		}
		b.WriteString(fmt.Sprintf("  • %s: %s\n", key, valueStr))
	}

	return strings.TrimSpace(b.String())
}

// formatValue formats a parameter value for display
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if len(v) > 200 {
			return fmt.Sprintf("%q (truncated)", v[:197]+"...")
		}
		return fmt.Sprintf("%q", v)
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case map[string]interface{}:
		return fmt.Sprintf("{...} (%d keys)", len(v))
	default:
		str := fmt.Sprintf("%v", v)
		if len(str) > 200 {
			return str[:197] + "..."
		}
		return str
	}
}

// getDraculaTheme returns a huh theme configured with Dracula colors
func (f *ToolApprovalForm) getDraculaTheme() *huh.Theme {
	t := huh.ThemeBase()

	// Use the theme instance colors which are already lipgloss.Color
	colors := f.theme.Colors

	// Configure with Dracula colors using lipgloss.Color
	// These match the Dracula theme palette
	t.Focused.Base = t.Focused.Base.
		BorderForeground(colors.Purple)

	t.Focused.Title = t.Focused.Title.
		Foreground(colors.Purple).
		Bold(true)

	t.Focused.Description = t.Focused.Description.
		Foreground(colors.Foreground)

	t.Focused.SelectSelector = t.Focused.SelectSelector.
		Foreground(colors.Pink)

	t.Focused.SelectedOption = t.Focused.SelectedOption.
		Foreground(colors.Cyan).
		Bold(true)

	t.Focused.UnselectedOption = t.Focused.UnselectedOption.
		Foreground(colors.Comment)

	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(colors.Background).
		Background(colors.Purple).
		Bold(true)

	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(colors.Foreground).
		Background(colors.CurrentLine)

	return t
}

// assessRiskLevel determines the risk level of a tool operation
func assessRiskLevel(toolUse *core.ToolUse) RiskLevel {
	if toolUse == nil {
		return RiskCaution
	}

	// Check tool name for dangerous operations
	dangerousTools := []string{"bash", "exec", "shell", "delete", "remove", "write"}
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
	safeTools := []string{"read", "list", "get", "search", "find", "view", "show"}
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
		"| sh", "chmod +x", "sudo",
	}

	cmdLower := strings.ToLower(cmd)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(cmdLower, dangerous) {
			return true
		}
	}
	return false
}
