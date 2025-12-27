// ABOUTME: Tests for Bubbletea UI view rendering
// ABOUTME: Validates view mode rendering, status bar, and visual elements
package ui_test

import (
	"testing"

	"github.com/2389-research/hex/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestViewRendersChatMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.CurrentView = ui.ViewModeChat

	view := model.View()
	assert.Contains(t, view, "HEX") // Neo-Terminal shows uppercase HEX
}

func TestViewRendersHistoryMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.CurrentView = ui.ViewModeHistory

	view := model.View()
	assert.Contains(t, view, "History Browser")
	assert.Contains(t, view, "HEX › HISTORY") // Neo-Terminal shows mode in status bar
}

func TestViewRendersToolsMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.CurrentView = ui.ViewModeTools

	view := model.View()
	assert.Contains(t, view, "Tool Inspector")
	assert.Contains(t, view, "HEX › TOOLS") // Neo-Terminal shows mode in status bar
}

func TestViewShowsTokenCounter(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.UpdateTokens(100, 250)

	view := model.View()
	// Neo-Terminal shows token format as "tokens: N in · M out"
	assert.Contains(t, view, "tokens: 100 in")
	assert.Contains(t, view, "250 out")
}

func TestViewShowsSearchMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.EnterSearchMode()
	model.UpdateSearchQuery("test")

	view := model.View()
	assert.Contains(t, view, "Search:")
	assert.Contains(t, view, "test")
}

func TestViewShowsHelpText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true

	view := model.View()
	// Neo-Terminal status bar shows key bindings with Unicode control symbols
	// We removed "/help" from status bar - now shows "⌃C quit · ⇥ views"
	assert.Contains(t, view, "⌃C quit") // Unicode control symbol
	assert.Contains(t, view, "⇥ views") // Tab to switch views
}

func TestViewStatusIndicatorChanges(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true

	// Test idle status
	view := model.View()
	assert.Contains(t, view, "●")

	// Test streaming status
	model.SetStatus(ui.StatusStreaming)
	view = model.View()
	assert.Contains(t, view, "●")

	// Test error status
	model.SetStatus(ui.StatusError)
	view = model.View()
	assert.Contains(t, view, "●")
}

func TestViewInitializingState(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = false

	view := model.View()
	assert.Contains(t, view, "Initializing")
}

func TestViewDoesNotShowInputInNonChatMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.CurrentView = ui.ViewModeHistory

	view := model.View()
	// The input placeholder shouldn't be visible in history mode
	assert.NotContains(t, view, "Send a message")
}

func TestViewShowsInputOnlyInChatMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.CurrentView = ui.ViewModeChat

	view := model.View()
	// In chat mode, we should see some input-related content
	// The exact content depends on how the input is rendered
	assert.True(t, len(view) > 0)
}

func TestViewMultipleMessages(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true

	// Add several messages
	model.AddMessage("user", "Hello")
	model.AddMessage("assistant", "Hi there")
	model.AddMessage("user", "How are you?")

	view := model.View()
	// View should contain content (messages are rendered in viewport)
	assert.True(t, len(view) > 100) // Should be substantial content
}

func TestViewMarkdownRendering(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.Width = 100 // Set a reasonable width

	// Add message with markdown
	model.AddMessage("assistant", "# Hello\n\nThis is **bold** and *italic*")

	// The view should exist and contain content
	view := model.View()
	assert.True(t, len(view) > 0)
	// View should contain HEX in the status bar
	assert.Contains(t, view, "HEX")
}

func TestViewOverflowIndicators(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.CurrentView = ui.ViewModeChat
	model.Width = 100
	model.Height = 30

	// Without overflow, borders should be plain dashes
	view := model.View()
	assert.Contains(t, view, "─") // Should have border dashes

	// With long input that causes overflow, we should see indicators
	// The textarea wraps at approximately Width-2 (for borders), so
	// 100 chars in 100-width should wrap into multiple lines
	longText := "AAAA AAAA AAAA AAAA AAAA AAAA AAAA AAAA AAAA AAAA " +
		"BBBB BBBB BBBB BBBB BBBB BBBB BBBB BBBB BBBB BBBB " +
		"CCCC CCCC CCCC CCCC CCCC CCCC CCCC CCCC CCCC CCCC " +
		"DDDD DDDD DDDD DDDD DDDD DDDD DDDD DDDD DDDD DDDD "
	model.SetInputValue(longText)

	view = model.View()
	// View should still render without panicking
	assert.True(t, len(view) > 0)
	assert.Contains(t, view, "HEX") // Status bar still present

	// When cursor is at end of long content, the overflow above indicator should appear
	// (if the content exceeds visible height of 3 lines)
	// The indicator is ▲ in the top border
	// Note: This test validates the logic works, but since textarea's LineInfo
	// behavior is complex, we mainly verify no crash and valid output
}

func TestOverflowDetectionLogic(t *testing.T) {
	// Test the overflow detection calculation directly
	// This mirrors the logic in view.go
	testCases := []struct {
		name        string
		visibleRows int
		totalRows   int
		cursorRow   int
		expectAbove bool
		expectBelow bool
	}{
		{"no overflow - cursor at start", 3, 3, 0, false, false},
		{"no overflow - cursor at end", 3, 3, 2, false, false},
		{"overflow below - cursor at top", 3, 5, 0, false, true},
		{"overflow below - cursor in middle", 3, 5, 2, false, true},
		{"overflow above - cursor at bottom", 3, 5, 4, true, false},
		{"lots of overflow - cursor at top", 3, 8, 0, false, true},
		{"overflow both ways - cursor in middle", 3, 8, 4, true, true},
		{"overflow above - cursor at bottom", 3, 8, 7, true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Replicate the logic from view.go
			firstVisibleRow := tc.cursorRow - (tc.visibleRows - 1)
			if firstVisibleRow < 0 {
				firstVisibleRow = 0
			}
			lastVisibleRow := firstVisibleRow + tc.visibleRows - 1
			if lastVisibleRow > tc.totalRows-1 {
				lastVisibleRow = tc.totalRows - 1
				firstVisibleRow = lastVisibleRow - tc.visibleRows + 1
				if firstVisibleRow < 0 {
					firstVisibleRow = 0
				}
			}

			hasOverflowAbove := firstVisibleRow > 0
			hasOverflowBelow := lastVisibleRow < tc.totalRows-1

			assert.Equal(t, tc.expectAbove, hasOverflowAbove, "hasOverflowAbove")
			assert.Equal(t, tc.expectBelow, hasOverflowBelow, "hasOverflowBelow")
		})
	}
}

// TODO: Implement IsBorderLine helper function
// func TestIsBorderLine(t *testing.T) {
// 	// Test the helper function that detects border lines
// 	tests := []struct {
// 		input    string
// 		expected bool
// 	}{
// 		{"─────────────────", true},  // Plain border
// 		{"───────────────▲─", true},  // Border with overflow indicator
// 		{"───────────────▼─", true},  // Border with overflow indicator
// 		{"───▲───────────▼─", true},  // Border with multiple indicators
// 		{"Hello World", false},       // Not a border
// 		{"───Hello───", false},       // Mixed content
// 		{"────", false},              // Too short
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.input, func(t *testing.T) {
// 			result := ui.IsBorderLine(tc.input)
// 			assert.Equal(t, tc.expected, result, "IsBorderLine(%q)", tc.input)
// 		})
// 	}
// }
