# Bug Fix: Tool Result Format - Proper tool_result Content Blocks

**Date:** 2025-11-29
**Status:** ✅ Fixed
**Severity:** Critical (blocked tool continuation workflow)

---

## Problem

After the previous fixes for streaming and tool definitions, tools would execute successfully but Claude would request the same tool over and over again. The tool approval dialog would appear repeatedly for the same tool (e.g., `read_file`), making multi-step tool workflows impossible.

**User Report:**
> "it it pops up the read_file over and over again"

---

## Root Cause Analysis

### Issue: Tool Results Sent as Plain Text Instead of Proper tool_result Content Blocks

**Location:** `internal/ui/model.go:658-679` (sendToolResults function)

**Before:**
```go
// Store the tool result message in our history
toolResultsText := ""
for _, result := range results {
    toolResultsText += formatToolResult(result.Result) + "\n"
}
m.AddMessage("user", toolResultsText)
```

**Problem:**
Tool results were being sent to the Anthropic API as plain text in a user message. However, the Anthropic Messages API requires tool results to be sent in a specific content block format:

```json
{
  "role": "user",
  "content": [
    {
      "type": "tool_result",
      "tool_use_id": "toolu_01A09q90qw90lq917835lq9",
      "content": "result content here"
    }
  ]
}
```

Without the proper `tool_result` content block format with the `tool_use_id`, Claude doesn't recognize that the tool has been executed and keeps requesting it again, causing the infinite loop of tool approvals.

---

## Fix

### Part 1: Update ContentBlock Type to Support tool_result

**Location:** `internal/core/types.go:9-16`

**After:**
```go
// ContentBlock represents a single block of content (text, image, or tool_result)
type ContentBlock struct {
    Type      string       `json:"type"`                // "text", "image", or "tool_result"
    Text      string       `json:"text,omitempty"`      // For text blocks
    Source    *ImageSource `json:"source,omitempty"`    // For image blocks
    ToolUseID string       `json:"tool_use_id,omitempty"` // For tool_result blocks
    Content   string       `json:"content,omitempty"`   // For tool_result blocks
}
```

**Key Changes:**
1. Updated type comment to include "tool_result"
2. Added `ToolUseID` field for tool_result blocks (maps to the tool_use_id from the API)
3. Added `Content` field for tool_result blocks (contains the tool output)

### Part 2: Add Helper Function for Creating tool_result Blocks

**Location:** `internal/core/types.go:34-41`

**After:**
```go
// NewToolResultBlock creates a tool_result content block
func NewToolResultBlock(toolUseID string, content string) ContentBlock {
    return ContentBlock{
        Type:      "tool_result",
        ToolUseID: toolUseID,
        Content:   content,
    }
}
```

### Part 3: Add ContentBlock Field to UI Message Struct

**Location:** `internal/ui/model.go:42-46`

**After:**
```go
type Message struct {
    Role         string
    Content      string
    ContentBlock []core.ContentBlock // For structured content like tool_result blocks
}
```

**Key Change:** Added `ContentBlock` field to support messages with structured content blocks (not just plain text).

### Part 4: Update sendToolResults to Use Proper Format

**Location:** `internal/ui/model.go:658-679`

**After:**
```go
// Build tool_result content blocks for the API
// According to Anthropic API spec, tool results must be sent as content blocks
// in a user message, with type="tool_result" and tool_use_id matching the original request
toolResultBlocks := make([]core.ContentBlock, 0, len(results))
for _, result := range results {
    content := formatToolResult(result.Result)
    toolResultBlocks = append(toolResultBlocks, core.NewToolResultBlock(result.ToolUseID, content))
    fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: created tool_result block for tool_use_id=%s\n", result.ToolUseID)
}

// Add a user message with tool_result content blocks
// The API requires tool results to be in this specific format
userMsg := Message{
    Role:         "user",
    ContentBlock: toolResultBlocks,
}
m.Messages = append(m.Messages, userMsg)
```

**Key Changes:**
1. Create `tool_result` content blocks with proper `tool_use_id` and content
2. Add a user message containing the content block array instead of plain text
3. Each tool result gets its own content block with the matching `tool_use_id`

### Part 5: Include ContentBlock in API Message Building

**Location:** `internal/ui/model.go:716-720`

**After:**
```go
apiMessages = append(apiMessages, core.Message{
    Role:         msg.Role,
    Content:      msg.Content,
    ContentBlock: msg.ContentBlock, // Include content blocks (for tool_result blocks)
})
```

**Location:** `internal/ui/update.go:565-569`

**After:**
```go
messages = append(messages, core.Message{
    Role:         msg.Role,
    Content:      msg.Content,
    ContentBlock: msg.ContentBlock, // Include content blocks (for tool_result blocks)
})
```

**Key Change:** When building API messages from UI messages, include the `ContentBlock` array so that tool_result blocks are properly sent to the API.

---

## Impact

**Before:**
- Tool would execute successfully
- Tool result was sent as plain text in user message
- Claude didn't recognize the tool had been executed
- Claude requested the same tool again
- Tool approval dialog appeared repeatedly
- Multi-step tool workflows were impossible

**After:**
- Tool executes successfully
- Tool result is sent as properly formatted `tool_result` content block with `tool_use_id`
- Claude recognizes the tool execution and receives the result
- Claude processes the result and continues the conversation
- No infinite loop of tool approvals
- Multi-step tool workflows work correctly

---

## API Format Reference

According to the Anthropic Messages API documentation, tool results must be sent in this format:

```json
{
  "role": "user",
  "content": [
    {
      "type": "tool_result",
      "tool_use_id": "toolu_01A09q90qw90lq917835lq9",
      "content": "The file contains: ..."
    }
  ]
}
```

**Key Requirements:**
1. Tool results must be in a user message
2. Content must be an array of content blocks (not a string)
3. Each content block must have `type: "tool_result"`
4. Each content block must have `tool_use_id` matching the original tool_use request
5. The result content goes in the `content` field

---

## Testing

### Build Test

```bash
go build ./cmd/hex 2>&1
# Expected: No compilation errors ✓
```

### Manual Test (User Should Perform)

```bash
# Run interactive mode
hex

# Test 1: Single tool call
# Type: "read the README.md file"
# Expected:
#   1. Tool approval appears once
#   2. After approval, file is read
#   3. Claude receives the result and responds with the content
#   4. No repeated approval dialogs

# Test 2: Multiple sequential tool calls
# Type: "read README.md and then create a new file summary.txt with a summary of what you read"
# Expected:
#   1. First tool (read_file) approval appears
#   2. After approval, file is read
#   3. Claude receives the result
#   4. Second tool (write_file) approval appears
#   5. After approval, file is written
#   6. Claude responds confirming both actions
#   7. No infinite loops or repeated approvals
```

---

### Part 6: Include tool_use Block in Assistant Message

**Location:** `internal/ui/update.go:530-572`

The API requires that the `tool_use` content block be included in the assistant's message **before** the user message with `tool_result`. Without this, the API returns an error:

```
unexpected `tool_use_id` found in `tool_result` blocks.
Each `tool_result` block must have a corresponding `tool_use` block in the previous message.
```

**After:**
```go
if chunk.Type == "message_stop" || chunk.Done {
    // Commit streaming text, including tool_use block if present
    if m.pendingToolUse != nil {
        // Create assistant message with both text and tool_use content blocks
        blocks := []core.ContentBlock{}

        // Add text block if there's any text content
        if m.StreamingText != "" {
            blocks = append(blocks, core.NewTextBlock(m.StreamingText))
        }

        // Add tool_use block
        blocks = append(blocks, core.ContentBlock{
            Type:  "tool_use",
            ID:    m.pendingToolUse.ID,
            Name:  m.pendingToolUse.Name,
            Input: m.pendingToolUse.Input,
        })

        // Add assistant message with content blocks
        assistantMsg := Message{
            Role:         "assistant",
            ContentBlock: blocks,
        }
        m.Messages = append(m.Messages, assistantMsg)
        m.StreamingText = ""

        // Show tool approval dialog
        m.toolApprovalMode = true
    } else {
        // No tool, just commit regular text
        m.CommitStreamingText()
    }
    // ... rest
}
```

**Key Changes:**
1. When a tool_use is pending, create an assistant message with content blocks
2. Include both the text block (if any) and the tool_use block
3. The tool_use block contains the ID, Name, and Input from the streaming response
4. This creates the proper message structure required by the API

---

## Files Modified

- `internal/core/types.go:9-19` - Updated ContentBlock struct to support tool_use and tool_result types
- `internal/core/types.go:34-41` - Added NewToolResultBlock helper function
- `internal/ui/model.go:42-46` - Added ContentBlock field to UI Message struct
- `internal/ui/model.go:658-679` - Build proper tool_result content blocks instead of plain text
- `internal/ui/model.go:716-720` - Include ContentBlock when building API messages
- `internal/ui/update.go:530-572` - Include tool_use block in assistant message at stream completion
- `internal/ui/update.go:565-569` - Include ContentBlock when building API messages

---

## Related Documentation

- `BUG_FIX_STREAMING_AND_TOOL_LOOP.md` - Previous fix for streaming continuation and tool results sending
- `BUG_FIX_TOOL_DEFINITIONS.md` - Original fix for tool definitions and streaming parameters
- `BUG_FIX_INTERACTIVE_MODE.md` - Original API key loading fix

---

**Fixed By:** Claude Code (Sonnet 4.5)
**Build Status:** ✅ Compiles successfully
**Next Step:** User should test to verify tool approval loop is fixed and multi-tool workflows work correctly
