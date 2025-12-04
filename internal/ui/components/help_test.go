// ABOUTME: Test suite for help overlay component
// ABOUTME: Ensures help text renders correctly and implements Component interface
package components

import (
	"testing"

	"github.com/harper/pagent/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestNewHelpOverlay(t *testing.T) {
	theme := themes.NewDracula()
	help := NewHelpOverlay(theme)

	assert.NotNil(t, help)
	assert.Equal(t, theme, help.theme)
}

func TestHelpOverlayView(t *testing.T) {
	theme := themes.NewDracula()
	help := NewHelpOverlay(theme)

	view := help.View()

	// View should render with help content
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Keyboard Shortcuts")
	assert.Contains(t, view, "ctrl+c")
	assert.Contains(t, view, "Quit")
}

func TestHelpOverlayImplementsComponent(_ *testing.T) {
	theme := themes.NewDracula()
	help := NewHelpOverlay(theme)

	// Verify it implements Component interface
	var _ Component = help
}

func TestHelpOverlaySetSize(t *testing.T) {
	theme := themes.NewDracula()
	help := NewHelpOverlay(theme)

	// Initially zero
	w, h := help.GetSize()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)

	// Set size
	cmd := help.SetSize(100, 40)
	assert.Nil(t, cmd) // Should not return a command

	// Verify size is stored
	w, h = help.GetSize()
	assert.Equal(t, 100, w)
	assert.Equal(t, 40, h)
}

func TestHelpOverlayGetSize(t *testing.T) {
	theme := themes.NewDracula()
	help := NewHelpOverlay(theme)

	// Set size and verify GetSize returns correct values
	help.SetSize(120, 50)
	w, h := help.GetSize()
	assert.Equal(t, 120, w)
	assert.Equal(t, 50, h)
}

func TestHelpOverlayInit(t *testing.T) {
	theme := themes.NewDracula()
	help := NewHelpOverlay(theme)

	cmd := help.Init()
	assert.Nil(t, cmd) // Should not return a command
}

func TestHelpOverlayUpdate(t *testing.T) {
	theme := themes.NewDracula()
	help := NewHelpOverlay(theme)

	// Update should return self with no command
	model, cmd := help.Update(nil)
	assert.Equal(t, help, model)
	assert.Nil(t, cmd)
}
