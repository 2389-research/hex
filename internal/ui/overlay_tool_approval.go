package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/2389-research/hex/internal/ui/forms"
)

// ToolApprovalOverlay implements the Overlay interface for tool approval
type ToolApprovalOverlay struct {
	model *Model
}

// NewToolApprovalOverlay creates a new tool approval overlay
func NewToolApprovalOverlay(m *Model) *ToolApprovalOverlay {
	return &ToolApprovalOverlay{model: m}
}

// GetHeader returns the overlay header
func (o *ToolApprovalOverlay) GetHeader() string {
	if len(o.model.pendingToolUses) > 1 {
		// Calculate current position (total - remaining + 1)
		// This shows which tool we're on in the sequence
		return fmt.Sprintf("Tool Approval Required (Tool 1 of %d)", len(o.model.pendingToolUses))
	}
	return "Tool Approval Required"
}

// GetContent returns the overlay content
func (o *ToolApprovalOverlay) GetContent() string {
	if !o.model.toolApprovalMode || len(o.model.pendingToolUses) == 0 {
		return "No tool pending approval"
	}

	var b strings.Builder

	// Always show only the first tool (individual approval)
	tool := o.model.pendingToolUses[0]
	riskLevel := forms.AssessRiskLevel(tool)
	coloredRisk := o.model.renderColoredRiskEmoji(riskLevel)
	paramPreview := o.model.getToolParamPreview(tool)
	if paramPreview != "" {
		b.WriteString(fmt.Sprintf("%s %s(%s)", coloredRisk, tool.Name, paramPreview))
	} else {
		b.WriteString(fmt.Sprintf("%s %s()", coloredRisk, tool.Name))
	}
	b.WriteString("\n\n")

	// Options - always singular since we show one tool at a time
	options := []string{
		"✓ Approve (run this time)",
		"✗ Deny (skip this time)",
		"✓✓ Always Allow (never ask again)",
		"✗✗ Never Allow (block permanently)",
	}

	selectedStyle := o.model.theme.AutocompleteSelected
	normalStyle := o.model.theme.AutocompleteItem

	for i, opt := range options {
		if i == o.model.selectedApprovalOpt {
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
	// Content: tool description (1-5 lines) + options (4 lines) + spacing
	baseHeight := 10
	if len(o.model.pendingToolUses) > 1 {
		baseHeight += len(o.model.pendingToolUses) // Additional lines for multiple tools
	}
	return baseHeight
}

// OnPush is called when the overlay is pushed onto the stack
func (o *ToolApprovalOverlay) OnPush(width, height int) {
	// No special initialization needed
}

// OnPop is called when the overlay is popped from the stack
func (o *ToolApprovalOverlay) OnPop() {
	// Clear state when dismissed
	o.model.toolApprovalMode = false
	o.model.toolApprovalForm = nil
	// NOTE: Don't clear pendingToolUses here! ApproveToolUse/DenyToolUse need it.
	// They will clear it after processing the approval decision.
	o.model.Status = StatusIdle
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
		return true, nil // Handled - caller should call DenyToolUse and Pop
	}

	// Handle navigation keys
	switch msg.Type {
	case tea.KeyUp:
		if o.model.selectedApprovalOpt > 0 {
			o.model.selectedApprovalOpt--
		}
		return true, nil
	case tea.KeyDown:
		if o.model.selectedApprovalOpt < 3 { // 4 options: 0-3
			o.model.selectedApprovalOpt++
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

// Cancel dismisses the tool approval (cleanup only, no API calls)
func (o *ToolApprovalOverlay) Cancel() {
	o.model.toolApprovalMode = false
	o.model.toolApprovalForm = nil
	o.model.pendingToolUses = nil
	o.model.Status = StatusIdle
}
