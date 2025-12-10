# Hex TUI Issues

Running list of UI issues discovered during testing. Do not fix yet - collecting feedback first.

---

## Visual / Color Issues

### 1. ~~`/help` context menu colors not legible~~ FIXED
- **Status**: FIXED in this session
- **Fix**: Improved autocomplete dropdown colors with dark background (DeepInk) and bright foreground (SoftPaper) for maximum contrast. Updated both NeoTerminal and Dracula themes.

---

## Layout / Behavior Issues

### 2. Context menu moves input box
- **Description**: When the `/help` context menu appears, it moves the input box up, causing the interface to juggle/shift
- **Impact**: Disorienting UX, layout instability
- **Expected**: Overlay should not displace other elements, or transition should be smooth

### 3. Previous content scrolls up when typing
- **Description**: When typing in the input box, previous conversation content scrolls up unexpectedly
- **Impact**: Confusing, loses context of what you're replying to
- **Expected**: Viewport should remain stable while typing unless explicitly scrolled

---

## Missing Features

### 4. ~~No status line support~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: Added status bar showing model name, shortened CWD, and dynamic key hints

### 5. ~~No visibility into which model is being used~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: Model name now displayed in status bar

### 6. ~~Not showing current working directory (CWD)~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: CWD now displayed in status bar (shortened format)

---

## Tool Execution Issues

### 7. ~~Tool usage shows no details~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: Tool output log with collapsed view (last 3 lines) and Ctrl+O overlay for full output. Compact tool calls show as `🛠 bash("command")` format.

### 12. ~~Tool use streaming appears broken - input_json_delta empty~~ FIXED
- **Status**: FIXED in this session
- **Root Cause**: When first `input_json_delta` chunk had empty `PartialJSON`, the handler returned `nil` instead of `continueReading()`, which stopped stream processing. Subsequent chunks with actual JSON never reached the UI.
- **Fix**: Changed `handleContentBlockDelta` to always call `continueReading()` for `input_json_delta` chunks, regardless of whether `PartialJSON` is empty.
- **File**: `internal/ui/update.go` lines 810-817

---

## Command Issues

### 8. ~~`/help` slash command not implemented~~ FIXED
- **Status**: FIXED in this session
- **Fix**: Created `/help` command (`commands/help.md`) showing available slash commands and keyboard shortcuts. Also refactored autocomplete to show slash commands (like `/plan`, `/help`) instead of internal tools (like `read_file`, `bash`). Autocomplete now triggers on `/` prefix.

---

## Safety / UX Issues

### 9. ~~Ctrl+C instantly quits without confirmation~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: Two-stage Ctrl+C quit confirmation - first press shows warning, second press quits

### 10. ~~No input blocking/staging during tool execution~~ FIXED
- **Status**: FIXED in this session
- **Root Cause**: Messages sent during tool execution were added to `m.Messages` immediately, appearing in API requests before tool_result, which violated Anthropic API protocol (tool_use must be immediately followed by tool_result).
- **Fix**: Messages typed during busy states (streaming, tool execution, tool approval) are now queued without being added to conversation history. Status bar shows "Message queued (N pending)". Queued messages are processed after the current operation completes and only then added to conversation history.
- **File**: `internal/ui/update.go` (input queuing logic and `processNextQueuedMessage`)

---

## Input / Navigation Issues

### 11. ~~Up arrow doesn't iterate through message history~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: Input history navigation with Up/Down keys now works like shell/REPL

---

## Tool Approval UI Issues

### 13. ~~Up/down arrows scroll options instead of moving highlight~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: Replaced huh form with custom approval menu that shows all 4 options at once with arrow key navigation

### 14. ~~Tool approval UI is too bulky~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: Compact approval UI now shows `⚠ bash("command")` format with all options visible

### 15. ~~Tool result display issues~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: Tool results now display cleanly with collapsed log view (`│ ─── bash("cmd") ───`) and are attributed to ASSISTANT. No more STDOUT prefix or incorrect YOU attribution.

### 16. ~~Tool execution needs "background threading" concept~~ FIXED
- **Status**: FIXED in PR #3
- **Fix**: Tool output log with collapsed view (last 3 lines) and Ctrl+O overlay. Internal tool messages filtered from main chat view.

### 17. ~~Enter key doesn't select option in tool approval~~ FIXED
- **Status**: FIXED in this session
- **Root Cause**: When embedding huh form in Bubble Tea, internal huh commands (NextField, etc.) weren't being routed back to the form. The form returned these commands, but only KeyMsg was forwarded to the form - other messages went to the parent model which didn't know how to handle them.
- **Fix**: Forward ALL messages to the embedded form when in approval mode, not just KeyMsg. This allows huh's internal state machine to properly transition from StateNormal to StateCompleted.
- **Files**: `internal/ui/update.go` (added message forwarding), `internal/ui/forms/approval.go` (cleanup)

---

## Issue Count: 17 (16 fixed, 1 open)

**Remaining open issues:**
- #2: Context menu moves input box
- #3: Previous content scrolls up when typing

Last updated: 2025-12-10
