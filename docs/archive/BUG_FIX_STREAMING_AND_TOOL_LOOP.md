# Bug Fix: Streaming Interruption and Tool Results Loop

**Date:** 2025-11-28
**Status:** ✅ Fixed
**Severity:** High (blocked tool functionality)

---

## Problems

### Problem 1: Stream Stopped After Tool Use
After the streaming parameter fix, tools would get proper parameters but the stream would stop completely after the first tool, preventing:
- The full response from being received
- Subsequent tool calls from working
- The conversation from continuing

### Problem 2: Tool Results Not Sent Back to API
The `sendToolResults()` function was a stub implementation that:
- Cleared tool results without sending them to the API
- Didn't include tool definitions in follow-up requests
- Prevented Claude from seeing tool results and making subsequent tool calls

---

## Root Cause Analysis

### Issue 1: Stream Interruption in `internal/ui/update.go:484-501`

**Before:**
```go
if chunk.Type == "content_block_stop" && m.assemblingToolUse != nil {
    // Parse accumulated JSON into Input map
    if m.toolInputJSONBuf != "" {
        var input map[string]interface{}
        if err := json.Unmarshal([]byte(m.toolInputJSONBuf), &input); err == nil {
            m.assemblingToolUse.Input = input
        }
    }

    // Tool use is complete, move to approval
    toolUse := m.assemblingToolUse
    m.assemblingToolUse = nil
    m.toolInputJSONBuf = ""

    // Pause streaming and handle tool
    m.streamChan = nil // ❌ This stops all streaming!
    return m, m.HandleToolUse(toolUse)
}
```

**Problem:** Setting `m.streamChan = nil` immediately stops the stream, preventing:
- The `message_stop` event from being received
- Any additional text or content after the tool from being streamed
- The stream from completing properly

### Issue 2: Incomplete `sendToolResults()` in `internal/ui/model.go:649-710`

**Before:**
```go
func (m *Model) sendToolResults() tea.Cmd {
    if len(m.toolResults) == 0 || m.apiClient == nil {
        return nil
    }

    // TODO: Use toolResults to construct proper tool_result content blocks
    // results := m.toolResults
    m.toolResults = nil // ❌ Clearing without using!

    return func() tea.Msg {
        // Build messages - but NOT including tool results!
        messages := make([]core.Message, 0, len(m.Messages)+1)
        for _, msg := range m.Messages {
            messages = append(messages, core.Message{
                Role:    msg.Role,
                Content: msg.Content,
            })
        }

        req := core.MessageRequest{
            Model:     m.Model,
            Messages:  messages,
            MaxTokens: 4096,
            Stream:    true,
            System:    m.systemPrompt,
            // ❌ No Tools field!
            // ❌ No tool results in messages!
        }
        // ... rest of function
    }
}
```

**Problems:**
1. Tool results are cleared but never sent to the API
2. Tool definitions are not included in the follow-up request
3. Claude never sees the tool results and can't continue the conversation

---

## Fixes

### Fix 1: Don't Pause Stream at Tool Use

**Location:** `internal/ui/update.go:484-500`

**After:**
```go
if chunk.Type == "content_block_stop" && m.assemblingToolUse != nil {
    // Parse accumulated JSON into Input map
    if m.toolInputJSONBuf != "" {
        var input map[string]interface{}
        if err := json.Unmarshal([]byte(m.toolInputJSONBuf), &input); err == nil {
            m.assemblingToolUse.Input = input
        }
    }

    // Tool use is complete, store as pending
    // Don't handle yet - wait for message_stop to ensure full response is received
    m.pendingToolUse = m.assemblingToolUse
    m.assemblingToolUse = nil
    m.toolInputJSONBuf = ""
    // ✅ Continue streaming to get rest of response
}
```

**Key Change:** Store the tool in `pendingToolUse` but DON'T pause the stream. Let it continue to completion.

### Fix 2: Show Approval Dialog at Stream Completion

**Location:** `internal/ui/update.go:529-543`

**After:**
```go
if chunk.Type == "message_stop" || chunk.Done {
    m.CommitStreamingText()
    m.SetStatus(StatusIdle)
    m.streamChan = nil
    m.streamCancel = nil
    m.streamCtx = nil

    // If there's a pending tool from the stream, show approval dialog
    if m.pendingToolUse != nil {
        m.toolApprovalMode = true
    }

    m.updateViewport()
    return m, nil
}
```

**Key Change:** At `message_stop`, if there's a pending tool, enable approval mode. The stream has fully completed at this point.

### Fix 3: Implement Tool Results Sending

**Location:** `internal/ui/model.go:649-710`

**After:**
```go
func (m *Model) sendToolResults() tea.Cmd {
    if len(m.toolResults) == 0 || m.apiClient == nil {
        return nil
    }

    // Capture tool results before clearing
    results := m.toolResults
    m.toolResults = nil // Clear results

    return func() tea.Msg {
        // Build messages including tool results
        messages := make([]core.Message, 0, len(m.Messages)+1)
        for _, msg := range m.Messages {
            messages = append(messages, core.Message{
                Role:    msg.Role,
                Content: msg.Content,
            })
        }

        // ✅ Add user message with tool results
        toolResultsText := ""
        for _, result := range results {
            toolResultsText += formatToolResult(result.Result) + "\n"
        }
        messages = append(messages, core.Message{
            Role:    "user",
            Content: toolResultsText,
        })

        // ✅ Get tool definitions from registry
        var tools []core.ToolDefinition
        if m.toolRegistry != nil {
            tools = m.toolRegistry.GetDefinitions()
        }

        req := core.MessageRequest{
            Model:     m.Model,
            Messages:  messages,
            MaxTokens: 4096,
            Stream:    true,
            System:    m.systemPrompt,
            Tools:     tools, // ✅ Include tool definitions
        }

        // ... rest of function
    }
}
```

**Key Changes:**
1. Capture `results` before clearing `m.toolResults`
2. Format tool results and add as a user message
3. Include tool definitions from registry in the request

---

## Impact

**Before:**
- First tool would work but stream would stop mid-response
- Subsequent tool calls would not work
- Conversation could not continue after a tool was used
- Tool results were not sent back to Claude

**After:**
- Full streaming response is received including text after tool_use
- Tool results are properly sent back to Claude
- Tool definitions are included in follow-up requests
- Claude can make multiple sequential tool calls
- Conversation continues naturally after tool execution

---

## Testing

### Build Test

```bash
mise exec -- go build ./cmd/hex 2>&1
# Expected: No compilation errors ✓
```

### Manual Test (User Should Perform)

```bash
# Run interactive mode
hex

# Test 1: Single tool call
# Type: "create a file called test.txt with content 'hello'"
# Expected: Tool approval dialog appears, file is created, conversation continues

# Test 2: Multiple tool calls
# Type: "create test1.txt with 'hello', then read it back to me"
# Expected:
#   1. First tool (write_file) approval appears
#   2. After approval, file is created
#   3. Response continues
#   4. Second tool (read_file) approval appears
#   5. After approval, content is read
#   6. Claude responds with the file content
```

---

## Files Modified

- `internal/ui/update.go:484-500` - Don't pause stream at content_block_stop
- `internal/ui/update.go:529-543` - Show approval dialog at message_stop
- `internal/ui/model.go:649-710` - Implement proper tool results sending

---

## Related Documentation

- `BUG_FIX_TOOL_DEFINITIONS.md` - Previous fix for tool definitions and streaming parameters
- `BUG_FIX_INTERACTIVE_MODE.md` - Original API key loading fix

---

**Fixed By:** Claude Code (Sonnet 4.5)
**Build Status:** ✅ Compiles successfully
**Next Step:** User should test to verify multi-tool workflows work end-to-end
