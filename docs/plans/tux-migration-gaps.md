# Tux Migration Gaps

Tracking gaps discovered during the tux migration that need to be addressed in the tux library.

## Chat Scrolling (Phase 3)

**Status:** Open
**Severity:** Medium - Affects user navigation experience
**Reported:** 2026-01-12

### Description

The tux library's `ChatContent` does not support scrolling. Users cannot:
- Scroll through chat history with `j`/`k` (vim-style)
- Jump to top/bottom with `gg`/`G`
- Page up/down with `Ctrl+u`/`Ctrl+d`
- Use arrow keys to scroll

### Root Cause Analysis

1. **ChatContent lacks viewport**: `tux.ChatContent` simply joins messages as strings and returns them from `View()`. There is no `viewport.Model` wrapping the content.

2. **No focus toggle mechanism**: The shell routes keyboard events to the active tab's content only when `s.focused == FocusTab`, but:
   - Focus defaults to `FocusInput` (the text input area)
   - There's no keyboard shortcut to toggle between `FocusInput` and `FocusTab`
   - Even if focus could be switched, `ChatContent.Update()` does nothing with key events

3. **Content.Viewport exists but unused**: Tux has a `content.Viewport` primitive that wraps bubbles/viewport with full scrolling support, but `ChatContent` doesn't use it.

### What Exists in Tux

- `content.Viewport` - Full scrolling support with:
  - `ScrollToTop()`, `ScrollToBottom()`
  - `ScrollUp(n)`, `ScrollDown(n)`
  - Proper keyboard handling via bubbles/viewport default keybindings (j/k, ctrl+u/ctrl+d, pgup/pgdn)

- `shell.FocusTarget` - Focus states exist (`FocusInput`, `FocusTab`, `FocusModal`)

- `shell.Shell.Focus()` - Method to set focus target

### Potential Solutions (for tux maintainer)

**Option A: ChatContent uses Viewport internally**
- ChatContent embeds `content.Viewport`
- Messages are rendered and set as viewport content
- ChatContent.Update() delegates to viewport for key handling
- Auto-scrolls to bottom on new messages

**Option B: Add focus toggle to Shell**
- Add `Esc` or other key to toggle between FocusInput and FocusTab
- When FocusTab is active, keys go to tab content
- Content implementations handle their own scrolling

**Option C: Global scroll keys**
- Shell handles scroll keys globally (when no modal active)
- Routes them to active tab content's viewport
- Requires content interface to expose scroll methods

### Impact on Hex Migration

Without chat scrolling, users cannot:
- Review long conversations
- Navigate to previous tool calls/results
- Jump to specific parts of the chat

This is functionality that exists in the current hex TUI that would be lost.

### Workaround (if needed)

Hex could create a custom ChatContent replacement that uses content.Viewport internally, but this would require:
- Duplicating ChatContent's message handling logic
- Custom wiring to replace tux's internal chat content
- Maintenance burden for hex-specific UI code

Per the migration design, gaps should be fixed in tux, not shimmed in hex.

---

## [Template for future gaps]

**Status:** [Open|Fixed|Won't Fix]
**Severity:** [High|Medium|Low]
**Reported:** [Date]

### Description
[What the gap is]

### Root Cause Analysis
[Why the gap exists]

### Potential Solutions
[Ideas for fixing]

### Impact on Hex Migration
[What functionality is affected]
