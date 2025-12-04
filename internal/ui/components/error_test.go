// ABOUTME: Test suite for error display component
// ABOUTME: Ensures errors render correctly with themed styling and implements Component interface
package components

import (
	"testing"

	"github.com/harper/jeff/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestNewErrorDisplay(t *testing.T) {
	theme := themes.NewDracula()
	title := "Test Error"
	message := "Something went wrong"
	details := "Stack trace here"

	err := NewErrorDisplay(theme, title, message, details)

	assert.NotNil(t, err)
	assert.Equal(t, theme, err.theme)
	assert.Equal(t, title, err.title)
	assert.Equal(t, message, err.message)
	assert.Equal(t, details, err.details)
}

func TestErrorDisplayView(t *testing.T) {
	theme := themes.NewDracula()
	err := NewErrorDisplay(theme, "API Error", "Request failed", "HTTP 500")

	view := err.View()

	// View should render with error content
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "API Error")
	assert.Contains(t, view, "Request failed")
	assert.Contains(t, view, "HTTP 500")
}

func TestErrorDisplayImplementsComponent(_ *testing.T) {
	theme := themes.NewDracula()
	err := NewErrorDisplay(theme, "Test", "Test message", "Test details")

	// Verify it implements Component interface
	var _ Component = err
}

func TestErrorDisplaySetSize(t *testing.T) {
	theme := themes.NewDracula()
	err := NewErrorDisplay(theme, "Test", "Test message", "Test details")

	// Initially zero
	w, h := err.GetSize()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)

	// Set size
	cmd := err.SetSize(100, 40)
	assert.Nil(t, cmd) // Should not return a command

	// Verify size is stored
	w, h = err.GetSize()
	assert.Equal(t, 100, w)
	assert.Equal(t, 40, h)
}

func TestErrorDisplayGetSize(t *testing.T) {
	theme := themes.NewDracula()
	err := NewErrorDisplay(theme, "Test", "Test message", "Test details")

	// Set size and verify GetSize returns correct values
	err.SetSize(120, 50)
	w, h := err.GetSize()
	assert.Equal(t, 120, w)
	assert.Equal(t, 50, h)
}

func TestErrorDisplayInit(t *testing.T) {
	theme := themes.NewDracula()
	err := NewErrorDisplay(theme, "Test", "Test message", "Test details")

	cmd := err.Init()
	assert.Nil(t, cmd) // Should not return a command
}

func TestErrorDisplayUpdate(t *testing.T) {
	theme := themes.NewDracula()
	err := NewErrorDisplay(theme, "Test", "Test message", "Test details")

	// Update should return self with no command
	model, cmd := err.Update(nil)
	assert.Equal(t, err, model)
	assert.Nil(t, cmd)
}
