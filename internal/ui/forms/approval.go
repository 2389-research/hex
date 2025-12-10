// Package forms provides beautiful huh-based forms for the hex TUI.
// ABOUTME: Huh-based tool approval form for interactive tool permission requests
// ABOUTME: Provides beautiful select-based approval UI with multiple options
package forms

import (
	"fmt"
	"strings"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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

// NOTE: ApprovalResultMsg and RunToolApprovalForm are defined in integration.go
// to provide the async pattern for running huh forms within bubbletea

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
	form      *huh.Form // The actual huh form instance
}

// NewToolApprovalForm creates a new tool approval form
func NewToolApprovalForm(toolUse *core.ToolUse) *ToolApprovalForm {
	return &ToolApprovalForm{
		toolUse:   toolUse,
		riskLevel: assessRiskLevel(toolUse),
		theme:     theme.DraculaTheme(),
	}
}

// BuildForm creates and returns the huh form as a tea.Model for embedding
// Use this instead of Run() to avoid terminal control conflicts with bubbletea
func (f *ToolApprovalForm) BuildForm() *huh.Form {
	// Build the approval options
	options := []huh.Option[ApprovalDecision]{
		huh.NewOption("✓ Approve (run this time)", DecisionApprove),
		huh.NewOption("✗ Deny (skip this time)", DecisionDeny),
		huh.NewOption("✓✓ Always Allow (never ask again)", DecisionAlwaysAllow),
		huh.NewOption("✗✗ Never Allow (block permanently)", DecisionNeverAllow),
	}

	// Build compact description: "⚠ bash: command="echo hello""
	description := f.formatCompactDescription()

	// Create the form with Dracula theme colors
	// Don't call Run() - return the form as a tea.Model for embedding
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[ApprovalDecision]().
				Title("🛠  Tool Approval Required").
				Description(description).
				Options(options...).
				Value(&f.decision).
				Height(8).
				Filtering(false), // Disable filtering so Enter submits immediately
		),
	).WithTheme(f.getDraculaTheme())
}

// GetDecision returns the user's decision after form completion
func (f *ToolApprovalForm) GetDecision() ApprovalFormResult {
	return ApprovalFormResult{
		Decision: f.decision,
		ToolUse:  f.toolUse,
	}
}

// GetRiskLevel returns the assessed risk level for this tool use
func (f *ToolApprovalForm) GetRiskLevel() RiskLevel {
	return f.riskLevel
}

// GetToolUse returns the tool use being approved
func (f *ToolApprovalForm) GetToolUse() *core.ToolUse {
	return f.toolUse
}

// Init implements tea.Model
func (f *ToolApprovalForm) Init() tea.Cmd {
	// Build form if not already built
	if f.form == nil {
		f.form = f.BuildForm()
	}
	return f.form.Init()
}

// Update implements tea.Model
func (f *ToolApprovalForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Forward messages to the huh form
	var cmd tea.Cmd
	updatedModel, cmd := f.form.Update(msg)
	f.form = updatedModel.(*huh.Form)
	return f, cmd
}

// View implements tea.Model
func (f *ToolApprovalForm) View() string {
	if f.form == nil {
		return "Loading form..."
	}
	return f.form.View()
}

// IsComplete checks if the form has been completed
func (f *ToolApprovalForm) IsComplete() bool {
	return f.form != nil && f.form.State == huh.StateCompleted
}

// Run displays the form and returns the user's decision
// Deprecated: Use BuildForm() instead to avoid terminal conflicts
func (f *ToolApprovalForm) Run() (ApprovalFormResult, error) {
	form := f.BuildForm()

	// Run the form
	err := form.Run()
	if err != nil {
		return ApprovalFormResult{}, err
	}

	return f.GetDecision(), nil
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

// formatCompactDescription creates a single-line compact description for the approval prompt
// Format: "⚠ bash("ls -la")" or "✓ read_file("/path/to/file")"
func (f *ToolApprovalForm) formatCompactDescription() string {
	// Defensive nil check
	if f.toolUse == nil {
		return "⚠ unknown_tool()"
	}

	// Get risk emoji
	var riskEmoji string
	switch f.riskLevel {
	case RiskSafe:
		riskEmoji = "✓"
	case RiskCaution:
		riskEmoji = "⚠"
	case RiskDanger:
		riskEmoji = "⚠⚠"
	default:
		riskEmoji = "?"
	}

	// Get the key parameter based on tool type
	var paramValue string
	toolName := f.toolUse.Name

	switch toolName {
	case "bash":
		if cmd, ok := f.toolUse.Input["command"].(string); ok {
			paramValue = formatCompactValue(cmd, 60)
		}
	case "read_file":
		// read_file uses "path" parameter
		if path, ok := f.toolUse.Input["path"].(string); ok {
			paramValue = formatCompactValue(path, 60)
		}
	case "write_file", "edit":
		if path, ok := f.toolUse.Input["file_path"].(string); ok {
			paramValue = formatCompactValue(path, 60)
		}
	case "grep", "glob":
		if pattern, ok := f.toolUse.Input["pattern"].(string); ok {
			paramValue = formatCompactValue(pattern, 50)
		}
	default:
		// For other tools, show first string parameter value
		for _, val := range f.toolUse.Input {
			if str, ok := val.(string); ok && str != "" {
				paramValue = formatCompactValue(str, 50)
				break
			}
		}
	}

	// Format as function call style: ⚠ bash("ls -la")
	if paramValue != "" {
		return fmt.Sprintf("%s %s(%s)", riskEmoji, toolName, paramValue)
	}
	return fmt.Sprintf("%s %s()", riskEmoji, toolName)
}

// formatCompactValue formats a value for compact display
func formatCompactValue(value string, maxLen int) string {
	// Escape newlines and handle multiline
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\t", "\\t")

	// Wrap in quotes for clarity
	if len(value) > maxLen {
		return fmt.Sprintf("%q", value[:maxLen-3]+"...")
	}
	return fmt.Sprintf("%q", value)
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

// AssessRiskLevel determines the risk level of a tool operation (exported for view.go)
func AssessRiskLevel(toolUse *core.ToolUse) RiskLevel {
	return assessRiskLevel(toolUse)
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
