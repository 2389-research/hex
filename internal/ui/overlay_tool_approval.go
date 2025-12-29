package ui

import (
	"fmt"
	"strings"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/ui/forms"
	tea "github.com/charmbracelet/bubbletea"
)

// ToolApprovalProvider provides content for the tool approval overlay.
// It implements BottomContentProvider, BottomKeyHandler, and BottomCancelHandler.
type ToolApprovalProvider struct {
	model       *Model
	tool        *core.ToolUse
	selectedOpt int
	remaining   int
}

// Header implements BottomContentProvider
func (p *ToolApprovalProvider) Header() string {
	// Show progress hint from queue if available
	if p.model.activeToolQueue != nil {
		hint := p.model.activeToolQueue.ProgressHint()
		if hint != "" {
			return fmt.Sprintf("🛠  Tool Approval %s", hint)
		}
	}
	return "🛠  Tool Approval Required"
}

// Content implements BottomContentProvider
func (p *ToolApprovalProvider) Content() string {
	if p.tool == nil {
		return "No tool pending approval"
	}

	var b strings.Builder

	riskLevel := forms.AssessRiskLevel(p.tool)
	coloredRisk := p.model.renderColoredRiskEmoji(riskLevel)
	paramPreview := p.model.getToolParamPreview(p.tool)
	if paramPreview != "" {
		b.WriteString(fmt.Sprintf("%s %s(%s)", coloredRisk, p.tool.Name, paramPreview))
	} else {
		b.WriteString(fmt.Sprintf("%s %s()", coloredRisk, p.tool.Name))
	}
	b.WriteString("\n\n")

	// Options
	options := []string{
		"✓ Approve (run this time)",
		"✗ Deny (skip this time)",
		"✓✓ Always Allow (never ask again)",
		"✗✗ Never Allow (block permanently)",
	}

	selectedStyle := p.model.theme.AutocompleteSelected
	normalStyle := p.model.theme.AutocompleteItem

	for i, opt := range options {
		if i == p.selectedOpt {
			b.WriteString(selectedStyle.Render("▸ " + opt))
		} else {
			b.WriteString(normalStyle.Render("  " + opt))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// Footer implements BottomContentProvider
func (p *ToolApprovalProvider) Footer() string {
	return "↑/↓: navigate • Enter: submit"
}

// DesiredHeight implements BottomContentProvider
func (p *ToolApprovalProvider) DesiredHeight() int {
	return 10
}

// HandleKey implements BottomKeyHandler
func (p *ToolApprovalProvider) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	// Handle Escape and Ctrl+C to trigger denial via ToolDecisionMsg
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		return true, func() tea.Msg {
			return ToolDecisionMsg{Decision: 1} // deny
		}
	}

	// Handle navigation keys
	switch msg.Type {
	case tea.KeyUp:
		if p.selectedOpt > 0 {
			p.selectedOpt--
		}
		return true, nil
	case tea.KeyDown:
		if p.selectedOpt < 3 { // 4 options: 0-3
			p.selectedOpt++
		}
		return true, nil
	case tea.KeyEnter:
		// Emit decision message for the queue system to handle
		decision := p.selectedOpt
		return true, func() tea.Msg {
			return ToolDecisionMsg{Decision: decision}
		}
	}

	// Modal: capture all other input to prevent leakage
	return true, nil
}

// Cancel implements BottomCancelHandler
func (p *ToolApprovalProvider) Cancel() tea.Cmd {
	return func() tea.Msg {
		return ToolDecisionMsg{Decision: 1} // deny
	}
}

// ToolApprovalOverlay wraps GenericBottomOverlay for tool approval.
// It provides access to the underlying provider for querying state.
type ToolApprovalOverlay struct {
	*GenericBottomOverlay
	provider *ToolApprovalProvider
}

// NewToolApprovalOverlay creates a new tool approval overlay for a specific tool
func NewToolApprovalOverlay(m *Model, tool *core.ToolUse, remaining int) *ToolApprovalOverlay {
	provider := &ToolApprovalProvider{
		model:       m,
		tool:        tool,
		selectedOpt: 0,
		remaining:   remaining,
	}
	return &ToolApprovalOverlay{
		GenericBottomOverlay: NewGenericBottomOverlay(provider, m.theme),
		provider:             provider,
	}
}

// NewToolApprovalOverlayFromQueue creates an overlay from a ToolDisposition in the queue
func NewToolApprovalOverlayFromQueue(m *Model, item *ToolDisposition) *ToolApprovalOverlay {
	remaining := 0
	if m.activeToolQueue != nil {
		remaining = m.activeToolQueue.Remaining()
	}
	return NewToolApprovalOverlay(m, item.Tool, remaining)
}

// GetTool returns the tool this overlay is for
func (o *ToolApprovalOverlay) GetTool() *core.ToolUse {
	return o.provider.tool
}

// GetSelectedOption returns the currently selected option (0-3)
func (o *ToolApprovalOverlay) GetSelectedOption() int {
	return o.provider.selectedOpt
}
