// ABOUTME: Test suite for Huh-based tool approval component
// ABOUTME: Ensures approval forms render correctly with theme colors
package components

import (
	"testing"

	"github.com/harper/pagent/internal/ui/themes"
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

	// Initialize the form (required for proper rendering)
	approval.Init()

	view := approval.View()

	// View should render (may not show full content until first update)
	// Huh forms render help text at minimum
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "submit")
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
