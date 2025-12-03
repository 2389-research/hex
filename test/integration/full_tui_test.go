// ABOUTME: Complete TUI integration test covering all phases
// ABOUTME: End-to-end test of theme, components, and interactions
package integration

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/pagent/internal/ui"
	"github.com/harper/pagent/internal/ui/components"
	"github.com/stretchr/testify/assert"
)

func TestFullTUI_AllPhasesIntegrated(t *testing.T) {
	// Phase 1: Theme system
	model := ui.NewModel("integration-test", "claude-sonnet-4", "dracula")
	assert.NotNil(t, model)

	// Initialize the UI with window size message
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Theme name is capitalized (Dracula, not dracula)
	assert.Equal(t, "Dracula", model.GetTheme().Name())

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

	// Verify messages were added
	assert.Len(t, model.Messages, 4) // 2 text + 2 component messages
	assert.Equal(t, "user", model.Messages[0].Role)
	assert.Equal(t, "Show me my emails", model.Messages[0].Content)

	// Verify components are embedded
	assert.NotNil(t, model.Messages[2].Component)
	assert.IsType(t, &components.Table{}, model.Messages[2].Component)
	assert.NotNil(t, model.Messages[3].Component)
	assert.IsType(t, &components.Progress{}, model.Messages[3].Component)

	// Render complete view
	view := model.View()

	// Verify basic UI elements are present
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Pagen")
	assert.Contains(t, view, "claude-sonnet-4")
}

func TestFullTUI_ThemeSwitching(t *testing.T) {
	// Create models with different themes
	dracula := ui.NewModel("test", "claude-sonnet-4", "dracula")
	gruvbox := ui.NewModel("test", "claude-sonnet-4", "gruvbox")
	nord := ui.NewModel("test", "claude-sonnet-4", "nord")

	// Initialize all models with window size
	dracula.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	gruvbox.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	nord.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Add a message to each to have content
	dracula.AddMessage("user", "Test message")
	gruvbox.AddMessage("user", "Test message")
	nord.AddMessage("user", "Test message")

	// Verify all models are initialized and ready
	draculaView := dracula.View()
	gruvboxView := gruvbox.View()
	nordView := nord.View()

	// All should render successfully
	assert.NotEmpty(t, draculaView)
	assert.NotEmpty(t, gruvboxView)
	assert.NotEmpty(t, nordView)

	// All should contain the title
	assert.Contains(t, draculaView, "Pagen")
	assert.Contains(t, gruvboxView, "Pagen")
	assert.Contains(t, nordView, "Pagen")
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

func TestFullTUI_HelpOverlay(t *testing.T) {
	// Test Phase 5: Help overlay
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	theme := model.GetTheme()

	helpOverlay := components.NewHelpOverlay(theme)
	view := helpOverlay.View()

	// Verify help content
	assert.Contains(t, view, "Keyboard Shortcuts")
	assert.Contains(t, view, "ctrl+c")
	assert.Contains(t, view, "Quit")
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "Send message")
}

func TestFullTUI_ErrorDisplay(t *testing.T) {
	// Test Phase 5: Error visualization
	model := ui.NewModel("test", "claude-sonnet-4", "dracula")
	theme := model.GetTheme()

	errorDisplay := components.NewErrorDisplay(
		theme,
		"Connection Failed",
		"Unable to connect to API",
		"Network timeout after 30 seconds",
	)
	view := errorDisplay.View()

	// Verify error content
	assert.Contains(t, view, "Connection Failed")
	assert.Contains(t, view, "Unable to connect to API")
	assert.Contains(t, view, "Network timeout")
}
