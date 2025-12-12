package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpOverlay_IsFullscreen(t *testing.T) {
	overlay := NewHelpOverlay()
	assert.True(t, overlay.IsFullscreen())
}

func TestHelpOverlay_GetContent(t *testing.T) {
	overlay := NewHelpOverlay()

	content := overlay.GetContent()
	assert.Contains(t, content, "Ctrl+O")
	assert.Contains(t, content, "Ctrl+H")
	assert.Contains(t, content, "Escape")
}
