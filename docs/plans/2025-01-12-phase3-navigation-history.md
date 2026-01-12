# Phase 3: Navigation & History Implementation Plan

**Goal:** Add conversation history, session management, and navigation to tux-based hex.

**Architecture:** Extend HexAgent and tux UI to support:
- Session persistence (save/load conversations)
- History browser (list and select past sessions)
- Message scrolling within current session
- Keyboard navigation (gg, G, up/down)

---

## Background

### Current State (Phase 2)
- HexAgent handles streaming, tool calls, approvals, execution
- Single-session conversations work end-to-end
- No persistence - history lost on restart

### Target State (Phase 3)
- Sessions persist to disk (JSON files)
- History tab shows past sessions
- Can resume any past session
- Keyboard navigation for scrolling
- Favorites marking for important sessions

### Key Data Flow
```
User input → HexAgent → API → Response
                ↓
         Session saved to disk
                ↓
         History browser updated
```

---

## Task 1: Session Data Model

**Files:**
- Create: `internal/tui/session.go`

**Step 1: Define session types**

```go
// Session represents a conversation session that can be persisted.
type Session struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`      // Auto-generated from first message
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Messages  []SessionMessage `json:"messages"`
    Favorite  bool      `json:"favorite"`
}

type SessionMessage struct {
    Role      string    `json:"role"`      // "user" or "assistant"
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    ToolCalls []SessionToolCall `json:"tool_calls,omitempty"`
}

type SessionToolCall struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Input  map[string]interface{} `json:"input"`
    Output string `json:"output"`
    Error  bool   `json:"error"`
}
```

**Step 2: Verify build**

```bash
go build ./internal/tui/...
```

**Step 3: Commit**

```bash
git add internal/tui/session.go
git commit -m "feat: add session data model"
```

---

## Task 2: Session Storage

**Files:**
- Create: `internal/tui/storage.go`

**Step 1: Implement storage interface**

```go
// SessionStorage handles session persistence.
type SessionStorage struct {
    dir string // ~/.hex/sessions/
}

// NewSessionStorage creates storage at the given directory.
func NewSessionStorage(dir string) (*SessionStorage, error)

// Save persists a session to disk.
func (s *SessionStorage) Save(session *Session) error

// Load retrieves a session by ID.
func (s *SessionStorage) Load(id string) (*Session, error)

// List returns all sessions, sorted by UpdatedAt descending.
func (s *SessionStorage) List() ([]*Session, error)

// Delete removes a session.
func (s *SessionStorage) Delete(id string) error
```

**Step 2: Implement methods**

- Save: Write JSON to `{dir}/{id}.json`
- Load: Read and parse JSON file
- List: Glob `{dir}/*.json`, parse all, sort by UpdatedAt
- Delete: Remove file

**Step 3: Verify build**

```bash
go build ./internal/tui/...
```

**Step 4: Commit**

```bash
git add internal/tui/storage.go
git commit -m "feat: add session storage"
```

---

## Task 3: Add Storage Tests

**Files:**
- Create: `internal/tui/storage_test.go`

**Step 1: Test save/load round-trip**

```go
func TestSessionStorage_SaveLoad(t *testing.T) {
    // Create temp dir
    // Save session
    // Load session
    // Verify fields match
}
```

**Step 2: Test list ordering**

```go
func TestSessionStorage_List(t *testing.T) {
    // Save multiple sessions with different timestamps
    // List should return newest first
}
```

**Step 3: Run tests**

```bash
go test ./internal/tui/... -v
```

**Step 4: Commit**

```bash
git add internal/tui/storage_test.go
git commit -m "test: add session storage tests"
```

---

## Task 4: Wire Storage to HexAgent

**Files:**
- Modify: `internal/tui/agent.go`
- Modify: `cmd/hex/tux.go`

**Step 1: Add storage and session fields to HexAgent**

```go
type HexAgent struct {
    // ... existing fields ...

    // Session management
    storage        *SessionStorage
    currentSession *Session
}
```

**Step 2: Update constructor**

```go
func NewHexAgent(client *core.Client, model string, systemPrompt string,
                 executor *tools.Executor, storage *SessionStorage) *HexAgent
```

**Step 3: Auto-save on message completion**

In processStream's message_stop handler, after adding to history:
- Update currentSession.Messages
- Call storage.Save(currentSession)

**Step 4: Update call sites**

In `cmd/hex/tux.go`:
```go
storage, err := tui.NewSessionStorage(sessionDir)
agent := tui.NewHexAgent(client, model, systemPrompt, executor, storage)
```

**Step 5: Verify build**

```bash
go build ./...
```

**Step 6: Commit**

```bash
git add internal/tui/agent.go cmd/hex/tux.go
git commit -m "feat: wire session storage to HexAgent"
```

---

## Task 5: Session Creation and Resume

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Add NewSession method**

```go
// NewSession starts a fresh session.
func (a *HexAgent) NewSession() {
    a.currentSession = &Session{
        ID:        uuid.New().String(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Messages:  make([]SessionMessage, 0),
    }
    a.messages = nil // Clear conversation history
}
```

**Step 2: Add ResumeSession method**

```go
// ResumeSession loads and continues an existing session.
func (a *HexAgent) ResumeSession(id string) error {
    session, err := a.storage.Load(id)
    if err != nil {
        return err
    }
    a.currentSession = session
    // Convert SessionMessages to core.Messages
    a.messages = convertToMessages(session.Messages)
    return nil
}
```

**Step 3: Verify build**

```bash
go build ./internal/tui/...
```

**Step 4: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: add session create and resume"
```

---

## Task 6: History Content for Tux

**Files:**
- Create: `internal/tui/history_content.go`

**Step 1: Implement tux content.Content for history browser**

```go
// HistoryContent displays session list for selection.
type HistoryContent struct {
    storage   *SessionStorage
    sessions  []*Session
    cursor    int
    width     int
    height    int
    onSelect  func(session *Session)
}

func NewHistoryContent(storage *SessionStorage, onSelect func(*Session)) *HistoryContent

// Implement content.Content interface:
// - Init() tea.Cmd
// - Update(msg tea.Msg) (content.Content, tea.Cmd)
// - View() string
// - Value() any
// - SetSize(width, height int)
```

**Step 2: Handle keyboard navigation**

- `j` / `down`: Move cursor down
- `k` / `up`: Move cursor up
- `enter`: Select session (call onSelect)
- `d`: Delete session (with confirmation)
- `f`: Toggle favorite

**Step 3: Render session list**

```go
func (h *HistoryContent) View() string {
    // Render list of sessions
    // Show: title, date, favorite star
    // Highlight selected
}
```

**Step 4: Verify build**

```bash
go build ./internal/tui/...
```

**Step 5: Commit**

```bash
git add internal/tui/history_content.go
git commit -m "feat: add history browser content"
```

---

## Task 7: Wire History Tab

**Files:**
- Modify: `cmd/hex/tux.go`

**Step 1: Create History tab in tux app setup**

```go
// Create history content
historyContent := tui.NewHistoryContent(storage, func(session *tui.Session) {
    agent.ResumeSession(session.ID)
    // Switch to Chat tab
})

// Add History tab
app.AddTab(shell.Tab{
    ID:       "history",
    Label:    "History",
    Shortcut: "ctrl+h",
    Content:  historyContent,
})
```

**Step 2: Verify build**

```bash
go build ./...
```

**Step 3: Commit**

```bash
git add cmd/hex/tux.go
git commit -m "feat: add history tab to tux UI"
```

---

## Task 8: Chat Scrolling

**Files:**
- Modify: `internal/tui/agent.go` or work with tux ChatContent

**Step 1: Check tux ChatContent capabilities**

Read tux's chat_content.go to understand scroll support.

**Step 2: If needed, add scroll support**

Options:
- Use tux's built-in viewport scrolling
- Add custom scroll handling via keyboard

**Step 3: Wire keyboard shortcuts**

- `gg`: Scroll to top
- `G`: Scroll to bottom
- `j` / `down`: Scroll down
- `k` / `up`: Scroll up
- `Ctrl+u`: Page up
- `Ctrl+d`: Page down

**Step 4: Verify build**

```bash
go build ./...
```

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add chat scrolling"
```

---

## Task 9: History Tests

**Files:**
- Create: `internal/tui/history_content_test.go`

**Step 1: Test list rendering**

```go
func TestHistoryContent_View(t *testing.T) {
    // Create mock storage with sessions
    // Verify view renders correctly
}
```

**Step 2: Test navigation**

```go
func TestHistoryContent_Navigation(t *testing.T) {
    // Send key messages
    // Verify cursor moves correctly
}
```

**Step 3: Run tests**

```bash
go test ./internal/tui/... -v
```

**Step 4: Commit**

```bash
git add internal/tui/history_content_test.go
git commit -m "test: add history content tests"
```

---

## Task 10: Manual Integration Test

**Files:**
- None (manual testing)

**Step 1: Build**

```bash
make build
```

**Step 2: Test workflow**

```bash
cd /Users/dylanr/work/hex/.worktrees/tux-migration
source .env
./bin/hex --tux
```

**Test sequence:**
1. Send a message, verify response
2. Press Ctrl+H to go to History tab
3. Verify current session appears
4. Start new session (if implemented)
5. Return to History, verify both sessions
6. Select old session, verify resume works
7. Test keyboard navigation (gg, G)

**Step 3: Verify**
- [ ] Sessions persist to disk
- [ ] History tab shows sessions
- [ ] Can resume sessions
- [ ] Keyboard navigation works
- [ ] Favorites toggle works

**Step 4: Final commit**

```bash
git add -A
git commit -m "chore: phase 3 navigation and history complete"
```

---

## Summary

Phase 3 adds session persistence and history browsing:

| Task | Description |
|------|-------------|
| 1 | Session data model |
| 2 | Session storage implementation |
| 3 | Storage tests |
| 4 | Wire storage to HexAgent |
| 5 | Session create/resume methods |
| 6 | History browser content |
| 7 | Wire history tab |
| 8 | Chat scrolling |
| 9 | History tests |
| 10 | Manual integration test |

After Phase 3, the tux-based hex will be able to:
- Stream text responses ✓ (Phase 1)
- Handle tool calls with approval ✓ (Phase 2)
- Persist sessions to disk
- Browse and resume past sessions
- Navigate with keyboard shortcuts
