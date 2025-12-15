# Tool Queue Refactor Design

**Date:** 2025-12-15
**Status:** Draft

## Problem

The current tool approval system is complex and buggy:
- Pushes ALL pending tools as overlays upfront
- Multiple code paths handle tool state (handleMessageStop, toolBatchExecutionMsg handler, handleApprovalResult)
- State scattered across 7+ fields on Model
- Duplicate overlays get pushed, stale overlays persist
- Auto-approval rules interact poorly with overlay stack

## Design Goals

1. Process tools sequentially (one at a time)
2. Pre-classify tools to know full picture before prompting
3. Single overlay pushed only when needed
4. Clean ephemeral state that doesn't leak between batches
5. Compact progress indicator showing what's happening

## Core Data Structures

```go
type ToolAction int

const (
    ActionNeedsApproval ToolAction = iota
    ActionAutoApprove
    ActionAutoDeny
)

type ToolOutcome int

const (
    OutcomePending ToolOutcome = iota
    OutcomeApproved
    OutcomeDenied
)

type ToolDisposition struct {
    Tool    *core.ToolUse
    Action  ToolAction   // initial classification
    Outcome ToolOutcome  // filled in after processed
    Reason  string       // "bypass mode", "always_allow rule", "no rule", etc.
}

type ToolQueue struct {
    items   []ToolDisposition
    current int
    results []ToolResult
}
```

## Classification

Tools are classified when the queue is created:

```go
func NewToolQueue(tools []*core.ToolUse, rules *ApprovalRules, permissionMode string) *ToolQueue {
    items := make([]ToolDisposition, len(tools))
    for i, tool := range tools {
        items[i] = classifyTool(tool, rules, permissionMode)
    }
    return &ToolQueue{items: items, current: 0}
}

func classifyTool(tool *core.ToolUse, rules *ApprovalRules, permissionMode string) ToolDisposition {
    // Permission mode bypass? Everything auto-approves
    if permissionMode == "bypass" {
        return ToolDisposition{Tool: tool, Action: ActionAutoApprove, Reason: "bypass mode"}
    }

    // Check rules
    if rules != nil {
        switch rules.Check(tool.Name) {
        case RuleAlwaysAllow:
            return ToolDisposition{Tool: tool, Action: ActionAutoApprove, Reason: "always_allow rule"}
        case RuleNeverAllow:
            return ToolDisposition{Tool: tool, Action: ActionAutoDeny, Reason: "never_allow rule"}
        }
    }

    return ToolDisposition{Tool: tool, Action: ActionNeedsApproval, Reason: "no rule"}
}
```

## Processing Flow

Single function drives everything:

```go
func (m *Model) ProcessNextTool() tea.Cmd {
    item := m.activeToolQueue.Current()
    if item == nil {
        // Queue done - send results to API
        return m.finalizeToolQueue()
    }

    switch item.Action {
    case ActionAutoApprove:
        item.Outcome = OutcomeApproved
        m.activeToolQueue.Advance()
        return m.executeToolThenProcessNext(item.Tool, ApprovalAlwaysAllow)

    case ActionAutoDeny:
        item.Outcome = OutcomeDenied
        m.activeToolQueue.AddResult(denialResult(item.Tool))
        m.activeToolQueue.Advance()
        return m.ProcessNextTool()  // immediately process next

    case ActionNeedsApproval:
        // Push single overlay, wait for user
        overlay := NewToolApprovalOverlay(m, item)
        m.overlayManager.Push(overlay, m.Width, m.Height)
        return nil  // wait for user input
    }
    return nil
}
```

## Progress Hint

Compact visual showing queue status:

```text
[.✓✗?]  = "you are here, then approve, deny, ask"
[✓✓✗.]  = "approved, approved, denied, you are here"
```

Symbols:
- `.` = current position (with dim background)
- `?` = needs approval (future)
- `✓` = approved (past) or will auto-approve (future)
- `✗` = denied (past) or will auto-deny (future)

```go
func (q *ToolQueue) ProgressHint() string {
    var hint strings.Builder
    hint.WriteString("[")
    for i, item := range q.items {
        if i == q.current {
            hint.WriteString(".")  // current position
        } else if i < q.current {
            // Past - show outcome
            switch item.Outcome {
            case OutcomeApproved:
                hint.WriteString("✓")
            case OutcomeDenied:
                hint.WriteString("✗")
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
```

## Overlay Integration

Overlay shows single tool, emits decision message:

```go
type ToolApprovalOverlay struct {
    model       *Model
    disposition *ToolDisposition
    selectedOpt int  // 0=approve, 1=deny, 2=always, 3=never
}

type ToolDecisionMsg struct {
    Decision int  // which option was selected
}

func (o *ToolApprovalOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
    switch msg.Type {
    case tea.KeyUp:
        if o.selectedOpt > 0 {
            o.selectedOpt--
        }
        return true, nil
    case tea.KeyDown:
        if o.selectedOpt < 3 {
            o.selectedOpt++
        }
        return true, nil
    case tea.KeyEnter:
        return true, func() tea.Msg {
            return ToolDecisionMsg{Decision: o.selectedOpt}
        }
    case tea.KeyEsc:
        return true, func() tea.Msg {
            return ToolDecisionMsg{Decision: 1}  // deny
        }
    }
    return true, nil  // capture all input (modal)
}

func (o *ToolApprovalOverlay) GetHeader() string {
    hint := o.model.activeToolQueue.ProgressHint()
    return fmt.Sprintf("Tool Approval %s", hint)
}
```

## Decision Handling

In update.go:

```go
case ToolDecisionMsg:
    return m.handleToolDecision(msg.Decision)

func (m *Model) handleToolDecision(decision int) tea.Cmd {
    item := m.activeToolQueue.Current()
    m.overlayManager.Pop()  // remove approval overlay

    switch decision {
    case 0: // Approve once
        item.Outcome = OutcomeApproved
        m.activeToolQueue.Advance()
        return m.executeToolThenProcessNext(item.Tool, ApprovalManual)

    case 1: // Deny once
        item.Outcome = OutcomeDenied
        m.activeToolQueue.AddResult(denialResult(item.Tool))
        m.activeToolQueue.Advance()
        return m.ProcessNextTool()

    case 2: // Always allow
        m.approvalRules.SetAlwaysAllow(item.Tool.Name)
        item.Outcome = OutcomeApproved
        m.activeToolQueue.Advance()
        return m.executeToolThenProcessNext(item.Tool, ApprovalAlwaysAllow)

    case 3: // Never allow
        m.approvalRules.SetNeverAllow(item.Tool.Name)
        item.Outcome = OutcomeDenied
        m.activeToolQueue.AddResult(denialResult(item.Tool))
        m.activeToolQueue.Advance()
        return m.ProcessNextTool()
    }
    return nil
}
```

## Model State Changes

**Remove:**
```go
pendingToolUses     []*core.ToolUse
toolResults         []ToolResult
toolApprovalMode    bool
toolApprovalForm    *huh.Form
approvalPrompt      *ApprovalPrompt
executingTool       bool
executingToolUses   []*core.ToolUse
```

**Add:**
```go
streamingTools      []*core.ToolUse  // accumulated during stream
activeToolQueue     *ToolQueue       // nil when not processing
toolResultHistory   []ToolResult     // persists for UI display
```

## Integration Points

### Stream Handling

```go
// During streaming - accumulate tools
func (m *Model) handleToolUseComplete() {
    m.streamingTools = append(m.streamingTools, m.assemblingToolUse)
    m.assemblingToolUse = nil
}

// When stream ends
func (m *Model) handleMessageStop() (tea.Model, tea.Cmd) {
    if len(m.streamingTools) > 0 {
        m.activeToolQueue = NewToolQueue(m.streamingTools, m.approvalRules, m.permissionMode)
        m.streamingTools = nil
        return m, m.ProcessNextTool()
    }
    // No tools - commit text
    m.CommitStreamingText()
    return m, nil
}
```

### Tool Execution Complete

```go
func (m *Model) handleToolExecutionComplete(result ToolResult) tea.Cmd {
    // Update queue
    m.activeToolQueue.AddResult(result)

    // Update persistent display
    m.toolResultHistory = append(m.toolResultHistory, result)

    // Continue processing
    return m.ProcessNextTool()
}
```

### Queue Completion

```go
func (m *Model) finalizeToolQueue() tea.Cmd {
    results := m.activeToolQueue.results
    m.activeToolQueue = nil  // clear ephemeral state

    return m.sendToolResultsToAPI(results)
}
```

## Summary

- **ToolQueue**: Ephemeral, created when stream ends with tools
- **ToolDisposition**: Pre-classifies each tool (auto-approve/deny/needs-approval)
- **ProcessNextTool()**: Single loop processes queue sequentially
- **Single overlay**: Pushed only when needed, emits `ToolDecisionMsg`
- **Progress hint**: `[✓✗.]` shows status at a glance
- **Separate display**: `toolResultHistory` persists for UI after queue gone

This design consolidates 7+ scattered fields into 3 clean fields, eliminates duplicate overlay bugs, and provides a predictable sequential flow with clear user feedback.
