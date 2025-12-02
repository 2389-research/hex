# Configuration

## Overview

Claude Code uses a layered configuration system that allows customization at both user and project levels. Configuration controls everything from API keys to UI appearance to tool behavior.

## Settings Hierarchy

Configuration is loaded in this priority order (later overrides earlier):

1. **Default Settings**: Built-in Claude Code defaults
2. **User Settings**: `~/.claude/settings.json`
3. **User Local Settings**: `~/.claude/settings.local.json`
4. **Project Settings**: `.claude/settings.json` (if exists)
5. **Environment Variables**: Runtime overrides

### Priority Example

```
Default: model = "claude-sonnet-4"
User:    model = "claude-sonnet-4.5"  ← Overrides default
Project: model = "claude-opus-4"      ← Overrides user setting
```

## User Configuration Directory

### Directory Structure

```
~/.claude/
├── settings.json              # Main user settings
├── settings.local.json        # Local overrides (gitignored)
├── config.json               # API keys and sensitive data
├── mcp.json                  # MCP server configuration
├── guidance.md               # Optional user instructions
├── commands/                 # Slash commands
│   ├── brainstorm.md
│   ├── plan.md
│   └── ...
├── skills/                   # User-defined skills
│   ├── skill-name/
│   │   └── SKILL.md
│   └── ...
├── plugins/                  # Downloaded plugins
├── hooks/                    # Event hooks
│   └── utilities/
├── docs/                     # User documentation
│   ├── python.md
│   ├── source-control.md
│   └── ...
├── history.jsonl            # Conversation history
├── debug/                   # Debug logs
└── todos/                   # Todo storage
```

## ~/.claude/settings.json Structure

### Complete Example

```json
{
  "cleanupPeriodDays": 999,
  "permissions": {
    "allow": [
      "mcp__zen__chat",
      "mcp__zen__thinkdeep",
      "mcp__socialmedia__login",
      "mcp__socialmedia__create_post",
      "mcp__private-journal__process_thoughts"
    ],
    "deny": []
  },
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "uv run /Users/harper/.claude/hooks/utilities/UserPromptSubmit/append_ultrathink.py"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "hooks": []
      },
      {
        "matcher": "clear",
        "hooks": [
          {
            "type": "command",
            "command": "uv run /Users/harper/.claude/hooks/utilities/BuddyNotification/buddy.py"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": []
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": []
      }
    ]
  },
  "statusLine": {
    "type": "command",
    "command": "bunx -y ccstatusline@latest",
    "padding": 0
  },
  "enabledPlugins": {
    "superpowers@superpowers-marketplace": true,
    "superpowers-developing-for-claude-code@superpowers-marketplace": true,
    "superpowers-chrome@superpowers-marketplace": true,
    "elements-of-style@superpowers-marketplace": true,
    "document-skills@anthropic-agent-skills": true
  },
  "alwaysThinkingEnabled": false,
  "feedbackSurveyState": {
    "lastShownTime": 1754360713440
  }
}
```

### Configuration Fields

#### cleanupPeriodDays

Controls automatic cleanup of old sessions and history.

```json
{
  "cleanupPeriodDays": 999
}
```

- **Type**: number
- **Default**: 30
- **Purpose**: Days to retain history before cleanup
- **Note**: Set to high value (999) to preserve all history

#### permissions

Controls which MCP tools are allowed or denied.

```json
{
  "permissions": {
    "allow": [
      "mcp__private-journal__process_thoughts",
      "mcp__socialmedia__create_post"
    ],
    "deny": [
      "mcp__dangerous_tool__*"
    ]
  }
}
```

- **allow**: Array of tool names to explicitly permit
- **deny**: Array of tool names to block (supports wildcards)
- **Evaluation**: Deny list takes precedence over allow list

#### hooks

Event-driven automation hooks.

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "python /path/to/script.py"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Session started'"
          }
        ]
      }
    ]
  }
}
```

**Available Hook Points:**
- `UserPromptSubmit`: Triggered when user submits a prompt
- `SessionStart`: Triggered at session initialization
- `PreToolUse`: Before any tool execution
- `PostToolUse`: After any tool execution
- `Stop`: When session ends
- `Notification`: On notification events

**Hook Configuration:**
- `matcher`: Optional filter (e.g., "startup", "clear", tool name)
- `type`: Hook type ("command")
- `command`: Shell command to execute

#### statusLine

Customizes the status line display.

```json
{
  "statusLine": {
    "type": "command",
    "command": "bunx -y ccstatusline@latest",
    "padding": 0
  }
}
```

- **type**: "command" or "text"
- **command**: Command to generate status line
- **padding**: Spacing around status line

Disable status line:
```json
{
  "statusLine": null
}
```

#### enabledPlugins

Controls which plugins are active.

```json
{
  "enabledPlugins": {
    "superpowers@superpowers-marketplace": true,
    "document-skills@anthropic-agent-skills": true,
    "my-plugin@local": false
  }
}
```

- **Key**: Plugin identifier (name@source)
- **Value**: Boolean (true = enabled, false = disabled)

#### alwaysThinkingEnabled

Enables continuous thinking mode.

```json
{
  "alwaysThinkingEnabled": false
}
```

- **Type**: boolean
- **Default**: false
- **Purpose**: Enable extended thinking for all responses

## ~/.claude/config.json Structure

Contains sensitive configuration like API keys.

```json
{
  "apiKey": "sk-ant-api03-xxxxxxxxxxxxxxxxxxxxx"
}
```

**Security Notes:**
- **Never commit**: Add to .gitignore
- **Permissions**: Set to 600 (user read/write only)
- **Rotation**: Rotate keys regularly
- **Environment**: Consider using environment variables in CI/CD

## ~/.claude/mcp.json Structure

Configures MCP servers (see [MCP Servers](09-MCP-SERVERS.md) for details).

```json
{
  "servers": {
    "server-name": {
      "command": "node",
      "args": ["/path/to/server/index.js"],
      "env": {
        "API_KEY": "value"
      },
      "disabled": false
    }
  }
}
```

## Project .claude Directory

Project-specific configuration in `.claude/` directory within project root.

### Structure

```
project/.claude/
├── settings.json          # Project-specific settings
├── guidance.md             # Project instructions
├── commands/             # Project commands
└── context/              # Project context files
```

### Project settings.json

Overrides user settings for specific project:

```json
{
  "model": "claude-opus-4",
  "permissions": {
    "allow": [
      "mcp__filesystem__*"
    ]
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "npm run lint-staged"
          }
        ]
      }
    ]
  }
}
```

### Project guidance

Project-specific instructions loaded automatically:

```markdown
# Project: MyApp

## Architecture
This is a microservices architecture with:
- Frontend: React + TypeScript
- Backend: Node.js + Express
- Database: PostgreSQL

## Coding Standards
- Use functional components
- Write tests first (TDD)
- No any types in TypeScript

## Deployment
Deploy to staging before production.
```

## Model Configuration

### Specifying Model

```json
{
  "model": "claude-sonnet-4-5-20250929"
}
```

**Available Models:**
- `claude-opus-4`: Most capable, slower
- `claude-sonnet-4`: Balanced performance
- `claude-sonnet-4-5-20250929`: Latest Sonnet 4.5
- `claude-haiku-4`: Fast, efficient

### Model Selection Strategy

```json
{
  "modelSelection": {
    "default": "claude-sonnet-4",
    "forComplexTasks": "claude-opus-4",
    "forSimpleTasks": "claude-haiku-4"
  }
}
```

### Temperature and Sampling

```json
{
  "temperature": 0.7,
  "maxTokens": 4096
}
```

## Memory and Context Management

### Context Window Configuration

```json
{
  "contextWindow": {
    "maxTokens": 200000,
    "preserveHistory": true,
    "summarizeOldMessages": false
  }
}
```

### History Retention

```json
{
  "history": {
    "retentionDays": 999,
    "maxSessions": 1000,
    "autoCleanup": false
  }
}
```

### Session Management

```json
{
  "session": {
    "autoSave": true,
    "saveInterval": 60000,
    "restoreOnStartup": true
  }
}
```

## Statusline Customization

### Built-in Status Line

```json
{
  "statusLine": {
    "type": "text",
    "text": "Claude Code v1.0.0"
  }
}
```

### Command-Based Status Line

```json
{
  "statusLine": {
    "type": "command",
    "command": "git branch --show-current",
    "padding": 1,
    "refreshInterval": 5000
  }
}
```

### Custom Status Line Script

```bash
#!/bin/bash
# ~/.claude/statusline.sh

BRANCH=$(git branch --show-current 2>/dev/null || echo "no git")
STATUS=$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')
TIME=$(date +%H:%M)

echo "[$BRANCH] Changes: $STATUS | $TIME"
```

Configuration:
```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh",
    "padding": 0
  }
}
```

### ccstatusline Package

Advanced status line with git integration:

```bash
npm install -g ccstatusline
# or use via bunx
```

```json
{
  "statusLine": {
    "type": "command",
    "command": "bunx -y ccstatusline@latest",
    "padding": 0
  }
}
```

## Output Styles

### Terminal Output Formatting

```json
{
  "output": {
    "colorEnabled": true,
    "syntaxHighlighting": true,
    "codeBlockStyle": "fenced",
    "tableStyle": "markdown"
  }
}
```

### Logging Configuration

```json
{
  "logging": {
    "level": "info",
    "destination": "~/.claude/debug/",
    "includeTimestamps": true,
    "includeToolCalls": true
  }
}
```

### Debug Mode

```json
{
  "debug": {
    "enabled": false,
    "verboseToolCalls": false,
    "logMCPMessages": false,
    "saveSnapshots": true
  }
}
```

## Tool Configuration

### Tool Timeouts

```json
{
  "tools": {
    "defaultTimeout": 120000,
    "Bash": {
      "timeout": 600000
    },
    "WebFetch": {
      "timeout": 30000
    }
  }
}
```

### Tool Permissions

```json
{
  "tools": {
    "Bash": {
      "allowDangerousCommands": false,
      "requireConfirmation": ["rm", "mv", "git push --force"]
    },
    "Write": {
      "requireBackup": true
    }
  }
}
```

## Environment-Specific Configuration

### Development Environment

```json
{
  "environment": "development",
  "debug": {
    "enabled": true
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "npm run lint"
          }
        ]
      }
    ]
  }
}
```

### Production Environment

```json
{
  "environment": "production",
  "debug": {
    "enabled": false
  },
  "tools": {
    "Bash": {
      "requireConfirmation": ["git push", "npm publish", "docker push"]
    }
  }
}
```

## Best Practices

### 1. Separate Sensitive Data

```
~/.claude/
├── settings.json       # Safe to commit (no secrets)
├── config.json         # NEVER commit (API keys)
└── settings.local.json # NEVER commit (local overrides)
```

### 2. Use Environment Variables

Instead of:
```json
{
  "servers": {
    "github": {
      "env": {
        "GITHUB_TOKEN": "ghp_xxxxx"
      }
    }
  }
}
```

Use:
```json
{
  "servers": {
    "github": {
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

### 3. Layer Configuration Appropriately

- **User settings**: Personal preferences, default tools
- **Project settings**: Project-specific requirements
- **Environment variables**: Secrets, deployment-specific values

### 4. Document Custom Configuration

```json
{
  "_comment": "Custom hook to run tests before committing",
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "npm test"
          }
        ]
      }
    ]
  }
}
```

### 5. Version Control Configuration

Add to `.gitignore`:
```
.claude/config.json
.claude/settings.local.json
.claude/history.jsonl
.claude/debug/
.claude/todos/
```

Commit to version control:
```
.claude/settings.json
# (include project guidance doc if versioned)
.claude/commands/
```

## Troubleshooting

### Configuration Not Loading

1. **Check JSON syntax**: Use `jq` to validate
   ```bash
   jq . ~/.claude/settings.json
   ```

2. **Check file permissions**: Ensure readable
   ```bash
   ls -la ~/.claude/settings.json
   ```

3. **Check hierarchy**: Verify which config is taking precedence

### Hooks Not Firing

1. **Verify hook command**: Run manually to test
2. **Check matcher**: Ensure matcher pattern is correct
3. **Review logs**: Check `~/.claude/debug/` for errors

### MCP Servers Not Starting

1. **Test command**: Run server command manually
2. **Check environment**: Verify env vars are set
3. **Review mcp.json**: Validate configuration syntax

### Plugin Issues

1. **Check enabledPlugins**: Verify plugin is enabled
2. **Reinstall plugin**: Remove and reinstall if corrupted
3. **Check compatibility**: Ensure plugin version is compatible

## See Also

- [MCP Servers](09-MCP-SERVERS.md) - External tool integration
- [Slash Commands](10-SLASH-COMMANDS.md) - Custom commands
- [Verification and Testing](12-VERIFICATION-AND-TESTING.md) - Testing workflows
