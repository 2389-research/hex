# Clem - Claude CLI

A Go implementation of Claude Code CLI with interactive mode and tool execution capabilities.

**Latest Version**: v0.2.0

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

### Coming in Phase 3
- [ ] Extended tools (Edit, Grep, Glob)
- [ ] MCP (Model Context Protocol) integration
- [ ] Plugin system
- [ ] Advanced debugging features

## Quick Start

### Installation

```bash
# Install latest version
go install github.com/harper/clem/cmd/clem@latest

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

### Read Tool
- Safely read file contents
- Approval required for sensitive paths (`/etc`, `~/.ssh`, etc.)
- UTF-8 content validation
- 10MB size limit

### Write Tool
- Three modes: create, overwrite, append
- Atomic writes with temp files
- Confirmation for overwrites
- Directory creation if needed

### Bash Tool
- Execute shell commands with timeout
- Sandboxed execution
- Dangerous command detection
- Real-time output streaming
- Default 30s timeout, max 5min

**See [TOOLS.md](docs/TOOLS.md) for complete tool documentation.**

## Project Status

| Phase | Status | Version | Description |
|-------|--------|---------|-------------|
| Phase 1 | ✅ Complete | v0.1.0 | Foundation (print mode, config, API) |
| Phase 2 | ✅ Complete | v0.2.0 | Interactive mode, tools, storage |
| Phase 3 | 🔄 Planned | v0.3.0 | Extended tools, MCP integration |
| Phase 4 | 📋 Planned | v0.4.0 | Plugin system |

## Documentation

- **[USER_GUIDE.md](docs/USER_GUIDE.md)** - Complete usage guide
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System design and internals
- **[TOOLS.md](docs/TOOLS.md)** - Tool system reference
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

**v0.3.0** (Q1 2026):
- Extended tools (Edit, Grep, Glob)
- Tool result persistence
- Multi-tool execution queueing
- MCP server integration (beta)

**v0.4.0** (Q2 2026):
- Full MCP support
- Plugin system
- Custom tool registration
- Advanced debugging features

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
