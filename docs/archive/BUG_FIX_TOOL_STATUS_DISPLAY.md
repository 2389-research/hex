# Bug Fix: Tool Status Display Issue

**Date:** 2025-11-29
**Status:** ✅ FIXED
**Severity:** Medium (UI display bug)

---

## Problem Discovered

During fresh-eyes code review, I found a bug in `renderToolStatus()` (view.go:315):

The function was checking `m.pendingToolUses` to determine what tools are executing, but `pendingToolUses` is **cleared** before execution starts!

### The Bug

```go
// In ApproveToolUse() - model.go:453
toolUses := m.pendingToolUses
m.pendingToolUses = nil  // ← CLEARED HERE!
m.executingTool = true

// Later in renderToolStatus() - view.go:315
if len(m.pendingToolUses) > 0 {  // ← ALWAYS EMPTY!
    // This code never runs during execution
}
```

**Result:** The tool status display would always show "⏳ Executing: unknown..." instead of the actual tool name(s).

---

## Root Cause

The execution flow is:

1. Tools collected in `pendingToolUses` during streaming
2. User approves → `ApproveToolUse()` called
3. `ApproveToolUse()` copies tools to local `toolUses` variable
4. `ApproveToolUse()` **clears** `m.pendingToolUses`
5. Tools execute in background
6. During execution, `renderToolStatus()` checks `m.pendingToolUses` → **EMPTY!**

---

## Solution

Added a new field `executingToolUses` to track which tools are currently being executed:

### Change 1: Add Field (model.go:106)

```go
pendingToolUses    []*core.ToolUse // Tools waiting for approval
executingToolUses  []*core.ToolUse // Tools currently being executed (for display) ← NEW
```

### Change 2: Save Tools Before Execution (model.go:455)

```go
toolUses := m.pendingToolUses
m.pendingToolUses = nil
m.executingToolUses = toolUses // ← Save for status display
m.toolApprovalMode = false
m.executingTool = true
```

### Change 3: Clear After Execution (update.go:86)

```go
case toolBatchExecutionMsg:
    m.executingTool = false
    m.executingToolUses = nil // ← Clear executing tools
    // ... rest of handler
```

### Change 4: Update Status Display (view.go:315)

```go
// Before:
if len(m.pendingToolUses) > 0 {  // Wrong - always empty!

// After:
if len(m.executingToolUses) > 0 {  // Correct!
    if len(m.executingToolUses) == 1 {
        toolName = m.executingToolUses[0].Name
    } else {
        toolName = fmt.Sprintf("%d tools", len(m.executingToolUses))
    }
}
```

---

## Files Modified

1. `internal/ui/model.go` - Added `executingToolUses` field, save tools before execution
2. `internal/ui/update.go` - Clear `executingToolUses` after batch completes
3. `internal/ui/view.go` - Use `executingToolUses` instead of `pendingToolUses`

---

## Expected Behavior After Fix

### Single Tool
```
⏳ Executing: write_file...
```

### Multiple Tools
```
⏳ Executing: 3 tools...
```

### No Longer Shows
```
⏳ Executing: unknown...  ← FIXED!
```

---

## Testing

Build verified:
```bash
$ go build ./cmd/hex
# Success!
```

Manual testing needed:
- Run hex
- Request multiple tools
- Verify status message shows correct tool name(s) during execution

---

**Impact:** Medium - UI display issue, doesn't affect functionality but improves UX
**Fix Type:** State management - proper tracking of executing tools
**Lines Changed:** ~5 lines across 3 files
