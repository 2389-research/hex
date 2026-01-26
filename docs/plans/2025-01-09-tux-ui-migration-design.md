# Tux UI Migration Design

**Date:** 2025-01-09
**Status:** Approved
**Goal:** Replace hex's custom TUI (~13,300 lines) with tux's higher-level abstractions

## Overview

Migrate hex's interactive TUI from a custom Bubbletea implementation to the tux library, achieving:

1. **Code sharing** - Enable hex and other projects (jeff, etc.) to share UI components
2. **Reduced complexity** - Replace monolithic 1,600-line Model with tux's cleaner architecture
3. **User configurability** - Enable TOML-based theme/keybinding customization

## Approach

**Strategy:** Feature-branch spike with collaborative gap-fixing

- Work on `tux-migration` branch
- When tux gaps are discovered, report to maintainer for fixes in tux
- Target: Complete interactive TUI feature parity

**What stays unchanged:**
- Print mode (`-p`) - doesn't use TUI
- Core agent logic (`internal/core/`)
- Tool implementations (`internal/tools/`)
- MCP/plugin systems
- CLI argument parsing

**What gets replaced:**
- `internal/ui/` - The entire TUI layer (~72 files, ~13,300 lines)
- Integration point in `cmd/hex/interactive.go`

## Architecture Mapping

| Hex Current | Tux Equivalent |
|-------------|----------------|
| `Model` struct (1,600 lines) | `shell.Shell` + `Backend` interface |
| `Update()` message handling | Shell's built-in event routing |
| `View()` rendering | Shell's composable `Content` primitives |
| Overlay stack system | `modal.Manager` with stacked modals |
| Streaming display | `StreamingController` |
| Status bar | `shell.Status` + `StatusBar` component |
| Input textarea | Shell's built-in `InputArea` |
| Tool approval UI | `modal.ApprovalModal` |
| Conversation browser | Tab with `content.SelectList` |
| Theme (Dracula) | `theme.Theme` + TOML config |
| Keyboard handling | Shell focus management + configurable keybindings |

### Key Architectural Shift

```
BEFORE (hex):
┌─────────────────────────────────┐
│ Monolithic Model (40+ fields)   │
│ - UI state                      │
│ - Agent state                   │
│ - Tool state                    │
│ - Everything coupled together   │
└─────────────────────────────────┘

AFTER (tux):
┌──────────────┐    ┌──────────────┐
│ shell.Shell  │◄───│ HexBackend   │
│ (UI only)    │    │ (agent logic)│
└──────────────┘    └──────────────┘
       │
       ▼
┌──────────────────────────────────┐
│ Tabs, Modals, Content primitives │
└──────────────────────────────────┘
```

The `HexBackend` implements tux's `Backend` interface, keeping agent/tool logic separate from UI concerns.

## Implementation Phases

### Phase 0: Acceptance Tests

Write high-level behavior tests that verify user-visible functionality regardless of UI implementation:

- Send message → receive streaming response
- Tool call → approval modal appears → approve → result shown
- Tool call → deny → appropriate handling
- Navigate conversation history (up/down through sessions)
- Keyboard shortcuts work (Ctrl+C, gg, G, etc.)
- Status bar shows correct state (model, tokens, streaming indicator)

These tests become the "definition of done" for the migration.

### Phase 1: Foundation

- Create `tux-migration` branch
- Implement `HexBackend` struct wrapping hex's existing agent/core logic
- Wire up basic shell with single "Chat" tab
- Get text input → agent → streaming response working
- Verify status bar shows model/tokens/streaming state

### Phase 2: Tool System

- Implement tool call events flowing to UI
- Wire up `ApprovalModal` for tool approvals
- Display tool results in chat
- Add "Activity" tab showing tool timeline (Ctrl+O equivalent)

### Phase 3: Navigation & History

- Conversation history browser (session picker)
- Session resume functionality
- Favorites system
- Message scroll/navigation (gg, G, etc.)

### Phase 4: Polish

- Autocomplete for input
- Help overlay
- Quick actions menu
- Suggestions system
- Keyboard shortcut parity

### Phase 5: Configuration

- TOML config file support (`~/.config/hex/ui.toml`)
- Theme selection (Dracula default, others available)
- Keybinding customization
- User preference persistence

## Testing Strategy

### Current State

- 343 unit tests in `internal/ui/` (tightly coupled to current implementation)
- 48 integration tests in `test/integration/` (more reusable)
- All passing with race detection

### Migration Testing Approach

| Test Type | Purpose | Approach |
|-----------|---------|----------|
| **Acceptance tests** | Verify user-visible behavior | New - test through shell interface |
| **Backend tests** | Verify HexBackend wraps agent correctly | New - test Backend implementation |
| **Integration tests** | End-to-end flows work | Adapt existing |

### Test Lifecycle

1. **Phase 0:** Write acceptance tests against current hex
2. **During migration:** Acceptance tests are the "green bar"
3. **After migration:** Archive old unit tests, write new ones for tux components

## Deliverables

1. **Branch:** `tux-migration` with all changes
2. **New directory:** `internal/tui/` - tux-based UI implementation
3. **New file:** `internal/tui/backend.go` - HexBackend implementing tux's Backend interface
4. **Updated:** `cmd/hex/interactive.go` - wire up tux shell
5. **New directory:** `test/acceptance/` - high-level behavior tests
6. **Deleted:** `internal/ui/` (after migration complete, ~13,300 lines removed)

## Gap Reporting Workflow

When a tux gap blocks progress:

```
Developer: "Blocked - tux is missing X" (reports to maintainer)
Maintainer: Fixes X in tux
Maintainer: "Fixed, pull latest tux"
Developer: Continues migration
```

Gaps are fixed in tux properly, not shimmed in hex.

## Success Criteria

- [ ] All acceptance tests pass
- [ ] All existing integration tests pass
- [ ] Feature parity with current TUI:
  - [ ] Chat with streaming responses
  - [ ] Text input with submit
  - [ ] Tool approval modals
  - [ ] Tool results display
  - [ ] Status bar (model, tokens, streaming state)
  - [ ] Keyboard shortcuts
  - [ ] Conversation history browser
  - [ ] Session picker/resume
  - [ ] Autocomplete
  - [ ] Overlays (help, tool timeline)
  - [ ] Theming
- [ ] hex builds and runs normally
- [ ] No regression in print mode (`-p`)

## Feature Parity Checklist

### Core Chat
- [ ] Streaming response display with token rate
- [ ] Message history with timestamps
- [ ] Input textarea with history navigation
- [ ] Scroll management (gg to top, G to bottom)

### Tool System
- [ ] Inline tool execution with approval prompts
- [ ] Tool result display with collapsed/expanded logs
- [ ] Tool timeline overlay (Ctrl+O)
- [ ] Tool approval form
- [ ] Risk assessment display

### Navigation
- [ ] Conversation history browser
- [ ] Session picker for resuming
- [ ] Favorites system
- [ ] Search functionality

### Polish
- [ ] Autocomplete with fuzzy search
- [ ] Quick actions menu
- [ ] Suggestions system
- [ ] Help overlay
- [ ] Dracula theme (and others via config)

### Configuration
- [ ] `~/.config/hex/ui.toml` support
- [ ] Theme selection
- [ ] Keybinding customization

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Tux has significant gaps | Collaborative fix workflow; maintainer available |
| Migration takes too long | Timeboxed spike approach; can pivot to incremental if needed |
| Subtle behavior regressions | Acceptance tests catch user-visible issues |
| Performance regression | Profile before/after; tux uses same Bubbletea foundation |

## Dependencies

- tux library at `github.com/2389-research/tux` (private)
- Tux must be pulled as a Go module dependency
