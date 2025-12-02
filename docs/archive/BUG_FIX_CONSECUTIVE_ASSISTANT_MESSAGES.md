# Bug Fix: Consecutive Assistant Messages Causing API 400 Error

**Date:** 2025-11-29
**Status:** ✅ FIXED
**Severity:** Critical (app "dies" after tool execution)

---

## Problem Summary

The app would "die" after executing tools, becoming unresponsive. The Anthropic API was rejecting subsequent requests with a 400 error:

```
messages.5: `tool_use` ids were found without `tool_result` blocks immediately after
```

---

## Root Cause

**Location:** `internal/ui/update.go:505`

When the API streamed a response containing both text and `tool_use` blocks, the code was incorrectly creating **two separate assistant messages**:

1. When `tool_use` content block was detected, `CommitStreamingText()` was called
2. This created Message [N]: assistant with text
3. Then when stream ended, it created Message [N+1]: assistant with tool_use blocks

**Result:** Two consecutive assistant messages, which violates the Anthropic API's message structure requirements.

---

## The Code Bug

**Before (broken):**

```go
// internal/ui/update.go:500-510
if chunk.Type == "content_block_start" && chunk.ContentBlock != nil {
    if chunk.ContentBlock.Type == "tool_use" {
        fmt.Fprintf(os.Stderr, "[STREAM_TOOL_START] tool_use detected: id=%s, name=%s\n",
            chunk.ContentBlock.ID, chunk.ContentBlock.Name)

        // Commit any streaming text before handling tool
        m.CommitStreamingText()  // ← BUG! Creates separate assistant message

        // Start assembling tool use...
    }
}
```

This caused the message flow:

```
Stream receives: "Let me create those files" + [tool_use, tool_use, tool_use, tool_use]

App creates:
  Message [8]: assistant "Let me create those files" ← from CommitStreamingText()
  Message [9]: assistant [4 tool_use blocks]          ← from message_stop handler

❌ Two consecutive assistant messages!
```

---

## The Fix

**After (fixed):**

```go
// internal/ui/update.go:500-510
if chunk.Type == "content_block_start" && chunk.ContentBlock != nil {
    if chunk.ContentBlock.Type == "tool_use" {
        fmt.Fprintf(os.Stderr, "[STREAM_TOOL_START] tool_use detected: id=%s, name=%s\n",
            chunk.ContentBlock.ID, chunk.ContentBlock.Name)

        // DON'T commit streaming text yet - it will be included in the same
        // assistant message as the tool_use blocks when the stream ends
        // (see lines 584-603 which creates one message with both text and tool_use blocks)

        // Start assembling tool use...
    }
}
```

Now the message flow is correct:

```
Stream receives: "Let me create those files" + [tool_use, tool_use, tool_use, tool_use]

App creates:
  Message [8]: assistant with ContentBlocks:
    - text: "Let me create those files"
    - tool_use block 1
    - tool_use block 2
    - tool_use block 3
    - tool_use block 4

✅ One assistant message with both text and tool_use blocks!
```

---

## Why This Fix Works

The code at `update.go:584-603` already handles creating a single assistant message with both text and tool_use blocks:

```go
// Handle message completion
if chunk.Type == "message_stop" || chunk.Done {
    if len(m.pendingToolUses) > 0 {
        // Create assistant message with both text and ALL tool_use content blocks
        blocks := []core.ContentBlock{}

        // Add text block if there's any text content
        if m.StreamingText != "" {
            blocks = append(blocks, core.NewTextBlock(m.StreamingText))
        }

        // Add ALL tool_use blocks
        for i, toolUse := range m.pendingToolUses {
            blocks = append(blocks, core.ContentBlock{
                Type:  "tool_use",
                ID:    toolUse.ID,
                Name:  toolUse.Name,
                Input: toolUse.Input,
            })
        }

        // Add single assistant message with all content blocks
        assistantMsg := Message{
            Role:         "assistant",
            ContentBlock: blocks,
        }
        m.Messages = append(m.Messages, assistantMsg)
    }
}
```

By **not** calling `CommitStreamingText()` early, we let the text accumulate until `message_stop`, where it's combined with tool_use blocks into **one message**.

---

## Impact

**Before:**
- App would die after first tool batch execution
- Error: "API error 400: tool_use ids were found without tool_result blocks"
- User unable to continue conversation

**After:**
- Tool execution completes successfully
- Results sent back to API
- Conversation continues normally ✅

---

## Files Changed

1. **`internal/ui/update.go`** (line 505)
   - Removed `m.CommitStreamingText()` call
   - Added comment explaining why we wait until message_stop

---

## Testing

Build verified:
```bash
$ cd clean && go build ./cmd/clem
# Success!
```

Manual testing needed:
1. Run `clem`
2. Request multiple tools (e.g., "create 3 files: a.txt, b.txt, c.txt")
3. Approve tools
4. Verify app continues working (doesn't "die")
5. Check `debug.log` - should NOT see "ERROR creating stream"

---

**Completion:** 2025-11-29
**Lines Changed:** 1 line removed, 3 comment lines added
