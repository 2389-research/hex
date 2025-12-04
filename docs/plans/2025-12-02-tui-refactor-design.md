# Jeff TUI Refactor Design

**Date:** 2025-12-02
**Status:** Approved
**Goal:** Transform Jeff into "Claude Code for Productivity" with rich inline components

## Vision

Create a chat-first terminal UI where an AI agent helps manage email, calendar, and tasks. The agent invokes tools that render rich interactive components (tables, forms, progress bars) directly inline in the conversation, similar to how Claude Code displays file contents and bash output.

## Core Concept

**"Agentic Mutt"** - Not replacing structured views with chat, but augmenting productivity tools with an AI copilot. The agent is the primary interface that summons and manipulates rich components as needed.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Status Bar (Dracula themed)                        │
│  📧 5 unread  📅 Meeting in 15m  ✓ 3 tasks today   │
├─────────────────────────────────────────────────────┤
│                                                      │
│  Chat Viewport (scrollable conversation)            │
│  ┌──────────────────────────────────────────┐      │
│  │ User: Show my unread emails              │      │
│  │                                           │      │
│  │ Agent: Here are your 5 unread emails:    │      │
│  │ ┌────────────────────────────────────┐   │      │
│  │ │ From          │ Subject       │ Date│   │      │
│  │ │ boss@co.com   │ Q4 Review     │ 2h  │◄─ Bubbles
│  │ │ sarah@co.com  │ Lunch?        │ 4h  │   Table
│  │ └────────────────────────────────────┘   │      │
│  │ [Interactive: ↑↓ select, Enter=open]     │      │
│  └──────────────────────────────────────────┘      │
│                                                      │
├─────────────────────────────────────────────────────┤
│  Input (textarea)                                   │
│  > Archive the Q4 review email_                     │
└─────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Chat-Primary, Tools-Secondary
- Conversation history is the main UI element
- Rich components (tables, lists) render inline as tool results
- Components are interactive but embedded in chat flow
- Scroll through history to see previous tool invocations

### 2. Three-Layer Architecture
```
internal/ui/
├── themes/          # Dracula, Gruvbox, Nord color schemes
│   ├── theme.go     # Theme interface and types
│   ├── dracula.go   # Dracula theme (default)
│   ├── gruvbox.go   # Gruvbox Dark theme
│   └── nord.go      # Nord theme
├── components/      # Rich bubbles-based components
│   ├── base.go      # Component interface
│   ├── table.go     # Email/task lists
│   ├── progress.go  # Streaming indicators
│   ├── forms.go     # Huh-based forms
│   ├── approval.go  # Tool approval using huh
│   └── list.go      # Lists with fuzzy search
├── model.go         # Main Bubbletea model (refactored)
└── views/           # Render logic for different states
```

## Theming System

### Theme Structure
```go
type Theme struct {
    Name        string

    // Base colors
    Background  lipgloss.Color
    Foreground  lipgloss.Color

    // Semantic colors
    Primary     lipgloss.Color  // Main accent
    Secondary   lipgloss.Color  // Secondary accent
    Success     lipgloss.Color  // Green for completed
    Warning     lipgloss.Color  // Yellow for pending
    Error       lipgloss.Color  // Red for errors

    // UI elements
    Border      lipgloss.Color
    BorderFocus lipgloss.Color
    Subtle      lipgloss.Color  // Muted text

    // Gradients
    TitleGradient []lipgloss.Color
}
```

### Three Built-in Themes

**1. Dracula (Default)**
- Background: #282a36
- Primary: #bd93f9 (purple)
- Secondary: #ff79c6 (pink)
- Success: #50fa7b (green)
- Warning: #f1fa8c (yellow)
- Error: #ff5555 (red)
- Title gradient: Purple → Pink

**2. Gruvbox Dark**
- Background: #282828
- Primary: #d79921 (orange)
- Secondary: #b16286 (purple)
- Warm, earthy tones
- Title gradient: Orange → Red

**3. Nord**
- Background: #2e3440
- Primary: #88c0d0 (cyan)
- Secondary: #81a1c1 (blue)
- Cool, nordic aesthetic
- Title gradient: Cyan → Blue

### Theme Configuration

**Config file (~/.jeff/config.yaml):**
```yaml
theme: dracula  # or "gruvbox", "nord"
```

**CLI flag:**
```bash
jefft --theme nord
```

## Rich Components

### Component System

Each rich component is a self-contained Bubbletea model embedded in chat messages.

**1. Interactive Table (emails, tasks, search results)**
```go
type TableComponent struct {
    table  table.Model  // From bubbles
    theme  *Theme
    data   [][]string

    // Interactions
    onSelect func(row int)
    onAction func(key string, row int)
}

// Keybindings in table:
// ↑↓: Navigate rows
// Enter: Select/open
// a: Archive (context-dependent)
// d: Delete
// r: Reply (for emails)
```

**2. Progress Indicators**
```go
type ProgressComponent struct {
    progress progress.Model  // From bubbles
    label    string
    value    float64  // 0.0 to 1.0
}

// Uses:
// - Streaming response progress
// - Token usage bar (tokens/limit)
// - Task completion percentage
```

**3. Huh Forms (tool approval, quick actions)**
```go
// Tool approval:
huh.NewForm(
    huh.NewConfirm().
        Title("Execute bash command?").
        Description("rm -rf /tmp/cache").
        Affirmative("Yes").
        Negative("No"),
)

// Quick actions:
huh.NewSelect().
    Title("Choose action").
    Options(
        huh.NewOption("Archive all", "archive"),
        huh.NewOption("Mark as read", "read"),
        huh.NewOption("Cancel", "cancel"),
    )
```

**4. List Component (conversation history, skill selector)**
```go
type ListComponent struct {
    list list.Model  // From bubbles
    items []list.Item
    showFilter bool  // Enable fuzzy search
}
```

### Component Lifecycle

**Rendering in Chat:**
1. Agent invokes tool (e.g., `search_emails`)
2. Tool returns structured data + component type
3. UI creates component instance with theme
4. Component renders inline in viewport
5. Component handles its own key events when focused

## Migration Plan

### Phase 1: Foundation (Days 1-2)

**Dependencies:**
```bash
go get github.com/charmbracelet/huh@latest
# bubbles already installed
# lipgloss already installed
```

**New Files:**
```
internal/ui/themes/
├── theme.go          # Theme interface and types
├── dracula.go        # Dracula theme
├── gruvbox.go        # Gruvbox Dark theme
└── nord.go           # Nord theme

internal/ui/components/
└── base.go           # Component interface
```

**Updates:**
- Refactor `model.go` to use theme system
- Update all existing lipgloss styles to use theme colors
- Add theme loading from config

**Goal:** All existing UI renders with Dracula theme, config-switchable

---

### Phase 2: Huh Integration (Days 2-3)

**New Files:**
```
internal/ui/components/
├── forms.go          # Huh form wrappers
└── approval.go       # Tool approval using huh (replaces old approval.go)
```

**Updates:**
- Replace tool approval dialog with huh confirm
- Replace quick actions with huh select
- Update model.go to handle huh form events

**Goal:** Tool approval and quick actions use huh forms with proper theming

---

### Phase 3: Rich Components (Days 3-4)

**New Files:**
```
internal/ui/components/
├── table.go          # Bubbles table wrapper
├── progress.go       # Progress bars for streaming/tokens
└── list.go           # Lists with fuzzy search
```

**Updates:**
- Modify message rendering to support embedded components
- Add component focus management in update loop
- Create example tool that returns table data

**Goal:** Can render interactive tables inline in chat

---

### Phase 4: Visual Polish (Day 5)

**Updates:**
- Add gradient support for title bar
- Improve borders and spacing throughout
- Enhanced markdown rendering with theme colors
- Smooth state transition animations

**Goal:** Professional, polished appearance with great attention to detail

## Technical Implementation Details

### Component Focus Management

Focus stack pattern to handle nested interactive components:

```go
type Model struct {
    focusStack []Focusable  // Stack of focusable components
    // ... existing fields
}

// When table appears: push to stack
// Table gets keys until ESC/blur
// Pop from stack returns focus to input
```

### Message with Embedded Components

```go
type Message struct {
    Role         string
    Content      string
    Component    Component  // Optional rich component
    ComponentID  string     // For event routing
}

// Rendering:
func (m Message) View() string {
    result := renderMarkdown(m.Content)
    if m.Component != nil {
        result += "\n" + m.Component.View()
    }
    return result
}
```

### Theme Hot-Reloading

```go
// When theme changes:
func (m *Model) SetTheme(theme *Theme) {
    m.theme = theme
    // Update all components
    for _, msg := range m.Messages {
        if msg.Component != nil {
            msg.Component.SetTheme(theme)
        }
    }
    m.Input.TextStyle = theme.InputStyle()
    m.Viewport.Style = theme.ViewportStyle()
}
```

### Streaming with Progress

```go
// When streaming:
type StreamingMessage struct {
    Content   string
    Progress  *ProgressComponent
    Tokens    int
    MaxTokens int
}

// Updates progress bar as tokens stream in
```

## Testing Strategy

**Unit Tests:**
- Theme color calculations
- Component rendering (snapshot tests)
- Focus stack management

**Integration Tests:**
- Full TUI flow with mock agent
- Component interactions
- Theme switching

**Manual Testing:**
- Visual regression testing for each theme
- Keyboard navigation flows
- Performance with large tables

## Success Criteria

✅ **Phase 1 Complete:** All existing UI renders beautifully with Dracula theme, configurable
✅ **Phase 2 Complete:** Tool approval and quick actions use polished huh forms
✅ **Phase 3 Complete:** Can display and interact with tables inline in chat
✅ **Phase 4 Complete:** Professional appearance with gradients, smooth animations, consistent spacing

## Future Enhancements (Phase 5+)

- Custom calendar grid component
- Plugin/MCP status dashboard
- Advanced data visualizations
- Complex form wizards
- Animated state transitions
- Help overlay with keybindings

## References

- [Charmbracelet Bubbles](https://github.com/charmbracelet/bubbles)
- [Charmbracelet Huh](https://github.com/charmbracelet/huh)
- [Charmbracelet Lipgloss](https://github.com/charmbracelet/lipgloss)
- [Dracula Theme](https://draculatheme.com/)
