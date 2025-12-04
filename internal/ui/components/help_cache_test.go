// Package components provides reusable UI components for the TUI.
// ABOUTME: Tests for Phase 1 Task 3 help overlay caching
// ABOUTME: Verifies cache behavior in HelpOverlay component
package components

import (
	"testing"

	"github.com/harper/jeff/internal/ui/themes"
)

// TestHelpOverlayCacheInitialization verifies cache starts empty
func TestHelpOverlayCacheInitialization(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)

	// Cache should start empty
	if help.cachedContent != "" {
		t.Error("Cached content should start empty")
	}

	// Dirty flag should start false (cache not yet built)
	if help.contentDirty {
		t.Error("Content dirty flag should start false")
	}
}

// TestHelpOverlayCachePopulation verifies cache is populated on first view
func TestHelpOverlayCachePopulation(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)
	help.SetSize(80, 24)

	// First view should populate cache
	view := help.View()

	// Cache should now have content
	if help.cachedContent == "" {
		t.Error("Cache should be populated after first view")
	}

	// Cached content should match returned view
	if help.cachedContent != view {
		t.Error("Cached content should match returned view")
	}

	// Dirty flag should be cleared after successful render
	if help.contentDirty {
		t.Error("Dirty flag should be cleared after view")
	}
}

// TestHelpOverlayCacheHitBehavior verifies subsequent views use cache
func TestHelpOverlayCacheHitBehavior(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)
	help.SetSize(80, 24)

	// First view
	view1 := help.View()
	cachedAfterFirst := help.cachedContent

	// Second view (should use cache)
	view2 := help.View()
	cachedAfterSecond := help.cachedContent

	// Results should be identical
	if view1 != view2 {
		t.Error("Cache hit should return identical content")
	}

	// Cached content should not change between calls
	if cachedAfterFirst != cachedAfterSecond {
		t.Error("Cached content should remain stable")
	}
}

// TestHelpOverlayCacheInvalidationOnWidthChange verifies width change invalidates cache
func TestHelpOverlayCacheInvalidationOnWidthChange(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)
	help.SetSize(80, 24)

	// Render and populate cache
	_ = help.View()

	// Verify cache is clean
	if help.contentDirty {
		t.Fatal("Cache should be clean after view")
	}

	// Change width
	help.SetSize(100, 24)

	// Cache should now be dirty
	if !help.contentDirty {
		t.Error("Cache should be dirty after width change")
	}
}

// TestHelpOverlayCacheInvalidationOnHeightChange verifies height change invalidates cache
func TestHelpOverlayCacheInvalidationOnHeightChange(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)
	help.SetSize(80, 24)

	// Render and populate cache
	_ = help.View()

	// Verify cache is clean
	if help.contentDirty {
		t.Fatal("Cache should be clean after view")
	}

	// Change height
	help.SetSize(80, 30)

	// Cache should now be dirty
	if !help.contentDirty {
		t.Error("Cache should be dirty after height change")
	}
}

// TestHelpOverlayCacheNoInvalidationOnSameSize verifies no invalidation when size doesn't change
func TestHelpOverlayCacheNoInvalidationOnSameSize(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)
	help.SetSize(80, 24)

	// Render and populate cache
	_ = help.View()

	// Verify cache is clean
	if help.contentDirty {
		t.Fatal("Cache should be clean after view")
	}

	// Set same size again
	help.SetSize(80, 24)

	// Cache should still be clean
	if help.contentDirty {
		t.Error("Cache should remain clean when size doesn't change")
	}
}

// TestHelpOverlayCacheRegenerationAfterInvalidation verifies cache regenerates after invalidation
func TestHelpOverlayCacheRegenerationAfterInvalidation(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)
	help.SetSize(80, 24)

	// First render
	view1 := help.View()

	// Change size to invalidate
	help.SetSize(100, 24)

	// Cache should be dirty
	if !help.contentDirty {
		t.Fatal("Cache should be dirty after size change")
	}

	// Second render (should regenerate)
	view2 := help.View()

	// Cache should be clean again
	if help.contentDirty {
		t.Error("Cache should be clean after regeneration")
	}

	// Cached content should be updated
	if help.cachedContent == "" {
		t.Error("Cache should have content after regeneration")
	}

	// Views should both have content (though may differ due to size)
	if view1 == "" || view2 == "" {
		t.Error("Both views should have content")
	}
}

// TestHelpOverlayMultipleResizes verifies cache behavior across multiple resizes
func TestHelpOverlayMultipleResizes(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)

	sizes := []struct {
		width  int
		height int
	}{
		{80, 24},
		{100, 30},
		{120, 40},
		{80, 24}, // Back to original
	}

	for i, size := range sizes {
		help.SetSize(size.width, size.height)
		view := help.View()

		if view == "" {
			t.Errorf("View %d: returned empty content", i)
		}

		// Cache should be clean after view
		if help.contentDirty {
			t.Errorf("View %d: cache should be clean after render", i)
		}

		// Cache should have content
		if help.cachedContent == "" {
			t.Errorf("View %d: cache should have content", i)
		}
	}
}

// TestHelpOverlayZeroSize verifies handling of zero/unset size
func TestHelpOverlayZeroSize(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)

	// Don't set size (starts at 0, 0)
	view := help.View()

	// Should still render (with no size constraints)
	if view == "" {
		t.Error("Should render content even with zero size")
	}

	// Cache should be populated
	if help.cachedContent == "" {
		t.Error("Cache should be populated")
	}
}

// TestHelpOverlaySetSizeZeroToNonZero verifies transition from zero to non-zero size
func TestHelpOverlaySetSizeZeroToNonZero(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)

	// Render at zero size
	_ = help.View()

	// Set non-zero size (should invalidate)
	help.SetSize(80, 24)

	// Cache should be dirty
	if !help.contentDirty {
		t.Error("Cache should be dirty after size change from zero")
	}

	// Render at new size
	view := help.View()

	if view == "" {
		t.Error("Should render at new size")
	}
}

// TestHelpOverlayCacheContentStructure verifies cached content has expected structure
func TestHelpOverlayCacheContentStructure(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := NewHelpOverlay(theme)
	help.SetSize(80, 24)

	view := help.View()

	// Cache should match view
	if help.cachedContent != view {
		t.Error("Cached content should match view output")
	}

	// Should not be empty
	if help.cachedContent == "" {
		t.Error("Cached content should not be empty")
	}

	// Should be a string (not nil/panic)
	_ = len(help.cachedContent)
}
