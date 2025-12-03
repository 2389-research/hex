// Package layout provides consistent border styles and spacing utilities for TUI layout.
// ABOUTME: Comprehensive tests for border styles and layout utilities
// ABOUTME: Tests border rendering, spacing, column layouts, and theme integration
package layout

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/pagent/internal/ui/themes"
)

// TestBorderSets verifies that all border character sets are defined correctly
func TestBorderSets(t *testing.T) {
	tests := []struct {
		name   string
		border BorderSet
	}{
		{"RoundedBorder", RoundedBorder},
		{"ThickBorder", ThickBorder},
		{"DoubleBorder", DoubleBorder},
		{"NormalBorder", NormalBorder},
		{"HiddenBorder", HiddenBorder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all border characters are set
			if tt.border.TopLeft == "" {
				t.Error("TopLeft should not be empty")
			}
			if tt.border.Top == "" {
				t.Error("Top should not be empty")
			}
			if tt.border.TopRight == "" {
				t.Error("TopRight should not be empty")
			}
			if tt.border.Right == "" {
				t.Error("Right should not be empty")
			}
			if tt.border.BottomRight == "" {
				t.Error("BottomRight should not be empty")
			}
			if tt.border.Bottom == "" {
				t.Error("Bottom should not be empty")
			}
			if tt.border.BottomLeft == "" {
				t.Error("BottomLeft should not be empty")
			}
			if tt.border.Left == "" {
				t.Error("Left should not be empty")
			}
		})
	}
}

// TestNewSpacing verifies uniform spacing configuration
func TestNewSpacing(t *testing.T) {
	spacing := NewSpacing(2, 1)

	if spacing.PaddingTop != 2 || spacing.PaddingRight != 2 ||
		spacing.PaddingBottom != 2 || spacing.PaddingLeft != 2 {
		t.Errorf("Expected uniform padding of 2, got top=%d right=%d bottom=%d left=%d",
			spacing.PaddingTop, spacing.PaddingRight, spacing.PaddingBottom, spacing.PaddingLeft)
	}

	if spacing.MarginTop != 1 || spacing.MarginRight != 1 ||
		spacing.MarginBottom != 1 || spacing.MarginLeft != 1 {
		t.Errorf("Expected uniform margin of 1, got top=%d right=%d bottom=%d left=%d",
			spacing.MarginTop, spacing.MarginRight, spacing.MarginBottom, spacing.MarginLeft)
	}
}

// TestNewPadding verifies padding-only configuration
func TestNewPadding(t *testing.T) {
	spacing := NewPadding(1, 2, 3, 4)

	if spacing.PaddingTop != 1 || spacing.PaddingRight != 2 ||
		spacing.PaddingBottom != 3 || spacing.PaddingLeft != 4 {
		t.Errorf("Expected padding top=1 right=2 bottom=3 left=4, got top=%d right=%d bottom=%d left=%d",
			spacing.PaddingTop, spacing.PaddingRight, spacing.PaddingBottom, spacing.PaddingLeft)
	}

	// Margins should be zero
	if spacing.MarginTop != 0 || spacing.MarginRight != 0 ||
		spacing.MarginBottom != 0 || spacing.MarginLeft != 0 {
		t.Error("Expected all margins to be 0 for NewPadding")
	}
}

// TestNewMargin verifies margin-only configuration
func TestNewMargin(t *testing.T) {
	spacing := NewMargin(1, 2, 3, 4)

	if spacing.MarginTop != 1 || spacing.MarginRight != 2 ||
		spacing.MarginBottom != 3 || spacing.MarginLeft != 4 {
		t.Errorf("Expected margin top=1 right=2 bottom=3 left=4, got top=%d right=%d bottom=%d left=%d",
			spacing.MarginTop, spacing.MarginRight, spacing.MarginBottom, spacing.MarginLeft)
	}

	// Padding should be zero
	if spacing.PaddingTop != 0 || spacing.PaddingRight != 0 ||
		spacing.PaddingBottom != 0 || spacing.PaddingLeft != 0 {
		t.Error("Expected all padding to be 0 for NewMargin")
	}
}

// TestNewBorderStyle verifies default border style creation
func TestNewBorderStyle(t *testing.T) {
	theme := themes.NewDracula()
	bs := NewBorderStyle(theme)

	if bs == nil {
		t.Fatal("NewBorderStyle returned nil")
	}

	if bs.Border != RoundedBorder {
		t.Error("Default border should be RoundedBorder")
	}

	if bs.Color != theme.Border() {
		t.Error("Default color should be theme's border color")
	}

	if bs.TitleAlign != lipgloss.Left {
		t.Error("Default title alignment should be Left")
	}

	if bs.Focused {
		t.Error("Default focused state should be false")
	}
}

// TestBorderStyleWithBorder verifies border set configuration
func TestBorderStyleWithBorder(t *testing.T) {
	theme := themes.NewDracula()
	bs := NewBorderStyle(theme).WithBorder(DoubleBorder)

	if bs.Border != DoubleBorder {
		t.Error("WithBorder should set DoubleBorder")
	}
}

// TestBorderStyleWithColor verifies custom color configuration
func TestBorderStyleWithColor(t *testing.T) {
	theme := themes.NewDracula()
	customColor := lipgloss.Color("#ff0000")
	bs := NewBorderStyle(theme).WithColor(customColor)

	if bs.Color != customColor {
		t.Errorf("WithColor should set custom color, got %v want %v", bs.Color, customColor)
	}
}

// TestBorderStyleWithSpacing verifies spacing configuration
func TestBorderStyleWithSpacing(t *testing.T) {
	theme := themes.NewDracula()
	spacing := NewSpacing(2, 1)
	bs := NewBorderStyle(theme).WithSpacing(spacing)

	if bs.Spacing != spacing {
		t.Error("WithSpacing should set spacing configuration")
	}
}

// TestBorderStyleWithSize verifies size configuration
func TestBorderStyleWithSize(t *testing.T) {
	theme := themes.NewDracula()
	bs := NewBorderStyle(theme).WithSize(80, 24)

	if bs.Width != 80 {
		t.Errorf("Expected width 80, got %d", bs.Width)
	}
	if bs.Height != 24 {
		t.Errorf("Expected height 24, got %d", bs.Height)
	}
}

// TestBorderStyleWithTitle verifies title configuration
func TestBorderStyleWithTitle(t *testing.T) {
	theme := themes.NewDracula()
	bs := NewBorderStyle(theme).WithTitle("Test Title", lipgloss.Center)

	if bs.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", bs.Title)
	}
	if bs.TitleAlign != lipgloss.Center {
		t.Error("Expected title alignment Center")
	}
}

// TestBorderStyleWithFocus verifies focus state changes border color
func TestBorderStyleWithFocus(t *testing.T) {
	theme := themes.NewDracula()

	// Test unfocused state
	bs := NewBorderStyle(theme).WithFocus(false)
	if bs.Focused {
		t.Error("Expected focused to be false")
	}
	if bs.Color != theme.Border() {
		t.Error("Unfocused border should use theme's border color")
	}

	// Test focused state
	bs = NewBorderStyle(theme).WithFocus(true)
	if !bs.Focused {
		t.Error("Expected focused to be true")
	}
	if bs.Color != theme.BorderFocus() {
		t.Error("Focused border should use theme's focus color")
	}
}

// TestBorderStyleRender verifies basic rendering
func TestBorderStyleRender(t *testing.T) {
	theme := themes.NewDracula()
	bs := NewBorderStyle(theme)
	content := "Hello, World!"

	rendered := bs.Render(content)

	if rendered == "" {
		t.Error("Rendered output should not be empty")
	}

	if !strings.Contains(rendered, content) {
		t.Error("Rendered output should contain original content")
	}

	// Should contain border characters
	lines := strings.Split(rendered, "\n")
	if len(lines) < 3 {
		t.Error("Rendered output should have at least 3 lines (top border, content, bottom border)")
	}
}

// TestBorderStyleRenderWithTitle verifies title rendering
func TestBorderStyleRenderWithTitle(t *testing.T) {
	theme := themes.NewDracula()
	bs := NewBorderStyle(theme).WithTitle("Title", lipgloss.Left)
	content := "Content"

	rendered := bs.Render(content)
	plainRendered := stripANSI(rendered)

	if !strings.Contains(plainRendered, "Title") {
		t.Errorf("Rendered output should contain title. Got: %s", plainRendered)
	}
	if !strings.Contains(plainRendered, content) {
		t.Errorf("Rendered output should contain content. Got: %s", plainRendered)
	}
}

// TestSeparator verifies horizontal separator creation
func TestSeparator(t *testing.T) {
	color := lipgloss.Color("#ffffff")

	tests := []struct {
		name     string
		width    int
		expected string
	}{
		{"zero width", 0, ""},
		{"negative width", -1, ""},
		{"positive width", 5, "─────"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Separator(tt.width, color)
			// Strip ANSI codes for length comparison
			plainResult := stripANSI(result)
			if plainResult != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, plainResult)
			}
		})
	}
}

// TestVerticalSeparator verifies vertical separator creation
func TestVerticalSeparator(t *testing.T) {
	color := lipgloss.Color("#ffffff")

	tests := []struct {
		name   string
		height int
		lines  int
	}{
		{"zero height", 0, 0},
		{"negative height", -1, 0},
		{"positive height", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerticalSeparator(tt.height, color)
			if tt.lines == 0 {
				if result != "" {
					t.Errorf("Expected empty string, got '%s'", result)
				}
				return
			}

			lines := strings.Split(result, "\n")
			if len(lines) != tt.lines {
				t.Errorf("Expected %d lines, got %d", tt.lines, len(lines))
			}
		})
	}
}

// TestBox verifies simple box creation
func TestBox(t *testing.T) {
	theme := themes.NewDracula()
	content := "Test content"

	result := Box(content, theme)

	if result == "" {
		t.Error("Box should produce output")
	}
	if !strings.Contains(result, content) {
		t.Error("Box should contain original content")
	}
}

// TestFocusedBox verifies focused box creation
func TestFocusedBox(t *testing.T) {
	theme := themes.NewDracula()
	content := "Test content"

	result := FocusedBox(content, theme)

	if result == "" {
		t.Error("FocusedBox should produce output")
	}
	if !strings.Contains(result, content) {
		t.Error("FocusedBox should contain original content")
	}
}

// TestTitledBox verifies titled box creation
func TestTitledBox(t *testing.T) {
	theme := themes.NewDracula()
	content := "Test content"
	title := "Test Title"

	result := TitledBox(content, title, theme)
	plainResult := stripANSI(result)

	if result == "" {
		t.Error("TitledBox should produce output")
	}
	if !strings.Contains(plainResult, content) {
		t.Errorf("TitledBox should contain original content. Got: %s", plainResult)
	}
	if !strings.Contains(plainResult, title) {
		t.Errorf("TitledBox should contain title. Got: %s", plainResult)
	}
}

// TestPaddedContent verifies padding application
func TestPaddedContent(t *testing.T) {
	content := "X"
	spacing := NewPadding(1, 1, 1, 1)

	result := PaddedContent(content, spacing)

	if result == "" {
		t.Error("PaddedContent should produce output")
	}

	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Error("PaddedContent with top and bottom padding should have at least 3 lines")
	}
}

// TestPadding verifies the Padding utility function
func TestPadding(t *testing.T) {
	content := "X"

	result := Padding(content, 1, 1, 1, 1)

	if result == "" {
		t.Error("Padding should produce output")
	}
	if !strings.Contains(result, content) {
		t.Error("Padding should contain original content")
	}
}

// TestMargin verifies margin application
func TestMargin(t *testing.T) {
	content := "X"

	// Test top margin
	result := Margin(content, 2, 0, 0, 0)
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Error("Top margin should add newlines before content")
	}

	// Test bottom margin
	result = Margin(content, 0, 0, 2, 0)
	lines = strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Error("Bottom margin should add newlines after content")
	}

	// Test left margin
	result = Margin(content, 0, 0, 0, 3)
	if !strings.HasPrefix(result, "   ") {
		t.Error("Left margin should add spaces before content")
	}

	// Test right margin
	result = Margin(content, 0, 3, 0, 0)
	if !strings.HasSuffix(result, "   ") {
		t.Error("Right margin should add spaces after content")
	}
}

// TestJoinHorizontal verifies horizontal joining
func TestJoinHorizontal(t *testing.T) {
	tests := []struct {
		name     string
		spacing  int
		elements []string
		expected string
	}{
		{"empty", 0, []string{}, ""},
		{"single element", 2, []string{"A"}, "A"},
		{"two elements", 2, []string{"A", "B"}, "A  B"},
		{"three elements", 1, []string{"A", "B", "C"}, "A B C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinHorizontal(tt.spacing, tt.elements...)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestJoinVertical verifies vertical joining
func TestJoinVertical(t *testing.T) {
	tests := []struct {
		name     string
		spacing  int
		elements []string
		lines    int
	}{
		{"empty", 0, []string{}, 0},
		{"single element", 2, []string{"A"}, 1},
		{"two elements no spacing", 0, []string{"A", "B"}, 2},
		{"two elements with spacing", 1, []string{"A", "B"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinVertical(tt.spacing, tt.elements...)
			if tt.lines == 0 {
				if result != "" {
					t.Errorf("Expected empty string, got '%s'", result)
				}
				return
			}

			lines := strings.Split(result, "\n")
			if len(lines) != tt.lines {
				t.Errorf("Expected %d lines, got %d", tt.lines, len(lines))
			}
		})
	}
}

// TestTwoColumn verifies two-column layout
func TestTwoColumn(t *testing.T) {
	theme := themes.NewDracula()
	left := "Left"
	right := "Right"

	result := TwoColumn(left, right, theme)

	if result == "" {
		t.Error("TwoColumn should produce output")
	}
	if !strings.Contains(result, left) {
		t.Error("TwoColumn should contain left content")
	}
	if !strings.Contains(result, right) {
		t.Error("TwoColumn should contain right content")
	}
}

// TestThreeColumn verifies three-column layout
func TestThreeColumn(t *testing.T) {
	theme := themes.NewDracula()
	left := "Left"
	middle := "Middle"
	right := "Right"

	result := ThreeColumn(left, middle, right, theme)

	if result == "" {
		t.Error("ThreeColumn should produce output")
	}
	if !strings.Contains(result, left) {
		t.Error("ThreeColumn should contain left content")
	}
	if !strings.Contains(result, middle) {
		t.Error("ThreeColumn should contain middle content")
	}
	if !strings.Contains(result, right) {
		t.Error("ThreeColumn should contain right content")
	}
}

// TestPlaceHorizontal verifies horizontal positioning
func TestPlaceHorizontal(t *testing.T) {
	content := "X"
	width := 10

	tests := []struct {
		name string
		pos  lipgloss.Position
	}{
		{"left", lipgloss.Left},
		{"center", lipgloss.Center},
		{"right", lipgloss.Right},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PlaceHorizontal(width, tt.pos, content)
			if result == "" {
				t.Error("PlaceHorizontal should produce output")
			}
			plainResult := stripANSI(result)
			if len(plainResult) < len(content) {
				t.Error("PlaceHorizontal should maintain or expand content width")
			}
		})
	}
}

// TestPlaceVertical verifies vertical positioning
func TestPlaceVertical(t *testing.T) {
	content := "X"
	height := 5

	tests := []struct {
		name string
		pos  lipgloss.Position
	}{
		{"top", lipgloss.Top},
		{"center", lipgloss.Center},
		{"bottom", lipgloss.Bottom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PlaceVertical(height, tt.pos, content)
			if result == "" {
				t.Error("PlaceVertical should produce output")
			}
			lines := strings.Split(result, "\n")
			if len(lines) != height {
				t.Errorf("PlaceVertical should produce exactly %d lines, got %d", height, len(lines))
			}
		})
	}
}

// TestMaxWidth verifies maximum width constraint
func TestMaxWidth(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		content string
	}{
		{"zero width", 0, "test"},
		{"negative width", -1, "test"},
		{"normal width", 10, "this is a long string that should wrap"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxWidth(tt.width, tt.content)
			if tt.width <= 0 {
				if result != tt.content {
					t.Error("MaxWidth with invalid width should return original content")
				}
				return
			}
			if result == "" {
				t.Error("MaxWidth should produce output")
			}
		})
	}
}

// TestThemeIntegration verifies theme integration across all themes
func TestThemeIntegration(t *testing.T) {
	themeNames := []string{"dracula", "gruvbox", "nord"}

	for _, themeName := range themeNames {
		t.Run(themeName, func(t *testing.T) {
			theme := themes.GetTheme(themeName)
			content := "Test content"

			// Test Box
			result := Box(content, theme)
			if !strings.Contains(result, content) {
				t.Errorf("%s theme: Box should contain content", themeName)
			}

			// Test FocusedBox
			result = FocusedBox(content, theme)
			if !strings.Contains(result, content) {
				t.Errorf("%s theme: FocusedBox should contain content", themeName)
			}

			// Test TitledBox
			result = TitledBox(content, "Title", theme)
			if !strings.Contains(result, content) || !strings.Contains(result, "Title") {
				t.Errorf("%s theme: TitledBox should contain content and title", themeName)
			}
		})
	}
}

// TestBorderStyleChaining verifies method chaining works correctly
func TestBorderStyleChaining(t *testing.T) {
	theme := themes.NewDracula()

	bs := NewBorderStyle(theme).
		WithBorder(ThickBorder).
		WithColor(lipgloss.Color("#ff0000")).
		WithSpacing(NewSpacing(1, 1)).
		WithSize(50, 10).
		WithTitle("Chained Title", lipgloss.Center).
		WithFocus(true)

	if bs.Border != ThickBorder {
		t.Error("Chained border should be ThickBorder")
	}
	if bs.Width != 50 {
		t.Error("Chained width should be 50")
	}
	if bs.Height != 10 {
		t.Error("Chained height should be 10")
	}
	if bs.Title != "Chained Title" {
		t.Error("Chained title should be set")
	}
	if !bs.Focused {
		t.Error("Chained focus should be true")
	}
}

// TestDifferentContentSizes verifies rendering with various content sizes
func TestDifferentContentSizes(t *testing.T) {
	theme := themes.NewDracula()

	tests := []struct {
		name    string
		content string
	}{
		{"empty", ""},
		{"single char", "X"},
		{"short", "Hello"},
		{"multiline", "Line 1\nLine 2\nLine 3"},
		{"long", strings.Repeat("This is a long line. ", 10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Box(tt.content, theme)
			if result == "" && tt.content != "" {
				t.Error("Box should produce output for non-empty content")
			}
		})
	}
}

// stripANSI removes ANSI escape codes from a string for testing
func stripANSI(s string) string {
	// Simple ANSI stripping - match ESC[ ... m sequences
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++ // Skip the '['
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
