# Overlay System Refactoring Design

**Date:** 2025-12-11
**Status:** Design Complete, Ready for Implementation

## Problem Statement

Current overlay system has two separate implementations:
- Bottom overlays (tool approval, autocomplete) use OverlayManager
- Fullscreen overlay (tool log) uses boolean flag with no scrolling
- Inconsistent input handling and no clear pattern for new overlays

## Architecture Overview

### Unified Stack-Based System

Single **OverlayManager** with stack-based model. Any overlay can be pushed onto the stack. Top of stack receives input first and acts as a true modal.

**Two overlay presentations:**
- **Bottom overlays**: Render between viewport and input, push viewport up
- **Fullscreen overlays**: Take over entire view with scrollable viewport

**Stack behavior:**
- Push/pop any overlay type: `m.overlayManager.Push(overlay)` / `Pop()`
- Model owns overlay instances, manager tracks stack
- Input capture prevents organic stacking of fullscreen modals (intentional)

**Designed for:**
- Bottom: Tool approval, autocomplete (existing)
- Fullscreen: Tool log (Ctrl+O), Help (Ctrl+H), History (Ctrl+R)

## Interface Design

### Base Overlay Interface

```go
type Overlay interface {
    // Structured rendering
    GetHeader() string
    GetContent() string
    GetFooter() string
    Render(width, height int) string

    // Input handling
    HandleKey(msg tea.KeyMsg) (handled bool, cmd tea.Cmd)

    // Lifecycle
    OnPush(width, height int)
    OnPop()

    // Height management
    GetDesiredHeight() int
}
```

### Scrollable Extension

```go
type Scrollable interface {
    Overlay
    Update(msg tea.Msg) tea.Cmd  // Receives viewport updates
}
```

Any overlay can implement Scrollable (bottom overlay with many items, all fullscreen overlays).

### Fullscreen Extension

```go
type FullscreenOverlay interface {
    Scrollable
    SetHeight(height int)
    IsFullscreen() bool  // Returns true
}
```

**Design rationale:**
- Composable interfaces allow flexible combinations
- Bottom overlays can be scrollable without being fullscreen
- Clear capability hierarchy
- Manager uses `IsFullscreen()` to determine rendering mode

## OverlayManager & Stack

```go
type OverlayManager struct {
    stack []Overlay
}

// Stack operations
func (om *OverlayManager) Push(overlay Overlay)
func (om *OverlayManager) Pop() Overlay
func (om *OverlayManager) Peek() Overlay
func (om *OverlayManager) Clear()

// Input routing
func (om *OverlayManager) Update(msg tea.Msg) tea.Cmd
func (om *OverlayManager) HandleKey(msg tea.KeyMsg) (handled bool, cmd tea.Cmd)

// Rendering
func (om *OverlayManager) Render(width, height int) string
func (om *OverlayManager) HasActive() bool
```

### Input Flow (Modal Behavior)

**All overlays are true modals** - capture all input:

1. If stack has overlay: `handled, cmd := m.overlayManager.HandleKey(msg)`
2. If overlay exists, handled=true ALWAYS (full capture)
3. Input never reaches normal handlers while overlay active
4. Overlay handles specific keys (Escape/Ctrl+C to dismiss, navigation, etc.)
5. All other keys captured and ignored

**No pass-through** - user must dismiss overlay to interact with main UI.

## Layout & Rendering

### Bottom Overlay Layout

```
┌─────────────────────┐
│ Header              │
├─────────────────────┤
│ Chat Viewport       │
│ (reduced height)    │
├─────────────────────┤
│ Bottom Overlay      │  ← Takes space from viewport
├─────────────────────┤
│ Input               │
├─────────────────────┤
│ Status              │
└─────────────────────┘
```

**Height management:**
```go
desiredHeight := overlay.GetDesiredHeight()
maxAllowed := m.Height * 0.4  // Cap at 40% of screen
overlayHeight := min(desiredHeight, maxAllowed)
viewportHeight = totalHeight - headerHeight - overlayHeight - inputHeight - statusHeight
```

Bottom overlays can become scrollable if content exceeds max height.

### Fullscreen Overlay Layout

```
┌─────────────────────┐
│ Fullscreen Overlay  │
│                     │
│ [scrollable]        │
│                     │
│                     │
└─────────────────────┘
```

Replaces entire view. Each fullscreen overlay embeds its own `viewport.Model`.

## Data Flow & Lifecycle

### Overlay Ownership

Model owns all overlay instances, created at initialization:

```go
type Model struct {
    overlayManager *OverlayManager

    // Bottom overlays
    toolApprovalOverlay *ToolApprovalOverlay
    autocompleteOverlay *AutocompleteOverlay

    // Fullscreen overlays
    toolLogOverlay   *ToolLogOverlay
    helpOverlay      *HelpOverlay
    historyOverlay   *HistoryOverlay

    // Data that overlays reference
    toolLogLines []string
    messages     []Message
}
```

### Data Updates

**Overlays reference Model data directly:**

```go
// At initialization
m.toolLogOverlay = NewToolLogOverlay(&m.toolLogLines)
m.historyOverlay = NewHistoryOverlay(&m.messages)
```

**Two update paths:**
1. **User input** → `m.overlayManager.HandleKey(msg)` → overlay handles internally
2. **Model data changes** → Model updates data → overlay reads on next render

No explicit sync needed - overlays are views over Model's data.

### Lifecycle Flow

```go
// Opening overlay
case tea.KeyCtrlO:
    m.overlayManager.Push(m.toolLogOverlay)
    // OnPush(width, height) called automatically

// Closing overlay (from inside overlay's HandleKey)
case tea.KeyEsc:
    m.overlayManager.Pop()
    // OnPop() called automatically

// Data updates while open - automatic
m.toolLogLines = append(m.toolLogLines, "new line")
// Next render shows new content
```

## Specific Overlay Implementations

### Bottom Overlays

**ToolApprovalOverlay** (refactor existing)
- Height: 5 lines (compact form)
- Keys: Enter (approve), Escape (deny), arrows (focus)
- Shows: tool name, param preview, Allow/Deny buttons

**AutocompleteOverlay** (refactor existing)
- Height: dynamic, max 40% screen (scrollable if exceeded)
- Keys: Enter (select), Escape (dismiss), arrows (navigate)
- Shows: filtered slash commands or suggestions

### Fullscreen Overlays

**ToolLogOverlay** (refactor from toollog.go)
- Content: Last 10,000 lines of tool output
- Header: "Tool Output Log"
- Footer: "Line 45/1250 • Esc to close"
- Keys: Escape/Ctrl+C (close), arrows/PageUp/PageDown (scroll), mouse wheel
- Hotkey: Ctrl+O

**HelpOverlay** (new)
- Content: Keyboard shortcuts, features, tips
- Header: "Help & Keyboard Shortcuts"
- Footer: "Esc to close"
- Keys: Same as tool log
- Hotkey: Ctrl+H

**HistoryOverlay** (new)
- Content: Last 1,000 messages with search
- Header: "Conversation History • / to search"
- Footer: "123 messages • Esc to close"
- Keys: Same as tool log, plus '/' for search
- Hotkey: Ctrl+R

**All fullscreen overlays:**
- Embed `viewport.Model` instance
- Handle viewport navigation
- Auto-scroll to bottom on open (configurable per overlay)

## Implementation Plan

### File Structure

```
internal/ui/
  overlay.go                    # Interfaces
  overlay_manager.go            # OverlayManager with stack
  overlay_tool_approval.go      # Refactored (bottom)
  overlay_autocomplete.go       # Refactored (bottom)
  overlay_tool_log.go           # Refactored (fullscreen)
  overlay_help.go               # New (fullscreen)
  overlay_history.go            # New (fullscreen)
  overlay_test.go               # Updated tests
```

### Migration Strategy

**Phase 1: Core refactor**
- Update interfaces (Overlay, Scrollable, FullscreenOverlay)
- Refactor OverlayManager to use stack
- Update existing bottom overlays (tool approval, autocomplete)
- Remove old implementation cruft

**Phase 2: Tool log conversion**
- Convert toollog.go to overlay_tool_log.go
- Remove `m.toolLogOverlay` boolean flag
- Implement fullscreen rendering with viewport
- Test Ctrl+O functionality

**Phase 3: New overlays**
- Implement HelpOverlay (Ctrl+H)
- Implement HistoryOverlay (Ctrl+R)
- Add comprehensive tests for all three fullscreen overlays

### Key Model Changes

```go
// Remove
m.toolLogOverlay bool  // Boolean flag

// Add
m.toolLogOverlay *ToolLogOverlay     // Instance
m.helpOverlay *HelpOverlay           // Instance
m.historyOverlay *HistoryOverlay     // Instance
```

### Testing Focus

- Stack push/pop operations
- Input capture (true modal behavior)
- Height calculations for bottom overlays
- Viewport scrolling in fullscreen overlays
- Overlay resize handling (tea.WindowSizeMsg)
- Data reference updates (verify overlays see Model changes)

## Benefits

1. **Unified system** - one pattern for all overlays
2. **Scrollable fullscreen** - proper viewport navigation
3. **Extensible** - clear pattern for adding new overlays
4. **True modals** - consistent input capture
5. **Flexible** - composable interfaces allow mixing capabilities
6. **Simple data flow** - overlays reference Model data directly

## Open Questions

None - design complete and validated.
