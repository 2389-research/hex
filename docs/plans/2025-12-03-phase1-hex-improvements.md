# Phase 1: Hex UI Improvements Implementation Plan

Date: 2025-12-03
Target: Jeff Agent
Source: Hex UI/TUI audit

## Overview

Implement the 4 highest-priority UI improvements from hex:
1. StatusBar Component
2. Session Picker
3. Token Visualization
4. Message Formatting Improvements

## Task Breakdown

### Task 1: StatusBar Component (NEW FILE)
**File:** `internal/ui/statusbar.go`
**Dependencies:** themes package

**Implementation:**
```go
package ui

type StatusBar struct {
    model         string
    tokensInput   int
    tokensOutput  int
    tokensTotal   int
    contextSize   int
    connection    ConnectionStatus
    currentMode   string
    width         int
    theme         themes.Theme
}

// Methods to implement:
- NewStatusBar(model string, width int, theme themes.Theme) *StatusBar
- SetWidth(width int)
- SetModel(model string)
- UpdateTokens(input, output int)
- SetTokens(input, output int)
- SetContextSize(size int)
- SetConnection(status ConnectionStatus)
- SetMode(mode string)
- Render() string
```

**Integration Points:**
- Add `statusBar *StatusBar` to Model
- Call `NewStatusBar()` in `NewModel()`
- Update token counts when receiving API responses
- Render at bottom of `View()`
- Update width on `WindowSizeMsg`

**Tests:**
- Test token counter updates
- Test render with different connection states
- Test width adaptation
- Test theme integration

---

### Task 2: Message Formatting Improvements
**Files:** `internal/ui/update.go`, `internal/ui/model.go`

**Changes:**
1. Change bullet from `•` to `●` for assistant messages
2. Ensure consistent spacing between messages
3. Fix multi-line message indentation
4. Remove empty indented lines

**Before:**
```
• Assistant response
  with bad indent


• Next response
```

**After:**
```
● Assistant response
  with proper indent

● Next response
```

**Implementation:**
- Search for `"•"` and replace with `"●"` in assistant message rendering
- Review spacing logic around `CommitStreamingText()`
- Test with multi-line messages
- Test with code blocks

**Tests:**
- Test assistant message bullet rendering
- Test user message rendering (should not have bullet)
- Test multi-line message formatting
- Test spacing between messages

---

### Task 3: Session Picker (NEW FILE)
**File:** `internal/ui/session_picker.go`
**Dependencies:** storage package, bubbles/list

**Implementation:**
```go
package ui

type SessionPicker struct {
    list     list.Model
    selected string
    quitting bool
    newSession bool
}

type sessionItem struct {
    conv *storage.Conversation
}

// Methods to implement:
- NewSessionPicker(conversations []*storage.Conversation) SessionPicker
- Init() tea.Cmd (tea.Model interface)
- Update(msg tea.Msg) (tea.Model, tea.Cmd) (tea.Model interface)
- View() string (tea.Model interface)
- SelectedID() string
- IsNewSession() bool

// Helper functions:
- formatTimeAgo(t time.Time) string
- truncateModel(model string) string
- truncateID(id string) string
```

**Features:**
- List recent conversations
- Show favorite indicator (★)
- Display "Updated X ago"
- Show model and conversation ID
- Fuzzy search/filtering
- Arrow key navigation
- Enter to select, Esc to cancel
- "New Session" option at top

**Integration:**
- Add flag `--resume` or prompt on startup
- Run as separate tea.Program before main UI
- Return selected conversation ID
- Load conversation and continue

**Flow:**
```
1. jefft starts
2. Check for existing conversations
3. If none, start new session
4. If any, show SessionPicker
5. User selects or creates new
6. Load selected conversation
7. Start main UI
```

**Tests:**
- Test sessionItem Title() and Description()
- Test navigation with arrow keys
- Test filtering/search
- Test selection
- Test cancellation
- Test with empty conversation list
- Test with favorite conversations

---

### Task 4: Token Visualization (NEW FILE + DIR)
**File:** `internal/ui/visualization/tokens.go`
**Dependencies:** bubbles/progress, themes

**Implementation:**
```go
package visualization

type TokenUsage struct {
    InputTokens  int
    OutputTokens int
    TotalTokens  int
    MaxTokens    int
    ModelName    string
}

type TokenVisualization struct {
    theme         themes.Theme
    current       TokenUsage
    history       []TokenUsage
    progress      progress.Model
    width         int
    warningShown  bool
}

// Methods to implement:
- NewTokenVisualization(theme themes.Theme) *TokenVisualization
- Update(usage TokenUsage)
- SetWidth(width int)
- Render() string
- RenderCompact() string
- ShouldWarn() bool (true if > 80% context used)

// Display:
Tokens: [████████░░░░░░░░] 45% (90K/200K)
Input: 60K | Output: 30K | Remaining: 110K
⚠️  Approaching limit - consider new session
```

**Integration:**
- Add `tokenViz *visualization.TokenVisualization` to Model
- Update on every API response
- Render in statusbar or as separate line
- Show warning when > 80% context used

**Tests:**
- Test token usage calculation
- Test progress bar rendering
- Test warning threshold
- Test width adaptation
- Test with different context sizes

---

## Implementation Order

1. **Message Formatting** (Easiest, immediate impact)
   - Quick win
   - No new files
   - Improves readability immediately

2. **StatusBar** (Foundation)
   - Provides infrastructure for token display
   - Needed by TokenVisualization
   - Good visual improvement

3. **Token Visualization** (Builds on StatusBar)
   - Can integrate into statusbar
   - Provides context awareness
   - Uses StatusBar positioning

4. **Session Picker** (Most Complex)
   - Requires storage integration
   - Separate tea.Program
   - Startup flow changes
   - Biggest user impact

## Technical Considerations

### Testing Strategy
- Unit tests for each component
- Integration tests for Model interactions
- Visual tests for rendering
- User flow tests for Session Picker

### Theme Integration
- All components use themes.Theme interface
- Support all existing themes (Dracula, Gruvbox, Nord)
- Test color rendering

### Storage Integration (Session Picker)
- Ensure storage package has ListConversations()
- Add favorite flag to Conversation if needed
- Add UpdatedAt timestamp if needed
- Handle empty conversation list gracefully

### Error Handling
- Session Picker: handle no conversations
- Token Viz: handle missing token data
- StatusBar: handle connection failures

### Performance
- Session Picker: limit to 50 most recent
- Token Viz: limit history to 50 entries
- StatusBar: efficient rendering

## Success Criteria

✅ Message formatting:
- All assistant messages use ● bullet
- Consistent spacing between messages
- Multi-line messages properly indented

✅ StatusBar:
- Shows model, tokens, connection status
- Updates in real-time
- Adapts to window width
- Matches theme colors

✅ Token Visualization:
- Accurate token tracking
- Visual progress bar
- Warning at 80% threshold
- Clear, readable display

✅ Session Picker:
- Lists all conversations
- Fuzzy search works
- Can select and resume
- Can create new session
- Shows favorites and metadata

## Rollback Plan

If any component causes issues:
1. Feature can be disabled with flag
2. Each component is isolated
3. StatusBar can be hidden
4. Session Picker can be bypassed with `--new` flag
5. Token Viz can be turned off

## Timeline Estimate

- Message Formatting: 1-2 hours
- StatusBar: 3-4 hours
- Token Visualization: 4-5 hours
- Session Picker: 6-8 hours

**Total: 14-19 hours of development time**

With subagents: Can complete in 1 session if parallelized correctly.

## Next Steps

1. Create branch: `feat/phase1-hex-improvements`
2. Implement in order (1→2→3→4)
3. Test each component individually
4. Integration test all together
5. Review and merge

---

## Code Review Checklist

- [ ] All components have unit tests
- [ ] Theme support verified for all themes
- [ ] No hardcoded colors
- [ ] Error handling in place
- [ ] Documentation/comments added
- [ ] Performance tested
- [ ] Works in tmux
- [ ] Works with window resize
- [ ] Backwards compatible
- [ ] No breaking changes to existing UI
