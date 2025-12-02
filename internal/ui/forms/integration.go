// Package forms provides beautiful huh-based forms for the clem TUI.
// ABOUTME: Integration helpers for using huh forms within bubbletea Model
// ABOUTME: Bridges huh's blocking Run() with bubbletea's message-based architecture
package forms

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/clem/internal/core"
)

// ApprovalResultMsg is sent when the approval form completes
type ApprovalResultMsg struct {
	Result ApprovalFormResult
	Error  error
}

// RunToolApprovalForm runs the approval form in a goroutine and returns a tea.Cmd
// This allows huh forms to work within bubbletea's message-based architecture
// Note: huh.Form.Run() is blocking, so we run it in a goroutine and send the result back
func RunToolApprovalForm(toolUse *core.ToolUse) tea.Cmd {
	return func() tea.Msg {
		form := NewToolApprovalForm(toolUse)
		result, err := form.Run()
		return ApprovalResultMsg{
			Result: result,
			Error:  err,
		}
	}
}

// RunToolApprovalFormBatch runs approval forms for multiple tools sequentially
// Each tool gets its own form, one at a time
func RunToolApprovalFormBatch(toolUses []*core.ToolUse) tea.Cmd {
	if len(toolUses) == 0 {
		return nil
	}

	// Run all tools sequentially, collecting results
	return func() tea.Msg {
		results := make([]ApprovalFormResult, 0, len(toolUses))

		for _, toolUse := range toolUses {
			form := NewToolApprovalForm(toolUse)
			result, err := form.Run()

			if err != nil {
				// Return error on first failure
				return ApprovalResultMsg{
					Error: err,
				}
			}

			results = append(results, result)
		}

		// For now, return just the first result
		// Future enhancement: return all results
		if len(results) > 0 {
			return ApprovalResultMsg{
				Result: results[0],
			}
		}

		return ApprovalResultMsg{}
	}
}
