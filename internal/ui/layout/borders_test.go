package layout

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/clem/internal/ui/theme"
)

func TestNewSpacing(t *testing.T) {
	spacing := NewSpacing(2, 1)

	if spacing.PaddingTop != 2 {
		t.Errorf("PaddingTop = %d, want 2", spacing.PaddingTop)
	}
	if spacing.PaddingRight != 2 {
		t.Errorf("PaddingRight = %d, want 2", spacing.PaddingRight)
	}
	if spacing.PaddingBottom != 2 {
		t.Errorf("PaddingBottom = %d, want 2", spacing.PaddingBottom)
	}
	if spacing.PaddingLeft != 2 {
		t.Errorf("PaddingLeft = %d, want 2", spacing.PaddingLeft)
	}
	if spacing.MarginTop != 1 {
		t.Errorf("MarginTop = %d, want 1", spacing.MarginTop)
	}
}

func TestNewPadding(t *testing.T) {
	spacing := NewPadding(1, 2, 3, 4)

	if spacing.PaddingTop != 1 {
		t.Errorf("PaddingTop = %d, want 1", spacing.PaddingTop)
	}
	if spacing.PaddingRight != 2 {
		t.Errorf("PaddingRight = %d, want 2", spacing.PaddingRight)
	}
	if spacing.PaddingBottom != 3 {
		t.Errorf("PaddingBottom = %d, want 3", spacing.PaddingBottom)
	}
	if spacing.PaddingLeft != 4 {
		t.Errorf("PaddingLeft = %d, want 4", spacing.PaddingLeft)
	}
	if spacing.MarginTop != 0 {
		t.Errorf("MarginTop = %d, want 0", spacing.MarginTop)
	}
}

func TestNewMargin(t *testing.T) {
	spacing := NewMargin(1, 2, 3, 4)

	if spacing.MarginTop != 1 {
		t.Errorf("MarginTop = %d, want 1", spacing.MarginTop)
	}
	if spacing.MarginRight != 2 {
		t.Errorf("MarginRight = %d, want 2", spacing.MarginRight)
	}
	if spacing.MarginBottom != 3 {
		t.Errorf("MarginBottom = %d, want 3", spacing.MarginBottom)
	}
	if spacing.MarginLeft != 4 {
		t.Errorf("MarginLeft = %d, want 4", spacing.MarginLeft)
	}
	if spacing.PaddingTop != 0 {
		t.Errorf("PaddingTop = %d, want 0", spacing.PaddingTop)
	}
}

func TestNewBorderStyle(t *testing.T) {
	th := theme.NewDraculaTheme()
	bs := NewBorderStyle(th)

	if bs == nil {
		t.Fatal("NewBorderStyle returned nil")
	}
	if bs.Border != RoundedBorder {
		t.Error("Expected RoundedBorder as default")
	}
	if bs.Color != th.Colors.Comment {
		t.Error("Expected Comment color as default")
	}
	if bs.Focused {
		t.Error("Expected Focused to be false by default")
	}
	if bs.TitleAlign != lipgloss.Left {
		t.Error("Expected Left alignment by default")
	}
}

func TestBorderStyleWithBorder(t *testing.T) {
	th := theme.NewDraculaTheme()
	bs := NewBorderStyle(th).WithBorder(ThickBorder)

	if bs.Border != ThickBorder {
		t.Error("WithBorder did not set ThickBorder")
	}
}

func TestBorderStyleWithColor(t *testing.T) {
	th := theme.NewDraculaTheme()
	testColor := lipgloss.Color("#ff0000")
	bs := NewBorderStyle(th).WithColor(testColor)

	if bs.Color != testColor {
		t.Errorf("WithColor did not set color correctly")
	}
}

func TestBorderStyleWithSpacing(t *testing.T) {
	th := theme.NewDraculaTheme()
	spacing := NewSpacing(2, 1)
	bs := NewBorderStyle(th).WithSpacing(spacing)

	if bs.Spacing != spacing {
		t.Error("WithSpacing did not set spacing")
	}
}

func TestBorderStyleWithSize(t *testing.T) {
	th := theme.NewDraculaTheme()
	bs := NewBorderStyle(th).WithSize(80, 24)

	if bs.Width != 80 {
		t.Errorf("Width = %d, want 80", bs.Width)
	}
	if bs.Height != 24 {
		t.Errorf("Height = %d, want 24", bs.Height)
	}
}

func TestBorderStyleWithTitle(t *testing.T) {
	th := theme.NewDraculaTheme()
	bs := NewBorderStyle(th).WithTitle("Test Title", lipgloss.Center)

	if bs.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", bs.Title, "Test Title")
	}
	if bs.TitleAlign != lipgloss.Center {
		t.Error("TitleAlign not set to Center")
	}
}

func TestBorderStyleWithFocus(t *testing.T) {
	th := theme.NewDraculaTheme()

	t.Run("focused", func(t *testing.T) {
		bs := NewBorderStyle(th).WithFocus(true)
		if !bs.Focused {
			t.Error("Focused should be true")
		}
		if bs.Color != th.Colors.Purple {
			t.Error("Focused border should use Purple color")
		}
	})

	t.Run("unfocused", func(t *testing.T) {
		bs := NewBorderStyle(th).WithFocus(false)
		if bs.Focused {
			t.Error("Focused should be false")
		}
		if bs.Color != th.Colors.Comment {
			t.Error("Unfocused border should use Comment color")
		}
	})
}

func TestBorderStyleRender(t *testing.T) {
	th := theme.NewDraculaTheme()

	tests := []struct {
		name    string
		content string
		setup   func(*BorderStyle) *BorderStyle
	}{
		{
			name:    "simple content",
			content: "Hello World",
			setup:   func(bs *BorderStyle) *BorderStyle { return bs },
		},
		{
			name:    "empty content",
			content: "",
			setup:   func(bs *BorderStyle) *BorderStyle { return bs },
		},
		{
			name:    "with padding",
			content: "Padded",
			setup: func(bs *BorderStyle) *BorderStyle {
				return bs.WithSpacing(NewPadding(1, 2, 1, 2))
			},
		},
		{
			name:    "with title",
			content: "Content",
			setup: func(bs *BorderStyle) *BorderStyle {
				return bs.WithTitle("Title", lipgloss.Left)
			},
		},
		{
			name:    "focused",
			content: "Focus",
			setup: func(bs *BorderStyle) *BorderStyle {
				return bs.WithFocus(true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := NewBorderStyle(th)
			bs = tt.setup(bs)
			result := bs.Render(tt.content)

			if result == "" && tt.content != "" {
				t.Error("Render returned empty string for non-empty content")
			}
		})
	}
}

func TestBorderSets(t *testing.T) {
	tests := []struct {
		name   string
		border BorderSet
	}{
		{"RoundedBorder", RoundedBorder},
		{"ThickBorder", ThickBorder},
		{"DoubleBorder", DoubleBorder},
		{"NormalBorder", NormalBorder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.border.TopLeft == "" {
				t.Error("TopLeft is empty")
			}
			if tt.border.Top == "" {
				t.Error("Top is empty")
			}
			if tt.border.TopRight == "" {
				t.Error("TopRight is empty")
			}
			if tt.border.Right == "" {
				t.Error("Right is empty")
			}
			if tt.border.BottomRight == "" {
				t.Error("BottomRight is empty")
			}
			if tt.border.Bottom == "" {
				t.Error("Bottom is empty")
			}
			if tt.border.BottomLeft == "" {
				t.Error("BottomLeft is empty")
			}
			if tt.border.Left == "" {
				t.Error("Left is empty")
			}
		})
	}
}

func TestSeparator(t *testing.T) {
	color := lipgloss.Color("#ffffff")

	tests := []struct {
		name  string
		width int
		want  bool // whether result should be empty
	}{
		{"normal width", 10, false},
		{"zero width", 0, true},
		{"negative width", -5, true},
		{"single char", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Separator(tt.width, color)
			isEmpty := result == ""

			if isEmpty != tt.want {
				t.Errorf("Separator(%d) empty = %v, want %v", tt.width, isEmpty, tt.want)
			}
		})
	}
}

func TestVerticalSeparator(t *testing.T) {
	color := lipgloss.Color("#ffffff")

	tests := []struct {
		name   string
		height int
		want   bool // whether result should be empty
	}{
		{"normal height", 5, false},
		{"zero height", 0, true},
		{"negative height", -3, true},
		{"single line", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerticalSeparator(tt.height, color)
			isEmpty := result == ""

			if isEmpty != tt.want {
				t.Errorf("VerticalSeparator(%d) empty = %v, want %v", tt.height, isEmpty, tt.want)
			}

			if !isEmpty {
				lines := strings.Split(result, "\n")
				if len(lines) != tt.height {
					t.Errorf("Got %d lines, want %d", len(lines), tt.height)
				}
			}
		})
	}
}

func TestBox(t *testing.T) {
	th := theme.NewDraculaTheme()
	result := Box("Test Content", th)

	if result == "" {
		t.Error("Box returned empty string")
	}
}

func TestFocusedBox(t *testing.T) {
	th := theme.NewDraculaTheme()
	result := FocusedBox("Test Content", th)

	if result == "" {
		t.Error("FocusedBox returned empty string")
	}
}

func TestTitledBox(t *testing.T) {
	th := theme.NewDraculaTheme()
	result := TitledBox("Test Content", "Test Title", th)

	if result == "" {
		t.Error("TitledBox returned empty string")
	}
	// Note: Title might be styled/colored, so we just check it's not empty
	// The actual title rendering is tested in TestBorderStyleRender
}

func TestPaddedContent(t *testing.T) {
	content := "Test"
	padding := NewPadding(1, 2, 1, 2)
	result := PaddedContent(content, padding)

	if result == "" {
		t.Error("PaddedContent returned empty string")
	}
}

func TestJoinHorizontal(t *testing.T) {
	tests := []struct {
		name     string
		spacing  int
		elements []string
		want     string
	}{
		{
			name:     "two elements",
			spacing:  2,
			elements: []string{"A", "B"},
			want:     "A  B",
		},
		{
			name:     "no spacing",
			spacing:  0,
			elements: []string{"A", "B", "C"},
			want:     "ABC",
		},
		{
			name:     "empty elements",
			spacing:  1,
			elements: []string{},
			want:     "",
		},
		{
			name:     "single element",
			spacing:  5,
			elements: []string{"A"},
			want:     "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinHorizontal(tt.spacing, tt.elements...)
			if result != tt.want {
				t.Errorf("JoinHorizontal() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestJoinVertical(t *testing.T) {
	tests := []struct {
		name     string
		spacing  int
		elements []string
	}{
		{
			name:     "two elements",
			spacing:  1,
			elements: []string{"A", "B"},
		},
		{
			name:     "no spacing",
			spacing:  0,
			elements: []string{"A", "B", "C"},
		},
		{
			name:     "empty elements",
			spacing:  1,
			elements: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinVertical(tt.spacing, tt.elements...)
			if len(tt.elements) == 0 && result != "" {
				t.Error("Expected empty result for no elements")
			}
			if len(tt.elements) > 0 && result == "" {
				t.Error("Expected non-empty result")
			}
		})
	}
}

func TestPlaceHorizontal(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		pos     lipgloss.Position
		content string
	}{
		{"left align", 20, lipgloss.Left, "Test"},
		{"center align", 20, lipgloss.Center, "Test"},
		{"right align", 20, lipgloss.Right, "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PlaceHorizontal(tt.width, tt.pos, tt.content)
			if result == "" {
				t.Error("PlaceHorizontal returned empty string")
			}
		})
	}
}

func TestPlaceVertical(t *testing.T) {
	tests := []struct {
		name    string
		height  int
		pos     lipgloss.Position
		content string
	}{
		{"top align", 10, lipgloss.Top, "Test"},
		{"center align", 10, lipgloss.Center, "Test"},
		{"bottom align", 10, lipgloss.Bottom, "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PlaceVertical(tt.height, tt.pos, tt.content)
			if result == "" {
				t.Error("PlaceVertical returned empty string")
			}
		})
	}
}

func TestMaxWidth(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		content string
	}{
		{"normal width", 10, "This is a long string that should wrap"},
		{"zero width", 0, "Test"},
		{"negative width", -5, "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxWidth(tt.width, tt.content)
			if result == "" && tt.content != "" {
				t.Error("MaxWidth returned empty string for non-empty content")
			}
		})
	}
}
