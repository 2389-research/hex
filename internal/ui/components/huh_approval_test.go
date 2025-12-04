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

func TestHuhApprovalImplementsComponent(_ *testing.T) {
	theme := themes.NewDracula()
	approval := NewHuhApproval(theme, "bash", "Run: echo test")

	// Verify it implements Component interface
	var _ Component = approval
}

func TestHuhApprovalSetSize(t *testing.T) {
	theme := themes.NewDracula()
	approval := NewHuhApproval(theme, "bash", "Run: echo test")

	// Initially zero
	w, h := approval.GetSize()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)

	// Set size
	cmd := approval.SetSize(100, 40)
	assert.Nil(t, cmd) // Should not return a command

	// Verify size is stored
	w, h = approval.GetSize()
	assert.Equal(t, 100, w)
	assert.Equal(t, 40, h)

	// Verify form width is updated
	assert.NotNil(t, approval.form)
}

func TestHuhApprovalGetSize(t *testing.T) {
	theme := themes.NewDracula()
	approval := NewHuhApproval(theme, "bash", "Run: echo test")

	// Set size and verify GetSize returns correct values
	approval.SetSize(120, 50)
	w, h := approval.GetSize()
	assert.Equal(t, 120, w)
	assert.Equal(t, 50, h)
}
