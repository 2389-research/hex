# Phase 6C: Quality-of-Life Features Implementation Plan

## Overview
Add smart productivity features to make Hex feel polished and professional.

## Tasks

### Task 1: Command History Storage and Repository
**Goal:** Store all user messages in searchable history table with FTS5

**Files to create:**
- `internal/storage/migrations/003_history.sql`
- `internal/storage/history_repository.go`
- `internal/storage/history_repository_test.go`

**Implementation:**
1. Create migration with:
   - `history` table: `(id, user_message, assistant_response, timestamp, conversation_id)`
   - FTS5 virtual table for full-text search on messages
   - Indexes on timestamp and conversation_id
2. Implement HistoryRepository with:
   - `AddHistoryEntry(db, convID, userMsg, assistantMsg)`
   - `SearchHistory(db, query, limit)` - FTS5 search
   - `GetRecentHistory(db, limit)` - Recent entries
3. Write tests covering:
   - Basic CRUD operations
   - FTS5 search functionality
   - Edge cases (empty messages, special characters)

**Definition of Done:**
- Migration executes cleanly
- Repository methods work correctly
- All tests pass
- FTS5 search returns relevant results

### Task 2: History Command Implementation
**Goal:** Add `hex history` command with search capabilities

**Files to create:**
- `cmd/hex/history.go`
- `cmd/hex/history_test.go`

**Implementation:**
1. Create history command with subcommands:
   - `hex history` - Show recent (default 20)
   - `hex history search "docker"` - FTS5 search
   - `hex history --limit 50` - Show more results
2. Format output with:
   - Timestamp (relative: "2 hours ago")
   - Message preview (truncated to 60 chars)
   - Conversation ID (clickable hint)
3. Add to root command registration
4. Integrate with storage layer

**Tests:**
- Command parsing and flags
- Output formatting
- Integration with history repository

**Definition of Done:**
- `hex history` shows recent commands
- Search works with FTS5
- Output is readable and useful
- Tests pass

### Task 3: Session Templates System
**Goal:** YAML-based session templates for common workflows

**Files to create:**
- `internal/templates/loader.go`
- `internal/templates/loader_test.go`
- `internal/templates/types.go`
- `cmd/hex/templates.go`
- `cmd/hex/templates_test.go`

**Implementation:**
1. Create template types:
   ```go
   type Template struct {
       Name           string
       SystemPrompt   string
       InitialMessages []Message
       ToolsEnabled   []string
   }
   ```
2. Implement template loader:
   - Read from `~/.hex/templates/`
   - Parse YAML files
   - Validate structure
3. Add commands:
   - `hex templates list` - Show available
   - `hex --template code-review` - Use template
4. Create example templates:
   - `~/.hex/templates/code-review.yaml`
   - `~/.hex/templates/debug-session.yaml`

**Tests:**
- YAML parsing
- Template validation
- Loading from directory
- Command integration

**Definition of Done:**
- Templates load from YAML
- `--template` flag works
- Example templates provided
- Tests pass

### Task 4: Auto-completion System
**Goal:** Tab completion for tools, files, and commands

**Files to create:**
- `internal/ui/autocomplete.go`
- `internal/ui/autocomplete_test.go`

**Implementation:**
1. Add fuzzy matching dependency: `github.com/sahilm/fuzzy`
2. Implement completion providers:
   - Tool name completion (when typing `:tool `)
   - File path completion (standard paths)
   - Command history completion (up arrow)
3. Integrate with Bubbletea UI:
   - Show dropdown on Tab key
   - Navigate with arrow keys
   - Accept with Enter
4. Add fuzzy search as user types

**Tests:**
- Fuzzy matching accuracy
- Completion filtering
- UI integration (mock Tea messages)

**Definition of Done:**
- Tab shows completions
- Fuzzy matching works
- Arrow keys navigate
- Tests pass

### Task 5: Conversation Favorites
**Goal:** Mark and filter favorite conversations

**Files to modify:**
- `internal/storage/migrations/004_favorites.sql` (new)
- `internal/storage/conversations.go`
- `internal/storage/conversations_test.go`
- `cmd/hex/favorites.go` (new)
- `cmd/hex/favorites_test.go` (new)

**Implementation:**
1. Add migration:
   - `ALTER TABLE conversations ADD COLUMN is_favorite BOOLEAN DEFAULT 0`
   - Index on is_favorite
2. Update ConversationsRepository:
   - `SetFavorite(db, convID, isFavorite)`
   - `ListFavorites(db)`
3. Add commands:
   - `hex favorite <conv-id>` - Toggle favorite
   - `hex favorites` - List all favorites
4. Update UI to show ⭐ for favorites
5. Add Ctrl+F keyboard shortcut in interactive mode

**Tests:**
- Favorite toggle
- List favorites
- UI display
- Keyboard shortcut

**Definition of Done:**
- Favorites persist in DB
- Commands work
- UI shows stars
- Keyboard shortcut works
- Tests pass

### Task 6: Quick Actions Menu
**Goal:** Press `:` for quick action menu

**Files to create:**
- `internal/ui/quickactions.go`
- `internal/ui/quickactions_test.go`

**Files to modify:**
- `internal/ui/update.go` (add `:` key handler)
- `internal/ui/view.go` (render quick actions)

**Implementation:**
1. Create quick action registry:
   - `:read <file>` - Trigger read tool
   - `:grep <pattern>` - Trigger grep tool
   - `:web <url>` - Trigger web fetch
   - `:attach <file>` - Attach image
   - `:save` - Save conversation
   - `:export` - Export as markdown
2. Add fuzzy search as user types
3. Show in overlay/modal
4. Execute action on Enter

**Tests:**
- Action registration
- Fuzzy search
- UI rendering
- Action execution

**Definition of Done:**
- `:` opens menu
- Actions execute correctly
- Fuzzy search works
- Tests pass

### Task 7: Export Features
**Goal:** Export conversations in multiple formats

**Files to create:**
- `internal/export/exporter.go`
- `internal/export/exporter_test.go`
- `internal/export/markdown.go`
- `internal/export/json.go`
- `internal/export/html.go`
- `cmd/hex/export.go`
- `cmd/hex/export_test.go`

**Implementation:**
1. Create exporters:
   - Markdown exporter with metadata header
   - JSON exporter with full structure
   - HTML exporter with syntax highlighting (use Chroma)
2. Add command:
   - `hex export <conv-id> --format markdown`
   - `hex export <conv-id> --format json`
   - `hex export <conv-id> --format html`
3. Include metadata:
   - Timestamp
   - Model used
   - Token usage (if available)
4. Write to stdout or file

**Tests:**
- Each format export
- Metadata inclusion
- Special character handling
- Round-trip for JSON

**Definition of Done:**
- All three formats work
- Metadata included
- Output is clean
- Tests pass

### Task 8: Smart Defaults and Suggestions
**Goal:** Auto-suggest tools based on user input

**Files to create:**
- `internal/suggestions/detector.go`
- `internal/suggestions/detector_test.go`

**Files to modify:**
- `internal/ui/model.go` (add suggestion system)
- `internal/ui/view.go` (render suggestions)

**Implementation:**
1. Create pattern detectors:
   - Code questions → suggest `read_file`, `grep`
   - Shell commands → suggest `bash`
   - URLs → suggest `web_fetch`
   - File paths → suggest `read_file`
2. Show non-intrusive tips:
   - "💡 Tip: Use :read to see file contents"
   - "💡 Tip: Use :grep to search files"
3. Don't auto-enable, just suggest
4. Dismissible with Esc

**Tests:**
- Pattern detection accuracy
- Suggestion display
- Dismissal logic

**Definition of Done:**
- Detectors work correctly
- Suggestions appear at right time
- Non-intrusive display
- Tests pass

### Task 9: Integration and Hook History Tracking
**Goal:** Hook history saving into conversation flow

**Files to modify:**
- `internal/ui/model.go` (save history on send)
- `internal/storage/schema.go` (add migration 003)

**Implementation:**
1. Update schema.go to include migration 003
2. In UI model, after assistant responds:
   - Save to history table
   - Use HistoryRepository.AddHistoryEntry()
3. Ensure history is saved for both interactive and print modes

**Tests:**
- History saves after each message
- Print mode also saves history
- No duplicates

**Definition of Done:**
- History automatically saves
- Works in all modes
- Tests pass

### Task 10: Documentation and Examples
**Goal:** Complete documentation for all features

**Files to create:**
- `docs/PRODUCTIVITY.md`

**Files to modify:**
- `README.md` (add features section)

**Implementation:**
1. Create PRODUCTIVITY.md with:
   - All keyboard shortcuts table
   - Template usage guide with examples
   - History search examples
   - Export format samples
   - Pro tips section
2. Update README.md:
   - Feature list
   - Quick start examples
   - Link to PRODUCTIVITY.md

**Definition of Done:**
- Documentation complete
- Examples are clear
- All features documented
- Links work

## Success Criteria

All features implemented with:
- Comprehensive test coverage
- Clean, maintainable code
- Fast performance (async where possible)
- Polished UX
- Complete documentation

## Dependencies

- `github.com/sahilm/fuzzy` - Fuzzy matching
- `gopkg.in/yaml.v3` - YAML parsing (already in go.mod)
- FTS5 extension for SQLite (built into modernc.org/sqlite)
