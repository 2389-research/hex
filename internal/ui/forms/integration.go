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

// RunToolApprovalFormBatch runs approval forms for multiple tools sequentially.
// Each tool gets its own form, one at a time.
//
// IMPORTANT: Currently, this uses the FIRST tool's decision for ALL tools in the batch.
// This means if the user approves tool 1, ALL tools are approved.
// If the user denies tool 1, ALL tools are denied.
// Individual per-tool decisions from subsequent forms are collected but not yet used.
//
// This is a known limitation - see the handler in model.go which calls ApproveToolUse()
// or DenyToolUse() which operate on ALL pending tools at once.
//
// Future enhancement: Support per-tool decisions with a more sophisticated result type.
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

			// LIMITATION: If user denies this tool, we should probably stop showing
			// subsequent forms since the entire batch will be denied anyway.
			// For now, we show all forms to collect feedback, but only use the first decision.
		}

		// Return the first result which will be used for ALL tools in the batch
		// See handleApprovalResult() in model.go for how this is processed
		if len(results) > 0 {
			return ApprovalResultMsg{
				Result: results[0],
			}
		}

		return ApprovalResultMsg{}
	}
}
