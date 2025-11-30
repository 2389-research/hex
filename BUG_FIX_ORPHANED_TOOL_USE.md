# Bug Fix: Orphaned tool_use Blocks Causing API 400 Error

**Date:** 2025-11-29
**Status:** 🔴 CRITICAL BUG - App "dies" after tool execution
**Severity:** High (blocks all tool execution workflows)

---

## Problem Summary

The app "dies" after executing tools because the Anthropic API rejects subsequent streaming requests with a 400 error:

```
messages.5: `tool_use` ids were found without `tool_result` blocks immediately after:
toolu_01327koFPAtuX3BJrKb9VNs8, toolu_0152RP5Gyf7r58PgxwnSQvpU,
toolu_01HEbhnbXNLf48Zfo9YVig9X, toolu_01KnFchXm8qMZPHSxLFuhKRB.
Each `tool_use` block must have a corresponding `tool_result` block
in the next message.
```

---

## Root Cause Analysis

### The Anthropic API Message Structure Requirement

The API requires strict message ordering:
1. Assistant message with `tool_use` content blocks
2. **IMMEDIATELY** followed by a user message with corresponding `tool_result` blocks
3. No other messages can intervene

### What Our App Is Doing Wrong

The app adds "tool" role messages for UI display between the assistant message and user message:

```
Message [8]: assistant with 4 tool_use blocks
Message [10]: role="tool" (UI display: "Tool write_file succeeded...")  ← PROBLEM!
Message [11]: role="tool" (UI display: "Tool write_file succeeded...")  ← PROBLEM!
Message [12]: role="tool" (UI display: "Tool write_file succeeded...")  ← PROBLEM!
Message [13]: role="tool" (UI display: "Tool write_file succeeded...")  ← PROBLEM!
Message [14]: role="tool" (UI display: "Tool write_file failed...")     ← PROBLEM!
Message [15]: user with 5 tool_result blocks
```

When `sendToolResults()` filters out "tool" role messages (line 819-822 in model.go):

```go
if msg.Role == "tool" {
    fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: skipping tool role message\n")
    continue
}
```

The API sees:

```
Message [5]: assistant with 4 tool_use blocks
Message [6]: user with 5 tool_result blocks  ← NOT immediately after!
```

But wait - there's a mismatch! The user message has 5 tool_result blocks, but the assistant message only has 4 tool_use blocks. This suggests **the tool_result blocks are for a DIFFERENT batch of tools**.

---

## The Actual Bug

Looking more carefully at the logs, the issue is that:

1. **First Stream**: API sends assistant message with 4 tool_use blocks
   - App adds this to message history as message [8]

2. **Tools Execute**: All 4 tools execute successfully

3. **Problem**: Instead of IMMEDIATELY adding a user message with 4 tool_result blocks to message history, the app:
   - Adds "tool" role messages for UI display
   - Then tries to send tool results back to API
   - But `sendToolResults()` creates a NEW user message with tool_result blocks
   - This user message gets appended AFTER the "tool" messages

4. **Result**: The message history looks like:
   ```
   [8] assistant: 4 tool_use blocks
   [9] assistant: "Not stuck! Let me continue..."  ← Another assistant message!
   [10-14] tool: UI display messages (filtered out)
   [15] user: 5 tool_result blocks (for a different batch!)
   ```

The REAL issue is that message [9] is **another assistant message** that comes BEFORE the tool results for message [8] are added. This violates the API's requirement that tool_use blocks must be immediately followed by tool_result blocks.

---

## The Fix

We need to ensure that after an assistant message with tool_use blocks is added to the message history, the VERY NEXT message (excluding "tool" role messages for UI) is a user message with the corresponding tool_result blocks.

### Option 1: Don't Add Tool Messages to History (Recommended)

Keep "tool" role messages separate from the main message history:
- Store them in a separate UI-only list (`m.ToolDisplayMessages`)
- Don't add them to `m.Messages` which gets sent to the API
- Only add user/assistant messages to `m.Messages`

### Option 2: Filter Messages Before Adding to History

When adding the user message with tool_result blocks, ensure it's added immediately after the assistant message with tool_use blocks, before any other assistant messages.

### Option 3: Validate Message Structure Before Sending

Before calling `CreateMessageStream`, validate that every assistant message with tool_use blocks is immediately followed by a user message with matching tool_result blocks.

---

## Implementation Plan

**Recommended approach: Option 1**

### Changes Needed

1. **model.go** - Add separate storage for UI display messages:
   ```go
   type Model struct {
       // ... existing fields ...
       Messages            []Message     // API messages only (user/assistant)
       ToolDisplayMessages []Message     // UI-only tool result displays
   }
   ```

2. **model.go:AddMessage()** - Route "tool" messages to separate storage:
   ```go
   func (m *Model) AddMessage(role string, content string) {
       if role == "tool" {
           m.ToolDisplayMessages = append(m.ToolDisplayMessages, Message{
               Role:    role,
               Content: content,
           })
       } else {
           m.Messages = append(m.Messages, Message{
               Role:    role,
               Content: content,
           })
       }
   }
   ```

3. **update.go** - When displaying tool results, use `AddMessage("tool", ...)` as before, but it now goes to separate storage

4. **view.go** - When rendering chat history, combine both lists:
   ```go
   allMessages := append(m.Messages, m.ToolDisplayMessages...)
   sort.Slice(allMessages, /* by timestamp or insertion order */)
   ```

5. **sendToolResults()** - No longer needs to filter out "tool" messages since they're not in `m.Messages`

---

## Expected Behavior After Fix

1. User requests tools
2. API sends assistant message with tool_use blocks → added to `m.Messages`
3. Tools execute
4. Tool results displayed in UI → added to `m.ToolDisplayMessages`
5. User message with tool_result blocks → added to `m.Messages` immediately after assistant message
6. `sendToolResults()` sends clean message history to API
7. API accepts request and continues conversation ✅

---

## Testing

1. Request multiple tools (e.g., "create 3 files")
2. Approve and execute
3. Verify app continues and doesn't "die"
4. Check debug.log for "ERROR creating stream" - should not appear

---

**Next Steps:** Implement Option 1 fix
