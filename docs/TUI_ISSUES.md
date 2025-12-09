# Hex TUI Issues

Running list of UI issues discovered during testing. Do not fix yet - collecting feedback first.

---

## Visual / Color Issues

### 1. `/help` context menu colors not legible
- **Description**: When typing `/help`, the autocomplete/context menu appears with light blue text on medium gray background
- **Impact**: Text is difficult to read
- **Expected**: High contrast colors for readability

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

### 4. No status line support
- **Description**: No visibility into current status, would benefit from ccstatusline support
- **Impact**: User doesn't know what's happening (streaming, waiting, etc.)
- **Expected**: Status line showing current state, model, tokens, etc.

### 5. No visibility into which model is being used
- **Description**: UI doesn't display which model (claude-sonnet-4, opus, etc.) is active
- **Impact**: User doesn't know what they're talking to
- **Expected**: Model name visible somewhere in UI (status bar, header, etc.)

### 6. Not showing current working directory (CWD)
- **Description**: UI doesn't display the current working directory
- **Impact**: User doesn't know what directory context Hex is operating in
- **Expected**: CWD visible in status bar or header

---

## Tool Execution Issues

### 7. Tool usage shows no details
- **Description**: When using a tool, only shows `Using tool: bash` with no information about WHAT command is being run, status, progress, etc.
- **Impact**: User is blind to what Hex is actually doing - no visibility into commands, no way to verify correctness
- **Expected**: Show the actual command/parameters being executed, execution status, output preview

### 12. ~~Tool use streaming appears broken - input_json_delta empty~~ FIXED
- **Status**: FIXED in this session
- **Root Cause**: When first `input_json_delta` chunk had empty `PartialJSON`, the handler returned `nil` instead of `continueReading()`, which stopped stream processing. Subsequent chunks with actual JSON never reached the UI.
- **Fix**: Changed `handleContentBlockDelta` to always call `continueReading()` for `input_json_delta` chunks, regardless of whether `PartialJSON` is empty.
- **File**: `internal/ui/update.go` lines 810-817

---

## Command Issues

### 8. `/help` slash command not implemented?
- **Description**: `/help` doesn't appear to be a real command
- **Additional error**: `Warning: failed to parse command commands/README.md: command missing required 'name' field`
- **Impact**: Basic help functionality missing, spurious warnings
- **Expected**: `/help` should show available commands and usage

---

## Safety / UX Issues

### 9. Ctrl+C instantly quits without confirmation
- **Description**: Pressing Ctrl+C immediately exits the application
- **Impact**: Easy to accidentally lose conversation/context
- **Expected**: First Ctrl+C should warn "Press Ctrl+C again to quit", second Ctrl+C actually quits

### 10. No input blocking/staging during tool execution
- **Description**: While Hex is working/using a tool, user can keep sending messages that have no chance of being processed
- **Impact**: Confusing UX, messages get lost or queued unexpectedly
- **Expected**: Input should be blocked or staged while processing, or clearly indicate messages are queued

---

## Input / Navigation Issues

### 11. Up arrow doesn't iterate through message history
- **Description**: Pressing up arrow in the input box doesn't cycle through previously sent messages
- **Impact**: No quick way to re-send or edit previous messages, common shell/REPL convention missing
- **Expected**: Up/down arrows should navigate through user's message history (like bash, zsh, Python REPL, etc.)

---

## Tool Approval UI Issues

### 13. Up/down arrows scroll options instead of moving highlight
- **Description**: In tool approval dialog, pressing up/down arrows changes the visibility of options (scrolls) instead of moving the selection highlight between options
- **Impact**: Confusing navigation, can't reliably select options with keyboard
- **Expected**: Up/down should move highlight between options (Approve/Deny/Always/Never)

### 14. Tool approval UI is too bulky
- **Description**: Current layout spreads information across too many lines:
  ```
  Tool: bash
  ID: toolu_01...
  Risk Level: Caution ⚠
  Parameters:
    • command: "..."
  ```
- **Impact**: Takes up too much screen space, ID is not user-relevant
- **Expected**: Compact layout with tool+params together, risk level in header:
  ```
  ⚠ bash: command="echo hello"
  [Approve] [Deny] [Always] [Never]
  ```

### 15. Tool result display issues
- **Description**: After tool execution, the result is displayed with several problems:
  1. Shows raw format including "STDOUT:" prefix instead of parsed/clean output
  2. Result is repeated as coming from "YOU" which is incorrect (tool results are not user messages)
- **Impact**: Confusing UX, cluttered display, incorrect attribution
- **Expected**: Clean tool output without raw prefixes, attributed correctly (not as "YOU")

### 16. Tool execution needs "background threading" concept
- **Description**: Tool results are displayed inline in the conversation, cluttering the view. Need a concept where:
  1. Tool execution happens "in background"
  2. Results are collapsed/hidden by default
  3. User can expand to see details if needed
  4. Or results only shown on request
- **Impact**: Conversation gets cluttered with verbose tool output, hard to follow the actual conversation flow
- **Expected**: Clean conversation view with tool activity minimized/collapsed, expandable on demand

### 17. ~~Enter key doesn't select option in tool approval~~ FIXED
- **Status**: FIXED in this session
- **Root Cause**: When embedding huh form in Bubble Tea, internal huh commands (NextField, etc.) weren't being routed back to the form. The form returned these commands, but only KeyMsg was forwarded to the form - other messages went to the parent model which didn't know how to handle them.
- **Fix**: Forward ALL messages to the embedded form when in approval mode, not just KeyMsg. This allows huh's internal state machine to properly transition from StateNormal to StateCompleted.
- **Files**: `internal/ui/update.go` (added message forwarding), `internal/ui/forms/approval.go` (cleanup)

---

## Issue Count: 17 (2 fixed, 15 open)

Last updated: 2025-12-09
