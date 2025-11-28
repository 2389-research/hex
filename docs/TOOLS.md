# Clem Tool System Reference

Complete reference for Clem's tool execution system and available tools.

## Table of Contents

- [Overview](#overview)
- [Tool Approval](#tool-approval)
- [Available Tools](#available-tools)
- [Read Tool](#read-tool)
- [Write Tool](#write-tool)
- [Bash Tool](#bash-tool)
- [Edit Tool](#edit-tool)
- [Grep Tool](#grep-tool)
- [Glob Tool](#glob-tool)
- [Safety Features](#safety-features)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Overview

Clem implements Claude's tool use capability, allowing Claude to:
- Read files from your filesystem
- Write or modify files
- Execute shell commands

All tool operations require explicit user approval for safety.

### How Tools Work

1. **Claude requests a tool** based on your conversation
2. **Clem intercepts the request** and shows you the details
3. **You approve or deny** the execution
4. **Tool executes** (if approved) and returns results
5. **Claude sees the results** and continues the conversation

### Tool Execution Flow

```
You: "Read config.yaml and explain it"
    │
    ▼
Claude: [Requests read_file tool]
    │
    ▼
Clem: ┌──────────────────────────────┐
      │ Tool Approval Required:       │
      │                              │
      │ Tool: read_file              │
      │ Path: /path/to/config.yaml   │
      │ Size: 1.2 KB                 │
      │                              │
      │ Approve? [y/N]               │
      └──────────────────────────────┘
    │
    ▼
You: y
    │
    ▼
Tool executes ──> Returns content
    │
    ▼
Claude: [Analyzes content and responds]
    │
    ▼
"This config file sets up..."
```

## Tool Approval

### Approval Prompt

When a tool requires approval, you'll see:

```
┌─────────────────────────────────────────┐
│ Tool Execution Request                  │
│                                         │
│ Tool: <tool_name>                       │
│ Parameters:                              │
│   param1: value1                        │
│   param2: value2                        │
│                                         │
│ Approve execution? [y/N]                │
└─────────────────────────────────────────┘
```

### Responding to Approval

**Approve**: Type `y` or `yes` and press Enter
**Deny**: Type `n`, `no`, or just press Enter

### When Approval Is Required

Different tools have different approval rules:

**Read Tool**:
- Always approved for normal files
- Requires approval for sensitive paths:
  - `/etc/*` (system config)
  - `~/.ssh/*` (SSH keys)
  - `~/.aws/*` (AWS credentials)
  - `.env` files (secrets)

**Write Tool**:
- Create mode: Auto-approved (won't overwrite)
- Overwrite mode: Requires approval (destructive)
- Append mode: Auto-approved (additive)

**Bash Tool**:
- Safe commands: Auto-approved (ls, pwd, echo, etc.)
- Dangerous commands: Requires approval (rm -rf, sudo, curl, etc.)
- Long-running commands (timeout > 60s): Requires approval

### Approval Security

**Clem never**:
- Executes tools without showing you details first
- Remembers approval decisions across sessions
- Runs dangerous operations silently

**You always**:
- See full tool parameters before execution
- Have final say on tool execution
- Can deny any operation

## Available Tools

### Summary

Clem provides 6 core tools for file operations, code editing, and search:

| Tool | Purpose | Approval Rules | Timeout |
|------|---------|----------------|---------|
| `read_file` | Read file contents | Sensitive paths only | N/A |
| `write_file` | Create/modify files | Overwrites only | N/A |
| `bash` | Execute commands | Dangerous commands | 30s default, 5min max |
| `edit` | Replace strings in files | Always | N/A |
| `grep` | Search code with ripgrep | Never (read-only) | N/A |
| `glob` | Find files by pattern | Never (read-only) | N/A |

## Read Tool

**Tool Name**: `read_file`

**Purpose**: Safely read file contents and return to Claude

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_path` | string | Yes | Absolute or relative path to file |

### Example Request

```json
{
  "tool": "read_file",
  "parameters": {
    "file_path": "/path/to/file.txt"
  }
}
```

### Safety Features

**Path Validation**:
- Resolves relative paths to absolute
- Prevents directory traversal (`../../../etc/passwd`)
- Checks file exists and is readable

**Size Limits**:
- Maximum file size: 10MB (configurable)
- Prevents memory exhaustion from huge files

**Content Validation**:
- UTF-8 encoding check
- Returns error for binary files

**Sensitive Path Detection**:
Requires approval for:
- `/etc/*` - System configuration
- `~/.ssh/*` - SSH keys and config
- `~/.aws/*` - AWS credentials
- `.env`, `*.env` - Environment secrets
- `~/.gnupg/*` - GPG keys
- `/proc/*`, `/sys/*` - System internals

### Return Value

**Success**:
```json
{
  "success": true,
  "output": "file contents here...",
  "metadata": {
    "size": 1234,
    "path": "/absolute/path/to/file.txt"
  }
}
```

**Failure**:
```json
{
  "success": false,
  "error": "file not found: /path/to/file.txt"
}
```

### Example Usage

**User**: "Read package.json and tell me what dependencies we use"

**Claude**: [Requests read_file with path: package.json]

**Clem**: Shows approval prompt

**User**: Approves

**Tool**: Returns package.json contents

**Claude**: "Based on package.json, you're using: React, Tailwind, TypeScript..."

### Error Handling

| Error | Reason | Solution |
|-------|--------|----------|
| File not found | Path doesn't exist | Check path spelling |
| Permission denied | No read access | Check file permissions |
| File too large | Size > 10MB | Read in chunks or use bash tool |
| Binary file | Non-UTF-8 content | Use bash tool with hexdump |

## Write Tool

**Tool Name**: `write_file`

**Purpose**: Create new files or modify existing files

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_path` | string | Yes | Path to write to |
| `content` | string | Yes | Content to write |
| `mode` | string | No | Operation mode (default: create) |

### Modes

**create** (default):
- Creates new file
- Fails if file already exists
- Auto-approved (safe)

**overwrite**:
- Replaces existing file
- Creates file if doesn't exist
- Requires approval (destructive)

**append**:
- Adds to end of existing file
- Creates file if doesn't exist
- Auto-approved (non-destructive)

### Example Request

```json
{
  "tool": "write_file",
  "parameters": {
    "file_path": "/path/to/output.txt",
    "content": "Hello, world!",
    "mode": "create"
  }
}
```

### Safety Features

**Atomic Writes**:
- Writes to temp file first
- Moves to final location only on success
- Prevents partial writes on error

**Directory Creation**:
- Creates parent directories if needed
- Uses 0755 permissions

**Validation**:
- Checks write permissions
- Validates content encoding
- Returns detailed errors

**Confirmation for Overwrites**:
```
┌─────────────────────────────────────────┐
│ Write Tool: Overwrite Confirmation      │
│                                         │
│ File: /path/to/existing.txt             │
│ Current size: 1.5 KB                    │
│ New size: 2.3 KB                        │
│                                         │
│ This will REPLACE the existing file.    │
│ Approve? [y/N]                          │
└─────────────────────────────────────────┘
```

### Return Value

**Success**:
```json
{
  "success": true,
  "output": "File written successfully",
  "metadata": {
    "path": "/absolute/path/to/file.txt",
    "size": 1234,
    "mode": "create"
  }
}
```

**Failure**:
```json
{
  "success": false,
  "error": "file already exists (use mode=overwrite to replace)"
}
```

### Example Usage

**User**: "Create a README.md with a basic project description"

**Claude**: [Requests write_file with README.md content]

**Clem**: Auto-approved (create mode, new file)

**Tool**: Creates README.md

**Claude**: "I've created README.md with..."

### Error Handling

| Error | Reason | Solution |
|-------|--------|----------|
| File exists | Create mode, file exists | Use overwrite or append mode |
| Permission denied | No write access | Check directory permissions |
| Disk full | Out of space | Free up disk space |
| Invalid path | Bad file path | Check path syntax |

## Bash Tool

**Tool Name**: `bash`

**Purpose**: Execute shell commands and capture output

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `command` | string | Yes | Shell command to execute |
| `working_dir` | string | No | Working directory (default: current) |
| `timeout` | number | No | Timeout in seconds (default: 30, max: 300) |

### Example Request

```json
{
  "tool": "bash",
  "parameters": {
    "command": "ls -la",
    "working_dir": "/path/to/dir",
    "timeout": 10
  }
}
```

### Safety Features

**Sandboxing**:
- No interactive shells
- No shell variable expansion (unless needed)
- Controlled environment

**Timeout Protection**:
- Default: 30 seconds
- Maximum: 5 minutes (300 seconds)
- Prevents hung processes

**Dangerous Command Detection**:
Requires approval for:
- `rm -rf` (recursive delete)
- `sudo` (privilege escalation)
- `curl`, `wget` (network access)
- `dd` (low-level disk operations)
- `mkfs`, `fdisk` (disk formatting)
- `kill`, `pkill` (process termination)
- `chmod 777` (insecure permissions)
- `> /dev/sda` (disk writing)

**Safe Commands** (auto-approved):
- `ls`, `pwd`, `echo`, `cat`, `grep`
- `find`, `wc`, `head`, `tail`
- `git status`, `git log`, `git diff`
- Read-only operations

### Return Value

**Success**:
```json
{
  "success": true,
  "output": "command output here...\n",
  "metadata": {
    "exit_code": 0,
    "duration_ms": 123,
    "command": "ls -la",
    "working_dir": "/path/to/dir"
  }
}
```

**Failure**:
```json
{
  "success": false,
  "error": "command failed with exit code 1",
  "output": "stderr output here...",
  "metadata": {
    "exit_code": 1
  }
}
```

### Example Usage

**User**: "List all Go files in the current directory"

**Claude**: [Requests bash with command: find . -name "*.go"]

**Clem**: Auto-approved (safe command)

**Tool**: Executes and returns file list

**Claude**: "Found 12 Go files: main.go, client.go, ..."

### Timeout Handling

**Default timeout (30s)**:
```
Command: ls -la
Timeout: 30s (default)
Status: Auto-approved
```

**Long timeout (requires approval)**:
```
┌─────────────────────────────────────────┐
│ Bash Tool: Long-Running Command         │
│                                         │
│ Command: npm install                    │
│ Timeout: 300s (5 minutes)               │
│                                         │
│ This may run for a while.               │
│ Approve? [y/N]                          │
└─────────────────────────────────────────┘
```

**Timeout exceeded**:
```json
{
  "success": false,
  "error": "command timeout after 30s",
  "output": "partial output before timeout...",
  "metadata": {
    "timeout": true,
    "duration_ms": 30000
  }
}
```

### Error Handling

| Error | Reason | Solution |
|-------|--------|----------|
| Command not found | Binary not in PATH | Check command spelling, install if needed |
| Permission denied | No execute access | Check permissions or use sudo (with approval) |
| Timeout | Command too slow | Increase timeout or optimize command |
| Exit code != 0 | Command failed | Check stderr output for error details |

---

## Edit Tool

The Edit tool performs exact string replacements in files - Claude's primary method for making code changes.

### Purpose

- Replace exact strings in files
- Support single or multiple occurrences
- Preserve file encoding and formatting
- Ensure changes are intentional (not ambiguous)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_path` | string | ✅ | Path to file to edit (absolute or relative) |
| `old_string` | string | ✅ | Exact string to find and replace |
| `new_string` | string | ✅ | Replacement string (must differ from old_string) |
| `replace_all` | boolean | ❌ | Replace all occurrences (default: false, requires unique match) |

### Behavior

**Single Replacement (default)**:
- `old_string` must appear EXACTLY ONCE in the file
- Fails if not found or appears multiple times
- Forces Claude to be specific and unambiguous

**Replace All Mode**:
- Replaces ALL occurrences of `old_string`
- Useful for renaming variables, updating imports, etc.
- Use when intentional bulk replacement is needed

### Approval

**ALWAYS requires approval** - editing files is destructive.

### Examples

#### Basic Edit
```json
{
  "file_path": "src/config.go",
  "old_string": "MaxRetries = 3",
  "new_string": "MaxRetries = 5"
}
```

#### Multiline Edit
```json
{
  "file_path": "main.go",
  "old_string": "func old() {\n\treturn nil\n}",
  "new_string": "func updated() {\n\treturn errors.New(\"not implemented\")\n}"
}
```

#### Replace All
```json
{
  "file_path": "package.json",
  "old_string": "\"version\": \"1.0.0\"",
  "new_string": "\"version\": \"1.1.0\"",
  "replace_all": true
}
```

#### Variable Rename
```json
{
  "file_path": "app.js",
  "old_string": "oldVariableName",
  "new_string": "newVariableName",
  "replace_all": true
}
```

### Error Handling

| Error | Reason | Solution |
|-------|--------|----------|
| String not found | `old_string` doesn't exist | Verify exact string (including whitespace) |
| Ambiguous match | Multiple occurrences found | Use more context or set `replace_all: true` |
| Identical strings | `old_string` == `new_string` | Must actually change something |
| File not found | Path doesn't exist | Check file path |
| Permission denied | No write access | Check file permissions |

### Safety Notes

- Edit preserves exact indentation from `old_string`
- Supports Unicode characters
- Respects line endings (LF/CRLF)
- No regex - exact literal matching only
- Changes are atomic (write succeeds or fails completely)

### Common Patterns

**Fix typo**:
```json
{
  "file_path": "README.md",
  "old_string": "recieve",
  "new_string": "receive"
}
```

**Update import**:
```json
{
  "file_path": "main.go",
  "old_string": "import \"old/package\"",
  "new_string": "import \"new/package\"",
  "replace_all": true
}
```

**Change function signature**:
```json
{
  "file_path": "api.go",
  "old_string": "func Process(data string) error {",
  "new_string": "func Process(ctx context.Context, data string) error {"
}
```

---

## Grep Tool

The Grep tool searches code using ripgrep - fast, powerful code search.

### Purpose

- Search code with regex patterns
- Filter by file type or glob pattern
- View context around matches
- Count occurrences

### Requirements

- **ripgrep must be installed**: `brew install ripgrep` (macOS) or `apt install ripgrep` (Linux)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `pattern` | string | ✅ | Regex pattern to search for |
| `path` | string | ❌ | Directory to search (default: current directory) |
| `output_mode` | string | ❌ | Output format: `content`, `files_with_matches`, `count` (default: `files_with_matches`) |
| `glob` | string | ❌ | Filter files by glob pattern (e.g., `*.go`) |
| `type` | string | ❌ | Filter by file type (e.g., `go`, `js`, `py`) |
| `-i` | boolean | ❌ | Case insensitive search |
| `-A` | number | ❌ | Lines of context after match (content mode only) |
| `-B` | number | ❌ | Lines of context before match (content mode only) |
| `-C` | number | ❌ | Lines of context around match (content mode only) |

### Output Modes

**`files_with_matches` (default)**:
- Shows only file paths containing matches
- Fast overview of where pattern appears
- Best for "where is X used?"

**`content`**:
- Shows matching lines with line numbers
- Supports context lines (`-A`, `-B`, `-C`)
- Best for seeing actual usage

**`count`**:
- Shows match count per file
- Best for statistics and coverage

### Approval

**Does NOT require approval** - grep is read-only.

### Examples

#### Find Function Definitions
```json
{
  "pattern": "func.*Test",
  "path": "internal/",
  "type": "go",
  "output_mode": "files_with_matches"
}
```

#### Search with Context
```json
{
  "pattern": "TODO",
  "glob": "*.go",
  "output_mode": "content",
  "-B": 2,
  "-A": 2
}
```

#### Case-Insensitive Search
```json
{
  "pattern": "error",
  "-i": true,
  "output_mode": "count"
}
```

#### Filter by File Type
```json
{
  "pattern": "import.*react",
  "type": "js",
  "output_mode": "files_with_matches"
}
```

#### Complex Regex
```json
{
  "pattern": "func \\w+\\(.*context\\.Context",
  "type": "go",
  "output_mode": "content"
}
```

### Error Handling

| Error | Reason | Solution |
|-------|--------|----------|
| ripgrep not installed | Binary not found | Install: `brew install ripgrep` |
| Invalid regex | Pattern syntax error | Check regex syntax |
| Path not found | Directory doesn't exist | Verify path |
| No matches | Pattern not found | Success with empty output |

### Output Format Examples

**files_with_matches**:
```
src/main.go
src/config.go
test/main_test.go
```

**content** (with line numbers):
```
src/main.go:15:func main() {
src/main.go:23:	return nil
src/config.go:8:const MaxRetries = 3
```

**count**:
```
src/main.go:5
src/config.go:2
test/main_test.go:8
```

### Common Patterns

**Find all TODOs**:
```json
{
  "pattern": "TODO|FIXME|HACK",
  "output_mode": "content",
  "-C": 1
}
```

**Find error handling**:
```json
{
  "pattern": "if err != nil",
  "type": "go",
  "output_mode": "count"
}
```

**Find API endpoints**:
```json
{
  "pattern": "router\\.(GET|POST|PUT|DELETE)",
  "type": "go",
  "output_mode": "content"
}
```

---

## Glob Tool

The Glob tool finds files by pattern matching - Claude's way to discover files.

### Purpose

- Find files by glob patterns
- Support recursive matching (`**`)
- Support brace expansion (`*.{ts,tsx}`)
- Sorted by modification time (newest first)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `pattern` | string | ✅ | Glob pattern (e.g., `*.go`, `**/*.test.js`) |
| `path` | string | ❌ | Directory to search (default: current directory) |

### Pattern Syntax

**Simple patterns**:
- `*.go` - All .go files in directory
- `test_*.py` - Files starting with test_
- `*.{js,ts}` - Files ending with .js OR .ts

**Recursive patterns**:
- `**/*.go` - All .go files in any subdirectory
- `src/**/*.tsx` - All .tsx files under src/
- `**/test/*.js` - All .js files in any test/ directory

**Brace expansion**:
- `*.{ts,tsx}` expands to `*.ts` and `*.tsx`
- `{foo,bar}/*.go` expands to `foo/*.go` and `bar/*.go`

### Approval

**Does NOT require approval** - glob is read-only.

### Examples

#### Find All Go Files
```json
{
  "pattern": "*.go",
  "path": "internal/"
}
```

#### Recursive Test Files
```json
{
  "pattern": "**/*_test.go"
}
```

#### Multiple Extensions
```json
{
  "pattern": "*.{ts,tsx}",
  "path": "src/components"
}
```

#### Specific Directory Pattern
```json
{
  "pattern": "src/**/*.test.js"
}
```

#### Find Config Files
```json
{
  "pattern": "**/{config,settings}.{json,yaml,yml}"
}
```

### Output Format

Results are sorted by modification time (newest first):

```
src/components/Button.tsx
src/components/Input.tsx
src/App.tsx
src/index.tsx
```

### Error Handling

| Error | Reason | Solution |
|-------|--------|----------|
| Invalid pattern | Syntax error | Check glob syntax |
| Path not found | Directory doesn't exist | Verify path |
| No matches | Pattern matched nothing | Success with empty output |

### Common Patterns

**Find all source files**:
```json
{
  "pattern": "**/*.{go,js,ts,py}"
}
```

**Find test files**:
```json
{
  "pattern": "**/*{_test.go,.test.js,.spec.ts}"
}
```

**Find config files**:
```json
{
  "pattern": "**/*.{json,yaml,yml,toml}"
}
```

**Find modified components**:
```json
{
  "pattern": "src/components/**/*.tsx"
}
```

### Comparison with Grep

Use **Glob** when you need to:
- Find files by name/extension
- Get newest files first
- Work with file paths (not content)

Use **Grep** when you need to:
- Search file contents
- Use regex patterns
- See actual code context

---

## Safety Features

### Overall Security Model

**Principle**: User has full control, always

**Implementation**:
1. **Explicit approval** for dangerous operations
2. **Detailed prompts** showing exactly what will happen
3. **No persistence** of approval decisions
4. **Atomic operations** where possible (write tool)
5. **Timeouts** to prevent hung processes
6. **Validation** before execution

### Permission Levels

**Auto-approved** (safe operations):
- Reading normal files
- Creating new files
- Appending to files
- Safe bash commands (ls, grep, etc.)

**Requires approval** (potentially dangerous):
- Reading sensitive files (/etc, ~/.ssh)
- Overwriting existing files
- Dangerous bash commands (rm -rf, sudo)
- Long-running commands (>60s timeout)

**Never allowed** (blocked):
- None currently (user can approve anything)
- Future: Could add blocklist for extremely dangerous ops

### Audit Trail

**Current**:
- Tool requests logged to console
- Approval/denial visible in UI
- Results saved to conversation history

**Future** (v0.3.0+):
- Tool execution log file
- Approval history
- Replay protection

## Examples

### Example 1: Code Review

**User**: "Review main.go and suggest improvements"

**Flow**:
1. Claude requests `read_file(main.go)`
2. User approves
3. Claude reads code
4. Claude analyzes and provides feedback
5. User: "Apply those changes"
6. Claude requests `write_file(main.go, improved_code, overwrite)`
7. User approves overwrite
8. File updated

### Example 2: Project Setup

**User**: "Set up a new Go project with basic structure"

**Flow**:
1. Claude requests `bash(mkdir -p cmd/app internal pkg)`
2. Auto-approved (safe command)
3. Claude requests `write_file(cmd/app/main.go, boilerplate_code, create)`
4. Auto-approved (create mode)
5. Claude requests `write_file(go.mod, module_config, create)`
6. Auto-approved
7. Structure created

### Example 3: Log Analysis

**User**: "Check the error logs and summarize issues"

**Flow**:
1. Claude requests `bash(tail -100 /var/log/app.log)`
2. Approval required (/var/log is sensitive)
3. User approves
4. Claude reads logs
5. Claude requests `bash(grep ERROR /var/log/app.log | wc -l)`
6. Auto-approved (read-only)
7. Claude summarizes findings

### Example 4: Bulk File Operations

**User**: "Find all TODOs in Go files and list them"

**Flow**:
1. Claude requests `bash(find . -name "*.go" -exec grep -Hn TODO {} \;)`
2. Auto-approved (safe command)
3. Returns list of TODOs
4. Claude formats and presents results

### Example 5: Configuration Update

**User**: "Add a new database connection to config.yaml"

**Flow**:
1. Claude requests `read_file(config.yaml)`
2. Auto-approved
3. Claude reads current config
4. Claude generates updated config
5. Claude requests `write_file(config.yaml, updated_config, overwrite)`
6. Approval required (overwrite)
7. User approves
8. Config updated

## Troubleshooting

### Tool Not Available

**Symptom**: "Tool not found: <tool_name>"

**Cause**: Tool not registered

**Solution**: Check tool name spelling, verify Clem version supports tool

### Permission Denied

**Symptom**: "Permission denied" in tool result

**Cause**: Insufficient filesystem or execution permissions

**Solution**:
- Check file/directory permissions
- For bash: verify command is in PATH
- For write: check directory is writable

### Timeout Issues

**Symptom**: "Command timeout after Xs"

**Cause**: Command takes longer than timeout

**Solution**:
- For bash: increase timeout parameter
- Optimize command (e.g., limit grep search scope)
- Break into smaller operations

### Large File Issues

**Symptom**: "File too large" or "Out of memory"

**Cause**: File exceeds 10MB limit

**Solution**:
- Use bash tool with `head -n 100 largefile.txt`
- Process file in chunks
- Use specialized tools (grep, awk) via bash

### Approval Not Working

**Symptom**: Tool denied even after typing "y"

**Cause**: Typo in approval response

**Solution**:
- Type exactly `y` or `yes`
- Check for extra spaces
- Press Enter after typing

### Binary File Errors

**Symptom**: "Invalid UTF-8 encoding"

**Cause**: Trying to read binary file with read_file

**Solution**:
- Use bash tool with `hexdump` or `xxd`
- Use `file` command to check file type first
- Use appropriate binary tools

---

## Advanced Topics

### Tool Composition

Claude can chain multiple tools:

```
User: "Find all Python files, read each one, and create a summary document"

Flow:
1. bash: find . -name "*.py"
2. read_file: file1.py
3. read_file: file2.py
4. ...
5. write_file: SUMMARY.md (with aggregated analysis)
```

### Custom Tool Timeout

Adjust timeout for long-running operations:

```
User: "Run the test suite (it takes 2 minutes)"

Claude: bash(
  command: "go test ./...",
  timeout: 150  # 2.5 minutes
)
```

### Working Directory Control

Execute commands in specific directories:

```
Claude: bash(
  command: "git status",
  working_dir: "/path/to/repo"
)
```

---

**See Also**:
- [USER_GUIDE.md](USER_GUIDE.md) - General usage guide
- [ARCHITECTURE.md](ARCHITECTURE.md) - Tool system architecture
- [CHANGELOG.md](../CHANGELOG.md) - Version history
