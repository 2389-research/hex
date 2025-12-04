# Crush UI Improvements - Phase 1

**Date:** 2025-12-03
**Objective:** Adopt key patterns from Bubbletea team's Crush agent to make our UI more robust
**Scope:** Foundation improvements - interfaces, caching, adaptive layout

## Background

Analysis of Crush agent revealed several production-grade patterns:
- Component interfaces for standardization
- Content caching for performance
- Adaptive layout for terminal size responsiveness
- Better size propagation
- Panic recovery

This phase focuses on foundation improvements that enable future architectural changes.

## Tasks

### Task 1: Define Core Component Interfaces

**Goal:** Create standard interfaces that all UI components implement

**Files to create:**
- `internal/ui/components/interfaces.go` - Core component interfaces

**Interfaces to define:**

```go
// Sizeable components can be resized
type Sizeable interface {
    SetSize(width, height int) tea.Cmd
    GetSize() (int, int)
}

// Focusable components can receive focus
type Focusable interface {
    Focus() tea.Cmd
    Blur() tea.Cmd
    IsFocused() bool
}

// Helpable components provide help text
type Helpable interface {
    HelpView() string
}

// Component is the base interface all UI components should implement
type Component interface {
    tea.Model
    Sizeable
}
```

**Success criteria:**
- Interfaces compile
- Clear documentation for each interface
- Examples in doc comments

**Testing:**
- No tests needed (just interface definitions)
- Will verify in next task when migrating components

---

### Task 2: Migrate Existing Components to Interfaces

**Goal:** Update HuhApproval, Help, and Error components to implement Component interface

**Files to modify:**
- `internal/ui/components/huh_approval.go`
- `internal/ui/components/help.go`
- `internal/ui/components/error.go`

**Changes for each component:**

1. Add `SetSize(width, height int) tea.Cmd` method
2. Add `GetSize() (int, int)` method
3. Store width/height as fields
4. Update rendering to respect stored dimensions

**Example (HuhApproval):**
```go
type HuhApproval struct {
    theme       themes.Theme
    toolName    string
    description string
    approved    bool
    form        *huh.Form
    width       int  // NEW
    height      int  // NEW
}

func (h *HuhApproval) SetSize(width, height int) tea.Cmd {
    h.width = width
    h.height = height
    if h.form != nil {
        h.form = h.form.WithWidth(width)
    }
    return nil
}

func (h *HuhApproval) GetSize() (int, int) {
    return h.width, h.height
}
```

**Success criteria:**
- All three components implement Component interface
- Size is properly stored and used in rendering
- Existing tests still pass

**Testing:**
- Update existing component tests to verify SetSize/GetSize
- Add test that components implement Component interface

---

### Task 3: Add Content Caching to Expensive Renders

**Goal:** Cache expensive rendering operations (markdown, help text) to improve performance

**Files to modify:**
- `internal/ui/model.go` - Add markdown cache
- `internal/ui/components/help.go` - Add help text cache

**Implementation:**

1. **Markdown caching in Model:**
```go
type Model struct {
    // ... existing fields ...

    // Markdown rendering cache
    markdownCache      map[string]string // messageID -> rendered markdown
    markdownCacheDirty bool
}

func (m *Model) RenderMessage(msg Message) string {
    cacheKey := msg.ID
    if cached, ok := m.markdownCache[cacheKey]; ok && !m.markdownCacheDirty {
        return cached
    }

    rendered := m.expensiveMarkdownRender(msg.Content)
    if m.markdownCache == nil {
        m.markdownCache = make(map[string]string)
    }
    m.markdownCache[cacheKey] = rendered
    return rendered
}

func (m *Model) InvalidateMarkdownCache() {
    m.markdownCacheDirty = true
}

func (m *Model) ClearMarkdownCache() {
    m.markdownCache = make(map[string]string)
    m.markdownCacheDirty = false
}
```

2. **Help text caching:**
```go
type Help struct {
    theme         themes.Theme
    cachedContent string
    contentDirty  bool
    width         int
    height        int
}

func (h *Help) View() string {
    if !h.contentDirty && h.cachedContent != "" {
        return h.cachedContent
    }

    content := h.generateHelpText()
    h.cachedContent = content
    h.contentDirty = false
    return content
}

func (h *Help) SetSize(width, height int) tea.Cmd {
    if width != h.width || height != h.height {
        h.contentDirty = true  // Size changed, invalidate cache
    }
    h.width = width
    h.height = height
    return nil
}
```

**Success criteria:**
- Markdown renders are cached per message ID
- Help text is cached and invalidated on resize
- Cache can be cleared/invalidated
- Performance improvement measurable (benchmark)

**Testing:**
- Unit test that verifies cache hit/miss
- Benchmark showing performance improvement
- Test cache invalidation on resize

---

### Task 4: Implement Adaptive Layout System

**Goal:** Detect terminal size and switch between compact/wide layouts

**Files to create:**
- `internal/ui/layout_mode.go` - Layout mode logic

**Files to modify:**
- `internal/ui/model.go` - Add layout mode tracking
- `internal/ui/view.go` - Render based on layout mode
- `internal/ui/update.go` - Handle layout mode changes on resize

**Implementation:**

1. **Layout mode types:**
```go
// layout_mode.go
type LayoutMode int

const (
    LayoutModeWide LayoutMode = iota    // Full width, all features
    LayoutModeCompact                    // Narrow terminals, simplified
)

const (
    CompactModeWidthBreakpoint  = 100
    CompactModeHeightBreakpoint = 24
)

func DetermineLayoutMode(width, height int) LayoutMode {
    if width < CompactModeWidthBreakpoint || height < CompactModeHeightBreakpoint {
        return LayoutModeCompact
    }
    return LayoutModeWide
}
```

2. **Model changes:**
```go
type Model struct {
    // ... existing fields ...
    layoutMode      LayoutMode
    forceLayoutMode bool  // User override
}

func (m *Model) UpdateLayoutMode(width, height int) {
    if m.forceLayoutMode {
        return
    }

    newMode := DetermineLayoutMode(width, height)
    if newMode != m.layoutMode {
        m.layoutMode = newMode
        // Trigger re-layout
    }
}
```

3. **View rendering:**
```go
func (m *Model) View() string {
    if m.layoutMode == LayoutModeCompact {
        return m.renderCompactLayout()
    }
    return m.renderWideLayout()
}

func (m *Model) renderCompactLayout() string {
    // Single column, minimal UI
    // No token viz, simplified status bar
}

func (m *Model) renderWideLayout() string {
    // Current full-featured layout
}
```

**Success criteria:**
- Layout automatically switches at breakpoints
- Compact mode is functional (all features work)
- Wide mode is unchanged
- User can force layout mode with flag/command

**Testing:**
- Test layout mode detection with various sizes
- Test compact mode rendering
- Test wide mode rendering
- Test forced layout mode

---

### Task 5: Add Panic Recovery to Components

**Goal:** Prevent component crashes from taking down the entire TUI

**Files to modify:**
- `internal/ui/update.go` - Add panic recovery wrapper
- `internal/ui/components/huh_approval.go` - Add recovery to Update()

**Implementation:**

1. **Recovery utility:**
```go
// internal/ui/recovery.go
func RecoverPanic(component string, fn func()) {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("Component panic recovered",
                "component", component,
                "panic", r,
                "stack", string(debug.Stack()))
        }
    }()
    fn()
}
```

2. **Update wrapper:**
```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var result tea.Model
    var cmd tea.Cmd

    RecoverPanic("Model.Update", func() {
        result, cmd = m.update(msg)
    })

    if result == nil {
        return m, cmd  // Panic occurred, return stable state
    }
    return result, cmd
}

func (m *Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Existing Update logic here
}
```

3. **Component recovery:**
```go
func (h *HuhApproval) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var result tea.Model = h
    var cmd tea.Cmd

    RecoverPanic("HuhApproval.Update", func() {
        // Existing update logic
        form, formCmd := h.form.Update(msg)
        if f, ok := form.(*huh.Form); ok {
            h.form = f
            cmd = formCmd
        }
    })

    return result, cmd
}
```

**Success criteria:**
- Panics in Update() are caught and logged
- TUI stays alive after component panic
- Stable state is returned after panic
- Stack traces logged for debugging

**Testing:**
- Test that induces panic in component Update()
- Verify TUI doesn't crash
- Verify error is logged
- Verify stable state returned

---

### Task 6: Implement Size Propagation Chain

**Goal:** Properly propagate size changes from parent to children components

**Files to modify:**
- `internal/ui/model.go` - Add size propagation to children
- `internal/ui/update.go` - Handle WindowSizeMsg properly

**Implementation:**

```go
func (m *Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) tea.Cmd {
    m.Width = msg.Width
    m.Height = msg.Height

    var cmds []tea.Cmd

    // Propagate to viewport
    if m.Viewport != nil {
        m.Viewport.Width = msg.Width
        m.Viewport.Height = msg.Height - m.getReservedHeight()
    }

    // Propagate to input
    if m.Input != nil {
        m.Input.SetWidth(msg.Width - 4)
    }

    // Propagate to components
    if m.helpComponent != nil {
        cmds = append(cmds, m.helpComponent.SetSize(msg.Width, msg.Height))
    }

    if m.errorComponent != nil {
        cmds = append(cmds, m.errorComponent.SetSize(msg.Width, msg.Height))
    }

    if m.huhApproval != nil {
        cmds = append(cmds, m.huhApproval.SetSize(msg.Width, msg.Height))
    }

    // Update layout mode based on new size
    m.UpdateLayoutMode(msg.Width, msg.Height)

    return tea.Batch(cmds...)
}

func (m *Model) getReservedHeight() int {
    reserved := 0
    reserved += 3  // Input area
    reserved += 1  // Status bar
    if m.tokenViz != nil && m.showTokenViz {
        reserved += 3  // Token visualization
    }
    return reserved
}
```

**Success criteria:**
- All components receive size updates on WindowSizeMsg
- Components render correctly at new sizes
- No components with incorrect dimensions
- Layout mode updates on resize

**Testing:**
- Test WindowSizeMsg propagation to all components
- Test components render correctly after resize
- Test layout mode switches on resize
- Test reserved height calculation

---

## Implementation Plan

**Execution order:**
1. Task 1 (Interfaces) - Foundation
2. Task 2 (Migrate components) - Apply foundation
3. Task 6 (Size propagation) - Critical for next tasks
4. Task 3 (Caching) - Performance
5. Task 4 (Adaptive layout) - UX
6. Task 5 (Panic recovery) - Robustness

**Estimated time:** 3-4 hours with subagents

**Dependencies:**
- Task 2 depends on Task 1
- Task 4 depends on Task 6
- Task 5 can run anytime

## Testing Strategy

- Unit tests for all new functions
- Integration tests for size propagation
- Visual tests for layout modes
- Performance benchmarks for caching
- Panic recovery tests

## Success Metrics

- All tests passing
- No regressions in existing features
- Measurable performance improvement (caching)
- Responsive UI at various terminal sizes
- No crashes from component panics

## Future Work (Phase 2)

After Phase 1 is complete:
- Pubsub event architecture
- Service layer extraction
- Full permission system (like Crush)
- Mouse event handling
- Layer-based rendering
