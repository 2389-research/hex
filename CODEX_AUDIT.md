# Codex CLI Architecture Audit

**Date**: 2025-12-02
**Auditor**: Claude (analyzing OpenAI's Codex CLI source code)
**Source**: `/Users/harper/workspace/2389/agent-class/agents/codex`

## Executive Summary

Codex CLI is OpenAI's local coding agent that runs in your terminal. It's a Rust-based TUI application that uses Claude/GPT models to autonomously write code, run tests, and execute commands on your local machine with sophisticated sandboxing and approval mechanisms.

---

## Core Architecture

### Language & Stack
- **Primary Language**: Rust (workspace with 40+ crates)
- **TUI Framework**: Ratatui
- **Core crates**:
  - `codex-core`: Main business logic
  - `codex-tui`: Terminal user interface
  - `codex-cli`: CLI entry point
  - `execpolicy`: Command execution safety
  - `mcp-types`: Model Context Protocol integration

### Key Design Principles
1. **Sandbox-first**: All commands run in sandboxed environments (Seatbelt on macOS, Landlock on Linux, Windows Sandbox on Windows)
2. **Approval-based**: User control over what the agent can execute
3. **Stateful**: Maintains conversation history with automatic compaction
4. **Modular**: Tool system with pluggable handlers

---

## Built-in Tools

From analyzing `core/src/tools/handlers/`:

| Tool | Purpose | Implementation |
|------|---------|----------------|
| `shell` / `unified_exec` | Execute shell commands | Runs in sandbox with approval flow |
| `read_file` | Read file contents | With encoding detection & truncation |
| `list_dir` | List directory contents | Recursive with gitignore awareness |
| `grep_files` | Search file contents | Powered by ripgrep |
| `apply_patch` | Apply code edits | Structured diff application |
| `view_image` | Process image input | For screenshot/diagram analysis |
| `plan` | Create execution plans | Task breakdown |
| `mcp_resource` | Access MCP servers | External tool integration |
| `web_search_request` | Search the web | (Feature-flagged) |

### Tool Execution Flow
1. Model requests tool via function calling
2. `tools/orchestrator.rs` routes to handler
3. Handler executes in `tools/runtimes/` (sandboxed or not)
4. `execpolicy` validates command safety
5. TUI shows approval prompt if needed
6. Result streams back to model

---

## Context Management & Compaction

**Location**: `core/src/compact.rs`, `core/src/context_manager/`

### How It Works
1. **Token Estimation**: ~4 chars per token heuristic
2. **Compaction Trigger**: When approaching model context limit
3. **Two modes**:
   - **Remote Compaction** (for ChatGPT auth): OpenAI's backend creates summary
   - **Local Compaction**: Uses Claude Haiku to summarize old messages

### Compaction Process (`compact.rs:44`)
```rust
pub(crate) async fn run_inline_auto_compact_task(
    sess: Arc<Session>,
    turn_context: Arc<TurnContext>,
)
```

- Summarizes old conversation turns
- Preserves system messages and recent context
- Creates summary message with prefix
- **Template**: `templates/compact/prompt.md` instructs summarization
- Uses Haiku model for cost efficiency (200 word max summary)
- Caches summaries to avoid re-summarization

### Key Constants
```rust
const COMPACT_USER_MESSAGE_MAX_TOKENS: usize = 20_000;
```

---

## Sandboxing Architecture

**Location**: `core/src/sandboxing/`, `core/src/seatbelt.rs`, `execpolicy/`

### Platform-Specific Sandboxes

#### macOS (Seatbelt)
- Uses `/usr/bin/sandbox-exec`
- Policy: `seatbelt_base_policy.sbpl`
- Network policy: `seatbelt_network_policy.sbpl`
- Sets `CODEX_SANDBOX=seatbelt` env var

#### Linux (Landlock)
- Kernel-based access control
- Restricts filesystem paths
- Network can be disabled

#### Windows
- Uses `codex-windows-sandbox-rs`
- Job objects for process isolation

### Execpolicy Rules
- Parses `.execpolicy.toml` files
- Defines allowed/denied commands
- Can require approval for specific patterns
- Supports wildcards and regex matching

---

## Authentication & API Integration

**Location**: `core/src/auth.rs`, `codex-api/`

### Supported Auth Methods
1. **ChatGPT** (preferred): OAuth flow with ChatGPT account
2. **API Key**: OpenAI API key (usage-based billing)
3. **OSS Providers**: LM Studio, Ollama, custom endpoints

### API Communication
- Uses Server-Sent Events (SSE) for streaming
- Protocol: `codex-protocol` crate defines message format
- Responses API: `/v1/responses` endpoint
- Handles function calls, tool results, content blocks

---

## Session Management & Persistence

**Location**: `core/src/rollout/`, `core/src/conversation_manager.rs`

### Storage Structure
```
~/.codex/
├── sessions/           # Active sessions
│   └── {session-id}/
│       ├── rollout.ndjson  # Event log
│       └── meta.json        # Session metadata
├── archived/          # Completed sessions
├── config.toml        # User configuration
└── AGENTS.md          # Global instructions
```

### Rollout Format (NDJSON)
Each line is a JSON event:
- `TurnContextItem`: CWD, approval policy, model
- `UserInputItem`: What user typed
- `ResponseInputItem`: What was sent to API
- `LocalShellAction`: Commands executed
- `ContextCompactedEvent`: When compaction occurred

### Resuming Sessions
```bash
codex resume              # Picker UI
codex resume --last       # Most recent
codex resume {id}         # Specific session
```

---

## AGENTS.md Memory System

**Location**: `core/src/project_doc.rs`, `core/src/user_instructions.rs`

### Lookup Order (merged top-down)
1. `~/.codex/AGENTS.md` - Global user preferences
2. Repo root `AGENTS.md`
3. Each parent directory up to CWD
4. CWD `AGENTS.md` or `AGENTS.override.md`

### Override Behavior
- `AGENTS.override.md` replaces inherited instructions for that directory
- Regular `AGENTS.md` merges with parent instructions

---

## TUI Features

**Location**: `codex-rs/tui/`

### Key Components
- **Chat View**: Main conversation interface
- **Approval Forms**: Huh-based interactive prompts
- **Transcript View**: Message history
- **File Search**: Fuzzy `@filename` autocomplete
- **Image Paste**: Ctrl+V for screenshots
- **Backtrack Mode**: Esc-Esc to edit previous messages

### Styling
- Uses Ratatui's `Stylize` trait
- Concise helpers: `.dim()`, `.cyan()`, `.bold()`
- Avoid hardcoded `.white()` - use default fg

### Snapshot Testing
- Uses `cargo-insta` for TUI rendering tests
- Validates terminal output with `.snap` files

---

## Skills System

**Location**: `core/src/skills/`

Codex has a "skills" concept (similar to prompts/instructions) that can be:
- Defined in config
- Invoked via slash commands
- Auto-applied based on context

---

## Custom Prompts & Slash Commands

**Location**: `docs/prompts.md`, `docs/slash_commands.md`

Users can define:
```toml
[prompts.refactor]
content = "Refactor this code to use modern patterns"

[prompts.review]
content = "Review this PR for security issues"
```

Invoked with `/refactor` or `/review` in TUI.

---

## Testing Infrastructure

### Test Types
1. **Unit tests**: Inline with `#[test]`
2. **Integration tests**: `core_test_support::responses`
3. **Snapshot tests**: `insta` for TUI rendering
4. **E2E tests**: Full agent runs with mocked SSE

### Test Helpers (`AGENTS.md:79-104`)
```rust
// Mount SSE response
let mock = responses::mount_sse_once(&server, responses::sse(vec![
    responses::ev_response_created("resp-1"),
    responses::ev_function_call(call_id, "shell", &args),
    responses::ev_completed("resp-1"),
])).await;

// Assert on request
let request = mock.single_request();
assert_eq!(request.function_call_output(call_id), expected);
```

---

## Configuration

**Location**: `~/.codex/config.toml`, `docs/config.md`

### Key Settings
```toml
[ai]
default_model = "claude-sonnet-4-5-20250929"

[sandbox]
enabled = true
network_disabled = false

[approval]
ask_for_approval = true

[mcp_servers.my_server]
command = "node"
args = ["server.js"]
```

---

## Notable Implementation Details

### Message History Truncation
- `core/src/truncate.rs` handles token budgets
- `TruncationPolicy` enum: `TruncateTokens`, `NoTruncation`
- Uses `approx_token_count()` for estimation

### Text Encoding Detection
- `core/src/text_encoding.rs`
- Uses `chardetng` to detect file encodings
- Gracefully handles binary files

### Git Integration
- `core/src/git_info.rs` (40KB!)
- Detects repo root, current branch
- Reads `.gitignore` for file filtering
- Shows branch in session picker

### Bash Command Parsing
- `core/src/bash.rs` + `tree-sitter-bash`
- Parses shell commands for safety analysis
- Detects dangerous patterns (rm -rf, etc.)

### Parallel Tool Execution
- `core/src/tools/parallel.rs`
- Can run multiple tools concurrently
- Coordinates results back to model

---

## Differences from hex

| Feature | Codex | hex |
|---------|-------|------|
| **Language** | Rust | Go |
| **TUI** | Ratatui | Bubbletea |
| **Sandboxing** | OS-specific (Seatbelt/Landlock) | Not yet implemented |
| **Compaction** | Automatic via API or Haiku | Infrastructure exists but unused |
| **MCP Support** | Full MCP client | Not yet |
| **Tool System** | Rich built-in + MCP | Basic set |
| **Approval Rules** | Execpolicy DSL | Just added persistent JSON rules |
| **Session Format** | NDJSON rollout | SQLite database |
| **Auth** | ChatGPT OAuth + API key | Direct API client |
| **Image Input** | Full support | Not yet |
| **Resuming** | Rich picker UI | Not yet |

---

## Security Model

### Defense-in-Depth Layers
1. **Sandbox**: OS-level process isolation
2. **Execpolicy**: Command pattern matching
3. **Approval Flow**: User confirmation
4. **Network Disable**: Optional internet blocking
5. **Escalation Requests**: Explicit permission for unsafe ops

### Environment Variables
- `CODEX_SANDBOX=seatbelt` - Indicates sandbox active
- `CODEX_SANDBOX_NETWORK_DISABLED=1` - Network disabled
- Tests check these to skip incompatible tests

---

## Build & Development

### Justfile Commands
```bash
just fmt              # Format code
just fix -p core      # Fix lints for specific crate
just fix              # Fix all workspace lints
cargo test -p codex-tui  # Run TUI tests
```

### Code Style
- Always inline format! args
- Collapse if statements
- Method references over closures
- No unsigned integers
- Compare whole objects in tests

---

## Open Source & Licensing

- **License**: Apache 2.0
- **Repository**: OpenAI GitHub
- Actively maintained
- Community contributions welcome

---

## Key Takeaways for hex

### What to Consider Adopting
1. **Auto-compaction**: Actually use the compaction infrastructure
2. **Execpolicy DSL**: More sophisticated than JSON rules
3. **NDJSON rollout**: Simpler than SQLite for append-only logs
4. **Snapshot testing**: Great for TUI validation
5. **MCP integration**: Industry standard for tool extensibility
6. **Image input**: Valuable for debugging screenshots
7. **Session resuming**: Better UX than current approach

### Where hex Differs (Good!)
1. **SQLite**: Better for complex queries than NDJSON
2. **Go simplicity**: Easier to build/deploy than Rust
3. **Simpler approval**: JSON rules vs. execpolicy DSL
4. **Direct focus**: Not trying to support multiple auth modes

### Architecture Insights
- Codex is **production-grade** but **complex**
- Heavy use of async Rust (tokio)
- Very modular (40+ crates)
- Strong separation: protocol, core, TUI, CLI
- Testing is **extensive** but not comprehensive

---

## References

- Source: `/Users/harper/workspace/2389/agent-class/agents/codex`
- Docs: `codex/docs/`
- Core: `codex-rs/core/src/`
- TUI: `codex-rs/tui/src/`
- Protocol: `codex-rs/protocol/src/`
