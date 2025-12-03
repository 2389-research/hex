# Bug Fix: Support Multiple Tool Calls in Single Response

**Date:** 2025-11-29
**Status:** 🔨 In Progress
**Severity:** Critical (blocks multi-tool workflows)

---

## Problem

From the debug.log, we discovered that Claude can request **multiple tools in a single streaming response**:

```
[STREAM_TOOL_START] tool_use detected: id=toolu_01YGwGvVnCH5iD1JhVZM53dx, name=write_file
[STREAM_TOOL_COMPLETE] tool_use complete, storing as pending
[STREAM_TOOL_START] tool_use detected: id=toolu_01UFR9reEDeEjoBNCUEurzR5, name=write_file
[STREAM_TOOL_COMPLETE] tool_use complete, storing as pending  ← OVERWRITES FIRST!
[STREAM_TOOL_START] tool_use detected: id=toolu_01YKpQSUDgpCoYDNB4FfwJqs, name=write_file
[STREAM_TOOL_COMPLETE] tool_use complete, storing as pending  ← OVERWRITES SECOND!
```

**Current behavior:**
- We store only ONE `pendingToolUse` at a time
- Each new tool_use block overwrites the previous one
- At stream completion, we only have the LAST tool
- All earlier tools are lost
- The app appears to "die" because those tools never execute

---

## Root Cause

**Location:** `internal/ui/update.go:516`

```go
// Handle content_block_stop - tool parameters are complete
if chunk.Type == "content_block_stop" && m.assemblingToolUse != nil {
    // ...
    m.pendingToolUse = m.assemblingToolUse  // ← OVERWRITES previous tool!
    m.assemblingToolUse = nil
    // ...
}
```

The code assumes only ONE tool per streaming response, but the API can send multiple.

---

## Solution Strategy

According to Anthropic's Messages API documentation, when multiple tools are requested:

1. The assistant message should contain **all** tool_use blocks
2. We should execute **all** tools
3. We should send back **all** tool_result blocks in a **single** user message

Two implementation options:

### Option A: Collect All Tools, Approve All Together

1. Change `pendingToolUse *core.ToolUse` → `pendingToolUses []*core.ToolUse`
2. Append each completed tool to the slice instead of overwriting
3. At stream completion, show approval dialog listing ALL tools
4. When approved, execute ALL tools sequentially
5. Send ALL results back in one user message

**Pros:**
- User approves once for the whole batch
- Matches API expectations (all results in one user message)

**Cons:**
- Need to update approval UI to show multiple tools
- More complex approval logic

### Option B: Approve and Execute One at a Time

1. Collect all tools during streaming
2. Show approval for first tool only
3. After first tool executes and result is sent, show approval for second tool
4. Continue until all tools are executed

**Pros:**
- Simpler approval UI (one tool at a time)
- User has fine-grained control

**Cons:**
- Requires multiple API round-trips
- Doesn't match the single-message-with-all-results pattern
- More complex state management

### Recommended: Option A

This matches the API's expected flow and is conceptually cleaner, even though it requires updating the approval UI.

---

## Implementation Plan

### Step 1: Change pendingToolUse to pendingToolUses slice

**File:** `internal/ui/model.go:105`

```go
// Before:
pendingToolUse  *core.ToolUse  // Tool waiting for approval/execution

// After:
pendingToolUses []*core.ToolUse // Tools waiting for approval/execution (can be multiple)
```

### Step 2: Append tools instead of overwriting

**File:** `internal/ui/update.go:516`

```go
// Before:
m.pendingToolUse = m.assemblingToolUse

// After:
m.pendingToolUses = append(m.pendingToolUses, m.assemblingToolUse)
fmt.Fprintf(os.Stderr, "[STREAM_TOOL_COMPLETE] added to pending tools (total pending: %d)\n", len(m.pendingToolUses))
```

### Step 3: Update stream completion to handle ALL tools

**File:** `internal/ui/update.go:553-588`

```go
// Create assistant message with ALL tool_use blocks
if len(m.pendingToolUses) > 0 {
    fmt.Fprintf(os.Stderr, "[STREAM_STOP_WITH_TOOLS] creating assistant message with %d tool_use blocks\n", len(m.pendingToolUses))

    blocks := []core.ContentBlock{}

    // Add text block if present
    if m.StreamingText != "" {
        blocks = append(blocks, core.NewTextBlock(m.StreamingText))
    }

    // Add ALL tool_use blocks
    for _, toolUse := range m.pendingToolUses {
        blocks = append(blocks, core.ContentBlock{
            Type:  "tool_use",
            ID:    toolUse.ID,
            Name:  toolUse.Name,
            Input: toolUse.Input,
        })
    }

    // Add assistant message with all content blocks
    assistantMsg := Message{
        Role:         "assistant",
        ContentBlock: blocks,
    }
    m.Messages = append(m.Messages, assistantMsg)
    m.StreamingText = ""

    // Show tool approval dialog (for all tools)
    m.toolApprovalMode = true
}
```

### Step 4: Update approval logic to handle multiple tools

**File:** `internal/ui/model.go:447-489`

```go
// ApproveToolUse executes ALL pending tools
func (m *Model) ApproveToolUse() tea.Cmd {
    if len(m.pendingToolUses) == 0 || m.toolExecutor == nil {
        m.toolApprovalMode = false
        return nil
    }

    fmt.Fprintf(os.Stderr, "[TOOL_APPROVAL] approving %d tools\n", len(m.pendingToolUses))

    // Execute all tools sequentially
    toolUses := m.pendingToolUses
    m.pendingToolUses = nil
    m.toolApprovalMode = false

    // Create a batch command that executes all tools
    return m.executeToolsBatch(toolUses)
}
```

### Step 5: Create batch execution function

**File:** `internal/ui/model.go` (new function)

```go
// executeToolsBatch executes multiple tools and collects all results
func (m *Model) executeToolsBatch(toolUses []*core.ToolUse) tea.Cmd {
    return func() tea.Msg {
        results := make([]ToolResult, 0, len(toolUses))

        for _, toolUse := range toolUses {
            fmt.Fprintf(os.Stderr, "[BATCH_EXEC] executing tool %s (id=%s)\n", toolUse.Name, toolUse.ID)

            ctx := context.Background()
            result, err := m.toolExecutor.Execute(ctx, toolUse.Name, toolUse.Input)

            if err != nil {
                fmt.Fprintf(os.Stderr, "[BATCH_EXEC_ERROR] tool %s failed: %v\n", toolUse.Name, err)
                results = append(results, ToolResult{
                    ToolUseID: toolUse.ID,
                    Result: &tools.Result{
                        ToolName: toolUse.Name,
                        Success:  false,
                        Error:    err.Error(),
                    },
                })
            } else {
                fmt.Fprintf(os.Stderr, "[BATCH_EXEC_SUCCESS] tool %s succeeded\n", toolUse.Name)
                results = append(results, ToolResult{
                    ToolUseID: toolUse.ID,
                    Result:    result,
                })
            }
        }

        return toolBatchExecutionMsg{results: results}
    }
}

// toolBatchExecutionMsg is sent when a batch of tools finishes executing
type toolBatchExecutionMsg struct {
    results []ToolResult
}
```

### Step 6: Update approval UI to show multiple tools

**File:** `internal/ui/view.go:262-290`

```go
if !m.toolApprovalMode || len(m.pendingToolUses) == 0 {
    return ""
}

var prompt strings.Builder
prompt.WriteString("\n")
prompt.WriteString(styles.ToolApprovalStyle.Render("=== Tool Approval Required ==="))
prompt.WriteString("\n\n")

if len(m.pendingToolUses) == 1 {
    tool := m.pendingToolUses[0]
    prompt.WriteString(fmt.Sprintf("Tool: %s\n", tool.Name))
    // ... rest of single tool display
} else {
    prompt.WriteString(fmt.Sprintf("The assistant wants to execute %d tools:\n\n", len(m.pendingToolUses)))
    for i, tool := range m.pendingToolUses {
        prompt.WriteString(fmt.Sprintf("%d. %s", i+1, tool.Name))
        if len(tool.Input) > 0 {
            prompt.WriteString(" (")
            // Show brief summary of inputs
            prompt.WriteString("...)")
        }
        prompt.WriteString("\n")
    }
}

prompt.WriteString("\nApprove? (y/n): ")
```

---

## Files to Modify

1. `internal/ui/model.go` - Change field, update approval/denial functions
2. `internal/ui/update.go` - Append tools, handle batch results, update stream completion
3. `internal/ui/view.go` - Update approval UI to show multiple tools

---

## Testing

After changes:

1. Build: `go build ./cmd/hex`
2. Run with debug: `./hex 2>debug.log`
3. Test multi-tool request: "create 3 files: test1.txt, test2.txt, test3.txt"
4. Expected in debug.log:
   ```
   [STREAM_TOOL_START] tool_use detected: id=..., name=write_file
   [STREAM_TOOL_COMPLETE] added to pending tools (total pending: 1)
   [STREAM_TOOL_START] tool_use detected: id=..., name=write_file
   [STREAM_TOOL_COMPLETE] added to pending tools (total pending: 2)
   [STREAM_TOOL_START] tool_use detected: id=..., name=write_file
   [STREAM_TOOL_COMPLETE] added to pending tools (total pending: 3)
   [STREAM_STOP_WITH_TOOLS] creating assistant message with 3 tool_use blocks
   [TOOL_APPROVAL] approving 3 tools
   [BATCH_EXEC] executing tool write_file (id=...)
   [BATCH_EXEC_SUCCESS] tool write_file succeeded
   [BATCH_EXEC] executing tool write_file (id=...)
   [BATCH_EXEC_SUCCESS] tool write_file succeeded
   [BATCH_EXEC] executing tool write_file (id=...)
   [BATCH_EXEC_SUCCESS] tool write_file succeeded
   ```

---

**Status:** Ready to implement
**Estimated effort:** ~30 minutes
