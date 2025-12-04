// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Layout mode detection and breakpoint logic for responsive UI
// ABOUTME: Determines whether to show compact or wide layout based on terminal size
package ui

// LayoutMode represents the current UI layout configuration
type LayoutMode int

const (
	// LayoutModeWide is used for terminals with sufficient space (>= 100 width, >= 24 height)
	// Displays full-featured UI with all components visible
	LayoutModeWide LayoutMode = iota

	// LayoutModeCompact is used for smaller terminals (< 100 width OR < 24 height)
	// Displays simplified UI with essential features only
	LayoutModeCompact
)

// String returns a human-readable representation of the layout mode
func (l LayoutMode) String() string {
	switch l {
	case LayoutModeWide:
		return "wide"
	case LayoutModeCompact:
		return "compact"
	default:
		return "unknown"
	}
}

const (
	// CompactModeWidthBreakpoint is the minimum width required for wide layout
	// Below this width, the UI switches to compact mode
	CompactModeWidthBreakpoint = 100

	// CompactModeHeightBreakpoint is the minimum height required for wide layout
	// Below this height, the UI switches to compact mode
	CompactModeHeightBreakpoint = 24
)

// DetermineLayoutMode determines the appropriate layout mode based on terminal dimensions
// Returns LayoutModeCompact if either dimension is below the breakpoint
// Returns LayoutModeWide if both dimensions meet or exceed their breakpoints
func DetermineLayoutMode(width, height int) LayoutMode {
	if width < CompactModeWidthBreakpoint || height < CompactModeHeightBreakpoint {
		return LayoutModeCompact
	}
	return LayoutModeWide
}
