# Multiple Tool Support - Implementation Status

**Date:** 2025-11-29
**Issue:** App "dies" after one tool when Claude requests multiple tools in a single response

---

## Discovery

From `debug.log` analysis, we found Claude sends **multiple tool_use blocks in a single streaming response**:

```
[STREAM_TOOL_START] tool_use detected: id=toolu_01YGwGvVnCH5iD1JhVZM53dx, name=write_file (package.json)
[STREAM_TOOL_COMPLETE] tool_use complete, storing as pending
[STREAM_TOOL_START] tool_use detected: id=toolu_01UFR9reEDeEjoBNCUEurzR5, name=write_file (index.html)
[STREAM_TOOL_COMPLETE] tool_use complete, storing as pending  ← OVERWRITES FIRST!
[STREAM_TOOL_START] tool_use detected: id=toolu_01YKpQSUDgpCoYDNB4FfwJqs, name=write_file (index.js)
[STREAM_TOOL_COMPLETE] tool_use complete, storing as pending  ← OVERWRITES SECOND!
[STREAM_TOOL_START] tool_use detected: id=toolu_01VZomqbXzbbySDjCKQ7xBaY, name=write_file (App.js)
time=2025-11-29T11:55:51.028-06:00 level=INFO msg="Clem shutting down"  ← APP DIES!
```

## Root Cause

We were using a single `pendingToolUse` field that gets overwritten with each new tool:
- Tool 1 arrives → stored in `pendingToolUse`
- Tool 2 arrives → OVERWRITES tool 1
- Tool 3 arrives → OVERWRITES tool 2
- At stream end, we only have Tool 3
- Tools 1 & 2 are lost

---

## Changes Made So Far

### ✅ 1. Model Field Change (`internal/ui/model.go:105`)

```go
// Before:
pendingToolUse  *core.ToolUse  // Tool waiting for approval

// After:
pendingToolUses []*core.ToolUse // Tools waiting for approval (can be multiple)
```

### ✅ 2. Append Tools Instead of Overwriting (`internal/ui/update.go:515-517`)

```go
// Before:
m.pendingToolUse = m.assemblingToolUse

// After:
m.pendingToolUses = append(m.pendingToolUses, m.assemblingToolUse)
fmt.Fprintf(os.Stderr, "[STREAM_TOOL_COMPLETE] tool_use complete, added to pending (total pending: %d): id=%s, name=%s\n",
    len(m.pendingToolUses), m.assemblingToolUse.ID, m.assemblingToolUse.Name)
```

### ✅ 3. Stream Completion with ALL Tools (`internal/ui/update.go:551-593`)

```go
// Check if we have pending tools
if len(m.pendingToolUses) > 0 {
    // Create assistant message with ALL tool_use blocks
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

    // Add assistant message to history
    m.Messages = append(m.Messages, Message{
        Role:         "assistant",
        ContentBlock: blocks,
    })

    m.toolApprovalMode = true
}
```

---

## Remaining Work

### 🔨 4. Update ApproveToolUse Function

**File:** `internal/ui/model.go:447-489`

**Needed:** Execute ALL pending tools sequentially, collect ALL results, send ALL results back in one user message

```go
func (m *Model) ApproveToolUse() tea.Cmd {
    if len(m.pendingToolUses) == 0 || m.toolExecutor == nil {
        m.toolApprovalMode = false
        return nil
    }

    fmt.Fprintf(os.Stderr, "[TOOL_APPROVAL] approving %d tools\n", len(m.pendingToolUses))

    // Capture tools and clear pending
    toolUses := m.pendingToolUses
    m.pendingToolUses = nil
    m.toolApprovalMode = false
    m.executingTool = true

    // Execute all tools in batch
    return m.executeToolsBatch(toolUses)
}
```

### 🔨 5. Create Batch Execution Function

**File:** `internal/ui/model.go` (new function)

```go
// executeToolsBatch executes multiple tools and collects all results
func (m *Model) executeToolsBatch(toolUses []*core.ToolUse) tea.Cmd {
    return func() tea.Msg {
        results := make([]ToolResult, 0, len(toolUses))

        for i, toolUse := range toolUses {
            fmt.Fprintf(os.Stderr, "[BATCH_EXEC] executing tool %d/%d: %s (id=%s)\n",
                i+1, len(toolUses), toolUse.Name, toolUse.ID)

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

        fmt.Fprintf(os.Stderr, "[BATCH_EXEC_DONE] executed %d tools\n", len(results))
        return toolBatchExecutionMsg{results: results}
    }
}

// toolBatchExecutionMsg is sent when a batch of tools finishes executing
type toolBatchExecutionMsg struct {
    results []ToolResult
}
```

### 🔨 6. Handle Batch Execution Results

**File:** `internal/ui/update.go` (add new case in Update function)

```go
case toolBatchExecutionMsg:
    m.executingTool = false

    fmt.Fprintf(os.Stderr, "[BATCH_RESULT_RECEIVED] received results for %d tools\n", len(msg.results))

    // Store all results
    m.toolResults = append(m.toolResults, msg.results...)

    // Display results in UI
    for _, result := range msg.results {
        resultMsg := formatToolResult(result.Result)
        m.AddMessage("tool", resultMsg)
    }

    m.updateViewport()

    // Send ALL tool results back to API in one user message
    return m, m.sendToolResults()
```

### 🔨 7. Update DenyToolUse Function

**File:** `internal/ui/model.go:492-524`

```go
func (m *Model) DenyToolUse() tea.Cmd {
    if len(m.pendingToolUses) == 0 {
        m.toolApprovalMode = false
        return nil
    }

    fmt.Fprintf(os.Stderr, "[TOOL_DENIAL] denying %d tools\n", len(m.pendingToolUses))

    // Create error results for all denied tools
    for _, toolUse := range m.pendingToolUses {
        m.toolResults = append(m.toolResults, ToolResult{
            ToolUseID: toolUse.ID,
            Result: &tools.Result{
                ToolName: toolUse.Name,
                Success:  false,
                Error:    "User denied permission",
            },
        })
        m.AddMessage("tool", "Tool denied: "+toolUse.Name)
    }

    m.pendingToolUses = nil
    m.toolApprovalMode = false
    m.updateViewport()

    // Send all denial results back to API
    return m.sendToolResults()
}
```

### 🔨 8. Update View Functions

**File:** `internal/ui/view.go`

Update all references to `m.pendingToolUse` to use `m.pendingToolUses[0]` or handle multiple tools:

- `renderToolApprovalPrompt()` - Show summary of all tools
- `renderStatus()` - Show "X tools pending"
- `getApprovalPrompt()` - List all tools

---

## Testing Plan

1. Build: `go build ./cmd/clem`
2. Run with debug: `./clem 2>debug.log`
3. Test multi-tool request: "create 3 files: test1.txt with content 'one', test2.txt with 'two', test3.txt with 'three'"
4. Expected output:
   - All 3 tool_use blocks collected
   - Single approval prompt for all 3
   - All 3 tools execute sequentially
   - All 3 results sent back in one user message
   - Claude continues normally

---

## Current Build Status

✅ **COMPILES SUCCESSFULLY!**

All changes have been implemented and the code builds without errors.

---

## Implementation Complete

All 8 steps have been completed:

1. ✅ Model field change (`pendingToolUse` → `pendingToolUses` slice)
2. ✅ Append tools instead of overwriting
3. ✅ Stream completion with ALL tools
4. ✅ Updated ApproveToolUse function (batch execution)
5. ✅ Created executeToolsBatch function
6. ✅ Added toolBatchExecutionMsg type
7. ✅ Added batch execution handler in update.go
8. ✅ Updated DenyToolUse function
9. ✅ Fixed all view.go references (3 functions)
10. ✅ Build verified

---

## Next Steps

1. **Test with multi-tool request** - Run clem and ask it to create 3-4 files
2. **Verify debug.log shows all tools collected** - Check for "total pending: N" messages
3. **Confirm all tools execute** - All files should be created
4. **Verify results sent back together** - Check for "BATCH_RESULTS_SENDING" message

---

## Expected Debug Log Pattern

```
[STREAM_TOOL_START] tool_use detected: id=..., name=write_file
[STREAM_TOOL_COMPLETE] added to pending (total pending: 1): id=..., name=write_file
[STREAM_TOOL_START] tool_use detected: id=..., name=write_file
[STREAM_TOOL_COMPLETE] added to pending (total pending: 2): id=..., name=write_file
[STREAM_TOOL_START] tool_use detected: id=..., name=write_file
[STREAM_TOOL_COMPLETE] added to pending (total pending: 3): id=..., name=write_file
[STREAM_STOP_WITH_TOOLS] creating assistant message with 3 tool_use block(s)
[TOOL_APPROVAL] approving 3 tool(s)
[BATCH_EXEC] executing tool 1/3: write_file (id=...)
[BATCH_EXEC_SUCCESS] tool write_file succeeded
[BATCH_EXEC] executing tool 2/3: write_file (id=...)
[BATCH_EXEC_SUCCESS] tool write_file succeeded
[BATCH_EXEC] executing tool 3/3: write_file (id=...)
[BATCH_EXEC_SUCCESS] tool write_file succeeded
[BATCH_EXEC_DONE] executed 3 tools
[BATCH_RESULT_RECEIVED] received results for 3 tool(s)
[BATCH_RESULTS_SENDING] sending 3 tool results back to API
```

---

**Status:** ✅ Ready for testing
**Complexity:** Medium - Completed successfully
**Time taken:** ~45 minutes
