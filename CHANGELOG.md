# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2025-11-28

### Added

#### Phase 5: Enhanced Tools + MCP Foundation

Four major enhancements expanding Clem's capabilities for background execution, streaming, persistence, and external tool integration:

**Phase 5A: Enhanced Tools**

**Bash Tool - Background Execution** (`run_in_background` parameter)
- Added optional `run_in_background` parameter (boolean, default false)
- When true: launches command in background, returns process ID immediately
- Integrates with existing BackgroundProcessRegistry
- Works seamlessly with BashOutput and KillShell tools
- 11 new tests, 38 total tests passing

**Task Tool - Streaming Updates** (`ExecuteStreaming()` method)
- New `ExecuteStreaming()` method returns `<-chan *Result` for incremental output
- Channel-based streaming with real-time progress updates
- Thread-safe concurrent stdout/stderr capture with mutex protection
- Progress metadata includes bytes_read, streaming status
- Respects timeout and context cancellation
- 8 new tests, 33 total tests passing

**TodoWrite Tool - SQLite Persistence**
- Database migration (002_todos.sql) creates todos table
- New repository layer: SaveTodos(), LoadTodos(), ClearCompleted()
- Auto-save functionality on Execute()
- `load_from_db` parameter to restore previous session
- Conversation scoping with CASCADE delete on conversation removal
- 13 repository tests + 8 tool tests, all passing

**Phase 5B: MCP Foundation**

**MCP (Model Context Protocol) Integration**
- JSON-RPC 2.0 client implementation with stdio transport
- Server registry with .mcp.json persistence (project-level config)
- Tool adapter bridging MCP tools to Clem Tool interface
- CLI commands:
  - `clem mcp add <name> <command> [args...]` - Register MCP server
  - `clem mcp list` - Show configured servers
  - `clem mcp remove <name>` - Unregister server
- 38 MCP tests (client, registry, adapter, CLI), all passing

**Phase 5C: MCP Tool Loading & Integration**

**Automatic MCP Tool Loading**
- Created `internal/mcp/loader.go` with `LoadMCPTools()` function
- Loads `.mcp.json` from project directory at startup
- Initializes MCP clients and registers tools automatically
- Graceful degradation when .mcp.json missing or invalid
- Built-in tools always available regardless of MCP status

**Integration**
- Modified `cmd/clem/root.go` to load MCP tools after built-in tools
- MCP tools seamlessly integrated into tool registry
- Works in both interactive and print modes
- 10+ integration tests verifying tool discovery and registration

**Documentation**
- Updated `docs/TOOLS.md` with MCP section (~360 lines)
- Created `docs/MCP_INTEGRATION.md` architecture guide (~850 lines)
- Created `examples/mcp/README.md` with practical examples (~670 lines)
- Created `.mcp.json.example` with well-commented configurations
- Updated main `README.md` with MCP features

**New Files**:
- internal/mcp/client.go - JSON-RPC 2.0 client
- internal/mcp/registry.go - Server configuration CRUD
- internal/mcp/tool_adapter.go - MCP→Clem bridge
- internal/mcp/loader.go - Automatic tool loading
- internal/mcp/mock_server_test.go - Test infrastructure
- cmd/clem/mcp.go - CLI commands
- cmd/clem/mcp_integration_test.go - Integration tests
- docs/MCP_INTEGRATION.md - Architecture documentation
- examples/mcp/README.md - Practical examples
- examples/mcp/.mcp.json.example - Example configuration
- internal/storage/migrations/002_todos.sql - Todos schema
- internal/storage/todo_repository.go - Todo persistence layer
- Plus test files for all new functionality

### Fixed

**Tool Interface Contract Compliance**
- Fixed WebFetchTool and WebSearchTool test expectations after Phase 4 bug fixes
- Updated 11 tests to check `Result.Success` instead of expecting Go errors
- All validation tests now correctly follow Tool interface contract

**MCP Test Infrastructure**
- Fixed mock server stdio communication using `io.Pipe()` instead of `bytes.Buffer`
- Fixed Cobra flag parsing in CLI tests with `DisableFlagParsing: true`
- Fixed error message validation to match actual Cobra output
- All MCP tests now passing

### Technical Details

- **Development Approach**: Parallel subagent development (6 agents total across 3 sessions)
  - Session 1: 4 agents for Phase 5A & 5B foundation
  - Session 2: Code review and bug fixes
  - Session 3: 2 agents for Phase 5C integration & documentation
- **Development Time**: ~6 hours (parallel) vs ~24 hours (sequential) - 4x speedup
- **Test Coverage**:
  - Phase 5A: 32 new tests across 3 enhanced tools
  - Phase 5B: 38 new tests for MCP foundation
  - Phase 5C: 10+ new integration tests
  - Total new tests: 80+ tests
- **Total Tests**: ~420+ tests across all packages, all passing
- **New Dependencies**: None (reused existing dependencies)
- **Lines Added**: ~8,500 lines (code + tests + docs)
- **Documentation**: 4 new/updated docs (~2,500 lines of documentation)

## [0.4.0] - 2025-11-28

### Added

#### Phase 4: Extended Capabilities

Seven new tools expanding Clem's capabilities for interactive decision-making, research, and advanced execution:

**Phase 4A: Interactive Tools**

**AskUserQuestion Tool** (`ask_user_question`)
- Interactive multiple-choice question prompts (1-4 questions per call)
- Single-select and multi-select support
- 2-4 options per question with automatic "Other: ..." option
- Always requires approval (user interaction)
- 14 comprehensive unit tests

**TodoWrite Tool** (`todo_write`)
- Structured task list management with visual indicators
- Three status levels: pending (☐), in_progress (⏳), completed (✅)
- Progress metadata tracking (total, pending, in_progress, completed)
- Never requires approval (display-only)
- 21 comprehensive unit tests

**Phase 4B: Research Tools**

**WebFetch Tool** (`web_fetch`)
- HTTP GET requests with HTML-to-markdown conversion
- 30-second timeout with context cancellation support
- Auto-detects content type (HTML, JSON, text, XML)
- User-Agent header set to "Clem/1.0"
- Always requires approval (network access)
- 12 comprehensive unit tests
- New dependency: `github.com/JohannesKaufmann/html-to-markdown v1.6.0`

**WebSearch Tool** (`web_search`)
- DuckDuckGo search integration (no API key required)
- Result limiting (default 10, configurable)
- Domain filtering (allowed/blocked lists, case-insensitive)
- Formatted markdown output with titles, URLs, snippets
- Always requires approval (network access)
- 12 comprehensive unit tests
- New dependency: `golang.org/x/net v0.47.0`

**Phase 4C: Advanced Execution Tools**

**Task Tool** (`task`)
- Sub-agent spawning for complex multi-step tasks
- Spawns `clem --print` as subprocess
- Environment inheritance (API keys, working directory, config)
- Configurable timeout (5min default, 30min max)
- Auto-builds clem binary if not in PATH
- Always requires approval (spawns processes, uses API)
- 21 comprehensive unit tests (15 validation, 6 integration)

**BashOutput Tool** (`bash_output`)
- Retrieves output from background bash processes by ID
- Incremental reading (only new output since last check)
- Optional regex filtering for stdout/stderr
- Thread-safe with mutex-protected read offsets
- Process metadata (done status, exit code)
- Never requires approval (read-only)
- 13 comprehensive unit tests

**KillShell Tool** (`kill_shell`)
- Terminates running background bash processes
- Two-stage shutdown: SIGTERM (graceful) → SIGKILL (force)
- Automatic registry cleanup
- Handles already-terminated processes gracefully
- Always requires approval (destructive operation)
- 10 comprehensive unit tests

**Background Process Registry**
- Shared thread-safe registry for background bash processes
- Supports both new BackgroundProcess struct (with output buffering) and legacy Process registration
- Mutex-protected operations for concurrent access
- Used by BashOutput and KillShell tools

### Changed

- Tool registry expanded from 6 to 13 tools
- Updated TOOLS.md with 1,546 lines of Phase 4 documentation
- Integration test suite updated to verify all 13 tools
- Table of contents and tools summary table updated

### Technical Details

- **Test Coverage**:
  - Phase 4A: 35 tests (14 AskUserQuestion + 21 TodoWrite)
  - Phase 4B: 24 tests (12 WebFetch + 12 WebSearch)
  - Phase 4C: 44 tests (21 Task + 13 BashOutput + 10 KillShell)
  - Total Phase 4: 103 new tests
- **Total Tests**: ~341 tests, all passing
- **New Files**: 14 implementation files + 14 test files + docs
- **Modified Files**: root.go, tools_test.go, TOOLS.md
- **New Dependencies**:
  - `github.com/JohannesKaufmann/html-to-markdown v1.6.0`
  - `golang.org/x/net v0.47.0`
- **Lines Added**: ~8,500 lines (code + tests + docs)

## [0.3.0] - 2025-11-28

### Added

#### Extended Tool Support (Phase 3)

Three new tools complete Clem's core tool suite:

**Edit Tool** (`edit`)
- Exact string replacement in files
- Single replacement mode (default) - requires unique match
- Replace-all mode for renaming variables, updating imports
- Preserves file encoding, indentation, and line endings
- Unicode support
- Always requires approval (destructive operation)
- 15 comprehensive unit tests

**Grep Tool** (`grep`)
- Code search powered by ripgrep
- Regex pattern matching across files
- Context lines support (-A, -B, -C)
- Multiple output modes: content, files_with_matches, count
- Case-insensitive search (-i flag)
- Glob and file type filtering
- Read-only (never requires approval)
- 17 comprehensive unit tests

**Glob Tool** (`glob`)
- File pattern matching with doublestar (`**`) support
- Brace expansion for multiple extensions (`*.{ts,tsx}`)
- Recursive directory traversal
- Results sorted by modification time (newest first)
- Directory prefix matching (`src/**/*.tsx`)
- Read-only (never requires approval)
- 11 comprehensive unit tests

### Fixed

- **Vim navigation conflict**: `/` key now only activates search mode when textarea is NOT focused, allowing users to type paths like `/Users/...` without triggering search
- Test updates for proper textarea focus handling

### Changed

- Tool registry now includes all 6 tools (Read, Write, Bash, Edit, Grep, Glob)
- Updated TOOLS.md documentation with comprehensive examples for all new tools
- Integration test suite expanded with Edit/Grep/Glob coverage

### Technical Details

- **Test Coverage**: 47 new tests added (43 unit + 4 integration)
- **Total Tests**: ~238 tests, all passing
- **New Files**: 6 (edit_tool.go/test, grep_tool.go/test, glob_tool.go/test)
- **Modified Files**: 4 (root.go, update.go, update_test.go, tools_test.go)

## [0.2.0] - 2025-11-27

### Added

#### Interactive Mode
- Full-featured TUI built with Bubbletea and Charm ecosystem libraries
- Streaming responses with progressive text rendering
- Real-time markdown formatting using Glamour
- Vim-style keyboard navigation (j/k for scroll, gg/G for jump, / for search)
- Multiple view modes: Chat, History, Tools Inspector
- Context-aware help text and status indicators (Idle/Streaming/Error)
- Token usage tracking with real-time counters

#### Storage System
- SQLite-based conversation persistence at `~/.clem/clem.db`
- Hybrid schema design: normalized tables + JSON for complex data
- Automatic schema migrations with embedded SQL files
- WAL mode enabled for better concurrency
- Foreign key constraints and optimized indexes
- Conversation CRUD operations (create, get, list, update timestamp)
- Message CRUD operations with conversation association
- `--continue` flag to resume last conversation
- `--resume <id>` flag to resume specific conversation by ID
- Automatic conversation title generation based on first message

#### Tool System
- Comprehensive tool execution framework with registry and executor
- Permission-based approval system for dangerous operations
- Three production-ready tools:

**Read Tool** (`read_file`)
- Safe file reading with path validation
- Approval required for sensitive paths (/etc, ~/.ssh, etc.)
- File size limits to prevent memory issues
- UTF-8 content validation
- Detailed error messages

**Write Tool** (`write_file`)
- Three operation modes: create, overwrite, append
- User confirmation for overwrite operations
- Atomic writes with temp files
- Directory creation if needed
- Content validation and error handling

**Bash Tool** (`bash`)
- Sandboxed shell command execution
- Configurable timeout (default 30s, max 5min)
- Real-time output streaming
- Working directory support
- Exit code capture
- Dangerous command detection (rm -rf, etc.)
- User approval for destructive operations

#### UI Enhancements
- Tool execution visualization with status updates
- Streaming text accumulation with live rendering
- Search mode with live query input
- Conversation title display in header
- Graceful error display in UI
- Window resize handling
- Proper cleanup on exit

### Changed
- API client now supports both streaming and non-streaming modes
- SSE (Server-Sent Events) parsing for streaming responses
- Database connection management improved with pragmas
- Context cancellation handling in streaming operations
- Tool parameter validation and nil handling
- Type unification for ToolUse across packages

### Fixed
- Context cancellation in streaming API client
- Nil parameter handling in tool executor
- ToolUse type consistency between core and storage packages
- Flag conflict detection (--continue and --resume are mutually exclusive)
- Viewport scrolling in UI with proper bottom anchoring
- Memory leaks in streaming goroutines

### Technical Details
- Go 1.24+ required
- Dependencies: Bubbletea, Lipgloss, Glamour, modernc.org/sqlite
- Test coverage: unit tests, integration tests, example-based tests
- All tests passing with comprehensive coverage
- TDD approach followed throughout implementation

## [0.1.0] - 2025-11-25

### Added
- Initial release with foundation features
- CLI framework using Cobra
- Configuration system with Viper and .env support
- Anthropic API client with Messages API integration
- Print mode (`--print`) for non-interactive usage
- Setup command (`clem setup-token`) for API key configuration
- Doctor command (`clem doctor`) for health checks
- JSON output format support
- Model selection via flags and config
- Unit and integration test suites
- VCR cassettes for API call recording/replay
- Comprehensive error handling

### Technical Details
- Go 1.24+ required
- Clean architecture with internal/pkg separation
- Real components over mocks in tests
- Integration tests with real filesystem interactions

[0.2.0]: https://github.com/harper/clem/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/harper/clem/releases/tag/v0.1.0
