package ui

import tea "github.com/charmbracelet/bubbletea"

// ToolApprovalOverlay implements the Overlay interface for tool approval
type ToolApprovalOverlay struct {
	model *Model
}

// NewToolApprovalOverlay creates a new tool approval overlay
func NewToolApprovalOverlay(m *Model) *ToolApprovalOverlay {
	return &ToolApprovalOverlay{model: m}
}

// Type returns the overlay type
func (o *ToolApprovalOverlay) Type() OverlayType {
	return OverlayToolApproval
}

// IsActive returns whether tool approval is currently shown
func (o *ToolApprovalOverlay) IsActive() bool {
	return o.model.toolApprovalMode
}

// Render returns the tool approval UI
func (o *ToolApprovalOverlay) Render() string {
	return o.model.renderToolApprovalPromptEnhanced()
}

// HandleKey processes key presses for tool approval
func (o *ToolApprovalOverlay) HandleKey(msg tea.KeyMsg) bool {
	// Tool approval handles its own up/down/enter navigation
	// Return false to let the main Update handle it
	return false
}

// Cancel dismisses the tool approval
func (o *ToolApprovalOverlay) Cancel() {
	o.model.toolApprovalMode = false
	o.model.toolApprovalForm = nil
	o.model.pendingToolUses = nil
	o.model.Status = StatusIdle
}

// Priority returns the precedence level
func (o *ToolApprovalOverlay) Priority() int {
	return 100 // High priority - tool approval is critical
}
