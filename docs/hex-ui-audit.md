# Hex UI/TUI Audit - Improvements to Incorporate

Date: 2025-12-03
Project: Pagen Agent (Productivity-focused AI assistant)
Source: Hex (Code-focused AI assistant at ../clem)

## Executive Summary

Audited 40+ commits from Hex focusing on UI/TUI improvements. Identified 10 key enhancements that would significantly improve pagen-agent's user experience, particularly for productivity workflows.

## High Priority Improvements

### 1. StatusBar Component ⭐⭐⭐
**File:** `internal/ui/statusbar.go`
**Value:** Essential for productivity - shows context at a glance

**Features:**
- Model name display
- Token counters (input/output/total)
- Context window usage
- Connection status indicators
- Current mode display
- Custom messages

**Why Important:**
For a productivity app, users need constant visibility into:
- How much context they have left
- Current model being used
- API connection health
- Token usage (cost awareness)

**Implementation:**
```go
type StatusBar struct {
    model         string
    tokensInput   int
    tokensOutput  int
    contextSize   int
    connection    ConnectionStatus
    currentMode   string
}
```

---

### 2. Session Picker ⭐⭐⭐
**File:** `internal/ui/session_picker.go`
**Value:** CRITICAL for productivity - quick access to past conversations

**Features:**
- Lists recent conversations with titles
- Shows "Updated X ago" with model info
- Fuzzy search/filtering
- Favorite indicators (★)
- Interactive selection with arrow keys

**Why Important:**
Productivity users constantly switch between tasks and conversations. Being able to quickly:
- Resume a meeting notes conversation
- Switch to yesterday's project planning
- Find that conversation about a specific topic

This is THE killer feature for a productivity AI assistant.

**Implementation Uses:**
- Bubbles list component
- Storage integration
- Tea.Program as a modal

---

### 3. Token Visualization ⭐⭐⭐
**File:** `internal/ui/visualization/tokens.go`
**Value:** Cost awareness and context management

**Features:**
- Real-time token usage progress bar
- Input vs output token breakdown
- Context window fill percentage
- Warning when approaching limit
- Historical usage graphs
- Detailed view toggle

**Why Important:**
- Users need to know when they're close to context limits
- Cost awareness (tokens = $)
- Helps decide when to start fresh vs continue conversation

**Visual:**
```
Tokens: [████████░░░░░░░░] 45% (90K/200K)
Input: 60K | Output: 30K | Remaining: 110K
⚠️  Approaching context limit - consider starting a new session
```

---

### 4. Message Formatting Improvements ⭐⭐
**Commits:**
- `691b4853` - Upgrade to BIGBOY bullet (● vs •)
- `2b300cfe` - Visual indicators for user/assistant
- `4fecf339` - Proper multi-line indentation
- `17e18b5c` - Remove indented empty lines
- `57459d6a` - Consistent spacing

**Features:**
- Bigger bullet for assistant messages (●)
- Proper indentation for multi-line messages
- Consistent spacing between messages
- Removed glamour paragraph indentation
- Fixed double-spacing bugs

**Why Important:**
- Better readability
- Clear visual distinction between user and assistant
- Professional appearance
- Less confusion when scrolling through history

**Before:**
```
• Here's my response
  with multiple lines
    that look weird


• Another response
```

**After:**
```
● Here's my response
  with multiple lines
  properly indented

● Another response
```

---

## Medium Priority Improvements

### 5. Tool Approval Form WindowSizeMsg Fixes ⭐⭐
**Commits:**
- `9fff2cc3` - Forward WindowSizeMsg to approval form
- `dd5f79ce` - Send initial WindowSizeMsg on creation
- `0533b264` - Forward ALL messages to approval form

**Problem Solved:**
Tool approval forms were freezing in tmux because they didn't receive window size messages on initialization or resize.

**Fix:**
```go
// Send initial window size when creating approval form
if m.toolApprovalForm != nil {
    cmd := m.toolApprovalForm.Update(tea.WindowSizeMsg{
        Width:  m.width,
        Height: m.height,
    })
    return m, cmd
}

// Forward ALL messages, not just KeyMsg
if m.toolApprovalMode && m.toolApprovalForm != nil {
    form, cmd := m.toolApprovalForm.Update(msg)
    m.toolApprovalForm = form.(*forms.ApprovalForm)
    return m, cmd
}
```

---

### 6. Layout/Borders System ⭐⭐
**File:** `internal/ui/layout/borders.go`
**Features:**
- Consistent border styles across components
- Spacing utilities
- Box rendering helpers
- Themed borders

**Why Important:**
- Professional, polished look
- Consistent visual language
- Easier to build complex layouts

---

### 7. Intro Screen ⭐
**Commits:**
- `83bfe451` - Add startup intro screen
- `175d6530` - Persist intro until first message

**Features:**
- ASCII logo on startup
- Quick tips/shortcuts
- Persists until first user message
- Sets expectation for new users

**Why Important:**
- Better first-run experience
- Teaches keyboard shortcuts
- Brand identity

---

## Lower Priority (Nice to Have)

### 8. Help Panel
**File:** `internal/ui/components/help.go`
- Context-aware key bindings
- Toggle with `?`
- Shows available commands

### 9. Gradient Animations
**File:** `internal/ui/animations/gradient.go`
- Smooth color transitions
- Animated title bars
- Visual polish

### 10. Plugin Dashboard
**File:** `internal/ui/dashboard/plugins.go`
- Shows MCP server status
- Plugin health monitoring
- Useful for debugging

---

## Implementation Plan

### Phase 1: Essential Features (Week 1)
1. Add StatusBar component
2. Implement Session Picker
3. Add Token Visualization
4. Apply message formatting improvements

### Phase 2: Quality of Life (Week 2)
5. Tool approval form WindowSizeMsg fixes
6. Layout/borders system
7. Intro screen

### Phase 3: Polish (Week 3)
8. Help panel
9. Optional: Gradient animations
10. Optional: Plugin dashboard

---

## Files to Create/Modify

### New Files:
- `internal/ui/statusbar.go` - Status bar component
- `internal/ui/session_picker.go` - Session picker modal
- `internal/ui/visualization/tokens.go` - Token visualization
- `internal/ui/layout/borders.go` - Layout utilities

### Files to Modify:
- `internal/ui/view.go` - Add status bar, integrate session picker
- `internal/ui/model.go` - Add status bar and token tracking state
- `internal/ui/update.go` - Message formatting improvements, WindowSizeMsg forwarding
- `internal/ui/components/huh_approval.go` - WindowSizeMsg handling

---

## Technical Dependencies

### Already Have:
- Bubbles (for list component in session picker)
- Huh (for forms)
- Lipgloss (for styling)
- Storage package (for conversations)

### May Need:
- Progress bar from bubbles (for token visualization)

---

## Risk Assessment

### Low Risk:
- Message formatting improvements (cosmetic only)
- StatusBar (additive, doesn't change existing behavior)
- Intro screen (only shown once)

### Medium Risk:
- Session picker (requires storage integration, could have bugs)
- Token visualization (needs accurate tracking)
- Layout/borders (could break existing layouts)

### High Risk:
- WindowSizeMsg forwarding (could cause race conditions if not careful)

---

## Success Metrics

After implementation, measure:
1. Time to resume past conversation (should be < 5 seconds)
2. User awareness of context limits (token viz visible?)
3. Reduced context limit errors (users start new sessions proactively)
4. User feedback on message readability
5. Tmux/resize bug reports (should decrease)

---

## Notes

- Hex is for code, pagen-agent is for productivity
- Not all hex features make sense (e.g., code-specific tooling)
- Focus on what helps productivity workflows:
  - Quick context switching (session picker)
  - Cost/context awareness (token viz)
  - Clear communication (message formatting)
  - Status visibility (status bar)

---

## Conclusion

Hex has implemented a mature, polished TUI with excellent UX. The most valuable improvements for pagen-agent are:

1. **Session Picker** - Game changer for productivity
2. **Token Visualization** - Essential for context management
3. **StatusBar** - Professional status display
4. **Message Formatting** - Better readability

These 4 features alone would significantly elevate pagen-agent's user experience and make it feel like a professional, production-ready tool.
