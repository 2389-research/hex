# Hex v0.2.0 Release Notes

**Release Date:** November 27, 2025

## Overview

Hex v0.2.0 is a major feature release that introduces **interactive mode** with full terminal UI, **conversation persistence**, and **tool execution capabilities**. This release transforms Hex from a simple print-mode CLI into a full-featured AI assistant with Claude's tool-using capabilities.

## Highlights

### Interactive Terminal UI

Launch Hex without flags to enter a beautiful, full-featured terminal interface:

```bash
hex
```

Features include:
- **Streaming responses** - See Claude's thoughts in real-time as they're generated
- **Markdown rendering** - Syntax highlighted code blocks and formatted text using Glamour
- **Vim-style navigation** - j/k to scroll, gg/G to jump to top/bottom, / to search
- **Status indicators** - Visual feedback for Idle, Streaming, and Error states
- **Token tracking** - Real-time input and output token counters
- **Multiple views** - Chat mode, conversation history, and tools inspector

### Conversation Persistence

Never lose your work again. All conversations are automatically saved to SQLite:

```bash
# Resume your last conversation
hex --continue

# Resume a specific conversation by ID
hex --resume conv-1234567890

# Database location: ~/.hex/hex.db
```

Features:
- Automatic conversation titling based on first message
- Full message history with timestamps
- Efficient indexing for fast lookups
- WAL mode for better concurrency

### Tool Execution

Claude can now interact with your system through three powerful tools:

**Read Tool** - Safe file reading
```bash
# Claude can read files with your approval
"Can you read config.yaml and explain it?"
```

**Write Tool** - Create and modify files
```bash
# Claude can write files with confirmation
"Create a new README.md with project description"
```

**Bash Tool** - Execute shell commands
```bash
# Claude can run commands with sandboxing
"List all Go files in the current directory"
```

All tools require user approval for dangerous operations and include:
- Timeout protection (configurable, max 5 minutes)
- Path validation and safety checks
- Real-time output streaming
- Detailed error messages

## What's New

### For Users

- **Interactive Mode**: Launch with just `hex` for rich terminal UI
- **Conversation History**: Resume past conversations seamlessly
- **Tool Approval System**: Explicit control over file/command operations
- **Better Visualization**: Markdown rendering, syntax highlighting, and formatted output
- **Search**: Find text in conversation history with / key
- **Responsive UI**: Handles window resizing gracefully

### For Developers

- **Storage Layer**: SQLite with hybrid schema (normalized + JSON)
- **Streaming API**: SSE parser with delta accumulation
- **Tool Framework**: Extensible registry and executor pattern
- **Comprehensive Tests**: Unit, integration, and example-based testing
- **Clean Architecture**: Separation of concerns across packages

## Upgrade Guide

### Installation

```bash
# Install or upgrade to v0.2.0
go install github.com/harper/hex/cmd/hex@v0.2.0

# Verify installation
hex doctor
```

### Migration

No migration needed! v0.2.0 is fully backward compatible with v0.1.0:

- Existing config files work unchanged
- Print mode (`--print`) still available
- All v0.1.0 commands work identically

New features are opt-in:
- Interactive mode: just run `hex` without flags
- Conversation persistence: automatic on first use
- Tools: activated when Claude requests them

### Configuration

Your existing configuration works as-is. New optional settings:

```yaml
# ~/.hex/config.yaml
api_key: sk-ant-api03-...
model: claude-sonnet-4-5-20250929

# New optional settings (not required)
database_path: ~/.hex/hex.db  # default location
tool_timeout: 30                 # seconds, default 30
```

## Breaking Changes

**None!** v0.2.0 is fully backward compatible with v0.1.0.

## Known Issues

1. **Tool Result Persistence**: Tool execution results are not yet saved to the database (planned for v0.3.0)
2. **File Size Limits**: Reading very large files (>10MB) may hit memory limits
3. **Tool Queueing**: Tools execute one at a time; no parallel execution yet
4. **Search Highlighting**: Search matches are not yet highlighted in viewport (coming soon)

Workarounds are documented in the user guide.

## Performance

- Streaming responses provide instant feedback (vs. waiting for full response)
- WAL mode enables concurrent database access
- Efficient SSE parsing with minimal overhead
- Tool execution timeout prevents hung commands

## Security

All tool operations include safety features:

- **Read Tool**: Approval required for sensitive paths (/etc, ~/.ssh, etc.)
- **Write Tool**: Confirmation required for overwriting existing files
- **Bash Tool**: Timeout limits, dangerous command detection, no shell expansion

Users maintain full control with explicit approval prompts.

## Documentation

Comprehensive documentation added:

- [USER_GUIDE.md](docs/USER_GUIDE.md) - Complete usage guide
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System design and internals
- [TOOLS.md](docs/TOOLS.md) - Tool system reference
- [CHANGELOG.md](CHANGELOG.md) - Full change history

## What's Next

Phase 3 roadmap (v0.3.0 planned):
- Extended tool support (Edit, Grep, Glob)
- MCP (Model Context Protocol) integration
- Plugin system for custom tools
- Tool result persistence
- Multi-tool execution queueing
- Advanced debugging features

## Acknowledgments

Built with these excellent libraries:
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [modernc.org/sqlite](https://modernc.org/sqlite) - Pure Go SQLite

## Feedback

Found a bug? Have a feature request?

- Open an issue on GitHub
- Check existing documentation
- Review the troubleshooting guide in USER_GUIDE.md

---

Thank you for using Hex! This release represents a major milestone in bringing Claude's capabilities to the terminal.

**Download:** `go install github.com/harper/hex/cmd/hex@v0.2.0`

**Full Changelog:** [CHANGELOG.md](CHANGELOG.md)
