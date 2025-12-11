package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolLogOverlay_IsFullscreen(t *testing.T) {
	lines := []string{"line 1", "line 2"}
	overlay := NewToolLogOverlay(&lines)

	assert.True(t, overlay.IsFullscreen())
}

func TestToolLogOverlay_GetDesiredHeight(t *testing.T) {
	lines := []string{}
	overlay := NewToolLogOverlay(&lines)

	// Fullscreen always wants max height
	assert.Equal(t, -1, overlay.GetDesiredHeight())
}

func TestToolLogOverlay_RefersToModelData(t *testing.T) {
	lines := []string{"initial"}
	overlay := NewToolLogOverlay(&lines)

	// Should reference lines, not copy
	lines = append(lines, "new line")

	content := overlay.GetContent()
	assert.Contains(t, content, "initial")
	assert.Contains(t, content, "new line")
}
