package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/ui/forms"
)

// ToolApprovalOverlay implements the Overlay interface for tool approval.
// Each instance represents approval for a SINGLE tool.
type ToolApprovalOverlay struct {
	model       *Model
	tool        *core.ToolUse // The specific tool this overlay is for
	selectedOpt int           // Which option is selected (0-3)
	remaining   int           // How many tools remain after this one (for display)
}

// NewToolApprovalOverlay creates a new tool approval overlay for a specific tool
func NewToolApprovalOverlay(m *Model, tool *core.ToolUse, remaining int) *ToolApprovalOverlay {
	return &ToolApprovalOverlay{
		model:       m,
		tool:        tool,
		selectedOpt: 0,
		remaining:   remaining,
	}
}

// GetTool returns the tool this overlay is for
func (o *ToolApprovalOverlay) GetTool() *core.ToolUse {
	return o.tool
}

// GetSelectedOption returns the currently selected option (0-3)
func (o *ToolApprovalOverlay) GetSelectedOption() int {
	return o.selectedOpt
}

// GetHeader returns the overlay header
func (o *ToolApprovalOverlay) GetHeader() string {
	if o.remaining > 0 {
		return fmt.Sprintf("Tool Approval Required (%d more after this)", o.remaining)
	}
	return "Tool Approval Required"
}

// GetContent returns the overlay content
func (o *ToolApprovalOverlay) GetContent() string {
	if o.tool == nil {
		return "No tool pending approval"
	}

	var b strings.Builder

	riskLevel := forms.AssessRiskLevel(o.tool)
	coloredRisk := o.model.renderColoredRiskEmoji(riskLevel)
	paramPreview := o.model.getToolParamPreview(o.tool)
	if paramPreview != "" {
		b.WriteString(fmt.Sprintf("%s %s(%s)", coloredRisk, o.tool.Name, paramPreview))
	} else {
		b.WriteString(fmt.Sprintf("%s %s()", coloredRisk, o.tool.Name))
	}
	b.WriteString("\n\n")

	// Options
	options := []string{
		"✓ Approve (run this time)",
		"✗ Deny (skip this time)",
		"✓✓ Always Allow (never ask again)",
		"✗✗ Never Allow (block permanently)",
	}

	selectedStyle := o.model.theme.AutocompleteSelected
	normalStyle := o.model.theme.AutocompleteItem

	for i, opt := range options {
		if i == o.selectedOpt {
			b.WriteString(selectedStyle.Render("▸ " + opt))
		} else {
			b.WriteString(normalStyle.Render("  " + opt))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// GetFooter returns the overlay footer
func (o *ToolApprovalOverlay) GetFooter() string {
	return "↑/↓: navigate • Enter: submit"
}

// GetDesiredHeight returns the desired height for this overlay
func (o *ToolApprovalOverlay) GetDesiredHeight() int {
	// Base height: header + footer + content
	// Content: tool description (1-2 lines) + options (4 lines) + spacing
	// We only render one tool at a time, so no extra height for multiple tools
	return 10
}

// OnPush is called when the overlay is pushed onto the stack
func (o *ToolApprovalOverlay) OnPush(width, height int) {
	// No special initialization needed
}

// OnPop is called when the overlay is popped from the stack
func (o *ToolApprovalOverlay) OnPop() {
	// Each overlay is independent - don't modify global state here.
	// The model will check if more overlays exist and update toolApprovalMode accordingly.
}

// Render returns the complete overlay rendering
func (o *ToolApprovalOverlay) Render(width, height int) string {
	var b strings.Builder

	dropdownWidth := width - 4
	if dropdownWidth < 40 {
		dropdownWidth = 40
	}
	boxStyle := o.model.theme.AutocompleteDropdown.Width(dropdownWidth)

	titleStyle := lipgloss.NewStyle().
		Foreground(o.model.theme.Colors.Purple).
		Bold(true)

	helpStyle := o.model.theme.AutocompleteHelp

	// Header
	if header := o.GetHeader(); header != "" {
		b.WriteString(titleStyle.Render("🛠  " + header))
		b.WriteString("\n")
	}

	// Content
	b.WriteString(o.GetContent())

	// Footer
	b.WriteString("\n")
	if footer := o.GetFooter(); footer != "" {
		b.WriteString(helpStyle.Render(footer))
	}

	return boxStyle.Render(b.String())
}

// HandleKey processes key presses for tool approval
func (o *ToolApprovalOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	// Handle Escape and Ctrl+C to trigger denial
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		return true, nil // Handled - caller will call Cancel() which denies this tool
	}

	// Handle navigation keys
	switch msg.Type {
	case tea.KeyUp:
		if o.selectedOpt > 0 {
			o.selectedOpt--
		}
		return true, nil
	case tea.KeyDown:
		if o.selectedOpt < 3 { // 4 options: 0-3
			o.selectedOpt++
		}
		return true, nil
	case tea.KeyEnter:
		// The actual approval logic is handled in update.go
		// Return false to let it through to the main handler
		return false, nil
	}

	// Modal: capture all other input to prevent leakage
	return true, nil
}

// Cancel dismisses the tool approval and sends denial to API for THIS tool only
func (o *ToolApprovalOverlay) Cancel() tea.Cmd {
	return o.model.DenySpecificTool(o.tool, DenialManual)
}
