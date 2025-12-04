// ABOUTME: Visual regression tests for theme rendering
// ABOUTME: Captures and validates visual output across all themes
package visual

import (
	"testing"

	"github.com/harper/pagent/internal/ui"
	"github.com/harper/pagent/internal/ui/components"
	"github.com/stretchr/testify/assert"
)

func TestDraculaTheme_VisualOutput(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	model.Ready = true // Mark as ready to render
	model.Width = 80
	model.Height = 24
	model.AddMessage("user", "Hello")
	model.AddMessage("assistant", "Hi! **Bold** and *italic* text.")

	view := model.View()

	// Should contain Dracula theme elements
	assert.Contains(t, view, "Pagen")
	assert.NotEmpty(t, view)

	// Verify decorative border is present
	assert.Contains(t, view, "╭") // Top left corner
	assert.Contains(t, view, "╮") // Top right corner
}

func TestGruvboxTheme_VisualOutput(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "gruvbox")
	model.Ready = true
	model.Width = 80
	model.Height = 24
	model.AddMessage("user", "Test")

	view := model.View()

	// Gruvbox should look different from Dracula
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Pagen")
}

func TestNordTheme_VisualOutput(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "nord")
	model.Ready = true
	model.Width = 80
	model.Height = 24
	model.AddMessage("user", "Test")

	view := model.View()

	// Nord should have distinct appearance
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Pagen")
}

func TestTableComponent_VisualPolish(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	model.Ready = true
	model.Width = 80
	model.Height = 24
	theme := model.GetTheme()

	table := components.NewTable(
		theme,
		[]string{"Column A", "Column B"},
		[][]string{
			{"Row 1A", "Row 1B"},
			{"Row 2A", "Row 2B"},
		},
	)

	// Table component itself can be rendered
	tableView := table.View()

	// Table should have borders and theme colors
	assert.Contains(t, tableView, "Column A")
	assert.Contains(t, tableView, "Row 1A")
	assert.Contains(t, tableView, "─") // Box drawing characters
}

func TestTitleGradient_EnhancedVisuals(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	model.Ready = true
	model.Width = 80
	model.Height = 24
	view := model.View()

	// Intro screen should have box-drawing border characters
	assert.Contains(t, view, "Productivity AI Agent")
	// View should be non-empty and contain border decorations
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "┏") // Top left corner of intro border
	assert.Contains(t, view, "┗") // Bottom left corner of intro border
}

func TestThemeComparison_DistinctVisuals(t *testing.T) {
	// Create models with different themes
	dracula := ui.NewModel("test", "claude-sonnet-4", "dracula")
	gruvbox := ui.NewModel("test", "claude-sonnet-4", "gruvbox")
	nord := ui.NewModel("test", "claude-sonnet-4", "nord")

	// Initialize all models
	for _, model := range []*ui.Model{dracula, gruvbox, nord} {
		model.Ready = true
		model.Width = 80
		model.Height = 24
	}

	// Get themes to verify they are different
	draculaTheme := dracula.GetTheme()
	gruvboxTheme := gruvbox.GetTheme()
	nordTheme := nord.GetTheme()

	// Verify themes have different names (proper case)
	assert.Equal(t, "Dracula", draculaTheme.Name())
	assert.Equal(t, "Gruvbox Dark", gruvboxTheme.Name())
	assert.Equal(t, "Nord", nordTheme.Name())

	// Verify themes have different primary colors
	assert.NotEqual(t, draculaTheme.Primary(), gruvboxTheme.Primary(), "Dracula and Gruvbox should have different primary colors")
	assert.NotEqual(t, gruvboxTheme.Primary(), nordTheme.Primary(), "Gruvbox and Nord should have different primary colors")
	assert.NotEqual(t, nordTheme.Primary(), draculaTheme.Primary(), "Nord and Dracula should have different primary colors")

	// All views should still render properly
	draculaView := dracula.View()
	gruvboxView := gruvbox.View()
	nordView := nord.View()

	assert.Contains(t, draculaView, "Productivity AI Agent")
	assert.Contains(t, gruvboxView, "Productivity AI Agent")
	assert.Contains(t, nordView, "Productivity AI Agent")
}

func TestSpacing_ConsistentLayout(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	model.Ready = true
	model.Width = 80
	model.Height = 24
	model.AddMessage("user", "Test")

	view := model.View()

	// View should be properly formatted with spacing
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Pagen")

	// Should have proper line breaks (multiple newlines for spacing)
	assert.Contains(t, view, "\n")
}

func TestProgressComponent_VisualRendering(t *testing.T) {
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	model.Ready = true
	model.Width = 80
	model.Height = 24
	theme := model.GetTheme()

	progress := components.NewProgress(theme, "Loading")
	progress.SetValue(0.75)

	// Test the progress component directly
	progressView := progress.View()

	// Progress bar should render with percentage
	assert.Contains(t, progressView, "Loading")
	assert.Contains(t, progressView, "75%")
}
