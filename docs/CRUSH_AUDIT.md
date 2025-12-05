# Crush Agent Audit - Learnings for Hex

**Date:** December 3, 2025
**Auditor:** Claude (via Hex)
**Source:** `/Users/harper/workspace/2389/agent-class/agents/crush`

## Executive Summary

Crush is Charm/bubbletea's official AI coding agent. This audit identifies architectural patterns, features, and implementation details that could significantly improve Hex's robustness and user experience.

**Key Findings:**
- Sophisticated page-based TUI architecture vs our view-mode system
- **LSP integration for code intelligence** (MAJOR missing feature in Hex)
- PubSub event architecture for decoupled communication
- Advanced permission system with persistent grants
- Hierarchical session management with parent/child relationships
- Database-backed storage using SQLite with migrations
- Categorized tool safety levels
- Progress bar support detection for modern terminals

---

## 1. Architecture Patterns

### 1.1 TUI Organization

**Crush Structure:**
```
/tui
  ├── tui.go              # Main orchestrator
  ├── /page               # Full-screen pages (chat, etc.)
  ├── /components         # Reusable UI components
  │   ├── /core           # Layout, status bar
  │   ├── /chat           # Chat-specific components
  │   ├── /dialogs        # Modal overlays
  │   └── /anim           # Animations
  └── /styles             # Theming system
```

**Key Differences from Hex:**
- **Pages vs View Modes**: Crush uses discrete Page IDs with lazy loading
- **Dialog System**: Overlays (permissions, model selector, quit confirm) separate from pages
- **Component Hierarchy**: Clear separation of concerns (layout, content, dialogs)

**appModel Structure:**
```go
type appModel struct {
    currentPage   page.PageID
    previousPage  page.PageID
    pages         map[page.PageID]util.Model
    loadedPages   map[page.PageID]bool // Lazy loading!

    dialog        dialogs.DialogCmp   // Modal overlays
    completions   completions.Completions
    status        status.StatusCmp
}
```

**Recommendation for Hex:**
- Consider refactoring from ViewMode enum to Page-based system
- Implement dialog overlay system for approvals, tool output
- Extract status bar into dedicated component
- Add lazy page loading for faster startup

### 1.2 Event Architecture

**PubSub Pattern:**
```go
// Decoupled event system
type Service interface {
    pubsub.Subscriber[Session]
    Create(ctx context.Context, title string) (Session, error)
    // ...
}

// Components subscribe to events
session.Subscribe(ctx) // Returns channel of Session events
```

**Benefits:**
- TUI updates reactively to backend changes
- Multiple components can react to same event
- Cleaner separation of business logic and UI

**Hex Current:** Direct coupling between model updates and UI rendering

**Recommendation for Hex:**
- Implement lightweight PubSub for: session changes, tool execution, LSP diagnostics
- Allow multiple UI components to subscribe to same events

---

## 2. **LSP Integration (CRITICAL FEATURE)**

### 2.1 Overview

Crush integrates Language Server Protocol (LSP) servers for **code intelligence**:
- Hover information
- Go-to-definition
- Find references
- Diagnostics (errors/warnings)
- Completions

**Impact:** LLM gets same context a developer has in their IDE!

### 2.2 Implementation

**Client Wrapper:** `/internal/lsp/client.go`
```go
type Client struct {
    client      *powernap.Client  // Using charmbracelet/x/powernap
    name        string
    fileTypes   []string          // Which extensions this LSP handles

    diagnostics *csync.VersionedMap[DocumentURI, []Diagnostic]
    openFiles   *csync.Map[string, *OpenFileInfo]
    serverState atomic.Value      // StateStarting/Ready/Error/Disabled
}
```

**Configuration:** `crush.json`
```json
{
  "lsp": {
    "go": {
      "command": "gopls",
      "env": { "GOTOOLCHAIN": "go1.24.5" }
    },
    "typescript": {
      "command": "typescript-language-server",
      "args": ["--stdio"]
    }
  }
}
```

**Tool Integration:**
- `references` tool: Uses LSP `textDocument/references`
- Hover data included in file context
- Diagnostics shown in status bar

### 2.3 Recommendation for Hex

**Priority: HIGH** - This would be a game-changer for code quality

**Implementation Plan:**
1. Add `powernap` dependency (Charm's LSP client library)
2. Create `/internal/lsp` package similar to Crush's
3. Add LSP config to our JSON schema
4. Integrate with `view` tool to include:
   - Hover information at cursor position
   - Diagnostics for viewed files
5. Create `references` tool for finding usages
6. Show LSP status in status bar

**Estimated Effort:** 2-3 days for basic integration

---

## 3. Permission System

### 3.1 Advanced Features

**Persistent Permissions:**
```go
// User can grant permission for entire session
GrantPersistent(permission PermissionRequest)

// Or grant once
Grant(permission PermissionRequest)
```

**Allowlist Support:**
```json
{
  "permissions": {
    "allowed_tools": [
      "view",           // Entire tool
      "ls",
      "grep",
      "edit:create"     // Specific tool:action combination
    ]
  }
}
```

**Safe Commands List:**
- Predefined whitelist of read-only commands (git status, ps, etc.)
- Auto-approved without prompting

**Per-Session Auto-Approve:**
```go
AutoApproveSession(sessionID string)  // --yolo mode per session
```

### 3.2 Hex Comparison

**Current Hex:**
- Simple per-tool approval
- No persistent permissions
- No granular action-level permissions

**Recommendation:**
- Add `allowed_tools` config support
- Implement persistent "Grant for session" option
- Add safe command whitelist for read-only tools
- Consider per-session YOLO mode

---

## 4. Session Management

### 4.1 Hierarchical Sessions

**Parent-Child Relationships:**
```go
type Session struct {
    ID               string
    ParentSessionID  string  // Links to parent session!
    Title            string
    MessageCount     int64
    PromptTokens     int64
    CompletionTokens int64
    SummaryMessageID string
    Cost             float64
}
```

**Use Cases:**
- Title generation in separate session
- Agent tool calls spawn child sessions
- Sub-tasks maintain separate context

**Agent Tool Sessions:**
```go
sessionID := CreateAgentToolSessionID(messageID, toolCallID)
// Allows tools to have their own conversation context
```

### 4.2 Token Tracking

**Detailed Metrics:**
- Prompt tokens per session
- Completion tokens per session
- Cost calculation per session
- Summary message tracking

**Recommendation for Hex:**
- Add token/cost tracking to Conversation table
- Consider hierarchical sessions for complex workflows
- Track which messages were summaries

---

## 5. Database & Storage

### 5.1 Migration System

**SQL Migrations:** `/internal/db/migrations/`
```
20250424200609_initial.sql
20250515105448_add_summary_message_id.sql
20250624000000_add_created_at_indexes.sql
20250627000000_add_provider_to_messages.sql
```

**Benefits:**
- Schema versioning
- Reproducible database state
- Easy rollback/upgrade path

**Hex Current:** Direct schema creation, no versioning

**Recommendation:**
- Adopt migration system (golang-migrate or similar)
- Version our schema changes
- Makes future schema changes safer

### 5.2 Code Generation

**SQLC Usage:**
- SQL queries in `.sql` files
- Type-safe Go code generated
- No ORM overhead

**Hex Current:** Manual SQL with string concatenation

**Recommendation:**
- Consider SQLC for type safety
- Or at minimum, extract queries to constants

---

## 6. Agent Execution Loop

### 6.1 Queuing System

**Message Queue:**
```go
messageQueue *csync.Map[string, []SessionAgentCall]

// If session busy, queue additional prompts
if a.IsSessionBusy(sessionID) {
    existing := messageQueue.Get(sessionID)
    existing = append(existing, call)
    messageQueue.Set(sessionID, existing)
    return nil  // Queued, not executed yet
}
```

**Benefits:**
- User can keep typing while agent works
- Preserves message order
- Prevents concurrent execution chaos

**Hex Current:** Single-threaded, blocks on user input

**Recommendation:**
- Implement message queuing per conversation
- Show queue depth in UI
- Allow queue clearing/cancellation

### 6.2 Auto-Summarization

**Token Management:**
```go
// Automatically summarize context when getting large
if shouldSummarize && !a.disableAutoSummarize {
    a.Summarize(ctx, sessionID, providerOptions)
}
```

**Hex Current:** Manual compaction trigger

**Recommendation:**
- Auto-trigger summarization at token threshold
- Make threshold configurable
- Show "context compressed" indicator in UI

---

## 7. UI/UX Enhancements

### 7.1 Mouse Event Throttling

**Performance Optimization:**
```go
func MouseEventFilter(m tea.Model, msg tea.Msg) tea.Msg {
    switch msg.(type) {
    case tea.MouseWheelMsg, tea.MouseMotionMsg:
        // Trackpad sends too many events, throttle to 15ms
        if now.Sub(lastMouseEvent) < 15*time.Millisecond {
            return nil
        }
    }
    return msg
}
```

**Recommendation for Hex:**
- Add mouse event throttling for smoother scrolling
- Prevents performance issues on trackpads

### 7.2 Terminal Capability Detection

**Progress Bar Support:**
```go
case tea.TerminalVersionMsg:
    termVersion := strings.ToLower(msg.Name)
    // Only show progress bars in modern terminals
    if strings.Contains(termVersion, "ghostty") ||
       slices.Contains(msg, "WT_SESSION") {
        sendProgressBar = true
    }
```

**Recommendation for Hex:**
- Detect terminal capabilities
- Gracefully degrade features in older terminals
- Could enable/disable: images, unicode, progress bars

### 7.3 Keyboard Enhancement Detection

**Ctrl+M vs Return:**
```go
case tea.KeyboardEnhancementsMsg:
    // Keyboard disambiguation available
    if msg.Flags > 0 {
        keyMap.Models.SetHelp("ctrl+m", "models")
    }
```

**Hex Current:** Can't distinguish Ctrl+M from Enter

**Recommendation:**
- Query terminal for keyboard enhancements
- Use Ctrl+M for model switcher if available
- Fallback to current bindings otherwise

---

## 8. Code Organization Patterns

### 8.1 Concurrency Primitives

**Custom Sync Types:** `/internal/csync/`
```go
// Thread-safe map with cleaner API
type Map[K comparable, V any] struct {
    m  map[K]V
    mu sync.RWMutex
}

// Versioned map for cache invalidation
type VersionedMap[K comparable, V any] struct {
    m       map[K]V
    version map[K]int64
    mu      sync.RWMutex
}
```

**Recommendation:**
- Extract our sync patterns into reusable primitives
- Reduces boilerplate, clearer intent

### 8.2 Domain Services

**Service Pattern:**
```go
// Define interface for behavior
type Service interface {
    Create(ctx context.Context, ...) (Session, error)
    Get(ctx context.Context, id string) (Session, error)
    // ...
}

// Concrete implementation
type service struct {
    q db.Querier  // Injected database
}
```

**Benefits:**
- Easy to mock for testing
- Clear boundaries between layers
- Dependency injection

**Hex Current:** Mixed business logic in UI model

**Recommendation:**
- Extract storage, session, message logic into services
- Makes testing WAY easier
- Cleaner separation of concerns

---

## 9. Configuration System

### 9.1 JSON Schema

**Schema Validation:**
```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": { ... },
  "lsp": { ... },
  "mcp": { ... }
}
```

**Benefits:**
- IDE autocomplete in config files
- Validation before app starts
- Documentation embedded in schema

**Hex Current:** No schema, manual validation

**Recommendation:**
- Create JSON schema for hex.json / crush.json
- Host schema online for IDE integration
- Validate config on load

### 9.2 Variable Expansion

**Environment Variable Support:**
```json
{
  "mcp": {
    "github": {
      "headers": {
        "Authorization": "Bearer $GH_PAT"
      }
    }
  }
}
```

**Recommendation:**
- Support $VAR and ${VAR} expansion in config
- Enables secure credential management

---

## 10. Testing Patterns

### 10.1 Mock Providers

**Test Mode:**
```go
config.UseMockProviders = true
defer config.ResetProviders()

// Now all LLM calls return mock data
```

**Hex Current:** Tests skip LLM-dependent code

**Recommendation:**
- Create mock LLM provider for testing
- Enables testing full conversation flows
- Faster, deterministic tests

### 10.2 Golden Files

**Output Testing:**
```bash
# Regenerate expected outputs
go test ./... -update

# Normal test compares to .golden files
go test ./...
```

**Recommendation:**
- Use golden files for rendering tests
- Makes TUI changes easier to review

---

## Priority Recommendations

### Immediate (This Sprint)

1. **Mouse event throttling** - Easy win, better UX
2. **Persistent permissions** - Users keep asking for this
3. **Safe command whitelist** - Reduce approval fatigue

### High Priority (Next Sprint)

4. **LSP Integration** - GAME CHANGER for code quality
5. **PubSub event system** - Cleaner architecture
6. **Message queuing** - Better UX when agent is busy

### Medium Priority

7. **Migration system** - Before next schema change
8. **Hierarchical sessions** - For complex workflows
9. **Token/cost tracking** - User requested feature
10. **JSON schema for config** - Better DX

### Low Priority (Nice to Have)

11. **Service layer refactor** - Improved testability
12. **Terminal capability detection** - Progressive enhancement
13. **SQLC adoption** - Type safety

---

## Conclusion

Crush demonstrates production-grade patterns that would significantly improve Hex:

**Biggest Opportunities:**
1. **LSP integration** - Would make code assistance dramatically better
2. **Event-driven architecture** - Cleaner, more maintainable code
3. **Permission UX** - Persistent grants, allowlists, safe commands
4. **Session hierarchy** - Better context management for complex tasks

**Architecture Lessons:**
- Page-based UI > Mode-based UI
- PubSub > Direct coupling
- Service layer > Mixed business logic
- Type-safe DB access > String SQL

**Next Steps:**
- Prioritize LSP integration (highest impact)
- Refactor permission system (quick wins)
- Plan migration to event-driven architecture (foundational improvement)

---

**End of Audit**
