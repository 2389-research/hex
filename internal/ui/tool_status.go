package ui

import "github.com/charmbracelet/lipgloss"

// getToolStatus returns the appropriate icon and style for a tool based on its
// ApprovalType and result success status.
//
// Status mapping:
// - Pending: ◷ (gray)
// - Manual Approve (success): ✓ (green)
// - Manual Approve (error): ✓ (red)
// - Always Allow (success): ✓✓ (green)
// - Always Allow (error): ✓✓ (red)
// - Manual Deny: ✗ (gray)
// - Never Allow: ✗✗ (gray)
func (m *Model) getToolStatus(toolUseID string) (icon string, style lipgloss.Style) {
	// Look up the tool result by ID
	for _, tr := range m.toolResults {
		if tr.ToolUseID == toolUseID {
			// Determine icon based on ApprovalType
			switch tr.ApprovalType {
			case ApprovalManual:
				icon = "✓"
			case ApprovalAlwaysAllow:
				icon = "✓✓"
			case DenialManual:
				icon = "✗"
			case DenialNeverAllow:
				icon = "✗✗"
			default:
				// Pending or unknown
				icon = "◷"
			}

			// Determine color based on approval/denial and success status
			switch tr.ApprovalType {
			case ApprovalManual, ApprovalAlwaysAllow:
				// Approved tools: green for success, red for error
				if tr.Result != nil && tr.Result.Success {
					style = m.theme.Success
				} else {
					style = m.theme.Error
				}
			case DenialManual, DenialNeverAllow:
				// Denied tools: always gray
				style = m.theme.Muted
			default:
				// Pending: gray
				style = m.theme.Muted
			}

			return icon, style
		}
	}

	// No result found = pending
	return "◷", m.theme.Muted
}
