# Clem Architecture Diagram

This directory contains a comprehensive architecture diagram of the Clem project, visualizing all components, data flows, and relationships.

## Files

- **`clem-architecture.dot`** (18KB) - Source GraphViz DOT file
- **`clem-architecture.png`** (605KB) - High-resolution PNG render
- **`clem-architecture.svg`** (98KB) - Scalable SVG render (recommended for viewing)

## Viewing the Diagram

### Recommended: SVG
Open `clem-architecture.svg` in your browser for the best experience:
```bash
open clem-architecture.svg  # macOS
xdg-open clem-architecture.svg  # Linux
start clem-architecture.svg  # Windows
```

### Alternative: PNG
View `clem-architecture.png` in any image viewer or browser.

## Regenerating the Diagram

If you modify `clem-architecture.dot`, regenerate the images with:

```bash
# Generate PNG
dot -Tpng clem-architecture.dot -o clem-architecture.png

# Generate SVG (scalable, recommended)
dot -Tsvg clem-architecture.dot -o clem-architecture.svg

# Generate PDF (for documentation)
dot -Tpdf clem-architecture.dot -o clem-architecture.pdf
```

## What's Visualized

The diagram shows:

### 1. External Systems (Light Blue)
- **User** - The developer using Clem
- **Anthropic API** - Claude messages endpoint with streaming
- **MCP Servers** - External tool servers (filesystem, fetch, custom)
- **File System** - User's project files and Clem config

### 2. CLI Layer (Light Green)
All entry point commands organized by category:

**Core Commands:**
- `main.go` - Cobra root and app entry
- `root.go` - Interactive mode (default)
- `print.go` - One-shot query mode (`--print`)
- `setup.go` - API key configuration
- `doctor.go` - Health check diagnostics

**Storage Commands (Phase 6C.2):**
- `history.go` - Search conversation history with FTS5
- `favorites.go` - Manage favorite conversations
- `export.go` - Export to MD/JSON/HTML

**Template Commands (Phase 6C.2):**
- `templates.go` - Session templates (YAML-based)

**MCP Commands:**
- `mcp.go` - MCP server management

### 3. Core Package (Light Yellow)
API client and configuration:
- `client.go` - HTTP client for Messages API
- `stream.go` - SSE parser with delta accumulation
- `types.go` - Request/response types
- `config.go` - Viper-based multi-source config

### 4. UI Package (Light Coral)
Bubbletea-based TUI:

**Core Components:**
- `model.go` - UI state (Elm Architecture)
- `update.go` - Event handlers (keyboard, messages)
- `view.go` - Renderer (Lipgloss + Glamour)
- `styles.go` - UI styling (colors, borders)

**Advanced Features (Phase 6C.2):**
- `autocomplete.go` - Tab completion with 3 providers (tool/file/history)
- `quickactions.go` - Command palette (`:` trigger, 6 actions)
- `suggestions.go` - Smart suggestions with pattern detection & learning

### 5. Storage Package (Light Gray)
SQLite persistence layer:
- `schema.go` - Database initialization
- `conversations.go` - Conversation CRUD + favorites + FTS5 search
- `messages.go` - Message CRUD with JSON tool calls
- `migrations/` - Embedded SQL migrations (001-004)
- `~/.clem/clem.db` - SQLite database (WAL mode)

### 6. Tools Package (Light Salmon)
Tool execution system with 13+ tools:

**Core System:**
- `registry.go` - Tool registry (name→tool map)
- `executor.go` - Execution engine with approval callbacks
- `tool.go` - Tool interface
- `result.go` - Result types

**Built-in Tools:**
- `read_tool.go` - Read files (sensitive path checking)
- `write_tool.go` - Write files (atomic writes)
- `edit_tool.go` - String replacement (single/bulk)
- `bash_tool.go` - Shell commands (timeout, sandbox)
- `grep_tool.go` - Code search (ripgrep)
- `glob_tool.go` - File pattern matching

**Advanced Tools (Phase 4):**
- `ask_user_question.go` - Interactive Q&A
- `todo_write.go` - Task list management
- `web_fetch.go` - HTTP GET with HTML→Markdown
- `web_search.go` - DuckDuckGo search
- `task.go` - Sub-agent spawning
- `bash_output.go` - Background process monitoring
- `kill_shell.go` - Process termination

### 7. MCP Package (Light Steel Blue)
Model Context Protocol integration:
- `registry.go` - Server configuration (`.mcp.json`)
- `client.go` - JSON-RPC client (stdio transport)
- `tool_adapter.go` - MCP→Clem tool bridge
- `tool_manager.go` - Multi-server tool management

### 8. Templates & Export (Lavender)
**Templates:**
- `templates/types.go` - Template structs
- `templates/loader.go` - YAML loading
- `~/.clem/templates/` - Template files (code-review.yaml, etc.)

**Export:**
- `export/exporter.go` - Exporter interface
- `export/markdown.go` - Markdown with YAML frontmatter
- `export/json.go` - JSON round-trip
- `export/html.go` - HTML with Chroma syntax highlighting (XSS-safe)

### 9. Smart Suggestions (Thistle)
Pattern detection and adaptive learning:
- `suggestions/detector.go` - 15+ pattern detectors
- `suggestions/learner.go` - Success/failure tracking with decay

## Data Flows

The diagram visualizes these key flows with colored edges:

1. **User Interaction** (Blue, bold) - User → CLI → UI
2. **API Communication** (Purple, bold) - Client → Anthropic API → Streaming
3. **Tool Execution** (Red, bold) - Executor → Tools → Results
4. **MCP Integration** (Dark Blue, bold) - MCP Client → Servers → Registry
5. **Storage Operations** (Black) - Components → SQLite DB
6. **File System** (Dashed) - Tools → Filesystem
7. **Network** (Dashed) - Web tools → External APIs

## Color Legend

| Color | Component Type |
|-------|---------------|
| Light Blue | User/External Systems |
| Light Green | CLI Commands |
| Light Yellow | Core/API |
| Light Coral | UI Components |
| Light Gray | Storage |
| Light Salmon | Tools |
| Light Steel Blue | MCP |
| Plum | Phase 6C.2 Features |
| Lavender | Templates & Export |
| Thistle | Smart Suggestions |

## Architecture Highlights

### Elm Architecture (UI)
The UI follows the Elm pattern:
- **Model** - Immutable state
- **Update** - Pure event handlers
- **View** - Pure render functions

### Registry Pattern (Tools)
- Extensible tool system
- Interface-based design
- Runtime registration
- MCP tools integrate seamlessly

### Hybrid Storage Schema
- Normalized tables (conversations, messages)
- JSON columns (tool_calls, metadata)
- FTS5 full-text search
- WAL mode for concurrency

### Streaming Architecture
- SSE (Server-Sent Events) from Anthropic API
- Channel-based delta accumulation
- Real-time UI updates
- Graceful cancellation

### MCP Integration
- JSON-RPC 2.0 over stdio
- Process lifecycle management
- Tool wrapping and registration
- Multi-server support

## Development Phases

The diagram shows features from all completed phases:

- **Phase 1** (Foundation) - CLI, Core, Config
- **Phase 2** (Interactive) - UI, Storage, Streaming
- **Phase 3** (Tools) - 13+ built-in tools
- **Phase 4** (MCP) - External tool integration
- **Phase 6C.2** (Smart Features) - 7 productivity enhancements

## For Contributors

When modifying the architecture:

1. Update the DOT file first
2. Regenerate PNG/SVG
3. Document changes in this README
4. Update ARCHITECTURE.md if needed

## Related Documentation

- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Detailed technical architecture
- **[docs/TOOLS.md](docs/TOOLS.md)** - Tool system reference
- **[docs/UI_GUIDE.md](docs/UI_GUIDE.md)** - UI features and shortcuts
- **[docs/MCP_INTEGRATION.md](docs/MCP_INTEGRATION.md)** - MCP architecture
- **[README.md](README.md)** - Project overview
- **[ROADMAP.md](ROADMAP.md)** - Development roadmap

---

**Last Updated:** 2025-11-28
**Diagram Version:** 1.0 (Complete through Phase 6C.2)
