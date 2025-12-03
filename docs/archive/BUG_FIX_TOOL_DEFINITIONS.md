# Bug Fix: Tool Definitions Not Sent to API

**Date:** 2025-11-28
**Status:** ✅ Fixed
**Severity:** High (blocks tool functionality)

---

## Problem

After fixing the API key loading bug, Claude was responding in interactive mode but didn't have access to any tools. The API was working but tools weren't being offered.

**User Report:**
> "it doesn't seem to have any tools"

---

## Root Cause Analysis

### Issue Location: `internal/tools/registry.go` and `internal/ui/update.go`

The tool registry existed and had 11+ tools registered in `cmd/hex/root.go:264-311`, but:

1. **No method to export tool definitions**: The `Registry` struct had no way to convert registered tools into the Claude API format (`[]core.ToolDefinition`)

2. **Not included in API requests**: `internal/ui/update.go:533-540` created the `MessageRequest` but didn't populate the `Tools` field

**Problem at `internal/ui/update.go:533-540` (BEFORE)**:
```go
// Create API request
req := core.MessageRequest{
    Model:     m.Model,
    Messages:  messages,
    MaxTokens: 4096,
    Stream:    true,
    System:    m.systemPrompt,
    // Tools field was missing!
}
```

---

## Fix

### Part 1: Added `GetDefinitions()` to Registry

Modified `internal/tools/registry.go` to add two new functions:

1. **`GetDefinitions()`**: Public method that returns `[]core.ToolDefinition` for all registered tools
2. **`getToolSchema(toolName string)`**: Helper function that returns proper JSON Schema for each tool's parameters

**Key tools with full schemas:**
- `read_file`: path (required), offset (optional), limit (optional)
- `write_file`: path, content (required), mode (optional: create/overwrite/append)
- `bash`: command (required), timeout, working_dir, run_in_background (optional)

**For other tools:** Falls back to minimal schema with empty properties (can be enhanced later)

**Code at `internal/tools/registry.go:66-155`:**
```go
// GetDefinitions returns tool definitions for all registered tools in API format
func (r *Registry) GetDefinitions() []core.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]core.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		def := core.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: getToolSchema(tool.Name()),
		}
		defs = append(defs, def)
	}
	return defs
}

// getToolSchema returns the JSON Schema for a specific tool's input parameters
func getToolSchema(toolName string) map[string]interface{} {
	switch toolName {
	case "read_file":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read",
				},
				// ... offset, limit ...
			},
			"required": []string{"path"},
		}
	// ... write_file, bash cases ...
	default:
		// For other tools, return minimal schema
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}
}
```

### Part 2: Updated UI to Include Tools in Request

Modified `internal/ui/update.go:533-547` to:
1. Call `m.toolRegistry.GetDefinitions()` if registry is not nil
2. Include tools in the `MessageRequest.Tools` field

**Code at `internal/ui/update.go:533-547` (AFTER)**:
```go
// Get tool definitions from registry
var tools []core.ToolDefinition
if m.toolRegistry != nil {
    tools = m.toolRegistry.GetDefinitions()
}

// Create API request
req := core.MessageRequest{
    Model:     m.Model,
    Messages:  messages,
    MaxTokens: 4096,
    Stream:    true,
    System:    m.systemPrompt,
    Tools:     tools,  // Now includes tool definitions!
}
```

---

## Design Decisions

### Why Manual Schema Mapping?

Per user's direction: **"skip the article for now. let's get it working without their new magic. then we can be model agnostic"**

Approach chosen:
- **Centralized schemas**: All tool schemas defined in `registry.go`'s `getToolSchema()` function
- **No interface changes**: The `Tool` interface remains simple (Name, Description, RequiresApproval, Execute)
- **Model agnostic**: Uses standard JSON Schema, not Anthropic-specific features
- **Easy to extend**: Other tools can be added to the switch statement as needed

Alternative approaches considered but rejected:
- Adding `Schema()` method to `Tool` interface → Would require changing every tool implementation
- Using reflection to auto-generate schemas → Complex, error-prone, harder to customize
- Advanced Anthropic tool features → User explicitly said skip for now

---

## Verification

### Build Test

```bash
mise exec -- go build ./cmd/hex 2>&1
# Expected: No compilation errors ✓
```

### Manual Test (User Should Perform)

```bash
# Run interactive mode
hex

# Type a message that requires a tool, e.g.:
# "read the file README.md"
# Expected: Claude should offer to use the read_file tool

# Try writing a file:
# "create a file called test.txt with content 'hello world'"
# Expected: Claude should offer to use the write_file tool
```

---

## Impact

**Before:** Interactive mode worked but Claude had no tool access - could only chat, not take actions

**After:** Claude receives tool definitions with every API request and can:
- Read files with `read_file`
- Write files with `write_file`
- Execute commands with `bash`
- Use 8+ other registered tools

---

## Related Code Paths

### Files Modified
- `internal/tools/registry.go:66-155` - Added GetDefinitions() and getToolSchema()
- `internal/ui/update.go:533-547` - Include tools in API request

### Files Involved (Not Changed)
- `internal/core/types.go:74-81` - ToolDefinition struct definition
- `internal/core/types.go:34-40` - MessageRequest struct with Tools field
- `cmd/hex/root.go:264-311` - Tool registration (ReadTool, WriteTool, BashTool, etc.)
- Individual tool files:
  - `internal/tools/read_tool.go` - Read tool implementation
  - `internal/tools/write_tool.go` - Write tool implementation
  - `internal/tools/bash_tool.go` - Bash tool implementation

---

## Next Steps (Future Enhancement)

The following tools currently use minimal schemas (empty properties):
- `edit_tool` (Phase 3)
- `grep_tool` (Phase 3)
- `glob_tool` (Phase 3)
- `ask_user_question` (Phase 4A)
- `todo_write` (Phase 4A)
- `web_fetch` (Phase 4B)
- `web_search` (Phase 4B)
- `task` (Phase 4C)
- `bash_output` (Phase 4C)
- `kill_shell` (Phase 4C)

To add schemas for these tools:
1. Examine the tool's `Execute()` method to understand its parameters
2. Add a case in `getToolSchema()` switch statement
3. Define proper JSON Schema with type, description, and required fields

---

**Fixed By:** Claude Code (Sonnet 4.5)
**Build Status:** ✅ Compiles successfully
**Next Step:** User should test interactive mode to verify tools are working

---

## Follow-up Fix: Streaming Tool Parameters

**Issue Discovered:** Tools were being approved before parameters finished streaming, causing "missing or invalid 'path' parameter" errors.

**Root Cause:** The streaming API sends tool_use incrementally:
1. `content_block_start` - Contains tool name and ID, but Input is empty
2. `content_block_delta` with `input_json_delta` - Contains partial JSON strings
3. `content_block_stop` - Signals parameters are complete

The code was triggering approval at step 1, before parameters arrived.

**Fix Applied:**

1. **Updated `internal/core/types.go:167-174`** - Added `PartialJSON` field to Delta struct for `input_json_delta` events

2. **Updated `internal/ui/model.go:100-109`** - Added fields:
   - `assemblingToolUse` - Tool being assembled from streaming chunks
   - `toolInputJSONBuf` - Buffer to accumulate JSON string fragments

3. **Updated `internal/ui/update.go:457-501`** - Changed streaming handler:
   - `content_block_start`: Initialize tool structure with empty Input
   - `content_block_delta`: Accumulate `partial_json` strings
   - `content_block_stop`: Parse complete JSON and trigger approval

Now the tool approval dialog waits until all parameters have been received and parsed.
