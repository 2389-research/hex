// Package layout provides consistent border styles and spacing utilities for TUI layout.
// ABOUTME: Border styles and spacing utilities for consistent TUI layout
// ABOUTME: Provides responsive layout components with Dracula theme integration
package layout

import (
	"strings"

	"github.com/2389-research/hex/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// BorderSet represents a complete set of border characters
type BorderSet struct {
	TopLeft     string
	Top         string
	TopRight    string
	Right       string
	BottomRight string
	Bottom      string
	BottomLeft  string
	Left        string
}

// Common border sets
var (
	// RoundedBorder uses rounded corners
	RoundedBorder = BorderSet{
		TopLeft:     "╭",
		Top:         "─",
		TopRight:    "╮",
		Right:       "│",
		BottomRight: "╯",
		Bottom:      "─",
		BottomLeft:  "╰",
		Left:        "│",
	}

	// ThickBorder uses thick lines
	ThickBorder = BorderSet{
		TopLeft:     "┏",
		Top:         "━",
		TopRight:    "┓",
		Right:       "┃",
		BottomRight: "┛",
		Bottom:      "━",
		BottomLeft:  "┗",
		Left:        "┃",
	}

	// DoubleBorder uses double lines
	DoubleBorder = BorderSet{
		TopLeft:     "╔",
		Top:         "═",
		TopRight:    "╗",
		Right:       "║",
		BottomRight: "╝",
		Bottom:      "═",
		BottomLeft:  "╚",
		Left:        "║",
	}

	// NormalBorder uses standard box drawing characters
	NormalBorder = BorderSet{
		TopLeft:     "┌",
		Top:         "─",
		TopRight:    "┐",
		Right:       "│",
		BottomRight: "┘",
		Bottom:      "─",
		BottomLeft:  "└",
		Left:        "│",
	}
)

// SpacingConfig defines padding and margin settings
type SpacingConfig struct {
	PaddingTop    int
	PaddingRight  int
	PaddingBottom int
	PaddingLeft   int
	MarginTop     int
	MarginRight   int
	MarginBottom  int
	MarginLeft    int
}

// NewSpacing creates a SpacingConfig with uniform padding and margin
func NewSpacing(padding, margin int) SpacingConfig {
	return SpacingConfig{
		PaddingTop:    padding,
		PaddingRight:  padding,
		PaddingBottom: padding,
		PaddingLeft:   padding,
		MarginTop:     margin,
		MarginRight:   margin,
		MarginBottom:  margin,
		MarginLeft:    margin,
	}
}

// NewPadding creates a SpacingConfig with only padding
func NewPadding(top, right, bottom, left int) SpacingConfig {
	return SpacingConfig{
		PaddingTop:    top,
		PaddingRight:  right,
		PaddingBottom: bottom,
		PaddingLeft:   left,
	}
}

// NewMargin creates a SpacingConfig with only margin
func NewMargin(top, right, bottom, left int) SpacingConfig {
	return SpacingConfig{
		MarginTop:    top,
		MarginRight:  right,
		MarginBottom: bottom,
		MarginLeft:   left,
	}
}

// BorderStyle creates a lipgloss border style with Dracula theme colors
type BorderStyle struct {
	Border       BorderSet
	Color        lipgloss.Color
	Style        lipgloss.Style
	Spacing      SpacingConfig
	Width        int
	Height       int
	Title        string
	TitleAlign   lipgloss.Position
	Focused      bool
	draculaTheme *theme.Theme
}

// NewBorderStyle creates a new border style with Dracula theme
func NewBorderStyle(t *theme.Theme) *BorderStyle {
	return &BorderStyle{
		Border:       RoundedBorder,
		Color:        t.Colors.Comment,
		Spacing:      NewSpacing(0, 0),
		TitleAlign:   lipgloss.Left,
		Focused:      false,
		draculaTheme: t,
	}
}

// WithBorder sets the border character set
func (b *BorderStyle) WithBorder(border BorderSet) *BorderStyle {
	b.Border = border
	return b
}

// WithColor sets the border color
func (b *BorderStyle) WithColor(color lipgloss.Color) *BorderStyle {
	b.Color = color
	return b
}

// WithSpacing sets the spacing configuration
func (b *BorderStyle) WithSpacing(spacing SpacingConfig) *BorderStyle {
	b.Spacing = spacing
	return b
}

// WithSize sets the width and height
func (b *BorderStyle) WithSize(width, height int) *BorderStyle {
	b.Width = width
	b.Height = height
	return b
}

// WithTitle sets the title and its alignment
func (b *BorderStyle) WithTitle(title string, align lipgloss.Position) *BorderStyle {
	b.Title = title
	b.TitleAlign = align
	return b
}

// WithFocus sets whether the border is focused
func (b *BorderStyle) WithFocus(focused bool) *BorderStyle {
	b.Focused = focused
	if focused {
		b.Color = b.draculaTheme.Colors.Purple
	} else {
		b.Color = b.draculaTheme.Colors.Comment
	}
	return b
}

// Render applies the border and spacing to content
func (b *BorderStyle) Render(content string) string {
	// Create base style with border
	style := lipgloss.NewStyle().
		Border(lipgloss.Border{
			Top:         b.Border.Top,
			Bottom:      b.Border.Bottom,
			Left:        b.Border.Left,
			Right:       b.Border.Right,
			TopLeft:     b.Border.TopLeft,
			TopRight:    b.Border.TopRight,
			BottomLeft:  b.Border.BottomLeft,
			BottomRight: b.Border.BottomRight,
		}).
		BorderForeground(b.Color).
		Padding(
			b.Spacing.PaddingTop,
			b.Spacing.PaddingRight,
			b.Spacing.PaddingBottom,
			b.Spacing.PaddingLeft,
		)

	// Add size constraints if specified
	if b.Width > 0 {
		style = style.Width(b.Width)
	}
	if b.Height > 0 {
		style = style.Height(b.Height)
	}

	// Add title if present
	if b.Title != "" {
		style = style.BorderTop(true).BorderTopForeground(b.Color)
	}

	rendered := style.Render(content)

	// Add title overlay if present
	if b.Title != "" {
		rendered = b.addTitle(rendered)
	}

	// Add margins
	if b.Spacing.MarginTop > 0 {
		rendered = strings.Repeat("\n", b.Spacing.MarginTop) + rendered
	}
	if b.Spacing.MarginBottom > 0 {
		rendered = rendered + strings.Repeat("\n", b.Spacing.MarginBottom)
	}
	if b.Spacing.MarginLeft > 0 {
		margin := strings.Repeat(" ", b.Spacing.MarginLeft)
		lines := strings.Split(rendered, "\n")
		for i, line := range lines {
			lines[i] = margin + line
		}
		rendered = strings.Join(lines, "\n")
	}

	return rendered
}

// addTitle overlays the title on the top border
func (b *BorderStyle) addTitle(bordered string) string {
	lines := strings.Split(bordered, "\n")
	if len(lines) == 0 {
		return bordered
	}

	// Style the title
	titleStyle := lipgloss.NewStyle().
		Foreground(b.draculaTheme.Colors.Purple).
		Bold(true)
	styledTitle := " " + titleStyle.Render(b.Title) + " "

	// Replace part of the top border with the title
	topLine := lines[0]
	titleLen := lipgloss.Width(styledTitle)

	// Calculate position based on alignment
	var position int
	lineWidth := lipgloss.Width(topLine)

	switch b.TitleAlign {
	case lipgloss.Left:
		position = 2
	case lipgloss.Center:
		position = (lineWidth - titleLen) / 2
	case lipgloss.Right:
		position = lineWidth - titleLen - 2
	default:
		position = 2
	}

	if position < 0 {
		position = 2
	}

	// Insert title into top border
	if position+titleLen < lineWidth {
		runes := []rune(topLine)
		result := string(runes[:position]) + styledTitle
		if position+titleLen < len(runes) {
			result += string(runes[position+titleLen:])
		}
		lines[0] = result
	}

	return strings.Join(lines, "\n")
}

// Separator creates a horizontal separator line
func Separator(width int, color lipgloss.Color) string {
	if width <= 0 {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(strings.Repeat("─", width))
}

// VerticalSeparator creates a vertical separator
func VerticalSeparator(height int, color lipgloss.Color) string {
	if height <= 0 {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(color)
	lines := make([]string, height)
	for i := range lines {
		lines[i] = style.Render("│")
	}
	return strings.Join(lines, "\n")
}

// Box creates a simple box around content
func Box(content string, t *theme.Theme) string {
	return NewBorderStyle(t).Render(content)
}

// FocusedBox creates a focused box around content
func FocusedBox(content string, t *theme.Theme) string {
	return NewBorderStyle(t).WithFocus(true).Render(content)
}

// TitledBox creates a box with a title
func TitledBox(content, title string, t *theme.Theme) string {
	return NewBorderStyle(t).
		WithTitle(title, lipgloss.Left).
		Render(content)
}

// PaddedContent adds padding to content without borders
func PaddedContent(content string, padding SpacingConfig) string {
	style := lipgloss.NewStyle().
		Padding(
			padding.PaddingTop,
			padding.PaddingRight,
			padding.PaddingBottom,
			padding.PaddingLeft,
		)
	return style.Render(content)
}

// JoinHorizontal joins multiple strings horizontally with spacing
func JoinHorizontal(spacing int, elements ...string) string {
	if len(elements) == 0 {
		return ""
	}
	spacer := strings.Repeat(" ", spacing)
	return strings.Join(elements, spacer)
}

// JoinVertical joins multiple strings vertically with spacing
func JoinVertical(spacing int, elements ...string) string {
	if len(elements) == 0 {
		return ""
	}
	spacer := strings.Repeat("\n", spacing+1)
	return strings.Join(elements, spacer)
}

// PlaceHorizontal positions content horizontally within a width
func PlaceHorizontal(width int, pos lipgloss.Position, content string) string {
	style := lipgloss.NewStyle().Width(width).Align(pos)
	return style.Render(content)
}

// PlaceVertical positions content vertically within a height
func PlaceVertical(height int, pos lipgloss.Position, content string) string {
	style := lipgloss.NewStyle().Height(height).AlignVertical(pos)
	return style.Render(content)
}

// MaxWidth constrains content to a maximum width with word wrapping
func MaxWidth(width int, content string) string {
	if width <= 0 {
		return content
	}
	style := lipgloss.NewStyle().MaxWidth(width)
	return style.Render(content)
}
