# Clem - Claude CLI

[![Test](https://github.com/harper/clem/workflows/Test/badge.svg)](https://github.com/harper/clem/actions/workflows/test.yml)
[![Release](https://github.com/harper/clem/workflows/Release/badge.svg)](https://github.com/harper/clem/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/harper/clem)](https://goreportcard.com/report/github.com/harper/clem)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go implementation of Claude Code CLI with interactive mode and tool execution capabilities.

**Latest Version**: v0.5.0

## Features

### Phase 1 (v0.1.0) - Foundation
- ✅ Print mode (non-interactive)
- ✅ Configuration management
- ✅ API client with Messages API
- ✅ Basic commands (setup, doctor)

### Phase 2 (v0.2.0) - Interactive Mode
- ✅ **Interactive TUI** with Bubbletea
- ✅ **Streaming responses** with progressive rendering
- ✅ **Conversation persistence** in SQLite
- ✅ **Tool execution** (Read, Write, Bash)
- ✅ **Advanced UI** features (glamour, vim navigation, search)
- ✅ **Conversation resumption** (--continue, --resume)

### Phase 3 (v0.3.0) - Extended Tools & MCP
- ✅ **Extended tools** (Edit, Grep, Glob, AskUserQuestion, TodoWrite, WebFetch, WebSearch, Task, BashOutput, KillShell)
- ✅ **MCP (Model Context Protocol) integration** - Use external tools from MCP servers
- ✅ **MCP CLI commands** - Add, list, and remove MCP servers
- ✅ **stdio transport** - Connect to MCP servers via standard I/O

### Coming in Phase 4
- [ ] HTTP/SSE transport for MCP servers
- [ ] MCP resources and prompts support
- [ ] Plugin system
- [ ] Advanced debugging features

## Quick Start

### Installation

**Method 1: Install Script (Recommended)**

```bash
# macOS and Linux
curl -sSL https://raw.githubusercontent.com/harper/clem/main/install.sh | bash

# Verify installation
clem --version
```

**Method 2: Homebrew (macOS/Linux)**

```bash
# Add tap and install
brew install harper/tap/clem

# Verify installation
clem --version
```

**Method 3: Go Install**

```bash
# Requires Go 1.24+
go install github.com/harper/clem/cmd/clem@latest

# Verify installation
clem --version
```

**Method 4: Download Binary**

Download pre-built binaries from the [releases page](https://github.com/harper/clem/releases):

1. Download the archive for your platform
2. Extract the binary
3. Move to a directory in your PATH (e.g., `/usr/local/bin`)
4. Run `clem --version` to verify

**Method 5: Build from Source**

```bash
# Clone repository
git clone https://github.com/harper/clem.git
cd clem

# Build and install
make install

# Verify installation
clem --version
```

### Setup

```bash
# Configure API key
clem setup-token sk-ant-api03-...

# Verify configuration
clem doctor
```

### Usage

**Interactive Mode** (full TUI):
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

**Print Mode** (quick one-off):
```bash
# Simple query
clem --print "What is the capital of France?"

# With JSON output
clem --print --output-format json "List 3 programming languages"
```

## What's New in v0.2.0

### Interactive Terminal UI
Beautiful, full-featured terminal interface:
- 📝 **Streaming responses** - See Claude's thoughts in real-time
- 🎨 **Markdown rendering** - Syntax highlighting and formatted text
- ⌨️ **Vim-style navigation** - j/k to scroll, gg/G to jump, / to search
- 📊 **Token tracking** - Real-time input and output token counters
- 🎯 **Multiple views** - Chat, History, and Tools inspector

### Conversation Persistence
Never lose your work:
- 💾 All conversations saved to `~/.clem/clem.db`
- 🔄 Resume with `--continue` or `--resume`
- 🏷️ Automatic conversation titling
- 📅 Full message history with timestamps

### Tool Execution
Claude can interact with your system:

**Read Tool** - Safe file reading
```bash
"Can you read config.yaml and explain it?"
```

**Write Tool** - Create and modify files
```bash
"Create a new README.md with project description"
```

**Bash Tool** - Execute shell commands
```bash
"List all Go files in the current directory"
```

All tools require user approval for dangerous operations.

## Interactive Mode Features

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Alt+Enter` | Insert newline |
| `j` / `k` | Scroll down/up |
| `gg` / `G` | Jump to top/bottom |
| `/` | Search |
| `Ctrl+C` | Quit |

### Tool Approval

When Claude requests a tool, you'll see:
```
┌─────────────────────────────────────────┐
│ Tool Execution Request                  │
│                                         │
│ Tool: read_file                         │
│ Path: /path/to/file.txt                 │
│                                         │
│ Approve? [y/N]                          │
└─────────────────────────────────────────┘
```

Type `y` to approve, `n` (or Enter) to deny.

### Status Indicators

- **Idle**: Ready for your input
- **Streaming**: Claude is responding
- **Error**: Something went wrong (check message)

## Configuration

Three ways to configure:

**1. Config file** (`~/.clem/config.yaml`):
```yaml
api_key: sk-ant-api03-...
model: claude-sonnet-4-5-20250929
```

**2. Environment variables**:
```bash
export ANTHROPIC_API_KEY=sk-ant-api03-...
export CLEM_MODEL=claude-sonnet-4-5-20250929
```

**3. `.env` file** (in project directory):
```bash
ANTHROPIC_API_KEY=sk-ant-api03-...
CLEM_MODEL=claude-sonnet-4-5-20250929
```

## Tools

### Built-in Tools

**Core Tools**:
- **Read** - Safely read file contents (sensitive path detection)
- **Write** - Create, overwrite, or append to files (atomic writes)
- **Bash** - Execute shell commands (timeout protection, dangerous command detection)
- **Edit** - Replace exact strings in files (single or bulk replacement)
- **Grep** - Search code with ripgrep (regex patterns, file filtering)
- **Glob** - Find files by pattern (recursive matching, brace expansion)

**Advanced Tools**:
- **AskUserQuestion** - Interactive multiple-choice questions
- **TodoWrite** - Task list management with progress tracking
- **WebFetch** - Fetch web content via HTTP GET
- **WebSearch** - Search the web via DuckDuckGo
- **Task** - Launch sub-agents for complex tasks
- **BashOutput** - Monitor background process output
- **KillShell** - Terminate background processes

### MCP Integration

**Extend Clem with MCP servers** - Use external tools from the Model Context Protocol ecosystem:

```bash
# Add an MCP server
clem mcp add filesystem npx -y @modelcontextprotocol/server-filesystem ~/Documents

# List configured servers
clem mcp list

# MCP tools are automatically available in conversations
clem
> "List all markdown files in my Documents directory"
```

**Official MCP Servers**:
- **@modelcontextprotocol/server-filesystem** - File operations (read, write, list, search)
- **@modelcontextprotocol/server-fetch** - HTTP requests and web scraping
- **@modelcontextprotocol/server-sqlite** - SQLite database queries
- **@modelcontextprotocol/server-postgres** - PostgreSQL database access

**Create Custom Servers**: Build your own MCP servers in Node.js, Python, or any language

**See [TOOLS.md](docs/TOOLS.md) for complete tool documentation and [MCP_INTEGRATION.md](docs/MCP_INTEGRATION.md) for MCP architecture details.**

## Project Status

| Phase | Status | Version | Description |
|-------|--------|---------|-------------|
| Phase 1 | ✅ Complete | v0.1.0 | Foundation (print mode, config, API) |
| Phase 2 | ✅ Complete | v0.2.0 | Interactive mode, core tools, storage |
| Phase 3 | ✅ Complete | v0.3.0 | Extended tools, MCP integration |
| Phase 4 | 📋 Planned | v0.4.0 | HTTP transport, MCP resources/prompts |
| Phase 5 | 📋 Planned | v0.5.0 | Plugin system, advanced debugging |

## Documentation

- **[USER_GUIDE.md](docs/USER_GUIDE.md)** - Complete usage guide
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System design and internals
- **[TOOLS.md](docs/TOOLS.md)** - Tool system reference (including MCP)
- **[MCP_INTEGRATION.md](docs/MCP_INTEGRATION.md)** - MCP architecture and server development
- **[examples/mcp/](examples/mcp/)** - MCP server examples and configurations
- **[CHANGELOG.md](CHANGELOG.md)** - Version history
- **[RELEASE_NOTES.md](RELEASE_NOTES.md)** - v0.2.0 release notes

## Development

### Build

```bash
# Clone repository
git clone https://github.com/harper/clem.git
cd clem

# Build
make build

# Run without building
make run -- --help
```

### Test

```bash
# Unit tests (fast)
make test

# All tests including integration
go test ./...

# With coverage
go test -cover ./...
```

### Project Structure

```
clem/
├── cmd/clem/           # CLI entry point
├── internal/           # Private implementation
│   ├── core/          # API client, types, config
│   ├── ui/            # Bubbletea TUI
│   ├── storage/       # SQLite persistence
│   └── tools/         # Tool system (Read, Write, Bash)
├── docs/              # Documentation
└── test/              # Integration tests
```

## Testing Philosophy

We use real components and avoid mocks:

- **Unit tests**: Fast, isolated logic tests
- **Integration tests**: End-to-end workflows with real filesystem
- **VCR cassettes**: Record/replay real API calls
- **Example-based tests**: Documentation + validation

**No mock mode**: We always use real data and real APIs.

## Architecture Highlights

- **Go 1.24+** for single binary distribution
- **Bubbletea** for terminal UI (Elm Architecture)
- **SQLite** for conversation persistence (hybrid schema)
- **SSE streaming** for real-time responses
- **Registry pattern** for extensible tool system

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed design.

## Performance

- **Streaming responses**: Instant feedback vs waiting for full response
- **WAL mode**: Efficient SQLite concurrency
- **Efficient SSE parsing**: Minimal overhead
- **Tool timeouts**: Prevents hung commands

## Security

All tool operations include safety features:

- **Read Tool**: Approval for sensitive paths
- **Write Tool**: Confirmation for overwrites
- **Bash Tool**: Timeout limits, dangerous command detection
- **No shell expansion**: Controlled execution environment
- **User approval**: Always required for dangerous ops

## Requirements

- Go 1.24 or later
- Anthropic API key ([get one here](https://console.anthropic.com))
- macOS, Linux, or Windows

## Contributing

Contributions welcome! Please:

1. Read the documentation
2. Follow the existing code style
3. Add tests for new features
4. Update documentation

## Troubleshooting

**API key not found**:
```bash
clem setup-token sk-ant-api03-...
clem doctor
```

**Database locked**:
```bash
# Close other Clem instances
lsof ~/.clem/clem.db
```

**Tool timeout**:
```yaml
# Increase in config
tool_timeout: 120
```

See [USER_GUIDE.md](docs/USER_GUIDE.md) for more troubleshooting.

## Roadmap

**v0.3.0** (Complete):
- ✅ Extended tools (Edit, Grep, Glob, AskUserQuestion, TodoWrite, WebFetch, WebSearch, Task, BashOutput, KillShell)
- ✅ MCP server integration (stdio transport)
- ✅ MCP CLI commands (add, list, remove)
- ✅ MCP tool adapter and registry

**v0.4.0** (Q1 2026):
- HTTP/SSE transport for MCP servers
- MCP resources support
- MCP prompts support
- Server lifecycle management
- Health checks and reconnection

**v0.5.0** (Q2 2026):
- Plugin system
- Custom tool registration
- Advanced debugging features
- Tool result persistence
- Multi-tool execution queueing

See [ROADMAP.md](ROADMAP.md) for detailed plans.

## Acknowledgments

Built with these excellent libraries:
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration
- [modernc.org/sqlite](https://modernc.org/sqlite) - Pure Go SQLite

## License

MIT

## Support

- **Documentation**: Check the `docs/` directory
- **Issues**: Open a GitHub issue
- **Questions**: See USER_GUIDE.md troubleshooting section

---

**Download**: `go install github.com/harper/clem/cmd/clem@latest`

**Documentation**: [docs/USER_GUIDE.md](docs/USER_GUIDE.md)

**Changelog**: [CHANGELOG.md](CHANGELOG.md)
