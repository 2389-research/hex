# Phase 2: Hex UI Improvements Implementation Plan

Date: 2025-12-03
Target: Jeff Agent
Source: Hex UI/TUI audit (docs/hex-ui-audit.md)

## Overview

Implement 3 medium-priority UI improvements from hex:
1. WindowSizeMsg Fixes for Tool Approval Forms
2. Layout/Borders System
3. Intro Screen

## Task Breakdown

### Task 1: WindowSizeMsg Fixes for Tool Approval Forms

**Problem:** Tool approval forms freeze in tmux because they don't receive window size messages on initialization or resize.

**Reference commits from hex:**
- `9fff2cc3` - Forward WindowSizeMsg to approval form
- `dd5f79ce` - Send initial WindowSizeMsg on creation
- `0533b264` - Forward ALL messages to approval form

**Files to modify:**
- `internal/ui/model.go` - Add initial WindowSizeMsg sending
- `internal/ui/update.go` - Forward ALL messages to approval form

**Implementation:**

1. In `model.go`, when entering approval mode (`EnterHuhApprovalMode()`):
   ```go
   func (m *Model) EnterHuhApprovalMode() {
       // ... existing code ...
       m.huhApproval = components.NewHuhApproval(m.theme, toolName, description)

       // NEW: Send initial WindowSizeMsg for tmux compatibility
       m.huhApproval.Update(tea.WindowSizeMsg{
           Width:  m.Width,
           Height: m.Height,
       })
   }
   ```

2. In `update.go`, update the approval mode handling:
   ```go
   // Phase 2: Forward ALL messages to approval form, not just KeyMsg
   if m.toolApprovalMode && m.huhApproval != nil {
       // Forward ALL messages (KeyMsg, WindowSizeMsg, etc.)
       approvalModel, cmd := m.huhApproval.Update(msg)
       if approval, ok := approvalModel.(*components.HuhApproval); ok {
           m.huhApproval = approval
           // ... handle completion ...
       }
       return m, cmd
   }
   ```

**Tests:**
- Test approval form receives initial WindowSizeMsg
- Test approval form receives resize WindowSizeMsg
- Test approval works in different terminal sizes
- Integration test simulating tmux scenario

**Success Criteria:**
✅ Approval forms don't freeze in tmux
✅ Approval forms resize properly
✅ All message types forwarded correctly

---

### Task 2: Layout/Borders System

**Purpose:** Consistent border styles and spacing utilities for professional, polished layouts.

**Reference:** `/Users/harper/workspace/2389/clem/internal/ui/layout/borders.go`

**Files to create:**
- `internal/ui/layout/borders.go`
- `internal/ui/layout/borders_test.go`

**Implementation:**

Create new package `internal/ui/layout` with:

```go
package layout

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/harper/jefft/internal/ui/themes"
)

// BorderStyle defines different border types
type BorderStyle string

const (
    BorderNone   BorderStyle = "none"
    BorderNormal BorderStyle = "normal"
    BorderRounded BorderStyle = "rounded"
    BorderDouble BorderStyle = "double"
    BorderHidden BorderStyle = "hidden"
)

// Box creates a bordered box with consistent styling
func Box(content string, theme themes.Theme, style BorderStyle) string {
    // Implementation
}

// Spacing utilities
func Padding(content string, top, right, bottom, left int) string {
    // Implementation
}

func Margin(content string, top, right, bottom, left int) string {
    // Implementation
}

// Common layouts
func TwoColumn(left, right string, theme themes.Theme) string {
    // Implementation
}

func ThreeColumn(left, middle, right string, theme themes.Theme) string {
    // Implementation
}
```

**Apply to existing components:**
- Use `Box()` for help overlay borders
- Use `Box()` for error visualization borders
- Use consistent spacing throughout

**Tests:**
- Test each border style renders correctly
- Test spacing utilities
- Test column layouts
- Test theme integration

**Success Criteria:**
✅ Consistent borders across all components
✅ Spacing utilities work correctly
✅ All themes supported
✅ Existing components updated to use layout package

---

### Task 3: Intro Screen

**Purpose:** Better first-run experience with ASCII logo and keyboard shortcuts.

**Reference commits from hex:**
- `83bfe451` - Add startup intro screen
- `175d6530` - Persist intro until first message

**Files to create/modify:**
- `internal/ui/intro.go` - Intro screen rendering
- `internal/ui/model.go` - Add intro state tracking
- `internal/ui/view.go` - Render intro when appropriate
- `internal/ui/update.go` - Hide intro on first message

**Implementation:**

1. Create `internal/ui/intro.go`:
   ```go
   package ui

   // RenderIntro returns the intro screen content
   func (m *Model) RenderIntro() string {
       // ASCII art logo
       logo := `
   ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
   ┃  ____                          ┃
   ┃ |  _ \ __ _  __ _  ___ _ __   ┃
   ┃ | |_) / _' |/ _' |/ _ \ '_ \  ┃
   ┃ |  __/ (_| | (_| |  __/ | | | ┃
   ┃ |_|   \__,_|\__, |\___|_| |_| ┃
   ┃             |___/              ┃
   ┃                                ┃
   ┃  Productivity AI Agent         ┃
   ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
   `

       // Keyboard shortcuts
       shortcuts := `
   Keyboard Shortcuts:
   • ctrl+c    - Quit
   • ctrl+s    - Save conversation
   • ctrl+f    - Toggle favorites
   • ?         - Toggle help
   • /         - Quick actions

   Type a message to begin...
   `

       return m.theme.Title.Render(logo) + "\n\n" + shortcuts
   }
   ```

2. Add state to `model.go`:
   ```go
   type Model struct {
       // ... existing fields ...
       showIntro bool  // NEW: Track if intro should be shown
   }

   func NewModel(...) *Model {
       return &Model{
           // ... existing init ...
           showIntro: true,  // Show intro initially
       }
   }
   ```

3. Update `view.go`:
   ```go
   func (m *Model) View() string {
       // ... existing code ...

       // Show intro if enabled and no messages yet
       if m.showIntro && len(m.Messages) == 0 {
           return m.RenderIntro()
       }

       // ... rest of view ...
   }
   ```

4. Update `update.go` to hide intro on first user message:
   ```go
   // When user sends first message
   if keyMsg.Type == tea.KeyEnter && m.Input.Value() != "" {
       m.showIntro = false  // Hide intro
       // ... rest of send logic ...
   }
   ```

**Tests:**
- Test intro renders on startup
- Test intro hides after first message
- Test intro doesn't show when resuming conversation
- Test intro respects theme

**Success Criteria:**
✅ Intro shows on first launch
✅ Intro hides after first message
✅ Intro doesn't show when resuming
✅ ASCII art renders correctly
✅ Shortcuts are helpful and accurate

---

## Implementation Order

1. **WindowSizeMsg Fixes** (30 min)
   - Immediate bug fix
   - Improves tmux usability
   - No new files

2. **Layout/Borders System** (1-2 hours)
   - Foundation for consistent styling
   - Can be applied incrementally
   - New package creation

3. **Intro Screen** (1 hour)
   - Quick win, good UX improvement
   - Standalone feature
   - Welcomes new users

## Testing Strategy

- Unit tests for each component
- Integration tests for message forwarding
- Visual tests for borders and intro
- Manual testing in tmux/screen

## Success Criteria

### WindowSizeMsg Fixes:
✅ Forms work in tmux without freezing
✅ Forms resize properly
✅ All message types forwarded

### Layout/Borders:
✅ Consistent styling across components
✅ All border styles work
✅ Theme integration complete

### Intro Screen:
✅ Welcoming first experience
✅ Helpful keyboard shortcuts
✅ Hides at right time

## Rollback Plan

Each feature is independent:
- WindowSizeMsg: Revert message forwarding changes
- Layout: Optional package, not required
- Intro: Can be disabled with flag

## Timeline

- Task 1: 30 minutes
- Task 2: 1-2 hours
- Task 3: 1 hour

**Total: 2.5-3.5 hours** (or faster with subagents)

---

## Notes

Phase 2 focuses on **quality of life** improvements:
- Better tmux compatibility
- Professional visual consistency
- Welcoming first-run experience

These are polish features that make jeff-agent feel production-ready.
