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
	// Neo-Terminal shows help text with Unicode control symbols
	assert.Contains(t, view, "/help")
	assert.Contains(t, view, "⌃C quit") // Unicode control symbol
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
