# Tool Log UI Design

**Date:** 2025-12-09
**Status:** Ready for implementation

## Overview

Tool output should be visually distinct and unobtrusive - like a live log tail. Users see the last 3 lines of output inline, and can expand to a full overlay view with `Ctrl+O`.

## Design

### Collapsed View (Inline)

**Location:** Inline in conversation, where tool output would normally appear.

**Visual treatment:**
- Dimmed/muted colors (less prominent than conversation text)
- Each line prefixed with `│` (log/quote aesthetic)
- Shows last 3 lines of raw output (`tail -n 3` style)

**Scope:** A "chunk" = all tool activity since the last user or assistant text message. Resets when conversation continues.

**Example:**
```
ASSISTANT [14:32]
    Let me check the build status.

│ go build -o bin/hexreplay ./cmd/hexreplay
│ ✅ Built bin/hexviz and bin/hexreplay
│ ✅ Built bin/hex

ASSISTANT [14:32]
    Build succeeded! All binaries are ready.
```

### Expanded View (Overlay)

**Trigger:** `Ctrl+O` toggles overlay. `Esc` also closes it.

**Behavior:**
- Modal overlay covering conversation area
- Shows full log of current tool chunk
- Scrollable with vi-style navigation (`j`/`k`, `gg`/`G`)
- Full output, not truncated

**Layout:**
```
┏━━ Tool Output Log ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ Ctrl+O or Esc to close ┓

─── grep("TODO") ───
src/main.go:42: // TODO: handle error
src/utils.go:18: // TODO: add tests

─── bash("make build") ───
Building visualization tools...
go build -o bin/hexviz ./cmd/hexviz
go build -o bin/hexreplay ./cmd/hexreplay
✅ Built bin/hexviz and bin/hexreplay
Building hex...
go build -ldflags "..." -o bin/hex ./cmd/hex
✅ Built bin/hex

┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Edge Cases

**No tool output:**
- Collapsed: Nothing shown (no empty box)
- `Ctrl+O`: Show "No tool output in current chunk"

**Tool running:**
- Collapsed view updates live as output streams
- Last 3 lines shift as new output arrives

**Tool failed:**
- Error output captured in log like any other output
- No special treatment needed

**Chunk transitions:**
- Assistant text response closes current chunk
- Next tool use starts new chunk
- Only current chunk displayed (data preserved for future features)

## Implementation Notes

### Data Model

Add to `Model` struct:
- `toolLogLines []string` - accumulated output lines for current chunk
- `toolLogOverlay bool` - whether overlay is visible

### Key Binding

- `Ctrl+O` - toggle tool log overlay (add to update.go key handling)

### Rendering

1. **Collapsed:** New function to render last 3 lines with `│` prefix and muted style
2. **Overlay:** New view mode or overlay component with bordered box and full content

### Chunk Management

- Clear `toolLogLines` when assistant sends text response
- Append to `toolLogLines` as tool output streams in
- Parse tool results to extract output lines

## Future Possibilities (Not in Scope)

- Scroll through previous chunks in overlay
- Search across all tool output
- Copy tool output to clipboard
- Persist tool logs to file
