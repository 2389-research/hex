# Tool Use Display Improvements

**Date:** 2025-12-12
**Status:** Ready for implementation

## Overview

Improve tool use visibility in the TUI:
1. Show approval status indicators on tools in message list
2. Ctrl+O shows timeline of all tool calls (fullscreen overlay)
3. Collapsed preview only on most recent tool

## Design

### 1. Status Indicators

| Status | Icon | Color | Meaning |
|--------|------|-------|---------|
| Pending | `◷` | Gray | Tool awaiting approval |
| Manual Approve (success) | `✓` | Green | User approved, ran successfully |
| Manual Approve (error) | `✓` | Red | User approved, ran with error |
| Always Allow (success) | `✓✓` | Green | Auto-approved, ran successfully |
| Always Allow (error) | `✓✓` | Red | Auto-approved, ran with error |
| Manual Deny | `✗` | Gray | User denied this call |
| Never Allow | `✗✗` | Gray | Denied by rule |

**Display:**
```text
✓✓ bash("echo hello")     # green
✓ bash("rm file")         # green
✗ bash("bad thing")       # gray
◷ bash("pending")         # gray
```

### 2. ApprovalType Tracking

Store approval decision at time of approval (not derived from current rules):

```go
type ApprovalType int

const (
    ApprovalPending     ApprovalType = iota
    ApprovalManual
    ApprovalAlwaysAllow
    DenialManual
    DenialNeverAllow
)

type ToolResult struct {
    ToolUseID    string
    Result       *tools.Result
    ApprovalType ApprovalType
}
```

### 3. Ctrl+O Timeline View

Fullscreen overlay showing ALL tool calls in conversation:

```text
┏━━ Tool Timeline ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ Ctrl+O or Esc to close ┓

[12:34:56] ✓✓ bash("echo hello")
└─ hello

[12:35:01] ✓ read_file("main.go")
└─ package main

   import "fmt"

   func main() {
       fmt.Println("hello")
   }

[12:35:15] ✗ bash("rm -rf /")
└─ DENIED

[12:36:00] ◷ bash("make build")
└─ (pending approval)

┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

- Full output for each tool (not truncated)
- Whole view scrollable
- Timestamps from parent message
- Empty state: "No tool calls in this conversation"

### 4. Collapsed Preview

- Only show on single most recent tool in entire conversation
- Not shown if tool is pending
- 3 lines with `│` prefix (existing style)

## Implementation Plan

### Task 1: Add ApprovalType to model.go

**File:** `internal/ui/model.go`

Add before ToolResult struct:
```go
type ApprovalType int

const (
    ApprovalPending     ApprovalType = iota
    ApprovalManual
    ApprovalAlwaysAllow
    DenialManual
    DenialNeverAllow
)
```

Update ToolResult:
```go
type ToolResult struct {
    ToolUseID    string
    Result       *tools.Result
    ApprovalType ApprovalType
}
```

### Task 2: Set ApprovalType when storing results

**File:** `internal/ui/update.go` and `internal/ui/model.go`

Find all places where ToolResult is created and set ApprovalType:

1. `toolResultMsg` handler in update.go (~line 143) - manual approval
2. `DenyToolUse()` in model.go (~line 1017) - manual denial
3. `executeBatchToolsWithProgress()` in model.go (~line 1102, 1112) - always-allow
4. Need to track never-allow denials (currently may not create ToolResult)

### Task 3: Create helper for status icon/color

**File:** `internal/ui/update.go` (or new file `internal/ui/tool_status.go`)

```go
func (m *Model) getToolStatus(toolUseID string) (icon string, style lipgloss.Style) {
    // Check if we have a result
    for _, tr := range m.toolResults {
        if tr.ToolUseID == toolUseID {
            // Determine icon based on ApprovalType
            // Determine color based on Result.Success
            return icon, style
        }
    }
    // No result yet = pending
    return "◷", grayStyle
}
```

### Task 4: Update tool_use rendering in message list

**File:** `internal/ui/update.go` (~line 1623-1665)

Replace:
```go
case "tool_use":
    paramPreview := getToolParamPreview(block.Name, block.Input)
    toolLine := fmt.Sprintf("🛠 %s(%s)", block.Name, paramPreview)
    b.WriteString(m.theme.ToolCall.Render(toolLine))
```

With:
```go
case "tool_use":
    paramPreview := getToolParamPreview(block.Name, block.Input)
    icon, style := m.getToolStatus(block.ID)
    toolLine := fmt.Sprintf("%s %s(%s)", icon, block.Name, paramPreview)
    b.WriteString(style.Render(toolLine))
```

### Task 5: Create ToolTimelineOverlay

**File:** `internal/ui/overlay_tool_timeline.go` (new file)

Based on existing `ToolLogOverlay` pattern:

```go
type ToolTimelineOverlay struct {
    model    *Model  // Reference to get messages and results
    viewport viewport.Model
    width    int
    height   int
}

func (o *ToolTimelineOverlay) IsFullscreen() bool { return true }
func (o *ToolTimelineOverlay) GetDesiredHeight() int { return -1 }
func (o *ToolTimelineOverlay) GetHeader() string { return "Tool Timeline" }

func (o *ToolTimelineOverlay) GetContent() string {
    // Iterate through model.Messages
    // Find tool_use blocks
    // Match with tool_result blocks
    // Format as timeline with status icons and full output
}
```

### Task 6: Update Ctrl+O to use new overlay

**File:** `internal/ui/update.go`

Change Ctrl+O handler to push ToolTimelineOverlay instead of ToolLogOverlay.

### Task 7: Scope collapsed preview to most recent tool

**File:** `internal/ui/view.go` or wherever collapsed preview is rendered

Add logic to:
1. Find the most recent tool_use across all messages
2. Only render collapsed preview for that tool
3. Skip if tool is pending (no result yet)

### Task 8: Delete or repurpose old ToolLogOverlay

**File:** `internal/ui/overlay_tool_log.go`

Either:
- Delete if no longer needed
- Keep for other purposes
- Rename/refactor into ToolTimelineOverlay

## Files to Modify

1. `internal/ui/model.go` - ApprovalType enum, ToolResult update
2. `internal/ui/update.go` - Set ApprovalType, status rendering, Ctrl+O handler
3. `internal/ui/overlay_tool_timeline.go` - New file
4. `internal/ui/view.go` - Collapsed preview scoping
5. `internal/ui/overlay_tool_log.go` - Delete or repurpose

## Testing

1. Manual approval flow - verify `✓` appears green
2. Always-allow flow - verify `✓✓` appears green
3. Manual deny - verify `✗` appears gray
4. Never-allow deny - verify `✗✗` appears gray
5. Error case - verify icon is red
6. Ctrl+O shows full timeline
7. Collapsed preview only on most recent tool
8. Empty conversation shows "No tool calls"
