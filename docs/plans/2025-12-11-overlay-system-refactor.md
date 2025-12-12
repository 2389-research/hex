# Overlay System Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Unify bottom and fullscreen overlay systems with stack-based management and scrollable fullscreen viewports.

**Architecture:** Replace dual systems (OverlayManager + boolean flag) with single stack-based OverlayManager. All overlays implement composable interfaces (Overlay, Scrollable, FullscreenOverlay). Bottom overlays push viewport up, fullscreen overlays take entire view with embedded viewport. Model owns overlay instances, manager tracks stack.

**Tech Stack:** Go, Bubble Tea, Charm Bracelet (viewport, lipgloss), existing hex TUI architecture

**Reference Design:** `docs/plans/2025-12-11-overlay-system-refactor-design.md`

---

## Phase 1: Core Interface & Manager Refactor

### Task 1: Update Base Overlay Interface

**Files:**
- Modify: `internal/ui/overlay.go:6-42`

**Step 1: Update Overlay interface with structured rendering**

Replace existing interface with:

```go
// Overlay represents a modal interface that appears over the main content
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

**Step 2: Add Scrollable interface**

Add below Overlay interface:

```go
// Scrollable adds viewport scrolling capability to any overlay
type Scrollable interface {
	Overlay
	Update(msg tea.Msg) tea.Cmd
}
```

**Step 3: Add FullscreenOverlay interface**

Add below Scrollable interface:

```go
// FullscreenOverlay represents a fullscreen modal with viewport
type FullscreenOverlay interface {
	Scrollable
	SetHeight(height int)
	IsFullscreen() bool
}
```

**Step 4: Remove OverlayType enum**

Delete lines 6-15 (OverlayType enum and constants) - no longer needed with interface hierarchy.

**Step 5: Commit interface updates**

```bash
git add internal/ui/overlay.go
git commit -m "refactor: update overlay interfaces for stack-based system

- Add GetHeader/GetContent/GetFooter for structured rendering
- Add GetDesiredHeight for bottom overlay height management
- Add Scrollable interface for viewport-enabled overlays
- Add FullscreenOverlay interface for fullscreen modals
- Remove OverlayType enum (replaced by IsFullscreen() check)"
```

---

### Task 2: Refactor OverlayManager to Stack

**Files:**
- Modify: `internal/ui/overlay.go:44-127`

**Step 1: Update OverlayManager struct**

Replace OverlayManager struct (line 45-47):

```go
// OverlayManager manages a stack of overlays
type OverlayManager struct {
	stack []Overlay
}
```

**Step 2: Update NewOverlayManager**

Replace function (line 50-54):

```go
// NewOverlayManager creates a new overlay manager
func NewOverlayManager() *OverlayManager {
	return &OverlayManager{
		stack: make([]Overlay, 0, 4), // Pre-allocate for common case
	}
}
```

**Step 3: Remove Register, add Push/Pop/Peek**

Delete Register function (lines 56-59), replace with:

```go
// Push adds an overlay to the top of the stack
func (om *OverlayManager) Push(overlay Overlay) {
	om.stack = append(om.stack, overlay)
	// OnPush will be called by Model with width/height
}

// Pop removes and returns the top overlay from the stack
func (om *OverlayManager) Pop() Overlay {
	if len(om.stack) == 0 {
		return nil
	}
	overlay := om.stack[len(om.stack)-1]
	om.stack = om.stack[:len(om.stack)-1]
	overlay.OnPop()
	return overlay
}

// Peek returns the top overlay without removing it
func (om *OverlayManager) Peek() Overlay {
	if len(om.stack) == 0 {
		return nil
	}
	return om.stack[len(om.stack)-1]
}

// Clear removes all overlays from the stack
func (om *OverlayManager) Clear() {
	for len(om.stack) > 0 {
		om.Pop()
	}
}
```

**Step 4: Update GetActive to use Peek**

Replace GetActive function (lines 62-74):

```go
// GetActive returns the top overlay on the stack, or nil if empty
func (om *OverlayManager) GetActive() Overlay {
	return om.Peek()
}
```

**Step 5: Update HandleKey for modal capture**

Replace HandleKey function (lines 82-89):

```go
// HandleKey passes a key event to the active overlay
// Returns (true, cmd) if key was handled, (false, nil) if no active overlay
func (om *OverlayManager) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	active := om.GetActive()
	if active == nil {
		return false, nil
	}
	// Modal behavior: overlay always captures input
	return active.HandleKey(msg)
}
```

**Step 6: Simplify HandleEscape and HandleCtrlC**

Replace both functions (lines 91-109) with:

```go
// HandleEscape is deprecated - use HandleKey instead
// Kept for backward compatibility during migration
func (om *OverlayManager) HandleEscape() tea.Cmd {
	active := om.GetActive()
	if active != nil {
		handled, cmd := active.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
		if handled {
			return cmd
		}
	}
	return nil
}

// HandleCtrlC is deprecated - use HandleKey instead
// Kept for backward compatibility during migration
func (om *OverlayManager) HandleCtrlC() tea.Cmd {
	active := om.GetActive()
	if active != nil {
		handled, cmd := active.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
		if handled {
			return cmd
		}
	}
	return nil
}
```

**Step 7: Update Render to check IsFullscreen**

Replace Render function (lines 111-118):

```go
// Render returns the content of the top overlay
// Returns empty string if no active overlay
func (om *OverlayManager) Render(width, height int) string {
	active := om.GetActive()
	if active == nil {
		return ""
	}
	return active.Render(width, height)
}

// IsFullscreen returns true if the active overlay is fullscreen
func (om *OverlayManager) IsFullscreen() bool {
	active := om.GetActive()
	if active == nil {
		return false
	}
	if fs, ok := active.(FullscreenOverlay); ok {
		return fs.IsFullscreen()
	}
	return false
}
```

**Step 8: Update CancelAll to use Pop**

Replace CancelAll function (lines 120-127):

```go
// CancelAll dismisses all active overlays
func (om *OverlayManager) CancelAll() {
	om.Clear()
}
```

**Step 9: Commit manager refactor**

```bash
git add internal/ui/overlay.go
git commit -m "refactor: convert OverlayManager to stack-based system

- Replace overlay list with stack
- Add Push/Pop/Peek/Clear operations
- Update HandleKey to return (handled bool, cmd)
- Add IsFullscreen() check for rendering mode
- Simplify CancelAll to use Clear
- Deprecate HandleEscape/HandleCtrlC (use HandleKey)"
```

---

### Task 3: Write Tests for Stack Operations

**Files:**
- Modify: `internal/ui/overlay_test.go`

**Step 1: Create mock overlay for testing**

Add at top of file after imports:

```go
// mockOverlay is a test overlay implementation
type mockOverlay struct {
	header        string
	content       string
	footer        string
	desiredHeight int
	onPushCalled  bool
	onPopCalled   bool
	lastWidth     int
	lastHeight    int
	keyHandler    func(tea.KeyMsg) (bool, tea.Cmd)
}

func newMockOverlay(header, content, footer string, height int) *mockOverlay {
	return &mockOverlay{
		header:        header,
		content:       content,
		footer:        footer,
		desiredHeight: height,
	}
}

func (m *mockOverlay) GetHeader() string              { return m.header }
func (m *mockOverlay) GetContent() string             { return m.content }
func (m *mockOverlay) GetFooter() string              { return m.footer }
func (m *mockOverlay) GetDesiredHeight() int          { return m.desiredHeight }
func (m *mockOverlay) Render(width, height int) string {
	return fmt.Sprintf("%s\n%s\n%s", m.header, m.content, m.footer)
}

func (m *mockOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if m.keyHandler != nil {
		return m.keyHandler(msg)
	}
	// Default: handle Escape to pop
	if msg.Type == tea.KeyEsc {
		return true, nil
	}
	return false, nil
}

func (m *mockOverlay) OnPush(width, height int) {
	m.onPushCalled = true
	m.lastWidth = width
	m.lastHeight = height
}

func (m *mockOverlay) OnPop() {
	m.onPopCalled = true
}
```

**Step 2: Write test for Push/Pop/Peek**

Add test:

```go
func TestOverlayManager_StackOperations(t *testing.T) {
	om := NewOverlayManager()

	// Empty stack
	assert.Nil(t, om.Peek())
	assert.Nil(t, om.GetActive())
	assert.False(t, om.HasActive())

	// Push first overlay
	overlay1 := newMockOverlay("Header 1", "Content 1", "Footer 1", 5)
	om.Push(overlay1)

	assert.Equal(t, overlay1, om.Peek())
	assert.Equal(t, overlay1, om.GetActive())
	assert.True(t, om.HasActive())

	// Push second overlay
	overlay2 := newMockOverlay("Header 2", "Content 2", "Footer 2", 10)
	om.Push(overlay2)

	assert.Equal(t, overlay2, om.Peek()) // Top of stack
	assert.Equal(t, overlay2, om.GetActive())

	// Pop should return top
	popped := om.Pop()
	assert.Equal(t, overlay2, popped)
	assert.True(t, overlay2.onPopCalled)
	assert.Equal(t, overlay1, om.Peek()) // Back to first

	// Pop last overlay
	om.Pop()
	assert.Nil(t, om.Peek())
	assert.False(t, om.HasActive())
}
```

**Step 3: Write test for Clear**

Add test:

```go
func TestOverlayManager_Clear(t *testing.T) {
	om := NewOverlayManager()

	overlay1 := newMockOverlay("H1", "C1", "F1", 5)
	overlay2 := newMockOverlay("H2", "C2", "F2", 10)
	overlay3 := newMockOverlay("H3", "C3", "F3", 15)

	om.Push(overlay1)
	om.Push(overlay2)
	om.Push(overlay3)

	assert.True(t, om.HasActive())

	om.Clear()

	assert.False(t, om.HasActive())
	assert.Nil(t, om.Peek())
	assert.True(t, overlay1.onPopCalled)
	assert.True(t, overlay2.onPopCalled)
	assert.True(t, overlay3.onPopCalled)
}
```

**Step 4: Write test for HandleKey modal capture**

Add test:

```go
func TestOverlayManager_HandleKeyModalCapture(t *testing.T) {
	om := NewOverlayManager()

	// No overlay - not handled
	handled, cmd := om.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, handled)
	assert.Nil(t, cmd)

	// Push overlay that handles Enter
	keyCaptured := false
	overlay := newMockOverlay("H", "C", "F", 5)
	overlay.keyHandler = func(msg tea.KeyMsg) (bool, tea.Cmd) {
		if msg.Type == tea.KeyEnter {
			keyCaptured = true
			return true, nil
		}
		return false, nil
	}
	om.Push(overlay)

	// Overlay captures Enter
	handled, cmd = om.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, handled)
	assert.True(t, keyCaptured)
	assert.Nil(t, cmd)

	// Overlay doesn't handle other keys but still captures (modal)
	handled, _ = om.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	assert.False(t, handled) // Overlay didn't handle it
}
```

**Step 5: Run tests to verify**

```bash
go test ./internal/ui -run TestOverlayManager -v
```

Expected: All tests pass

**Step 6: Commit tests**

```bash
git add internal/ui/overlay_test.go
git commit -m "test: add stack operation tests for OverlayManager

- Test Push/Pop/Peek with multiple overlays
- Test Clear removes all overlays and calls OnPop
- Test HandleKey modal capture behavior
- Add mockOverlay helper for testing"
```

---

### Task 4: Update ToolApprovalOverlay to New Interface

**Files:**
- Modify: `internal/ui/overlay_tool_approval.go`

**Step 1: Add structured rendering methods**

Add after existing methods:

```go
// GetHeader returns the overlay header
func (o *ToolApprovalOverlay) GetHeader() string {
	return "Tool Approval Required"
}

// GetContent returns the overlay content
func (o *ToolApprovalOverlay) GetContent() string {
	// Extract content from existing Render()
	// This is the tool name, params, and buttons section
	if o.toolName == "" {
		return "No tool pending approval"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Tool: %s\n", o.toolName))
	b.WriteString(fmt.Sprintf("Params: %s\n", o.paramPreview))
	return b.String()
}

// GetFooter returns the overlay footer
func (o *ToolApprovalOverlay) GetFooter() string {
	return "↑/↓: Focus • Enter: Approve • Esc: Deny"
}

// GetDesiredHeight returns the desired height
func (o *ToolApprovalOverlay) GetDesiredHeight() int {
	return 5 // Compact form
}
```

**Step 2: Add lifecycle methods**

Add:

```go
// OnPush is called when overlay is pushed to stack
func (o *ToolApprovalOverlay) OnPush(width, height int) {
	// No special initialization needed
}

// OnPop is called when overlay is popped from stack
func (o *ToolApprovalOverlay) OnPop() {
	// Clear state when dismissed
	o.toolName = ""
	o.paramPreview = ""
}
```

**Step 3: Update HandleKey signature**

Update HandleKey to return (bool, tea.Cmd):

```go
// HandleKey processes key input
func (o *ToolApprovalOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Deny and pop handled by caller
		return true, o.HandleEscape()
	case tea.KeyEnter:
		// Approve handled by specific logic
		// Return handled=true to capture input
		return true, nil
	case tea.KeyUp, tea.KeyDown:
		// Focus navigation
		return true, nil
	default:
		// Modal: capture all other keys
		return true, nil
	}
}
```

**Step 4: Update Render to use structured methods**

Refactor Render() to compose from Get methods:

```go
// Render returns the complete overlay rendering
func (o *ToolApprovalOverlay) Render(width, height int) string {
	var b strings.Builder

	// Header
	if header := o.GetHeader(); header != "" {
		b.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("cyan")).
			Render(header))
		b.WriteString("\n")
	}

	// Content
	b.WriteString(o.GetContent())
	b.WriteString("\n")

	// Footer
	if footer := o.GetFooter(); footer != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(footer))
	}

	return b.String()
}
```

**Step 5: Commit tool approval updates**

```bash
git add internal/ui/overlay_tool_approval.go
git commit -m "refactor: update ToolApprovalOverlay to new interface

- Add GetHeader/GetContent/GetFooter methods
- Add GetDesiredHeight (returns 5 lines)
- Add OnPush/OnPop lifecycle methods
- Update HandleKey to return (bool, cmd)
- Refactor Render to use structured methods"
```

---

### Task 5: Update AutocompleteOverlay to New Interface

**Files:**
- Modify: `internal/ui/overlay_autocomplete.go`

**Step 1-5: Apply same refactoring as ToolApprovalOverlay**

Follow identical pattern from Task 4:
- Add GetHeader/GetContent/GetFooter
- Add GetDesiredHeight (dynamic based on items, max 40% screen)
- Add OnPush/OnPop
- Update HandleKey signature
- Update Render

**Step 6: Commit autocomplete updates**

```bash
git add internal/ui/overlay_autocomplete.go
git commit -m "refactor: update AutocompleteOverlay to new interface

- Add structured rendering methods
- Add dynamic GetDesiredHeight with 40% cap
- Add lifecycle methods
- Update HandleKey signature
- Align with ToolApprovalOverlay pattern"
```

---

## Phase 2: Tool Log Fullscreen Conversion

### Task 6: Create ToolLogOverlay as FullscreenOverlay

**Files:**
- Create: `internal/ui/overlay_tool_log.go`

**Step 1: Write test for ToolLogOverlay**

Create `internal/ui/overlay_tool_log_test.go`:

```go
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestToolLogOverlay_IsFullscreen(t *testing.T) {
	lines := []string{"line 1", "line 2"}
	overlay := NewToolLogOverlay(&lines)

	assert.True(t, overlay.IsFullscreen())
}

func TestToolLogOverlay_GetDesiredHeight(t *testing.T) {
	lines := []string{}
	overlay := NewToolLogOverlay(&lines)

	// Fullscreen always wants max height
	assert.Equal(t, -1, overlay.GetDesiredHeight())
}

func TestToolLogOverlay_RefersToModelData(t *testing.T) {
	lines := []string{"initial"}
	overlay := NewToolLogOverlay(&lines)

	// Should reference lines, not copy
	lines = append(lines, "new line")

	content := overlay.GetContent()
	assert.Contains(t, content, "initial")
	assert.Contains(t, content, "new line")
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui -run TestToolLogOverlay -v
```

Expected: FAIL - NewToolLogOverlay not defined

**Step 3: Implement ToolLogOverlay struct**

Create `internal/ui/overlay_tool_log.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToolLogOverlay displays tool output in a fullscreen scrollable view
type ToolLogOverlay struct {
	lines    *[]string      // Reference to Model's tool log lines
	viewport viewport.Model // Embedded viewport for scrolling
	width    int
	height   int
}

// NewToolLogOverlay creates a new tool log overlay
func NewToolLogOverlay(lines *[]string) *ToolLogOverlay {
	return &ToolLogOverlay{
		lines:    lines,
		viewport: viewport.New(0, 0), // Initialized in OnPush
	}
}

// IsFullscreen returns true (this is a fullscreen overlay)
func (o *ToolLogOverlay) IsFullscreen() bool {
	return true
}

// GetDesiredHeight returns -1 (fullscreen wants all available height)
func (o *ToolLogOverlay) GetDesiredHeight() int {
	return -1
}

// GetHeader returns the overlay header
func (o *ToolLogOverlay) GetHeader() string {
	return "Tool Output Log"
}

// GetContent returns the current tool log lines
func (o *ToolLogOverlay) GetContent() string {
	if len(*o.lines) == 0 {
		return "No tool output in current chunk"
	}

	// Apply 10k line limit
	lines := *o.lines
	if len(lines) > 10000 {
		lines = lines[len(lines)-10000:]
	}

	return strings.Join(lines, "\n")
}

// GetFooter returns the overlay footer with line count
func (o *ToolLogOverlay) GetFooter() string {
	totalLines := len(*o.lines)
	if totalLines > 10000 {
		return fmt.Sprintf("Showing last 10,000 of %d lines • Esc to close", totalLines)
	}
	return fmt.Sprintf("%d lines • Esc to close", totalLines)
}

// OnPush initializes the viewport with dimensions
func (o *ToolLogOverlay) OnPush(width, height int) {
	o.width = width
	o.height = height

	// Initialize viewport (leave space for header and footer)
	o.viewport = viewport.New(width-4, height-6)
	o.viewport.SetContent(o.GetContent())

	// Auto-scroll to bottom
	o.viewport.GotoBottom()
}

// OnPop cleans up
func (o *ToolLogOverlay) OnPop() {
	// Nothing to clean up
}

// SetHeight updates the viewport height
func (o *ToolLogOverlay) SetHeight(height int) {
	o.height = height
	o.viewport.Height = height - 6
	o.viewport.SetContent(o.GetContent())
}

// Update handles viewport messages
func (o *ToolLogOverlay) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	o.viewport, cmd = o.viewport.Update(msg)

	// Update content on window size changes
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		o.viewport.SetContent(o.GetContent())
	}

	return cmd
}

// HandleKey processes keyboard input
func (o *ToolLogOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		// Pop will be handled by caller
		return true, nil

	case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDn:
		// Viewport navigation
		cmd := o.Update(msg)
		return true, cmd

	default:
		// Modal: capture all other keys
		return true, nil
	}
}

// Render returns the complete fullscreen view
func (o *ToolLogOverlay) Render(width, height int) string {
	var b strings.Builder

	// Update content if lines changed
	o.viewport.SetContent(o.GetContent())

	// Header with border
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("cyan"))

	closeHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Ctrl+O or Esc to close")

	header := headerStyle.Render(o.GetHeader())
	headerLine := fmt.Sprintf("┏━━ %s %s %s ┓",
		header,
		strings.Repeat("━", width-len(o.GetHeader())-len("Ctrl+O or Esc to close")-12),
		closeHint)
	b.WriteString(headerLine)
	b.WriteString("\n\n")

	// Viewport content
	b.WriteString(o.viewport.View())
	b.WriteString("\n\n")

	// Footer with border
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	b.WriteString(footerStyle.Render(o.GetFooter()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("┗%s┛", strings.Repeat("━", width-2)))

	return b.String()
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/ui -run TestToolLogOverlay -v
```

Expected: PASS

**Step 5: Commit ToolLogOverlay**

```bash
git add internal/ui/overlay_tool_log.go internal/ui/overlay_tool_log_test.go
git commit -m "feat: implement ToolLogOverlay as fullscreen overlay

- Implements FullscreenOverlay interface
- References Model's tool log lines directly
- Embeds viewport for scrolling (arrows, PageUp/Down)
- 10k line limit with truncation message
- Auto-scrolls to bottom on open
- Modal input capture"
```

---

### Task 7: Integrate ToolLogOverlay into Model

**Files:**
- Modify: `internal/ui/model.go`

**Step 1: Add ToolLogOverlay instance to Model**

Find Model struct and add field:

```go
type Model struct {
	// ... existing fields ...

	// Overlays
	overlayManager      *OverlayManager
	toolApprovalOverlay *ToolApprovalOverlay
	autocompleteOverlay *AutocompleteOverlay
	toolLogOverlay      *ToolLogOverlay  // Add this

	// Tool log data
	toolLogLines []string

	// ... rest of fields ...
}
```

**Step 2: Remove old tool log boolean**

Find and remove:

```go
toolLogOverlay bool  // Remove this line
```

**Step 3: Initialize ToolLogOverlay in NewModel**

Find NewModel function and add initialization:

```go
func NewModel(...) *Model {
	m := &Model{
		// ... existing initializations ...
	}

	// Initialize overlays
	m.overlayManager = NewOverlayManager()
	m.toolApprovalOverlay = NewToolApprovalOverlay(...)
	m.autocompleteOverlay = NewAutocompleteOverlay(...)
	m.toolLogOverlay = NewToolLogOverlay(&m.toolLogLines)  // Add this

	return m
}
```

**Step 4: Update Ctrl+O handler in update.go**

Find the Ctrl+O handler and replace:

```go
// Old code:
if msg.Type == tea.KeyCtrlO {
	m.toggleToolLogOverlay()
	return m, nil
}

// New code:
if msg.Type == tea.KeyCtrlO {
	if m.overlayManager.GetActive() == m.toolLogOverlay {
		// Already open, close it
		m.overlayManager.Pop()
	} else {
		// Open tool log
		m.overlayManager.Push(m.toolLogOverlay)
		m.toolLogOverlay.OnPush(m.Width, m.Height)
	}
	return m, nil
}
```

**Step 5: Remove old toollog.go functions**

Delete or comment out in `internal/ui/toollog.go`:
- `toggleToolLogOverlay()`
- `renderToolLogOverlay()`

Keep the helper functions:
- `appendToolLogLine()`
- `appendToolLogOutput()`
- `getToolLogLastN()`
- `clearToolLogChunk()`
- `startToolLogEntry()`
- `renderCollapsedToolLog()`

**Step 6: Update view.go rendering**

Find in `internal/ui/view.go` and remove:

```go
// Remove this:
if m.toolLogOverlay {
	return m.renderToolLogOverlay()
}
```

Add after checking onboarding:

```go
// Check for fullscreen overlay first
if m.overlayManager.IsFullscreen() {
	return m.overlayManager.Render(m.Width, m.Height)
}
```

**Step 7: Test manually**

```bash
make build
./hex
# Press Ctrl+O - should show tool log overlay
# Press Esc - should close
# Press Ctrl+O again - should reopen
```

**Step 8: Commit integration**

```bash
git add internal/ui/model.go internal/ui/update.go internal/ui/view.go internal/ui/toollog.go
git commit -m "feat: integrate ToolLogOverlay into Model

- Add toolLogOverlay instance to Model
- Remove toolLogOverlay boolean flag
- Update Ctrl+O to Push/Pop overlay from stack
- Update view.go to check IsFullscreen before rendering
- Keep helper functions for tool log data management"
```

---

## Phase 3: New Fullscreen Overlays

### Task 8: Implement HelpOverlay

**Files:**
- Create: `internal/ui/overlay_help.go`

**Step 1: Write test for HelpOverlay**

Create `internal/ui/overlay_help_test.go`:

```go
package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpOverlay_IsFullscreen(t *testing.T) {
	overlay := NewHelpOverlay()
	assert.True(t, overlay.IsFullscreen())
}

func TestHelpOverlay_GetContent(t *testing.T) {
	overlay := NewHelpOverlay()

	content := overlay.GetContent()
	assert.Contains(t, content, "Ctrl+O")
	assert.Contains(t, content, "Ctrl+H")
	assert.Contains(t, content, "Escape")
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui -run TestHelpOverlay -v
```

Expected: FAIL

**Step 3: Implement HelpOverlay**

Create `internal/ui/overlay_help.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpOverlay displays keyboard shortcuts and features
type HelpOverlay struct {
	viewport viewport.Model
	width    int
	height   int
}

// NewHelpOverlay creates a new help overlay
func NewHelpOverlay() *HelpOverlay {
	return &HelpOverlay{
		viewport: viewport.New(0, 0),
	}
}

// IsFullscreen returns true
func (o *HelpOverlay) IsFullscreen() bool {
	return true
}

// GetDesiredHeight returns -1 (fullscreen)
func (o *HelpOverlay) GetDesiredHeight() int {
	return -1
}

// GetHeader returns the header
func (o *HelpOverlay) GetHeader() string {
	return "Help & Keyboard Shortcuts"
}

// GetContent returns the help text
func (o *HelpOverlay) GetContent() string {
	return `# Keyboard Shortcuts

## Navigation
- **↑/↓**: Scroll viewport
- **PageUp/PageDown**: Page up/down
- **Ctrl+D/U**: Half page down/up
- **Home/End**: Go to top/bottom

## Overlays
- **Ctrl+O**: Toggle tool output log
- **Ctrl+H**: Toggle this help screen
- **Ctrl+R**: Open conversation history
- **Escape**: Close active overlay

## Input
- **Enter**: Send message
- **Shift+Enter**: New line in message
- **Ctrl+C**: Cancel stream or close overlay

## Tools
- **Enter**: Approve tool (when prompted)
- **Escape**: Deny tool (when prompted)

## Other
- **Ctrl+L**: Clear screen
- **Ctrl+Q**: Quit application

# Tips

- Use overlays to view detailed information without losing context
- All overlays are scrollable with arrow keys and PageUp/PageDown
- Tool output log shows the last 10,000 lines of tool execution
- Conversation history is limited to the last 1,000 messages
`
}

// GetFooter returns the footer
func (o *HelpOverlay) GetFooter() string {
	return "Press Escape or Ctrl+H to close"
}

// OnPush initializes viewport
func (o *HelpOverlay) OnPush(width, height int) {
	o.width = width
	o.height = height
	o.viewport = viewport.New(width-4, height-6)
	o.viewport.SetContent(o.GetContent())
}

// OnPop cleans up
func (o *HelpOverlay) OnPop() {}

// SetHeight updates viewport height
func (o *HelpOverlay) SetHeight(height int) {
	o.height = height
	o.viewport.Height = height - 6
}

// Update handles messages
func (o *HelpOverlay) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	o.viewport, cmd = o.viewport.Update(msg)
	return cmd
}

// HandleKey processes input
func (o *HelpOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC, tea.KeyCtrlH:
		return true, nil // Pop handled by caller

	case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDn,
		tea.KeyHome, tea.KeyEnd, tea.KeyCtrlD, tea.KeyCtrlU:
		cmd := o.Update(msg)
		return true, cmd

	default:
		return true, nil // Modal capture
	}
}

// Render returns the complete view
func (o *HelpOverlay) Render(width, height int) string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("cyan"))
	b.WriteString(fmt.Sprintf("┏━━ %s ┓\n\n", headerStyle.Render(o.GetHeader())))

	// Content
	b.WriteString(o.viewport.View())
	b.WriteString("\n\n")

	// Footer
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(footerStyle.Render(o.GetFooter()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("┗%s┛", strings.Repeat("━", width-2)))

	return b.String()
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/ui -run TestHelpOverlay -v
```

Expected: PASS

**Step 5: Integrate HelpOverlay into Model**

In `internal/ui/model.go`:

```go
// Add to Model struct
helpOverlay *HelpOverlay

// Add to NewModel
m.helpOverlay = NewHelpOverlay()
```

In `internal/ui/update.go`, add handler:

```go
// Handle Ctrl+H: toggle help overlay
if msg.Type == tea.KeyCtrlH {
	if m.overlayManager.GetActive() == m.helpOverlay {
		m.overlayManager.Pop()
	} else {
		m.overlayManager.Push(m.helpOverlay)
		m.helpOverlay.OnPush(m.Width, m.Height)
	}
	return m, nil
}
```

**Step 6: Commit HelpOverlay**

```bash
git add internal/ui/overlay_help.go internal/ui/overlay_help_test.go internal/ui/model.go internal/ui/update.go
git commit -m "feat: implement help overlay (Ctrl+H)

- Fullscreen overlay with keyboard shortcuts
- Scrollable markdown-style help content
- Integrated with Ctrl+H hotkey
- Modal input capture"
```

---

### Task 9: Implement HistoryOverlay

**Files:**
- Create: `internal/ui/overlay_history.go`

**Step 1: Write test for HistoryOverlay**

Create `internal/ui/overlay_history_test.go`:

```go
package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHistoryOverlay_IsFullscreen(t *testing.T) {
	messages := []Message{}
	overlay := NewHistoryOverlay(&messages)
	assert.True(t, overlay.IsFullscreen())
}

func TestHistoryOverlay_RefersToModelMessages(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "Hello", Timestamp: time.Now()},
	}
	overlay := NewHistoryOverlay(&messages)

	// Should reference messages, not copy
	messages = append(messages, Message{
		Role:      "assistant",
		Content:   "Hi there",
		Timestamp: time.Now(),
	})

	content := overlay.GetContent()
	assert.Contains(t, content, "Hello")
	assert.Contains(t, content, "Hi there")
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui -run TestHistoryOverlay -v
```

Expected: FAIL

**Step 3: Implement HistoryOverlay**

Create `internal/ui/overlay_history.go`:

```go
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HistoryOverlay displays conversation history
type HistoryOverlay struct {
	messages *[]Message
	viewport viewport.Model
	width    int
	height   int
}

// NewHistoryOverlay creates a new history overlay
func NewHistoryOverlay(messages *[]Message) *HistoryOverlay {
	return &HistoryOverlay{
		messages: messages,
		viewport: viewport.New(0, 0),
	}
}

// IsFullscreen returns true
func (o *HistoryOverlay) IsFullscreen() bool {
	return true
}

// GetDesiredHeight returns -1 (fullscreen)
func (o *HistoryOverlay) GetDesiredHeight() int {
	return -1
}

// GetHeader returns the header
func (o *HistoryOverlay) GetHeader() string {
	return fmt.Sprintf("Conversation History (%d messages)", len(*o.messages))
}

// GetContent returns formatted message history
func (o *HistoryOverlay) GetContent() string {
	if len(*o.messages) == 0 {
		return "No messages in conversation"
	}

	// Apply 1000 message limit
	messages := *o.messages
	if len(messages) > 1000 {
		messages = messages[len(messages)-1000:]
	}

	var b strings.Builder
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	assistantStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	for i, msg := range messages {
		// Timestamp
		timestamp := msg.Timestamp.Format("15:04:05")
		b.WriteString(timeStyle.Render(timestamp))
		b.WriteString(" ")

		// Role
		if msg.Role == "user" {
			b.WriteString(userStyle.Render("[YOU]"))
		} else {
			b.WriteString(assistantStyle.Render("[ASSISTANT]"))
		}
		b.WriteString("\n")

		// Content (truncate long messages)
		content := msg.Content
		if len(content) > 500 {
			content = content[:497] + "..."
		}
		b.WriteString(content)
		b.WriteString("\n")

		// Separator between messages
		if i < len(messages)-1 {
			b.WriteString(strings.Repeat("─", 80))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// GetFooter returns the footer
func (o *HistoryOverlay) GetFooter() string {
	totalMessages := len(*o.messages)
	if totalMessages > 1000 {
		return fmt.Sprintf("Showing last 1,000 of %d messages • Escape to close", totalMessages)
	}
	return fmt.Sprintf("%d messages • Escape to close", totalMessages)
}

// OnPush initializes viewport
func (o *HistoryOverlay) OnPush(width, height int) {
	o.width = width
	o.height = height
	o.viewport = viewport.New(width-4, height-6)
	o.viewport.SetContent(o.GetContent())
	o.viewport.GotoBottom() // Start at most recent
}

// OnPop cleans up
func (o *HistoryOverlay) OnPop() {}

// SetHeight updates viewport height
func (o *HistoryOverlay) SetHeight(height int) {
	o.height = height
	o.viewport.Height = height - 6
}

// Update handles messages
func (o *HistoryOverlay) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	o.viewport, cmd = o.viewport.Update(msg)

	// Update content on window size changes
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		o.viewport.SetContent(o.GetContent())
	}

	return cmd
}

// HandleKey processes input
func (o *HistoryOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC, tea.KeyCtrlR:
		return true, nil // Pop handled by caller

	case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDn,
		tea.KeyHome, tea.KeyEnd, tea.KeyCtrlD, tea.KeyCtrlU:
		cmd := o.Update(msg)
		return true, cmd

	default:
		return true, nil // Modal capture
	}
}

// Render returns the complete view
func (o *HistoryOverlay) Render(width, height int) string {
	var b strings.Builder

	// Update content if messages changed
	o.viewport.SetContent(o.GetContent())

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("cyan"))
	b.WriteString(fmt.Sprintf("┏━━ %s ┓\n\n", headerStyle.Render(o.GetHeader())))

	// Content
	b.WriteString(o.viewport.View())
	b.WriteString("\n\n")

	// Footer
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(footerStyle.Render(o.GetFooter()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("┗%s┛", strings.Repeat("━", width-2)))

	return b.String()
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/ui -run TestHistoryOverlay -v
```

Expected: PASS

**Step 5: Integrate HistoryOverlay into Model**

In `internal/ui/model.go`:

```go
// Add to Model struct
historyOverlay *HistoryOverlay

// Add to NewModel
m.historyOverlay = NewHistoryOverlay(&m.messages)
```

In `internal/ui/update.go`, add handler:

```go
// Handle Ctrl+R: toggle history overlay
if msg.Type == tea.KeyCtrlR {
	if m.overlayManager.GetActive() == m.historyOverlay {
		m.overlayManager.Pop()
	} else {
		m.overlayManager.Push(m.historyOverlay)
		m.historyOverlay.OnPush(m.Width, m.Height)
	}
	return m, nil
}
```

**Step 6: Commit HistoryOverlay**

```bash
git add internal/ui/overlay_history.go internal/ui/overlay_history_test.go internal/ui/model.go internal/ui/update.go
git commit -m "feat: implement history overlay (Ctrl+R)

- Fullscreen overlay showing last 1000 messages
- Scrollable conversation history with timestamps
- Color-coded roles (user/assistant)
- Integrated with Ctrl+R hotkey
- Modal input capture"
```

---

## Task 10: Update Bottom Overlay Rendering in View

**Files:**
- Modify: `internal/ui/view.go`

**Step 1: Find bottom overlay rendering section**

Locate the section that renders bottom overlays (currently around line 85-90).

**Step 2: Update to calculate height and push viewport up**

Replace bottom overlay rendering:

```go
// Render bottom overlays between viewport and input
var bottomOverlayContent string
var bottomOverlayHeight int
if m.overlayManager.HasActive() && !m.overlayManager.IsFullscreen() {
	active := m.overlayManager.GetActive()

	// Calculate desired height with 40% cap
	desiredHeight := active.GetDesiredHeight()
	maxAllowed := int(float64(m.Height) * 0.4)
	if desiredHeight > maxAllowed {
		bottomOverlayHeight = maxAllowed
	} else {
		bottomOverlayHeight = desiredHeight
	}

	// Render overlay
	bottomOverlayContent = active.Render(m.Width, bottomOverlayHeight)
}

// Adjust viewport height if bottom overlay present
viewportHeight := m.Height - headerHeight - inputHeight - statusHeight - bottomOverlayHeight
m.viewport.Height = viewportHeight
```

**Step 3: Render viewport, then overlay, then input**

Update rendering order:

```go
// Render viewport
b.WriteString(m.viewport.View())
b.WriteString("\n")

// Render bottom overlay if present
if bottomOverlayContent != "" {
	b.WriteString(bottomOverlayContent)
	b.WriteString("\n")
}

// Render input
b.WriteString(m.renderInput())
```

**Step 4: Test manually**

```bash
make build
./hex
# Test tool approval - should push viewport up
# Test autocomplete - should push viewport up
# Test Ctrl+O - should take full screen
```

**Step 5: Commit view updates**

```bash
git add internal/ui/view.go
git commit -m "refactor: update bottom overlay rendering to push viewport

- Calculate bottom overlay height with 40% cap
- Adjust viewport height dynamically
- Render bottom overlays between viewport and input
- Maintain fullscreen overlay check at top"
```

---

## Task 11: Update Input Routing in Update

**Files:**
- Modify: `internal/ui/update.go`

**Step 1: Find main keyboard handler**

Locate the keyboard input section (likely in `handleKeyboard` or main `Update` switch).

**Step 2: Add overlay input routing at top**

Add at the beginning of keyboard handling:

```go
// Route input to active overlay first (modal behavior)
if m.overlayManager.HasActive() {
	handled, cmd := m.overlayManager.HandleKey(msg)
	if handled {
		// Check if overlay was dismissed (need to pop)
		if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
			active := m.overlayManager.GetActive()
			// Let overlay handle first, then pop
			if cmd == nil {
				m.overlayManager.Pop()
			}
		}
		return m, cmd
	}
	// Overlay didn't handle - shouldn't happen with modal capture
	// but fall through anyway
}
```

**Step 3: Route viewport updates to scrollable overlays**

After overlay key routing, add:

```go
// Route viewport updates to scrollable overlays
if m.overlayManager.HasActive() {
	active := m.overlayManager.GetActive()
	if scrollable, ok := active.(Scrollable); ok {
		// Let overlay handle viewport navigation
		if cmd := scrollable.Update(msg); cmd != nil {
			return m, cmd
		}
	}
}
```

**Step 4: Remove old overlay-specific handlers**

Remove or comment out:
- Old `handleEscape()` calls for overlays
- Old `HandleCtrlC()` calls for overlays
- Specific overlay key routing (now handled by overlay itself)

**Step 5: Test all hotkeys**

```bash
make build
./hex
# Ctrl+O - tool log opens, Esc closes
# Ctrl+H - help opens, Esc closes
# Ctrl+R - history opens, Esc closes
# Arrow keys scroll in fullscreen
# Tool approval: Enter/Esc work
# Autocomplete: Enter/Esc work
```

**Step 6: Commit input routing updates**

```bash
git add internal/ui/update.go
git commit -m "refactor: unify input routing through overlay manager

- Route all input to overlay manager first
- Handle Escape/Ctrl+C to pop overlays
- Route viewport updates to scrollable overlays
- Remove old overlay-specific handlers
- Modal behavior: overlay captures all input"
```

---

## Task 12: Integration Testing & Cleanup

**Files:**
- Test all overlay functionality
- Clean up old code
- Update documentation

**Step 1: Write integration tests**

Create `internal/ui/overlay_integration_test.go`:

```go
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestOverlayIntegration_StackMultipleBottom(t *testing.T) {
	om := NewOverlayManager()

	// Push two bottom overlays
	overlay1 := newMockOverlay("H1", "C1", "F1", 5)
	overlay2 := newMockOverlay("H2", "C2", "F2", 10)

	om.Push(overlay1)
	om.Push(overlay2)

	// Top should be active
	assert.Equal(t, overlay2, om.GetActive())
	assert.False(t, om.IsFullscreen())

	// Pop one
	om.Pop()
	assert.Equal(t, overlay1, om.GetActive())
}

func TestOverlayIntegration_FullscreenTakesOver(t *testing.T) {
	om := NewOverlayManager()

	// Push bottom overlay
	bottom := newMockOverlay("Bottom", "Content", "Footer", 5)
	om.Push(bottom)

	// Push fullscreen overlay (simulated)
	lines := []string{"line1", "line2"}
	fullscreen := NewToolLogOverlay(&lines)
	om.Push(fullscreen)

	// Should be fullscreen
	assert.True(t, om.IsFullscreen())
	assert.Equal(t, fullscreen, om.GetActive())

	// Pop fullscreen
	om.Pop()
	assert.False(t, om.IsFullscreen())
	assert.Equal(t, bottom, om.GetActive())
}
```

**Step 2: Run integration tests**

```bash
go test ./internal/ui -run TestOverlayIntegration -v
```

Expected: All pass

**Step 3: Clean up old overlay code**

Remove deprecated files/functions:
- Delete old rendering functions in toollog.go (if not removed)
- Remove any old overlay-specific handlers in update.go
- Clean up commented-out code

**Step 4: Update help text and documentation**

Update internal docs:
- Overlay architecture
- Adding new overlays guide
- Testing patterns

**Step 5: Final manual testing checklist**

Test all functionality:
```
[ ] Ctrl+O opens tool log
[ ] Tool log shows recent output
[ ] Tool log scrolls with arrows/PageUp/PageDown
[ ] Esc closes tool log
[ ] Ctrl+H opens help
[ ] Help scrolls properly
[ ] Ctrl+H or Esc closes help
[ ] Ctrl+R opens history
[ ] History shows messages
[ ] History scrolls properly
[ ] Esc closes history
[ ] Tool approval appears when tool requested
[ ] Enter approves tool
[ ] Esc denies tool
[ ] Autocomplete appears on slash command
[ ] Enter selects command
[ ] Esc dismisses autocomplete
[ ] Bottom overlays push viewport up
[ ] Multiple overlays can stack (bottom only)
[ ] All overlays are modal (capture input)
```

**Step 6: Commit integration tests and cleanup**

```bash
git add internal/ui/overlay_integration_test.go
git commit -m "test: add integration tests for overlay system

- Test stacking multiple bottom overlays
- Test fullscreen overlay takes precedence
- Test modal input capture across overlay types
- Manual testing checklist complete"
```

**Step 7: Final commit - mark as complete**

```bash
git commit --allow-empty -m "feat: complete overlay system refactor

All three phases complete:
- Phase 1: Core interfaces and stack-based manager
- Phase 2: Tool log conversion to fullscreen overlay
- Phase 3: New help and history fullscreen overlays

Features:
- Unified stack-based overlay management
- Composable interfaces (Overlay, Scrollable, FullscreenOverlay)
- True modal behavior with input capture
- Scrollable fullscreen overlays with viewport
- Bottom overlays push viewport up
- 10k line tool log, 1k message history limits

All tests passing. Ready for review."
```

---

## Implementation Complete

Plan saved to: `docs/plans/2025-12-11-overlay-system-refactor.md`

**Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration with quality gates

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with review checkpoints

**Which approach?**
