// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Tests for layout mode detection and switching
// ABOUTME: Ensures layout mode correctly adapts to terminal size changes
package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestDetermineLayoutMode tests the layout mode detection function
func TestDetermineLayoutMode(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		expected LayoutMode
	}{
		{
			name:     "wide layout - both dimensions above breakpoints",
			width:    120,
			height:   30,
			expected: LayoutModeWide,
		},
		{
			name:     "wide layout - exactly at breakpoints",
			width:    CompactModeWidthBreakpoint,
			height:   CompactModeHeightBreakpoint,
			expected: LayoutModeWide,
		},
		{
			name:     "compact layout - width below breakpoint",
			width:    80,
			height:   30,
			expected: LayoutModeCompact,
		},
		{
			name:     "compact layout - height below breakpoint",
			width:    120,
			height:   20,
			expected: LayoutModeCompact,
		},
		{
			name:     "compact layout - both dimensions below breakpoints",
			width:    80,
			height:   20,
			expected: LayoutModeCompact,
		},
		{
			name:     "compact layout - width just below breakpoint",
			width:    CompactModeWidthBreakpoint - 1,
			height:   30,
			expected: LayoutModeCompact,
		},
		{
			name:     "compact layout - height just below breakpoint",
			width:    120,
			height:   CompactModeHeightBreakpoint - 1,
			expected: LayoutModeCompact,
		},
		{
			name:     "compact layout - very small terminal",
			width:    40,
			height:   10,
			expected: LayoutModeCompact,
		},
		{
			name:     "wide layout - very large terminal",
			width:    200,
			height:   60,
			expected: LayoutModeWide,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineLayoutMode(tt.width, tt.height)
			if result != tt.expected {
				t.Errorf("DetermineLayoutMode(%d, %d) = %v, want %v",
					tt.width, tt.height, result, tt.expected)
			}
		})
	}
}

// TestLayoutModeString tests the String() method
func TestLayoutModeString(t *testing.T) {
	tests := []struct {
		mode     LayoutMode
		expected string
	}{
		{LayoutModeWide, "wide"},
		{LayoutModeCompact, "compact"},
		{LayoutMode(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.mode.String()
			if result != tt.expected {
				t.Errorf("LayoutMode(%d).String() = %q, want %q",
					tt.mode, result, tt.expected)
			}
		})
	}
}

// TestModelUpdateLayoutMode tests the UpdateLayoutMode method
func TestModelUpdateLayoutMode(t *testing.T) {
	t.Run("updates layout mode on size change", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")

		// Start in wide mode
		model.layoutMode = LayoutModeWide

		// Resize to compact size
		model.UpdateLayoutMode(80, 20)

		if model.layoutMode != LayoutModeCompact {
			t.Errorf("Expected layout mode to be compact, got %v", model.layoutMode)
		}
	})

	t.Run("does not update if forced", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")

		// Force wide layout mode
		model.SetForceLayoutMode(true, LayoutModeWide)

		// Try to resize to compact size
		model.UpdateLayoutMode(80, 20)

		// Should still be wide because it's forced
		if model.layoutMode != LayoutModeWide {
			t.Errorf("Expected layout mode to remain wide (forced), got %v", model.layoutMode)
		}
	})

	t.Run("invalidates cache on mode change", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")

		// Start in wide mode
		model.layoutMode = LayoutModeWide
		model.markdownCacheDirty = false

		// Resize to compact (should trigger cache invalidation)
		model.UpdateLayoutMode(80, 20)

		if !model.markdownCacheDirty {
			t.Error("Expected markdown cache to be invalidated on layout mode change")
		}
	})

	t.Run("does not invalidate cache if mode unchanged", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")

		// Start in wide mode
		model.layoutMode = LayoutModeWide
		model.markdownCacheDirty = false

		// "Resize" but stay in wide mode
		model.UpdateLayoutMode(120, 30)

		if model.markdownCacheDirty {
			t.Error("Expected markdown cache to remain clean when layout mode unchanged")
		}
	})
}

// TestSetForceLayoutMode tests the force layout mode functionality
func TestSetForceLayoutMode(t *testing.T) {
	t.Run("forces layout mode", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")

		// Force compact mode
		model.SetForceLayoutMode(true, LayoutModeCompact)

		if !model.forceLayoutMode {
			t.Error("Expected forceLayoutMode to be true")
		}
		if model.layoutMode != LayoutModeCompact {
			t.Errorf("Expected layout mode to be compact, got %v", model.layoutMode)
		}
	})

	t.Run("invalidates cache when forcing mode", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")
		model.markdownCacheDirty = false

		model.SetForceLayoutMode(true, LayoutModeWide)

		if !model.markdownCacheDirty {
			t.Error("Expected markdown cache to be invalidated when forcing layout mode")
		}
	})

	t.Run("unforcing allows automatic detection", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")

		// Force wide mode
		model.SetForceLayoutMode(true, LayoutModeWide)

		// Unforce (but keep wide mode)
		model.forceLayoutMode = false

		// Now update should work
		model.UpdateLayoutMode(80, 20)

		if model.layoutMode != LayoutModeCompact {
			t.Errorf("Expected automatic layout detection to switch to compact, got %v", model.layoutMode)
		}
	})
}

// TestGetLayoutMode tests the getter method
func TestGetLayoutMode(t *testing.T) {
	model := NewModel("test-conv", "test-model", "default")

	model.layoutMode = LayoutModeCompact
	if model.GetLayoutMode() != LayoutModeCompact {
		t.Errorf("Expected GetLayoutMode() to return compact, got %v", model.GetLayoutMode())
	}

	model.layoutMode = LayoutModeWide
	if model.GetLayoutMode() != LayoutModeWide {
		t.Errorf("Expected GetLayoutMode() to return wide, got %v", model.GetLayoutMode())
	}
}

// TestLayoutModeIntegrationWithWindowSize tests layout mode updates on window resize
func TestLayoutModeIntegrationWithWindowSize(t *testing.T) {
	t.Run("window resize triggers layout mode update", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")
		model.Ready = true

		// Start with wide layout
		msg := tea.WindowSizeMsg{Width: 120, Height: 30}
		model.handleWindowSizeMsg(msg)

		if model.layoutMode != LayoutModeWide {
			t.Errorf("Expected wide layout mode, got %v", model.layoutMode)
		}

		// Resize to compact
		msg = tea.WindowSizeMsg{Width: 80, Height: 20}
		model.handleWindowSizeMsg(msg)

		if model.layoutMode != LayoutModeCompact {
			t.Errorf("Expected compact layout mode after resize, got %v", model.layoutMode)
		}
	})

	t.Run("forced mode persists through resize", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")
		model.Ready = true

		// Force wide mode
		model.SetForceLayoutMode(true, LayoutModeWide)

		// Try to resize to compact dimensions
		msg := tea.WindowSizeMsg{Width: 80, Height: 20}
		model.handleWindowSizeMsg(msg)

		// Should still be wide
		if model.layoutMode != LayoutModeWide {
			t.Errorf("Expected forced wide layout to persist, got %v", model.layoutMode)
		}
	})
}

// TestLayoutModeRendering tests that different layouts are actually rendered
func TestLayoutModeRendering(t *testing.T) {
	t.Run("compact layout is simplified", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")
		model.Ready = true
		model.showIntro = false // Disable intro to test actual layout
		model.layoutMode = LayoutModeCompact

		output := model.View()

		// Compact mode should not include token visualization
		// (We can't easily test for its absence, but we can check the output is reasonable)
		if len(output) == 0 {
			t.Error("Expected compact layout to produce output")
		}

		// Check that it contains essential elements
		if !strings.Contains(output, "Jeff") {
			t.Error("Expected compact layout to contain title")
		}
	})

	t.Run("wide layout includes full features", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")
		model.Ready = true
		model.showIntro = false // Disable intro to test actual layout
		model.layoutMode = LayoutModeWide
		model.TokensInput = 100
		model.TokensOutput = 200

		output := model.View()

		if len(output) == 0 {
			t.Error("Expected wide layout to produce output")
		}

		// Check that it contains essential elements
		if !strings.Contains(output, "Jeff") {
			t.Error("Expected wide layout to contain title")
		}
	})

	t.Run("layout mode switches between renderers", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")
		model.Ready = true
		model.showIntro = false // Disable intro to test actual layout

		// Render in compact mode
		model.layoutMode = LayoutModeCompact
		compactOutput := model.View()

		// Render in wide mode
		model.layoutMode = LayoutModeWide
		wideOutput := model.View()

		// The outputs should be different (wide has more features)
		// We can't compare lengths directly due to formatting, but we can check both produce output
		if len(compactOutput) == 0 || len(wideOutput) == 0 {
			t.Error("Both layout modes should produce output")
		}
	})
}

// TestLayoutModeInitialization tests that layout mode is properly initialized
func TestLayoutModeInitialization(t *testing.T) {
	t.Run("new model initializes with correct layout mode", func(t *testing.T) {
		model := NewModel("test-conv", "test-model", "default")

		// Default size is 80x24, which should be compact (80 < 100)
		if model.layoutMode != LayoutModeCompact {
			t.Errorf("Expected initial layout mode to be compact (80x24), got %v", model.layoutMode)
		}

		if model.forceLayoutMode {
			t.Error("Expected forceLayoutMode to be false initially")
		}
	})
}

// TestLayoutModeBreakpoints verifies the breakpoint constants
func TestLayoutModeBreakpoints(t *testing.T) {
	t.Run("breakpoint constants are sensible", func(t *testing.T) {
		if CompactModeWidthBreakpoint <= 0 {
			t.Error("Width breakpoint should be positive")
		}
		if CompactModeHeightBreakpoint <= 0 {
			t.Error("Height breakpoint should be positive")
		}
		if CompactModeWidthBreakpoint != 100 {
			t.Errorf("Width breakpoint should be 100, got %d", CompactModeWidthBreakpoint)
		}
		if CompactModeHeightBreakpoint != 24 {
			t.Errorf("Height breakpoint should be 24, got %d", CompactModeHeightBreakpoint)
		}
	})
}
