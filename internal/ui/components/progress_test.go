// ABOUTME: Test suite for progress bar component
// ABOUTME: Tests progress rendering and value updates
package components

import (
	"testing"

	"github.com/harper/jeff/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestNewProgress(t *testing.T) {
	theme := themes.NewDracula()
	label := "Processing"

	progress := NewProgress(theme, label)

	assert.NotNil(t, progress)
	assert.Equal(t, theme, progress.theme)
	assert.Equal(t, label, progress.label)
	assert.Equal(t, 0.0, progress.value)
}

func TestProgressSetValue(t *testing.T) {
	theme := themes.NewDracula()
	progress := NewProgress(theme, "Loading")

	progress.SetValue(0.5)
	assert.Equal(t, 0.5, progress.value)

	progress.SetValue(1.0)
	assert.Equal(t, 1.0, progress.value)
}

func TestProgressView(t *testing.T) {
	theme := themes.NewDracula()
	progress := NewProgress(theme, "Upload")
	progress.SetValue(0.75)

	view := progress.View()
	assert.NotEmpty(t, view)
}
