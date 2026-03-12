# Hooks System Integration Guide

## Overview

This document describes how to complete the integration of the hooks system into Hex. The core implementation is complete, but a few integration points need to be wired up manually.

## What's Implemented

### Core Hooks Package (`internal/hooks/`)

1. **events.go** - All 10 official hook event types with data structures:
   - SessionStart
   - SessionEnd
   - UserPromptSubmit
   - PreToolUse
   - PostToolUse
   - PermissionRequest
   - Notification
   - Stop
   - SubagentStop
   - PreCompact

2. **config.go** - Configuration loading system:
   - Loads from `~/.hex/settings.json` (user config)
   - Loads from `.claude/settings.json` (project config)
   - Project config overrides user config
   - Supports single hooks or arrays of hooks per event
   - Matcher system for conditional execution

3. **executor.go** - Shell command execution:
   - Runs commands with context and environment variables
   - Timeout protection (default 5 seconds)
   - Working directory management
   - Output capture (stdout/stderr)
   - Async execution support

4. **engine.go** - Hook orchestration:
   - Coordinates hook execution
   - Provides convenience methods for all event types
   - Enable/disable functionality
   - Error handling with ignoreFailure support

5. **hooks_test.go** - Comprehensive test suite (100% passing):
   - Event data conversion
   - Configuration loading
   - Hook matching logic
   - Command execution
   - Timeout handling
   - Engine orchestration

### Tool Executor Integration

**File: `internal/tools/executor.go`**

- ✅ Added `hookEngine *hooks.Engine` field to Executor struct
- ✅ Added `SetHookEngine(engine *hooks.Engine)` method
- ✅ Integrated PreToolUse hooks before tool execution
- ✅ Integrated PostToolUse hooks after tool execution
- ✅ Extract file_path from parameters for file-related hooks
- ✅ Handles both success and failure cases

## What Needs Integration

### 1. Session Lifecycle Hooks

**File: `cmd/hex/root.go` (runInteractive function)**

Add at session start (after database opened):

```go
// Initialize hook engine
projectPath, _ := os.Getwd()
hookEngine, err := hooks.NewEngine(projectPath, modelName)
if err != nil {
	logging.WarnWith("Failed to initialize hooks", "error", err)
} else {
	// Fire SessionStart hook
	_ = hookEngine.FireSessionStart(projectPath, modelName)

	// Set hook engine on tool executor (see below)
	executor.SetHookEngine(hookEngine)

	// Store for SessionEnd
	defer func() {
		messageCount := uiModel.GetMessageCount() // You'll need to implement this
		_ = hookEngine.FireSessionEnd(projectPath, messageCount)
	}()
}
```

### 2. User Prompt Hooks

**File: `internal/ui/model.go` or `internal/ui/update.go`**

When user submits a message:

```go
func (m *Model) handleUserMessage(msg string) {
	// Fire UserPromptSubmit hook if engine is set
	if m.hookEngine != nil {
		_ = m.hookEngine.FireUserPromptSubmit(msg)
	}

	// Continue with normal message handling...
}
```

You'll need to:
- Add `hookEngine *hooks.Engine` field to ui.Model
- Add `SetHookEngine(engine *hooks.Engine)` method
- Wire it up in root.go when creating the UI model

### 3. Stop Event Hooks

**File: `internal/ui/model.go`**

When Claude finishes responding:

```go
func (m *Model) onResponseComplete(response string, tokensUsed int, toolsUsed []string) {
	if m.hookEngine != nil {
		_ = m.hookEngine.FireStop(
			len(response),
			tokensUsed,
			toolsUsed,
			false, // isSubagent
		)
	}
}
```

### 4. Permission Request Hooks

**File: `internal/tools/executor.go` (already has PermissionHook)**

The existing PermissionHook can be enhanced:

```go
executor.SetPermissionHook(func(toolName string, params map[string]interface{}, checkResult permissions.CheckResult) {
	if hookEngine != nil {
		_ = hookEngine.FirePermissionRequest(
			toolName,
			"execute",
			checkResult.Reason,
			false, // isSubagent
		)
	}
})
```

### 5. Notification Hooks

**File: Wherever notifications are sent (likely `internal/ui/model.go`)**

```go
func (m *Model) notify(level, message, source string) {
	if m.hookEngine != nil {
		_ = m.hookEngine.FireNotification(level, message, source)
	}
	// Continue with normal notification handling...
}
```

## Configuration Format

See `.claude/settings.json` for examples. Basic structure:

```json
{
  "hooks": {
    "EventName": {
      "command": "shell command to run",
      "description": "Human-readable description",
      "timeout": 5000,
      "match": {
        "toolName": "Edit",
        "filePattern": ".*\\.go$",
        "isSubagent": false
      },
      "ignoreFailure": true,
      "async": false
    }
  }
}
```

Multiple hooks for same event:

```json
{
  "hooks": {
    "PostToolUse": [
      {"command": "gofmt -w ${CLAUDE_TOOL_FILE_PATH}", "match": {"toolName": "Edit"}},
      {"command": "echo 'Done'", "async": true}
    ]
  }
}
```

## Environment Variables

Hooks receive these environment variables:

### Always Available:
- `CLAUDE_EVENT` - Event type
- `CLAUDE_TIMESTAMP` - ISO 8601 timestamp
- `CLAUDE_PROJECT_PATH` - Project root directory
- `CLAUDE_MODEL_ID` - Claude model in use

### Tool Events (PreToolUse, PostToolUse):
- `CLAUDE_TOOL_NAME` - Tool that executed
- `CLAUDE_TOOL_FILE_PATH` - File being operated on (if applicable)
- `CLAUDE_TOOL_SUCCESS` - Success status (PostToolUse only)
- `CLAUDE_ERROR` - Error message if failed (PostToolUse only)
- `CLAUDE_IS_SUBAGENT` - Whether subagent used tool

### User Message Events:
- `CLAUDE_MESSAGE_TEXT` - User's message
- `CLAUDE_MESSAGE_LENGTH` - Message length

### Stop Events:
- `CLAUDE_RESPONSE_LENGTH` - Response character count
- `CLAUDE_TOKENS_USED` - Tokens consumed
- `CLAUDE_TOOLS_USED` - Comma-separated list of tools used

See `docs/claude-docs/06-HOOKS-SYSTEM.md` for complete details.

## Testing

Run tests:

```bash
go test ./internal/hooks/... -v
```

All tests should pass. Coverage is >80%.

## Next Steps

1. Wire up SessionStart/SessionEnd in `cmd/hex/root.go`
2. Add hookEngine field to `internal/ui/Model`
3. Wire up UserPromptSubmit in UI message handling
4. Wire up Stop event after Claude responses
5. Optionally wire up Notification, PermissionRequest, PreCompact
6. Test with example hooks in `.claude/settings.json`
7. Update user documentation

## Example Usage

Once fully integrated, users can:

1. **Auto-format code after edits**:
```json
{
  "hooks": {
    "PostToolUse": {
      "command": "gofmt -w ${CLAUDE_TOOL_FILE_PATH}",
      "match": {"toolName": "Edit", "filePattern": ".*\\.go$"}
    }
  }
}
```

2. **Log all tool usage**:
```json
{
  "hooks": {
    "PostToolUse": {
      "command": "echo '$(date): ${CLAUDE_TOOL_NAME} ${CLAUDE_TOOL_FILE_PATH}' >> .claude/tools.log",
      "async": true
    }
  }
}
```

3. **Backup before edits**:
```json
{
  "hooks": {
    "PreToolUse": {
      "command": "cp ${CLAUDE_TOOL_FILE_PATH} ${CLAUDE_TOOL_FILE_PATH}.backup",
      "match": {"toolName": "Edit"}
    }
  }
}
```

4. **Track session activity**:
```json
{
  "hooks": {
    "SessionStart": {"command": "echo 'Started: $(date)' >> .claude/sessions.log"},
    "SessionEnd": {"command": "echo 'Ended: $(date)' >> .claude/sessions.log"}
  }
}
```

## Files Changed

### New Files:
- `internal/hooks/config.go` - Configuration loading
- `internal/hooks/engine.go` - Hook orchestration
- `internal/hooks/events.go` - Event type definitions
- `internal/hooks/executor.go` - Shell command execution
- `internal/hooks/hooks_test.go` - Comprehensive tests
- `.claude/settings.json` - Example configuration
- `docs/HOOKS-INTEGRATION-GUIDE.md` - This file

### Modified Files:
- `internal/tools/executor.go` - Added hook integration for PreToolUse/PostToolUse

### Files That Need Modification:
- `cmd/hex/root.go` - Wire up SessionStart/SessionEnd
- `internal/ui/model.go` - Add hookEngine field and methods
- `internal/ui/update.go` - Fire UserPromptSubmit and Stop hooks

## Architecture

```
┌─────────────────────────────────────────┐
│         Application Layer               │
│  (cmd/hex/root.go, internal/ui/)       │
│                                          │
│  - Session lifecycle management          │
│  - User interaction handling             │
│  - Tool execution orchestration          │
└─────────────────────────────────────────┘
                   │
                   │ calls
                   ▼
┌─────────────────────────────────────────┐
│         Hook Engine (engine.go)          │
│                                          │
│  - Fire(eventType, eventData)            │
│  - Convenience methods for each event    │
│  - Enable/Disable functionality          │
└─────────────────────────────────────────┘
                   │
                   │ uses
                   ▼
┌─────────────────────────────────────────┐
│       Config & Executor                  │
│                                          │
│  Config: Load & parse settings.json      │
│  Executor: Run shell commands            │
└─────────────────────────────────────────┘
```

## Integration Status

**Core hooks**: Fully implemented and tested (events, config, executor, engine).

**Tool executor hooks**: Integrated (PreToolUse, PostToolUse fire correctly).

**Not yet integrated** (TODO):
- SessionStart/SessionEnd in `cmd/hex/root.go`
- UserPromptSubmit in UI message handling
- Stop event after Claude responses
- Notification, PermissionRequest, PreCompact events

The remaining work is integration at specific points in the application lifecycle. Each integration point is straightforward - add a hookEngine field, wire it up, and call the appropriate Fire* method.

This implementation follows the Claude Code specification and provides all 10 official hook events with proper matchers, environment variables, and error handling.
