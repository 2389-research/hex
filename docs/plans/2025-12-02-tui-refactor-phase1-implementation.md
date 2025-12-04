# TUI Refactor Phase 1: Foundation & Theming - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add huh dependency, create Dracula/Gruvbox/Nord themes, refactor existing UI to use theme system with gradient support

**Architecture:** Create a theme system with Theme interface and three implementations. Refactor all existing lipgloss styles in model.go, view.go, statusbar.go to use theme colors. Add gradient rendering for headers.

**Tech Stack:**
- github.com/charmbracelet/huh (new)
- github.com/charmbracelet/lipgloss (existing)
- github.com/charmbracelet/bubbles (existing)

---

## Task 1: Add Dependencies & Create Theme Foundation

**Files:**
- Modify: `go.mod`
- Create: `internal/ui/themes/theme.go`
- Create: `internal/ui/themes/theme_test.go`

**Step 1: Add huh dependency**

```bash
go get github.com/charmbracelet/huh@latest
```

Expected: go.mod updated with huh dependency

**Step 2: Write test for Theme interface**

Create `internal/ui/themes/theme_test.go`:

```go
package themes_test

import (
	"testing"

	"github.com/harper/jefft/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestThemeInterface(t *testing.T) {
	// Test that we can get a theme by name
	theme := themes.GetTheme("dracula")
	assert.NotNil(t, theme)
	assert.Equal(t, "Dracula", theme.Name())
}

func TestDefaultTheme(t *testing.T) {
	// Unknown theme should return Dracula as default
	theme := themes.GetTheme("unknown")
	assert.NotNil(t, theme)
	assert.Equal(t, "Dracula", theme.Name())
}

func TestAllThemes(t *testing.T) {
	// Should have exactly 3 themes
	all := themes.AllThemes()
	assert.Len(t, all, 3)

	names := make(map[string]bool)
	for _, theme := range all {
		names[theme.Name()] = true
	}

	assert.True(t, names["Dracula"])
	assert.True(t, names["Gruvbox Dark"])
	assert.True(t, names["Nord"])
}
```

**Step 3: Run test to verify it fails**

```bash
go test ./internal/ui/themes -v
```

Expected: FAIL - package does not exist

**Step 4: Create Theme interface**

Create `internal/ui/themes/theme.go`:

```go
// Package themes provides color schemes for the Jeff TUI.
// ABOUTME: Theme system with Dracula, Gruvbox, Nord color schemes
// ABOUTME: Provides semantic colors and gradient support for UI elements
package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the interface for UI color schemes
type Theme interface {
	Name() string

	// Base colors
	Background() lipgloss.Color
	Foreground() lipgloss.Color

	// Semantic colors
	Primary() lipgloss.Color
	Secondary() lipgloss.Color
	Success() lipgloss.Color
	Warning() lipgloss.Color
	Error() lipgloss.Color

	// UI element colors
	Border() lipgloss.Color
	BorderFocus() lipgloss.Color
	Subtle() lipgloss.Color

	// Gradients (returns array of colors for gradient rendering)
	TitleGradient() []lipgloss.Color
}

// GetTheme returns a theme by name, defaulting to Dracula
func GetTheme(name string) Theme {
	switch name {
	case "gruvbox":
		return NewGruvboxTheme()
	case "nord":
		return NewNordTheme()
	default:
		return NewDraculaTheme()
	}
}

// AllThemes returns all available themes
func AllThemes() []Theme {
	return []Theme{
		NewDraculaTheme(),
		NewGruvboxTheme(),
		NewNordTheme(),
	}
}
```

**Step 5: Run test to verify it still fails**

```bash
go test ./internal/ui/themes -v
```

Expected: FAIL - NewDraculaTheme, NewGruvboxTheme, NewNordTheme undefined

**Step 6: Commit foundation**

```bash
git add go.mod go.sum internal/ui/themes/
git commit -m "feat(ui): add theme system foundation and huh dependency"
```

---

## Task 2: Implement Dracula Theme

**Files:**
- Create: `internal/ui/themes/dracula.go`
- Create: `internal/ui/themes/dracula_test.go`

**Step 1: Write Dracula theme tests**

Create `internal/ui/themes/dracula_test.go`:

```go
package themes_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jefft/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestDraculaTheme(t *testing.T) {
	theme := themes.NewDraculaTheme()

	assert.Equal(t, "Dracula", theme.Name())
	assert.Equal(t, lipgloss.Color("#282a36"), theme.Background())
	assert.Equal(t, lipgloss.Color("#f8f8f2"), theme.Foreground())
	assert.Equal(t, lipgloss.Color("#bd93f9"), theme.Primary())
	assert.Equal(t, lipgloss.Color("#ff79c6"), theme.Secondary())
	assert.Equal(t, lipgloss.Color("#50fa7b"), theme.Success())
	assert.Equal(t, lipgloss.Color("#f1fa8c"), theme.Warning())
	assert.Equal(t, lipgloss.Color("#ff5555"), theme.Error())
}

func TestDraculaGradient(t *testing.T) {
	theme := themes.NewDraculaTheme()
	gradient := theme.TitleGradient()

	assert.Len(t, gradient, 2)
	assert.Equal(t, lipgloss.Color("#bd93f9"), gradient[0]) // Purple
	assert.Equal(t, lipgloss.Color("#ff79c6"), gradient[1]) // Pink
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/themes -v -run Dracula
```

Expected: FAIL - NewDraculaTheme undefined

**Step 3: Implement Dracula theme**

Create `internal/ui/themes/dracula.go`:

```go
package themes

import "github.com/charmbracelet/lipgloss"

// DraculaTheme implements the Dracula color scheme
type DraculaTheme struct{}

// NewDraculaTheme creates a new Dracula theme
func NewDraculaTheme() Theme {
	return &DraculaTheme{}
}

func (t *DraculaTheme) Name() string {
	return "Dracula"
}

func (t *DraculaTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282a36")
}

func (t *DraculaTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#f8f8f2")
}

func (t *DraculaTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#bd93f9") // Purple
}

func (t *DraculaTheme) Secondary() lipgloss.Color {
	return lipgloss.Color("#ff79c6") // Pink
}

func (t *DraculaTheme) Success() lipgloss.Color {
	return lipgloss.Color("#50fa7b") // Green
}

func (t *DraculaTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#f1fa8c") // Yellow
}

func (t *DraculaTheme) Error() lipgloss.Color {
	return lipgloss.Color("#ff5555") // Red
}

func (t *DraculaTheme) Border() lipgloss.Color {
	return lipgloss.Color("#6272a4") // Comment gray
}

func (t *DraculaTheme) BorderFocus() lipgloss.Color {
	return lipgloss.Color("#bd93f9") // Purple
}

func (t *DraculaTheme) Subtle() lipgloss.Color {
	return lipgloss.Color("#6272a4") // Comment gray
}

func (t *DraculaTheme) TitleGradient() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color("#bd93f9"), // Purple
		lipgloss.Color("#ff79c6"), // Pink
	}
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/ui/themes -v -run Dracula
```

Expected: PASS

**Step 5: Commit Dracula theme**

```bash
git add internal/ui/themes/dracula.go internal/ui/themes/dracula_test.go
git commit -m "feat(ui): implement Dracula theme"
```

---

## Task 3: Implement Gruvbox Theme

**Files:**
- Create: `internal/ui/themes/gruvbox.go`
- Create: `internal/ui/themes/gruvbox_test.go`

**Step 1: Write Gruvbox theme tests**

Create `internal/ui/themes/gruvbox_test.go`:

```go
package themes_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jefft/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestGruvboxTheme(t *testing.T) {
	theme := themes.NewGruvboxTheme()

	assert.Equal(t, "Gruvbox Dark", theme.Name())
	assert.Equal(t, lipgloss.Color("#282828"), theme.Background())
	assert.Equal(t, lipgloss.Color("#ebdbb2"), theme.Foreground())
	assert.Equal(t, lipgloss.Color("#d79921"), theme.Primary())
	assert.Equal(t, lipgloss.Color("#b16286"), theme.Secondary())
}

func TestGruvboxGradient(t *testing.T) {
	theme := themes.NewGruvboxTheme()
	gradient := theme.TitleGradient()

	assert.Len(t, gradient, 2)
	assert.Equal(t, lipgloss.Color("#d79921"), gradient[0]) // Orange
	assert.Equal(t, lipgloss.Color("#cc241d"), gradient[1]) // Red
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/themes -v -run Gruvbox
```

Expected: FAIL - NewGruvboxTheme undefined

**Step 3: Implement Gruvbox theme**

Create `internal/ui/themes/gruvbox.go`:

```go
package themes

import "github.com/charmbracelet/lipgloss"

// GruvboxTheme implements the Gruvbox Dark color scheme
type GruvboxTheme struct{}

// NewGruvboxTheme creates a new Gruvbox theme
func NewGruvboxTheme() Theme {
	return &GruvboxTheme{}
}

func (t *GruvboxTheme) Name() string {
	return "Gruvbox Dark"
}

func (t *GruvboxTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282828")
}

func (t *GruvboxTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#ebdbb2")
}

func (t *GruvboxTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#d79921") // Orange
}

func (t *GruvboxTheme) Secondary() lipgloss.Color {
	return lipgloss.Color("#b16286") // Purple
}

func (t *GruvboxTheme) Success() lipgloss.Color {
	return lipgloss.Color("#b8bb26") // Green
}

func (t *GruvboxTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#fabd2f") // Yellow
}

func (t *GruvboxTheme) Error() lipgloss.Color {
	return lipgloss.Color("#fb4934") // Red
}

func (t *GruvboxTheme) Border() lipgloss.Color {
	return lipgloss.Color("#504945") // Gray
}

func (t *GruvboxTheme) BorderFocus() lipgloss.Color {
	return lipgloss.Color("#d79921") // Orange
}

func (t *GruvboxTheme) Subtle() lipgloss.Color {
	return lipgloss.Color("#928374") // Gray
}

func (t *GruvboxTheme) TitleGradient() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color("#d79921"), // Orange
		lipgloss.Color("#cc241d"), // Red
	}
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/ui/themes -v -run Gruvbox
```

Expected: PASS

**Step 5: Commit Gruvbox theme**

```bash
git add internal/ui/themes/gruvbox.go internal/ui/themes/gruvbox_test.go
git commit -m "feat(ui): implement Gruvbox Dark theme"
```

---

## Task 4: Implement Nord Theme

**Files:**
- Create: `internal/ui/themes/nord.go`
- Create: `internal/ui/themes/nord_test.go`

**Step 1: Write Nord theme tests**

Create `internal/ui/themes/nord_test.go`:

```go
package themes_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jefft/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestNordTheme(t *testing.T) {
	theme := themes.NewNordTheme()

	assert.Equal(t, "Nord", theme.Name())
	assert.Equal(t, lipgloss.Color("#2e3440"), theme.Background())
	assert.Equal(t, lipgloss.Color("#eceff4"), theme.Foreground())
	assert.Equal(t, lipgloss.Color("#88c0d0"), theme.Primary())
	assert.Equal(t, lipgloss.Color("#81a1c1"), theme.Secondary())
}

func TestNordGradient(t *testing.T) {
	theme := themes.NewNordTheme()
	gradient := theme.TitleGradient()

	assert.Len(t, gradient, 2)
	assert.Equal(t, lipgloss.Color("#88c0d0"), gradient[0]) // Cyan
	assert.Equal(t, lipgloss.Color("#81a1c1"), gradient[1]) // Blue
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/themes -v -run Nord
```

Expected: FAIL - NewNordTheme undefined

**Step 3: Implement Nord theme**

Create `internal/ui/themes/nord.go`:

```go
package themes

import "github.com/charmbracelet/lipgloss"

// NordTheme implements the Nord color scheme
type NordTheme struct{}

// NewNordTheme creates a new Nord theme
func NewNordTheme() Theme {
	return &NordTheme{}
}

func (t *NordTheme) Name() string {
	return "Nord"
}

func (t *NordTheme) Background() lipgloss.Color {
	return lipgloss.Color("#2e3440")
}

func (t *NordTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#eceff4")
}

func (t *NordTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#88c0d0") // Cyan
}

func (t *NordTheme) Secondary() lipgloss.Color {
	return lipgloss.Color("#81a1c1") // Blue
}

func (t *NordTheme) Success() lipgloss.Color {
	return lipgloss.Color("#a3be8c") // Green
}

func (t *NordTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#ebcb8b") // Yellow
}

func (t *NordTheme) Error() lipgloss.Color {
	return lipgloss.Color("#bf616a") // Red
}

func (t *NordTheme) Border() lipgloss.Color {
	return lipgloss.Color("#4c566a") // Polar Night
}

func (t *NordTheme) BorderFocus() lipgloss.Color {
	return lipgloss.Color("#88c0d0") // Cyan
}

func (t *NordTheme) Subtle() lipgloss.Color {
	return lipgloss.Color("#4c566a") // Polar Night
}

func (t *NordTheme) TitleGradient() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color("#88c0d0"), // Cyan
		lipgloss.Color("#81a1c1"), // Blue
	}
}
```

**Step 4: Run all theme tests**

```bash
go test ./internal/ui/themes -v
```

Expected: ALL PASS

**Step 5: Commit Nord theme**

```bash
git add internal/ui/themes/nord.go internal/ui/themes/nord_test.go
git commit -m "feat(ui): implement Nord theme"
```

---

## Task 5: Add Theme Configuration Loading

**Files:**
- Modify: `internal/core/config.go`
- Modify: `internal/core/config_test.go`

**Step 1: Write test for theme config**

Add to `internal/core/config_test.go`:

```go
func TestConfigTheme(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `theme: gruvbox`
	err := os.WriteFile(configPath, []byte(configYAML), 0600)
	require.NoError(t, err)

	_ = os.Setenv("JEFF_CONFIG_PATH", configPath)
	defer func() { _ = os.Unsetenv("JEFF_CONFIG_PATH") }()

	cfg, err := core.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "gruvbox", cfg.Theme)
}

func TestConfigThemeDefault(t *testing.T) {
	_ = os.Unsetenv("JEFF_CONFIG_PATH")
	_ = os.Unsetenv("JEFF_THEME")

	cfg, err := core.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "dracula", cfg.Theme) // Default theme
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/core -v -run TestConfigTheme
```

Expected: FAIL - cfg.Theme undefined

**Step 3: Add Theme field to Config struct**

In `internal/core/config.go`, add to Config struct:

```go
type Config struct {
	APIKey         string            `mapstructure:"api_key"`
	Model          string            `mapstructure:"model"`
	DefaultTools   []string          `mapstructure:"default_tools"`
	PermissionMode string            `mapstructure:"permission_mode"`
	Theme          string            `mapstructure:"theme"`  // Add this line
	Hooks          hooks.HooksConfig `mapstructure:"hooks"`
}
```

**Step 4: Add theme default and binding**

In `internal/core/config.go` LoadConfig function, add:

```go
// After other SetDefault calls:
v.SetDefault("theme", "dracula")

// After other BindEnv calls:
_ = v.BindEnv("theme")
```

**Step 5: Run test to verify it passes**

```bash
go test ./internal/core -v -run TestConfigTheme
```

Expected: PASS

**Step 6: Commit theme configuration**

```bash
git add internal/core/config.go internal/core/config_test.go
git commit -m "feat(config): add theme configuration support"
```

---

## Task 6: Integrate Theme into UI Model

**Files:**
- Modify: `internal/ui/model.go`
- Modify: `cmd/jefft/root.go`

**Step 1: Add theme field to Model**

In `internal/ui/model.go`, add import and field:

```go
import (
	// ... existing imports
	"github.com/harper/jefft/internal/ui/themes"
)

type Model struct {
	// ... existing fields
	theme themes.Theme  // Add this field after renderer
}
```

**Step 2: Update NewModel to accept theme**

In `internal/ui/model.go`, modify NewModel signature and implementation:

```go
// NewModel creates a new Bubbletea model with the specified theme
func NewModel(conversationID, model string, theme themes.Theme) *Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()
	ta.ShowLineNumbers = false
	ta.CharLimit = 0

	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true

	return &Model{
		ConversationID: conversationID,
		Model:          model,
		Input:          ta,
		Viewport:       vp,
		Messages:       make([]Message, 0),
		CurrentView:    ViewModeChat,
		Status:         StatusIdle,
		theme:          theme,  // Add this line
	}
}
```

**Step 3: Update root.go to pass theme**

In `cmd/jefft/root.go`, update where NewModel is called (around line 227):

```go
// Load theme
themeName := cfg.Theme
if themeName == "" {
	themeName = "dracula"
}
theme := themes.GetTheme(themeName)

// Create UI model with theme
if uiModel == nil {
	conversationID = fmt.Sprintf("conv-%d", time.Now().Unix())
	uiModel = ui.NewModel(conversationID, modelName, theme)
	// ... rest of the code
}
```

Also add import:

```go
import (
	// ... existing imports
	"github.com/harper/jefft/internal/ui/themes"
)
```

**Step 4: Update other NewModel calls**

Find all other places NewModel is called and update them. Check:

```bash
grep -r "ui.NewModel" --include="*.go"
```

Update each to pass `themes.NewDraculaTheme()` as the third argument.

**Step 5: Build to verify no compile errors**

```bash
go build ./cmd/jefft
```

Expected: Success

**Step 6: Commit theme integration**

```bash
git add internal/ui/model.go cmd/jefft/root.go
git commit -m "feat(ui): integrate theme system into UI model"
```

---

## Task 7: Add Gradient Rendering Utility

**Files:**
- Create: `internal/ui/themes/gradient.go`
- Create: `internal/ui/themes/gradient_test.go`

**Step 1: Write gradient rendering tests**

Create `internal/ui/themes/gradient_test.go`:

```go
package themes_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jefft/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestRenderGradient(t *testing.T) {
	colors := []lipgloss.Color{
		lipgloss.Color("#ff0000"),
		lipgloss.Color("#00ff00"),
	}

	result := themes.RenderGradient("Hello", colors)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Hello")
}

func TestRenderGradientSingleColor(t *testing.T) {
	colors := []lipgloss.Color{
		lipgloss.Color("#ff0000"),
	}

	result := themes.RenderGradient("Hello", colors)
	assert.NotEmpty(t, result)
}

func TestRenderGradientEmpty(t *testing.T) {
	result := themes.RenderGradient("", []lipgloss.Color{lipgloss.Color("#ff0000")})
	assert.Empty(t, result)
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/themes -v -run Gradient
```

Expected: FAIL - RenderGradient undefined

**Step 3: Implement gradient rendering**

Create `internal/ui/themes/gradient.go`:

```go
package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// RenderGradient renders text with a color gradient
func RenderGradient(text string, colors []lipgloss.Color) string {
	if len(text) == 0 {
		return ""
	}

	if len(colors) == 0 {
		return text
	}

	if len(colors) == 1 {
		return lipgloss.NewStyle().Foreground(colors[0]).Render(text)
	}

	// Calculate how many characters per color step
	runes := []rune(text)
	numRunes := len(runes)
	numColors := len(colors)

	var result string
	for i, r := range runes {
		// Calculate which color to use based on position
		colorIndex := (i * (numColors - 1)) / (numRunes - 1)
		if colorIndex >= numColors {
			colorIndex = numColors - 1
		}

		style := lipgloss.NewStyle().Foreground(colors[colorIndex])
		result += style.Render(string(r))
	}

	return result
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/ui/themes -v -run Gradient
```

Expected: PASS

**Step 5: Commit gradient utility**

```bash
git add internal/ui/themes/gradient.go internal/ui/themes/gradient_test.go
git commit -m "feat(ui): add gradient text rendering utility"
```

---

## Task 8: Update View to Use Theme Colors

**Files:**
- Modify: `internal/ui/view.go`

**Step 1: Update title bar to use theme gradient**

In `internal/ui/view.go`, find the title rendering section and update it:

```go
// Find and replace the title rendering code with:
func (m Model) renderTitle() string {
	title := "Jeff - Productivity AI Agent"
	gradientTitle := themes.RenderGradient(title, m.theme.TitleGradient())

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	return titleStyle.Render(gradientTitle)
}
```

**Step 2: Update View() to use renderTitle()**

In the `View()` function, replace the title section with:

```go
title := m.renderTitle()
```

**Step 3: Update border styles to use theme**

Find all `lipgloss.NewStyle()` calls that set borders and update them:

```go
// Example - update border colors:
borderStyle := lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(m.theme.Border())

// For focused elements:
borderFocusStyle := lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(m.theme.BorderFocus())
```

**Step 4: Update text colors**

Find all color references and update to use theme methods:

```go
// Replace hardcoded colors like:
// lipgloss.Color("#ff0000")
// With:
// m.theme.Error()
```

**Step 5: Build and visually test**

```bash
go build ./cmd/jefft
./jefft --help
```

Check that help text appears (UI not broken).

**Step 6: Commit themed view**

```bash
git add internal/ui/view.go
git commit -m "feat(ui): apply theme colors to view rendering"
```

---

## Task 9: Update Status Bar to Use Theme

**Files:**
- Modify: `internal/ui/statusbar.go`

**Step 1: Update status bar styles**

In `internal/ui/statusbar.go`, update the Render() function to use theme colors:

```go
func (s *StatusBar) Render(theme themes.Theme) string {
	// Status indicator styling with theme colors
	var statusStyle lipgloss.Style
	switch s.status {
	case StatusIdle:
		statusStyle = lipgloss.NewStyle().Foreground(theme.Success())
	case StatusStreaming:
		statusStyle = lipgloss.NewStyle().Foreground(theme.Primary())
	case StatusError:
		statusStyle = lipgloss.NewStyle().Foreground(theme.Error())
	default:
		statusStyle = lipgloss.NewStyle().Foreground(theme.Foreground())
	}

	// ... rest of implementation using theme colors
}
```

**Step 2: Update status bar call sites**

In `view.go`, update status bar rendering to pass theme:

```go
statusBar := m.statusBar.Render(m.theme)
```

**Step 3: Build to verify**

```bash
go build ./cmd/jefft
```

Expected: Success

**Step 4: Commit themed status bar**

```bash
git add internal/ui/statusbar.go internal/ui/view.go
git commit -m "feat(ui): apply theme colors to status bar"
```

---

## Task 10: Add CLI Flag for Theme Selection

**Files:**
- Modify: `cmd/jefft/root.go`

**Step 1: Add theme flag**

In `cmd/jefft/root.go` init() function, add flag:

```go
var themeFlag string

func init() {
	// ... existing flags
	rootCmd.PersistentFlags().StringVar(&themeFlag, "theme", "", "UI theme: dracula, gruvbox, nord (default: dracula)")
}
```

**Step 2: Update theme loading in runInteractive**

In runInteractive(), update theme loading:

```go
// Determine theme (CLI flag overrides config)
themeName := themeFlag
if themeName == "" {
	themeName = cfg.Theme
}
if themeName == "" {
	themeName = "dracula"
}
theme := themes.GetTheme(themeName)
```

**Step 3: Test CLI flag**

```bash
go build ./cmd/jefft
./jefft --theme nord --help
./jefft --theme gruvbox --help
./jefft --theme dracula --help
```

Expected: Each should work without errors

**Step 4: Commit theme flag**

```bash
git add cmd/jefft/root.go
git commit -m "feat(cli): add --theme flag for runtime theme selection"
```

---

## Phase 1 Complete!

All tests should pass:

```bash
go test ./internal/ui/themes -v
go test ./internal/core -v -run Theme
go build ./cmd/jefft
```

Expected output:
- All theme tests PASS
- Config theme tests PASS
- Binary builds successfully
- Can run with `--theme` flag

**Final verification:**

```bash
./jefft --theme dracula
./jefft --theme gruvbox
./jefft --theme nord
```

Each should launch with appropriate color scheme.

---

## Next Steps

Phase 1 is complete! Ready to move to Phase 2: Huh Forms Integration.

Create a new plan file: `2025-12-02-tui-refactor-phase2-implementation.md`
