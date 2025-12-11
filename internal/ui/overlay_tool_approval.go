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

// GetHeader returns the header content
func (o *ToolApprovalOverlay) GetHeader() string {
	return ""
}

// GetContent returns the main content
func (o *ToolApprovalOverlay) GetContent() string {
	return o.Render(0, 0)
}

// GetFooter returns the footer content
func (o *ToolApprovalOverlay) GetFooter() string {
	return ""
}

// GetDesiredHeight returns the desired height for this overlay
func (o *ToolApprovalOverlay) GetDesiredHeight() int {
	return 5
}

// OnPush is called when the overlay is pushed onto the stack
func (o *ToolApprovalOverlay) OnPush(width, height int) {}

// OnPop is called when the overlay is popped from the stack
func (o *ToolApprovalOverlay) OnPop() {}

// Render returns the tool approval UI
func (o *ToolApprovalOverlay) Render(width, height int) string {
	return o.model.renderToolApprovalPromptEnhanced()
}

// HandleKey processes key presses for tool approval
func (o *ToolApprovalOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	// Handle Escape and Ctrl+C to trigger denial
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		return true, nil // Handled - caller should call DenyToolUse and Pop
	}
	// Other tool approval navigation (up/down/enter) is handled in main Update
	// Modal: capture all input, even if not specifically handled
	return true, nil
}

// Cancel dismisses the tool approval (cleanup only, no API calls)
func (o *ToolApprovalOverlay) Cancel() {
	o.model.toolApprovalMode = false
	o.model.toolApprovalForm = nil
	o.model.pendingToolUses = nil
	o.model.Status = StatusIdle
}
