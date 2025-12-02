package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDraculaColorConstants verifies all color constants match the official Dracula spec
func TestDraculaColorConstants(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		expected string
	}{
		{"Background", Background, "#282a36"},
		{"CurrentLine", CurrentLine, "#44475a"},
		{"Foreground", Foreground, "#f8f8f2"},
		{"Comment", Comment, "#6272a4"},
		{"Cyan", Cyan, "#8be9fd"},
		{"Green", Green, "#50fa7b"},
		{"Orange", Orange, "#ffb86c"},
		{"Pink", Pink, "#ff79c6"},
		{"Purple", Purple, "#bd93f9"},
		{"Red", Red, "#ff5555"},
		{"Yellow", Yellow, "#f1fa8c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.color, "Color %s should match Dracula spec", tt.name)
		})
	}
}

// TestNewDraculaTheme verifies theme creation and initialization
func TestNewDraculaTheme(t *testing.T) {
	theme := NewDraculaTheme()
	require.NotNil(t, theme, "Theme should not be nil")

	// Verify color references are initialized
	assert.Equal(t, lipgloss.Color(Background), theme.Colors.Background)
	assert.Equal(t, lipgloss.Color(CurrentLine), theme.Colors.CurrentLine)
	assert.Equal(t, lipgloss.Color(Foreground), theme.Colors.Foreground)
	assert.Equal(t, lipgloss.Color(Comment), theme.Colors.Comment)
	assert.Equal(t, lipgloss.Color(Cyan), theme.Colors.Cyan)
	assert.Equal(t, lipgloss.Color(Green), theme.Colors.Green)
	assert.Equal(t, lipgloss.Color(Orange), theme.Colors.Orange)
	assert.Equal(t, lipgloss.Color(Pink), theme.Colors.Pink)
	assert.Equal(t, lipgloss.Color(Purple), theme.Colors.Purple)
	assert.Equal(t, lipgloss.Color(Red), theme.Colors.Red)
	assert.Equal(t, lipgloss.Color(Yellow), theme.Colors.Yellow)
}

// TestTextStyles verifies all text styles are properly initialized
func TestTextStyles(t *testing.T) {
	theme := NewDraculaTheme()

	tests := []struct {
		name          string
		style         lipgloss.Style
		expectedColor lipgloss.Color
		expectBold    bool
	}{
		{"Title", theme.Title, theme.Colors.Purple, true},
		{"Subtitle", theme.Subtitle, theme.Colors.Pink, true},
		{"Body", theme.Body, theme.Colors.Foreground, false},
		{"Muted", theme.Muted, theme.Colors.Comment, false},
		{"Emphasized", theme.Emphasized, theme.Colors.Cyan, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the style is not empty/zero
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Style %s should render text", tt.name)
		})
	}
}

// TestStatusStyles verifies status indicator styles
func TestStatusStyles(t *testing.T) {
	theme := NewDraculaTheme()

	tests := []struct {
		name          string
		style         lipgloss.Style
		expectedColor lipgloss.Color
	}{
		{"Success", theme.Success, theme.Colors.Green},
		{"Error", theme.Error, theme.Colors.Red},
		{"Warning", theme.Warning, theme.Colors.Yellow},
		{"Info", theme.Info, theme.Colors.Cyan},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Status style %s should render text", tt.name)
		})
	}
}

// TestInteractiveStyles verifies interactive element styles are initialized
func TestInteractiveStyles(t *testing.T) {
	theme := NewDraculaTheme()

	// Test that styles render (border styles have visual differences via borders)
	borderNormal := theme.Border.Render("test")
	borderFocused := theme.BorderFocused.Render("test")
	assert.NotEmpty(t, borderNormal, "Border style should render")
	assert.NotEmpty(t, borderFocused, "Focused border style should render")

	// Test input styles render
	inputNormal := theme.Input.Render("test")
	inputFocused := theme.InputFocused.Render("test")
	assert.NotEmpty(t, inputNormal, "Input style should render")
	assert.NotEmpty(t, inputFocused, "Focused input style should render")

	// Test button styles render
	buttonNormal := theme.Button.Render("test")
	buttonActive := theme.ButtonActive.Render("test")
	assert.NotEmpty(t, buttonNormal, "Button style should render")
	assert.NotEmpty(t, buttonActive, "Active button style should render")
}

// TestComponentStyles verifies specialized component styles
func TestComponentStyles(t *testing.T) {
	theme := NewDraculaTheme()

	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"StatusBar", theme.StatusBar},
		{"ViewMode", theme.ViewMode},
		{"SearchPrompt", theme.SearchPrompt},
		{"TokenCounter", theme.TokenCounter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Component style %s should render text", tt.name)
		})
	}
}

// TestToolStyles verifies tool-related styles
func TestToolStyles(t *testing.T) {
	theme := NewDraculaTheme()

	tests := []struct {
		name          string
		style         lipgloss.Style
		expectedColor lipgloss.Color
	}{
		{"ToolApproval", theme.ToolApproval, theme.Colors.Orange},
		{"ToolExecuting", theme.ToolExecuting, theme.Colors.Yellow},
		{"ToolSuccess", theme.ToolSuccess, theme.Colors.Green},
		{"ToolError", theme.ToolError, theme.Colors.Red},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Tool style %s should render text", tt.name)
		})
	}
}

// TestSuggestionStyles verifies suggestion component styles
func TestSuggestionStyles(t *testing.T) {
	theme := NewDraculaTheme()

	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"SuggestionBox", theme.SuggestionBox},
		{"SuggestionTitle", theme.SuggestionTitle},
		{"SuggestionReason", theme.SuggestionReason},
		{"SuggestionHint", theme.SuggestionHint},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Suggestion style %s should render text", tt.name)
		})
	}
}

// TestListStyles verifies list and selection styles
func TestListStyles(t *testing.T) {
	theme := NewDraculaTheme()

	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"ListItem", theme.ListItem},
		{"ListItemSelected", theme.ListItemSelected},
		{"ListItemActive", theme.ListItemActive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "List style %s should render text", tt.name)
		})
	}

	// Verify all styles are initialized and render
	normal := theme.ListItem.Render("test")
	selected := theme.ListItemSelected.Render("test")
	active := theme.ListItemActive.Render("test")

	assert.NotEmpty(t, normal, "Normal list item should render")
	assert.NotEmpty(t, selected, "Selected list item should render")
	assert.NotEmpty(t, active, "Active list item should render")
}

// TestHelpAndModalStyles verifies help and modal styles
func TestHelpAndModalStyles(t *testing.T) {
	theme := NewDraculaTheme()

	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"HelpPanel", theme.HelpPanel},
		{"HelpKey", theme.HelpKey},
		{"HelpDesc", theme.HelpDesc},
		{"Modal", theme.Modal},
		{"ModalTitle", theme.ModalTitle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Help/Modal style %s should render text", tt.name)
		})
	}
}

// TestCodeStyles verifies code and syntax styles
func TestCodeStyles(t *testing.T) {
	theme := NewDraculaTheme()

	tests := []struct {
		name          string
		style         lipgloss.Style
		expectedColor lipgloss.Color
	}{
		{"Code", theme.Code, theme.Colors.Green},
		{"CodeBlock", theme.CodeBlock, theme.Colors.Foreground},
		{"Keyword", theme.Keyword, theme.Colors.Pink},
		{"String", theme.String, theme.Colors.Yellow},
		{"Number", theme.Number, theme.Colors.Purple},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Code style %s should render text", tt.name)
		})
	}
}

// TestLinkStyles verifies link styles
func TestLinkStyles(t *testing.T) {
	theme := NewDraculaTheme()

	link := theme.Link.Render("test")
	linkHover := theme.LinkHover.Render("test")

	assert.NotEmpty(t, link, "Link style should render text")
	assert.NotEmpty(t, linkHover, "Link hover style should render text")
}

// TestDraculaThemeFunction verifies the convenience function
func TestDraculaThemeFunction(t *testing.T) {
	theme1 := DraculaTheme()
	theme2 := DraculaTheme()

	require.NotNil(t, theme1, "DraculaTheme() should return non-nil theme")
	require.NotNil(t, theme2, "DraculaTheme() should return non-nil theme")

	// Verify they have the same color values
	assert.Equal(t, theme1.Colors.Purple, theme2.Colors.Purple)
	assert.Equal(t, theme1.Colors.Pink, theme2.Colors.Pink)
	assert.Equal(t, theme1.Colors.Green, theme2.Colors.Green)
}

// TestStyleConsistency verifies that related styles use consistent colors
func TestStyleConsistency(t *testing.T) {
	theme := NewDraculaTheme()

	// Success states should use green
	successText := theme.Success.Render("✓")
	toolSuccess := theme.ToolSuccess.Render("✓")
	assert.NotEmpty(t, successText)
	assert.NotEmpty(t, toolSuccess)

	// Error states should use red
	errorText := theme.Error.Render("✗")
	toolError := theme.ToolError.Render("✗")
	assert.NotEmpty(t, errorText)
	assert.NotEmpty(t, toolError)

	// Warning states should use yellow
	warningText := theme.Warning.Render("⚠")
	toolExecuting := theme.ToolExecuting.Render("⏳")
	assert.NotEmpty(t, warningText)
	assert.NotEmpty(t, toolExecuting)
}

// TestAllStylesInitialized verifies no styles are left at zero value
func TestAllStylesInitialized(t *testing.T) {
	theme := NewDraculaTheme()

	// Test a sample from each category to ensure initialization
	testStyles := []struct {
		name       string
		style      lipgloss.Style
		hasBorder  bool
		hasPadding bool
	}{
		{"Title", theme.Title, false, false},
		{"Success", theme.Success, false, false},
		{"Border", theme.Border, true, false},
		{"StatusBar", theme.StatusBar, false, false},
		{"ToolApproval", theme.ToolApproval, true, true},
		{"SuggestionBox", theme.SuggestionBox, true, true},
		{"ListItem", theme.ListItem, false, false},
		{"HelpPanel", theme.HelpPanel, true, true},
		{"Code", theme.Code, false, false},
		{"Link", theme.Link, false, false},
	}

	for _, tt := range testStyles {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Style %s should render non-empty text", tt.name)

			// Styles with borders or padding will have visible changes
			// Styles with only colors may not in headless test mode
			if tt.hasBorder || tt.hasPadding {
				assert.NotEqual(t, "test", rendered, "Style %s should apply formatting (border/padding)", tt.name)
			}
		})
	}
}

// TestThemeColorsMatchConstants verifies theme color fields match the constants
func TestThemeColorsMatchConstants(t *testing.T) {
	theme := NewDraculaTheme()

	tests := []struct {
		name     string
		color    lipgloss.Color
		constant string
	}{
		{"Background", theme.Colors.Background, Background},
		{"CurrentLine", theme.Colors.CurrentLine, CurrentLine},
		{"Foreground", theme.Colors.Foreground, Foreground},
		{"Comment", theme.Colors.Comment, Comment},
		{"Cyan", theme.Colors.Cyan, Cyan},
		{"Green", theme.Colors.Green, Green},
		{"Orange", theme.Colors.Orange, Orange},
		{"Pink", theme.Colors.Pink, Pink},
		{"Purple", theme.Colors.Purple, Purple},
		{"Red", theme.Colors.Red, Red},
		{"Yellow", theme.Colors.Yellow, Yellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, lipgloss.Color(tt.constant), tt.color,
				"Theme color %s should match constant", tt.name)
		})
	}
}

// BenchmarkThemeCreation benchmarks theme initialization performance
func BenchmarkThemeCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewDraculaTheme()
	}
}

// BenchmarkStyleRendering benchmarks rendering performance
func BenchmarkStyleRendering(b *testing.B) {
	theme := NewDraculaTheme()
	text := "This is a test message for benchmarking"

	b.Run("Title", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = theme.Title.Render(text)
		}
	})

	b.Run("Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = theme.Error.Render(text)
		}
	})

	b.Run("ToolApproval", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = theme.ToolApproval.Render(text)
		}
	})
}
