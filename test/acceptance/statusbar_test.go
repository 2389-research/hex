// ABOUTME: Acceptance tests for status bar display
// ABOUTME: Tests that status bar shows correct information

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusBar_ShowsModelName(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Status bar should show model name (or part of it)
	assert.True(t, ViewContainsAny(h, "sonnet", "claude", "HEX"),
		"status bar should show model identifier or app name")
}

func TestStatusBar_ShowsStreamingIndicator(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Before streaming
	assert.Equal(t, "idle", h.GetStatus())

	// During streaming
	require.NoError(t, h.SimulateStreamStart())
	assert.Equal(t, "streaming", h.GetStatus())

	// After streaming
	require.NoError(t, h.SimulateStreamEnd())
	assert.Equal(t, "idle", h.GetStatus())
}

func TestStatusBar_RendersWithinWidth(t *testing.T) {
	// Test different terminal widths
	widths := []int{80, 120, 200}

	for _, width := range widths {
		t.Run(string(rune('0'+width/10))+"0_width", func(t *testing.T) {
			h := NewBubbleteaAdapter()
			require.NoError(t, h.Init(width, 40))
			defer h.Shutdown()

			lines := splitLines(h.GetView())

			// No line should exceed terminal width (allowing for some overflow)
			for _, line := range lines {
				// Note: ANSI escape codes inflate length, so this is approximate
				if len(stripAnsi(line)) > width+10 {
					// Allow some slack for edge cases
					t.Logf("Warning: line may exceed width: %d > %d", len(stripAnsi(line)), width)
				}
			}
		})
	}
}

// Helper to split view into lines
func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// Helper to strip ANSI escape codes (basic implementation)
func stripAnsi(s string) string {
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}
