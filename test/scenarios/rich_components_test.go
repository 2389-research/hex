// ABOUTME: Scenario test for rich inline components (tables, progress)
// ABOUTME: Validates component rendering in chat conversation flow
package scenarios

import (
	"testing"

	"github.com/harper/jeff/internal/ui"
	"github.com/harper/jeff/internal/ui/components"
	"github.com/stretchr/testify/assert"
)

func TestTableInChatMessage_RendersCorrectly(t *testing.T) {
	// Setup model with Dracula theme
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")
	model.Ready = true

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
	model.UpdateViewport() // Manually update viewport to render messages

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
	model.Ready = true

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
	model.UpdateViewport() // Manually update viewport to render messages

	// Render view
	view := model.View()

	// Verify progress appears
	assert.Contains(t, view, "Uploading")
	assert.Contains(t, view, "65%")
}

func TestMultipleComponents_InDifferentMessages(t *testing.T) {
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")
	model.Ready = true
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

	model.UpdateViewport() // Manually update viewport to render messages

	// Render
	view := model.View()

	// Both components should render
	assert.Contains(t, view, "Task")
	assert.Contains(t, view, "Review PR")
	assert.Contains(t, view, "Processing")
	assert.Contains(t, view, "33%")
}
