# Hex - Claude CLI

<p align="center">
  <img src="docs/hex.png" alt="Hex Logo" width="400"/>
</p>

[![Test](https://github.com/2389-research/hex/workflows/Test/badge.svg)](https://github.com/2389-research/hex/actions/workflows/test.yml)
[![Release](https://github.com/2389-research/hex/workflows/Release/badge.svg)](https://github.com/2389-research/hex/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/2389-research/hex)](https://goreportcard.com/report/github.com/2389-research/hex)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go implementation of Claude Code CLI with interactive mode and tool execution capabilities.

**Latest Version**: v1.0.0

## Features

Hex v1.0 is a production-ready Claude CLI with comprehensive tool support, interactive TUI, and MCP integration.

### Core Features
- ✅ **Print Mode** - Non-interactive command-line queries
- ✅ **Interactive TUI** - Full-featured Bubbletea interface with streaming
- ✅ **Conversation Persistence** - SQLite storage with resume support
- ✅ **Tool System** - 11 built-in tools + extensible via MCP
- ✅ **MCP Integration** - Model Context Protocol for external tools
- ✅ **Structured Logging** - JSON and text formats with multiple log levels
- ✅ **CI/CD Pipeline** - GitHub Actions with comprehensive quality checks
- ✅ **Multi-Platform** - macOS, Linux, Windows support

### Built-in Tools (11 Total)
**Core**: Read, Write, Bash, Edit, Grep, Glob
**Advanced**: AskUserQuestion, TodoWrite, WebFetch, WebSearch, Task
**Process Management**: BashOutput, KillShell

### Distribution Channels (6)
1. Homebrew (macOS/Linux)
2. Install scripts (curl/PowerShell)
3. Docker images (GHCR)
4. Binary releases (GitHub)
5. Linux packages (.deb, .rpm, .apk)
6. Go install

## Quick Start

### Installation

**Method 1: Install Script (Recommended)**

```bash
# macOS and Linux
curl -sSL https://raw.githubusercontent.com/harper/hex/main/install.sh | bash

# Windows (PowerShell as Administrator)
iwr -useb https://raw.githubusercontent.com/harper/hex/main/install.ps1 | iex

# Verify installation
hex --version
```

**Method 2: Homebrew (macOS/Linux)**

```bash
# Add tap and install
brew install harper/tap/hex

# Verify installation
hex --version
```

**Method 3: Go Install**

```bash
# Requires Go 1.24+
go install github.com/2389-research/hex/cmd/hex@latest

# Verify installation
hex --version
```

**Method 4: Download Binary**

Download pre-built binaries from the [releases page](https://github.com/2389-research/hex/releases):

1. Download the archive for your platform
2. Extract the binary
3. Move to a directory in your PATH (e.g., `/usr/local/bin`)
4. Run `hex --version` to verify

**Method 5: Build from Source**

```bash
# Clone repository
git clone https://github.com/2389-research/hex.git
cd hex

# Build and install
make install

# Verify installation
hex --version
```

### Setup

```bash
# Configure API key
hex setup-token sk-ant-api03-...

# Verify configuration
hex doctor
```

### Usage

**Interactive Mode** (full TUI):
```bash
# Start new conversation
hex

# Start with initial prompt
hex "Help me debug this code"

# Resume last conversation
hex --continue

# Resume specific conversation
hex --resume conv-1234567890
```

**Print Mode** (quick one-off):
```bash
# Simple query
hex --print "What is the capital of France?"

# With JSON output
hex --print --output-format json "List 3 programming languages"
```

## What's New in v1.0.0

### Production-Ready Release
After 6 phases of development, Hex v1.0 is production-ready with:
- ✅ **94.7% project completion** (Grade A)
- ✅ **73.8% test coverage** across 115+ test files
- ✅ **29,000+ lines of code** with comprehensive documentation
- ✅ **27 benchmarks** with exceptional performance (2-1000x better than targets)
- ✅ **6 distribution channels** for easy installation

### Key Capabilities

**Interactive Terminal UI**
- 📝 **Streaming responses** with progressive rendering
- 🎨 **Markdown syntax highlighting** via Glamour
- ⌨️ **Vim navigation** (j/k scroll, gg/G jump, / search)
- 📊 **Real-time token tracking**
- 🎯 **Multiple views** (Chat, History, Tools)

**Conversation Persistence**
- 💾 SQLite storage (`~/.hex/hex.db`)
- 🔄 Resume with `--continue` or `--resume <id>`
- 🏷️ Automatic conversation titles
- 📅 Full message history

**Tool System**
- 🛠️ **11 built-in tools** with user approval system
- 🔌 **MCP integration** for external tool servers
- ⚡ **Background processes** (BashOutput, KillShell)
- 🧠 **Sub-agents** via Task tool

**Logging & Observability**
- 📝 **Structured logging** (JSON/text formats)
- 🎚️ **Multiple log levels** (debug, info, warn, error)
- 📄 **File and console output**
- 🔍 **Context propagation** (conversation ID, request ID)

**Multi-Agent Capabilities**
- 📊 **Event-Sourcing** - Complete audit trail of all agent activities
- 💰 **Cost Tracking** - Automatic cost calculation per agent and tree totals
- 🔍 **Visualization** - Tree, timeline, and cost views with HTML export
- 🛡️ **Process Management** - Graceful shutdown with cascading cleanup
- 📈 **Analytics** - Analyze multi-agent performance and costs

See [Multi-Agent Features Guide](docs/MULTIAGENT_FEATURES.md) for details.

### Security Notes
⚠️ **Important**: v1.0 requires **Go 1.24.9+** to address 12 known vulnerabilities in Go stdlib. See [SECURITY_AUDIT.md](SECURITY_AUDIT.md) for details.

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

**1. Config file** (`~/.hex/config.yaml`):
```yaml
api_key: sk-ant-api03-...
model: claude-sonnet-4-5-20250929
```

**2. Environment variables**:
```bash
export ANTHROPIC_API_KEY=sk-ant-api03-...
export HEX_MODEL=claude-sonnet-4-5-20250929
```

**3. `.env` file** (in project directory):
```bash
ANTHROPIC_API_KEY=sk-ant-api03-...
HEX_MODEL=claude-sonnet-4-5-20250929
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

**Extend Hex with MCP servers** - Use external tools from the Model Context Protocol ecosystem:

```bash
# Add an MCP server
hex mcp add filesystem npx -y @modelcontextprotocol/server-filesystem ~/Documents

# List configured servers
hex mcp list

# MCP tools are automatically available in conversations
hex
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

**Current Release:** v1.0.0 (Production Ready) - 94.7% Complete (Grade A)

| Phase | Completion | Grade | Description |
|-------|-----------|-------|-------------|
| Phase 1 | 95% | A | Foundation (print mode, config, API) |
| Phase 2 | 100% | A+ | Interactive mode, core tools, storage |
| Phase 3 | 95% | A | Extended tools, MCP integration |
| Phase 4 | 88% | B+ | Advanced features (sub-agents, smart tools) |
| Phase 6A | 90% | A- | Logging, CI/CD, quality infrastructure |
| Phase 6C.2 | 100% | A+ | Smart features (autocomplete, favorites) |

**Quality Metrics:**
- 73.8% test coverage (exceeds 70% target)
- 27 benchmarks with exceptional performance
- 115+ test files, 341+ test functions
- Pre-commit hooks with comprehensive checks

## Documentation

**Getting Started:**
- **[USER_GUIDE.md](docs/USER_GUIDE.md)** - Complete usage guide
- **[TOOLS.md](docs/TOOLS.md)** - Tool system reference (including MCP)
- **[MCP_INTEGRATION.md](docs/MCP_INTEGRATION.md)** - MCP architecture and server development

**Development:**
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System design and internals
- **[ARCHITECTURE_DIAGRAM.md](ARCHITECTURE_DIAGRAM.md)** - Visual architecture guide
- **[examples/mcp/](examples/mcp/)** - MCP server examples and configurations

**Release Information:**
- **[CHANGELOG.md](CHANGELOG.md)** - Version history
- **[RELEASE_NOTES.md](RELEASE_NOTES.md)** - v1.0.0 release notes
- **[ROADMAP_UPDATED.md](ROADMAP_UPDATED.md)** - Project roadmap and status
- **[SECURITY_AUDIT.md](SECURITY_AUDIT.md)** - Security vulnerability report
- **[KNOWN_ISSUES.md](KNOWN_ISSUES.md)** - Known non-blocking issues

## Development

### Prerequisites

```bash
# Install pre-commit for Git hooks
# macOS
brew install pre-commit

# Linux
pip install pre-commit

# Verify installation
pre-commit --version
```

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/2389-research/hex.git
cd hex

# Install pre-commit hooks
pre-commit install

# Run hooks manually (optional)
pre-commit run --all-files
```

### Build

```bash
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

### Pre-commit Hooks

We use pre-commit hooks to maintain code quality. Hooks run automatically on `git commit`:

**Configured Hooks**:
- `go fmt` - Format Go code
- `go vet` - Check for suspicious constructs
- `goimports` - Organize imports
- `go test` - Run all tests (60s timeout)
- `go mod tidy` - Keep dependencies clean
- `golangci-lint` - Comprehensive linting (uses `.golangci.yml`)
- File checks (trailing whitespace, YAML syntax, etc.)

**Manual Execution**:
```bash
# Run all hooks
pre-commit run --all-files

# Run specific hook
pre-commit run go-fmt --all-files

# Skip hooks for emergency commits (not recommended)
git commit --no-verify
```

**Note**: Hooks ensure code quality and prevent common mistakes. They run quickly (typically < 10s).

### Project Structure

```
hex/
├── cmd/hex/           # CLI entry point
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

- **Go 1.24.9 or later** (required for security fixes - see [SECURITY_AUDIT.md](SECURITY_AUDIT.md))
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
hex setup-token sk-ant-api03-...
hex doctor
```

**Database locked**:
```bash
# Close other Hex instances
lsof ~/.hex/hex.db
```

**Tool timeout**:
```yaml
# Increase in config
tool_timeout: 120
```

See [USER_GUIDE.md](docs/USER_GUIDE.md) for more troubleshooting.

## Roadmap

**v1.0.0** (Current - Production Ready ✅):
- ✅ All core features complete
- ✅ Interactive TUI with streaming
- ✅ 11 built-in tools + MCP integration
- ✅ Structured logging and observability
- ✅ CI/CD with comprehensive checks
- ✅ 6 distribution channels

**v1.1** (Q1 2026):
- Enhanced MCP support (HTTP/SSE transport, resources, prompts)
- Conversation search and filtering
- Tool execution history and replay
- Advanced UI features (split panes, tabs)
- Performance optimizations

**v1.2** (Q2 2026):
- Plugin system for custom tools
- Multi-agent orchestration
- Background task management
- Advanced debugging features
- Tool result persistence

**v2.0** (Q3 2026):
- Distributed conversation sync
- Team collaboration features
- Advanced security features
- Enterprise integrations

See [ROADMAP_UPDATED.md](ROADMAP_UPDATED.md) for detailed plans.

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

**Download**: `go install github.com/2389-research/hex/cmd/hex@latest`

**Documentation**: [docs/USER_GUIDE.md](docs/USER_GUIDE.md)

**Changelog**: [CHANGELOG.md](CHANGELOG.md)
