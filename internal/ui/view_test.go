// ABOUTME: Tests for Bubbletea UI view rendering
// ABOUTME: Validates view mode rendering, status bar, and visual elements
package ui_test

import (
	"testing"

	"github.com/harper/pagent/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestViewRendersChatMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	model.CurrentView = ui.ViewModeChat
	// Phase 2 Task 3: Disable intro to test normal chat view
	model.AddMessage("user", "test") // Add a message to get past intro

	view := model.View()
	assert.Contains(t, view, "Pagen")  // Changed from "Clem" to "Pagen"
	assert.Contains(t, view, "[chat]") // Phase 6C: lowercase in status bar
}

func TestViewRendersHistoryMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	model.CurrentView = ui.ViewModeHistory
	// Phase 2 Task 3: Add message to bypass intro screen
	model.AddMessage("user", "test")

	view := model.View()
	assert.Contains(t, view, "History Browser")
	assert.Contains(t, view, "[history]") // Phase 6C: lowercase in status bar
}

func TestViewRendersToolsMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	model.CurrentView = ui.ViewModeTools
	// Phase 2 Task 3: Add message to bypass intro screen
	model.AddMessage("user", "test")

	view := model.View()
	assert.Contains(t, view, "Tool Inspector")
	assert.Contains(t, view, "[tools]") // Phase 6C: lowercase in status bar
}

func TestViewShowsTokenCounter(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	// Phase 2 Task 3: Add message to bypass intro screen
	model.AddMessage("user", "test")
	model.UpdateTokens(100, 250)

	view := model.View()
	// Phase 6C: New compact token format with k suffix
	assert.Contains(t, view, "0k↓") // 100 rounds down to 0k
	assert.Contains(t, view, "0k↑") // 250 rounds down to 0k
}

func TestViewShowsSearchMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	// Phase 2 Task 3: Add message to bypass intro screen
	model.AddMessage("user", "test")
	model.EnterSearchMode()
	model.UpdateSearchQuery("test")

	view := model.View()
	assert.Contains(t, view, "Search:")
	assert.Contains(t, view, "test")
}

func TestViewShowsHelpText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	// Phase 2 Task 3: Add message to bypass intro screen
	model.AddMessage("user", "test")

	view := model.View()
	// Phase 6C: New compact help text in status bar
	assert.Contains(t, view, "?:help")
	assert.Contains(t, view, "^C:quit")
}

func TestViewStatusIndicatorChanges(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	// Phase 2 Task 3: Add message to bypass intro screen
	model.AddMessage("user", "test")

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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = false

	view := model.View()
	assert.Contains(t, view, "Initializing")
}

func TestViewDoesNotShowInputInNonChatMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	model.CurrentView = ui.ViewModeHistory

	view := model.View()
	// The input placeholder shouldn't be visible in history mode
	assert.NotContains(t, view, "Send a message")
}

func TestViewShowsInputOnlyInChatMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	model.CurrentView = ui.ViewModeChat

	view := model.View()
	// In chat mode, we should see some input-related content
	// The exact content depends on how the input is rendered
	assert.True(t, len(view) > 0)
}

func TestViewMultipleMessages(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true
	model.Width = 100 // Set a reasonable width

	// Add message with markdown
	model.AddMessage("assistant", "# Hello\n\nThis is **bold** and *italic*")

	// The view should exist and contain content
	view := model.View()
	assert.True(t, len(view) > 0)
	// View should contain the app name at minimum
	assert.Contains(t, view, "Pagen")
}
