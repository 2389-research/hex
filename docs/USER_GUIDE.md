# Clem User Guide

Complete guide to using Clem, the Go-based Claude CLI with interactive mode and tool execution.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Quick Start](#quick-start)
- [Interactive Mode](#interactive-mode)
- [Conversation Management](#conversation-management)
- [Tool System](#tool-system)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

## Installation

### Prerequisites

- Go 1.24 or later
- Anthropic API key (get one at [console.anthropic.com](https://console.anthropic.com))

### Install Clem

```bash
# Install latest version
go install github.com/harper/clem/cmd/clem@latest

# Or install specific version
go install github.com/harper/clem/cmd/clem@v0.2.0

# Verify installation
clem --version
```

### Build from Source

```bash
# Clone repository
git clone https://github.com/harper/clem.git
cd clem

# Build
make build

# Or run without building
make run -- --help
```

## Configuration

### API Key Setup

Three ways to configure your API key:

**Option 1: Command (Recommended)**
```bash
clem setup-token sk-ant-api03-...
```

**Option 2: Environment Variable**
```bash
export ANTHROPIC_API_KEY=sk-ant-api03-...
```

**Option 3: Config File**

Create `~/.clem/config.yaml`:
```yaml
api_key: sk-ant-api03-...
model: claude-sonnet-4-5-20250929
```

**Option 4: .env File**

Create `.env` in your project directory:
```bash
ANTHROPIC_API_KEY=sk-ant-api03-...
JEFF_MODEL=claude-sonnet-4-5-20250929
```

### Configuration Priority

Configuration is loaded in this order (later overrides earlier):

1. Config file (`~/.clem/config.yaml`)
2. Environment variables
3. .env file in current directory
4. Command-line flags

### Available Settings

```yaml
# ~/.clem/config.yaml

# Required
api_key: sk-ant-api03-...

# Optional (with defaults)
model: claude-sonnet-4-5-20250929
database_path: ~/.clem/clem.db
tool_timeout: 30  # seconds
max_tokens: 4096
temperature: 1.0
```

### Health Check

Verify your configuration:

```bash
clem doctor
```

This checks:
- API key configuration
- Database accessibility
- Network connectivity
- Model availability

## Quick Start

### Print Mode (Non-Interactive)

Simple one-off questions without entering interactive mode:

```bash
# Basic usage
clem --print "What is the capital of France?"

# With different model
clem --print --model claude-opus-4-5-20250929 "Explain quantum computing"

# JSON output
clem --print --output-format json "List 3 programming languages"
```

### Interactive Mode

Launch the terminal UI:

```bash
# Start new conversation
clem

# Start with initial prompt
clem "Help me debug this code"

# Resume last conversation
clem --continue

# Resume specific conversation
clem --resume conv-1234567890
```

## Interactive Mode

### Overview

Interactive mode provides a full-featured terminal UI with:
- Real-time streaming responses
- Markdown formatting with syntax highlighting
- Conversation history
- Tool execution with approval
- Search functionality
- Multiple view modes

### UI Layout

```
┌─────────────────────────────────────────────────┐
│ Clem • claude-sonnet-4-5-20250929              │  ← Title bar
│                                                 │
│ ┌─────────────────────────────────────────────┐ │
│ │                                             │ │
│ │  Conversation viewport                      │ │  ← Messages
│ │  (scrollable with j/k)                      │ │
│ │                                             │ │
│ └─────────────────────────────────────────────┘ │
│                                                 │
│ ┌─────────────────────────────────────────────┐ │
│ │ Your message here...                        │ │  ← Input area
│ └─────────────────────────────────────────────┘ │
│                                                 │
│ Status: Idle | Tokens: 1234 in, 5678 out      │  ← Status bar
│ ctrl+c: quit • enter: send • /: search         │  ← Help text
└─────────────────────────────────────────────────┘
```

### Sending Messages

1. Type your message in the input area (bottom section)
2. Press `Enter` to send (without modifiers)
3. Use `Alt+Enter` for newlines without sending
4. Watch the streaming response appear in real-time

### Markdown Rendering

Clem renders markdown with full formatting:

**Code blocks** with syntax highlighting:
````
```python
def hello():
    print("Hello, world!")
```
````

**Formatting**:
- **Bold**, *italic*, `inline code`
- Lists (bulleted and numbered)
- Headers, blockquotes, links

**Tables**, horizontal rules, and more.

### Streaming

Responses stream in real-time:
- See text as Claude generates it
- Status indicator shows "Streaming..."
- Token counters update live
- Cancel with `Ctrl+C` if needed

## Conversation Management

### Creating Conversations

Every time you run `clem` without `--continue` or `--resume`, a new conversation starts.

Conversations are automatically titled based on your first message.

### Resuming Conversations

**Resume last conversation:**
```bash
clem --continue
```

**Resume specific conversation:**
```bash
# List conversations first (in interactive mode, switch to History view)
clem --resume conv-1234567890
```

### Conversation Storage

All conversations stored in SQLite at `~/.clem/clem.db`:
- Automatically created on first use
- Indexed for fast retrieval
- WAL mode for better concurrency
- Survives across sessions

### Finding Conversations

In interactive mode:
1. Press `h` to switch to History view
2. Browse past conversations
3. Press `Enter` to resume selected conversation

## Tool System

Claude can execute three types of tools with your approval:

### Read Tool

**Purpose**: Read file contents safely

**Example request:**
```
"Can you read config.yaml and explain what it does?"
```

**Approval prompt:**
```
Tool: read_file
Path: /path/to/config.yaml

Approve execution? [y/N]
```

**Safety features:**
- Approval required for sensitive paths (`/etc`, `~/.ssh`, etc.)
- File size limits (default 10MB)
- UTF-8 validation
- Path validation (no parent directory traversal)

### Write Tool

**Purpose**: Create or modify files

**Three modes:**

**Create** - Create new file (fails if exists)
```
"Create a new README.md with project description"
```

**Overwrite** - Replace existing file (requires confirmation)
```
"Update the config.yaml with new settings"
```

**Append** - Add to end of file
```
"Add a new section to the documentation"
```

**Approval prompt:**
```
Tool: write_file
Path: /path/to/file.txt
Mode: overwrite
Size: 1234 bytes

Approve execution? [y/N]
```

**Safety features:**
- Confirmation required for overwrites
- Atomic writes using temp files
- Directory creation if needed
- Content validation

### Bash Tool

**Purpose**: Execute shell commands

**Example request:**
```
"List all Go files in the current directory"
```

**Approval prompt:**
```
Tool: bash
Command: find . -name "*.go"
Working dir: /current/dir
Timeout: 30s

Approve execution? [y/N]
```

**Safety features:**
- Timeout protection (default 30s, max 5min)
- Dangerous command detection (`rm -rf`, `sudo`, etc.)
- Working directory control
- Exit code capture
- Real-time output streaming

**Dangerous commands** require approval:
- File deletion (`rm -rf`)
- System modification (`sudo`)
- Network operations (`curl`, `wget`)
- Process manipulation (`kill`, `pkill`)

### Tool Approval

**Approving tools:**
- Type `y` or `yes` to approve
- Press `Enter` or `n` to deny
- Approval is per-execution (not remembered)

**Tool execution flow:**
1. Claude requests tool
2. Clem displays approval prompt with details
3. You approve or deny
4. Tool executes (if approved)
5. Results shown in conversation
6. Claude continues based on results

## Keyboard Shortcuts

### Global

| Key | Action |
|-----|--------|
| `Ctrl+C` | Quit Clem |
| `Esc` | Cancel current operation |
| `?` | Show help (context-aware) |

### Navigation (Chat View)

| Key | Action |
|-----|--------|
| `j` | Scroll down one line |
| `k` | Scroll up one line |
| `d` | Scroll down half page |
| `u` | Scroll up half page |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `Ctrl+D` | Scroll down page |
| `Ctrl+U` | Scroll up page |

### Input

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Alt+Enter` | Insert newline |
| `Ctrl+W` | Delete word backward |
| `Ctrl+U` | Delete to start of line |

### Search

| Key | Action |
|-----|--------|
| `/` | Enter search mode |
| `Enter` | Execute search |
| `Esc` | Exit search |
| `n` | Next match (planned) |
| `N` | Previous match (planned) |

### View Modes

| Key | Action |
|-----|--------|
| `c` | Chat view (default) |
| `h` | History view |
| `t` | Tools inspector |
| `Tab` | Cycle views |

## Advanced Usage

### Custom Models

Use different Claude models:

```bash
# Via flag
clem --model claude-opus-4-5-20250929

# Via config
echo "model: claude-opus-4-5-20250929" >> ~/.clem/config.yaml
```

Available models:
- `claude-sonnet-4-5-20250929` (default, balanced)
- `claude-opus-4-5-20250929` (most capable)
- `claude-haiku-4-5-20250929` (fastest)

### Token Limits

Control maximum response length:

```bash
# Via flag
clem --max-tokens 8192

# Via config
echo "max_tokens: 8192" >> ~/.clem/config.yaml
```

Defaults:
- Print mode: 4096 tokens
- Interactive mode: 4096 tokens
- Maximum: Model-dependent (check API docs)

### Temperature

Adjust response randomness (0.0 to 1.0):

```bash
# Via flag
clem --temperature 0.5

# Via config
echo "temperature: 0.5" >> ~/.clem/config.yaml
```

- `0.0`: Deterministic, focused
- `1.0`: Creative, varied (default)

### Database Management

**Location**: `~/.clem/clem.db`

**Backup:**
```bash
cp ~/.clem/clem.db ~/.clem/clem.db.backup
```

**Reset:**
```bash
rm ~/.clem/clem.db
# Database recreated on next run
```

**Inspect:**
```bash
sqlite3 ~/.clem/clem.db
> .schema
> SELECT * FROM conversations;
```

### Batch Processing

Process multiple files using tools:

```
"Read all .go files in the current directory and summarize them"

"Find all TODO comments in the codebase and list them"

"Check all markdown files for broken links"
```

Claude will request appropriate tools sequentially.

## Troubleshooting

### Common Issues

**API Key Not Found**

```
Error: API key not configured
```

**Solution:**
```bash
clem setup-token sk-ant-api03-...
# Or verify: clem doctor
```

**Database Locked**

```
Error: database is locked
```

**Solution:**
```bash
# Close other Clem instances
# Or check for stale connections:
lsof ~/.clem/clem.db
```

**Tool Timeout**

```
Tool execution timeout after 30s
```

**Solution:**
```bash
# Increase timeout in config:
echo "tool_timeout: 120" >> ~/.clem/config.yaml
```

**Large File Read Failure**

```
Error: file too large (>10MB)
```

**Solution:**
- Read file in chunks
- Use bash tool with `head` or `tail`
- Process files programmatically

### Debug Mode

Enable verbose logging:

```bash
# Set environment variable
export JEFF_DEBUG=1
clem

# Check logs
tail -f ~/.clem/debug.log
```

### Network Issues

**Proxy configuration:**
```bash
export HTTP_PROXY=http://proxy:8080
export HTTPS_PROXY=http://proxy:8080
```

**Timeout issues:**
Check your network:
```bash
curl -I https://api.anthropic.com
```

### Performance Tips

1. **Streaming**: Keep enabled for better responsiveness
2. **Token limits**: Lower max_tokens for faster responses
3. **Database**: WAL mode is automatic (no tuning needed)
4. **Tools**: Approve quickly to avoid blocking Claude

### Getting Help

1. **Check documentation**: `docs/` directory
2. **Run doctor**: `clem doctor`
3. **Check logs**: Debug mode logs to `~/.clem/debug.log`
4. **Report issues**: GitHub issues with reproduction steps

## Tips and Tricks

### Productivity

**Multiline input:**
- Use `Alt+Enter` for newlines
- Paste code blocks directly
- Format with markdown in your message

**Conversation hygiene:**
- Start new conversations for unrelated topics
- Use `--continue` to build on context
- Descriptive first messages help auto-titling

**Tool efficiency:**
- Approve tools quickly to maintain flow
- Deny if unsure (you can rephrase and try again)
- Check tool parameters before approving

### Workflow Examples

**Code review:**
```
"Read main.go and suggest improvements"
# Approve read tool
# Claude analyzes and suggests
"Apply those changes to main.go"
# Approve write tool
```

**Documentation:**
```
"Read all files in cmd/ and create API documentation"
# Approve multiple read tools
# Claude generates docs
"Write the documentation to docs/API.md"
# Approve write tool
```

**Debugging:**
```
"Run the test suite and analyze failures"
# Approve bash tool
# Claude reads output
"The error is in auth.go line 42. Read that file"
# Approve read tool
"Fix the bug"
# Approve write tool
```

### Configuration Profiles

Different configs for different projects:

```bash
# Personal project
cd ~/personal
echo "ANTHROPIC_API_KEY=sk-..." > .env
clem

# Work project
cd ~/work
echo "ANTHROPIC_API_KEY=sk-work-..." > .env
echo "JEFF_MODEL=claude-opus-4-5-20250929" >> .env
clem
```

---

**Next Steps:**
- Explore [ARCHITECTURE.md](ARCHITECTURE.md) for system design
- Read [TOOLS.md](TOOLS.md) for tool system details
- Check [CHANGELOG.md](../CHANGELOG.md) for version history

**Questions?** Open an issue on GitHub or check the documentation.
