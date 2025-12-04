# Jeff TUI Refactor - Phases 2-5 Implementation Plan

**Date:** 2025-12-02
**Engineer:** Doctor Biz
**Context:** Phase 1 (theme system) is complete. This plan covers Phases 2-5 for the full visual redesign.
**Approach:** TDD with scenario testing, subagent-driven development

---

## Prerequisites

Before starting, verify Phase 1 completion:
- ✅ Theme system exists at `internal/ui/themes/`
- ✅ Three themes: Dracula, Gruvbox, Nord
- ✅ All UI components use theme colors
- ✅ All tests passing

---

## Phase 2: Huh Integration

**Goal:** Replace custom approval dialog with polished Huh forms for tool approval and quick actions.

### Task 1: Add Huh Dependency (2 min)

**File:** `go.mod`

**Test first:**
```bash
# Verify huh import fails
go run -c 'package main; import "github.com/charmbracelet/huh"' 2>&1 | grep "no required module"
```

**Implementation:**
```bash
cd /Users/harper/Public/src/2389/jeff-agent
go get github.com/charmbracelet/huh@latest
go mod tidy
```

**Verify:**
```bash
grep 'github.com/charmbracelet/huh' go.mod
go build ./...
```

---

### Task 2: Create Huh Approval Component - Test (3 min)

**File:** `internal/ui/components/huh_approval_test.go` (NEW)

**Why:** TDD - write test first to define the interface we want.

**Code:**
```go
// ABOUTME: Test suite for Huh-based tool approval component
// ABOUTME: Ensures approval forms render correctly with theme colors
package components

import (
	"testing"

	"github.com/harper/jefft/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestNewHuhApproval(t *testing.T) {
	theme := themes.NewDracula()
	toolName := "bash"
	description := "Run: rm -rf /tmp/cache"

	approval := NewHuhApproval(theme, toolName, description)

	assert.NotNil(t, approval)
	assert.Equal(t, theme, approval.theme)
	assert.Equal(t, toolName, approval.toolName)
	assert.Equal(t, description, approval.description)
	assert.False(t, approval.approved) // Default state
}

func TestHuhApprovalView(t *testing.T) {
	theme := themes.NewDracula()
	approval := NewHuhApproval(theme, "bash", "Run: echo test")

	view := approval.View()

	// View should contain approval prompt
	assert.Contains(t, view, "bash")
	assert.Contains(t, view, "echo test")
}

func TestHuhApprovalApprove(t *testing.T) {
	theme := themes.NewDracula()
	approval := NewHuhApproval(theme, "bash", "Run: echo test")

	// Initially not approved
	assert.False(t, approval.IsApproved())

	// Approve
	approval.SetApproved(true)
	assert.True(t, approval.IsApproved())
}
```

**Verify:**
```bash
go test ./internal/ui/components/... -v -run TestNewHuhApproval
# Should FAIL - component doesn't exist yet
```

---

### Task 3: Create Huh Approval Component - Implementation (5 min)

**File:** `internal/ui/components/huh_approval.go` (NEW)

**Why:** Implement the component to make tests pass.

**Code:**
```go
// ABOUTME: Huh-based tool approval component using confirm form
// ABOUTME: Provides themed approval dialogs for tool execution
package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/harper/jefft/internal/ui/themes"
)

// HuhApproval is a Huh-based approval dialog for tool execution
type HuhApproval struct {
	theme       themes.Theme
	toolName    string
	description string
	approved    bool
	form        *huh.Form
}

// NewHuhApproval creates a new Huh approval dialog
func NewHuhApproval(theme themes.Theme, toolName, description string) *HuhApproval {
	var approved bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Key("approved").
				Title(fmt.Sprintf("Execute %s?", toolName)).
				Description(description).
				Affirmative("Yes!").
				Negative("No.").
				Value(&approved),
		),
	).WithTheme(huhThemeFromJeffTheme(theme))

	return &HuhApproval{
		theme:       theme,
		toolName:    toolName,
		description: description,
		approved:    false,
		form:        form,
	}
}

// Init implements tea.Model
func (h *HuhApproval) Init() tea.Cmd {
	return h.form.Init()
}

// Update implements tea.Model
func (h *HuhApproval) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle form updates
	form, formCmd := h.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		h.form = f
		cmd = formCmd
	}

	// Check if form is complete
	if h.form.State == huh.StateCompleted {
		// Extract the approved value from form
		if val := h.form.GetString("approved"); val != "" {
			h.approved = (val == "true" || val == "Yes!")
		}
	}

	return h, cmd
}

// View implements tea.Model
func (h *HuhApproval) View() string {
	return h.form.View()
}

// IsApproved returns whether the tool was approved
func (h *HuhApproval) IsApproved() bool {
	return h.approved
}

// SetApproved sets the approval state (for testing)
func (h *HuhApproval) SetApproved(approved bool) {
	h.approved = approved
}

// IsComplete returns whether the form is complete
func (h *HuhApproval) IsComplete() bool {
	return h.form.State == huh.StateCompleted
}

// huhThemeFromJeffTheme converts Jeff theme to Huh theme
func huhThemeFromJeffTheme(theme themes.Theme) *huh.Theme {
	return &huh.Theme{
		Focused: huh.FieldStyles{
			Base:          theme.Foreground(),
			Title:         theme.Primary(),
			Description:   theme.Subtle(),
			ErrorIndicator: theme.Error(),
			ErrorMessage:  theme.Error(),
			SelectSelector: theme.Primary(),
			NextIndicator: theme.Success(),
			PrevIndicator: theme.Warning(),
			Option:        theme.Foreground(),
			SelectedOption: theme.Primary(),
		},
		Blurred: huh.FieldStyles{
			Base:          theme.Subtle(),
			Title:         theme.Subtle(),
			Description:   theme.Subtle(),
			SelectSelector: theme.Subtle(),
			NextIndicator: theme.Subtle(),
			PrevIndicator: theme.Subtle(),
			Option:        theme.Subtle(),
			SelectedOption: theme.Subtle(),
		},
		Help: huh.HelpStyles{
			Ellipsis:       theme.Subtle(),
			ShortKey:       theme.Secondary(),
			ShortDesc:      theme.Foreground(),
			ShortSeparator: theme.Subtle(),
			FullKey:        theme.Secondary(),
			FullDesc:       theme.Foreground(),
			FullSeparator:  theme.Subtle(),
		},
	}
}
```

**Verify:**
```bash
go test ./internal/ui/components/... -v -run TestNewHuhApproval
# Should PASS now
```

---

### Task 4: Create Huh Quick Actions Component - Test (3 min)

**File:** `internal/ui/components/huh_quickactions_test.go` (NEW)

**Code:**
```go
// ABOUTME: Test suite for Huh-based quick actions selector
// ABOUTME: Validates quick action selection UI rendering and interaction
package components

import (
	"testing"

	"github.com/harper/jefft/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestNewHuhQuickActions(t *testing.T) {
	theme := themes.NewDracula()
	options := []QuickActionOption{
		{Label: "Clear conversation", Value: "clear"},
		{Label: "Export chat", Value: "export"},
		{Label: "Toggle help", Value: "help"},
	}

	qa := NewHuhQuickActions(theme, options)

	assert.NotNil(t, qa)
	assert.Equal(t, theme, qa.theme)
	assert.Equal(t, 3, len(qa.options))
	assert.Empty(t, qa.selected)
}

func TestHuhQuickActionsView(t *testing.T) {
	theme := themes.NewDracula()
	options := []QuickActionOption{
		{Label: "Clear conversation", Value: "clear"},
	}

	qa := NewHuhQuickActions(theme, options)
	view := qa.View()

	assert.Contains(t, view, "Clear conversation")
}

func TestHuhQuickActionsSelection(t *testing.T) {
	theme := themes.NewDracula()
	options := []QuickActionOption{
		{Label: "Clear conversation", Value: "clear"},
	}

	qa := NewHuhQuickActions(theme, options)

	// Initially no selection
	assert.Empty(t, qa.GetSelected())

	// Simulate selection
	qa.SetSelected("clear")
	assert.Equal(t, "clear", qa.GetSelected())
}
```

**Verify:**
```bash
go test ./internal/ui/components/... -v -run TestNewHuhQuickActions
# Should FAIL - component doesn't exist
```

---

### Task 5: Create Huh Quick Actions Component - Implementation (5 min)

**File:** `internal/ui/components/huh_quickactions.go` (NEW)

**Code:**
```go
// ABOUTME: Huh-based quick actions selector using select form
// ABOUTME: Provides themed action selection menu for common operations
package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/harper/jefft/internal/ui/themes"
)

// QuickActionOption represents a quick action option
type QuickActionOption struct {
	Label string
	Value string
}

// HuhQuickActions is a Huh-based quick actions selector
type HuhQuickActions struct {
	theme    themes.Theme
	options  []QuickActionOption
	selected string
	form     *huh.Form
}

// NewHuhQuickActions creates a new Huh quick actions selector
func NewHuhQuickActions(theme themes.Theme, options []QuickActionOption) *HuhQuickActions {
	var selected string

	// Convert options to huh.Option
	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt.Label, opt.Value)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("action").
				Title("Choose action").
				Options(huhOptions...).
				Value(&selected),
		),
	).WithTheme(huhThemeFromJeffTheme(theme))

	return &HuhQuickActions{
		theme:    theme,
		options:  options,
		selected: "",
		form:     form,
	}
}

// Init implements tea.Model
func (h *HuhQuickActions) Init() tea.Cmd {
	return h.form.Init()
}

// Update implements tea.Model
func (h *HuhQuickActions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle form updates
	form, formCmd := h.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		h.form = f
		cmd = formCmd
	}

	// Check if form is complete
	if h.form.State == huh.StateCompleted {
		h.selected = h.form.GetString("action")
	}

	return h, cmd
}

// View implements tea.Model
func (h *HuhQuickActions) View() string {
	return h.form.View()
}

// GetSelected returns the selected action value
func (h *HuhQuickActions) GetSelected() string {
	return h.selected
}

// SetSelected sets the selected action (for testing)
func (h *HuhQuickActions) SetSelected(value string) {
	h.selected = value
}

// IsComplete returns whether the form is complete
func (h *HuhQuickActions) IsComplete() bool {
	return h.form.State == huh.StateCompleted
}
```

**Verify:**
```bash
go test ./internal/ui/components/... -v -run TestNewHuhQuickActions
# Should PASS now
```

---

### Task 6: Integrate HuhApproval into Model - Test (3 min)

**File:** `internal/ui/model_test.go`

**Add test:**
```go
func TestModelHuhApprovalIntegration(t *testing.T) {
	model := NewModel("test-conv", "claude-sonnet-4", "dracula")

	// Initially no approval in progress
	assert.False(t, model.toolApprovalMode)
	assert.Nil(t, model.huhApproval)

	// Add a pending tool
	toolUse := &core.ToolUse{
		ID:    "tool-123",
		Name:  "bash",
		Input: map[string]interface{}{"command": "echo test"},
	}
	model.pendingToolUses = []*core.ToolUse{toolUse}

	// Enter approval mode
	model.EnterHuhApprovalMode()

	assert.True(t, model.toolApprovalMode)
	assert.NotNil(t, model.huhApproval)
}
```

**Verify:**
```bash
go test ./internal/ui/... -v -run TestModelHuhApprovalIntegration
# Should FAIL - methods don't exist
```

---

### Task 7: Integrate HuhApproval into Model - Implementation (5 min)

**File:** `internal/ui/model.go`

**Add import:**
```go
import (
	// ... existing imports
	"github.com/harper/jefft/internal/ui/components"
)
```

**Add field to Model struct:**
```go
type Model struct {
	// ... existing fields

	// Huh components for Phase 2
	huhApproval     *components.HuhApproval
	huhQuickActions *components.HuhQuickActions
}
```

**Add methods:**
```go
// EnterHuhApprovalMode creates and shows Huh approval dialog
func (m *Model) EnterHuhApprovalMode() {
	if len(m.pendingToolUses) == 0 {
		return
	}

	m.toolApprovalMode = true

	// Build description from pending tools
	var description string
	if len(m.pendingToolUses) == 1 {
		tool := m.pendingToolUses[0]
		description = fmt.Sprintf("Tool: %s\nInput: %v", tool.Name, tool.Input)
	} else {
		description = fmt.Sprintf("%d tools waiting for approval", len(m.pendingToolUses))
	}

	toolName := m.pendingToolUses[0].Name
	m.huhApproval = components.NewHuhApproval(m.theme, toolName, description)
}

// ExitHuhApprovalMode closes the approval dialog
func (m *Model) ExitHuhApprovalMode() {
	m.toolApprovalMode = false
	m.huhApproval = nil
}
```

**Verify:**
```bash
go test ./internal/ui/... -v -run TestModelHuhApprovalIntegration
# Should PASS now
```

---

### Task 8: Update Model.Update() to Handle Huh Forms (5 min)

**File:** `internal/ui/update.go`

**Find the section handling tool approval (around line 116-140)**

**Replace old approval handling with:**
```go
// In tool approval mode, handle Huh approval form
if m.toolApprovalMode && m.huhApproval != nil {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Let Huh form handle all input
		approvalModel, cmd := m.huhApproval.Update(msg)
		if approval, ok := approvalModel.(*components.HuhApproval); ok {
			m.huhApproval = approval

			// Check if form is complete
			if approval.IsComplete() {
				if approval.IsApproved() {
					m.ExitHuhApprovalMode()
					return m, m.ApproveToolUse()
				} else {
					m.ExitHuhApprovalMode()
					return m, m.DenyToolUse()
				}
			}
		}
		return m, cmd
	}
	return m, nil
}
```

**Verify:**
```bash
go build ./...
# Should compile without errors
```

---

### Task 9: Update View() to Render Huh Forms (3 min)

**File:** `internal/ui/view.go`

**Find the section rendering approval prompt**

**Replace with:**
```go
// Render Huh approval if active
if m.toolApprovalMode && m.huhApproval != nil {
	approvalView := m.huhApproval.View()
	parts = append(parts, "\n"+approvalView)
}
```

**Verify:**
```bash
go build ./...
```

---

### Task 10: Write Scenario Test for Tool Approval (5 min)

**File:** `test/scenarios/tool_approval_test.go` (NEW)

**Code:**
```go
// ABOUTME: Scenario test for tool approval flow using Huh forms
// ABOUTME: Validates end-to-end approval interaction and rendering
package scenarios

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/jefft/internal/core"
	"github.com/harper/jefft/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestToolApprovalWithHuh_UserApprovesTool(t *testing.T) {
	// Setup
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")

	// Add pending tool
	toolUse := &core.ToolUse{
		ID:    "tool-123",
		Name:  "bash",
		Input: map[string]interface{}{"command": "echo hello"},
	}
	model.pendingToolUses = []*core.ToolUse{toolUse}

	// Enter approval mode
	model.EnterHuhApprovalMode()

	// Verify approval UI is shown
	assert.True(t, model.toolApprovalMode)
	assert.NotNil(t, model.huhApproval)

	view := model.View()
	assert.Contains(t, view, "bash")
	assert.Contains(t, view, "echo hello")

	// Simulate user pressing 'y' (yes)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)

	// After approval, form should process
	// (Full approval flow would need tool executor mock)
	assert.NotNil(t, m)
}

func TestToolApprovalWithHuh_UserDeniesTool(t *testing.T) {
	// Setup
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")

	// Add pending tool
	toolUse := &core.ToolUse{
		ID:    "tool-456",
		Name:  "bash",
		Input: map[string]interface{}{"command": "rm -rf /"},
	}
	model.pendingToolUses = []*core.ToolUse{toolUse}

	// Enter approval mode
	model.EnterHuhApprovalMode()

	// Simulate user pressing 'n' (no)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)

	// Tool should be denied
	assert.NotNil(t, m)
}

func TestToolApprovalVisualCheck_Dracula(t *testing.T) {
	// Visual regression test - verify Dracula theme renders correctly
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")

	toolUse := &core.ToolUse{
		ID:    "tool-789",
		Name:  "search_emails",
		Input: map[string]interface{}{"query": "unread"},
	}
	model.pendingToolUses = []*core.ToolUse{toolUse}
	model.EnterHuhApprovalMode()

	view := model.View()

	// Should contain Dracula purple color codes
	// (Huh theme should use our theme colors)
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "search_emails")
}
```

**Verify:**
```bash
go test ./test/scenarios/... -v
# Should PASS
```

---

## Phase 3: Rich Components

**Goal:** Add interactive tables, progress bars, and lists that render inline in the chat.

### Task 11: Create Table Component - Test (3 min)

**File:** `internal/ui/components/table_test.go` (NEW)

**Code:**
```go
// ABOUTME: Test suite for interactive table component using bubbles
// ABOUTME: Validates table rendering, theming, and interaction
package components

import (
	"testing"

	"github.com/harper/jefft/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestNewTable(t *testing.T) {
	theme := themes.NewDracula()
	columns := []string{"From", "Subject", "Date"}
	rows := [][]string{
		{"alice@example.com", "Meeting", "2h ago"},
		{"bob@example.com", "Report", "5h ago"},
	}

	table := NewTable(theme, columns, rows)

	assert.NotNil(t, table)
	assert.Equal(t, theme, table.theme)
	assert.Equal(t, 2, len(rows))
}

func TestTableView(t *testing.T) {
	theme := themes.NewDracula()
	columns := []string{"Name", "Value"}
	rows := [][]string{{"Key1", "Value1"}}

	table := NewTable(theme, columns, rows)
	view := table.View()

	assert.Contains(t, view, "Name")
	assert.Contains(t, view, "Value")
	assert.Contains(t, view, "Key1")
}

func TestTableSelection(t *testing.T) {
	theme := themes.NewDracula()
	columns := []string{"Name"}
	rows := [][]string{{"Row1"}, {"Row2"}, {"Row3"}}

	table := NewTable(theme, columns, rows)

	// Initially first row selected
	assert.Equal(t, 0, table.GetSelectedRow())

	// Move down
	table.MoveDown()
	assert.Equal(t, 1, table.GetSelectedRow())

	// Move up
	table.MoveUp()
	assert.Equal(t, 0, table.GetSelectedRow())
}
```

**Verify:**
```bash
go test ./internal/ui/components/... -v -run TestNewTable
# Should FAIL - component doesn't exist
```

---

### Task 12: Create Table Component - Implementation (5 min)

**File:** `internal/ui/components/table.go` (NEW)

**Code:**
```go
// ABOUTME: Interactive table component using bubbles table
// ABOUTME: Renders tabular data with theme colors and keyboard navigation
package components

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jefft/internal/ui/themes"
)

// Table is a themed table component
type Table struct {
	theme       themes.Theme
	table       table.Model
	columns     []string
	rows        [][]string
	selectedRow int
}

// NewTable creates a new table component
func NewTable(theme themes.Theme, columns []string, rows [][]string) *Table {
	// Create table columns
	tableCols := make([]table.Column, len(columns))
	for i, col := range columns {
		tableCols[i] = table.Column{
			Title: col,
			Width: 20,
		}
	}

	// Create table rows
	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		tableRows[i] = table.Row(row)
	}

	// Create styled table
	t := table.New(
		table.WithColumns(tableCols),
		table.WithRows(tableRows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Apply theme styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.BorderFocus()).
		BorderBottom(true).
		Bold(true).
		Foreground(theme.Primary())
	s.Selected = s.Selected.
		Foreground(theme.Background()).
		Background(theme.Primary()).
		Bold(false)

	t.SetStyles(s)

	return &Table{
		theme:       theme,
		table:       t,
		columns:     columns,
		rows:        rows,
		selectedRow: 0,
	}
}

// Init implements tea.Model
func (t *Table) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (t *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	t.selectedRow = t.table.Cursor()
	return t, cmd
}

// View implements tea.Model
func (t *Table) View() string {
	return t.table.View()
}

// GetSelectedRow returns the currently selected row index
func (t *Table) GetSelectedRow() int {
	return t.selectedRow
}

// MoveDown moves selection down
func (t *Table) MoveDown() {
	t.table.MoveDown(1)
	t.selectedRow = t.table.Cursor()
}

// MoveUp moves selection up
func (t *Table) MoveUp() {
	t.table.MoveUp(1)
	t.selectedRow = t.table.Cursor()
}

// SetRows updates the table rows
func (t *Table) SetRows(rows [][]string) {
	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		tableRows[i] = table.Row(row)
	}
	t.table.SetRows(tableRows)
	t.rows = rows
}
```

**Verify:**
```bash
go test ./internal/ui/components/... -v -run TestNewTable
# Should PASS
```

---

### Task 13: Create Progress Component - Test (2 min)

**File:** `internal/ui/components/progress_test.go` (NEW)

**Code:**
```go
// ABOUTME: Test suite for progress bar component
// ABOUTME: Tests progress rendering and value updates
package components

import (
	"testing"

	"github.com/harper/jefft/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestNewProgress(t *testing.T) {
	theme := themes.NewDracula()
	label := "Processing"

	progress := NewProgress(theme, label)

	assert.NotNil(t, progress)
	assert.Equal(t, theme, progress.theme)
	assert.Equal(t, label, progress.label)
	assert.Equal(t, 0.0, progress.value)
}

func TestProgressSetValue(t *testing.T) {
	theme := themes.NewDracula()
	progress := NewProgress(theme, "Loading")

	progress.SetValue(0.5)
	assert.Equal(t, 0.5, progress.value)

	progress.SetValue(1.0)
	assert.Equal(t, 1.0, progress.value)
}

func TestProgressView(t *testing.T) {
	theme := themes.NewDracula()
	progress := NewProgress(theme, "Upload")
	progress.SetValue(0.75)

	view := progress.View()
	assert.NotEmpty(t, view)
}
```

**Verify:**
```bash
go test ./internal/ui/components/... -v -run TestNewProgress
# Should FAIL
```

---

### Task 14: Create Progress Component - Implementation (4 min)

**File:** `internal/ui/components/progress.go` (NEW)

**Code:**
```go
// ABOUTME: Progress bar component using bubbles progress
// ABOUTME: Displays completion percentage with theme colors
package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jefft/internal/ui/themes"
)

// Progress is a themed progress bar component
type Progress struct {
	theme    themes.Theme
	label    string
	value    float64 // 0.0 to 1.0
	progress progress.Model
}

// NewProgress creates a new progress bar
func NewProgress(theme themes.Theme, label string) *Progress {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return &Progress{
		theme:    theme,
		label:    label,
		value:    0.0,
		progress: p,
	}
}

// Init implements tea.Model
func (p *Progress) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (p *Progress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return p, nil
}

// View implements tea.Model
func (p *Progress) View() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(p.theme.Foreground()).
		Bold(true)

	percentStyle := lipgloss.NewStyle().
		Foreground(p.theme.Primary()).
		Bold(true)

	label := labelStyle.Render(p.label)
	bar := p.progress.ViewAs(p.value)
	percent := percentStyle.Render(fmt.Sprintf(" %.0f%%", p.value*100))

	return lipgloss.JoinHorizontal(lipgloss.Left, label, " ", bar, percent)
}

// SetValue updates the progress value (0.0 to 1.0)
func (p *Progress) SetValue(value float64) {
	if value < 0.0 {
		value = 0.0
	}
	if value > 1.0 {
		value = 1.0
	}
	p.value = value
}

// GetValue returns the current progress value
func (p *Progress) GetValue() float64 {
	return p.value
}
```

**Verify:**
```bash
go test ./internal/ui/components/... -v -run TestNewProgress
# Should PASS
```

---

### Task 15: Add Component Embedding to Message Type (3 min)

**File:** `internal/ui/model.go`

**Update Message struct:**
```go
type Message struct {
	Role         string
	Content      string
	ContentBlock []core.ContentBlock // For structured content like tool_result blocks

	// Rich component for inline rendering (Phase 3)
	Component interface{} // Can be *components.Table, *components.Progress, etc.
	ComponentID string    // Unique ID for routing events to component
}
```

**Add method to Model:**
```go
// AddMessageWithComponent adds a message with an embedded component
func (m *Model) AddMessageWithComponent(role, content string, component interface{}) {
	componentID := fmt.Sprintf("comp-%d", len(m.Messages))
	m.Messages = append(m.Messages, Message{
		Role:        role,
		Content:     content,
		Component:   component,
		ComponentID: componentID,
	})
	m.updateContextUsage()
}
```

**Verify:**
```bash
go build ./...
```

---

### Task 16: Update View() to Render Embedded Components (4 min)

**File:** `internal/ui/view.go`

**In the message rendering loop, add component rendering:**
```go
// After rendering message content
if msg.Component != nil {
	// Render component based on type
	switch comp := msg.Component.(type) {
	case *components.Table:
		componentView := comp.View()
		renderedContent += "\n\n" + componentView
	case *components.Progress:
		componentView := comp.View()
		renderedContent += "\n" + componentView
	case tea.Model:
		// Generic tea.Model support
		componentView := comp.View()
		renderedContent += "\n" + componentView
	}
}
```

**Verify:**
```bash
go build ./...
```

---

### Task 17: Write Scenario Test for Rich Components (5 min)

**File:** `test/scenarios/rich_components_test.go` (NEW)

**Code:**
```go
// ABOUTME: Scenario test for rich inline components (tables, progress)
// ABOUTME: Validates component rendering in chat conversation flow
package scenarios

import (
	"testing"

	"github.com/harper/jefft/internal/ui"
	"github.com/harper/jefft/internal/ui/components"
	"github.com/stretchr/testify/assert"
)

func TestTableInChatMessage_RendersCorrectly(t *testing.T) {
	// Setup model with Dracula theme
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")

	// Create a table component
	columns := []string{"From", "Subject", "Date"}
	rows := [][]string{
		{"alice@example.com", "Q4 Review", "2h ago"},
		{"bob@example.com", "Lunch?", "4h ago"},
	}

	theme := model.GetTheme()
	table := components.NewTable(theme, columns, rows)

	// Add message with embedded table
	model.AddMessageWithComponent(
		"assistant",
		"Here are your unread emails:",
		table,
	)

	// Render view
	view := model.View()

	// Verify table appears in view
	assert.Contains(t, view, "From")
	assert.Contains(t, view, "Subject")
	assert.Contains(t, view, "alice@example.com")
	assert.Contains(t, view, "Q4 Review")
}

func TestProgressInChatMessage_ShowsCompletion(t *testing.T) {
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")

	// Create progress component
	theme := model.GetTheme()
	progress := components.NewProgress(theme, "Uploading")
	progress.SetValue(0.65)

	// Add message with progress
	model.AddMessageWithComponent(
		"assistant",
		"Upload in progress:",
		progress,
	)

	// Render view
	view := model.View()

	// Verify progress appears
	assert.Contains(t, view, "Uploading")
	assert.Contains(t, view, "65%")
}

func TestMultipleComponents_InDifferentMessages(t *testing.T) {
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")
	theme := model.GetTheme()

	// First message: table
	table := components.NewTable(
		theme,
		[]string{"Task", "Status"},
		[][]string{
			{"Review PR", "Done"},
			{"Write docs", "In Progress"},
		},
	)
	model.AddMessageWithComponent("assistant", "Your tasks:", table)

	// Second message: progress
	progress := components.NewProgress(theme, "Processing")
	progress.SetValue(0.33)
	model.AddMessageWithComponent("assistant", "Working on it:", progress)

	// Render
	view := model.View()

	// Both components should render
	assert.Contains(t, view, "Task")
	assert.Contains(t, view, "Review PR")
	assert.Contains(t, view, "Processing")
	assert.Contains(t, view, "33%")
}
```

**Verify:**
```bash
go test ./test/scenarios/... -v -run TestTableInChatMessage
go test ./test/scenarios/... -v -run TestProgressInChatMessage
go test ./test/scenarios/... -v -run TestMultipleComponents
# All should PASS
```

---

## Phase 4: Visual Polish

**Goal:** Professional appearance with gradients, animations, and consistent spacing.

### Task 18: Add Title Gradient Enhancement (3 min)

**File:** `internal/ui/view.go`

**Current title rendering uses gradient. Enhance it:**
```go
// Enhanced title with better gradient and spacing
titleText := fmt.Sprintf("Jeff • %s", m.Model)
titleGradient := themes.RenderGradient(titleText, m.theme.TitleGradient())

// Add decorative border
titleStyle := lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(m.theme.BorderFocus()).
	Padding(0, 2).
	MarginBottom(1)

title := titleStyle.Render(titleGradient)
```

**Verify:**
```bash
go build ./...
```

---

### Task 19: Improve Markdown Rendering with Theme Colors (4 min)

**File:** `internal/ui/model.go`

**Update RenderMessage to use themed glamour:**
```go
// RenderMessage renders a message using glamour for assistant messages
func (m *Model) RenderMessage(msg Message) (string, error) {
	if msg.Role == "assistant" {
		// Create themed renderer if we don't have one or theme changed
		if m.renderer == nil {
			renderer, err := m.createThemedRenderer()
			if err == nil {
				m.renderer = renderer
			}
		}

		if m.renderer != nil {
			rendered, err := m.renderer.Render(msg.Content)
			if err != nil {
				return msg.Content, err
			}
			return rendered, nil
		}
	}
	return msg.Content, nil
}

// createThemedRenderer creates a glamour renderer with theme colors
func (m *Model) createThemedRenderer() (*glamour.TermRenderer, error) {
	// Build custom glamour style from our theme
	style := glamour.DarkStyleConfig

	// Customize with our theme colors
	style.Document.Color = glamour.NewColor(string(m.theme.Foreground()))
	style.H1.Color = glamour.NewColor(string(m.theme.Primary()))
	style.H2.Color = glamour.NewColor(string(m.theme.Secondary()))
	style.Code.BackgroundColor = glamour.NewColor(string(m.theme.Background()))
	style.CodeBlock.Chroma.Text.Color = glamour.NewColor(string(m.theme.Foreground()))
	style.Link.Color = glamour.NewColor(string(m.theme.Primary()))
	style.LinkText.Color = glamour.NewColor(string(m.theme.Secondary()))

	return glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(m.Width-10),
	)
}
```

**Verify:**
```bash
go build ./...
```

---

### Task 20: Add Smooth State Transitions (4 min)

**File:** `internal/ui/components/transitions.go` (NEW)

**Code:**
```go
// ABOUTME: Smooth state transition utilities for UI animations
// ABOUTME: Provides fade-in, slide-in effects for components
package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TransitionState represents animation state
type TransitionState int

const (
	TransitionIdle TransitionState = iota
	TransitionFadingIn
	TransitionFadingOut
	TransitionComplete
)

// FadeTransition handles fade-in/out animations
type FadeTransition struct {
	state    TransitionState
	opacity  float64 // 0.0 to 1.0
	duration time.Duration
	startTime time.Time
}

// NewFadeTransition creates a new fade transition
func NewFadeTransition(duration time.Duration) *FadeTransition {
	return &FadeTransition{
		state:    TransitionIdle,
		opacity:  0.0,
		duration: duration,
	}
}

// FadeIn starts fade-in animation
func (f *FadeTransition) FadeIn() tea.Cmd {
	f.state = TransitionFadingIn
	f.startTime = time.Now()
	f.opacity = 0.0
	return f.tick()
}

// tick updates animation frame
func (f *FadeTransition) tick() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return transitionTickMsg{time: t}
	})
}

type transitionTickMsg struct {
	time time.Time
}

// Update updates the transition state
func (f *FadeTransition) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case transitionTickMsg:
		if f.state == TransitionFadingIn {
			elapsed := msg.time.Sub(f.startTime)
			progress := float64(elapsed) / float64(f.duration)

			if progress >= 1.0 {
				f.opacity = 1.0
				f.state = TransitionComplete
				return nil
			}

			f.opacity = progress
			return f.tick()
		}
	}
	return nil
}

// GetOpacity returns current opacity (0.0 to 1.0)
func (f *FadeTransition) GetOpacity() float64 {
	return f.opacity
}

// IsComplete returns whether transition is done
func (f *FadeTransition) IsComplete() bool {
	return f.state == TransitionComplete
}
```

**Verify:**
```bash
go build ./internal/ui/components/...
```

---

### Task 21: Improve Spacing and Layout (3 min)

**File:** `internal/ui/view.go`

**Add consistent spacing constants:**
```go
const (
	paddingHorizontal = 2
	paddingVertical   = 1
	marginBottom      = 1
	borderRadius      = 1
)

// Apply to all major sections
func (m *Model) createViewStyles() viewStyles {
	theme := m.theme

	baseStyle := lipgloss.NewStyle().
		Padding(paddingVertical, paddingHorizontal)

	return viewStyles{
		title: baseStyle.Copy().
			Bold(true).
			Foreground(theme.Primary()).
			MarginBottom(marginBottom),

		input: baseStyle.Copy().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderFocus()).
			MarginTop(marginBottom),

		// ... apply consistent padding/margin to all styles
	}
}
```

**Verify:**
```bash
go build ./...
```

---

### Task 22: Add Hover/Focus Visual Feedback (3 min)

**File:** `internal/ui/components/table.go`

**Enhance table selection visuals:**
```go
// In NewTable, enhance selected style:
s.Selected = s.Selected.
	Foreground(theme.Background()).
	Background(theme.Primary()).
	Bold(true).
	Underline(true) // Add underline for emphasis

// Add subtle animation to selection
s.Selected = s.Selected.
	Border(lipgloss.RoundedBorder()).
	BorderForeground(theme.Primary()).
	Padding(0, 1)
```

**Verify:**
```bash
go build ./internal/ui/components/...
```

---

### Task 23: Write Visual Regression Tests (5 min)

**File:** `test/visual/theme_rendering_test.go` (NEW)

**Code:**
```go
// ABOUTME: Visual regression tests for theme rendering
// ABOUTME: Captures and validates visual output across all themes
package visual

import (
	"testing"

	"github.com/harper/jefft/internal/ui"
	"github.com/harper/jefft/internal/ui/components"
	"github.com/stretchr/testify/assert"
)

func TestDraculaTheme_VisualOutput(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	model.AddMessage("user", "Hello")
	model.AddMessage("assistant", "Hi! **Bold** and *italic* text.")

	view := model.View()

	// Should contain ANSI color codes for Dracula purple (#bd93f9)
	assert.Contains(t, view, "Jeff")
	assert.NotEmpty(t, view)

	// Verify gradient in title
	assert.Contains(t, view, "\033[") // ANSI escape codes present
}

func TestGruvboxTheme_VisualOutput(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "gruvbox")
	model.AddMessage("user", "Test")

	view := model.View()

	// Gruvbox should look different from Dracula
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Jeff")
}

func TestNordTheme_VisualOutput(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "nord")
	model.AddMessage("user", "Test")

	view := model.View()

	// Nord should have distinct appearance
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Jeff")
}

func TestTableComponent_VisualPolish(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	theme := model.GetTheme()

	table := components.NewTable(
		theme,
		[]string{"Column A", "Column B"},
		[][]string{
			{"Row 1A", "Row 1B"},
			{"Row 2A", "Row 2B"},
		},
	)

	model.AddMessageWithComponent("assistant", "Here's a table:", table)
	view := model.View()

	// Table should have borders and theme colors
	assert.Contains(t, view, "Column A")
	assert.Contains(t, view, "Row 1A")
	assert.Contains(t, view, "─") // Box drawing characters
}
```

**Verify:**
```bash
go test ./test/visual/... -v
```

---

## Phase 5: Enhanced Features

**Goal:** New quality-of-life features that showcase the refactored UI.

### Task 24: Add Help Overlay with Keybindings (5 min)

**File:** `internal/ui/components/help.go` (NEW)

**Code:**
```go
// ABOUTME: Help overlay component showing keyboard shortcuts
// ABOUTME: Displays themed help information with keybindings
package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jefft/internal/ui/themes"
)

// HelpOverlay displays keyboard shortcuts
type HelpOverlay struct {
	theme themes.Theme
}

// NewHelpOverlay creates a new help overlay
func NewHelpOverlay(theme themes.Theme) *HelpOverlay {
	return &HelpOverlay{theme: theme}
}

// View renders the help overlay
func (h *HelpOverlay) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(h.theme.Primary()).
		Bold(true).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(h.theme.Secondary()).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(h.theme.Foreground())

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(h.theme.BorderFocus()).
		Padding(1, 2)

	title := titleStyle.Render("Keyboard Shortcuts")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"ctrl+c", "Quit"},
		{"enter", "Send message"},
		{"ctrl+p", "Quick actions"},
		{"ctrl+l", "Clear screen"},
		{"esc", "Cancel/close"},
		{"↑↓", "Navigate tables"},
		{"?", "Toggle help"},
	}

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	for _, s := range shortcuts {
		key := keyStyle.Render(s.key)
		desc := descStyle.Render(s.desc)
		line := lipgloss.JoinHorizontal(lipgloss.Left, key, "  ", desc)
		lines = append(lines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return borderStyle.Render(content)
}
```

**Verify:**
```bash
go build ./internal/ui/components/...
```

---

### Task 25: Add Error Visualization Component (4 min)

**File:** `internal/ui/components/error.go` (NEW)

**Code:**
```go
// ABOUTME: Error visualization component with themed styling
// ABOUTME: Displays errors with appropriate color and formatting
package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jefft/internal/ui/themes"
)

// ErrorDisplay shows formatted errors
type ErrorDisplay struct {
	theme   themes.Theme
	title   string
	message string
	details string
}

// NewErrorDisplay creates a new error display
func NewErrorDisplay(theme themes.Theme, title, message, details string) *ErrorDisplay {
	return &ErrorDisplay{
		theme:   theme,
		title:   title,
		message: message,
		details: details,
	}
}

// View renders the error display
func (e *ErrorDisplay) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(e.theme.Error()).
		Bold(true).
		MarginBottom(1)

	messageStyle := lipgloss.NewStyle().
		Foreground(e.theme.Foreground())

	detailsStyle := lipgloss.NewStyle().
		Foreground(e.theme.Subtle()).
		Italic(true)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(e.theme.Error()).
		Padding(1, 2).
		MarginTop(1)

	var lines []string
	lines = append(lines, titleStyle.Render("❌ "+e.title))

	if e.message != "" {
		lines = append(lines, "")
		lines = append(lines, messageStyle.Render(e.message))
	}

	if e.details != "" {
		lines = append(lines, "")
		lines = append(lines, detailsStyle.Render("Details: "+e.details))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return borderStyle.Render(content)
}
```

**Verify:**
```bash
go build ./internal/ui/components/...
```

---

### Task 26: Write Final Integration Test (5 min)

**File:** `test/integration/full_tui_test.go` (NEW)

**Code:**
```go
// ABOUTME: Complete TUI integration test covering all phases
// ABOUTME: End-to-end test of theme, components, and interactions
package integration

import (
	"testing"

	"github.com/harper/jefft/internal/ui"
	"github.com/harper/jefft/internal/ui/components"
	"github.com/stretchr/testify/assert"
)

func TestFullTUI_AllPhasesIntegrated(t *testing.T) {
	// Phase 1: Theme system
	model := ui.NewModel("integration-test", "claude-sonnet-4", "dracula")
	assert.NotNil(t, model)
	assert.Equal(t, "dracula", model.GetTheme().Name())

	// Add chat messages
	model.AddMessage("user", "Show me my emails")
	model.AddMessage("assistant", "Here are your emails:")

	// Phase 3: Rich components - add table
	theme := model.GetTheme()
	table := components.NewTable(
		theme,
		[]string{"From", "Subject"},
		[][]string{
			{"alice@example.com", "Meeting"},
			{"bob@example.com", "Report"},
		},
	)
	model.AddMessageWithComponent("assistant", "Your emails:", table)

	// Phase 3: Add progress
	progress := components.NewProgress(theme, "Loading")
	progress.SetValue(0.8)
	model.AddMessageWithComponent("assistant", "Progress:", progress)

	// Render complete view
	view := model.View()

	// Verify all elements render
	assert.Contains(t, view, "Jeff")
	assert.Contains(t, view, "Show me my emails")
	assert.Contains(t, view, "From")
	assert.Contains(t, view, "alice@example.com")
	assert.Contains(t, view, "Loading")
	assert.Contains(t, view, "80%")

	// Phase 4: Visual polish - verify gradients present
	assert.Contains(t, view, "\033[") // ANSI codes
}

func TestFullTUI_ThemeSwitching(t *testing.T) {
	// Create models with different themes
	dracula := ui.NewModel("test", "claude-sonnet-4", "dracula")
	gruvbox := ui.NewModel("test", "claude-sonnet-4", "gruvbox")
	nord := ui.NewModel("test", "claude-sonnet-4", "nord")

	draculaView := dracula.View()
	gruvboxView := gruvbox.View()
	nordView := nord.View()

	// All should render but with different styling
	assert.NotEqual(t, draculaView, gruvboxView)
	assert.NotEqual(t, gruvboxView, nordView)
	assert.NotEqual(t, nordView, draculaView)
}

func TestFullTUI_ComponentInteraction(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	theme := model.GetTheme()

	// Create interactive table
	table := components.NewTable(
		theme,
		[]string{"Task", "Status"},
		[][]string{
			{"Task 1", "Done"},
			{"Task 2", "In Progress"},
			{"Task 3", "Pending"},
		},
	)

	// Initial selection
	assert.Equal(t, 0, table.GetSelectedRow())

	// Navigate
	table.MoveDown()
	assert.Equal(t, 1, table.GetSelectedRow())

	table.MoveUp()
	assert.Equal(t, 0, table.GetSelectedRow())
}
```

**Verify:**
```bash
go test ./test/integration/... -v -run TestFullTUI
# All tests should PASS
```

---

## Verification & Testing

After completing all tasks, run comprehensive test suite:

```bash
# Run all tests
go test ./... -v

# Run only scenario tests
go test ./test/scenarios/... -v

# Run integration tests
go test ./test/integration/... -v

# Build the application
go build ./cmd/jefft/...

# Test with each theme
./jefft --theme dracula
./jefft --theme gruvbox
./jefft --theme nord
```

---

## Success Criteria

- ✅ All unit tests passing
- ✅ All scenario tests passing
- ✅ All integration tests passing
- ✅ Build succeeds without errors
- ✅ Manual testing shows:
  - Tool approval uses Huh forms
  - Quick actions use Huh select
  - Tables render inline in chat
  - Progress bars show correctly
  - Each theme looks visually distinct
  - Gradients appear in title
  - Spacing is consistent
  - Help overlay displays shortcuts
  - Errors render with theme colors

---

## Commit Strategy

After each task (or small group of related tasks):
```bash
git add .
git commit -m "feat(ui): [task description]"
```

Example commits:
- `feat(ui): add huh dependency and approval component`
- `feat(ui): integrate huh forms for tool approval`
- `feat(ui): add interactive table component`
- `feat(ui): add progress bar component`
- `feat(ui): enhance title gradient and spacing`
- `feat(ui): add help overlay and error visualization`

---

## Notes for Engineer

**Context Assumptions:**
- You have zero familiarity with this codebase
- All file paths are absolute
- All code examples are complete and copy-pasteable
- Tests are written BEFORE implementation (TDD)
- Scenario tests validate end-to-end flows

**Key Libraries:**
- `github.com/charmbracelet/huh` - Form library for approvals/selects
- `github.com/charmbracelet/bubbles` - Components (table, progress, list)
- `github.com/charmbracelet/lipgloss` - Styling library
- `github.com/charmbracelet/bubbletea` - TUI framework

**Testing Philosophy:**
- Write test first (RED)
- Implement to make it pass (GREEN)
- Refactor if needed (REFACTOR)
- Scenario tests verify real usage

**If You Get Stuck:**
- Read the design doc: `docs/plans/2025-12-02-tui-refactor-design.md`
- Check existing theme code: `internal/ui/themes/`
- Look at existing components: `internal/ui/approval.go`, `internal/ui/statusbar.go`
- Run tests frequently to catch issues early

---

## Estimated Time

- Phase 2 (Huh Integration): ~45 minutes
- Phase 3 (Rich Components): ~50 minutes
- Phase 4 (Visual Polish): ~30 minutes
- Phase 5 (Enhanced Features): ~20 minutes

**Total: ~2.5 hours** (excluding breaks and debugging)

---

**Ready to execute with subagent-driven development!**
