# Hooks System

## Overview

Hooks are shell commands that execute automatically at specific points in Claude Code's lifecycle. They enable automation, integration with external tools, and customization of Claude's behavior.

## What Are Hooks?

A hook is a command that runs when a specific event occurs:

- **Event-driven**: Triggered by Claude Code lifecycle events
- **Shell commands**: Any executable command or script
- **Context-aware**: Receive data about the triggering event
- **Conditional**: Can be filtered with matchers
- **Chainable**: Multiple hooks can run for the same event

Think of hooks as automated reactions to Claude's actions - running formatters after edits, logging user requests, managing sessions, etc.

## Hook Events

Claude Code provides 10 official hook events across the interaction lifecycle:

### 1. SessionStart
**When**: Claude Code session begins (before first user message)

**Use cases**:
- Initialize session logging
- Check for updates
- Start monitoring processes
- Load session-specific configuration

**Available data**:
```json
{
  "event": "SessionStart",
  "timestamp": "2025-12-01T10:30:00Z",
  "projectPath": "/Users/harper/myproject",
  "modelId": "claude-sonnet-4-5-20250929"
}
```

**Example**:
```json
{
  "hooks": {
    "SessionStart": {
      "command": "echo 'Session started at $(date)' >> .claude/session.log"
    }
  }
}
```

### 2. SessionEnd
**When**: Claude Code session terminates (clean shutdown)

**Use cases**:
- Commit uncommitted changes
- Generate session reports
- Clean up temporary files
- Stop monitoring processes

**Available data**:
```json
{
  "event": "SessionEnd",
  "timestamp": "2025-12-01T11:45:00Z",
  "projectPath": "/Users/harper/myproject",
  "messageCount": 47
}
```

**Example**:
```json
{
  "hooks": {
    "SessionEnd": {
      "command": "npm run build && git add -A"
    }
  }
}
```

### 3. UserPromptSubmit
**When**: User sends a message (before Claude processes it)

**Use cases**:
- Log user requests
- Validate message format
- Trigger pre-processing
- Start timers for analytics

**Available data**:
```json
{
  "event": "UserPromptSubmit",
  "timestamp": "2025-12-01T10:31:00Z",
  "messageText": "Fix the authentication bug",
  "messageLength": 26
}
```

**Example**:
```json
{
  "hooks": {
    "UserPromptSubmit": {
      "command": "echo '$(date): ${CLAUDE_MESSAGE_TEXT}' >> requests.log"
    }
  }
}
```

### 4. PreToolUse
**When**: After Claude creates tool parameters, before tool execution

**Use cases**:
- Validate tool usage
- Log tool calls
- Backup files before edits
- Check permissions

**Available data**:
```json
{
  "event": "PreToolUse",
  "timestamp": "2025-12-01T10:31:15Z",
  "toolName": "Edit",
  "parameters": {
    "file_path": "/Users/harper/myproject/src/auth.ts",
    "old_string": "const timeout = 30",
    "new_string": "const timeout = 60"
  },
  "isSubagent": false
}
```

**Example**:
```json
{
  "hooks": {
    "PreToolUse": {
      "command": "cp ${CLAUDE_TOOL_FILE_PATH} ${CLAUDE_TOOL_FILE_PATH}.backup",
      "match": {
        "toolName": "Edit"
      }
    }
  }
}
```

### 5. PostToolUse
**When**: Immediately after a tool completes successfully

**Use cases**:
- Format code after edits
- Run linters
- Update indexes
- Trigger builds

**Available data**:
```json
{
  "event": "PostToolUse",
  "timestamp": "2025-12-01T10:31:16Z",
  "toolName": "Edit",
  "parameters": {
    "file_path": "/Users/harper/myproject/src/auth.ts"
  },
  "success": true,
  "error": null,
  "isSubagent": false
}
```

**Example**:
```json
{
  "hooks": {
    "PostToolUse": {
      "command": "biome format --write ${CLAUDE_TOOL_FILE_PATH}",
      "match": {
        "toolName": "Edit",
        "filePattern": ".*\\.ts$"
      }
    }
  }
}
```

### 6. PermissionRequest
**When**: User is shown a permission dialog for a tool or action

**Use cases**:
- Log all permission requests
- Auto-approve common operations
- Send notifications for sensitive actions
- Track security decisions

**Available data**:
```json
{
  "event": "PermissionRequest",
  "timestamp": "2025-12-01T10:31:20Z",
  "toolName": "Bash",
  "action": "execute_command",
  "description": "Run npm test",
  "isSubagent": false
}
```

**Example**:
```json
{
  "hooks": {
    "PermissionRequest": {
      "command": "echo '$(date): Permission requested for ${CLAUDE_TOOL_NAME}' >> permissions.log",
      "async": true
    }
  }
}
```

### 7. Notification
**When**: Claude Code sends notifications to the user

**Use cases**:
- Forward notifications to external systems
- Log important events
- Trigger alerts
- Archive notifications

**Available data**:
```json
{
  "event": "Notification",
  "timestamp": "2025-12-01T10:31:25Z",
  "level": "info",
  "message": "Tool execution completed successfully",
  "source": "PostToolUse"
}
```

**Example**:
```json
{
  "hooks": {
    "Notification": {
      "command": "curl -X POST https://notifications.example.com/log -d '{\"msg\": \"${CLAUDE_NOTIFICATION_MESSAGE}\"}'",
      "async": true,
      "ignoreFailure": true
    }
  }
}
```

### 8. Stop
**When**: Main Claude Code agent finishes responding to user

**Use cases**:
- Clean up resources after response
- Log response metrics
- Trigger post-processing
- Update tracking systems

**Available data**:
```json
{
  "event": "Stop",
  "timestamp": "2025-12-01T10:31:45Z",
  "responseLength": 1247,
  "tokensUsed": 892,
  "toolsUsed": ["Read", "Edit"],
  "isSubagent": false
}
```

**Example**:
```json
{
  "hooks": {
    "Stop": {
      "command": "notify-send 'Claude finished responding'",
      "async": true,
      "ignoreFailure": true
    }
  }
}
```

### 9. SubagentStop
**When**: A subagent (spawned via Task tool call) finishes responding

**Use cases**:
- Track subagent performance
- Aggregate subagent results
- Log parallel work completion
- Cleanup subagent resources

**Available data**:
```json
{
  "event": "SubagentStop",
  "timestamp": "2025-12-01T10:32:10Z",
  "taskDescription": "Implement user authentication",
  "responseLength": 2100,
  "tokensUsed": 1450,
  "isSubagent": true
}
```

**Example**:
```json
{
  "hooks": {
    "SubagentStop": {
      "command": "echo '$(date): Subagent completed - ${CLAUDE_TASK_DESCRIPTION}' >> subagent.log",
      "async": true
    }
  }
}
```

### 10. PreCompact
**When**: Before Claude Code runs a compact operation (like compacting history)

**Use cases**:
- Backup session state before compacting
- Log compaction events
- Archive conversation history
- Trigger maintenance tasks

**Available data**:
```json
{
  "event": "PreCompact",
  "timestamp": "2025-12-01T10:35:00Z",
  "compactionType": "history",
  "currentSize": 2500,
  "estimatedNewSize": 1200
}
```

**Example**:
```json
{
  "hooks": {
    "PreCompact": {
      "command": "cp .claude/session.json .claude/session-backup-$(date +%s).json"
    }
  }
}
```

## Configuration Format

Hooks are configured in `.claude/settings.json`:

### Basic Hook

```json
{
  "hooks": {
    "PostToolUse": {
      "command": "echo 'Tool used'"
    }
  }
}
```

### Multiple Hooks for Same Event

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "command": "biome format --write ${CLAUDE_TOOL_FILE_PATH}",
        "match": { "toolName": "Edit" }
      },
      {
        "command": "npm run build",
        "match": { "filePattern": "src/.*\\.ts$" }
      }
    ]
  }
}
```

### Full Configuration

```json
{
  "hooks": {
    "PostToolUse": {
      "command": "biome format --write ${CLAUDE_TOOL_FILE_PATH}",
      "description": "Auto-format TypeScript files after editing",
      "timeout": 10000,
      "env": {
        "NODE_ENV": "development"
      },
      "match": {
        "toolName": "Edit",
        "filePattern": ".*\\.ts$",
        "isSubagent": false
      },
      "ignoreFailure": true
    }
  }
}
```

### Configuration Fields

| Field | Required | Description |
|-------|----------|-------------|
| `command` | Yes | Shell command to execute |
| `description` | No | Human-readable description |
| `timeout` | No | Max execution time in milliseconds (default: 30000) |
| `env` | No | Environment variables to set |
| `match` | No | Conditions for when hook runs |
| `ignoreFailure` | No | Continue even if command fails (default: false) |
| `async` | No | Run in background without blocking (default: false) |

## Matchers and Targeting

Matchers control when hooks execute:

### Tool Name Matching

```json
{
  "match": {
    "toolName": "Edit"
  }
}
```

Matches exact tool name. Common tools:
- `Edit`, `Write`, `Read`
- `Bash`, `Grep`, `Glob`
- `mcp__playwright__browser_navigate`

### File Pattern Matching

```json
{
  "match": {
    "filePattern": ".*\\.ts$"
  }
}
```

Regex pattern against `file_path` parameter. Examples:
- `.*\\.ts$` - TypeScript files
- `src/.*` - Files in src directory
- `.*/test/.*` - Test files
- `.*\\.(ts|tsx)$` - TypeScript and TSX files

### Multiple Conditions (AND)

```json
{
  "match": {
    "toolName": "Edit",
    "filePattern": "src/.*\\.ts$",
    "isSubagent": false
  }
}
```

All conditions must match.

### Subagent Filtering

```json
{
  "match": {
    "isSubagent": true
  }
}
```

- `true`: Only when subagent uses tool
- `false`: Only when main agent uses tool
- Omitted: Triggers for both

## Environment Variables

Hooks receive context via environment variables:

### Always Available

| Variable | Description | Example |
|----------|-------------|---------|
| `CLAUDE_EVENT` | Event type | `PostToolUse` |
| `CLAUDE_TIMESTAMP` | ISO 8601 timestamp | `2025-12-01T10:31:16Z` |
| `CLAUDE_PROJECT_PATH` | Project root directory | `/Users/harper/myproject` |
| `CLAUDE_MODEL_ID` | Claude model in use | `claude-sonnet-4-5-20250929` |

### Tool Events (PreToolUse, PostToolUse)

| Variable | Description | Example |
|----------|-------------|---------|
| `CLAUDE_TOOL_NAME` | Tool that executed | `Edit` |
| `CLAUDE_TOOL_FILE_PATH` | File being operated on | `/Users/harper/myproject/src/auth.ts` |
| `CLAUDE_TOOL_SUCCESS` | Success status | `true` or `false` |
| `CLAUDE_ERROR` | Error message (if failed) | `File not found` |
| `CLAUDE_IS_SUBAGENT` | Whether subagent used tool | `true` or `false` |

### User Message Events (UserPromptSubmit)

| Variable | Description | Example |
|----------|-------------|---------|
| `CLAUDE_MESSAGE_TEXT` | User's message | `Fix the bug in auth.ts` |
| `CLAUDE_MESSAGE_LENGTH` | Message length | `23` |

### Stop Events (Stop, SubagentStop)

| Variable | Description | Example |
|----------|-------------|---------|
| `CLAUDE_RESPONSE_LENGTH` | Response character count | `1247` |
| `CLAUDE_TOKENS_USED` | Tokens consumed | `892` |
| `CLAUDE_TASK_DESCRIPTION` | Task description (SubagentStop) | `Implement authentication` |

### Notification Events

| Variable | Description | Example |
|----------|-------------|---------|
| `CLAUDE_NOTIFICATION_MESSAGE` | Notification text | `Tool completed successfully` |
| `CLAUDE_NOTIFICATION_LEVEL` | Severity level | `info`, `warning`, `error` |

## Common Use Cases

### 1. Auto-formatting After Edits

```json
{
  "hooks": {
    "PostToolUse": {
      "command": "prettier --write ${CLAUDE_TOOL_FILE_PATH}",
      "description": "Format code after editing",
      "match": {
        "toolName": "Edit",
        "filePattern": ".*\\.(ts|tsx|js|jsx)$"
      },
      "ignoreFailure": true
    }
  }
}
```

### 2. Logging All Tool Usage

```json
{
  "hooks": {
    "PostToolUse": {
      "command": "echo '$(date -Iseconds) ${CLAUDE_TOOL_NAME} ${CLAUDE_TOOL_FILE_PATH}' >> .claude/tool-usage.log",
      "description": "Log all tool usage",
      "async": true,
      "ignoreFailure": true
    }
  }
}
```

### 3. Backup Before Edits

```json
{
  "hooks": {
    "PreToolUse": {
      "command": "cp ${CLAUDE_TOOL_FILE_PATH} ${CLAUDE_TOOL_FILE_PATH}.backup",
      "description": "Create backup before editing",
      "match": { "toolName": "Edit" }
    }
  }
}
```

### 4. Desktop Notifications on Completion

```json
{
  "hooks": {
    "Stop": {
      "command": "osascript -e 'display notification \"Claude finished responding\" with title \"Claude Code\"'",
      "description": "Notify when response complete (macOS)",
      "async": true,
      "ignoreFailure": true
    }
  }
}
```

### 5. Continuous Build on Changes

```json
{
  "hooks": {
    "PostToolUse": {
      "command": "npm run build",
      "description": "Rebuild after source changes",
      "match": {
        "toolName": "Edit",
        "filePattern": "src/.*"
      },
      "async": true,
      "timeout": 120000
    }
  }
}
```

### 6. Track Session Duration

```json
{
  "hooks": {
    "SessionStart": {
      "command": "echo '{\"start\": \"$(date -Iseconds)\"}' > .claude/session.json"
    },
    "SessionEnd": {
      "command": "echo ',\"end\": \"$(date -Iseconds)\"}' >> .claude/session.json"
    }
  }
}
```

### 7. Log Permission Requests

```json
{
  "hooks": {
    "PermissionRequest": {
      "command": "echo '$(date): Permission for ${CLAUDE_TOOL_NAME}' >> .claude/permissions.log",
      "async": true
    }
  }
}
```

### 8. Alert on Errors

```json
{
  "hooks": {
    "Notification": {
      "command": "curl -X POST https://hooks.slack.com/... -d '{\"text\": \"${CLAUDE_NOTIFICATION_MESSAGE}\"}'",
      "match": {
        "level": "error"
      },
      "async": true,
      "ignoreFailure": true
    }
  }
}
```

## Security Considerations

### 1. Command Injection

**Vulnerable**:
```json
{
  "command": "echo ${CLAUDE_MESSAGE_TEXT}"
}
```

User could inject: `$(rm -rf /)`

**Safe**:
```json
{
  "command": "python -c 'import os; print(os.environ[\"CLAUDE_MESSAGE_TEXT\"])'"
}
```

Use programming languages to safely handle untrusted data.

### 2. Path Traversal

**Vulnerable**:
```json
{
  "command": "cat ${CLAUDE_TOOL_FILE_PATH}"
}
```

Path could be: `../../../../etc/passwd`

**Safe**:
```json
{
  "command": "realpath -e ${CLAUDE_TOOL_FILE_PATH} | grep -q ^${CLAUDE_PROJECT_PATH} && cat ${CLAUDE_TOOL_FILE_PATH}"
}
```

Validate paths are within project.

### 3. Sensitive Data Exposure

**Risky**:
```json
{
  "command": "curl https://analytics.example.com -d \"${CLAUDE_MESSAGE_TEXT}\""
}
```

User messages might contain secrets, API keys, etc.

**Better**:
```json
{
  "command": "echo '${CLAUDE_EVENT}' | curl https://analytics.example.com -d @-"
}
```

Only send necessary, non-sensitive metadata.

### 4. Resource Exhaustion

**Problematic**:
```json
{
  "command": "npm run build",
  "timeout": 0  // No timeout!
}
```

**Safe**:
```json
{
  "command": "npm run build",
  "timeout": 120000,  // 2 minute limit
  "async": true       // Don't block Claude
}
```

Always set reasonable timeouts.

### 5. Credential Handling

**Never**:
```json
{
  "env": {
    "API_KEY": "sk-1234567890abcdef"  // Hard-coded secret!
  }
}
```

**Instead**:
```json
{
  "command": "source ~/.secrets && npm run deploy"
}
```

Load secrets from secure external sources.

## Debugging Hooks

### Check Hook Execution

Enable debug logging:
```bash
CLAUDE_DEBUG=hooks claude
```

### Test Hooks Manually

Extract and run commands:
```bash
# Get the command from settings.json
COMMAND=$(jq -r '.hooks.PostToolUse.command' .claude/settings.json)

# Set environment variables
export CLAUDE_TOOL_FILE_PATH="src/auth.ts"
export CLAUDE_TOOL_NAME="Edit"

# Run the command
eval "$COMMAND"
```

### Common Issues

**Hook doesn't run**:
- Check matcher conditions
- Verify event name is correct (use official hooks only)
- Ensure command is executable
- Check file permissions

**Hook fails silently**:
- Remove `ignoreFailure: true`
- Check stderr output
- Add explicit error handling
- Increase timeout

**Hook blocks Claude**:
- Set `async: true`
- Reduce command complexity
- Optimize slow operations
- Use background processes

**Environment variable empty**:
- Verify event provides that variable
- Check variable name spelling
- Ensure matcher allows execution
- Review tool parameters

## Advanced Patterns

### Conditional Execution

```bash
#!/bin/bash
if [ "${CLAUDE_TOOL_NAME}" = "Edit" ]; then
  npm run lint "${CLAUDE_TOOL_FILE_PATH}"
fi
```

### Multi-step Workflows

```bash
#!/bin/bash
set -e  # Exit on error

# 1. Format
prettier --write "${CLAUDE_TOOL_FILE_PATH}"

# 2. Lint
eslint --fix "${CLAUDE_TOOL_FILE_PATH}"

# 3. Type check
tsc --noEmit "${CLAUDE_TOOL_FILE_PATH}"
```

### Error Recovery

```bash
#!/bin/bash

npm test || {
  echo "Tests failed, rolling back..."
  git checkout HEAD -- "${CLAUDE_TOOL_FILE_PATH}"
  exit 1
}
```

### Parallel Hooks

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "command": "npm run lint",
        "async": true
      },
      {
        "command": "npm run type-check",
        "async": true
      },
      {
        "command": "npm test",
        "async": true
      }
    ]
  }
}
```

All run simultaneously without blocking each other.

### Hook Scripts

Instead of inline commands, use scripts:

`.claude/hooks/post-edit.sh`:
```bash
#!/bin/bash
set -euo pipefail

FILE="${CLAUDE_TOOL_FILE_PATH}"
PROJECT="${CLAUDE_PROJECT_PATH}"

# Change to project directory
cd "${PROJECT}"

# Format
prettier --write "${FILE}"

# Run file-specific tests
TEST_FILE="${FILE%.ts}.test.ts"
if [ -f "${TEST_FILE}" ]; then
  npm test -- "${TEST_FILE}"
fi

# Update build
npm run build
```

In `settings.json`:
```json
{
  "hooks": {
    "PostToolUse": {
      "command": "${CLAUDE_PROJECT_PATH}/.claude/hooks/post-edit.sh",
      "match": { "toolName": "Edit" }
    }
  }
}
```

## Integration with Other Systems

### Hooks + MCP Servers

Trigger MCP server actions:
```json
{
  "hooks": {
    "Stop": {
      "command": "curl -X POST http://localhost:3000/mcp/notify -d '{\"event\": \"response_complete\"}'"
    }
  }
}
```

### Hooks + External Logging

Forward all events to a logging service:
```json
{
  "hooks": {
    "Stop": {
      "command": "curl -X POST https://logs.example.com/events -d '{\"type\": \"claude_stop\", \"tokens\": ${CLAUDE_TOKENS_USED}}'",
      "async": true,
      "ignoreFailure": true
    }
  }
}
```

### Hooks + File Watchers

Start file watchers on session start:
```json
{
  "hooks": {
    "SessionStart": {
      "command": "npm run watch &",
      "async": true
    },
    "SessionEnd": {
      "command": "pkill -f 'npm run watch'"
    }
  }
}
```

## Best Practices

1. **Use official hooks only**: The 10 hooks documented here are the complete set
2. **Keep commands simple**: Complex logic goes in scripts
3. **Use timeouts**: Prevent hanging hooks
4. **Enable async for slow operations**: Don't block Claude
5. **Ignore failures for non-critical hooks**: Use `ignoreFailure: true`
6. **Validate inputs**: Don't trust environment variables
7. **Use matchers**: Only run when relevant
8. **Test hooks manually**: Before relying on them
9. **Document hook purpose**: Use `description` field
10. **Monitor hook performance**: Log execution times

## Summary

Hooks enable powerful automation in Claude Code:

- **10 official events** covering the full lifecycle
- **Shell command execution** for maximum flexibility
- **Conditional matching** for precise targeting
- **Environment variables** for context awareness
- **Security considerations** for safe execution

Use hooks to:
- Automate formatting and linting
- Log activity for analytics
- Manage session lifecycle
- Integrate with external systems
- Customize Claude's workflow

With careful design, hooks become invisible automation that enhances your development workflow without interrupting Claude's operation.
