# Tux Migration Gaps

Tracking gaps discovered during the tux migration that need to be addressed in the tux library.

## Chat Scrolling (Phase 3)

**Status:** Fixed
**Severity:** Medium - Affects user navigation experience
**Reported:** 2026-01-12
**Fixed:** 2026-01-12

### Description

The tux library's `ChatContent` did not support scrolling.

### Solution Implemented

Added viewport support to `ChatContent` in tux:

1. **Embedded `viewport.Model`** in ChatContent for scrolling
2. **Keyboard navigation** in `Update()`:
   - `j`/`k` or arrows for line scrolling
   - `g` for top, `G` for bottom
   - `Ctrl+u`/`Ctrl+d` for half-page scrolling
   - `PgUp`/`PgDn` for full page scrolling
3. **Auto-scroll behavior**:
   - Auto-scrolls to bottom on new messages by default
   - Disables auto-scroll when user manually scrolls up
   - Re-enables auto-scroll when user scrolls to bottom or presses `G`
4. **Focus toggle** (added earlier):
   - `Esc` toggles focus between input and tab content
   - When tab content has focus, keys route to ChatContent for scrolling

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
