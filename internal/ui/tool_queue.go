package ui

import (
	"strings"

	"github.com/2389-research/hex/internal/approval"
	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/tools"
)

// ToolAction represents how a tool should be handled
type ToolAction int

const (
	ActionNeedsApproval ToolAction = iota
	ActionAutoApprove
	ActionAutoDeny
)

// ToolOutcome represents what happened to a tool after processing
type ToolOutcome int

const (
	OutcomePending ToolOutcome = iota
	OutcomeApproved
	OutcomeDenied
)

// ToolDisposition represents a tool and how it should be/was handled
type ToolDisposition struct {
	Tool    *core.ToolUse
	Action  ToolAction  // initial classification
	Outcome ToolOutcome // filled in after processed
	Reason  string      // "bypass mode", "always_allow rule", "no rule", etc.
}

// ToolQueue manages sequential processing of tool approvals
type ToolQueue struct {
	items   []ToolDisposition
	current int
	results []ToolResult
}

// NewToolQueue creates a queue from pending tools, classifying each upfront
func NewToolQueue(tools []*core.ToolUse, rules *approval.Rules, permissionMode string) *ToolQueue {
	items := make([]ToolDisposition, len(tools))
	for i, tool := range tools {
		items[i] = classifyTool(tool, rules, permissionMode)
	}
	return &ToolQueue{items: items, current: 0}
}

// classifyTool determines how a tool should be handled based on rules and permission mode
func classifyTool(tool *core.ToolUse, rules *approval.Rules, permissionMode string) ToolDisposition {
	// Permission mode auto? Everything auto-approves
	if permissionMode == "auto" {
		return ToolDisposition{Tool: tool, Action: ActionAutoApprove, Reason: "auto mode"}
	}

	// Check rules
	if rules != nil {
		switch rules.Check(tool.Name) {
		case approval.RuleAlwaysAllow:
			return ToolDisposition{Tool: tool, Action: ActionAutoApprove, Reason: "always_allow rule"}
		case approval.RuleNeverAllow:
			return ToolDisposition{Tool: tool, Action: ActionAutoDeny, Reason: "never_allow rule"}
		}
	}

	return ToolDisposition{Tool: tool, Action: ActionNeedsApproval, Reason: "no rule"}
}

// Current returns the current tool disposition, or nil if done
func (q *ToolQueue) Current() *ToolDisposition {
	if q.current >= len(q.items) {
		return nil
	}
	return &q.items[q.current]
}

// Advance moves to the next tool in the queue
func (q *ToolQueue) Advance() {
	q.current++
}

// AddResult adds a tool result to the accumulated results
func (q *ToolQueue) AddResult(result ToolResult) {
	q.results = append(q.results, result)
}

// Results returns all accumulated results
func (q *ToolQueue) Results() []ToolResult {
	return q.results
}

// IsDone returns true if all tools have been processed
func (q *ToolQueue) IsDone() bool {
	return q.current >= len(q.items)
}

// Len returns the total number of tools in the queue
func (q *ToolQueue) Len() int {
	return len(q.items)
}

// Remaining returns the number of tools after the current one
func (q *ToolQueue) Remaining() int {
	remaining := len(q.items) - q.current - 1
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ProgressHint returns a compact string showing queue status
// Example: "[●✓✗?]" = "you are here, then approve, deny, ask"
// Example: "[✓✓✗●]" = "approved, approved, denied, you are here"
// The current position "●" gets a subtle background highlight
func (q *ToolQueue) ProgressHint() string {
	if len(q.items) <= 1 {
		return "" // no hint needed for single tool
	}

	// Highlight current position with background only (inherits foreground from context)
	// Use raw ANSI escape to avoid lipgloss reset breaking subsequent text
	bgStart := "\x1b[48;2;61;90;128m" // #3d5a80 lighter blue background
	bgEnd := "\x1b[49m"               // reset background only

	var hint strings.Builder
	hint.WriteString("[")
	for i, item := range q.items {
		if i == q.current {
			hint.WriteString(bgStart + "●" + bgEnd)
		} else if i < q.current {
			// Past - show outcome
			switch item.Outcome {
			case OutcomeApproved:
				hint.WriteString("✓")
			case OutcomeDenied:
				hint.WriteString("✗")
			default:
				hint.WriteString("?")
			}
		} else {
			// Future - show disposition
			switch item.Action {
			case ActionAutoApprove:
				hint.WriteString("✓")
			case ActionAutoDeny:
				hint.WriteString("✗")
			case ActionNeedsApproval:
				hint.WriteString("?")
			}
		}
	}
	hint.WriteString("]")
	return hint.String()
}

// CountByAction returns counts of each action type in remaining (unprocessed) items
func (q *ToolQueue) CountByAction() (needsApproval, autoApprove, autoDeny int) {
	for i := q.current; i < len(q.items); i++ {
		switch q.items[i].Action {
		case ActionNeedsApproval:
			needsApproval++
		case ActionAutoApprove:
			autoApprove++
		case ActionAutoDeny:
			autoDeny++
		}
	}
	return
}

// denialResult creates a ToolResult for a denied tool
func denialResult(tool *core.ToolUse, approvalType ApprovalType) ToolResult {
	return ToolResult{
		ToolUseID: tool.ID,
		Result: &tools.Result{
			ToolName: tool.Name,
			Success:  false,
			Error:    "User denied permission",
		},
		ApprovalType: approvalType,
	}
}
