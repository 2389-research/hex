// ABOUTME: Test suite for Huh-based quick actions selector
// ABOUTME: Validates quick action selection UI rendering and interaction
package components

import (
	"testing"

	"github.com/harper/pagent/internal/ui/themes"
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
	qa.Init()
	view := qa.View()

	// View should render with help text
	assert.NotEmpty(t, view)
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
